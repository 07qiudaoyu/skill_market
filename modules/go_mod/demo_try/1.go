package main

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

func main() {
	err := keyring.Set(
		"skill-market",
		"user-token",
		"this is token",
	)
	if err != nil {
		fmt.Println("保存失败:", err)
		return
	}
	token, err := keyring.Get("skill-market", "user-token")
	if err != nil {
		fmt.Println("读取失败:", err)
		return
	}

	fmt.Println("读取成功:", token)
}
