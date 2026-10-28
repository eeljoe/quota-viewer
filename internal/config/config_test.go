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
			{ID: "opencode-go", Enabled: true, Creds: map[string]string{"api_key": "oc_sk_test"}},
			{ID: "mimo", Enabled: false},
			{ID: "deepseek", Enabled: true, Creds: map[string]string{"api_key": "d1"}, Budget: 500.00},
			{ID: "ollama", Enabled: false, Creds: map[string]string{"cookie": "wos-session=o1"}},
			{ID: "command-code", Enabled: false},
			{ID: "factory-droid", Enabled: false},
			{ID: "minimax", Enabled: false},
			{ID: "open-design", Enabled: false, Creds: map[string]string{"cookie": "session=od1"}},
			{ID: "devin", Enabled: false, Creds: map[string]string{"api_key": "dv1"}},
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
	if last.ID != "devin" || last.Enabled {
		t.Errorf("expected disabled devin appended, got %+v", last)
	}
	// minimax / ollama / command-code / factory-droid / open-design 也应被补全且默认关闭
	for _, id := range []string{"ollama", "command-code", "factory-droid", "minimax", "open-design"} {
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

// 旧版扁平格式(含 mimo_cookie 与已废弃的 opencode_go_* 死 key)迁移到 providers 结构。
// opencode_go_workspace_id/session_token 已无人读取:opencode-go 迁移后 enabled 状态照旧
// (它本来就是默认启用项)、creds 为空。
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
	// mimo 有凭证且只占 1 个监控对象(上限 9)⇒ 迁移后保持启用。
	// 上限按对象数计后,旧版「4 个启用超 3」的钳制场景不再触发;
	// 钳制不丢凭证这条由 TestLoad_ClampsObjectsOverNine 继续守护。
	if !byID["mimo"].Enabled {
		t.Errorf("mimo should stay enabled (1 object, limit 9): %+v", byID["mimo"])
	}
	if byID["mimo"].Creds["cookie"] != "session=abc" {
		t.Errorf("mimo creds should migrate: %+v", byID["mimo"])
	}
	// 死 key 不再被读取:opencode-go 保持默认启用,但 creds 必须为空
	oc := byID["opencode-go"]
	if !oc.Enabled {
		t.Errorf("opencode-go should stay enabled (default provider), got %+v", oc)
	}
	if len(oc.Creds) != 0 {
		t.Errorf("opencode-go creds must be empty (dead legacy keys ignored), got %+v", oc.Creds)
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

// 旧配置里只有那两个死 key 有值(其他扁平字段全空):不该因它们把 opencode-go 当成
// "有凭证"、更不该写进 creds —— 结果与"旧格式全空"一致(默认启用前三个,creds 为空)。
func TestLoad_MigrateLegacyOnlyDeadKeys_WritesNoCreds(t *testing.T) {
	writeConfig(t, `{
	  "opencode_go_workspace_id": "ws1",
	  "opencode_go_session_token": "tok1",
	  "refresh_interval_min": 15,
	  "ball_x": -1,
	  "ball_y": -1
	}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	byID := map[string]ProviderConfig{}
	for _, p := range cfg.Providers {
		byID[p.ID] = p
	}
	oc := byID["opencode-go"]
	if !oc.Enabled {
		t.Errorf("opencode-go should stay enabled as a default provider, got %+v", oc)
	}
	if len(oc.Creds) != 0 {
		t.Errorf("opencode-go creds must be empty (dead legacy keys ignored), got %+v", oc.Creds)
	}
	if !byID["kimi"].Enabled || !byID["xfyun"].Enabled {
		t.Errorf("expected default enabled providers to stay enabled: kimi=%v xfyun=%v",
			byID["kimi"].Enabled, byID["xfyun"].Enabled)
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

// TestLoad_ClampsObjectsOverNine 复现手工编辑配置绕过 SaveConfig 钳制的现场。
// 上限按**监控对象**计(普通模式 9):11 家各 1 组凭证 = 11 个对象,Load 必须按顺序
// 钳回 9 个,否则配置面板会渲染出不可能的状态,保存时被静默钳掉。
// 第 10 家 open-design 与第 11 家 devin 会**同时**被钳制关闭(clampEnabled 是
// used+cost > limit 就 continue,超出后不会"跳过前一个再试下一个")。
func TestLoad_ClampsObjectsOverNine(t *testing.T) {
	ids := make([]string, 0, len(AllProviderIDs))
	for i, id := range AllProviderIDs {
		extra := ""
		if i == len(AllProviderIDs)-1 {
			extra = `,"creds":{"api_key":"fk_keep"}`
		}
		ids = append(ids, fmt.Sprintf(`{"id":"%s","enabled":true%s}`, id, extra))
	}
	writeConfig(t, fmt.Sprintf(`{
	  "providers": [%s],
	  "refresh_interval_min": 15,
	  "ball_x": -1,
	  "ball_y": -1
	}`, strings.Join(ids, ",")))

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
	if enabledCount != 9 {
		t.Errorf("expected exactly 9 enabled after clamp, got %d", enabledCount)
	}
	// 注册表顺序在前的保留启用
	if !byID["kimi"].Enabled || !byID["ollama"].Enabled || !byID["command-code"].Enabled {
		t.Error("expected first three enabled providers to survive clamp")
	}
	// 第 10 家 open-design 与第 11 家 devin 都被钳制关闭(11 个对象只装得下前 9 家)
	if byID["open-design"].Enabled {
		t.Error("expected open-design clamped to disabled (10th object)")
	}
	if byID["devin"].Enabled {
		t.Error("expected devin clamped to disabled (11th object)")
	}
	// 钳制只关启用,不丢凭证
	if byID["devin"].Creds["api_key"] != "fk_keep" {
		t.Errorf("clamping must preserve credentials, got %+v", byID["devin"])
	}
}

// TestLoad_ClampsByCredentialGroupCount 上限按凭证组数计:一家挂 3 组凭证就占 3 格。
// 4 家 × 3 组 = 12 个对象 > 9 ⇒ 第 4 家整家关闭(不切半组,避免留下无主的对象)。
func TestLoad_ClampsByCredentialGroupCount(t *testing.T) {
	three := `"creds":{"api_key":"k0"},` +
		`"extra":[{"key":"a","creds":{"api_key":"k1"}},{"key":"b","creds":{"api_key":"k2"}}]`
	writeConfig(t, `{
	  "providers": [
	    {"id":"kimi","enabled":true,`+three+`},
	    {"id":"xfyun","enabled":true,`+three+`},
	    {"id":"opencode-go","enabled":true,`+three+`},
	    {"id":"mimo","enabled":true,`+three+`}
	  ],
	  "refresh_interval_min": 15,
	  "ball_x": -1,
	  "ball_y": -1
	}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	byID := map[string]ProviderConfig{}
	for _, p := range cfg.Providers {
		byID[p.ID] = p
	}
	for _, id := range []string{"kimi", "xfyun", "opencode-go"} {
		if !byID[id].Enabled {
			t.Errorf("%s 应保留启用(9 个对象刚好装下前三家)", id)
		}
		if got := len(byID[id].CredSets()); got != 3 {
			t.Errorf("%s 凭证组数: got %d, want 3", id, got)
		}
	}
	if byID["mimo"].Enabled {
		t.Error("mimo(第 4 家,12 个对象超 9)应被整家钳制关闭")
	}
	// 关闭的那家凭证与凭证组都要留着,重新勾选即恢复
	if len(byID["mimo"].Extra) != 2 || byID["mimo"].Creds["api_key"] != "k0" {
		t.Errorf("钳制不得丢凭证组: %+v", byID["mimo"])
	}
}

// TestLoad_ExtendedMode_AllowsSixteenObjects 扩展模式把对象上限从 9 放开到 16:
// 6 家 × 3 组 = 18 个对象 ⇒ 前 5 家(15 个)保留,第 6 家关闭。
func TestLoad_ExtendedMode_AllowsSixteenObjects(t *testing.T) {
	three := `"creds":{"api_key":"k0"},` +
		`"extra":[{"key":"a","creds":{"api_key":"k1"}},{"key":"b","creds":{"api_key":"k2"}}]`
	buildJSON := func(extended bool) string {
		entries := make([]string, 0, 6)
		for _, id := range []string{"kimi", "xfyun", "opencode-go", "mimo", "deepseek", "ollama"} {
			entries = append(entries, fmt.Sprintf(`{"id":"%s","enabled":true,%s}`, id, three))
		}
		return fmt.Sprintf(`{
		  "providers": [%s],
		  "extended_mode": %t,
		  "refresh_interval_min": 15,
		  "ball_x": -1,
		  "ball_y": -1
		}`, strings.Join(entries, ","), extended)
	}

	writeConfig(t, buildJSON(true))
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	byID := map[string]ProviderConfig{}
	for _, p := range cfg.Providers {
		byID[p.ID] = p
	}
	for _, id := range []string{"kimi", "xfyun", "opencode-go", "mimo", "deepseek"} {
		if !byID[id].Enabled {
			t.Errorf("扩展模式: %s 应保留启用(15 个对象 ≤ 16)", id)
		}
	}
	if byID["ollama"].Enabled {
		t.Error("扩展模式: 第 6 家(18 个对象)应被钳制关闭")
	}

	// 同样的配置关掉扩展模式 ⇒ 上限 9,只装得下 3 家
	writeConfig(t, buildJSON(false))
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got := CountObjects(cfg.Providers); got != 9 {
		t.Errorf("normal mode 对象数: got %d, want 9", got)
	}
}

// TestCredSets_SkipsEmptyGroups 加了行没填内容的凭证组不该凭空多出监控对象
// (否则会多一格"未配置 Key"的红色报错格)。
func TestCredSets_SkipsEmptyGroups(t *testing.T) {
	p := ProviderConfig{
		Creds: map[string]string{"api_key": "k0"},
		Extra: []CredGroup{
			{Key: "a", Creds: map[string]string{"api_key": ""}},
			{Key: "b", Creds: map[string]string{"api_key": "k1"}},
			{Key: "c", Creds: nil},
		},
	}
	sets := p.CredSets()
	if len(sets) != 2 {
		t.Fatalf("凭证组数: got %d, want 2(空组应被跳过)", len(sets))
	}
	if sets[0]["api_key"] != "k0" || sets[1]["api_key"] != "k1" {
		t.Errorf("凭证组内容不对: %+v", sets)
	}
}

// TestLoad_ExtraCredGroupsRoundTrip 附加凭证组要能原样往返(不被迁移逻辑吃掉)。
func TestLoad_ExtraCredGroupsRoundTrip(t *testing.T) {
	writeConfig(t, `{
	  "providers": [
	    {"id":"kimi","enabled":true,"creds":{"api_key":"k0"},
	     "extra":[{"key":"kimi#1","creds":{"api_key":"k1"}},{"key":"kimi#2","creds":{"api_key":"k2"}}]},
	    {"id":"xfyun","enabled":false}
	  ],
	  "refresh_interval_min": 15,
	  "ball_x": -1,
	  "ball_y": -1
	}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	var kimi ProviderConfig
	for _, p := range cfg.Providers {
		if p.ID == "kimi" {
			kimi = p
		}
	}
	if len(kimi.Extra) != 2 {
		t.Fatalf("附加凭证组: got %d, want 2", len(kimi.Extra))
	}
	if kimi.Extra[1].Key != "kimi#2" || kimi.Extra[1].Creds["api_key"] != "k2" {
		t.Errorf("附加凭证组内容不对: %+v", kimi.Extra)
	}
	if got := len(kimi.CredSets()); got != 3 {
		t.Errorf("CredSets(): got %d, want 3", got)
	}
}
