// 登出：只清本地 token，不动其他配置（邮箱留着，方便用户知道上次登的是谁）。
package oneself

import (
	"go_mod/output"
	"go_mod/security"

	"github.com/spf13/cobra"
)

// logoutResult 给 --json 用。
type logoutResult struct {
	LoggedOut bool   `json:"logged_out"`
	Email     string `json:"email,omitempty"`
	Message   string `json:"message"`
}

var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "清除本地保存的 token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := security.LoadConfig()
		if err != nil {
			return err
		}
		if cfg.Token == "" {
			if output.Enabled() {
				return output.Data(logoutResult{LoggedOut: false, Email: cfg.Email, Message: "当前没有保存 token"})
			}
			output.Say("当前没有保存 token，无需登出")
			return nil
		}
		cfg.Token = ""
		if err := security.SaveConfig(cfg); err != nil {
			return err
		}
		if output.Enabled() {
			return output.Data(logoutResult{LoggedOut: true, Email: cfg.Email, Message: "已清除本地 token"})
		}
		output.Say("已清除本地 token")
		return nil
	},
}
