package fetcher

import "testing"

func TestGetAll_ContainsFiveProviders_InStableOrder(t *testing.T) {
	all := GetAll()
	if len(all) != 5 {
		t.Fatalf("expected 5 providers, got %d", len(all))
	}
	want := []string{"kimi", "xfyun", "opencode-go", "mimo", "deepseek"}
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
	for _, d := range GetAll() {
		creds := make(map[string]string)
		for _, f := range d.Fields {
			creds[f.Key] = "dummy-" + f.Key
		}
		f := d.Build(creds)
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
