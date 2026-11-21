// 升级：把已安装的技能换到新版本。
//
// 复用 install 的覆盖逻辑，区别有两点：
// 1. 只下载一次，给该技能装过的每个 agent 各装一遍；
// 2. 先用 public_id（就是账本里的 uuid）问出最新版本号，全部已是最新时一次下载都不发生。
//
// 修复被删坏的文件不是 upgrade 的职责，用 install 重装。
package skills

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	constantquantity "go_mod/constant_quantity"
	"go_mod/output"

	"github.com/spf13/cobra"
)

// UpgradeResult 一个 target 的升级结果。
type UpgradeResult struct {
	Slug    string `json:"slug"`
	Target  string `json:"target"`
	From    string `json:"from"`
	To      string `json:"to"`
	Path    string `json:"path"`
	Copied  int    `json:"copied"`
	Removed int    `json:"removed"`
	Skipped bool   `json:"skipped"` // 版本已一致，没动磁盘
}

// upgradePayload --json 时的 data 结构。
type upgradePayload struct {
	Count int             `json:"count"`
	Items []UpgradeResult `json:"items"`
}

var UpgradeCmd = &cobra.Command{
	Use:   "upgrade <slug>",
	Short: "把已安装的技能升到最新版",
	Long: "升级该技能装过的每一个 agent。已经是最新版的会跳过，不重复下载覆盖。\n" +
		"想修复被删坏的文件请用 install，不是 upgrade 的职责。",
	Example: "  skillctl skill upgrade go-review\n" +
		"  skillctl skill upgrade go-review --version 1.2.0",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version, _ := cmd.Flags().GetString("version")

		output.Say("升级 %s", args[0])
		// 部分成功也要先打印再返回错误
		results, err := Upgrade(args[0], version)
		if output.Enabled() {
			payload := upgradePayload{Count: len(results), Items: results}
			if err != nil {
				return output.FailWithData(err, payload)
			}
			return output.Data(payload)
		}
		for _, r := range results {
			printUpgrade(r)
		}
		if err != nil {
			return err
		}
		printUpgradeSummary(results)
		return nil
	},
}

func init() {
	UpgradeCmd.Flags().String("version", "", "升到指定版本，不传则升到最新版本")
}

// Upgrade 把 slug 装过的所有 target 升到 desired 版本。
func Upgrade(slug, versionFlag string) ([]UpgradeResult, error) {
	slug, err := cleanSlug(slug)
	if err != nil {
		return nil, err
	}

	reg, err := LoadRegistry()
	if err != nil {
		return nil, err
	}
	byTarget := reg.Skills[slug]
	if len(byTarget) == 0 {
		return nil, fmt.Errorf("本机没有安装 %s（先 skillctl skill install %s）", slug, slug)
	}
	targets := sortedTargets(byTarget)

	// 先算出目标版本：知道版本号，就能在"全都已是最新"时一次下载都不发生。
	desired := strings.TrimSpace(versionFlag)
	if desired == "" {
		if id := anyUUID(byTarget, targets); id != "" {
			// 查不到不致命：退回"下载后从文件名解析"，只是多下载一次
			if v, err := fetchLatestVersion(id); err == nil {
				desired = v
			}
		}
	}

	if desired != "" && allAtVersion(byTarget, targets, desired) {
		results := make([]UpgradeResult, 0, len(targets))
		for _, t := range targets {
			results = append(results, UpgradeResult{
				Slug:    slug,
				Target:  t,
				From:    desired,
				To:      desired,
				Path:    byTarget[t].Path,
				Skipped: true,
			})
		}
		return results, nil
	}

	// 只下载一次，所有 target 共用这个 zip
	zipPath, filename, uuid, err := downloadZip(slug, desired)
	if err != nil {
		return nil, err
	}
	defer os.Remove(zipPath)

	actual := pickVersion(desired, filename)

	var results []UpgradeResult
	for _, t := range targets {
		rec := byTarget[t]
		if rec.Version == actual {
			results = append(results, UpgradeResult{
				Slug:    slug,
				Target:  t,
				From:    rec.Version,
				To:      actual,
				Path:    rec.Path,
				Skipped: true,
			})
			continue
		}
		// 账本里的 target 已不在白名单（账本被改 / 映射回退）：数据异常，不要静默跳过
		if _, err := TargetDir(t); err != nil {
			return results, fmt.Errorf("账本里 %s 记在 %q 上，但它已不是合法 target:%w", slug, t, err)
		}

		res, err := installFromZip(zipPath, filename, uuid, slug, t, actual)
		if err != nil {
			return results, fmt.Errorf("升级 %s@%s 失败:%w", slug, t, err)
		}
		results = append(results, UpgradeResult{
			Slug:    slug,
			Target:  t,
			From:    rec.Version,
			To:      res.Version,
			Path:    res.Path,
			Copied:  res.Copied,
			Removed: res.Removed,
		})
	}
	return results, nil
}

// fetchLatestVersion 问后端某技能有哪些版本，取最新上传的一条。
// uuid 形如 uuid-5，尾巴的数字就是 skills.id，后端直接当 public_id 用。
func fetchLatestVersion(publicID string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, FindVersionsURL(publicID), nil)
	if err != nil {
		return "", err
	}
	resp, err := constantquantity.NewHTTPClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("查询版本失败:%s", resp.Status)
	}

	var out struct {
		Versions []struct {
			Version string `json:"version"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("解析版本列表失败:%w", err)
	}
	if len(out.Versions) == 0 {
		return "", fmt.Errorf("服务端没有可用版本")
	}
	// 后端按 created_at desc 排序，第一条就是最新上传的
	return out.Versions[0].Version, nil
}

// anyUUID 取该 slug 任意一条记录上的 uuid。同一个技能装到多个 agent，
// uuid 是同一个（都是那个包），所以取第一个非空的就够。
func anyUUID(byTarget map[string]InstallRecord, targets []string) string {
	for _, t := range targets {
		if id := byTarget[t].UUID; id != "" {
			return id
		}
	}
	return ""
}

func allAtVersion(byTarget map[string]InstallRecord, targets []string, version string) bool {
	for _, t := range targets {
		if byTarget[t].Version != version {
			return false
		}
	}
	return true
}

func printUpgrade(r UpgradeResult) {
	if r.Skipped {
		fmt.Printf("  %-12s %s（已是最新，跳过）\n", r.Target, r.From)
		return
	}
	fmt.Printf("  %-12s %s → %s（复制 %d，删除 %d）\n", r.Target, r.From, r.To, r.Copied, r.Removed)
}

func printUpgradeSummary(results []UpgradeResult) {
	var up, skip int
	for _, r := range results {
		if r.Skipped {
			skip++
		} else {
			up++
		}
	}
	if up == 0 {
		fmt.Printf("\n全部已是最新，无需操作\n")
		return
	}
	fmt.Printf("\n升级 %d 个，跳过 %d 个\n", up, skip)
}
