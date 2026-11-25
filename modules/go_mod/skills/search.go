// 关键词搜索：跟 list 共用同一套请求和打印，区别只在于必须给一个关键词。
package skills

import (
	"go_mod/output"

	"github.com/spf13/cobra"
)

var SearchCmd = &cobra.Command{
	Use:     "search <keyword>",
	Short:   "按关键词搜索市场技能",
	Example: "  skillctl skill search email\n  skillctl skill search crawler --category automation --limit 20",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		sortBy, _ := cmd.Flags().GetString("sort")
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		if err := checkListArgs(page, limit, sortBy); err != nil {
			return err
		}
		resp, err := SearchSkills(args[0], category, sortBy, page, limit)
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
	SearchCmd.Flags().String("category", "", "按分类过滤，填分类 slug（用 skillctl skill categories 查看）")
	SearchCmd.Flags().String("sort", "newest", "排序方式:newest/downloads/rating")
	SearchCmd.Flags().Int("page", 1, "页码")
	SearchCmd.Flags().Int("limit", 10, "每页数量")
}
