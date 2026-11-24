// 登入：把邮箱密码交给后端换 token，成功后写入 ~/.skillctl/config.json。
// 密码只用于这一次请求：不落盘、不进 config.json。
// 写进命令行会留在 shell 历史里，这是明知的取舍 —— 换少一条取值分支。
package oneself

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	constantquantity "go_mod/constant_quantity"
	"go_mod/output"
	"go_mod/security"

	"github.com/spf13/cobra"
)

// loginRequest 对应后端 POST /api/v1/auth/login 的请求体
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse 对应后端的 UserResponse 模型
type loginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
	Name    string `json:"name"`
}

// loginResult 给 --json 用。刻意不含 token：token 只落盘，不出现在任何输出里。
type loginResult struct {
	UserID  int    `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// LoginCmd 命令层：只负责取参数和打印，真正的请求在 Login 里。
var LoginCmd = &cobra.Command{
	Use:     "login",
	Short:   "登入并把 token 保存到 ~/.skillctl/config.json",
	Example: "  skillctl auth login --email me@example.com --password 12345678",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		if email == "" || password == "" {
			return fmt.Errorf("必须提供 --email 和 --password")
		}

		res, err := Login(email, password)
		if err != nil {
			return err
		}
		if output.Enabled() {
			return output.Data(loginResult{UserID: res.UserID, Name: res.Name, Email: email, Message: res.Message})
		}
		fmt.Printf("登入成功:%s\n", res.Message)
		fmt.Printf("用户:%s (user_id=%d)\n", res.Name, res.UserID)
		fmt.Println("token 已保存到 ~/.skillctl/config.json")
		return nil
	},
}

// Login 调用后端登入接口，成功后落地 token，返回响应体供命令层打印。
func Login(email, password string) (*loginResponse, error) {
	body, err := json.Marshal(loginRequest{Email: email, Password: password})
	if err != nil {
		return nil, fmt.Errorf("请求体序列化失败:%w", err)
	}

	url := constantquantity.ServerURL() + "/api/v1/auth/login"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("构建登入请求失败:%w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := constantquantity.NewHTTPClient().Do(req)
	if err != nil {
		// 连接层面的错误(超时/拒绝)单独提示，不然用户看到一堆 dial tcp 不知道怎么办
		return nil, fmt.Errorf("连不上后端 %s，请确认 uvicorn 已启动:%w", constantquantity.ServerURL(), err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取登入响应失败:%w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("登入失败，状态码 %d: %s", resp.StatusCode, string(data))
	}

	var out loginResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("解析登入响应失败:%w", err)
	}
	if out.Token == "" {
		return nil, fmt.Errorf("后端没有返回 token，请检查帐号状态")
	}

	// 保留已有的其他配置（比如 defaultTarget），只覆盖登入相关字段
	cfg, err := security.LoadConfig()
	if err != nil {
		return nil, err
	}
	cfg.Token = out.Token
	cfg.Email = email
	cfg.UserID = out.UserID
	cfg.Name = out.Name
	if err := security.SaveConfig(cfg); err != nil {
		return nil, err
	}
	return &out, nil
}

func init() {
	LoginCmd.Flags().String("email", "", "登入邮箱")
	LoginCmd.Flags().String("password", "", "登入密码")
}
