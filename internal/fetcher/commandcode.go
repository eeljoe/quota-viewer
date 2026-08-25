package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CommandCodeFetcher 通过 Command Code 私有 API(与官方 CLI 同款路由)获取额度数据。
// 端点均在 https://api.commandcode.ai 下,界面并不公开,路由来自官方 CLI
// (npm 包 command-code 的 dist/cli.mjs,函数 fetchUsageData):
//   - GET /alpha/whoami          → user/org 信息(取 orgId)
//   - GET /alpha/billing/credits → 月度剩余 credits + 5 小时/周滚动窗口限制
//
// 认证: Authorization: Bearer <API Key>。Key 在 Studio 生成(`user_...` 格式),
// 与官方 CLI `~/.commandcode/auth.json` 里保存的 apiKey 相同。
// 备注: Command Code 无公开额度查询 API,这些 /alpha/* 路由是 CLI 内部接口,
// 结构或字段名可能随 CLI 升级变化;失效时对照官方 cli.mjs 更新本文件。
type CommandCodeFetcher struct {
	apiKey       string
	baseURL      string // 可重写,用于测试
	readAuthFile bool   // apiKey 为空时是否尝试读取 ~/.commandcode/auth.json
}

// NewCommandCodeFetcher 创建一个新的 CommandCodeFetcher。
// apiKey 为空时,若本机存在 ~/.commandcode/auth.json(已登录官方 CLI)自动读取。
func NewCommandCodeFetcher(apiKey string) *CommandCodeFetcher {
	return &CommandCodeFetcher{
		apiKey:       strings.TrimSpace(apiKey),
		baseURL:      "https://api.commandcode.ai",
		readAuthFile: true,
	}
}

// 以下响应结构对应官方 CLI 的 /alpha/* 接口(字段名取自 dist/cli.mjs)。
type commandCodeWhoami struct {
	Success bool `json:"success"`
	Org     *struct {
		ID string `json:"id"`
	} `json:"org"`
}

type commandCodeWindow struct {
	Used    float64 `json:"used"`
	Cap     float64 `json:"cap"`
	ResetAt int64   `json:"resetAt"` // epoch 毫秒
}

type commandCodeCreditsResp struct {
	Credits *struct {
		MonthlyCredits   float64 `json:"monthlyCredits"`
		PurchasedCredits float64 `json:"purchasedCredits"`
		FreeCredits      float64 `json:"freeCredits"`
	} `json:"credits"`
	WindowLimits *struct {
		FiveHour commandCodeWindow `json:"fiveHour"`
		Weekly   commandCodeWindow `json:"weekly"`
	} `json:"windowLimits"`
}

// Fetch 拉取并解析 Command Code 额度。主展示 = 5 小时窗口的使用量(与 Ollama 一致,
// 最贴近"本次还能不能用"),月度剩余与周窗口写入 Remaining 辅助。
func (f *CommandCodeFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "Command Code",
		Kind:        KindUsage,
		LastUpdated: time.Now(),
	}

	key := f.apiKey
	if key == "" && f.readAuthFile {
		key = readCommandCodeAuthKey()
	}
	if key == "" {
		result.Error = "未配置 Command Code API Key(在 Studio 生成填入,或先登录官方 CLI)"
		return result
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// 1) whoami: 个人用户 org 为 null,团队用户需带 orgId 查询。
	whoamiBody, err := f.getJSON(client, key, "/alpha/whoami")
	if err != nil {
		result.Error = err.Error()
		return result
	}
	var whoami commandCodeWhoami
	if err := json.Unmarshal(whoamiBody, &whoami); err != nil {
		result.Error = "解析 whoami 响应失败"
		return result
	}
	var orgID string
	if whoami.Org != nil {
		orgID = whoami.Org.ID
	}

	// 2) credits: 月度剩余 + 5h/周窗口。
	credRoute := "/alpha/billing/credits"
	if orgID != "" {
		credRoute += "?orgId=" + orgID
	}
	credBody, err := f.getJSON(client, key, credRoute)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	var cr commandCodeCreditsResp
	if err := json.Unmarshal(credBody, &cr); err != nil {
		result.Error = "无法解析额度数据(响应结构可能已变化)"
		return result
	}
	if cr.Credits == nil || cr.WindowLimits == nil {
		result.Error = "无法解析额度数据(响应结构可能已变化)"
		return result
	}

	five := cr.WindowLimits.FiveHour
	if five.Cap <= 0 {
		result.Error = "未找到 5 小时窗口额度,响应结构可能已变化"
		return result
	}
	result.Used = five.Used
	result.Total = five.Cap
	result.Percent = five.Used / five.Cap * 100
	if five.ResetAt > 0 {
		result.ResetAt = time.UnixMilli(five.ResetAt).UTC().Format(time.RFC3339)
	}

	remain := fmt.Sprintf("5小时 %.2f/%.2f 已用", five.Used, five.Cap)
	if wk := cr.WindowLimits.Weekly; wk.Cap > 0 {
		remain += fmt.Sprintf(" · 周 %.2f/%.2f", wk.Used, wk.Cap)
	}
	total := cr.Credits.MonthlyCredits + cr.Credits.PurchasedCredits + cr.Credits.FreeCredits
	remain += fmt.Sprintf(" · 余额 $%.2f", total)
	result.Remaining = remain

	return result
}

// getJSON 发起一次带 Bearer 认证的 GET 请求,返回 200 的响应体。
// 401/403 判定为 Key 失效,其余非 200 透出状态码。
func (f *CommandCodeFetcher) getJSON(client *http.Client, key, route string) ([]byte, error) {
	url := strings.TrimRight(f.baseURL, "/") + route
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 401 || resp.StatusCode == 403:
		return nil, fmt.Errorf("API Key 无效或已过期,请在 Studio 重新生成")
	case resp.StatusCode != 200:
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// readCommandCodeAuthKey 从 ~/.commandcode/auth.json 读取官方 CLI 已保存的 API Key。
// 文件缺失或解析失败时返回空字符串(调用方回退到"未配置"错误)。
func readCommandCodeAuthKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".commandcode", "auth.json"))
	if err != nil {
		return ""
	}
	var auth struct {
		APIKey string `json:"apiKey"`
	}
	if json.Unmarshal(data, &auth) != nil {
		return ""
	}
	return strings.TrimSpace(auth.APIKey)
}
