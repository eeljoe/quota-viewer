package fetcher

import "testing"

func TestGetAll_ContainsElevenProviders_InStableOrder(t *testing.T) {
	all := GetAll()
	if len(all) != 11 {
		t.Fatalf("expected 11 providers, got %d", len(all))
	}
	want := []string{"kimi", "xfyun", "opencode-go", "mimo", "deepseek", "ollama", "command-code", "factory-droid", "minimax", "open-design", "devin"}
	for i, id := range want {
		if all[i].ID != id {
			t.Errorf("expected providers[%d].ID=%s, got %s", i, id, all[i].ID)
		}
	}
}

func TestGetAll_EachProvider_HasCompleteDefinition(t *testing.T) {
	for _, d := range GetAll() {
		if d.DisplayName == "" {
			t.Errorf("provider %s missing DisplayName", d.ID)
		}
		if d.Abbr == "" {
			t.Errorf("provider %s missing Abbr", d.ID)
		}
		if len(d.Fields) == 0 {
			t.Errorf("provider %s missing credential fields", d.ID)
		}
		for _, f := range d.Fields {
			if f.Key == "" || f.Label == "" || f.Type == "" {
				t.Errorf("provider %s has incomplete field %+v", d.ID, f)
			}
		}
		if d.Build == nil {
			t.Errorf("provider %s missing Build", d.ID)
		}
	}
}

func TestGetAll_Build_ReturnsFetcherForEach(t *testing.T) {
	// 隔离用户目录:任何 Provider 的空凭证 Build 都不应读取/访问真实用户凭证或网络。
	t.Setenv("USERPROFILE", t.TempDir())
	// APPDATA 也要隔离:Devin 留空时读本机 Devin Desktop 登录态(票 02),
	// 不隔离会读到开发机真实登录态并真的发起网络请求。
	t.Setenv("APPDATA", t.TempDir())
	for _, d := range GetAll() {
		// 使用空凭证：只验证 Build 可执行且 Fetch 不 panic，不访问真实网络。
		f := d.Build(map[string]string{})
		if f == nil {
			t.Errorf("provider %s Build returned nil", d.ID)
		}
		// 每个 Build 出来的 fetcher 必须可执行且不 panic
		r := f.Fetch()
		if r.Platform == "" {
			t.Errorf("provider %s Fetch returned empty platform", d.ID)
		}
	}
}

func TestGet_KnownAndUnknown(t *testing.T) {
	if _, ok := Get("kimi"); !ok {
		t.Error("expected Get('kimi') to succeed")
	}
	if _, ok := Get("unknown-provider"); ok {
		t.Error("expected Get('unknown-provider') to fail")
	}
}

// TestGet_Devin 第 11 家的定义契约(票 01 AS-10)。
func TestGet_Devin(t *testing.T) {
	d, ok := Get("devin")
	if !ok {
		t.Fatal("expected Get('devin') to succeed")
	}
	if d.DisplayName != "Devin" {
		t.Errorf("DisplayName: got %q, want Devin", d.DisplayName)
	}
	if d.Abbr != "DV" {
		t.Errorf("Abbr: got %q, want DV", d.Abbr)
	}
	if d.Kind != KindUsage {
		t.Errorf("Kind: got %q, want %q", d.Kind, KindUsage)
	}
	if d.LoginURL != "https://app.devin.ai" {
		t.Errorf("LoginURL: got %q, want https://app.devin.ai", d.LoginURL)
	}
	if len(d.Fields) != 1 {
		t.Fatalf("凭证字段数: got %d, want 恰好 1 个", len(d.Fields))
	}
	if d.Fields[0].Key != "api_key" || d.Fields[0].Type != "password" {
		t.Errorf("凭证字段: got %+v, want api_key/password", d.Fields[0])
	}
	if d.Fields[0].Label == "" {
		t.Error("凭证字段 Label 不得为空")
	}
	// Build 必须能造出 DevinFetcher。
	if f, ok := d.Build(map[string]string{"api_key": "x"}).(*DevinFetcher); !ok || f.apiKey != "x" {
		t.Errorf("Build 应返回带 apiKey 的 *DevinFetcher, got %T", d.Build(map[string]string{"api_key": "x"}))
	}
}
