package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
)

type SearchResponse1 struct {
	Public_id      string   `json:"public_id"`
	Slug           string   `json:"slug"`
	Display_name   string   `json:"Display_name"`
	Summary        string   `json:"summary"`
	Category_name  string   `json:"category_name"`
	Tags           []string `json:"tags"`
	Download_count int      `json:"download_count"`
	Rating_avg     float32  `json:"rating_avg"`
	Created_at     string   `json:"created_at"`
}
type SearchResponse2 struct {
	Message string            `json:"message"`
	Total   int               `json:"total"`
	Items   []SearchResponse1 `json:"items"`
	Page    int               `json:"page"`
	Size    int               `json:"size"`
}

func SearchSkill(ak []string) error {
	//给几个值？
	//就给三个值，一个大致搜索q，排序，页数，这类的，其他tags及category不给
	q := ak[0]
	sort := ak[1]
	page := ak[2]
	size := ak[3]
	if page == "" {
		page = "1"
	}
	if size == "" {
		size = "10"
	}
	category := ""
	tags := ""
	SKILL_MARKET_SEVERS := "http://127.0.0.1:8000"
	baseURL := SKILL_MARKET_SEVERS + "/api/v1/skills/search"

	params := url.Values{}

	params.Set("q", q)
	params.Set("category", category)
	params.Set("tags", tags)
	params.Set("sort", sort)
	params.Set("page", page)
	params.Set("size", size)

	s_url := baseURL + "?" + params.Encode()

	resp, err := http.Get(s_url)
	if err != nil {
		return fmt.Errorf("url请求失败：%w", err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("请求失败：", resp.StatusCode)
		return fmt.Errorf("error")

	}
	resBOBY, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("数据读取错误:%w", err)

	}
	var result SearchResponse2
	err = json.Unmarshal(resBOBY, &result)
	if err != nil {
		return fmt.Errorf("解析JSON失败:%w", err)

	}
	fmt.Println("API 状态码:", resp.StatusCode)
	fmt.Println("Message:", result.Message)
	fmt.Printf("找到了 %d skills:\n", result.Total)
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.AppendHeader(table.Row{
		"Public_id", "Slug", "Display_name", "Summary", "Category_name", "Tags", "Download_count", "Rating_avg", "Created_at"})
	// 表写完了，就将数据写进表里面
	for _, v := range result.Items {
		t.AppendRow(table.Row{v.Public_id, v.Slug, v.Display_name, v.Summary, v.Category_name, v.Tags, v.Download_count, v.Rating_avg, v.Created_at})
	}
	t.Render()
	fmt.Println(t.Render())
	return nil
}
