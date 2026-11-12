package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
)

// 返回的数据体versions带有的数据组
type Version struct {
	ID                 int    `json:"id"`                 // "id": 0,
	SKILL_NAME         string `json:"skill_name"`         // "skill_name": "string",
	VERSION            string `json:"version"`            // "version": "string",
	SIZE_BYTES         int    `json:"size_bytes"`         //"size_bytes": 0,
	EXTRACT_FILE_BYTES int    `json:"extract_file_bytes"` //"extract_file_bytes": 0,
	FILE_NAME          string `json:"file_name"`          //"file_name": "string",
	CREATED_AT         string `json:"created_at"`         // "created_at": "2026-10-08T17:19:08.624Z",
	FOUNDER_NAME       string `json:"founder_name"`       //"founder_name": "string"
}
type APIReponse struct {
	Message  string    `json:"message"`
	Versions []Version `json:"versions"`
} //这个就是我们要的数据
func FindVersions(ak []string) error {

	uuid := ak[0]

	url := "http://127.0.0.1:8000/api/v1/skills/find_versions?public_id=" + uuid

	resp, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("url请求失败:%w", err)

	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("请求失败：", resp.StatusCode)
		return fmt.Errorf("error")

	}
	//读取数据流
	resBOBY, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("数据读取错误:%w", err)

	}
	var result APIReponse                  //动态吗，在我的印象里面var就是一个动态，也就是说result是APIResponse类型
	err = json.Unmarshal(resBOBY, &result) //这步是啥意思？将resBOBY的数据加载到result，这里的&就像c语言的scanf
	if err != nil {
		return fmt.Errorf("解析JSON失败:%w", err)

	} //也就是上步有错的时候，就会有错误爆出
	fmt.Println("API 状态码:", resp.StatusCode)
	fmt.Println("Message:", result.Message)
	fmt.Printf("找到了 %d 个版本:\n", len(result.Versions))
	//for _, v := range result.Versions {
	//fmt.Printf("id:%d|\nskill_name:%s|\nversion:%s|\nszie_bytes:%d|\nextract_file_bytes:%d|\nfile_name:%s|\ncreated_at:%s|\nfounder_name:%s|\n",
	//	v.ID, v.SKILL_NAME, v.VERSION, v.SIZE_BYTES, v.EXTRACT_FILE_BYTES, v.FILE_NAME, v.CREATED_AT, v.FOUNDER_NAME)
	//}
	//进行修改输出格式
	//fmt.Printf("%5s %-10s %-20s %-8s %10s %10s\n", "ID", "FOUNDER_NAME", "SKILL_NAME", "VERSION", "SIZE_BYTES", "EXTRACT_FILE_BYTES")
	//for _, v := range result.Versions {
	//fmt.Printf("%5d %-10s %-15s %-8s %10d %10d\n", v.ID, v.FOUNDER_NAME, v.SKILL_NAME, v.VERSION, v.SIZE_BYTES, v.EXTRACT_FILE_BYTES)
	//}中文占字符不一，有点难受，直接去找库
	t := table.NewWriter()      //建立空表
	t.SetStyle(table.StyleBold) //设定表格类型，这个事ASCII 风格，也就想mysql的一样
	t.AppendHeader(table.Row{
		"ID", "FOUNDER_NAME", "SKILL_NAME", "VERSION", "SIZE_BYTES", "EXTRACT_FILE_BYTES"})
	// 表写完了，就将数据写进表里面
	for _, v := range result.Versions {
		t.AppendRow(table.Row{v.ID, v.FOUNDER_NAME, v.SKILL_NAME, v.VERSION, v.SIZE_BYTES, v.EXTRACT_FILE_BYTES})
	}

	t.Render()              //渲染表格
	fmt.Println(t.Render()) //但我们渲染后还要输出呢
	defer resp.Body.Close()
	return nil
}
