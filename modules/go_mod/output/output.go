// CLI 的输出层：默认人类可读，--json 时输出结构化信封。
//
// 约定（方便脚本 / agent 消费）：
//   - 成功：stdout 打 {"ok":true,"data":...}，退出码 0
//   - 失败：stderr 打 {"ok":false,"error":{"message":"..."}}，退出码 1
//   - 杂项提示（自动检测、进度）在 --json 时改走 stderr，不能污染 stdout
package output

import (
	"encoding/json"
	"fmt"
	"os"
)

var jsonMode bool

// SetJSON 由 root 的 PersistentPreRunE 调用，决定这一条命令怎么输出。
func SetJSON(on bool) { jsonMode = on }

// Enabled 当前是否 JSON 模式。
func Enabled() bool { return jsonMode }

// Say 打印人类可读的提示。--json 时改走 stderr，保持 stdout 只有 JSON。
func Say(format string, a ...any) {
	if jsonMode {
		fmt.Fprintf(os.Stderr, format+"\n", a...)
		return
	}
	fmt.Printf(format+"\n", a...)
}

type successEnvelope struct {
	OK   bool `json:"ok"`
	Data any  `json:"data"`
}

type errorBody struct {
	Message string `json:"message"`
}

type errorEnvelope struct {
	OK    bool      `json:"ok"`
	Error errorBody `json:"error"`
	Data  any       `json:"data,omitempty"`
}

// PrintedError 表示错误信息已经输出过了，上层只需要设置退出码，别再打一遍。
type PrintedError struct{ Err error }

func (e PrintedError) Error() string { return e.Err.Error() }
func (e PrintedError) Unwrap() error { return e.Err }

// Data 输出成功结果到 stdout。
func Data(v any) error {
	return encode(os.Stdout, successEnvelope{OK: true, Data: v})
}

// Fail 输出失败结果到 stderr。
func Fail(err error) {
	_ = encode(os.Stderr, errorEnvelope{OK: false, Error: errorBody{Message: err.Error()}})
}

// FailWithData 输出失败结果，同时带上已经完成的部分（比如卸载了 2 个 target 后第 3 个失败）。
func FailWithData(err error, data any) error {
	_ = encode(os.Stderr, errorEnvelope{OK: false, Error: errorBody{Message: err.Error()}, Data: data})
	return PrintedError{Err: err}
}

func encode(f *os.File, v any) error {
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false) // 中文和 < > 原样输出，别变成 \u003c
	return enc.Encode(v)
}
