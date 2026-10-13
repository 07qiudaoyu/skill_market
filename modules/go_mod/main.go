package main

import (
	"fmt"
	"go_mod/auth"
	"go_mod/skills"
	"os"
)

func main() {
	ak := os.Args
	// --------------------------------------------------------------------
	if ak[1] == "login" {
		if err := auth.Login(ak); err != nil {
			fmt.Println("登录错误:", err)
			return
		}
	}
	// --------------------------------------------------------------------
	if ak[1] == "register" {
		if err := auth.Register(ak); err != nil {
			fmt.Println("注册错误:", err)
			return
		}
	}
	//-----------------------------------------------------------------------
	if ak[1] == "find" {
		if err := skills.Find(ak); err != nil {
			fmt.Println("skill查找错误:", err)
			return
		}
	}
	//---------------------------------------------------------------

}
