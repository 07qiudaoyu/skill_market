package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zalando/go-keyring"
)

type login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
} //发送的go数据类型，后面要转换为json
type Login_Response struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	User_id int    `json:"user_id"`
	Name    string `json:"name"`
}

func Login(ak []string) error {

	if len(ak) != 4 {
		return fmt.Errorf("格式错误:<command> <email> <password>")

	}

	Input := login{
		Email:    ak[2],
		Password: ak[3],
	}

	fmt.Println("登入中：......")
	//开始将我们的Input转为json
	JsonData, err := json.Marshal(Input)
	if err != nil {
		return fmt.Errorf("go转为json错误:%w", err)

	}
	//搭建post请求,也就是包装快递
	url := "http://127.0.0.1:8000/api/v1/auth/login"
	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(JsonData), //能送
	)
	if err != nil {
		return fmt.Errorf("构建post请求失败:%w", err)

	}
	//头文件表明
	req.Header.Set("Content-Type", "application/json")

	//发派快递员
	client := &http.Client{}
	//发快递
	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("post请求体发送错误:%w", err)

	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("回复请求错误:%w", err)

	}

	fmt.Println("状态码:", resp.StatusCode)
	if resp.StatusCode != 200 {
		fmt.Println("error:", string(body))
		return fmt.Errorf("登入失败") //不知道这错没错

	}
	fmt.Println("登入成功:", string(body))
	// 单独拿出我们要的token，将json在转换为go
	var respData Login_Response
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return fmt.Errorf("token提取错误:%w", err)

	}
	err = keyring.Set(
		"skill_market/login/token",
		"user_email",
		ak[2],
	)
	if err != nil {
		return fmt.Errorf("token查询保存错误:%w", err)

	}
	err = keyring.Set(
		"skill_market/login/token",
		ak[2], //用户邮件
		respData.Token,
	)
	if err != nil {
		return fmt.Errorf("token保存错误:%w", err)

	}
	fmt.Println("登录凭证已安全保存")
	return nil
}
