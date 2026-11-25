// 列表：本机已安装（读账本）和市场（调 search 接口）两种模式，共用一套打印。
package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"

	constantquantity "go_mod/constant_quantity"
	"go_mod/output"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// SkillListItem 对应后端 SkillListItem 模型
type SkillListItem struct {
	PublicID      string   `json:"public_id"`
	Slug          string   `json:"slug"`
	DisplayName   string   `json:"display_name"`
	Summary       string   `json:"summary"`
	CategoryName  *string  `json:"category_name"`
	Tags          []string `json:"tags"`
	DownloadCount int      `json:"download_count"`
	RatingAvg     float64  `json:"rating_avg"`
	CreatedAt     string   `json:"created_at"`
}

// SkillSearchResponse 对应后端 SkillSearchResponse
type SkillSearchResponse struct {
	Message string          `json:"message"`
	Total   int             `json:"total"`
	Items   []SkillListItem `json:"items"`
	Page    int             `json:"page"`
	Size    int             `json:"size"`
}

// maxPageSize 本地兜的上限。后端对 size 没有校验，传多大就查多大，这里替它挡一下。
const maxPageSize = 100

var validSorts = map[string]bool{"newest": true, "downloads": true, "rating": true}

// checkListArgs 在发请求前把非法的分页/排序参数拦下来。
//
// 不拦的后果都是"静默出错"：
//   - limit < 0：后端拼出 `LIMIT -1`，MySQL 直接抛异常 → 500 Internal Server Error
//   - limit = 0：后端正常返回空列表，用户以为"没有数据"，其实是他写错了
//   - sort 写错：后端 sort_map.get(sort, "s.created_at DESC") 静默回退成按时间排序
func checkListArgs(page, limit int, sortBy string) error {
	if page < 1 {
		return fmt.Errorf("--page 必须 ≥ 1（收到 %d）", page)
	}
	if limit < 1 || limit > maxPageSize {
		return fmt.Errorf("--limit 必须在 1~%d 之间（收到 %d）", maxPageSize, limit)
	}
	if !validSorts[sortBy] {
		return fmt.Errorf("--sort 只能是 newest / downloads / rating（收到 %q）", sortBy)
	}
	return nil
}

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出市场的技能，或本机已安装的技能",
	Example: "  skillctl skill list\n" +
		"  skillctl skill list --installed\n" +
		"  skillctl skill list --category automation --sort downloads",
	RunE: func(cmd *cobra.Command, args []string) error {
		installed, _ := cmd.Flags().GetBool("installed")
		if installed {
			items, err := collectInstalled()
			if err != nil {
				return err
			}
			if output.Enabled() {
				return output.Data(installedPayload{Count: len(items), Items: items})
			}
			printInstalled(items)
			return nil
		}

		category, _ := cmd.Flags().GetString("category")
		sortBy, _ := cmd.Flags().GetString("sort")
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		if err := checkListArgs(page, limit, sortBy); err != nil {
			return err
		}
		resp, err := SearchSkills("", category, sortBy, page, limit)
		if err != nil {
			return err
		}
		if output.Enabled() {
			return output.Data(resp)
		}
		printMarketList(resp)
		return nil
	},
}

func init() {
	ListCmd.Flags().Bool("installed", false, "只列本机已安装的技能(读 ~/.skillctl/registry.json)")
	ListCmd.Flags().String("category", "", "按分类过滤，填分类 slug（用 skillctl skill categories 查看）")
	ListCmd.Flags().String("sort", "newest", "排序方式:newest/downloads/rating")
	ListCmd.Flags().Int("page", 1, "页码")
	ListCmd.Flags().Int("limit", 10, "每页数量")
}

// SearchSkills 调 GET /api/v1/skills/search。q 为空就是"列全部"。
func SearchSkills(q, category, sortBy string, page, limit int) (*SkillSearchResponse, error) {
	u, err := url.Parse(constantquantity.ServerURL() + "/api/v1/skills/search")
	if err != nil {
		return nil, fmt.Errorf("构建搜索地址失败:%w", err)
	}
	qs := u.Query()
	if q != "" {
		qs.Set("q", q)
	}
	if category != "" {
		qs.Set("category", category)
	}
	if sortBy != "" {
		qs.Set("sort", sortBy)
	}
	qs.Set("page", strconv.Itoa(page))
	qs.Set("size", strconv.Itoa(limit))
	u.RawQuery = qs.Encode()

	// search 是公开接口，不带 token
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("构建搜索请求失败:%w", err)
	}

	resp, err := constantquantity.NewHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("连不上后端 %s，请确认 uvicorn 已启动:%w", constantquantity.ServerURL(), err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取搜索响应失败:%w", err)
	}
	if resp.StatusCode != http.StatusOK {
		// 后端 500 时可能是纯文本，直接截断展示，避免糊一屏
		return nil, fmt.Errorf("搜索失败，状态码 %d: %s", resp.StatusCode, truncate(string(data), 200))
	}

	var out SkillSearchResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败(服务端可能返回了非 JSON):%w", err)
	}
	return &out, nil
}

// installedItem --json 时的一条已安装记录。
type installedItem struct {
	Slug        string `json:"slug"`
	Target      string `json:"target"`
	Version     string `json:"version"`
	UUID        string `json:"uuid,omitempty"`
	Path        string `json:"path"`
	InstalledAt string `json:"installed_at"`
	Files       int    `json:"files"`
}

type installedPayload struct {
	Count int             `json:"count"`
	Items []installedItem `json:"items"`
}

// collectInstalled 从本地账本读，不发任何请求。
func collectInstalled() ([]installedItem, error) {
	reg, err := LoadRegistry()
	if err != nil {
		return nil, err
	}
	items := make([]installedItem, 0, len(reg.Skills))
	for slug, byTarget := range reg.Skills {
		for target, rec := range byTarget {
			items = append(items, installedItem{
				Slug:        slug,
				Target:      target,
				Version:     rec.Version,
				UUID:        rec.UUID,
				Path:        rec.Path,
				InstalledAt: rec.InstalledAt,
				Files:       len(rec.Files),
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Slug != items[j].Slug {
			return items[i].Slug < items[j].Slug
		}
		return items[i].Target < items[j].Target
	})
	return items, nil
}

// printInstalled 把已安装列表画成表格。
func printInstalled(items []installedItem) {
	if len(items) == 0 {
		fmt.Println("本机还没有安装任何技能，用 skillctl skill install <slug> 开始")
		return
	}
	fmt.Printf("已安装 %d 个（技能×agent）\n\n", len(items))

	// 用 go-pretty 而不是 text/tabwriter：tabwriter 按字节算宽度，中文会错位
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"SLUG", "VERSION", "TARGET", "INSTALLED", "PATH"})
	for _, r := range items {
		t.AppendRow(table.Row{r.Slug, r.Version, r.Target, shortDate(r.InstalledAt), humanPath(r.Path)})
	}
	t.Render()
}

// printMarketList 打印市场搜索结果。
func printMarketList(resp *SkillSearchResponse) {
	if resp.Total == 0 || len(resp.Items) == 0 {
		fmt.Println("没有找到匹配的技能")
		return
	}
	fmt.Printf("共 %d 个技能（第 %d 页，每页 %d）\n\n", resp.Total, resp.Page, resp.Size)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"SLUG", "NAME", "CATEGORY", "DOWNLOADS", "RATING"})
	for _, it := range resp.Items {
		category := "-"
		if it.CategoryName != nil && *it.CategoryName != "" {
			category = *it.CategoryName
		}
		t.AppendRow(table.Row{it.Slug, it.DisplayName, category, it.DownloadCount, fmt.Sprintf("%.1f", it.RatingAvg)})
	}
	t.Render()
}

// shortDate 取 RFC3339 的日期部分，表格里不需要秒。
func shortDate(ts string) string {
	if len(ts) >= 10 {
		return ts[:10]
	}
	return ts
}

// truncate 截断长错误体，避免把整个 HTML 错误页打到终端。
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
