// 通过id来找到zip是否存在
package skills

import (
	"fmt"
	"go_mod/security"
	"io"
	"net/http"
	"os"
)

func DownloadZip(ak []string) error {
	if len(ak) != 3 {
		return fmt.Errorf("格式错误:<command> <id>")
	}
	url := "http://127.0.0.1:8000/api/v1/skills/Download_file?id=" + ak[2]
	res, err := http.NewRequest(
		http.MethodGet,
		url,
		nil, //get的接口，不需要json请求体，所以这个为nil
	)
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
	defer res.Body.Close() //关闭请求体
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取错误响应失败: %w", err)
		}
		return fmt.Errorf("请求失败，状态码 %d: %s",
			resp.StatusCode, string(body))
	}
	//文件的下载
	fileName := fmt.Sprintf("skill-%s.zip", ak[2])

	file, err := os.Create(fileName) //建立文件
	if err != nil {
		return fmt.Errorf("文件建立错误：%w", err)
	}

	defer file.Close() // 兜底

	n, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("copy to %s: %w", fileName, err)
	}
	defer resp.Body.Close()
	// 显式 Close 一次，捕获写入类错误
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", fileName, err)
	}

	fmt.Printf("下载成功，文件名为 %s（%d 字节）\n", fileName, n)

	return nil
}
