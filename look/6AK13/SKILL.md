---
name: how-use-to-skillctl
description: 用于指示用户Agent怎么快速使用skillctl
license: MIT
---

# In what use cases is it applicable

agent 自己用本产品的 CLI（skillctl）去下载、安装、更新、搜索、卸载 skill，也可以打开附属文件核对内容。所有命令都支持 `--json`，便于 Agent 提取信息。

## agent 使用方式要求

1. 每条命令都带 `--json`，便于读取报错信息。
2. 每次报错时，先读回复、弄清错误再决定下一步。拿不准就立即结束，把错误原样交回管理层，由管理者决定怎么改，**不要擅自尝试**。
3. **任何附属文件严禁修改。**
4. 想找相关 skill，先看下方「用户查找技能层面」一节。

## 30 秒上手

```
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

## 命令速查

```
skillctl auth register  --email <邮箱> --name <名字> [--password <密码>]
skillctl auth login     --email <邮箱> [--password <密码>]
skillctl auth logout
skillctl version
skillctl skill categories

skillctl skill list      [--installed] [--category <slug>] [--sort newest|downloads|rating] [--page N] [--limit N]
skillctl skill search    <关键词> [--category <slug>] [--limit N]
skillctl skill install   <slug> [--target <agent>] [--version <版本>]
skillctl skill uninstall <slug> [--target <agent>] [--all] [--force]
skillctl skill upgrade   <slug> [--version <版本>]
```

## 快速查找详细使用目录


| 命令 | 在哪 |
|---|---|
| register | 用户登入层面 → 1. register |
| login | 用户登入层面 → 2. login |
| logout | 用户登入层面 → 3. logout |
| search | 用户查找技能层面 → 1. search |
| list | 用户查找技能层面 → 2. list |
| install | 用户查找技能层面 → 3. install |
| uninstall | 用户查找技能层面 → 4. uninstall |
| upgrade | 用户查找技能层面 → 5. upgrade |

## 用户登入层面

### 1. register

```
skillctl auth register --email <用户邮件> --name <名字> --password <用户密码> (--json)
```

成功有两种结果：一种自动登入成功，一种还需手动 login 一次。

个人用户信息查询：

```
cat ~/.skillctl/config.json
```

### 2. login

```
skillctl auth login --email <用户邮件> --password <用户密码> (--json)
```

个人用户信息查询：

```
cat ~/.skillctl/config.json
```

**不得修改这个文件。**

### 3. logout

```
skillctl auth logout
```

只清掉 `~/.skillctl/config.json` 里的 token，环境里留的不动。可以查看上次登入的使用者。

## 用户查找技能层面

这一层分两条指令：`search` 和 `list`。

市场体积还小，**不建议用 `--limit`、`--page`**。

`--category` **只支持 slug，中文不被允许**。查 slug 输入：

```
skillctl skill categories
```

### 1. search

```
skillctl skill search <关键词> (--category <分类的 slug>) (--limit <int> --page <int>) (--sort <newest|downloads|rating>)
```

### 2. list

```
skillctl skill list (--installed) (--category <分类的 slug>) (--limit <int> --page <int>) (--sort <newest|downloads|rating>)
```

`--installed` 用于查看本机已下载 skill 的详细内容，也可以直接 cat `~/.skillctl/registry.json` 自查。

```
cat ~/.skillctl/registry.json
```

**不得修改这个文件。**

### 3. install

```
skillctl skill install <skill 的 slug> (--version <版本>) --target <agent 名称>
```

1. 失败时，请详细看 json 回复。
2. 目前不具备快速检验 skill 功能。
3. 每次下载默认安装最新版本。
4. `--target` 一定得是白名单内的，目前不支持添加白名单。
5. 设有 global 公共目录，可以下载到这里，后续手动添加到 agent 里。
6. 下载后 cat `~/.skillctl/registry.json` 查看详细，**禁止修改**。

### 4. uninstall

```
skillctl skill uninstall <skill 的 slug> (--target <agent 名称>) (--all) (--force)
```

1. 多个 agent 下都有这个 skill 时，输入 `--target` 指定删除哪一个。
2. 所有 agent 都删，用 `--all`。
3. 默认删除下载文件名，不检查文件是否被修改过，统一删除。
4. 用户添加的文件默认保留；输入 `--force` 会直接删掉整个文件夹，用户文件也一并删除。

### 5. upgrade

```
skillctl skill upgrade <skill 的 slug> (--version <版本>)
```

1. 只能用于单个 skill 的全量更新，不具备指定单个 agent 更新的能力。
2. 可用于降级，传 `--version`。
