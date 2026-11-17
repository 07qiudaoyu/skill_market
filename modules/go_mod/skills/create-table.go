package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
)

func CreateTable(dir1 string) error {

	entries, err := os.ReadDir(dir1)
	if err != nil {
		return err
	}

	var records []Record
	for _, entry := range entries {
		if entry.IsDir() { //看文件夹

			count1 := filepath.Join(dir1, entry.Name(), marker)
			_, err := os.Stat(count1)
			if err != nil {
				continue
			}
			count2 := filepath.Join(dir1, entry.Name(), "record.json")
			data, err := os.ReadFile(count2)
			if err != nil {
				return err
			}
			var record Record

			err = json.Unmarshal(data, &record)
			if err != nil {
				return err
			}
			records = append(records, record)
		}
	}
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.AppendHeader(table.Row{"SERVERS|来源", "SLUG|名字", "VERSION|版本", "UUID|uuid", "LOCATION|zip地址目录"})
	for _, v := range records {
		t.AppendRow(table.Row{v.Servers, v.Slug, v.Version, v.Uuid, v.Source})
	}
	t.Render()              //渲染表格
	fmt.Println(t.Render()) //但我们渲染后还要输出呢

	return nil
}
