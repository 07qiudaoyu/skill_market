package main

import (
	"fmt"
	"go_mod/auth"
	"os"
)

func main() {
	ak := os.Args
	if err := auth.Login(ak); err != nil {
		fmt.Println("登录错误:", err)
		return
	}
}
