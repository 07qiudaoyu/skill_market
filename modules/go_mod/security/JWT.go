package security

import (
	"fmt"
	"net/http"

	"github.com/zalando/go-keyring"
)

func JwtHeader(res *http.Request) {

	result1, err := keyring.Get(
		"skill_market/login/token",
		"user_email",
	)
	if err != nil {
		fmt.Println("用户邮件读取错误:", err)
		return
	}
	result2, err := keyring.Get(
		"skill_market/login/token",
		result1,
	)
	if err != nil {
		fmt.Println("Token读取错误:", err)
		return
	}
	res.Header.Set("Authorization", "Bearer "+result2)

}
