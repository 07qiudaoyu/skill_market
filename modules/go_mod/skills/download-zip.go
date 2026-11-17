// 通过id来找到zip是否存在
package skills

import (
	"encoding/json"
	"fmt"
	"go_mod/security"
	"io"
	"mime" //用来从header里面拿文件名
	"net/http"
	"os"
	"path/filepath" //文件路径拼接
	"strconv"       //用于强制转换的，将string变为int
	"strings"

	"github.com/jedib0t/go-pretty/v6/table" //打印附属文件json信息
)

const marker = "07qiudaoyu_SkillMarket.txt"

type Record struct {
	Slug    string `json:"slug"`
	Version string `json:"version"`
	Servers string `json:"servers"`
	Uuid    string `json:"uuid"`
	Source  string `json:"source"`
}

func DownloadZip(ak []string, dir string, Als bool, Ols bool) error {
	ck := false
	if dir != "" {
		ck = true
	}

	//=====================================

	id, err := strconv.Atoi(ak[0])
	if err != nil {
		return fmt.Errorf("输入命令转换错误:%w", err)
	}

	//===========================================
	url := "http://127.0.0.1:8000/api/v1/skills/Download_file?id=" + strconv.Itoa(id)

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
	uuid := parmas["uuid"]
	nam := strings.TrimSuffix(filename, ".zip") //取除字符串后缀
	svs := strings.Split(nam, "_v")             //拿什么当间隔号
	slug := svs[0]
	sv := svs[1]

	filename_4 := "record.json"

	//2.开始进行下载，同时下载没成功要删除

	var filename_1 string
	var filename_2 string
	var filename_91 string
	var dir1 string

	if ck == false {
		dir1 = filepath.Join("..", "..", "download-zip", nam)
		err = os.MkdirAll(dir1, 0755)
		if err != nil {
			return fmt.Errorf("文件夹创建失败：%w", err)
		}
		filename_1 = filepath.Join(dir1, filename)
		filename_2 = filepath.Join(dir1, filename_4)
		filename_91 = filepath.Join(dir1, marker)
	} else {
		dir1 = filepath.Join(dir, nam)
		err = os.MkdirAll(dir1, 0755)
		if err != nil {
			return fmt.Errorf("文件夹创建失败：%w", err)
		}
		filename_1 = filepath.Join(dir1, filename)
		filename_2 = filepath.Join(dir1, filename_4)
		filename_91 = filepath.Join(dir1, marker)

	}

	filename_3, err := filepath.Abs(filename_1)
	if err != nil {
		fmt.Printf("绝对路径获取错误")
		return nil
	}
	data_json := Record{
		Servers: "http://127.0.0.1:8000",
		Slug:    slug,
		Version: sv,
		Uuid:    uuid,
		Source:  filename_3,
	}
	file, err := os.Create(filename_1) //建立文件

	if err != nil {
		return fmt.Errorf("主文件创建错误:%w", err)
	}
	file_2, err := os.Create(filename_2)
	if err != nil {
		return fmt.Errorf("附属文件创建错误:%w", err)
	}
	file91, err := os.Create(filename_91) //建立就行，不看内容-->改，前面是隐藏文件，但考虑到用户清理不方便，显示，要用内容
	if err != nil {
		return fmt.Errorf("标记文件创建失败:%w", err)
	}

	//=============================================
	success := false
	defer func() {
		file.Close()
		if !success {
			os.Remove(filename_1)
			os.Remove(filename_2)
			os.Remove(filename_91)
		}
	}() //函数最后结束时执行的
	n, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("下载错误: %w", err)
	}

	data, err := json.MarshalIndent(data_json, "", "    ")
	if err != nil {
		return fmt.Errorf("附属文件转换为json失败：%w", err)
	}

	err = os.WriteFile(filename_2, data, 0644)
	if err != nil {
		return fmt.Errorf("附属文件写入失败失败：%w", err)
	}
	data_biaoji := []byte("该文件用于标记文件夹，若删除，将导致其绑定的skill，在其目录下，不能进入列表")

	err = os.WriteFile(filename_91, data_biaoji, 0644)
	if err != nil {
		return fmt.Errorf("标记文件写入失败失败：%w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("主文件关闭错误: %w", err)
	}
	if err := file_2.Close(); err != nil {
		return fmt.Errorf("附属文件关闭错误: %w", err)
	}
	if err := file91.Close(); err != nil {
		return fmt.Errorf("标记文件关闭错误: %w", err)
	}

	//+++++++++++++++++++++++++++++++++++完善区域+++++++++++++++++++++++++++
	if Als != false {
		dirk := filepath.Dir(dir1)

		err = CreateTable(dirk) //------------------------
		if err != nil {
			fmt.Print("建表失败:%w", err)
		}
	}
	//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	if Ols != false {
		t := table.NewWriter()
		t.SetStyle(table.StyleBold)
		t.AppendHeader(table.Row{"SERVERS|来源", "SLUG|名字", "VERSION|版本", "UUID|uuid", "LOCATION|zip地址目录"})
		t.AppendRow(table.Row{data_json.Servers, data_json.Slug, data_json.Version, data_json.Uuid, data_json.Source})
		t.Render()              //渲染表格
		fmt.Println(t.Render()) //但我们渲染后还要输出呢
	}
	success = true
	//====================================================

	fmt.Printf("下载成功: %s (%d 字节)\n", filename, n)

	return nil

}
