package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"quota-viewer/internal/fetcher"
)

// writeConfig 把 JSON 写入测试 APPDATA 下的配置文件。
func writeConfig(t *testing.T, jsonStr string) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	dir := filepath.Join(tmpDir, "quota-viewer")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(jsonStr), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_FileNotExists_ReturnsDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.RefreshIntervalMin != 15 {
		t.Errorf("expected default RefreshIntervalMin=15, got %d", cfg.RefreshIntervalMin)
	}
	if cfg.BallX != -1 || cfg.BallY != -1 {
		t.Errorf("expected default BallX=-1, BallY=-1, got %d,%d", cfg.BallX, cfg.BallY)
	}
	if len(cfg.Providers) != len(AllProviderIDs) {
		t.Fatalf("expected %d providers, got %d", len(AllProviderIDs), len(cfg.Providers))
	}
	// 默认启用前三个,其余关闭
	for _, p := range cfg.Providers {
		wantEnabled := false
		for _, d := range DefaultProviderIDs {
			if p.ID == d {
				wantEnabled = true
				break
			}
		}
		if p.Enabled != wantEnabled {
			t.Errorf("provider %s: expected enabled=%v, got %v", p.ID, wantEnabled, p.Enabled)
		}
	}
}

func TestSaveThenLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	original := &Config{
		Providers: []ProviderConfig{
			{ID: "kimi", Enabled: true, Creds: map[string]string{"api_key": "k1"}},
			{ID: "xfyun", Enabled: false},
			{ID: "opencode-go", Enabled: true, Creds: map[string]string{"workspace_id": "w1", "session_token": "s1"}},
			{ID: "mimo", Enabled: false},
			{ID: "deepseek", Enabled: true, Creds: map[string]string{"api_key": "d1"}, Budget: 500.00},
			{ID: "ollama", Enabled: false, Creds: map[string]string{"cookie": "wos-session=o1"}},
			{ID: "command-code", Enabled: false},
			{ID: "factory-droid", Enabled: false},
			{ID: "minimax", Enabled: false},
		},
		RefreshIntervalMin: 30,
		BallX:              100,
		BallY:              200,
	}

	err := Save(original)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// 验证文件确实创建在正确路径
	expectedPath := filepath.Join(tmpDir, "quota-viewer", "config.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("config file not created at %s", expectedPath)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded.Providers) != len(original.Providers) {
		t.Fatalf("Providers length mismatch: got %d", len(loaded.Providers))
	}
	for i, p := range loaded.Providers {
		o := original.Providers[i]
		if p.ID != o.ID || p.Enabled != o.Enabled {
			t.Errorf("Providers[%d] mismatch: got %+v, want %+v", i, p, o)
		}
		for k, v := range o.Creds {
			if p.Creds[k] != v {
				t.Errorf("Providers[%d].Creds[%s] mismatch: got %s, want %s", i, k, p.Creds[k], v)
			}
		}
		if p.Budget != o.Budget {
			t.Errorf("Providers[%d].Budget mismatch: got %f, want %f", i, p.Budget, o.Budget)
		}
	}
	if loaded.RefreshIntervalMin != 30 {
		t.Errorf("RefreshIntervalMin mismatch: got %d", loaded.RefreshIntervalMin)
	}
	if loaded.BallX != 100 || loaded.BallY != 200 {
		t.Errorf("Ball position mismatch: got %d,%d", loaded.BallX, loaded.BallY)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	// 确保目录不存在
	dir := filepath.Join(tmpDir, "quota-viewer")
	os.RemoveAll(dir)

	cfg := &Config{Providers: Default().Providers}
	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() should create directory, got error: %v", err)
	}
}

func TestLoad_NewProvider_AppendedToExistingV2Config(t *testing.T) {
	writeConfig(t, `{
	  "providers": [
	    {"id":"kimi","enabled":true,"creds":{"api_key":"k"}},
	    {"id":"xfyun","enabled":false},
	    {"id":"opencode-go","enabled":true},
	    {"id":"mimo","enabled":false},
	    {"id":"deepseek","enabled":false}
	  ],
	  "refresh_interval_min": 15,
	  "ball_x": -1,
	  "ball_y": -1
	}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(cfg.Providers) != len(AllProviderIDs) {
		t.Fatalf("expected %d providers after migration, got %d", len(AllProviderIDs), len(cfg.Providers))
	}
	last := cfg.Providers[len(cfg.Providers)-1]
	if last.ID != "minimax" || last.Enabled {
		t.Errorf("expected disabled minimax appended, got %+v", last)
	}
	// ollama / command-code / factory-droid 也应被补全且默认关闭
	for _, id := range []string{"ollama", "command-code", "factory-droid"} {
		for _, p := range cfg.Providers {
			if p.ID == id && p.Enabled {
				t.Errorf("expected disabled %s appended, got %+v", id, p)
			}
		}
	}
	if cfg.Providers[0].Creds["api_key"] != "k" {
		t.Errorf("existing credentials should be preserved: %+v", cfg.Providers[0])
	}
}

// 旧版扁平格式(含 mimo_cookie 与 opencode_go 字段)迁移到 providers 结构。
func TestLoad_MigrateLegacyJSON(t *testing.T) {
	writeConfig(t, `{
	  "kimi_api_key": "k",
	  "xfyun_cookie": "x",
	  "mimo_cookie": "session=abc",
	  "opencode_go_workspace_id": "ws1",
	  "opencode_go_session_token": "tok1",
	  "refresh_interval_min": 5,
	  "ball_x": 10,
	  "ball_y": 20
	}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// 有凭证的 provider 都 enabled,凭证迁移正确
	byID := map[string]ProviderConfig{}
	for _, p := range cfg.Providers {
		byID[p.ID] = p
	}

	if !byID["kimi"].Enabled || byID["kimi"].Creds["api_key"] != "k" {
		t.Errorf("kimi migrate wrong: %+v", byID["kimi"])
	}
	if !byID["xfyun"].Enabled || byID["xfyun"].Creds["cookie"] != "x" {
		t.Errorf("xfyun migrate wrong: %+v", byID["xfyun"])
	}
	// mimo 有凭证但被钳制(4 个 enabled 超上限,按注册表顺序保留前 3):
	// 凭证保留,enabled 为 false,用户可在设置中重新启用
	if byID["mimo"].Enabled {
		t.Errorf("mimo should be clamped to disabled (4 enabled > 3): %+v", byID["mimo"])
	}
	if byID["mimo"].Creds["cookie"] != "session=abc" {
		t.Errorf("mimo creds should survive clamping: %+v", byID["mimo"])
	}
	oc := byID["opencode-go"]
	if !oc.Enabled || oc.Creds["workspace_id"] != "ws1" || oc.Creds["session_token"] != "tok1" {
		t.Errorf("opencode-go migrate wrong: %+v", oc)
	}
	// deepseek 无旧字段 → 不启用
	if byID["deepseek"].Enabled {
		t.Errorf("deepseek should not be enabled after migration: %+v", byID["deepseek"])
	}
	// 通用项保留
	if cfg.RefreshIntervalMin != 5 {
		t.Errorf("RefreshIntervalMin mismatch: got %d", cfg.RefreshIntervalMin)
	}
	if cfg.BallX != 10 || cfg.BallY != 20 {
		t.Errorf("Ball position mismatch: got %d,%d", cfg.BallX, cfg.BallY)
	}

	// 迁移后文件已回写为新格式
	tmpDir := os.Getenv("APPDATA")
	newPath := filepath.Join(tmpDir, "quota-viewer", "config.json")
	data, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("migrated config not written back: %v", err)
	}
	var probe struct {
		Providers []ProviderConfig `json:"providers"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		t.Fatalf("migrated config not valid JSON: %v", err)
	}
	if len(probe.Providers) == 0 {
		t.Error("migrated config missing providers key")
	}
}

// 旧格式全部字段为空 → 默认启用前三个。
func TestLoad_MigrateLegacyEmpty_ReturnsDefaults(t *testing.T) {
	writeConfig(t, `{
	  "kimi_api_key": "",
	  "xfyun_cookie": "",
	  "refresh_interval_min": 5,
	  "ball_x": -1,
	  "ball_y": -1
	}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.RefreshIntervalMin != 5 {
		t.Errorf("RefreshIntervalMin mismatch: got %d", cfg.RefreshIntervalMin)
	}
	enabledCount := 0
	for _, p := range cfg.Providers {
		if p.Enabled {
			enabledCount++
		}
	}
	if enabledCount != 3 {
		t.Errorf("expected 3 default enabled providers, got %d", enabledCount)
	}
}

// TestAllProviderIDs_MatchesFetcherRegistry 守护 config.AllProviderIDs 与 fetcher
// 注册表同步。两份清单漂移时:Default()/ensureKnownProviders 漏掉新 Provider,
// 全新安装与旧配置升级路径都不会带上它(2026-09-30 接入 factory-droid 时实际踩坑)。
func TestAllProviderIDs_MatchesFetcherRegistry(t *testing.T) {
	all := fetcher.GetAll()
	if len(AllProviderIDs) != len(all) {
		t.Fatalf("AllProviderIDs has %d entries but registry has %d", len(AllProviderIDs), len(all))
	}
	for i, def := range all {
		if AllProviderIDs[i] != def.ID {
			t.Errorf("AllProviderIDs[%d]=%s, registry[%d]=%s (order must match)",
				i, AllProviderIDs[i], i, def.ID)
		}
	}
}

// TestLoad_ClampsEnabledOverThree 复现 2026-09-30 现场:手工编辑配置造成 4 个同时
// 启用(绕过 SaveConfig 的钳制)。Load 必须按顺序钳制回 3 个,否则配置面板会渲染出
// 「勾了 4 个」的不可能状态,保存时被静默钳掉一个,用户反复保存才收敛。
func TestLoad_ClampsEnabledOverThree(t *testing.T) {
	writeConfig(t, `{
	  "providers": [
	    {"id":"kimi","enabled":true},
	    {"id":"xfyun","enabled":false},
	    {"id":"opencode-go","enabled":false},
	    {"id":"mimo","enabled":false},
	    {"id":"deepseek","enabled":false},
	    {"id":"ollama","enabled":true},
	    {"id":"command-code","enabled":true},
	    {"id":"factory-droid","enabled":true,"creds":{"api_key":"fk_keep"}}
	  ],
	  "refresh_interval_min": 15,
	  "ball_x": -1,
	  "ball_y": -1
	}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	enabledCount := 0
	byID := map[string]ProviderConfig{}
	for _, p := range cfg.Providers {
		byID[p.ID] = p
		if p.Enabled {
			enabledCount++
		}
	}
	if enabledCount != 3 {
		t.Errorf("expected exactly 3 enabled after clamp, got %d", enabledCount)
	}
	// 注册表顺序在前的保留启用,最后的 factory-droid 被钳制关闭
	if !byID["kimi"].Enabled || !byID["ollama"].Enabled || !byID["command-code"].Enabled {
		t.Error("expected first three enabled providers to survive clamp")
	}
	if byID["factory-droid"].Enabled {
		t.Error("expected factory-droid clamped to disabled (4th enabled)")
	}
	// 钳制只关启用,不丢凭证
	if byID["factory-droid"].Creds["api_key"] != "fk_keep" {
		t.Errorf("clamping must preserve credentials, got %+v", byID["factory-droid"])
	}
}

// TestLoad_ExtendedMode_AllowsNineEnabled 验证扩展模式把启用上限从 3 放开到 9:
// extended_mode=true 时全部 Provider 可同时启用;普通模式仍然钳回 3。
func TestLoad_ExtendedMode_AllowsNineEnabled(t *testing.T) {
	buildJSON := func(extended bool) string {
		ids := make([]string, 0, len(AllProviderIDs))
		for _, id := range AllProviderIDs {
			ids = append(ids, fmt.Sprintf(`{"id":"%s","enabled":true}`, id))
		}
		return fmt.Sprintf(`{
		  "providers": [%s],
		  "extended_mode": %t,
		  "refresh_interval_min": 15,
		  "ball_x": -1,
		  "ball_y": -1
		}`, strings.Join(ids, ","), extended)
	}

	writeConfig(t, buildJSON(true))
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	enabled := 0
	for _, p := range cfg.Providers {
		if p.Enabled {
			enabled++
		}
	}
	if enabled != len(AllProviderIDs) {
		t.Errorf("extended mode: expected all %d enabled, got %d", len(AllProviderIDs), enabled)
	}

	// 同样的启用状态,关闭扩展模式 → 钳回 3
	writeConfig(t, buildJSON(false))
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	enabled = 0
	for _, p := range cfg.Providers {
		if p.Enabled {
			enabled++
		}
	}
	if enabled != 3 {
		t.Errorf("normal mode: expected 3 enabled after clamp, got %d", enabled)
	}
}
