// 安装：下载 zip → 解压到 ~/.skillctl/temp 校验 → 逐文件覆盖到 agent 目录 → 写账本。
//
// 三个刻意的选择：
// 1. staging 只用于"解压和校验"，落地是逐文件覆盖 —— 校验失败时目标目录一个字节都没动。
// 2. 不做备份、不做回滚：覆盖是幂等的，中途失败重跑一次 install 就收敛了。
// 3. 只删"上一版账本里有、这个包里没有"的文件，用户自己加的文件不动。
package skills

import (
	"archive/zip"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	constantquantity "go_mod/constant_quantity"
	"go_mod/output"
	"go_mod/security"

	"github.com/spf13/cobra"
)

// InstallResult 安装结果，给命令层打印用。
type InstallResult struct {
	Slug    string `json:"slug"`
	Version string `json:"version"`
	Target  string `json:"target"`
	Path    string `json:"path"`
	Copied  int    `json:"copied"`
	Removed int    `json:"removed"`
}

var InstallCmd = &cobra.Command{
	Use:   "install <slug>",
	Short: "下载并安装技能到指定 agent 的目录",
	Long: "把技能装到 agent 会读的目录里。不传 --target 时会自动检测已安装的 agent，\n" +
		"只有唯一命中才自动使用；命中多个或一个都没有时会要求显式指定。",
	Example: "  skillctl skill install go-review\n" +
		"  skillctl skill install go-review --target claude-code --version 1.0.0",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version, _ := cmd.Flags().GetString("version")
		target, _ := cmd.Flags().GetString("target")

		res, err := Install(args[0], version, target)
		if err != nil {
			return err
		}
		if output.Enabled() {
			return output.Data(res)
		}
		fmt.Printf("已安装 %s@%s\n", res.Slug, res.Version)
		fmt.Printf("  target  %s\n", res.Target)
		fmt.Printf("  路径    %s\n", humanPath(res.Path))
		fmt.Printf("  文件    复制 %d 个，删除 %d 个\n", res.Copied, res.Removed)
		fmt.Printf("  校验    SKILL.md 在位\n")
		return nil
	},
}

func init() {
	InstallCmd.Flags().String("version", "", "指定版本，不传则安装最新版本")
	InstallCmd.Flags().String("target", "", "安装给哪个 agent:"+strings.Join(ValidTargets(), "/"))
}

// Install 完成一次安装动作。version 为空表示最新版，target 为空表示自动检测。
func Install(slug, version, targetFlag string) (*InstallResult, error) {
	slug, err := cleanSlug(slug)
	if err != nil {
		return nil, err
	}

	target, err := ResolveTarget(targetFlag)
	if err != nil {
		return nil, err
	}

	// 下载一次。upgrade 要给多个 target 装同一个包，会跳过本函数直接复用同一个 zip。
	zipPath, filename, uuid, err := downloadZip(slug, version)
	if err != nil {
		return nil, err
	}
	defer os.Remove(zipPath)

	return installFromZip(zipPath, filename, uuid, slug, target, version)
}

// cleanSlug 校验 slug。它会拼进目录名，必须挡掉路径穿越。
func cleanSlug(slug string) (string, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", fmt.Errorf("slug 不能为空")
	}
	if slug == "." || slug == ".." || strings.ContainsAny(slug, `/\`) {
		return "", fmt.Errorf("非法 slug %q", slug)
	}
	return slug, nil
}

// installFromZip 把已下载的 zip 装进某个 target，zip 的删除由调用方负责。
// 拆出来是为了让多个 target 复用同一次下载。
func installFromZip(zipPath, filename, uuid, slug, target, requestVersion string) (*InstallResult, error) {
	baseDir, err := TargetDir(target)
	if err != nil {
		return nil, err
	}
	installDir := filepath.Join(baseDir, slug)

	// 1. 解压到 ~/.skillctl/temp/<slug>@<target>
	tempRoot, err := security.SubDir("temp")
	if err != nil {
		return nil, err
	}
	stageDir := filepath.Join(tempRoot, slug+"@"+target)
	os.RemoveAll(stageDir)
	defer os.RemoveAll(stageDir)

	files, err := extractZip(zipPath, stageDir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("这个包里没有任何文件，安装中止")
	}

	// 2. 校验：必须是一个像样的 skill
	if err := validateSkill(stageDir); err != nil {
		return nil, err
	}

	// 3. 读上一版清单，用于算出该删掉哪些旧文件
	reg, err := LoadRegistry()
	if err != nil {
		return nil, err
	}
	var oldFiles []string
	if rec, ok := reg.Get(slug, target); ok {
		oldFiles = rec.Files
	}

	// 4. 覆盖落地
	copied, removed, err := overlay(stageDir, installDir, files, oldFiles)
	if err != nil {
		return nil, err
	}

	version := pickVersion(requestVersion, filename)

	// 5. 全部成功之后才写账本。先写账本再落盘的话，
	//    中途失败会留下"账本说有、磁盘上没有"的孤儿记录。
	reg.Put(slug, target, InstallRecord{
		Version: version,
		UUID:    uuid,
		Path:    installDir,
		Files:   files,
	})
	if err := SaveRegistry(reg); err != nil {
		return nil, err
	}

	return &InstallResult{
		Slug:    slug,
		Version: version,
		Target:  target,
		Path:    installDir,
		Copied:  copied,
		Removed: removed,
	}, nil
}

// downloadZip 请求后端拿 zip，返回本地缓存路径、服务端给的文件名和 uuid。
//
// 必须把 body 读完再关：后端是在流的尾部写下载计数的，提前中断就不记账。
func downloadZip(slug, version string) (zipPath, filename, uuid string, err error) {
	url := CliDownloadURL(slug, version)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("构建下载请求失败:%w", err)
	}
	if err := security.JwtHeader(req); err != nil {
		return "", "", "", err
	}

	resp, err := constantquantity.NewHTTPClient().Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("连不上后端 %s，请确认 uvicorn 已启动:%w", constantquantity.ServerURL(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return "", "", "", fmt.Errorf("鉴权失败，请重新登入:skillctl auth login")
		case http.StatusNotFound:
			return "", "", "", fmt.Errorf("服务端找不到 slug=%s 的该版本:%s", slug, strings.TrimSpace(string(body)))
		}
		return "", "", "", fmt.Errorf("下载失败，状态码 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// 后端的 uuid 塞在 Content-Disposition 里，标准头装私有字段，只能这样解析
	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return "", "", "", fmt.Errorf("解析 Content-Disposition 失败:%w", err)
	}
	filename = params["filename"]
	uuid = params["uuid"]
	if filename == "" {
		filename = slug + ".zip"
	}

	cacheDir, err := security.SubDir("cache")
	if err != nil {
		return "", "", "", err
	}
	// 用 Base 兜一层，防止服务端文件名里带路径
	zipPath = filepath.Join(cacheDir, filepath.Base(filename))

	out, err := os.Create(zipPath)
	if err != nil {
		return "", "", "", fmt.Errorf("创建缓存文件失败:%w", err)
	}
	// 读完整条流，后端尾部记账依赖这个
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(zipPath)
		return "", "", "", fmt.Errorf("下载中断:%w", err)
	}
	if err := out.Close(); err != nil {
		os.Remove(zipPath)
		return "", "", "", fmt.Errorf("关闭缓存文件失败:%w", err)
	}
	return zipPath, filename, uuid, nil
}

// extractZip 解压到 destDir，返回包内文件相对路径清单（统一用 / 分隔）。
func extractZip(zipPath, destDir string) ([]string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("文件不是合法的 zip:%w", err)
	}
	defer r.Close()

	var files []string
	for _, f := range r.File {
		rel, err := safeRelPath(f.Name)
		if err != nil {
			return nil, err
		}
		if rel == "" {
			continue
		}
		// macOS 打包会带这两个，不算技能内容
		if strings.HasPrefix(rel, "__MACOSX/") || filepath.Base(rel) == ".DS_Store" {
			continue
		}

		dest := filepath.Join(destDir, filepath.FromSlash(rel))
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, 0755); err != nil {
				return nil, fmt.Errorf("创建目录 %s 失败:%w", dest, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return nil, fmt.Errorf("创建目录失败:%w", err)
		}
		if err := writeZipFile(f, dest); err != nil {
			return nil, err
		}
		files = append(files, rel)
	}
	return files, nil
}

// safeRelPath 防 zip slip：拒绝绝对路径和跳出解压目录的 ..。
func safeRelPath(name string) (string, error) {
	clean := path.Clean(strings.ReplaceAll(name, `\`, "/"))
	if clean == "." {
		return "", nil
	}
	if path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("zip 含非法路径 %q，安装中止", name)
	}
	return clean, nil
}

// writeZipFile 把单个 zip 条目写成文件。
func writeZipFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("读取 zip 条目 %s 失败:%w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败:%w", dest, err)
	}
	if _, err := io.Copy(out, rc); err != nil {
		out.Close()
		return fmt.Errorf("写入 %s 失败:%w", dest, err)
	}
	return out.Close()
}

// validateSkill 校验"这确实是一个 skill"：
// 根目录必须有 SKILL.md，且大小写完全一致（macOS 文件系统不区分大小写，
// 光靠 Stat 会把 skill.md 当成 SKILL.md 放过去，但 agent 只认大写的）。
func validateSkill(stageDir string) error {
	entries, err := os.ReadDir(stageDir)
	if err != nil {
		return fmt.Errorf("读取解压目录失败:%w", err)
	}
	for _, e := range entries {
		if e.Name() == "SKILL.md" && !e.IsDir() {
			return nil
		}
	}
	return fmt.Errorf("这个包里没有 SKILL.md，不算一个可安装的技能")
}

// overlay 把 staging 里的文件逐个覆盖到安装目录，并删掉旧版遗留的文件。
func overlay(stageDir, installDir string, newFiles, oldFiles []string) (copied, removed int, err error) {
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return 0, 0, fmt.Errorf("创建安装目录失败:%w", err)
	}

	for _, rel := range newFiles {
		src := filepath.Join(stageDir, filepath.FromSlash(rel))
		dst := filepath.Join(installDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return copied, removed, fmt.Errorf("创建目录失败:%w", err)
		}
		if err := copyFile(src, dst); err != nil {
			return copied, removed, err
		}
		copied++
	}

	// 上一版有、这一版没有的文件要删掉，否则升级会留下垃圾。
	// 用户自己加的文件不在 oldFiles 里，永远不会被碰。
	newSet := make(map[string]bool, len(newFiles))
	for _, f := range newFiles {
		newSet[f] = true
	}
	for _, rel := range oldFiles {
		if newSet[rel] {
			continue
		}
		p := filepath.Join(installDir, filepath.FromSlash(rel))
		if err := os.Remove(p); err == nil {
			removed++
			removeEmptyParents(filepath.Dir(p), installDir)
		}
	}
	return copied, removed, nil
}

// copyFile 覆盖式复制单个文件。
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开 %s 失败:%w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("创建 %s 失败:%w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("写 %s 失败:%w", dst, err)
	}
	return out.Close()
}

// removeEmptyParents 自底向上删掉空目录，删到 stopAt 就停。
func removeEmptyParents(dir, stopAt string) {
	for {
		if dir == stopAt || !strings.HasPrefix(dir, stopAt+string(os.PathSeparator)) {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

// pickVersion 用户显式指定的版本优先，否则从服务端文件名 <slug>_v<version>.zip 里解析。
func pickVersion(version, filename string) string {
	if version != "" {
		return version
	}
	base := strings.TrimSuffix(filepath.Base(filename), ".zip")
	if i := strings.LastIndex(base, "_v"); i >= 0 {
		return base[i+2:]
	}
	return "unknown"
}

// humanPath 把家目录前缀换成 ~，输出短一点。
func humanPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}
