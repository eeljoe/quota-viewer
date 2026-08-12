package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
	if last.ID != "ollama" || last.Enabled {
		t.Errorf("expected disabled ollama appended, got %+v", last)
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
