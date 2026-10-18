package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 是应用配置:动态 Provider 列表(顺序 = 展示顺序)+ 通用项。
type Config struct {
	Providers          []ProviderConfig `json:"providers"`
	RefreshIntervalMin int              `json:"refresh_interval_min"`
	BallX              int              `json:"ball_x"`
	BallY              int              `json:"ball_y"`
	ExtendedMode       bool             `json:"extended_mode,omitempty"` // 扩展模式:悬浮球最多 9 格(默认关闭,3 格)
}

// ProviderConfig 描述单个 Provider 的启用状态与凭证。
type ProviderConfig struct {
	ID      string            `json:"id"`
	Enabled bool              `json:"enabled"`
	Creds   map[string]string `json:"creds,omitempty"`  // 主凭证(第 1 组)
	Budget  float64           `json:"budget,omitempty"` // 余额型 Provider 的预算总量(0 = 未设)
	Extra   []CredGroup       `json:"extra,omitempty"`  // 第 2..N 组凭证(同平台多 Key / 多账号)
}

// CredGroup 是一组额外的凭证。
//
// Key 是稳定标识(前端生成、保存时原样带回):保存时按它对齐到已有组,
// 避免"删掉中间一行"导致后面各行的凭证值整体错位。空串的组由后端补一个。
type CredGroup struct {
	Key   string            `json:"key,omitempty"`
	Creds map[string]string `json:"creds,omitempty"`
}

// CredSets 返回该 Provider 的全部凭证组:主凭证在前,附加组在后。
//
// 全空的附加组会被跳过——用户点了"添加一行"却没填内容时,不该凭空多出一个
// 报"未配置 Key"的监控对象。主凭证恒返回(可能为空 map),保持单凭证 Provider
// 既有语义:启用了但没配凭证 = 一格错误提示。
func (p ProviderConfig) CredSets() []map[string]string {
	sets := make([]map[string]string, 0, 1+len(p.Extra))
	sets = append(sets, p.Creds)
	for _, g := range p.Extra {
		if CredGroupEmpty(g.Creds) {
			continue
		}
		sets = append(sets, g.Creds)
	}
	return sets
}

// CredGroupEmpty 判断一组凭证是否完全没有内容(用户加了行但没填)。
func CredGroupEmpty(creds map[string]string) bool {
	for _, v := range creds {
		if v != "" {
			return false
		}
	}
	return true
}

// AllProviderIDs 全部已知 Provider id(与 fetcher 注册表一致,顺序 = 展示顺序)。
// 同步由 config_test.go 的 TestAllProviderIDs_MatchesFetcherRegistry 守护。
var AllProviderIDs = []string{"kimi", "xfyun", "opencode-go", "mimo", "deepseek", "ollama", "command-code", "factory-droid", "minimax", "open-design"}

// DefaultProviderIDs 默认启用的 Provider(与现状一致:Kimi/讯飞/OpenCode Go)。
var DefaultProviderIDs = []string{"kimi", "xfyun", "opencode-go"}

// 悬浮球展示上限:按**监控对象**计,不再按 Provider 计。
//
// 一个监控对象 = 一组凭证 × 一张套餐卡。60x60 的悬浮球里:
//   - 普通模式 9 个 = 3x3 网格,每格约 20px(沿用 v1.3.0 扩展模式的观感)
//   - 扩展模式 16 个 = 4x4 网格,每格约 15px
//
// 再往上(5x5 每格 12px)CJK 缩写与两字缩写就糊成一团,故 16 是硬上限。
const (
	maxObjectsNormal   = 9
	maxObjectsExtended = 16
)

// MaxObjects 返回悬浮球可显示的监控对象数上限(扩展模式放开到 16)。
func MaxObjects(extended bool) int {
	if extended {
		return maxObjectsExtended
	}
	return maxObjectsNormal
}

// CountObjects 统计一组 Provider 配置当前占用的监控对象数(各启用 Provider 的凭证组数之和)。
func CountObjects(providers []ProviderConfig) int {
	n := 0
	for _, p := range providers {
		if p.Enabled {
			n += len(p.CredSets())
		}
	}
	return n
}

// clampEnabled 把启用的监控对象数钳制到上限(超出的 Provider 按顺序关闭)。
// 按对象数而非 Provider 数计:一个 Provider 挂 3 组凭证就占 3 格。
func clampEnabled(providers []ProviderConfig, limit int) {
	used := 0
	for i := range providers {
		if !providers[i].Enabled {
			continue
		}
		cost := len(providers[i].CredSets())
		if used+cost > limit {
			providers[i].Enabled = false
			continue
		}
		used += cost
	}
}

// legacyConfig 旧版扁平配置结构(仅用于 Load 时迁移)。
type legacyConfig struct {
	KimiAPIKey         string `json:"kimi_api_key"`
	XfyunCookie        string `json:"xfyun_cookie"`
	MimoCookie         string `json:"mimo_cookie"`
	RefreshIntervalMin int    `json:"refresh_interval_min"`
	BallX              int    `json:"ball_x"`
	BallY              int    `json:"ball_y"`
}

func configDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", os.ErrNotExist
	}
	dir := filepath.Join(appData, "quota-viewer")
	return dir, nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Default 返回带默认值的配置:全部 Provider 注册,默认启用前三个。
func Default() *Config {
	cfg := &Config{
		Providers:          make([]ProviderConfig, 0, len(AllProviderIDs)),
		RefreshIntervalMin: 15,
		BallX:              -1,
		BallY:              -1,
	}
	for _, id := range AllProviderIDs {
		enabled := false
		for _, d := range DefaultProviderIDs {
			if d == id {
				enabled = true
				break
			}
		}
		cfg.Providers = append(cfg.Providers, ProviderConfig{ID: id, Enabled: enabled})
	}
	return cfg
}

// Load 读取配置文件。文件不存在时返回带默认值的空配置(不报错)。
// 旧版扁平字段格式(config v1)会自动迁移为 providers 结构并回写。
func Load() (*Config, error) {
	cfg := Default()

	path, err := configPath()
	if err != nil {
		return cfg, nil // APPDATA 不存在,返回默认配置
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // 文件不存在不算错误
		}
		return cfg, err
	}

	// 探测是新格式(providers 键)还是旧格式(扁平字段)
	var probe struct {
		Providers []ProviderConfig `json:"providers"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return cfg, err
	}

	if len(probe.Providers) > 0 {
		// 新格式
		if err := json.Unmarshal(data, cfg); err != nil {
			return cfg, err
		}
	} else {
		// 旧格式:迁移
		var legacy legacyConfig
		if err := json.Unmarshal(data, &legacy); err != nil {
			return cfg, err
		}
		cfg = migrateFromLegacy(legacy)
	}

	// 确保默认值
	if cfg.RefreshIntervalMin <= 0 {
		cfg.RefreshIntervalMin = 15
	}
	// 确保 providers 非空(防御:损坏文件),并为已有 v2 配置补上新 Provider。
	if len(cfg.Providers) == 0 {
		cfg.Providers = Default().Providers
	}
	if ensureKnownProviders(cfg) {
		_ = Save(cfg) // 新 Provider 迁移回写失败不阻塞启动
	}

	// 钳制监控对象上限(普通 9 / 扩展 16):手工编辑配置/异常状态可能超过,
	// 不钳制会让配置面板渲染出不可能的勾选状态,保存时被静默钳掉。
	clampEnabled(cfg.Providers, MaxObjects(cfg.ExtendedMode))

	return cfg, nil
}

// ensureKnownProviders 给已有动态配置追加新注册的 Provider,默认关闭且不影响现有选择。
func ensureKnownProviders(cfg *Config) bool {
	seen := make(map[string]bool, len(cfg.Providers))
	for _, p := range cfg.Providers {
		seen[p.ID] = true
	}

	changed := false
	for _, id := range AllProviderIDs {
		if seen[id] {
			continue
		}
		cfg.Providers = append(cfg.Providers, ProviderConfig{ID: id})
		changed = true
	}
	return changed
}

// migrateFromLegacy 把旧版扁平配置迁移为 providers 结构。
// 有值的旧字段 → 对应 Provider enabled + 凭证迁移;全部为空 → 默认启用前三个。
// 迁移后立即回写新格式(失败静默,下次 Load 会再尝试)。
func migrateFromLegacy(l legacyConfig) *Config {
	cfg := Default()
	cfg.RefreshIntervalMin = l.RefreshIntervalMin
	cfg.BallX = l.BallX
	cfg.BallY = l.BallY

	set := func(id string, has bool, creds map[string]string) {
		if !has {
			return
		}
		for i := range cfg.Providers {
			if cfg.Providers[i].ID == id {
				cfg.Providers[i].Enabled = true
				cfg.Providers[i].Creds = creds
				return
			}
		}
	}

	set("kimi", l.KimiAPIKey != "", map[string]string{"api_key": l.KimiAPIKey})
	set("xfyun", l.XfyunCookie != "", map[string]string{"cookie": l.XfyunCookie})
	set("mimo", l.MimoCookie != "", map[string]string{"cookie": l.MimoCookie})

	// opencode-go 无旧扁平字段可迁移:它改用 API Key 通道后,旧版的
	// opencode_go_workspace_id / opencode_go_session_token 已无人读取(死 key),
	// 故不在此处 set —— 该家保持默认启用状态、creds 为空。

	// 钳制监控对象上限(旧格式无扩展模式,按普通模式 9 个对象)
	clampEnabled(cfg.Providers, MaxObjects(false))

	_ = Save(cfg) // 回写新格式,失败静默
	return cfg
}

// Save 写入配置文件。目录不存在时自动创建。
func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
