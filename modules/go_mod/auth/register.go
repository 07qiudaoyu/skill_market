package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type register struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func Register(ak []string) error {
	if len(ak) != 5 {
		fmt.Println("格式错误：<command> <email> <name> <password>")
		return fmt.Errorf("error")
	}

	Input := register{
		Email:    ak[2],
		Name:     ak[3],
		Password: ak[4],
	}

	//将go转换为json，用于post输入
	JsonData, err := json.Marshal(Input)
	if err != nil {
		return fmt.Errorf("go数据转换json错误:%w", err)

	}
	url := "http://127.0.0.1:8000/api/v1/auth/register"
	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(JsonData), //使其具有io.reader能力
	) //进行req包装，成一个快递，能被api接收的快递
	if err != nil {
		return fmt.Errorf("构建post请求错误:%w", err)

	}
	req.Header.Set("Content-Type", "application/json") //头文件设立，表明json

	client := &http.Client{}    //怕一个人当快递员client
	resp, err := client.Do(req) //让client进行req快递的发送，给api，也就是url（快递地址），这个操作被视为resp

	if err != nil {
		return fmt.Errorf("请求失败:%w", err)

	}
	body, err := io.ReadAll(resp.Body) //看顾客的回复，也就是api回复
	if err != nil {
		return fmt.Errorf("回应失败:%w", err)

	}
	fmt.Println("状态码:", resp.StatusCode)
	fmt.Println("响应内容:", string(body))
	defer resp.Body.Close()
	return nil
}
