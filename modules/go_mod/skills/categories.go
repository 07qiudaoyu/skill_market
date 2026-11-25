// 分类枚举：只把后端的 /All_categories 搬过来，不做任何加工。
//
// 存在的理由：--category 要填的是 slug（web-dev），但 list 的表格里显示的是中文名
// （Web开发）。没有这个命令，用户只能靠猜 —— 而猜错不会报错，只会得到空结果。
package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	constantquantity "go_mod/constant_quantity"
	"go_mod/output"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// categoryItem 对应后端 GET /api/v1/skills/All_categories 的一项
type categoryItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type categoriesPayload struct {
	Count int            `json:"count"`
	Items []categoryItem `json:"items"`
}

var CategoriesCmd = &cobra.Command{
	Use:     "categories",
	Short:   "列出全部分类(slug 就是 --category 要填的值)",
	Example: "  skillctl skill categories\n  skillctl skill list --category web-dev",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := fetchCategories()
		if err != nil {
			return err
		}
		if output.Enabled() {
			return output.Data(categoriesPayload{Count: len(items), Items: items})
		}
		printCategories(items)
		return nil
	},
}

// fetchCategories 调 All_categories。注意这个接口返回的是裸数组，没有信封。
func fetchCategories() ([]categoryItem, error) {
	req, err := http.NewRequest(http.MethodGet, AllCategoriesURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("构建分类请求失败:%w", err)
	}

	resp, err := constantquantity.NewHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("连不上后端 %s，请确认 uvicorn 已启动:%w", constantquantity.ServerURL(), err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取分类响应失败:%w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取分类失败，状态码 %d:%s", resp.StatusCode, string(data))
	}

	// 显式 make：分类为空时也要序列化成 []，不能是 null
	items := make([]categoryItem, 0)
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("解析分类响应失败:%w", err)
	}
	return items, nil
}

func printCategories(items []categoryItem) {
	if len(items) == 0 {
		fmt.Println("后端还没有任何分类")
		return
	}
	fmt.Printf("共 %d 个分类（--category 填 SLUG 这一列）\n\n", len(items))

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"SLUG", "NAME", "ID"})
	for _, c := range items {
		t.AppendRow(table.Row{c.Slug, c.Name, c.ID})
	}
	t.Render()
}
