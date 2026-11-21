// 命令树的根：main.go 只调 Execute()，所有子命令在这里挂上去。
//
// 全局 flag 只在 root 上定义一次（PersistentFlags），子命令自动继承。
// SilenceUsage / SilenceErrors：出错时只打一行 Error，不喷一整屏 usage。
package main

import (
	constantquantity "go_mod/constant_quantity"
	"go_mod/oneself"
	"go_mod/output"
	"go_mod/security"
	"go_mod/skills"

	"github.com/spf13/cobra"
)

var (
	flagToken  string
	flagServer string
	flagJSON   bool
)

var rootCmd = &cobra.Command{
	Use:           "skillctl",
	Short:         "Skill Market 命令行工具:搜索、安装、管理技能",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// flag 值转交给下层包，业务代码就不用再关心 cobra
		security.FlagToken = flagToken
		constantquantity.FlagServer = flagServer
		output.SetJSON(flagJSON)
		return nil
	},
}

// auth 层：账号相关。单独一层是为了跟 skill 区分，避免 login/install 混在一起。
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "账号:登入、登出",
}

// skill 层：技能相关。子命令在 pkg 里各自定义，这里只做挂载。
var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "技能:搜索、安装、卸载、升级",
}

func init() {
	// 支持 skillctl --version（一行版本号）。详细信息仍然用 skillctl version。
	// 提前注册这个 flag，cobra 就不会用它的英文默认说明覆盖掉。
	rootCmd.Version = oneself.Version
	rootCmd.Flags().Bool("version", false, "显示版本号(详细环境信息用 skillctl version)")

	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "鉴权 token(也可用 DR_ACCESS_TOKEN 环境变量)")
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "", "后端地址(默认 "+constantquantity.Default_url+")")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "输出 JSON（给脚本 / agent 消费）")

	authCmd.AddCommand(oneself.LoginCmd, oneself.LogoutCmd)
	skillCmd.AddCommand(skills.ListCmd, skills.SearchCmd, skills.InstallCmd, skills.UninstallCmd, skills.UpgradeCmd)

	rootCmd.AddCommand(authCmd, skillCmd, oneself.VersionCmd)
}

// Execute 是 main.go 唯一要调的东西。
func Execute() error {
	return rootCmd.Execute()
}

// JSONRequested 直接读 flag 本身，不依赖 PersistentPreRunE：
// 参数校验失败时 PreRunE 不会执行，但 main 仍然要能知道该不该输出 JSON。
func JSONRequested() bool {
	v, err := rootCmd.PersistentFlags().GetBool("json")
	return err == nil && v
}
