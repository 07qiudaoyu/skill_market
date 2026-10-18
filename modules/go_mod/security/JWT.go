package security

import (
	"fmt"
	"net/http"

	"github.com/zalando/go-keyring"
)

func JwtHeader(res *http.Request) error {

	result1, err := keyring.Get(
		"skill_market/login/token",
		"user_email",
	)
	if err != nil {
		return fmt.Errorf("用户邮件读取错误:%w", err)

	}
	result2, err := keyring.Get(
		"skill_market/login/token",
		result1,
	)
	if err != nil {
		return fmt.Errorf("Token读取错误:%w", err)

	}
	res.Header.Set("Authorization", "Bearer "+result2)
	return nil
}
