package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "greet",
		Short: "跟人打招呼的小工具",
	}
	hello := &cobra.Command{
		Use:   "hello <name>",
		Short: "跟一个人问好",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			loud, err := cmd.Flags().GetBool("loud")
			if err != nil {
				return err
			}
			if loud {
				fmt.Printf("你好哦哦哦！，%s\n", args[0])
			} else {

				fmt.Printf("你好，%s\n", args[0])
			}

			return nil
		},
	}
	hello.Flags().Bool("loud", false, "喊出来")
	bye := &cobra.Command{
		Use:   "bye <名字>",
		Short: "跟一个人道别",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("再见，%s\n", args[0])
			return nil
		},
	}
	root.AddCommand(hello, bye)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
