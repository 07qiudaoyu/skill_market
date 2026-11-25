// 命令树的根：main.go 只调 Execute()，所有子命令在这里挂上去。
//
// 全局 flag 只在 root 上定义一次（PersistentFlags），子命令自动继承。
// SilenceUsage / SilenceErrors：出错时只打一行 Error，不喷一整屏 usage。
package main

import (
	"fmt"

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

// groupRunE 给"只做分组、本身不执行动作"的命令（root / auth / skill）用。
//
// 不设 RunE 时 cobra 会打印整屏 usage 并以退出码 0 结束，脚本会把帮助文本当成成功返回。
// 所以这里分开：人类模式照常打帮助（rc=0 是合理的），JSON 模式返回错误（rc=1 + 结构化信封）。
func groupRunE(cmd *cobra.Command, args []string) error {
	if output.Enabled() {
		return fmt.Errorf("%s 后面要跟子命令，用 %s --help 查看", cmd.CommandPath(), cmd.CommandPath())
	}
	return cmd.Help()
}

var rootCmd = &cobra.Command{
	Use:           "skillctl",
	Short:         "Skill Market 命令行工具:搜索、安装、管理技能",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          groupRunE,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// flag 值转交给下层包，业务代码就不用再关心 cobra
		security.FlagToken = flagToken
		constantquantity.FlagServer = flagServer
		output.SetJSON(flagJSON)
		return nil
	},
}

// auth 层：账号相关。单独一层是为了跟 skill 区分，避免 login/install 混在一起。
// Args: NoArgs 让 "skillctl auth logn" 这种拼错变成明确报错，
// 而不是被 cobra 当成多余参数忽略、再打一屏帮助然后返回 0。
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "账号:登入、登出",
	Args:  cobra.NoArgs,
	RunE:  groupRunE,
}

// skill 层：技能相关。子命令在 pkg 里各自定义，这里只做挂载。
var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "技能:搜索、安装、卸载、升级",
	Args:  cobra.NoArgs,
	RunE:  groupRunE,
}

func init() {
	// 支持 skillctl --version（一行版本号）。详细信息仍然用 skillctl version。
	// 提前注册这个 flag，cobra 就不会用它的英文默认说明覆盖掉。
	rootCmd.Version = oneself.Version
	rootCmd.Flags().Bool("version", false, "显示版本号(详细环境信息用 skillctl version)")

	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "鉴权 token(也可用 DR_ACCESS_TOKEN 环境变量)")
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "", "后端地址(默认 "+constantquantity.Default_url+")")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "输出 JSON（给脚本 / agent 消费）")

	authCmd.AddCommand(oneself.RegisterCmd, oneself.LoginCmd, oneself.LogoutCmd)
	skillCmd.AddCommand(skills.ListCmd, skills.SearchCmd, skills.CategoriesCmd,
		skills.InstallCmd, skills.UninstallCmd, skills.UpgradeCmd)

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

// jsonInArgs 从原始命令行里找 --json，兜住"命令名写错"的情况。
//
// 为什么需要它：命令名拼错时 cobra 在 Find() 阶段就返回错误，ParseFlags() 根本没执行，
// JSONRequested() 只能读到默认值 false —— 用户明确要了 JSON，却收到一行中文错误。
// 这里不解析所有 flag，只认 --json 三个写法，取最后一次出现。
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
	return want
}
