// 通过id来找到zip是否存在
package skills

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_mod/security"
	"io"
	"mime" //用来从header里面拿文件名
	"net/http"
	"os"
	"strconv" //用于强制转换的，将string变为int
)

type updata_download_request struct {
	ID int `json:"skill_versions_id"`
}

func DownloadZip(ak []string) error {
	if len(ak) != 3 {
		return fmt.Errorf("格式错误:<command> <id>")
	}
	//=====================================
	id, err := strconv.Atoi(ak[2])
	if err != nil {
		return fmt.Errorf("输入命令转换错误:%w", err)
	}
	Input := updata_download_request{
		ID: id,
	}
	JsonData, err := json.Marshal(Input)
	if err != nil {
		return fmt.Errorf("go数据转换json错误:%w", err)

	}
	//===========================================
	url := "http://127.0.0.1:8000/api/v1/skills/Download_file?id=" + ak[2]
	res, err := http.NewRequest(
		http.MethodGet,
		url,
		nil, //get的接口，不需要json请求体，所以这个为nil
	) //没有请求体,也就不用关闭body
	if err != nil {
		return fmt.Errorf("构建get请求失败:%w", err)

	}
	err = security.JwtHeader(res) //加入jwt进我们的header里面
	if err != nil {
		return fmt.Errorf("jwt加入失败，请重新登入:%w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(res)
	if err != nil {
		return fmt.Errorf("请求发送错误:%w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取错误响应失败: %w", err)
		}
		return fmt.Errorf("请求失败，状态码 %d: %s",
			resp.StatusCode, string(body))
	}
	//文件的下载,也就是当状态为ok时，进行
	//1.首先，我们先去拿下文件名，在我们的后端接口处
	_, parmas, err := mime.ParseMediaType(
		resp.Header.Get("Content-Disposition"),
	)
	if err != nil {
		return fmt.Errorf("文件名读取错误：%w", err)
	}
	filename := parmas["filename"]
	//2.开始进行下载，同时下载没成功要删除
	file, err := os.Create(filename) //建立文件
	if err != nil {
		return fmt.Errorf("文件创建错误:%w", err)
	}
	//=============================================
	success := false
	defer func() {
		file.Close()
		if !success {
			os.Remove(filename)
		}
	}() //函数最后结束时执行的
	n, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("下载错误: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("文件关闭错误: %w", err)
	}

	success = true
	//====================================================
	url2 := "http://127.0.0.1:8000/api/v1/skills/Updata_Download"

	res1, err := http.NewRequest(
		http.MethodPost,
		url2,
		bytes.NewBuffer(JsonData),
	)
	if err != nil {
		return fmt.Errorf("更新数据库URL请求失败:%w", err)
	}
	err = security.JwtHeader(res1) //加入jwt进我们的header里面

	if err != nil {
		return fmt.Errorf("jwt加入失败，请重新登入:%w", err)
	}
	res1.Header.Set("Content-Type", "application/json")

	resp1, err := client.Do(res1)
	if err != nil {
		return fmt.Errorf("zip下载成功,但更新下载数据库请求失败:%w", err)
	}
	defer resp1.Body.Close()
	body, err := io.ReadAll(resp1.Body) //看顾客的回复，也就是api回复
	if err != nil {
		return fmt.Errorf("zip下载成功,但数据库回应失败:%w", err)

	}
	fmt.Println("状态码:", resp1.StatusCode)
	fmt.Println("响应内容:", string(body))

	if resp1.StatusCode != 200 {
		fmt.Println("响应内容:", string(body))

		return fmt.Errorf("下载zip成功，但数据库更新失败")

	}
	fmt.Printf("下载成功: %s (%d 字节)\n", filename, n)
	return nil

}
