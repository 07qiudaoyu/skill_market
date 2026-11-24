# skillctl 改造记录（2026-11-22）

> 这份文档不是使用手册（那本在 `README.md`），而是**改造过程的技术复盘**：
> 每一步到底改了什么、为什么这么改、还有哪些别的做法被排除了、我用什么办法确认它真的对。
>
> 目的只有一个：让没参与实现的人（包括三个月后的你自己）能看懂这些代码为什么长这样。

---

## 0. 怎么读

全文分三个层次，建议按顺序看：

| 层次 | 在哪 | 你得到的 |
|---|---|---|
| **发生了什么** | 第 1 节：改动总表 | 一眼看清范围 |
| **为什么这么设计** | 第 2~5 节：四批改动逐批讲 | 每个决定的取舍和代价 |
| **我用的方法** | 第 6 节：可复用技巧 | 能搬到别的项目上的思路 |

如果时间紧，只看第 1 节 + 第 6 节；要给别人讲，重点讲第 2 节的五个设计决定。

**本次对话开始前，已经完成的东西**（不在本文范围内，但影响后面所有判断）：

- 三层命令树：`skillctl` → `auth` / `skill` → 具体动作
- 五个业务命令：`list` / `search` / `install` / `uninstall` / `upgrade`
- 安装账本 `~/.skillctl/registry.json`，target 白名单，overlay 安装模型

本次对话做的是**收尾四件事**：让输出能被机器读、把源码变成能直接跑的二进制、写交接文档、补一个 `--version`。

---

## 1. 改动总表

按性质分四批。注意"性质"这一列 —— 它能解释为什么改动量差别这么大。

| 批次 | 内容 | 文件 | 性质 |
|---|---|---|---|
| **A** | `--json` 输出层 | 12 个 Go 文件（2 新建，10 修改） | 横切改造：改遍所有命令，但**不动业务逻辑** |
| **B** | 构建与交付 | `Makefile`（新建）、`.gitignore`（+1 行） | 新增能力：从"有源码"到"有产品" |
| **C** | 文档 | `modules/go_mod/README.md`（新建） | 交付物：给三个不同读者看 |
| **D** | `--version` | `root.go`（+5 行）、README（+4 行） | 小增强：补齐一个约定俗成的入口 |

A 批的文件清单：

```
新建  output/output.go          输出层，JSON 信封 / 人机分流
新建  oneself/logout.go         logout 只有 30 行，但需要结构化输出
改写  main.go                   只管退出码
改写  root.go                   全局 flag + JSON 开关
修改  oneself/version.go        version 的 JSON 形态
修改  oneself/login.go          login 的 JSON 形态（不含 token）
修改  skills/target.go          "检测到 xxx" 这句提示改成走统一出口
修改  skills/install.go         InstallResult 加 json tag + JSON 分支
修改  skills/uninstall.go       部分成功要把已完成的数据带回去
修改  skills/upgrade.go         同上
修改  skills/list.go            拆成"取数据"和"打印"两步
修改  skills/search.go          加 JSON 分支
```

一句话概括 A 批：**业务逻辑一行没动，但每个命令的"出口"都被改造了一遍。**

---

## 2. A 批：`--json` 输出层

### 2.1 需求从哪来

问题不是"用户想要 JSON"，而是更具体的两件事：

1. **脚本拿不到数据。** 之前只有表格，脚本想取版本号只能 `awk '{print $3}'`。我只要把列顺序调一下，脚本就全崩 —— 而且是静默崩，输出错的值，不报错。
2. **AI agent 也拿不到。** agent 调 CLI 时同样是解析文本，表格对它是最差的输入格式（列宽随终端变、中文按显示宽度对齐会漂）。

所以目标定得很清楚：**同一份数据，人要能看，机器要能读，两者不能互相污染。**

### 2.2 先做的判断：为什么它该排在最后做

这不是一句"反正都要做"，而是一个次序判断，理由是：

`--json` 是典型的**横切改造**（cross-cutting change）—— 它横穿所有命令，但每个命令里只占几行。改一个命令很快，麻烦在于"必须每个命令都改到"，漏一个就出 bug。

而命令集当时还在变（刚加完 uninstall、upgrade）。如果边加命令边改输出，每加一个命令都要同时想两套输出，注意力被劈成两半：一半想"功能对不对"，一半想"格式对不对"。

排到最后的好处：**命令集定型后，输出改造只需要做一遍。** 以后加新命令，照着现成的三个套路抄就行（见 2.6）。

代价是：如果命令集一直不稳定，这个改造会一直往后推。这是真实的取舍，不是没有成本。

### 2.3 五个设计决定

这一节是本文的核心。每个决定我都写清楚：问题是什么、考虑过哪些别的做法、为什么选这个、代价在哪。

---

#### 决定一：`--json` 放全局，不放每个命令

**问题**：8 个命令，如果每个都定义一次 `--json`，要写 8 遍；以后加第 9 个命令，准会忘。

**做法**：cobra 的 `PersistentFlags()` —— 定义在根命令上，整棵树继承。

```go
rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "输出 JSON（给脚本 / agent 消费）")
```

**但这里立刻撞上一个时序问题。** cobra 执行一个命令的顺序是：

```
解析 flag  →  校验参数个数(Args)  →  PersistentPreRunE  →  RunE  →  返回 error  →  main
```

我原本打算在 `PersistentPreRunE` 里把开关转交给输出包：

```go
PersistentPreRunE: func(...) error {
    output.SetJSON(flagJSON)   // 看起来没问题
    return nil
}
```

**坑在哪**：如果用户忘了给参数 —— `skillctl skill install --json`（没写 slug）—— cobra 会在 **Args 校验阶段就失败**，`PersistentPreRunE` 根本不会执行。可这时 flag 早就解析完了，用户的意图非常明确：他就是要 JSON。

结果就是：一个本该结构化的错误，变成了一行人类可读的红字。脚本抓过来解析，直接挂。

**修法**：给 `main.go` 一个不依赖 PreRunE 的查询入口，直接读 flag 本身：

```go
func JSONRequested() bool {
    v, err := rootCmd.PersistentFlags().GetBool("json")
    return err == nil && v
}
```

注意它读的是 `rootCmd` 的 flag，**不是当前命令的**。因为 flag 已经解析完了，值就存在那里，谁读都一样。

**这个坑的通用教训**：用框架的"钩子"（PreRun、beforeEach、middleware）时，先搞清楚它在整个执行链里的**位置**，而不是假设"它总会在业务代码之前跑"。凡是"校验失败就直接短路"的框架，钩子都可能被跳过。

---

#### 决定二：用信封，不用裸 JSON

**三种常见做法**：

| 做法 | 长什么样 | 问题 |
|---|---|---|
| A. 裸 JSON | 成功：直接返回数据<br>失败：打一行纯文本错误 | 脚本得先试着解析；"成功但列表为空"和"失败"长得一样（都是空输出） |
| B. 永远返回数据 + 状态码字段 | `{"code":0,"msg":"ok","data":...}` 无论成败 | 退出码就能判断成败，没必要再绕一层 |
| C. 信封 + 退出码 | 成功 `{"ok":true,"data":...}`<br>失败 `{"ok":false,"error":{...}}` | 两个通道各司其职（我选的） |

**为什么选 C**：因为"成败"和"细节"是两类信息，该走两个通道：

- **退出码**：`echo $?` → `0` / `1`。这是操作系统给的、最可靠的信号，shell 的 `if` 直接能用。
- **JSON 内容**：装了什么、删了几个、为什么失败。

用信封而不是裸 JSON，是为了让"成功"和"失败"**在结构上就长得不一样**。脚本看一眼有没有 `ok` 字段就知道该往哪走，不需要猜。

**代价**：多一层嵌套，`jq` 里要写 `.data.items` 而不是 `.items`。我认为值得 —— 省下的那三个字符，换来的是不会误判。

**一个细节**：`data` 字段在成功时永远存在，哪怕内容为空。这样脚本可以无条件写 `.data`，不用判 null。

---

#### 决定三：stdout 只放结果，stderr 放提示和错误（最重要）

这是 Unix 四十年来的约定，但第一次写 CLI 的人几乎都会踩，因为**在终端里看不出区别** —— stdout 和 stderr 都打在屏幕上，混在一起很正常。

一旦走管道，区别就致命了：

```bash
skillctl skill list --json | jq '.data.total'
```

如果我把"正在加载…"或"检测到 codex"打进 stdout，`jq` 立刻报：

```
jq: error (at <stdin>:1): parse error
```

**这时脚本作者的第一反应是"你的 JSON 坏了"，而不是"你的提示混进来了"。** 排查方向就错了。

所以这么做：

| 通道 | 放什么 | 谁消费 |
|---|---|---|
| stdout | 只有最终结果（JSON 或表格） | 管道、`jq`、agent |
| stderr | 提示语、进度、错误 | 人的眼睛 |

**真正的工程量不在这条约定，而在于让所有人类提示走同一个出口。**

概念上很简单，加一个函数：

```go
func Say(format string, a ...any) {
    if jsonMode {
        fmt.Fprintf(os.Stderr, format+"\n", a...)  // JSON 模式：改道 stderr
        return
    }
    fmt.Printf(format+"\n", a...)                  // 普通模式：照常打屏幕
}
```

它的价值是：**分流逻辑只写一次。** 如果每条命令各自 `fmt.Printf`，我得改 20 处，漏一处就出 bug。有了 `Say`，改 20 处变成"把 20 处 `fmt.Printf` 换成 `output.Say`"，是机械替换，不会漏逻辑。

**最容易漏的一处值得单独说**：`target.go` 里的

```go
output.Say("检测到 %s，本次安装到 %s", name, dir)
```

这行藏在 `Install()` 内部两层调用之下，从命令层（`install.go` 的 `RunE`）**完全看不到**。如果只顺着 `RunE` 改，这行必漏，而它一漏，`install --json` 的输出就是"一行中文 + 一个 JSON"，`jq` 直接挂。

**这就是为什么我做横切改造时先 grep、不靠记忆。** 具体做法见第 6 节技巧 1。

---

#### 决定四：失败走 stderr，且 stdout 一个字节都不写

失败时的样子是这样的：

```
退出码  : 1
stdout  : （0 字节，什么都没有）
stderr  : {"ok":false,"error":{"message":"未登入或缺少 token，请先 skillctl auth login"}}
```

**为什么不让错误也进 stdout？** 为了让 shell 逻辑只有一条路径：

```bash
if out=$(skillctl skill list --json); then
    echo "$out" | jq '.data.items[]'
else
    echo "失败了"        # 不需要解析任何东西，退出码已经说明
fi
```

**别的做法也有**：AWS CLI 就把错误 JSON 打到 stdout，调用方必须读 `ok` 字段才能判断成败。两种都能工作，但我选前者的理由是"不与退出码争夺判断权" —— 一条命令的成败，永远只看退出码，不看你解析到了什么。

**代价**：如果有人在 `2>/dev/null` 的情况下跑，失败时什么都看不到，只有一个非零退出码。对脚本来说这没问题（它本来就该看退出码）；对在终端里调试的人来说，需要知道别随便丢弃 stderr。

---

#### 决定五：部分成功，批量操作独有的问题

单条命令要么成功要么失败。但 `upgrade` 是**遍历一个技能的多个安装位置**（同一个技能可能同时装在 claude-code 和 codex 上），天然会出现"升了 2 个，第 3 个账本记录坏了"。

这时候如果只返回错误，会发生什么：

- 前两个 target 其实已经升级完、账本已经写了
- 但脚本只看到"失败"，以为什么都没发生
- 重跑一遍（幂等，结果正确，但白跑一次）

所以错误信封里加一个可选的 `data`：

```json
{
  "ok": false,
  "error": { "message": "账本里 skill-x 记在 \"foo\" 上，但它已不是合法 target..." },
  "data":  { "count": 2, "items": [ ... 已经成功的两个 ... ] }
}
```

**但这引出一个隐藏问题：错误会被打印两次。**

cobra 的标准模式是"`RunE` 返回 error，`main` 统一打印"。现在 `RunE` 已经自己输出过一次结构化错误了，`main` 收到 error 又会输出一遍。结果 stderr 上出现**两个 JSON 对象**：

```
{"ok":false,"error":{...}}
{"ok":false,"error":{...}}      ← 多出来的
```

`jq` 解析第一个就成功，然后报 `invalid character '{' after top-level value` —— 跟 stdout 被污染是同一类事故。

**先想到的错误做法**：搞一个全局变量 `已打印 = true`，`main` 看到就跳过。

**为什么不行**：这是**隐式状态**。一个谁都能改的包级变量，main 靠它和子命令通信。将来某处不小心设了一下，或者并发跑，行为就变得不可预测，而且极难排查 —— 因为"它该不该打印"这件事在代码里没有类型层面的表达。

**正确做法**：用一个**类型**来表达"这个错误已经输出过了"。

```go
// PrintedError 表示错误信息已经输出过了，上层只需要设置退出码
type PrintedError struct{ Err error }

func (e PrintedError) Error() string { return e.Err.Error() }
func (e PrintedError) Unwrap() error { return e.Err }
```

`main` 里：

```go
var printed output.PrintedError
if errors.As(err, &printed) {
    os.Exit(1)          // 已经输出过了，只设退出码
}
```

这是 Go 处理"错误需要有类别"的标准手法：**哨兵类型 + `errors.As` + `Unwrap`**。

- `errors.As` 能穿过包装层找到它（因为实现了 `Unwrap`）
- 它是**显式的**：只有明确返回 `PrintedError` 的地方才会被跳过打印，别的地方照常
- 编译器帮你查：类型写错立刻编译失败，不像全局 bool 那样静默

**这里有一条可以通用的判断标准**：当你想用一个 `bool` 或 `string` 在两层代码之间传"状态"时，先想想能不能用一个**类型**。类型能被编译器检查，bool 不能。

---

### 2.4 逐文件说明

按"改了什么 → 为什么"排列。

#### `output/output.go`（新建，核心）

整个输出层的全部内容，约 75 行：

| 成员 | 作用 |
|---|---|
| `jsonMode` / `SetJSON` / `Enabled` | 全局开关 |
| `Say` | 人类提示的统一出口，JSON 模式下改道 stderr |
| `successEnvelope` / `errorEnvelope` | 两种信封的结构定义 |
| `Data(v)` | 成功信封 → stdout |
| `Fail(err)` | 失败信封 → stderr |
| `FailWithData(err, data)` | 失败信封带已完成部分，返回 `PrintedError` |
| `PrintedError` | 已输出标记 |
| `encode(f, v)` | 统一编码：2 空格缩进 + 关闭 HTML 转义 |

**为什么单独建一个包，不塞进 `security` 或 `constant_quantity`？**

这两个包一边管凭据、一边管地址，都已经被 `skills` 和 `oneself` 引用。把输出塞进去能跑，但会造成"业务包因为一个输出函数而依赖安全模块"这种看不懂的耦合 —— 半年后有人问"为什么 skills 依赖 security？"答案是"为了拿一个打印函数"，这不合理。

输出是横切关注点，独立成包最干净。代价只是多一个目录。

`encode` 里那个 `SetEscapeHTML(false)` 下面单独讲。

#### `root.go`（改写）

三件事：

```go
PersistentPreRunE: func(...) error {
    security.FlagToken = flagToken
    constantquantity.FlagServer = flagServer
    output.SetJSON(flagJSON)      // ①②③ 都在这里
    return nil
}
```

再加 `JSONRequested()`（见决定一）和 `--version`（见第 5 节）。

**为什么 token 和 server 也是在这里转交？** 因为它们和 JSON 一样，是"全局 flag → 包级变量"的转交。放在同一个地方，读代码的人只需要看一处就知道"全局配置在哪落地"。

#### `main.go`（改写）

现在只有 8 行逻辑，但它是所有错误的最终出口：

```go
func main() {
    if err := Execute(); err != nil {
        var printed output.PrintedError
        if errors.As(err, &printed) {
            os.Exit(1)                 // 子命令已经输出过

        }
        if JSONRequested() {
            output.Fail(err)           // JSON 信封 → stderr
        } else {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        }
        os.Exit(1)
    }
}
```

三个分支各对应一种情况，顺序不能换：先看"是否已输出"，再看"要不要 JSON"，最后才是人类可读。

#### `skills/list.go`（改动最有代表性）

改前，这个函数把两件事焊在一起：

```go
func listInstalled() error {
    reg, _ := LoadRegistry()
    for ... { rows = append(rows, ...) }      // 取数据
    fmt.Printf("已安装 %d 个\n", len(rows))   // 直接打印
    t.Render()                                // 直接打印
}
```

它对外的能力就是"打印一张表"，没有第二种用法。要出 JSON，只能复制一份循环逻辑 —— 那就是两份代码，以后改排序得改两处，早晚不一致。

改后拆成两个函数：

```go
func collectInstalled() ([]installedItem, error)   // 只取数据
func printInstalled(items []installedItem)         // 只画表格
```

`RunE` 里先拿到 items，再决定怎么呈现：

```go
items, err := collectInstalled()
if err != nil { return err }

if output.Enabled() {
    return output.Data(installedPayload{Count: len(items), Items: items})
}
printInstalled(items)
return nil
```

**判断标准可以记住这一条**：*当同一份数据有两种呈现方式时，"取数据"和"呈现"就必须分开。*

这跟后端的 `router → service → repository` 是同一个道理：service 不写 SQL，不是因为它不会，而是因为它只负责"数据怎么来"，不负责"数据怎么存"。职责一混，两种需求就开始互相牵制。

**顺带一个细节**：`installedItem` 里的字段比表格列多。表格有 6 列（slug/target/版本/文件数/路径/时间），JSON 里还有 `uuid`。**JSON 的字段可以比表格多** —— 因为读者不同，表格怕挤，JSON 不怕。

#### `skills/install.go` / `uninstall.go` / `upgrade.go`

这三个的改法完全一样，都是三段式：

1. Result 结构体加 `json` tag（字段名改成小驼峰：`Target` → `"target"`）
2. `RunE` 里最前面加分支
3. 人类可读的输出留在分支后面（这样就不需要改动它们）

以 install 为例：

```go
res, err := Install(slug, version, targetFlag)

if output.Enabled() {
    if err != nil {
        return err                            // 让 main 统一发失败信封
    }
    return output.Data(res)                   // 成功信封
}

// ↓ 以下完全不用改，原来怎么写还怎么写
fmt.Printf("已安装 %s %s → %s\n", ...)
```

**这么做的好处**：原来那 5 行人类可读输出的代码**一行都没动**。改造是"在前面加一个出口"，不是"改写原有逻辑"。改动量小，回归风险也小。

`uninstall` 和 `upgrade` 因为可能部分成功，所以走的是 `FailWithData` 而不是 `return err` —— 这是决定五的内容。

#### `oneself/login.go`

特意说明一点：**`loginResult` 里没有 token 字段。**

```go
type loginResult struct {
    Email  string `json:"email"`
    UserID int    `json:"userId"`
}
```

为什么不带上？因为 token 是凭据。一旦进入 JSON，它就可能被：贴到 issue 里、被 CI 日志持久化、被 agent 记进上下文、被 `> out.json` 落到磁盘。而这些场景里没有一个需要 token —— 它已经存进 `config.json` 了。

**"最小暴露"原则**：默认不输出敏感字段，需要时再显式导出（比如 `config.json` 就在那里，想看自己看，而且有 0600 权限保护）。

#### `oneself/logout.go`（新建）

只有 30 行，为什么还要单独一个文件？因为它需要**结构化输出**：

```go
type logoutResult struct {
    Server string `json:"server"`
    Email  string `json:"email"`
}
```

脚本想知道"刚才登出的是哪个账号"，就不会只有一行中文可以抓。小功能也值得有自己的 Result 类型 —— 一致性比省文件更重要。

#### `skills/target.go`

只改了一行 `fmt.Printf` → `output.Say`。看起来微不足道，但它就是决定三里说的"藏在两层调用下面、最容易漏"的那一处。单独列出来，是为了强调：**横切改造的验收，不能只测最外层。**

#### `skills/search.go`

加了一个 JSON 分支。这里没有 Result 结构体，因为 `SearchSkills` 返回的 `SkillSearchResponse` 本身就是后端的响应结构，字段齐全，直接透传就行：

```go
if output.Enabled() {
    return output.Data(resp)
}
printMarketList(resp)
```

**这里有一个可以记住的判断**：如果数据本来就是结构化的（来自 API 的响应），直接透传；只有在数据是"本地拼装出来给人看的"时候，才需要单独的 Result 类型。

---

### 2.5 A 批里另外三个坑

#### 坑一：`null` 和 `[]` 不是一回事

Go 里 `var x []T` 是 nil 切片，`json.Marshal` 出来是 `null`：

```json
{"count": 0, "items": null}
```

脚本写 `.items[]` 时，`null` 会报错：

```
jq: error: Cannot iterate over null
```

而 `[]` 不会。所以列表**必须**初始化为空切片：

```go
items := make([]installedItem, 0)     // 不是 var items []installedItem
```

这不是理论问题 —— "本机没装任何技能"是最常见的第一次使用场景，恰好就是这个分支。

#### 坑二：`SetEscapeHTML` 默认开着

Go 的 `json.Encoder` 默认把 `<` `>` `&` 转义成 `\u003c` `\u003e` `\u0026`。这是为了安全地往 HTML 里嵌 JSON，但我们的输出是给终端和脚本看的，不嵌 HTML。

不改的话，路径或技能描述里带个 `<` 就变成一堆转义码，人能读但很痛苦，而且容易让人误以为编码坏了。

```go
enc.SetEscapeHTML(false)
```

**语义等价，但可读性差很多。** 这类"默认值是为另一个场景设的"要特别注意 —— 用一门技术前先问"这个默认值是为谁设的"。

#### 坑三：缩进 2 空格是我主动选的

其实脚本不在乎缩进（`jq` 会重新排版）。选 2 空格是因为这份输出**首先被人看** —— 用户在终端里直接看到它，或者贴到对话里给 agent 看。单行 JSON 在终端里会被折行折断，非常难读。

代价是多了几十个字节。忽略不计。

---

### 2.6 验收方式和结果

**验收思路**：不看代码，只看行为。分三类场景，每类都要有"成功的"和"失败的"。

| 类别 | 场景 | 期望 |
|---|---|---|
| 成功路径 | `version` / `list` / `list --installed` / `search` / `install` / `upgrade` / `login` / `logout` 带 `--json` | `ok:true` + 数据完整 |
| 失败路径 | token 缺失、slug 不存在、版本不存在 | stdout **0 字节**、stderr 是 JSON、退出码 1 |
| 部分成功 | `uninstall --all` 中途失败 | `ok:false` 但 `data` 里有已完成的项 |

另外三条"体检"：

```bash
# 1. stdout 是否纯净（最关键的一条）
./skillctl skill list --json 2>/dev/null | python3 -m json.tool
# 能解析 → 说明没有杂质；报错 → 说明有提示语混进去了

# 2. 空列表是否是 []
HOME=$PWD/.tmp-home ./skillctl skill list --installed --json | grep items

# 3. token 有没有漏
./skillctl auth login --json | grep -c token     # 必须 0
```

全部通过。**其中第 1 条是这套设计唯一的"一票否决"指标** —— 只要 stdout 能整段通过 `json.tool`/`jq`，说明分流是对的。

**以后加新命令时，记住三个固定动作**：

1. `RunE` 最前面加 `if output.Enabled() { return output.Data(res) }`
2. 所有人类提示用 `output.Say`，不用 `fmt.Printf`
3. 列表用 `make([]T, 0)`

---

## 3. B 批：构建与交付

### 3.1 为什么这一步是"必须做"，不是"锦上添花"

被打回的第一条原因是：**只给了源码，没给能跑的东西。**

对写 Go 的人，`go build` 是一秒钟的事。但对使用者，这意味着：

1. 得先装 Go（几百 MB，还得配对版本）
2. 得知道在哪执行 `go build`
3. 编出来的东西放哪、怎么进 PATH
4. 换个平台还得再编一次

**任何一步没做好，用户就会卡住，然后来问你 —— 而这个问题本来不该存在。** 交付二进制，本质上是把"环境搭建"的责任从使用者身上拿回来。

### 3.2 Makefile 逐行讲

```make
BINARY  := skillctl
VERSION ?= 1.0.0
LDFLAGS := -s -w -X go_mod/oneself.Version=$(VERSION)
DIST    := dist
```

`VERSION ?= 1.0.0` 里的 `?=` 是"没传才用默认值"，所以可以覆盖：

```bash
make dist VERSION=1.1.0
```

一个 Makefile 管全年，不用改文件。

`LDFLAGS` 里三个参数各有分工，这是理解 Go 构建的关键：

| 参数 | 作用 | 为什么需要 |
|---|---|---|
| `-s` | 去掉符号表 | 体积小一点，也稍难被逆向 |
| `-w` | 去掉 DWARF 调试信息 | 同上。两个合起来通常省 20~30% |
| `-X pkg.Var=value` | **链接期修改字符串变量的值** | 版本号不用写进源码 |

`-X` 是最有意思的一个。它做的事是：在编译已经完成、链接还没结束的时候，直接改写某个包里某个 `string` 变量的初始值。

所以 `oneself.Version` 在源码里是：

```go
var Version = "dev"        // 源码里永远是 dev
```

编完之后二进制里是 `1.0.0`。好处：

- 源码里不出现版本号 → 不用每次发版都改文件、不用怕忘了改
- 也不生成 `version.go` 这种由工具写入的文件（那会让 git 每次发版都有噪音）
- `go run` 的时候还是 `dev`，一眼能看出"这不是发布版"

```make
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64
```

六个组合。为什么是这六个：

- macOS：M 系列（arm64）+ Intel（amd64），都还大量在用
- Linux：服务器基本是这两个架构
- Windows：同理

有意识地**没做** 32 位（`386`）、FreeBSD、Linux 的 `mips`/`riscv64` —— 目标用户里几乎不存在，编出来是浪费。范围要克制，否则发布目录会变成二十个文件，反而没人知道该下哪个。

```make
dist: dist-clean
	@mkdir -p $(DIST)
	@set -e; for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		out="$(DIST)/$(BINARY)-$(VERSION)-$${os}-$${arch}$${ext}"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			go build -trimpath -ldflags "$(LDFLAGS)" -o "$$out" .; \
	done
```

几个 Makefile 的坑值得记：

- **`$$` 是转义**。Makefile 里 `$` 是它自己的变量语法，要传给 shell 得写 `$$`。所以 `$${p%/*}` 到 shell 里才是 `${p%/*}`（去掉最后一个 `/` 及之后，取 OS）。
- **`@` 前缀**表示"不打印这条命令本身"，只打印它的输出。否则屏幕上全是命令回显。
- **`set -e`** 让循环中任何一条失败就立刻中断。不加的话，某个平台编译失败会被忽略，最后你以为成功了 —— 但 `dist/` 里少一个文件。**这是"看起来成功实际上失败"的典型**。
- **`dist-clean` 单独一个 target**，不是复用 `clean`。因为 `clean` 会连 `skillctl`（开发用的那个）一起删掉，每次打包都重编一遍开发版没必要。

```make
check: fmt vet build
fmt:
	gofmt -l . | grep . && echo "上面这些文件需要 gofmt -w" && exit 1 || true
```

`gofmt -l` 列出"格式不合规"的文件，什么都不输出就代表干净。所以这里用 `grep .`（匹配任意字符）来判断"有没有输出"：

- 有输出 → `grep` 返回 0 → 打印提示 → `exit 1`（失败）
- 无输出 → `grep` 返回 1 → `|| true` 兜住，让整条命令成功

这是 shell 里"如果没有输出就算通过"的标准写法。

### 3.3 三个构建参数，为什么都必要

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags "..." -o out .
```

**`GOOS` / `GOARCH`**：交叉编译。Go 自带全套目标平台的标准库，`go build` 时按这两个变量选择，就得到对应平台的机器码。不需要装虚拟机、不需要 Docker、不需要目标机器。

这也是 Go 相对 C/C++ 的巨大优势 —— 后者交叉编译要配交叉工具链，往往是项目里最麻烦的部分。

**`CGO_ENABLED=0`**：关掉 C 语言互操作。默认情况下 Go 某些包（比如 DNS 解析）可能链接系统 libc，那就不是纯静态的了。

关掉之后：

- Linux 上得到真正静态链接的 ELF，`file` 会明确写 `statically linked`
- 拿过去就能跑，不怕目标机器 glibc 版本老

**代价**：DNS 解析改用 Go 自己的实现（纯 Go resolver），行为跟系统 resolver 略有差别（比如不读某些 `nsswitch.conf` 配置）。对一个只访问固定后端地址的 CLI 完全无所谓。

这个取舍值得记：**静态链接换来"到处能跑"，代价是"不跟随系统配置"**。多数分发场景选静态。

**`-trimpath`**：把二进制里记录的源文件路径，从绝对路径改成模块内路径。

不加的话，二进制里会写进去 `/Users/07deqiudaoyu/Desktop/skill_market/...`。这不影响运行，但有两个问题：

1. **泄露你的用户名和目录结构** —— 交付给别人的文件里带着你的本机信息
2. 别人的 panic 堆栈里会显示一个他机器上不存在的路径，看着莫名其妙

改完之后，堆栈里是 `go_mod/skills/install.go`，干净且有意义。

### 3.4 产物与验收

六个文件 + 一个校验和：

| 文件 | `file` 实测输出 | 体积 |
|---|---|---|
| darwin-arm64 | Mach-O 64-bit executable arm64 | 7.1M |
| darwin-amd64 | Mach-O 64-bit executable x86_64 | 7.7M |
| linux-amd64 | ELF 64-bit LSB, x86-64, **statically linked** | 7.5M |
| linux-arm64 | ELF 64-bit LSB, ARM aarch64, **statically linked** | 6.9M |
| windows-amd64.exe | PE32+ executable (console) x86-64 | 7.8M |
| windows-arm64.exe | PE32+ executable (console) Aarch64 | 7.1M |

**为什么 7MB 这么大？** 一个 "Hello World" 的 Go 程序约 1.5MB。多出来的主要是 `net/http`（完整的 HTTP 客户端 + TLS 实现）和 `crypto` 那一堆曲线、哈希算法。这部分是 Go 静态链接的固定开销，`-s -w` 已经省了一部分。

对 CLI 来说 7MB 完全可接受（对比：Node.js 的 CLI 要先装个 80MB 运行时）。真要压到 2~3MB 可以用 UPX，但会让杀毒软件误报，不值。

**四个验收点**，每个都在验证一个具体假设：

| 验收 | 命令 | 验证的假设 |
|---|---|---|
| 平台对不对 | `file dist/*` | 交叉编译真的产出了目标格式，不是把 darwin 二进制改了个名字 |
| 是否静态 | 上面输出里的 `statically linked` | 目标机器不需要 libc 版本匹配 |
| 路径有没有漏 | `strings dist/*-linux-arm64 \| grep -c "/Users/07deqiudaoyu"` → **0** | `-trimpath` 生效了 |
| 校验和自洽 | `shasum -a 256 -c checksums.txt` → 六个 OK | 传输过程中文件没坏 |

第二个"strings 检查"是我专门加的。原因是：`-trimpath` 有没有生效，**从命令是否成功完全看不出来** —— 它不加也能编过。只能去二进制里搜。这类"静默生效"的参数，必须有独立的验证手段。

### 3.5 一个容易忘记的事

**改了源码之后，`dist/` 里的二进制就旧了。**

这次加 `--version` 就遇到了：改完 `root.go`，`make build` 出了新的 `skillctl`，但 `dist/` 里六个还是老的（里面没有 `--version`）。如果不重新 `make dist` 就交付，对方会发现"文档写了 `--version`，但我的程序不认识"。

**规则**：交付前一定确认"产物比最近的源码改动新"。粗糙但有效的办法是发版时永远从 `make dist` 开始，别手动复制。

---

## 4. C 批：文档

### 4.1 这份文档要服务三种读者

写文档最容易犯的错是"按代码顺序写"，结果对谁都不好用。这次先定读者：

| 读者 | 他关心什么 | 对应章节 |
|---|---|---|
| **想用的人** | 怎么装上、常用命令是什么 | 30 秒上手、命令速查、target 表 |
| **要接手/讲解的人** | 代码怎么组织、关键设计为什么这样 | 代码结构、四个关键设计 |
| **报问题的人** | 我需要提供什么信息 | `skillctl version` 的输出说明 |
| **前后端都要看的人** | CLI 依赖哪些接口、有什么坑 | 后端接口依赖 |

同一份文档服务四种读者是有代价的（会长），但拆成四份文件更糟 —— 没人会去翻第二份。

### 4.2 结构上的几个刻意安排

**① 开头是"30 秒上手"，不是"项目简介"。**

读者第一眼想知道的是"这东西我要怎么用"，不是"它的架构理念"。简介可以放后面，甚至不放。

**② 命令速查用代码块堆叠，不用表格。**

表格适合"字段 → 说明"这种二维关系；一长串命令用表格反而挤。所以命令用代码块（复制方便），flag 和 env 用表格（对照查找方便）。

**③ 账本给了一个完整的 JSON 实例。**

比写十行"包含 version、uuid、path、files 字段"有效得多。真实的例子自带结构、层次和数据类型，读者一眼就知道该怎么解析。

**④ "四个关键设计"直接对应被打回的四个问题。**

这不是巧合 —— 文档要能回答问题，而对方脑子里只有那四个问题：

| 他的问题 | 文档里对应的节 |
|---|---|
| 只给源码没二进制 | 发布产物 |
| 命令缩写乱 | 命令速查（全称命令 + 统一前缀） |
| `--dir` 任意目录 | 设计 1：安装目录由 agent 决定 |
| 每包一个 record.json | 设计 2：一个账本 |

**写文档时先想"对方会问什么"，比想"我实现了什么"有效。**

### 4.3 文档里埋的两条"防踩坑"提示

写在"后端接口依赖"一节，因为它们是跨系统约定，最容易在别人手里出错：

**① 版本参数叫 `versions`，复数。**

```go
q.Set("versions", version)     // 不是 "version"
```

写错不会报错 —— FastAPI 那边参数对不上就当没传，静默返回最新版。于是出现最难受的 bug：**用户指定了版本，装上的却是最新版，而且没有任何提示。**

这类 bug 的可怕之处是"静默"。所以文档必须显式写出来。

**② `public_id` 就是账本里的 `uuid`。**

这个关系此前是模糊的（一度以为需要先下载才能知道版本）。确认之后，`upgrade` 才能做到"先问版本，已经是最新就一次下载都不发生"。

文档里写出来，是为了让接手的人不用重新推导一遍。

---

## 5. D 批：`--version`

### 5.1 为什么明明有 `version` 子命令，还要加 `--version`

因为**这是肌肉记忆**。`go --version`、`docker --version`、`node --version`（虽然报错）、`python --version` —— 所有人第一反应都是敲这个。

`skillctl --version` 之前会怎样？cobra 会认为 `--version` 是个不认识的 flag，报：

```
Error: unknown flag: --version
```

然后打印一大段帮助。**用户的第一印象就是"这工具很粗糙"。** 加这 5 行的收益，比它看上去大得多。

### 5.2 cobra 的执行顺序，决定了怎么加

cobra 里 `--version` 不是普通 flag，它有特殊处理。执行顺序是：

```
解析 flag
  ↓
--help ？  → 打印帮助，返回
  ↓
--version ？ → 打印版本，返回          ← 在这里就结束了
  ↓
参数个数校验
  ↓
PreRunE → RunE
```

所以只要设置 `rootCmd.Version`，cobra 就自动接管 `--version`，连 `RunE` 都不用写。

### 5.3 为什么要"提前注册"这个 flag

cobra 的默认行为是：发现 `Version` 非空且还没有 `version` flag，就自动加一个，说明文字是英文的 `version for skillctl`。

我们的 CLI 其他地方都是中文，突然冒出一句英文不协调。所以**抢在 cobra 前面自己注册**：

```go
rootCmd.Version = oneself.Version
rootCmd.Flags().Bool("version", false, "显示版本号(详细环境信息用 skillctl version)")
```

cobra 的逻辑是"如果已经存在同名 flag 就不再加"，所以它看到我们注册了，就跳过，保留我们的中文说明。

**这是一个通用手法**：框架的"自动补全"行为，通常都允许你先手动做一遍来接管。找一下它判断的条件（这里是"flag 是否存在"），提前满足条件就行。

### 5.4 四个验收点

| 场景 | 结果 | 说明 |
|---|---|---|
| `skillctl --version` | `skillctl version 1.0.0` | 主目标 |
| `skillctl version` | 多行详情 | 没被破坏 |
| `skillctl skill install --help` | `--version string 指定版本` | **子命令的 flag 没被抢** |
| `skillctl --version --json` | 仍是纯文本 | 见下 |

第三个点值得说明：`skill install` 自己也有一个 `--version`（指定技能版本）。两者不冲突，因为：

- 根命令的 `--version` 是 `rootCmd.Flags()`（本地 flag），**只在根命令上有效**
- 子命令的 `--version` 是它自己的本地 flag

如果当初写成 `PersistentFlags()`，根命令的 `--version` 会**继承到所有子命令**，那 `skill install x --version 1.0.0` 就歧义了（是"显示版本"还是"装 1.0.0 版"？）。cobra 会优先用子命令自己的（因为本地优先于继承），但这是靠优先级"碰巧"没出事，不是设计上干净。

**规则：flag 的作用域要按"谁需要"来定，不要图省事一律用 Persistent。**

第四个点是一个**行为边界**，不是 bug：

```bash
$ skillctl --version --json
skillctl version 1.0.0
```

因为 `--version` 在"参数校验"之前就被 cobra 处理并返回了，根本走不到我们的 JSON 分支。`docker --version` 也一样不吃 `--json`。

我在文档里明确写了这一点，因为**边界要写在文档里，不能靠用户猜**。要 JSON 版本信息，用 `skillctl version --json`（走子命令，正常）。

---

## 6. 可复用的技巧（重点）

前面讲的是"这次怎么做的"，这节是"下次还能怎么用"。

### 技巧 1：横切改造靠搜索，不靠记忆

**场景**：要改的东西散落在所有文件里（比如所有 `fmt.Printf` 都要换成 `output.Say`）。

**做法**：动手前先 grep 全量，拿到完整清单，再逐个判断。这次是先搜 `fmt.Print`，拿到 25 处匹配，然后分类：

- 在 JSON 路径上的 → 必须改
- 只在人类可读路径上的 → 可以不改
- 表格渲染内部的 → 不动（那是呈现逻辑本身）

**为什么不能靠记忆**：人会记得"我改过的那几处"，记不住"藏在我没想到的地方的那一处"。`target.go` 里那句提示就是典型 —— 它在两层调用之下，从命令入口根本看不到。

**判断标准**：如果这次改动的正确性依赖"我记得改全了"，那它就不可靠。必须有一个**外部的、完整的清单**（grep 结果）来兜底。

### 技巧 2：动手前先画"执行顺序"

**场景**：要在框架的某个钩子里挂东西（PreRun、middleware、生命周期函数）。

**做法**：先搞清楚框架的执行链，再决定挂哪。

cobra 的链是 `解析 flag → 校验 → PreRun → Run → error → main`。知道这个之后立刻能看出两个结论：

- 参数校验失败时 PreRun 不执行 → 不能让 PreRun 承担"必须发生"的职责
- `--version` 在校验前就返回 → 它不可能吃到 JSON 分支

**两个 bug 都是"没画顺序图"造成的。** 这次的 `JSONRequested()` 就是补这个的。

### 技巧 3：两种呈现 = 必须拆开

**场景**：函数既要能打印给人看，又要能输出数据给机器。

**做法**：拆成 `collect*()` + `print*()`。

**判断标准**：*如果这个函数将来可能有第二种输出方式，现在就拆。* 不用等"真的需要"，因为拆的成本很低（就是把循环和打印分开），而合并的成本很高（一旦有人依赖它的打印行为，改动就会牵连一片）。

这跟后端分层是同一个思路，只是粒度更小：**service 不写 SQL、collect 不打印。**

### 技巧 4：用类型表达状态，不用全局变量

**场景**：需要让上层知道"下层已经处理过了"。

**做法**：定义哨兵类型 + `errors.As`。

```go
type PrintedError struct{ Err error }
```

**为什么不用 `var printed bool`**：

| | 全局 bool | 哨兵类型 |
|---|---|---|
| 编译期检查 | 无 | 有（类型写错编不过） |
| 作用域 | 全局，谁都能改 | 只有返回值传递 |
| 出错时 | 静默地多打一次 / 少打一次 | 要么命中要么不命中，逻辑清晰 |
| 可组合 | 不能 | 能（`Unwrap` 保留错误链） |

**通用规律**：任何"用一个变量在模块之间传状态"的设计，先想能不能换成一个类型、一个接口、一个返回值。**显式的代价是一次性多写几行；隐式的代价是将来某天排查两小时。**

### 技巧 5：交付物要能"自证"

**场景**：把东西交给别人，需要他能自己确认"拿到了正确的东西"。

**做法**：让交付物自己证明自己。

| 证明什么 | 靠什么 |
|---|---|
| 版本对不对 | `skillctl version` 自报版本 + Go 版本 + 平台 |
| 文件有没有坏 | `checksums.txt` + `shasum -c` |
| 有没有泄露本机信息 | `strings` + `grep` 搜用户名 |
| 是不是真的静态链接 | `file` 输出里的 `statically linked` |

**核心想法**：不要让对方"相信你说它是好的"，而是给他一个能自己验证的手段。这也是省时间 —— 对方不用来问你，自己就能确认。

### 技巧 6：隔离环境 + 反向证明

**场景**：测试会写用户目录（`~/.skillctl`）、测试需要验证"某个操作没有发生"。

**做法一：隔离 HOME。**

```bash
HOME=$PWD/.tmp-home ./skillctl skill install demo
```

Go 的 `os.UserHomeDir()` 读的是 `$HOME` 环境变量，所以改它就把所有写操作关进了一个临时目录。测完 `rm -rf` 就行，不污染真实环境、不影响别的测试。

这比"先备份再恢复"简单得多，也比"直接往真家目录写"安全得多。

**做法二：用"不可能成功"的凭据反向证明。**

要验证 `upgrade` 在"已是最新"时**零下载**，怎么证明？直接看输出说"已是最新"不够 —— 万一它偷偷下载了只是没提示呢？

做法是给一个必然失败的 token：

```bash
DR_ACCESS_TOKEN=invalid-token ./skillctl skill upgrade demo
```

如果它真的尝试下载，就会被鉴权拦下并报错；现在它成功返回"已是最新"，就**证明了下载这一步确实没发生**。

**这叫反证法**：不证明"A 发生了"，而是构造一个"如果 A 发生就必然暴露"的条件，然后观察它没暴露。

这个手法在测试"某操作被跳过"时特别好用 —— 因为"没发生"这件事本身很难直接观察。

### 技巧 7：静默生效的参数，必须有独立验证

`-trimpath`、`-s`、`GOMAXPROCS` 这类参数，**生效与否都能编译成功**。所以"命令没报错"完全不能说明它们起作用了。

做法：为每个静默参数准备一个独立检查。

| 参数 | 检查方式 |
|---|---|
| `-trimpath` | `strings` 搜本机路径，期望 0 |
| `-s -w` | `ls -lh` 看体积变化 |
| `CGO_ENABLED=0` | `file` 看 `statically linked` |
| `-X` 版本注入 | 跑 `--version` 看输出 |

**通用原则**：验证要针对"假设"，而不是针对"命令执行成功"。`go build` 成功只说明代码能编，不说明编出来的东西符合你的预期。

### 技巧 8：改完源码，产物要重编

这条太朴素了，但确实容易忘，而且后果是"文档和程序不一致"。

**做法**：把重编放进流程里 —— 发版永远从 `make dist` 开始，不手动拷文件。

**可以记住的检查**：交付前跑一次 `make dist && ./dist/<本机平台> --version`，版本对得上就说明产物是新的。

---

## 7. 全部验收汇总

一张表看完这次验了什么、怎么验的、结论。

| # | 验什么 | 手段 | 结果 |
|---|---|---|---|
| 1 | 每个命令的 JSON 输出能解析 | `python3 -m json.tool` | 通过 |
| 2 | 失败时 stdout 干净 | `2>/dev/null` 后看输出长度 | 0 字节 |
| 3 | 空列表是 `[]` 不是 `null` | 空账本 + `--installed --json` | `[]` |
| 4 | token 不进 JSON | `grep -c token` | 0 |
| 5 | 部分成功带回数据 | 多 target 中途失败 | `data` 有已完成项 |
| 6 | 编译无告警 | `gofmt -l` + `go vet ./...` | 全空 |
| 7 | 六平台产物格式正确 | `file` | 3 种格式都对 |
| 8 | Linux 版静态链接 | `file` 输出 | `statically linked` |
| 9 | 二进制不含本机路径 | `strings \| grep -c` | 0 |
| 10 | 校验和自洽 | `shasum -a 256 -c` | 6 个 OK |
| 11 | 发布版能跑 | 直接执行 `dist/darwin-arm64 version` | `1.0.0` |
| 12 | `--version` 生效 | `skillctl --version` | `skillctl version 1.0.0` |
| 13 | 子命令 `--version` 未被抢 | `skill install --help` | `--version string 指定版本` |
| 14 | 人类可读输出没被破坏 | 不带 `--json` 跑 list/upgrade | 表格与文案正常 |

**这 14 项全部通过，但它们有一个共同的盲区：全部在测"正常情况下 CLI 有没有做对"，没有一项在测"用户写错的时候 CLI 说了什么"。**

后者后来被外部评审打出了 4 个真问题（其中 2 个是"静默返回成功"）。复核与修复见第 9 节 —— 那一节比这一节更值得读，因为它解释了为什么"全绿"不等于"没问题"。

---

## 8. 已知边界与还没做的事

**主动没做**（不是忘了，是判断价值不够）：

| 项 | 为什么不做 |
|---|---|
| `install.sh`（装进 PATH 的脚本） | 收益是"省一步手动复制"，但引入"脚本会不会改错你的 PATH"的信任成本。先让人手放 `/usr/local/bin` 更透明。要做得先想清楚：装哪、要不要 sudo、怎么卸载 |
| 32 位 / FreeBSD 等平台 | 目标用户几乎不存在，多编一个是纯浪费 |
| `--json` 支持 `--version` | 框架层面就在业务逻辑之前返回了，强行支持要绕过 cobra 机制，不值得。文档已说明用 `skillctl version --json` |
| 二进制压缩（UPX） | 7MB 可接受，压到 3MB 换来杀毒软件误报，不划算 |

**已知但未解决的**：

1. **后端 8000 端口偶发卡死**。表现为：端口在 LISTEN，但任意请求 8 秒超时返回 `HTTP=000`。初步判断是 `--reload` 的 worker 进程死了，父进程还占着端口。这次测试时遇到两次，一次用 8001 起干净实例绕过，一次重启解决。**没找到根因**，下次再遇到应该先看 worker 子进程是否还活着。
2. **`install.sh` 缺失**，见上表。
3. **Windows 版只验证了文件格式**（`file` 显示 PE32+），没有在真 Windows 机器上跑过。理论上 Go 标准库跨平台没问题，但路径分隔符、换行符相关的细节应该实机验证一次再交付。
4. **零个 Go 测试文件**。`go test ./...` 六个包全是 `no test files`。所有结论都靠手工跑，任何人改一行都可能悄悄退化。这是目前最大的工程缺口，排在修复顺序的最后（理由见第 9.4 节）。
5. **测试数据太少导致分支没被触发**：`upgrade` 里"删除旧版多出来的文件"那一段，因为现有两对测试包的文件清单完全相同（big-demo 1.0.0/2.0.0 都是 77 个，frontend-e2e 1.0.0/1.0.1 也相同），正常路径下 `removed` 永远是 0，这段逻辑没有被真实数据验证过。

---

## 9. 事后审计：外部反馈逐条核验

这一节是文档写成**之后**补的。起因是另一个人（DeepSeek）读了代码、跑了命令，给出 11 条问题清单。逐条核验下来：**9 条完全成立，2 条方向对但表述需要修正**，另外它还漏了 3 处更严重的地方。

这一节比前面任何一节都值得细读，因为它记录的是"我以为验完了，其实没有"。

### 9.1 逐条核验

| # | 反馈 | 结论 | 证据 |
|---|---|---|---|
| 1 | `skillctl --json` 无子命令时把整屏 usage 打到 stdout，退出码 0 | **成立** | 实测 stdout 785 字节、stderr 0、`rc=0` |
| 2 | 写错命令时 `--json` 失效，输出人类可读错误 | **成立，但比它说的更糟** | 见 9.2 |
| 3 | 分页参数没有本地校验，`--limit=-1` 打 500、`--limit=0` 静默空 | **成立** | 实测 `500 Internal Server Error` / `没有找到匹配的技能` |
| 4 | `search demo` 搜不到 `big-demo-skill`，后端只 LIKE `display_name` 和 `summary` | **成立** | `search demo` → total=0；`search 大仓库` → total=1 |
| 5 | 重复安装同一版本 `download_count` 不涨 | **成立**（设计如此） | 连装 3 次，始终停在 5 |
| 6 | 零个测试文件 | **成立** | 6 个包全是 `no test files` |
| 7 | `upgrade` 的 `removed` 永远是 0，没被真实数据验证过 | **成立** | 两对测试包的文件清单 diff 全为空 |
| 8 | 坏包拒绝 / 下载断流 / Windows 实跑 三条没测 | **成立** | 确实没有 |
| 9 | `make build` 和 `make dist` 的 darwin-arm64 不是同一个二进制 | **成立，数字要修正** | 不是"差 203 字节大小"，是**大小完全相同（7460658）、203 个字节内容不同**，差异来自 `CGO_ENABLED` 1 vs 0 |
| 10 | `skill_url.go` 的 `Ddownload_url` 是死代码 | **成立** | 全仓仅定义处，零引用 |
| 11 | 工作区没有 `STATE.md`，但 `AGENTS.md` 要求先读它 | **成立** | 文件确实不存在 |

### 9.2 它说漏、也说轻了的地方

**① 第二层命令写错时不报错，静默返回成功。**

反馈只说"`--json` 失效、输出人类可读错误"。实际有两种情况，第二种更严重：

| 场景 | 修之前 | 修之后 |
|---|---|---|
| `skillctl lst --json` | rc=1，stderr 纯文本 | rc=1，stderr JSON |
| `skillctl skill lst --json` | **rc=0**，stdout 打 720 字节帮助 | rc=1，stderr JSON |

原因是 cobra 的 `legacyArgs` 只在**根命令**上做 unknown-command 检查，`skill` 有父命令就直接放行 → 因为 `skill` 不可运行 → 打印 help → 返回 nil → rc=0。对脚本来说这是"命令成功执行了，输出是一堆中文"。

**② JSON 字段命名有两套风格。**

市场数据直接透传后端（`display_name` / `download_count` / `created_at`），本机安装数据是 CLI 自己拼的（`installedAt`）。同一个 CLI 的输出，消费方要记两套规则。

**③ `--sort` 也静默回退。**

反馈提到 `--limit`，但同一个文件的 `sort` 有一样的问题：后端 `sort_map.get(sort, "s.created_at DESC")`，传 `--sort=stars` 不报错，默默按时间排序。属于同一类病：**用户的意图被吞掉，而 CLI 表现得很正常。**

### 9.3 修了什么

一共 6 处行为改动 + 1 处清理，全部在 CLI 侧，**没有碰后端**。

#### ① 无子命令 → 分清"人"和"脚本"

给 root / auth / skill 三个"只分组、不执行"的命令都挂上 `RunE`：

```go
func groupRunE(cmd *cobra.Command, args []string) error {
	if output.Enabled() {
		return fmt.Errorf("%s 后面要跟子命令，用 %s --help 查看", cmd.CommandPath(), cmd.CommandPath())
	}
	return cmd.Help()
}
```

**为什么不是简单地"一律报错"**：`skillctl` 直接敲回车，人的预期就是看帮助，这时 rc=0 是对的。只有脚本要明确"我没让你成功"。两个模式给两个答案，是这个函数存在的全部理由。

#### ② 二层命令拼错 → 加 `Args: cobra.NoArgs`

```go
var skillCmd = &cobra.Command{Use: "skill", Args: cobra.NoArgs, RunE: groupRunE}
```

`NoArgs` 在参数校验阶段就把 `lst` 判为非法：

```
unknown command "lst" for "skillctl skill"
```

关键是**参数校验发生在 flag 解析之后**，所以这时 `--json` 已经落进 flag，`JSONRequested()` 能读到 true，错误能正确输出成 JSON。

#### ③ 命令名拼错 → 从原始命令行找 `--json`

根层写错（`skillctl lst --json`）时 cobra 在 `Find()` 阶段就失败，`ParseFlags()` 没跑，flag 读不到。补一个只认 `--json` 的扫描：

```go
func jsonInArgs(args []string) bool {
	want := false
	for _, a := range args {
		switch a {
		case "--json", "--json=true":
			want = true
		case "--json=false":
			want = false
		}
	}
	return want     // 取最后一次出现，--json=false 能正确覆盖前面的 --json
}
```

`main.go` 里两个条件取或：`if JSONRequested() || jsonInArgs(os.Args[1:])`。

**为什么不去重新 ParseFlags**：会跟 cobra 内部状态打架（重复解析、重复报错）。这个函数只做一件事 —— 在 flag 系统不可用时，从字符串里捞一个布尔值。范围小到不会出错。

#### ④ 分页 / 排序 → 本地拦截

```go
const maxPageSize = 100

func checkListArgs(page, limit int, sortBy string) error {
	if page < 1 { ... }
	if limit < 1 || limit > maxPageSize { ... }
	if !validSorts[sortBy] { ... }
	return nil
}
```

`list` 和 `search` 共用（两者本来就共用 `SearchSkills`）。放在**发请求之前**，所以：

| 输入 | 修之前 | 修之后 |
|---|---|---|
| `--limit=-1` | 后端 500，`Internal Server Error` | 本地报错，一个请求都不发 |
| `--limit=0` | rc=0，"没有找到匹配的技能" | 本地报错 |
| `--limit=101` | 后端照查，可能很慢 | 本地报错（上限 100） |
| `--page=0` | 后端 offset = -10，行为未定义 | 本地报错 |
| `--sort=stars` | 静默按时间排序 | 本地报错，列出三个合法值 |

**上限 100 是新增的**，原来没有。理由是：后端对 `size` 完全不校验，`--limit=100000` 会真的去查十万行。与其等它出事，不如在 CLI 挡一下。

**为什么不在后端加校验**：也是合理的做法，但需要改 Python 代码、重新部署、还要考虑所有老客户端的兼容。CLI 侧加校验的成本是几行代码、立刻生效。两边都加最稳，但当前阶段 CLI 侧足够（后端是内网自用的）。

#### ⑤ JSON 字段统一成 snake_case

```go
installedAt → installed_at      // list.go
dirGone     → dir_gone          // uninstall.go
userId      → user_id           // login.go / version.go
loggedIn    → logged_in         // version.go
loggedOut   → logged_out        // logout.go
```

**关键判断：只改输出结构体，不改磁盘结构体。**

同样是 json tag，位置不同，含义完全不同：

| 结构体 | 序列化到哪 | 能不能改 |
|---|---|---|
| `installedItem` / `UninstallResult` / `loginResult` … | stdout 的 JSON 信封 | **能改**（这是我们新定的对外契约） |
| `Registry`（`registry.go`） | `~/.skillctl/registry.json` | **不能改**（用户机器上已有文件） |
| `Config`（`security.go`） | `~/.skillctl/config.json` | **不能改**（同上） |

`registry.json` 里的 `schemaVersion` / `installedAt` 一旦改名，旧账本读出来就是空值，`SchemaVersion` 变 0 还可能触发版本不兼容逻辑。**"输出格式"和"持久化格式"是两套契约，变更代价完全不同。**

#### ⑥ `make build` 统一 `CGO_ENABLED=0`

```make
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .
```

修完实测两个 sha256 **完全相同**：

```
2be185c251ec11fd4c5e82b4781cefa471ae8c1fc502ef60529bb36dc72c7e32  两个文件
```

这样"本地 `make build` 验过，再核对发布包 sha"才是可信的流程。

#### ⑦ 删掉死代码

`Ddownload_url` 已无引用，删除。注释里"临时保留，避免这一版编译不过"是我自己当初写的 —— **临时代码必须带清除条件**，否则它会活到下一个版本。

### 9.4 没修的

| 项 | 决定 | 理由 |
|---|---|---|
| 后端 `q` 加上 `slug` 匹配 | **不改** | 项目负责人拍板：搜索按内容（display_name / summary）语义更合理；slug 是标识符，不是描述 |
| 补 Go 单元测试 | 延后 | 工作量大；且先把行为改对，再用测试锁住，顺序反了会锁住错的行为 |
| Windows 实机验证 | 未做 | 手边没有 Windows 机器 |
| `STATE.md` | 未建 | 属于协作流程，等负责人确认 |

### 9.5 为什么之前 14 项验收全绿，却有 4 个真问题

这是本文最该记住的部分。

**因为那 14 项测的全是"我做的东西对不对"，没有一项测"用户做错了会发生什么"。**

手测有三种系统性偏差：

| 偏差 | 表现 | 这次的具体例子 |
|---|---|---|
| **只走 happy path** | 按文档的正确用法敲命令 | `list --limit 10`、`search demo` 都是对的 |
| **只验自己的假设** | 我想到了"参数校验失败时 PreRunE 不执行"，就只测了那种失败 | 却没测"参数名本身写错"，那是另一条完全不同的失败路径 |
| **路径依赖** | 我心里知道哪条路走得通，手会避开走不通的 | 从没想过敲 `skillctl --json` 不带子命令 |

外部评审之所以有效，不是因为更聪明，而是因为**他读代码时没有我的路径依赖**。他会问"如果用户这么敲会怎样"，而我问的是"我这样敲对不对"。

**具体到操作层面，能立刻用的补丁是**：

> 验证一个命令，至少准备三类输入：**合法的、非法的、空白的**。
> 每类都检查四件东西：`rc`、`stdout 内容`、`stdout 长度`、`stderr 内容`。

这次的 4 个问题，全部落在"非法输入"这一类里 —— 而这一类我一项都没准备。

### 9.6 修完的验证

| # | 场景 | rc | stdout | stderr |
|---|---|---|---|---|
| 1 | `skillctl`（无参数） | 0 | 帮助文本 | 空 |
| 2 | `skillctl --json` | 1 | **0 字节** | JSON 错误 |
| 3 | `skillctl skill --json` | 1 | 0 字节 | JSON 错误 |
| 4 | `skillctl skill lst --json` | 1 | 0 字节 | JSON 错误（修前 rc=0 + 帮助） |
| 5 | `skillctl lst --json` | 1 | 0 字节 | JSON 错误（修前纯文本） |
| 6 | `skillctl lst`（无 `--json`） | 1 | 0 字节 | `Error: unknown command "lst"` |
| 7 | `--limit=-1` | 1 | 0 字节 | `--limit 必须在 1~100 之间（收到 -1）` |
| 8 | `--limit=0` | 1 | 0 字节 | 同上 |
| 9 | `--limit=101` | 1 | 0 字节 | 同上 |
| 10 | `--page=0` | 1 | 0 字节 | `--page 必须 ≥ 1（收到 0）` |
| 11 | `--sort=stars` | 1 | 0 字节 | `--sort 只能是 newest / downloads / rating` |
| 12 | `--limit=2`（正常） | 0 | `total=3 items=2 size=2` | 空 |

回归（改字段名后原有行为不能变）：

| 场景 | 结果 |
|---|---|
| `version --json` | `logged_in` / `user_id` 生效 |
| `install --json` | `copied=77 removed=0` |
| `list --installed --json` | `installed_at` 生效 |
| `uninstall --json` | `dir_gone` 生效，`removed=77` |
| 人类模式 `list` / `list --installed` | 表格与中文对齐未受影响 |
| `make build` vs `make dist` | sha256 完全相同 |
| `shasum -a 256 -c checksums.txt` | 6 个全 OK |

### 9.7 补充：`auth register`（审计之后才发现的缺口）

审计的 11 条里没有这一条，是项目负责人自己看命令列表时发现的：**`auth` 只有 `login` 和 `logout`，没有注册。**

漏掉的原因值得记：我当时的默认假设是"用户已经在网页端注册过了"。但 CLI 是一个独立入口 —— 一个只有命令行的人在第一步就走不下去。

#### 后端接口

```
POST /api/v1/auth/register
请求 {email, name(2~20字), password(8~32位)}
响应 201 {message: "注册成功"}
```

**和 login 的三个关键差异**：

| | login | register |
|---|---|---|
| 成功状态码 | 200 | **201** |
| 是否返回 token | 是 | **否** |
| 请求必填 | email + password | email + name + password |

第一条最容易踩：如果照抄 `login` 里的 `if resp.StatusCode != http.StatusOK`，注册成功会被当成失败。

#### 三个设计决定

**① 注册成功后自动登入。**

因为 register 不返回 token，如果不自动登录，用户面对的是一句"注册成功"+ 一个仍然是未登录状态的 CLI，还得再敲一条命令。所以在 `Register()` 之后顺手调用已有的 `Login()`。

**② 自动登入失败，不算注册失败。**

账号已经在数据库里建好了。这时候返回错误会让用户以为"没注册上"，然后去重试 —— 而重试会撞上 409（邮箱已存在）。

所以：人类模式打印"注册成功"+ 明确提示请手动登录，退出码 0；JSON 模式里 `logged_in: false` 把这个中间态显式表达出来：

```json
{"ok": true, "data": {"email": "...", "user_id": 0, "message": "注册成功", "logged_in": false}}
```

**一个业务动作可能产生"部分成功"，这件事在 register 上比 upgrade 更隐蔽** —— 因为从用户视角看它只是一条命令。

**③ 本地校验长度，并把 422 转成人话。**

密码长度这类约束后端有（`schemas.py` 的 `Field(min_length=8, max_length=32)`），但后端的 422 响应长这样：

```json
{"detail":[{"loc":["body","password"],"msg":"String should have at least 8 characters","type":"string_too_short"}]}
```

直接打出来没人看得懂。所以两件事都做：

- 本地先挡一遍（`checkRegisterArgs`），错误信息是中文、带实际收到的长度
- 万一后端仍然返回 422（比如邮箱格式），用 `friendlyErrorBody` 把 `detail` 数组压成一行人话

实测效果：

```
$ skillctl auth register --email not-an-email --name 小明 --password 12345678
Error: 注册失败，状态码 422:email value is not a valid email address: An email address must have an @-sign.

$ skillctl auth register --email 已存在@example.com --name 小明 --password 12345678
Error: 注册失败，状态码 409:邮件已被注册
```

两条的错误体形状完全不同（一个是数组、一个是字符串），`friendlyErrorBody` 两种都接住了。

#### 实测 8 项

| # | 场景 | 结果 |
|---|---|---|
| 1 | 缺 `--password` | 本地报错，rc=1 |
| 2 | `--name` 只有 1 个字 | 本地报错「必须在 2~20 个字之间」 |
| 3 | `--password` 只有 3 位 | 本地报错「必须在 8~32 位之间」 |
| 4 | 正常注册 | `注册成功` → `已登入:测试用户 (user_id=16)`，token 落地 |
| 5 | 同一邮箱再注册 | 409 → `邮件已被注册` |
| 6 | 邮箱格式非法 | 422 → 可读的英文 detail |
| 7 | `--json` | `logged_in: true`，`user_id: 17` |
| 8 | 注册后用 `version` 查 | 显示新账号，说明 token 真的可用 |

（这次测试在本地库里留下了两个账号：`cli-test-<时间戳>@example.com`、`cli-json-<时间戳>@example.com`。）

#### 这一条真正的教训

**功能清单应该按"用户旅程"列，而不是按"我想到的模块"列。**

用户旅程是：

```
注册 → 登入 → 搜索 → 安装 → 升级 → 卸载
```

我实际是从第二步开始做的，因为"注册"在我脑子里属于网页端的事。**只要一开始把这条线写出来，第 1 步缺失是看一眼就能发现的。**

这和 9.5 是同一类错误：9.5 是"没测非法输入"，这里是"没列全用户路径"。两者都是**验证的覆盖面问题，不是实现能力问题**。

### 9.8 删掉密码的环境变量通路

原来的取值顺序是两条线：

```
--password  有值就用
否则       读 SKILLCTL_PASSWORD 环境变量
```

项目负责人看过之后决定**砍掉环境变量那条，只留 `--password` 明文**。理由很直接：

> 我不会为了这一点安全性，大费周章去搞这个。

这个判断我认同，而且它符合一个容易忽略的标准：**一条通路要维护的成本，不只在于它本身的几行代码**。

| 成本 | 具体是什么 |
|---|---|
| 代码 | 每个用到密码的命令多 4 行（判空 + 读 env） |
| 错误信息 | 要写成"必须提供 --password(或设置 SKILLCTL_PASSWORD)"，两种失败混在一句话里 |
| 帮助文本 | flag 说明要多半句，`--help` 变长 |
| 文档 | 环境变量表多一行，还得解释"为什么推荐用它" |
| 认知 | 用户要判断"我该用哪种"，而不是"填密码" |

收益是什么？密码不进 shell 历史。但这个收益只在"有人真的长期用这种方式"时才兑现，而实际上：

- 密码**只在 login / register 时用一次**，之后凭证是 `config.json` 里的 token
- CI 场景用的是 `DR_ACCESS_TOKEN`（token），不是密码
- 真要防历史泄露，临时变量 `PASSWORD=xxx skillctl auth login ...` 也做得到

**用一次性场景的便利，换五处长期维护成本，不划算。**

删完之后的实际行为（实测）：

| 命令 | 结果 |
|---|---|
| `SKILLCTL_PASSWORD=12345678 skillctl auth login --email a@example.com` | `Error: 必须提供 --email 和 --password`，rc=1 |
| 同上，register | `Error: 必须提供 --email、--name 和 --password` |
| `skillctl auth login --email x --password 12345678` | 请求真的发出去了，后端回 `401 {"detail":"邮件或密码错误"}` |

第一条故意把环境变量设上再跑 —— **如果还有 env 通路，它就该成功**。它报错了，才证明那条线真的删干净了（不是删了代码但还留着提示）。

这就是第 6 节"反证法"的又一次应用：验证"某条路已经不存在"，比验证"另一条路存在"更需要构造巧妙的输入。

---

## 10. 术语表

文档里出现的词，一句话解释。

| 词 | 意思 |
|---|---|
| **cobra** | Go 里最常用的命令行框架。管命令树、flag 解析、帮助信息 |
| **persistent flag** | 定义在父命令上、所有子命令都能用的 flag（如 `--json`） |
| **local flag** | 只在当前命令有效的 flag（如 `install --version` 指技能版本） |
| **横切改造** | 穿过多处代码、但不改变业务逻辑的改动（如统一输出格式） |
| **信封（envelope）** | 在数据外面包一层固定结构，用来携带元信息（成功与否、错误原因） |
| **stdout / stderr** | 进程的两个输出通道。前者走数据（可被管道消费），后者走提示（给人看） |
| **退出码** | 进程结束时返回给 shell 的数字。0 = 成功，非 0 = 失败 |
| **哨兵类型** | 用来标记某类错误的特殊类型，配合 `errors.As` 判断 |
| **静态链接** | 把依赖的库代码全部打进可执行文件，运行时不需要系统再提供 |
| **交叉编译** | 在一台机器上编译出另一个平台的可执行文件 |
| **overlay** | 安装时不删目录，直接往上面覆盖新文件 |
| **幂等** | 同一个操作做一次和做多次，结果一样 |
| **契约 / 契约测试** | 前后端/调用方与被调用方约定的接口行为，及其对应测试 |
| **反证** | 不直接证明"某事发生"，而是构造一个"若发生就必然暴露"的条件 |

---

*本文档记录的是 2026-11-22 的改造。之后如果有新的改动，建议在末尾追加一节，而不是改写原文 —— 保留决策当时的样子，比写一份"永远正确"的文档更有价值。*
