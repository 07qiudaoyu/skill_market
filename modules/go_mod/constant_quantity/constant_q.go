// 常量与全局地址：所有对外请求的地址都从这里取，避免散落在各业务文件里。
package constantquantity

import (
	"net/http"
	"os"
	"time"
)

// Default_url 后端地址默认值。本地开发指向 uvicorn 的 8000 端口。
const Default_url = "http://127.0.0.1:8000"

// HTTPTimeout 单个请求的超时时间。
// 不设的话，后端没起时要等系统 TCP 超时(约 75 秒)才报错，用户会以为程序卡死。
const HTTPTimeout = 10 * time.Second

// NewHTTPClient 返回带超时的 client。所有对外请求都用它，不要用裸的 &http.Client{}。
func NewHTTPClient() *http.Client {
	return &http.Client{Timeout: HTTPTimeout}
}

// FlagServer 由 root 的 --server 写入，用来临时切环境。
var FlagServer string

// ServerURL 解析优先级：--server > SKILLCTL_SERVER 环境变量 > Default_url。
func ServerURL() string {
	if FlagServer != "" {
		return FlagServer
	}
	if s := os.Getenv("SKILLCTL_SERVER"); s != "" {
		return s
	}
	return Default_url
}
