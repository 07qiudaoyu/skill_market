# skillctl

Skill Market 的命令行客户端：搜索技能、装进 agent 会读的目录、升级、卸载。

装的是**文件**，不是配置。装完之后 agent 下次启动就能发现这个技能，不需要任何额外注册。

---

## 30 秒上手

```bash
# 1. 没有账号先注册（注册成功会顺手登入）；已有账号直接 login
skillctl auth register --email you@example.com --name 小明 --password ******
skillctl auth login    --email you@example.com --password ******
# token 落到 ~/.skillctl/config.json，权限 0600

# 2. 看看市场有什么
skillctl skill list

# 3. 安装。不传 --target 时会自动检测你机器上装了哪个 agent
skillctl skill install go-review

# 4. 看本机装了什么
skillctl skill list --installed

# 5. 有新版本时升级，不想要了就卸载
skillctl skill upgrade go-review
skillctl skill uninstall go-review
```

---

## 发布产物

```bash
make dist          # 交叉编译 6 个平台到 dist/
```

| 文件 | 目标 |
|---|---|
| `skillctl-1.0.0-darwin-arm64` | macOS Apple Silicon |
| `skillctl-1.0.0-darwin-amd64` | macOS Intel |
| `skillctl-1.0.0-linux-amd64` | Linux x86_64 |
| `skillctl-1.0.0-linux-arm64` | Linux ARM64（服务器） |
| `skillctl-1.0.0-windows-amd64.exe` | Windows x86_64 |
| `skillctl-1.0.0-windows-arm64.exe` | Windows ARM64 |

全部 `CGO_ENABLED=0` 静态链接，下载就能跑，不需要装 Go 或任何运行时。
`dist/checksums.txt` 是 SHA-256，用来确认下载的文件没被破坏：

```bash
shasum -a 256 -c checksums.txt
```

版本号由 `make` 注入（`-X go_mod/oneself.Version=1.0.0`），源码里不写死；
`-trimpath` 保证二进制里不含打包机器的绝对路径。

---

## 命令速查

```
skillctl auth register --email <邮箱> --name <名字> [--password <密码>]
skillctl auth login    --email <邮箱> [--password <密码>]
skillctl auth logout
skillctl version

skillctl skill list      [--installed] [--category <slug>] [--sort newest|downloads|rating] [--page N] [--limit N]
skillctl skill search    <关键词> [--category <slug>] [--limit N]
skillctl skill categories
skillctl skill install   <slug> [--target <agent>] [--version <版本>]
skillctl skill uninstall <slug> [--target <agent>] [--all] [--force]
skillctl skill upgrade   <slug> [--version <版本>]
```

全局 flag（所有命令通用）：

| flag | 说明 |
|---|---|
| `--server <url>` | 后端地址，默认 `http://127.0.0.1:8000` |
| `--token <token>` | 临时鉴权，优先于配置文件 |
| `--json` | 输出机器可读的 JSON |

`--version` 和 `version` 子命令的区别：前者只打一行 `skillctl version 1.0.0`，
后者打完整环境信息（Go 版本、平台、后端地址、登入账号）—— 报问题时贴后者。
`--version` 由框架在进入命令之前就处理完，所以 `--json` 对它不生效。

环境变量（优先级高于配置文件）：

| 变量 | 用途 |
|---|---|
| `DR_ACCESS_TOKEN` | 鉴权 token，CI 里用这个 |
| `DR_SKILL_MARKET_TOKEN` | 同上，备用名 |
| `SKILLCTL_SERVER` | 后端地址 |

> 密码没有环境变量入口，只能 `--password` 传。写进命令行会留在 shell 历史里，这是为了少一条取值分支接受的代价；`config.json` 里只存 token，不存密码。

### 分页、排序与搜索的边界

- `--limit` 取值 **1~100**（默认 10），`--page` 从 1 开始。超范围在本地直接报错，不会发给后端
  —— 以前 `--limit=-1` 会让后端回 500，`--limit=0` 会静默返回空列表。
- `--sort` 只有三个值：`newest` / `downloads` / `rating`。写错会报错（以前是静默按时间排）。
- `search` 的关键词只匹配**技能名字**（display_name）和**简介**（summary），**不匹配 slug**。
  slug 是标识符不是描述，所以搜 `demo` 找不到 `big-demo-skill`，搜"大仓库"才能找到。
- `--category` 填的是分类的 **slug**（`web-dev`），不是列表里显示的中文名（Web开发）。
  用 `skillctl skill categories` 查全部 slug —— 填错不会报错，只会得到空列表。
- 列表里的 `DOWNLOADS` 是**去重后的 (用户, 版本) 组合数**，不是总下载次数。
  同一个人反复装同一版本，这个数字不会涨。

### 支持的 target

| target | 安装目录 | 说明 |
|---|---|---|
| `global` | `~/.agents/skills/<slug>` | 通用目录，不参与自动检测 |
| `claude-code` | `~/.claude/skills/<slug>` | |
| `codex` | `~/.codex/skills/<slug>` | |
| `cursor` | `~/.cursor/skills/<slug>` | |
| `qoder` | `~/.qoder/skills/<slug>` | |

---

## 本地状态：`~/.skillctl/`

| 路径 | 内容 |
|---|---|
| `config.json` | token、邮箱、user_id（0600，只有你能读） |
| `registry.json` | 安装账本：哪个技能装在哪个 agent、版本、文件清单 |
| `cache/` | 下载中的 zip，装完就删 |
| `temp/` | 解压校验用的临时目录，装完就删 |

账本结构：

```json
{
  "schemaVersion": 1,
  "skills": {
    "go-review": {
      "codex": {
        "version": "2.0.0",
        "uuid": "uuid-5",
        "path": "/Users/you/.codex/skills/go-review",
        "files": ["SKILL.md", "scripts/check.py"],
        "installedAt": "2026-11-22T00:18:42+08:00"
      }
    }
  }
}
```

`skills[slug][target]` 两层 map：一个技能可以同时装给多个 agent，每个 agent 只有一条记录。
`files` 记的是"这个包里有哪些文件"，用途见下面的安装模型。

---

## 代码结构

```
modules/go_mod/
├── main.go                 只管退出码，不写业务
├── root.go                 命令树：挂载 auth / skill / version，定义全局 flag
├── constant_quantity/      后端地址、HTTP 超时、全局运行期配置
├── security/               config.json 读写、token 解析优先级
├── output/                 输出层：人类可读 / --json 信封、stdout 与 stderr 分流
├── oneself/                auth login、logout、version
└── skills/
    ├── skill_url.go        所有后端接口地址（接口改名只改这里）
    ├── target.go           target 白名单与自动检测
    ├── registry.go         账本读写
    ├── install.go          下载 → 解压 → 校验 → 覆盖 → 记账
    ├── uninstall.go        路径校验 → 删文件 → 清账本
    ├── upgrade.go          查最新版 → 下载一次 → 逐个 target 安装
    ├── list.go             市场列表 / 本机已安装
    └── search.go           关键词搜索
```

三层命令树：`skillctl → skill → install`。每个子命令一个文件、自己 `init()` 挂到父节点上，
`root.go` 只做拼装。加命令时不用改任何已有文件的逻辑。

---

## 四个关键设计

### 1. 安装目录由 agent 决定，不由用户决定

agent 只认自己约定的目录（Claude Code 只读 `~/.claude/skills/`）。
如果允许用户 `--dir` 装到任意位置，结果就是"装成功但 agent 读不到"，
用户看到的是"我明明装了却不起作用"。

所以 target 是**白名单**，路径写死在 `target.go` 里。想支持新 agent，加一行映射，
而不是把自由度丢给用户去猜。

### 2. 一个账本，不是每包一个文件

旧做法是每个技能目录里塞一个 `record.json` 加一个标记文件 —— 那是把元数据混进了内容目录，
卸载时得靠"猜哪些文件是元数据"来排除，也会被 zip 覆盖掉。

现在所有记录集中在 `~/.skillctl/registry.json`：
- 卸载只需要读一个文件
- 升级时能算出"上一版有、这一版没有"的文件
- agent 的 skills 目录里只有技能本身的文件，干净

### 3. 安装 = 覆盖式落地（overlay），不做备份

```
下载 zip → 解压到 ~/.skillctl/temp/<slug>@<target>
         → 校验 SKILL.md 必须在
         → 逐文件覆盖到 <agent 目录>/<slug>/
         → 删除「上一版账本里有、这个包里没有」的文件
         → 最后写账本
```

三个刻意的选择：
- **先校验后落地**：校验失败时目标目录一个字节都没动
- **不做备份、不回滚**：覆盖是幂等的，中途失败重跑一次就收敛了
- **先写磁盘后写账本**：反过来的话，中途失败会留下"账本说有、磁盘上没有"的孤儿记录

用户自己加进技能目录的文件不在 `files` 清单里，所以升级永远不会碰它们。

### 4. 只给二进制，版本号编译进去

用户拿到的是 6 个平台的可执行文件，不需要装 Go。`skillctl version` 能自报版本、
Go 版本、目标平台和后端地址 —— 对方报问题时贴这一段就够了。

---

## 后端接口依赖

| 命令 | 接口 | 鉴权 |
|---|---|---|
| `auth login` | `POST /api/v1/auth/login` | — |
| `skill list` / `search` | `GET /api/v1/skills/search` | 公开 |
| `skill install` / `upgrade` | `GET /api/v1/skills/Cli_downloadzip?slug=&versions=` | Bearer |
| `skill upgrade` | `GET /api/v1/skills/find_versions?public_id=` | 公开 |

注意两件事：

1. 下载接口的版本参数叫 **`versions`（复数）**，写成 `version` 会被静默忽略，永远拿最新版。
2. `public_id` 就是 `skills` 表的 public_id，形如 `uuid-5`；账本里记的 `uuid` 就是它。
   `upgrade` 用它先问出最新版号，已经是最新时**一次下载都不发生**。

下载接口在响应流的尾部写下载计数，所以 CLI 必须把 body 读完 ——
提前中断会导致这次下载不计数。

---

## 开发

```bash
make build     # 当前平台，带版本号
make dist      # 6 个平台 + 校验和
make check     # gofmt + vet + build
```

跑本地后端（注意用项目自己的 venv，系统 Python 缺依赖）：

```bash
~/venv/bin/uvicorn app.main:app --reload     # 8000
```
