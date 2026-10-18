package fetcher

import "testing"

// TestSettleIdentity_SingleObject 单对象时清空后缀、沿用 Provider 缩写 —— 与旧观感一致。
func TestSettleIdentity_SingleObject(t *testing.T) {
	def := ProviderDef{ID: "kimi", Abbr: "K"}
	r := QuotaResult{SubLabel: "不该出现的后缀"}
	SettleIdentity(&r, def, 0, 0, 1, 1)

	if r.SubLabel != "" {
		t.Errorf("单对象不应带后缀: got %q", r.SubLabel)
	}
	if r.Abbr != "K" {
		t.Errorf("单对象应沿用 Provider 缩写: got %q", r.Abbr)
	}
	if r.Slot != "kimi/0" {
		t.Errorf("slot: got %q, want kimi/0", r.Slot)
	}
}

// TestSettleIdentity_MultiPlan 多套餐:主卡保留 Provider 缩写,套餐序号进 slot。
func TestSettleIdentity_MultiPlan(t *testing.T) {
	def := ProviderDef{ID: "minimax", Abbr: "MX"}

	head := QuotaResult{}
	SettleIdentity(&head, def, 0, 0, 2, 1)
	if head.Abbr != "MX" || head.Slot != "minimax/0/p0" {
		t.Errorf("主套餐: abbr=%q slot=%q", head.Abbr, head.Slot)
	}

	tail := QuotaResult{SubLabel: "视频赠送卡", Abbr: "视"}
	SettleIdentity(&tail, def, 0, 1, 2, 1)
	if tail.Abbr != "视" {
		t.Errorf("套餐卡缩写应沿用 fetcher 派生值: got %q", tail.Abbr)
	}
	if tail.SubLabel != "视频赠送卡" {
		t.Errorf("单 Key 时后缀就是套餐名: got %q", tail.SubLabel)
	}
	if tail.Slot != "minimax/0/p1" {
		t.Errorf("slot: got %q, want minimax/0/p1", tail.Slot)
	}
}

// TestSettleIdentity_MultiKey 多 Key:第 2..N 把用「首字母 + 序号」区分,后缀标 Key 号。
func TestSettleIdentity_MultiKey(t *testing.T) {
	def := ProviderDef{ID: "kimi", Abbr: "K"}

	first := QuotaResult{}
	SettleIdentity(&first, def, 0, 0, 1, 3)
	if first.Abbr != "K" || first.SubLabel != "Key 1" {
		t.Errorf("第 1 把 Key: abbr=%q sub=%q", first.Abbr, first.SubLabel)
	}

	second := QuotaResult{}
	SettleIdentity(&second, def, 1, 0, 1, 3)
	if second.Abbr != "K2" {
		t.Errorf("第 2 把 Key 缩写: got %q, want K2", second.Abbr)
	}
	if second.SubLabel != "Key 2" || second.Slot != "kimi/1" {
		t.Errorf("第 2 把 Key: sub=%q slot=%q", second.SubLabel, second.Slot)
	}
}

// TestSettleIdentity_MultiKeyMultiPlan 两个维度同时存在时,后缀要同时说清是哪把 Key 的哪张卡。
func TestSettleIdentity_MultiKeyMultiPlan(t *testing.T) {
	def := ProviderDef{ID: "minimax", Abbr: "MX"}

	r := QuotaResult{SubLabel: "视频赠送卡", Abbr: "视"}
	SettleIdentity(&r, def, 1, 1, 2, 2)
	if r.SubLabel != "Key 2 · 视频赠送卡" {
		t.Errorf("后缀: got %q, want \"Key 2 · 视频赠送卡\"", r.SubLabel)
	}
	if r.Slot != "minimax/1/p1" {
		t.Errorf("slot: got %q, want minimax/1/p1", r.Slot)
	}
}

// TestFetchAll_FallsBackToSingle 未实现 MultiFetcher 的 Provider 仍是单元素切片。
func TestFetchAll_FallsBackToSingle(t *testing.T) {
	def := ProviderDef{
		ID:   "kimi",
		Abbr: "K",
		Build: func(map[string]string) Fetcher {
			return plainFetcher{QuotaResult{Platform: "Kimi"}}
		},
	}
	rs := FetchAll(def.Build(map[string]string{"api_key": "k"}))
	if len(rs) != 1 {
		t.Fatalf("单对象 Provider 应返回 1 条: got %d", len(rs))
	}
}

// plainFetcher 只实现 Fetcher(不实现 MultiFetcher),模拟 8/10 家未做多套餐的 Provider。
type plainFetcher struct{ r QuotaResult }

func (f plainFetcher) Fetch() QuotaResult { return f.r }

// TestFetchAll_EmptySliceFallsBack MultiFetcher 违约返回空切片时退回 Fetch(),
// 避免上层拿到「零监控对象」而静默不显示这一家。
func TestFetchAll_EmptySliceFallsBack(t *testing.T) {
	got := FetchAll(emptyMultiFetcher{})
	if len(got) != 1 || got[0].Error == "" {
		t.Fatalf("空切片应退回单结果: %+v", got)
	}
}

type emptyMultiFetcher struct{}

func (emptyMultiFetcher) Fetch() QuotaResult {
	return QuotaResult{Platform: "x", Error: "boom"}
}
func (emptyMultiFetcher) FetchAll() []QuotaResult { return nil }

// TestShortPlanAbbr 缩写派生规则(CJK 首字 / ASCII 两字符 / 派生不出为空)。
func TestShortPlanAbbr(t *testing.T) {
	cases := []struct{ in, want string }{
		{"视频赠送卡", "视"},
		{"音乐卡", "音"},
		{"MiniMax-M3", "MI"},
		{"m3", "M3"},
		{"", ""},
		{"---", ""},
	}
	for _, c := range cases {
		if got := shortPlanAbbr(c.in, 0); got != c.want {
			t.Errorf("shortPlanAbbr(%q): got %q, want %q", c.in, got, c.want)
		}
	}
}

// TestPlanAbbrSetter_NoDuplicates 同名/无名卡片必须退化成互不重复的序号缩写,
// 否则同屏会出现两个一模一样的球格。
func TestPlanAbbrSetter_NoDuplicates(t *testing.T) {
	s := &planAbbrSetter{}
	if got := s.next(0, "主套餐"); got != "" {
		t.Errorf("主卡应把缩写留给 Provider: got %q", got)
	}
	a := s.next(1, "视频卡")
	b := s.next(2, "视频卡")
	c := s.next(3, "")
	if a == b || b == c || a == c {
		t.Errorf("缩写必须互不相同: %q %q %q", a, b, c)
	}
	if a != "视" {
		t.Errorf("首个可派生缩写应直接用: got %q", a)
	}
}
