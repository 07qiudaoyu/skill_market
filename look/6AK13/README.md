# skillctl

一个命令行工具，管你本机 AI agent 的「技能」：逛市场、搜索、装、升、卸。

不用开网页。一条命令做一件事，所有命令都支持 `--json`，出结构化结果，方便脚本和 Agent 读。

---

## 目录

- [装好之后先确认](#装好之后先确认)
- [30 秒上手](#30-秒上手)
- [命令速查](#命令速查)
- [账号：register / login / logout](#账号register--login--logout)
- [找技能：search / list](#找技能search--list)
- [装技能：install](#装技能install)
- [卸技能：uninstall](#卸技能uninstall)
- [升技能：upgrade](#升技能upgrade)
- [本地文件都在哪](#本地文件都在哪)
- [给 Agent 的硬性约定](#给-agent-的硬性约定)
- [报错时看哪里](#报错时看哪里)

---

## 装好之后先确认

```
skillctl version
```

会打出工具版本、Go 环境、后端地址和登录状态。看到「未登入」是正常的，下一步再登。

> Windows 用户先看 [skillctl-Windows-使用说明.md](./skillctl-Windows-使用说明.md)，选对架构、改名、跑起来，都在那份里。

---

## 30 秒上手

```
# 1. 没账号先注册，注册成功会顺手登入；已有账号直接 login
skillctl auth register --email you@example.com --name 小明 --password ******
skillctl auth login    --email you@example.com --password ******

# 2. 看看市场里有什么
skillctl skill list

# 3. 装一个。不传 --target 会自动检测你机器上装了哪个 agent
skillctl skill install go-review

# 4. 看本机装了什么
skillctl skill list --installed

# 5. 有新版本就升，不要了就卸
skillctl skill upgrade go-review
skillctl skill uninstall go-review
```

---

## 命令速查

```
skillctl auth register  --email <邮箱> --name <名字> [--password <密码>]
skillctl auth login     --email <邮箱> [--password <密码>]
skillctl auth logout
skillctl version
skillctl skill categories

skillctl skill list     [--installed] [--category <slug>] [--sort newest|downloads|rating] [--limit N] [--page N]
skillctl skill search   <关键词> [--category <slug>] [--sort newest|downloads|rating] [--limit N] [--page N]
skillctl skill install  <slug> [--target <agent>] [--version <版本>]
skillctl skill uninstall <slug> [--target <agent>] [--all] [--force]
skillctl skill upgrade  <slug> [--version <版本>]
```

`--server <地址>` 是全局参数，可以临时指定后端，比如：

```
skillctl --server http://<服务端地址> skill list
```

---

## 账号：register / login / logout

### register

```
skillctl auth register --email <用户邮件> --name <名字> --password <用户密码>
```

成功有两种结果：一种注册完直接登入，一种还要再手动 `login` 一次。看到「未登入」就补一句 login。

### login

```
skillctl auth login --email <用户邮件> --password <用户密码>
```

token 会落到 `~/.skillctl/config.json`，权限 `0600`。

### logout

```
skillctl auth logout
```

只清掉 `config.json` 里的 token，环境里的不动。想确认现在是谁在登，`cat ~/.skillctl/config.json` 看一眼。

---

## 找技能：search / list

这两个是同一个层面的两条路：`search` 按关键词，`list` 按分类和热度。

市场现在还小，**不建议用 `--limit` 和 `--page`**，一页能翻完。

`--category` **只认 slug，中文不认**。不知道 slug 先查：

```
skillctl skill categories
```

### search

```
skillctl skill search <关键词> [--category <slug>] [--sort newest|downloads|rating] [--limit N] [--page N]
```

### list

```
skillctl skill list [--installed] [--category <slug>] [--sort newest|downloads|rating] [--limit N] [--page N]
```

`--installed` 看的是本机已装的清单，完整内容在 `~/.skillctl/registry.json`。

---

## 装技能：install

```
skillctl skill install <slug> [--version <版本>] [--target <agent>]
```

几条要紧的：

- **不传 `--target` 会自动检测**你机器上装了哪个 agent，装进去。
- **默认装最新版**。想要旧版才传 `--version`。
- `--target` 必须落在白名单里，白名单目前不能自己加。能填的值跑一遍就知道，填错会报错。
- 有一个 `global` 公共目录，可以装到这里，之后手动挂到你自己的 agent 上。
- 装完去 `~/.skillctl/registry.json` 看详细记录。
- **目前没有「装前检查」**，装完不校验技能能不能跑。

---

## 卸技能：uninstall

```
skillctl skill uninstall <slug> [--target <agent>] [--all] [--force]
```

| 你的情况 | 加什么 |
|---|---|
| 同一个技能装在好几个 agent 下，只想删一个 | `--target <agent>` |
| 所有 agent 下的都删 | `--all` |
| 连你自己往里加的文件一起删掉 | `--force` |

默认行为是**只删下载下来的那些文件，你自己加的保留**。注意它不检查文件有没有被你改过，默认统一删。

`--force` 是连文件夹整个删，用户文件也不留。用之前想清楚。

---

## 升技能：upgrade

```
skillctl skill upgrade <slug> [--version <版本>]
```

- 一次只升一个技能，**不能指定某个 agent**，装了的地方一起升。
- 传 `--version` 可以**降级**，不只是升。
- 不传就升到最新版。

---

## 本地文件都在哪

| 路径 | 是什么 |
|---|---|
| `~/.skillctl/config.json` | 账号 token，权限 `0600` |
| `~/.skillctl/registry.json` | 本机装了哪些技能、装在哪、什么版本 |

两个都是工具自己维护的。**看可以，不要手改。**

---

## 给 Agent 的硬性约定

如果你是个 Agent，拿这些命令干活，按下面四条走：

1. **每条命令都带 `--json`**，输出结构化结果，报错也能读到具体原因。
2. **报错先读回复，再决定下一步。** 自己拿不准的，直接把错误原样交回管理层，**不要自行重试或猜着改**。
3. **任何附属文件严禁修改**——`config.json`、`registry.json`、技能目录里的文件，一个都不许动。
4. 想找技能，先 `skillctl skill categories` 拿 slug，再用 `search` / `list` 定位，别凭印象编 slug。

一句话记：**带 `--json`、报错就停、只读不改。**

---

## 报错时看哪里

先跑这一条，把输出贴给作者：

```
skillctl version
```

版本、Go 环境、系统架构、后端地址、登录状态全在里面，基本能定位一半问题。

剩下的一半：

- 连不上 → 后端没起，或 `--server` 指错了地址
- 「不是内部或外部命令」→ 当前目录没这个可执行文件，`cd` 进去或把目录加进 PATH
- 分类查不到 → 用了中文，换成 `skillctl skill categories` 里的 slug
- 装不上 → 多半是 `--target` 不在白名单，看 `--json` 那条报错里写了哪些能填