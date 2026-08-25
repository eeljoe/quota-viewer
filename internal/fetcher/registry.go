package fetcher

// CredentialField 描述一个 Provider 的凭证字段(驱动前端动态渲染输入框)。
type CredentialField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"` // "password" | "text" | "textarea"
}

// ProviderDef 描述一个可配置的 Provider(注册表条目)。
type ProviderDef struct {
	ID          string
	DisplayName string
	Abbr        string // 球格缩写
	Kind        string // KindUsage | KindBalance
	LoginURL    string // 打开登录页按钮 URL(空 = 不显示按钮)
	Fields      []CredentialField
	Build       func(creds map[string]string) Fetcher
}

// registry 是全部已知 Provider 的注册表,顺序固定 = 推荐展示顺序。
var registry = []ProviderDef{
	{
		ID:          "kimi",
		DisplayName: "Kimi",
		Abbr:        "K",
		Kind:        KindUsage,
		Fields: []CredentialField{
			{Key: "api_key", Label: "API Key", Type: "password"},
		},
		Build: func(creds map[string]string) Fetcher {
			return NewKimiFetcher(creds["api_key"])
		},
	},
	{
		ID:          "xfyun",
		DisplayName: "讯飞星辰",
		Abbr:        "讯",
		Kind:        KindUsage,
		LoginURL:    "https://maas.xfyun.cn/packageSubscription",
		Fields: []CredentialField{
			{Key: "cookie", Label: "Cookie(浏览器 F12 复制)", Type: "textarea"},
		},
		Build: func(creds map[string]string) Fetcher {
			return NewXfyunFetcher(creds["cookie"], "")
		},
	},
	{
		ID:          "opencode-go",
		DisplayName: "OpenCode Go",
		Abbr:        "Go",
		Kind:        KindUsage,
		LoginURL:    "https://opencode.ai",
		Fields: []CredentialField{
			{Key: "workspace_id", Label: "Workspace ID", Type: "text"},
			{Key: "session_token", Label: "Session Token", Type: "password"},
		},
		Build: func(creds map[string]string) Fetcher {
			return NewOpenCodeGoFetcher(creds["workspace_id"], creds["session_token"])
		},
	},
	{
		ID:          "mimo",
		DisplayName: "小米 MiMo",
		Abbr:        "M",
		Kind:        KindUsage,
		LoginURL:    "https://platform.xiaomimimo.com/console/plan-manage",
		Fields: []CredentialField{
			{Key: "cookie", Label: "Cookie(浏览器 F12 复制)", Type: "textarea"},
		},
		Build: func(creds map[string]string) Fetcher {
			return NewMiMoFetcher(creds["cookie"], "")
		},
	},
	{
		ID:          "deepseek",
		DisplayName: "DeepSeek",
		Abbr:        "D",
		Kind:        KindBalance,
		LoginURL:    "https://platform.deepseek.com",
		Fields: []CredentialField{
			{Key: "api_key", Label: "API Key", Type: "password"},
		},
		Build: func(creds map[string]string) Fetcher {
			return NewDeepSeekFetcher(creds["api_key"])
		},
	},
	{
		ID:          "ollama",
		DisplayName: "Ollama",
		Abbr:        "O",
		Kind:        KindUsage,
		LoginURL:    "https://ollama.com/settings",
		Fields: []CredentialField{
			{Key: "cookie", Label: "Cookie(登录 ollama.com 后复制)", Type: "textarea"},
		},
		Build: func(creds map[string]string) Fetcher {
			return NewOllamaFetcher(creds["cookie"])
		},
	},
	{
		ID:          "command-code",
		DisplayName: "Command Code",
		Abbr:        "C",
		Kind:        KindUsage,
		LoginURL:    "https://commandcode.ai/studio",
		Fields: []CredentialField{
			{Key: "api_key", Label: "API Key(留空自动读 ~/.commandcode/auth.json)", Type: "password"},
		},
		Build: func(creds map[string]string) Fetcher {
			return NewCommandCodeFetcher(creds["api_key"])
		},
	},
}

// GetAll 返回全部注册 Provider 的副本(固定顺序)。
func GetAll() []ProviderDef {
	out := make([]ProviderDef, len(registry))
	copy(out, registry)
	return out
}

// Get 按 id 查找 Provider;不存在返回 false。
func Get(id string) (ProviderDef, bool) {
	for _, d := range registry {
		if d.ID == id {
			return d, true
		}
	}
	return ProviderDef{}, false
}
