// 技能相关接口的地址组装。集中在这个文件，接口改名时只改这里。
package skills

import (
	"net/url"

	constantquantity "go_mod/constant_quantity"
)

// CliDownloadURL 组装 /api/v1/skills/Cli_downloadzip。
//
// 后端签名是 slug + versions（注意是复数），versions 为空表示"取最新版本"，
// 由后端的 Re_f_newest_versions 决定。不要写成 version，会被静默忽略。
func CliDownloadURL(slug, version string) string {
	q := url.Values{}
	q.Set("slug", slug)
	if version != "" {
		q.Set("versions", version)
	}
	return constantquantity.ServerURL() + "/api/v1/skills/Cli_downloadzip?" + q.Encode()
}

// FindVersionsURL 组装 /api/v1/skills/find_versions。
//
// 注意参数是 public_id，不是 slug：public_id 就是账本里的 uuid（形如 uuid-5，
// 尾巴的数字等于 skills.id）。upgrade 用它先问出最新版本号，省掉无意义的下载。
func FindVersionsURL(publicID string) string {
	q := url.Values{}
	q.Set("public_id", publicID)
	return constantquantity.ServerURL() + "/api/v1/skills/find_versions?" + q.Encode()
}

// Ddownload_url 旧常量，install.go 重写前临时保留，避免这一版编译不过。
const Ddownload_url = constantquantity.Default_url + "/api/v1/skills/Cli_downloadzip?slug="
