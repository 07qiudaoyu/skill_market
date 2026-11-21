// 本地状态与鉴权：~/.skillctl/config.json 的读写，以及请求要用的 Authorization 头。
//
// 为什么放 ~/.skillctl 而不是项目目录：用户可能把二进制拷到别处、或者换一个
// clone 目录，账本和 token 不能跟着源码走。os.UserHomeDir() 也方便测试时用
// HOME=/tmp/xxx 整体隔离。
//
// token 取值顺序：DR_ACCESS_TOKEN > DR_SKILL_MARKET_TOKEN > --token > config.json。
// 已弃用 go-keyring：Linux 无桌面环境和 CI 上必挂，跨平台失败模式太多。
package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// FlagToken 由 root 的 --token 写入，优先级排在环境变量之后、配置文件之前。
var FlagToken string

// Config 对应 ~/.skillctl/config.json。不保存密码。
type Config struct {
	Token         string `json:"token,omitempty"`
	Email         string `json:"email,omitempty"`
	UserID        int    `json:"userId,omitempty"`
	Name          string `json:"name,omitempty"`
	DefaultTarget string `json:"defaultTarget,omitempty"`
}

// SkillctlDir 返回 ~/.skillctl，CLI 的全部本地状态都住在这里。
func SkillctlDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败:%w", err)
	}
	return filepath.Join(home, ".skillctl"), nil
}

// SubDir 返回 ~/.skillctl/<name> 并确保目录存在。temp 和 cache 都走这里，
// 跟 agent 的 skill 目录同在用户家目录下，避免跨磁盘 rename 失败。
func SubDir(name string) (string, error) {
	dir, err := SkillctlDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(p, 0700); err != nil {
		return "", fmt.Errorf("创建目录 %s 失败:%w", p, err)
	}
	return p, nil
}

// ConfigPath 返回 ~/.skillctl/config.json 的完整路径。
func ConfigPath() (string, error) {
	dir, err := SkillctlDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig 读取配置。文件不存在不算错误，返回空配置，代表"还没登入过"。
func LoadConfig() (*Config, error) {
	p, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("读取配置失败:%w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败:%w", err)
	}
	return &cfg, nil
}

// SaveConfig 写配置。目录 0700、文件 0600：里面有 token。
func SaveConfig(cfg *Config) error {
	p, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("创建配置目录失败:%w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("配置文件序列化失败:%w", err)
	}
	if err := os.WriteFile(p, data, 0600); err != nil {
		return fmt.Errorf("写入配置文件失败:%w", err)
	}
	return nil
}

// ResolveToken 按优先级取 token，取不到返回空串，由调用方决定怎么提示。
func ResolveToken() string {
	if t := os.Getenv("DR_ACCESS_TOKEN"); t != "" {
		return t
	}
	if t := os.Getenv("DR_SKILL_MARKET_TOKEN"); t != "" {
		return t
	}
	if FlagToken != "" {
		return FlagToken
	}
	cfg, err := LoadConfig()
	if err != nil {
		return ""
	}
	return cfg.Token
}

// JwtHeader 给请求加上 Authorization: Bearer <token>。
func JwtHeader(res *http.Request) error {
	token := ResolveToken()
	if token == "" {
		return fmt.Errorf("未登入或缺少 token:请先执行 skillctl auth login,或设置 DR_ACCESS_TOKEN 环境变量")
	}
	res.Header.Set("Authorization", "Bearer "+token)
	return nil
}
