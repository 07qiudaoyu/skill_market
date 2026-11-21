// 卸载：把某个 (slug, target) 从磁盘和账本里一起去掉。
//
// 安全原则：账本里的 path 只当线索，不当命令。真实路径永远由 target 重算，
// 和账本不一致就拒绝删除——账本可能被手改，也可能来自旧的 target 映射。
// --force 只放宽"删多少"（允许连用户自建文件一起删），绝不放宽"删哪里"。
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go_mod/output"

	"github.com/spf13/cobra"
)

// UninstallResult 卸载结果，给命令层打印用。
type UninstallResult struct {
	Slug    string   `json:"slug"`
	Target  string   `json:"target"`
	Path    string   `json:"path"`
	Removed int      `json:"removed"`        // 删掉的文件数
	Kept    []string `json:"kept,omitempty"` // 目录里剩下的、不是 skillctl 装的文件（相对路径）
	DirGone bool     `json:"dirGone"`        // 技能目录是否已经整个消失
}

// uninstallPayload --json 时的 data 结构。
type uninstallPayload struct {
	Count int               `json:"count"`
	Items []UninstallResult `json:"items"`
}

var UninstallCmd = &cobra.Command{
	Use:   "uninstall <slug>",
	Short: "卸载技能，同时清掉本地账本记录",
	Long: "只删 skillctl 自己装进去的文件，你在技能目录里手写的文件会保留并提示。\n" +
		"技能装在多个 agent 上时需要 --target 指定，或 --all 全部卸载。",
	Example: "  skillctl skill uninstall go-review\n" +
		"  skillctl skill uninstall go-review --target codex\n" +
		"  skillctl skill uninstall go-review --all --force",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetString("target")
		all, _ := cmd.Flags().GetBool("all")
		force, _ := cmd.Flags().GetBool("force")

		// 部分成功也要先打印再返回错误，否则用户看不到删掉了什么
		results, err := Uninstall(args[0], target, all, force)
		if output.Enabled() {
			payload := uninstallPayload{Count: len(results), Items: results}
			if err != nil {
				return output.FailWithData(err, payload)
			}
			return output.Data(payload)
		}
		for _, r := range results {
			printUninstall(r)
		}
		return err
	},
}

func init() {
	UninstallCmd.Flags().String("target", "", "只卸载这个 agent 上的:"+strings.Join(ValidTargets(), "/"))
	UninstallCmd.Flags().Bool("all", false, "卸载该技能在所有 agent 上的安装")
	UninstallCmd.Flags().Bool("force", false, "连目录里不是你安装的文件一起删掉")
}

// Uninstall 卸载 slug。target 为空且只装在一处时，自动选那一个。
func Uninstall(slug, targetFlag string, all bool, force bool) ([]UninstallResult, error) {
	slug, err := cleanSlug(slug)
	if err != nil {
		return nil, err
	}
	if targetFlag != "" && all {
		return nil, fmt.Errorf("--target 和 --all 不能同时用")
	}

	reg, err := LoadRegistry()
	if err != nil {
		return nil, err
	}
	byTarget := reg.Skills[slug]
	if len(byTarget) == 0 {
		return nil, fmt.Errorf("本机没有安装 %s（用 skillctl skill list --installed 看已装列表）", slug)
	}

	var targets []string
	switch {
	case all:
		targets = sortedTargets(byTarget)
	case targetFlag != "":
		if _, ok := byTarget[targetFlag]; !ok {
			return nil, fmt.Errorf("%s 没有装在 %s 上（已装在:%s）",
				slug, targetFlag, strings.Join(sortedTargets(byTarget), ", "))
		}
		targets = []string{targetFlag}
	case len(byTarget) > 1:
		// 装了好几处又不指定：猜哪个都是错，让用户显式选
		return nil, fmt.Errorf("%s 装在多个 agent 上（%s），请用 --target 指定，或用 --all 全部卸载",
			slug, strings.Join(sortedTargets(byTarget), ", "))
	default:
		targets = sortedTargets(byTarget)
	}

	var results []UninstallResult
	for _, t := range targets {
		res, err := uninstallOne(reg, slug, t, force)
		if err != nil {
			return results, fmt.Errorf("卸载 %s@%s 失败:%w", slug, t, err)
		}
		// 每个 target 删完立刻记账：中途失败时，已成功的那些不该还留在账本里
		reg.Delete(slug, t)
		if err := SaveRegistry(reg); err != nil {
			return results, fmt.Errorf("%s@%s 文件已删但账本没写成功:%w", slug, t, err)
		}
		results = append(results, res)
	}
	return results, nil
}

// uninstallOne 删一个 (slug, target)。四道校验不做 --force 豁免，顺序从便宜到贵。
func uninstallOne(reg *Registry, slug, target string, force bool) (UninstallResult, error) {
	res := UninstallResult{Slug: slug, Target: target}
	rec, ok := reg.Get(slug, target)
	if !ok {
		return res, fmt.Errorf("账本里没有这条记录")
	}

	// ① 目标目录只认白名单，路径现算，不用账本里的 path
	baseDir, err := TargetDir(target)
	if err != nil {
		return res, err
	}
	want := filepath.Join(baseDir, slug)

	// ② 算出来的路径先过三连否决，挡掉白名单常量配错这种连锁反应
	if !filepath.IsAbs(want) {
		return res, fmt.Errorf("算出来的安装路径不是绝对路径(%q)，拒绝操作", want)
	}
	if home, err := os.UserHomeDir(); err == nil {
		if want == home || want == filepath.Dir(home) || want == string(os.PathSeparator) {
			return res, fmt.Errorf("算出来的安装路径指向家目录或上层目录(%q)，拒绝操作", want)
		}
	}
	if filepath.Base(want) != slug {
		return res, fmt.Errorf("目标目录名 %q 和 slug 不符，拒绝操作", filepath.Base(want))
	}
	res.Path = want

	// ③ 账本 path 必须和现算的一致。不一致就拒绝，而不是"以 want 为准继续删"：
	// 映射变了的时候，want 上可能正好住着另一个同名技能。
	if filepath.Clean(rec.Path) != filepath.Clean(want) {
		return res, fmt.Errorf("账本记录的路径是 %s，但 %s 的正确位置是 %s，不一致，拒绝删除（账本可能被手改，或 target 映射变过）",
			humanPath(rec.Path), target, humanPath(want))
	}

	// ④ files 只允许文件级相对路径，且清理后不能跳出安装目录
	files := make([]string, 0, len(rec.Files))
	for _, rel := range rec.Files {
		if rel == "" {
			continue
		}
		if filepath.IsAbs(rel) {
			return res, fmt.Errorf("账本里出现绝对路径 %q，拒绝删除", rel)
		}
		full := filepath.Join(want, rel)
		if full == filepath.Clean(want) {
			return res, fmt.Errorf("账本里的路径 %q 指向安装目录本身，拒绝删除", rel)
		}
		if !isInside(want, full) {
			return res, fmt.Errorf("账本里的路径 %q 会跳出安装目录，拒绝删除", rel)
		}
		files = append(files, full)
	}

	// 目录早就没了（用户手删过）：只剩清账本，目的已达成
	if _, err := os.Stat(want); err != nil {
		if os.IsNotExist(err) {
			res.DirGone = true
			return res, nil
		}
		return res, fmt.Errorf("检查 %s 失败:%w", humanPath(want), err)
	}

	if force {
		// --force：整个目录移除，连不是你装的文件一起
		count, err := countFiles(want)
		if err != nil {
			return res, err
		}
		if err := os.RemoveAll(want); err != nil {
			return res, fmt.Errorf("删除 %s 失败:%w", humanPath(want), err)
		}
		res.Removed = count
		res.DirGone = true
		return res, nil
	}

	// 默认：只删账本清单里的文件
	for _, p := range files {
		if err := os.Remove(p); err != nil {
			if os.IsNotExist(err) {
				continue // 用户手删过，不算错
			}
			return res, fmt.Errorf("删除 %s 失败:%w", humanPath(p), err)
		}
		res.Removed++
	}

	// 清掉删空的子目录，不碰技能目录本身
	for _, p := range files {
		removeEmptyParents(filepath.Dir(p), want)
	}

	// 看还剩什么：空就整目录清掉，非空就保留并报告
	leftover, err := listFiles(want)
	if err != nil {
		return res, err
	}
	if len(leftover) == 0 {
		if err := os.RemoveAll(want); err != nil {
			return res, fmt.Errorf("清理空目录 %s 失败:%w", humanPath(want), err)
		}
		res.DirGone = true
	} else {
		res.Kept = leftover
	}
	return res, nil
}

func printUninstall(r UninstallResult) {
	fmt.Printf("已卸载 %s（target %s）\n", r.Slug, r.Target)
	if r.Path != "" {
		fmt.Printf("  路径  %s\n", humanPath(r.Path))
	}
	if r.DirGone {
		fmt.Printf("  文件  删除 %d 个，技能目录已清空\n", r.Removed)
	} else {
		fmt.Printf("  文件  删除 %d 个\n", r.Removed)
	}
	if len(r.Kept) == 0 {
		return
	}
	fmt.Printf("  保留  目录里还有 %d 个不是 skillctl 装的文件:\n", len(r.Kept))
	show := r.Kept
	if len(show) > 5 {
		show = show[:5]
	}
	for _, f := range show {
		fmt.Printf("        %s\n", f)
	}
	if len(r.Kept) > len(show) {
		fmt.Printf("        ...另有 %d 个\n", len(r.Kept)-len(show))
	}
	// 账本已经清掉了，所以不能提示"重跑加 --force"；只能给路径让人自己处理
	fmt.Printf("  提示  这些文件与 skillctl 无关，已保留；需要清理请手动处理上面的路径\n")
}

// sortedTargets 把某个 slug 下的 target 排序输出，保证报错和打印顺序稳定。
func sortedTargets(m map[string]InstallRecord) []string {
	out := make([]string, 0, len(m))
	for name := range m {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// isInside 判断 p 是否落在 root 之内（root 自身算在内）。
func isInside(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if filepath.IsAbs(rel) {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// listFiles 列出 root 下所有文件的相对路径（含隐藏文件）。
func listFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("扫描 %s 失败:%w", humanPath(root), err)
	}
	sort.Strings(out)
	return out, nil
}

// countFiles 数 root 下的文件个数，用于 --force 时报告删了多少。
func countFiles(root string) (int, error) {
	files, err := listFiles(root)
	if err != nil {
		return 0, err
	}
	return len(files), nil
}
