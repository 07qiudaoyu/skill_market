// 版本信息：CLI 版本、编译进来的 Go 环境、当前后端地址、登入状态。
// 排查"在别人机器上跑不起来"时，让对方把这条命令的输出贴过来就够了。
package oneself

import (
	"fmt"
	"runtime"

	constantquantity "go_mod/constant_quantity"
	"go_mod/output"
	"go_mod/security"

	"github.com/spf13/cobra"
)

// Version 编译时注入：go build -ldflags "-X go_mod/oneself.Version=1.0.0"
var Version = "dev"

// versionInfo 给 --json 用。注意不含 token，只报告有没有。
type versionInfo struct {
	Version  string `json:"version"`
	Go       string `json:"go"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Server   string `json:"server"`
	LoggedIn bool   `json:"logged_in"`
	Email    string `json:"email,omitempty"`
	UserID   int    `json:"user_id,omitempty"`
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示 CLI 版本、运行环境与登入状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := security.LoadConfig()
		if err != nil {
			return err
		}
		info := versionInfo{
			Version:  Version,
			Go:       runtime.Version(),
			OS:       runtime.GOOS,
			Arch:     runtime.GOARCH,
			Server:   constantquantity.ServerURL(),
			LoggedIn: cfg.Token != "",
			Email:    cfg.Email,
			UserID:   cfg.UserID,
		}
		if output.Enabled() {
			return output.Data(info)
		}

		fmt.Printf("skillctl    %s\n", info.Version)
		fmt.Printf("go          %s %s/%s\n", info.Go, info.OS, info.Arch)
		fmt.Printf("server      %s\n", info.Server)
		if info.Email == "" {
			fmt.Println("login       未登入")
		} else if info.LoggedIn {
			fmt.Printf("login       %s (user_id=%d)\n", info.Email, info.UserID)
		} else {
			fmt.Printf("login       %s (token 已清除)\n", info.Email)
		}
		return nil
	},
}
