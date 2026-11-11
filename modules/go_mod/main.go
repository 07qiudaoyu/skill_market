package main

import (
	"fmt"
	"go_mod/auth"
	"go_mod/skills"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "greet",
		Short: "提供注册，登入，下载及搜索功能",
	}
	lg := &cobra.Command{
		Use:   "lg <邮件email> <密码password>",
		Short: "用户登入auth/login",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.Login(args); err != nil {
				fmt.Println("登录错误:", err)
				return err
			}
			fmt.Printf("登入成功")
			return nil
		},
	}
	rg := &cobra.Command{
		Use:   "rg <邮件email> <名字name> <密码password>",
		Short: "用户注册auth/register",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.Register(args); err != nil {
				fmt.Println("注册错误:", err)
				return err
			}
			return nil
		},
	}
	dl := &cobra.Command{
		Use:   "dl <技能id>",
		Short: "用户下载zip技能skills/download-zip",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := skills.DownloadZip(args); err != nil {
				fmt.Println("skill下载失败:", err)
				return err
			}
			fmt.Print("下载成功")
			return nil
		},
	}
	fv := &cobra.Command{
		Use:   "fv <uuid>",
		Short: "用户通过客户端拿到uuid进行版本搜索skills/find-versions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := skills.FindVersions(args); err != nil {
				fmt.Println("skill查找错误:", err)
				return err
			}
			fmt.Print("搜索成功")
			return nil
		},
	}
	root.AddCommand(lg, rg, dl, fv)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}

}
