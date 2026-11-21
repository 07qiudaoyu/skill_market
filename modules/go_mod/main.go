// 总开关：不写任何业务，只把控制权交给 root.Execute()。
package main

import (
	"errors"
	"fmt"
	"os"

	"go_mod/output"
)

func main() {
	if err := Execute(); err != nil {
		// 子命令可能已经自己输出过结构化错误，这里只负责退出码
		var printed output.PrintedError
		if errors.As(err, &printed) {
			os.Exit(1)
		}
		// 错误统一在这里打一次，不打 usage、不打堆栈
		if JSONRequested() {
			output.Fail(err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
