// 本地安装账本 ~/.skillctl/registry.json。
//
// 结构 skills[slug][target]：一个 skill 可以同时装给多个 agent，
// 每个 agent 只保留一份记录。用两层 map 而不是数组，天然保证唯一，
// 也不用写循环去查重。Files 记的是"这个包里有哪些文件"，
// 供下次升级时算出"上一版有、新版没有"该删掉的文件，不是目录里所有文件。
package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go_mod/security"
)

// InstallRecord 一条安装记录 = 一个 (slug, target) 组合。
type InstallRecord struct {
	Version     string   `json:"version"`
	UUID        string   `json:"uuid,omitempty"`
	Path        string   `json:"path"`
	Files       []string `json:"files"`
	InstalledAt string   `json:"installedAt"`
}

// Registry 对应 ~/.skillctl/registry.json。
type Registry struct {
	SchemaVersion int                                 `json:"schemaVersion"`
	Skills        map[string]map[string]InstallRecord `json:"skills"`
}

// RegistryPath 返回账本路径。
func RegistryPath() (string, error) {
	dir, err := security.SkillctlDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "registry.json"), nil
}

// LoadRegistry 读账本。文件不存在返回空账本，不报错。
func LoadRegistry() (*Registry, error) {
	p, err := RegistryPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{SchemaVersion: 1, Skills: map[string]map[string]InstallRecord{}}, nil
		}
		return nil, fmt.Errorf("读取账本失败:%w", err)
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("账本格式错误(%s):%w", p, err)
	}
	if reg.Skills == nil {
		reg.Skills = map[string]map[string]InstallRecord{}
	}
	if reg.SchemaVersion == 0 {
		reg.SchemaVersion = 1
	}
	return &reg, nil
}

// SaveRegistry 写账本，0600：里面含安装路径和 uuid。
func SaveRegistry(reg *Registry) error {
	p, err := RegistryPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("创建账本目录失败:%w", err)
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("账本序列化失败:%w", err)
	}
	if err := os.WriteFile(p, data, 0600); err != nil {
		return fmt.Errorf("写账本失败:%w", err)
	}
	return nil
}

// Get 取某个 (slug, target) 的记录。
func (r *Registry) Get(slug, target string) (InstallRecord, bool) {
	byTarget, ok := r.Skills[slug]
	if !ok {
		return InstallRecord{}, false
	}
	rec, ok := byTarget[target]
	return rec, ok
}

// Put 写入一条记录，同时补上安装时间。
func (r *Registry) Put(slug, target string, rec InstallRecord) {
	if r.Skills == nil {
		r.Skills = map[string]map[string]InstallRecord{}
	}
	if r.Skills[slug] == nil {
		r.Skills[slug] = map[string]InstallRecord{}
	}
	rec.InstalledAt = time.Now().Format(time.RFC3339)
	r.Skills[slug][target] = rec
}

// Delete 删除一条记录，返回是否真的存在过。
func (r *Registry) Delete(slug, target string) bool {
	byTarget, ok := r.Skills[slug]
	if !ok {
		return false
	}
	if _, ok := byTarget[target]; !ok {
		return false
	}
	delete(byTarget, target)
	if len(byTarget) == 0 {
		delete(r.Skills, slug)
	}
	return true
}
