// 安装目标（target）的白名单与解析。
//
// 核心约束：安装目录由 agent 决定，不由用户决定。
// 用户如果能把 skill 放到自己的目录，agent 根本不会去读，装完等于没装。
// 所以这里只认白名单，用户要新 agent 只能等我们加一行映射。
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go_mod/output"
)

// targetPaths target 名 → 相对用户主目录的安装目录。
var targetPaths = map[string]string{
	"global":      ".agents/skills",
	"claude-code": ".claude/skills",
	"codex":       ".codex/skills",
	"cursor":      ".cursor/skills",
	"qoder":       ".qoder/skills",
	"dsh":         ".dsh/skills",
}

// ValidTargets 返回排好序的合法 target，用于报错时列给用户。
func ValidTargets() []string {
	names := make([]string, 0, len(targetPaths))
	for name := range targetPaths {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ResolveTarget 决定这次装给哪个 agent。
//
// 显式传了 --target 就用它；没传就扫一遍 agent 专属目录：
//   - 命中唯一一个 → 本次使用它（只对本次生效，不写进配置当永久默认）
//   - 命中多个 / 零个 → 报错要求显式指定，绝不瞎猜
//
// global 是兜底目录，不参与自动检测（~/.agents/skills 不存在是常态）。
func ResolveTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target != "" {
		if _, ok := targetPaths[target]; !ok {
			return "", fmt.Errorf("未知的 target %q，可选值:%s", target, strings.Join(ValidTargets(), ", "))
		}
		return target, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败:%w", err)
	}

	var found []string
	for name, rel := range targetPaths {
		if name == "global" {
			continue
		}
		if st, err := os.Stat(filepath.Join(home, rel)); err == nil && st.IsDir() {
			found = append(found, name)
		}
	}
	sort.Strings(found)

	switch len(found) {
	case 1:
		output.Say("检测到 %s，本次安装到 %s（可用 --target 指定其他）", found[0], found[0])
		return found[0], nil
	case 0:
		return "", fmt.Errorf("没有检测到任何 agent 目录，请用 --target 指定:%s", strings.Join(ValidTargets(), ", "))
	default:
		return "", fmt.Errorf("检测到多个 agent（%s），请用 --target 指定装给谁", strings.Join(found, ", "))
	}
}

// TargetDir 返回某个 target 的 skill 根目录，不创建。
func TargetDir(target string) (string, error) {
	rel, ok := targetPaths[target]
	if !ok {
		return "", fmt.Errorf("未知的 target %q，可选值:%s", target, strings.Join(ValidTargets(), ", "))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败:%w", err)
	}
	return filepath.Join(home, rel), nil
}
