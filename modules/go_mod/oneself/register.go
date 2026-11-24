// 注册：新账号第一次用 CLI 的入口。
//
// 后端 POST /api/v1/auth/register 只返回一句 message，不返回 token，
// 所以注册成功后顺手调一次 Login 把 token 落地 —— 否则用户还得再敲一条命令，
// 而他根本不知道自己现在算不算"登上了"。
package oneself

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	constantquantity "go_mod/constant_quantity"
	"go_mod/output"

	"github.com/spf13/cobra"
)

// 与后端 modules/auth/schemas.py 的 RegisterRequest 保持一致。
// 本地挡一遍是为了错误信息可读：后端的 422 是一坨 pydantic JSON，用户看不懂。
const (
	minNameLen     = 2
	maxNameLen     = 20
	minPasswordLen = 8
	maxPasswordLen = 32
)

type registerRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type registerResponse struct {
	Message string `json:"message"`
}

// registerResult 给 --json 用。同样不含 token。
// LoggedIn 表示"注册后自动登入"有没有成功 —— 注册成功但登入失败时两者会不一致。
type registerResult struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	UserID   int    `json:"user_id,omitempty"`
	Message  string `json:"message"`
	LoggedIn bool   `json:"logged_in"`
}

var RegisterCmd = &cobra.Command{
	Use:     "register",
	Short:   "注册新账号并自动登入",
	Example: "  skillctl auth register --email me@example.com --name 小明 --password 12345678",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		name, _ := cmd.Flags().GetString("name")
		password, _ := cmd.Flags().GetString("password")
		if email == "" || name == "" || password == "" {
			return fmt.Errorf("必须提供 --email、--name 和 --password")
		}
		if err := checkRegisterArgs(name, password); err != nil {
			return err
		}

		msg, err := Register(email, name, password)
		if err != nil {
			return err
		}

		res := registerResult{Email: email, Name: name, Message: msg}

		// 注册接口不给 token，顺手登入一次。
		// 这里刻意不让登录失败变成命令失败：账号已经建好了，报错会误导用户以为没注册成功。
		login, loginErr := Login(email, password)
		if loginErr == nil {
			res.UserID = login.UserID
			res.Name = login.Name
			res.LoggedIn = true
		}

		if output.Enabled() {
			return output.Data(res)
		}

		// 后端的 message 本身就是"注册成功"，直接打，不要加前缀
		fmt.Println(msg)
		if loginErr != nil {
			fmt.Printf("但自动登入失败:%v\n", loginErr)
			fmt.Printf("账号已经建好了，请手动登入:skillctl auth login --email %s\n", email)
			return nil
		}
		fmt.Printf("已登入:%s (user_id=%d)\n", res.Name, res.UserID)
		fmt.Println("token 已保存到 ~/.skillctl/config.json")
		return nil
	},
}

// checkRegisterArgs 本地校验长度。只校验"后端一定会拒绝"的那几条，不做额外发明。
func checkRegisterArgs(name, password string) error {
	if n := len([]rune(name)); n < minNameLen || n > maxNameLen {
		return fmt.Errorf("--name 长度必须在 %d~%d 个字之间（收到 %d 个）", minNameLen, maxNameLen, n)
	}
	if n := len(password); n < minPasswordLen || n > maxPasswordLen {
		return fmt.Errorf("--password 长度必须在 %d~%d 位之间（收到 %d 位）", minPasswordLen, maxPasswordLen, n)
	}
	return nil
}

// Register 调 POST /api/v1/auth/register，成功返回后端给的提示语。
func Register(email, name, password string) (string, error) {
	body, err := json.Marshal(registerRequest{Email: email, Name: name, Password: password})
	if err != nil {
		return "", fmt.Errorf("请求体序列化失败:%w", err)
	}

	url := constantquantity.ServerURL() + "/api/v1/auth/register"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("构建注册请求失败:%w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := constantquantity.NewHTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("连不上后端 %s，请确认 uvicorn 已启动:%w", constantquantity.ServerURL(), err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取注册响应失败:%w", err)
	}
	// 注意是 201 不是 200：用 200 判断会把这个接口的成功当成失败
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("注册失败，状态码 %d:%s", resp.StatusCode, friendlyErrorBody(data))
	}

	var out registerResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("解析注册响应失败:%w", err)
	}
	if out.Message == "" {
		out.Message = "注册成功"
	}
	return out.Message, nil
}

// friendlyErrorBody 把 FastAPI 的错误体转成一行人话。
//
// 后端有两种长相：
//
//	{"detail":[{"loc":["body","password"],"msg":"String should have at least 8 characters"}]}
//	{"detail":"邮箱已被注册"}
//
// 直接 string(data) 打出来，用户只会看到一堆 JSON。
func friendlyErrorBody(data []byte) string {
	var listErr struct {
		Detail []struct {
			Loc []string `json:"loc"`
			Msg string   `json:"msg"`
		} `json:"detail"`
	}
	if err := json.Unmarshal(data, &listErr); err == nil && len(listErr.Detail) > 0 {
		parts := make([]string, 0, len(listErr.Detail))
		for _, d := range listErr.Detail {
			field := ""
			if len(d.Loc) > 0 {
				field = d.Loc[len(d.Loc)-1] // 取 "body"/"password" 里最后一段
			}
			parts = append(parts, strings.TrimSpace(field+" "+d.Msg))
		}
		return strings.Join(parts, ";")
	}

	var strErr struct {
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(data, &strErr); err == nil && strErr.Detail != "" {
		return strErr.Detail
	}
	return strings.TrimSpace(string(data))
}

func init() {
	RegisterCmd.Flags().String("email", "", "注册邮箱")
	RegisterCmd.Flags().String("name", "", "显示名称(2~20 个字)")
	RegisterCmd.Flags().String("password", "", "密码(8~32 位)")
}
