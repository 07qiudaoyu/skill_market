package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	if len(ak) != 3 {
		return fmt.Errorf("格式错误:<command> <uuid>")
	}
	uuid := ak[2]

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
	for _, v := range result.Versions {
		fmt.Printf("id:%d|\nskill_name:%s|\nversion:%s|\nszie_bytes:%d|\nextract_file_bytes:%d|\nfile_name:%s|\ncreated_at:%s|\nfounder_name:%s|\n",
			v.ID, v.SKILL_NAME, v.VERSION, v.SIZE_BYTES, v.EXTRACT_FILE_BYTES, v.FILE_NAME, v.CREATED_AT, v.FOUNDER_NAME)
	}
	defer resp.Body.Close()
	return nil
}
