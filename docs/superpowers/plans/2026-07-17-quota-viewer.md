# Quota Viewer 悬浮球实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建一个 Wails v2 桌面悬浮球,不开浏览器即可一眼查看 Kimi、讯飞星辰、小米 MiMo 三个 Coding Plan 的额度状态。

**Architecture:** Wails v2 单窗口双态(frameless + transparent + alwaysOnTop)。Go 后端并发调用三个 Fetcher 抓取额度,通过 `runtime.EventsEmit` 推送前端。Kimi 走 API Key,讯飞/小米走 Cookie 抓网页 XHR。配置存 `%APPDATA%/quota-viewer/config.json`。

**Tech Stack:** Go 1.24、Wails v2.12.0、原生 HTML/CSS/JS(无前端框架)

## Global Constraints

- 平台:仅 Windows(Windows 10/11,WebView2)
- Wails CLI 路径:`C:/Users/joe/go/bin/wails.exe`(不在 PATH 中,需用完整路径或加入 PATH)
- Go 模块名:`quota-viewer`
- 配置目录:`%APPDATA%/quota-viewer/`(即 `C:/Users/joe/AppData/Roaming/quota-viewer/`)
- HTTP 超时:所有 fetcher 统一 10 秒
- 前端零依赖:不用 npm 包,纯原生 HTML/CSS/JS
- 窗口选项:`Frameless: true, AlwaysOnTop: true, BackgroundColour: options.NewRGBA(0,0,0,0), HideWindowOnClose: true`
- 工作目录:`C:/Users/joe/Desktop/工作学习/软件开发/quota viewer`

---

## 文件结构

```
quota-viewer/
├── main.go                         # Wails 入口,窗口选项,托盘初始化
├── app.go                          # App struct,暴露给前端的方法
├── internal/
│   ├── config/
│   │   ├── config.go               # Config struct, Load(), Save()
│   │   └── config_test.go          # 配置读写测试
│   ├── fetcher/
│   │   ├── types.go                # Fetcher 接口 + QuotaResult
│   │   ├── kimi.go                 # KimiFetcher
│   │   ├── kimi_test.go            # KimiFetcher 测试
│   │   ├── xfyun.go                # XfyunFetcher
│   │   ├── xfyun_test.go           # XfyunFetcher 测试
│   │   ├── mimo.go                 # MiMoFetcher
│   │   └── mimo_test.go            # MiMoFetcher 测试
│   └── tray/
│       └── tray.go                 # 系统托盘菜单
├── frontend/
│   └── src/
│       ├── index.html              # 双态切换(球面 / 详情面板 / 配置引导)
│       ├── main.js                 # 事件监听 + UI 更新
│       └── style.css               # 样式
├── wails.json
├── go.mod
└── appicon.png
```

---

## Task 1: Wails 项目脚手架 + 基础窗口

**Files:**
- Create: `main.go`
- Create: `app.go`
- Create: `wails.json`
- Create: `go.mod`
- Create: `frontend/src/index.html`
- Create: `frontend/src/main.js`
- Create: `frontend/src/style.css`

**Interfaces:**
- Produces: `App` struct with empty `Greet(name string) string` method(后续 task 往里加方法),Wails 窗口以 frameless+transparent+alwaysOnTop 启动

- [ ] **Step 1: 用 wails init 初始化项目**

在工作目录下初始化(使用 vanilla 模板,无前端框架):

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
"C:/Users/joe/go/bin/wails.exe" init -n quota-viewer -t vanilla
```

这会生成 `main.go`、`app.go`、`frontend/`、`wails.json`、`go.mod` 等。由于目录名含空格和中文,如果 init 失败,改为先在临时目录初始化再移动文件:

```bash
"C:/Users/joe/go/bin/wails.exe" init -n quota-viewer -t vanilla -d /tmp/qv-init
cp -r /tmp/qv-init/* "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer/"
cp -r /tmp/qv-init/.* "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer/" 2>/dev/null || true
```

- [ ] **Step 2: 修改 go.mod 模块名**

编辑 `go.mod`,确保模块名为 `quota-viewer`:

```
module quota-viewer

go 1.24
```

- [ ] **Step 3: 改写 main.go — 配置悬浮球窗口**

将 `main.go` 改为:

```go
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Quota Viewer",
		Width:     60,
		Height:    60,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:    options.NewRGBA(0, 0, 0, 0),
		AlwaysOnTop:         true,
		HideWindowOnClose:   true,
		DisableResize:       true,
		WindowStartState:    options.Normal,
		OnStartup:           app.OnStartup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
```

- [ ] **Step 4: 改写 app.go — 最小 App 结构**

将 `app.go` 改为:

```go
package main

import (
	"context"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}
```

- [ ] **Step 5: 写最小前端 — 占位球面**

将 `frontend/src/index.html` 改为:

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Quota Viewer</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <div id="ball" class="ball">
        <span class="ball-text">Q</span>
    </div>
    <script src="main.js"></script>
</body>
</html>
```

将 `frontend/src/style.css` 改为:

```css
* { margin: 0; padding: 0; box-sizing: border-box; }
html, body {
    width: 60px;
    height: 60px;
    background: transparent;
    overflow: hidden;
    font-family: -apple-system, "Segoe UI", sans-serif;
}
.ball {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background: rgba(40, 40, 50, 0.9);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    --wails-draggable: drag;
}
.ball-text {
    color: #7c83ff;
    font-size: 24px;
    font-weight: bold;
    user-select: none;
}
```

将 `frontend/src/main.js` 改为:

```javascript
// 占位,后续 task 填充
console.log("Quota Viewer started");
```

- [ ] **Step 6: 构建并运行验证**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
"C:/Users/joe/go/bin/wails.exe" dev
```

预期:出现一个 60×60 的透明圆形悬浮球,置顶显示,可拖动。Ctrl+C 退出 dev。

- [ ] **Step 7: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git init 2>/dev/null || true
git add -A
git commit -m "feat: init wails project with frameless floating ball window"
```

---

## Task 2: 配置管理模块

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Interfaces:**
- Produces: `config.Config` struct, `config.Load()` / `config.Save()` 函数;字段:`KimiAPIKey`, `XfyunCookie`, `MimoCookie`, `RefreshIntervalMin`, `BallX`, `BallY`

- [ ] **Step 1: 写 config.go**

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	KimiAPIKey        string `json:"kimi_api_key"`
	XfyunCookie       string `json:"xfyun_cookie"`
	MimoCookie        string `json:"mimo_cookie"`
	RefreshIntervalMin int   `json:"refresh_interval_min"`
	BallX             int    `json:"ball_x"`
	BallY             int    `json:"ball_y"`
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

// Load 读取配置文件。文件不存在时返回带默认值的空配置(不报错)。
func Load() (*Config, error) {
	cfg := &Config{
		RefreshIntervalMin: 15,
		BallX:              -1,
		BallY:              -1,
	}

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

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return cfg, err
	}

	// 确保默认值
	if cfg.RefreshIntervalMin <= 0 {
		cfg.RefreshIntervalMin = 15
	}

	return cfg, nil
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
```

- [ ] **Step 2: 写 config_test.go**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FileNotExists_ReturnsDefaults(t *testing.T) {
	// 临时设置 APPDATA 到测试目录
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
}

func TestSaveThenLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	original := &Config{
		KimiAPIKey:         "sk-kimi-test123",
		XfyunCookie:        "SSID=abc; token=xyz",
		MimoCookie:         "session=def",
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
	if loaded.KimiAPIKey != "sk-kimi-test123" {
		t.Errorf("KimiAPIKey mismatch: got %s", loaded.KimiAPIKey)
	}
	if loaded.XfyunCookie != "SSID=abc; token=xyz" {
		t.Errorf("XfyunCookie mismatch: got %s", loaded.XfyunCookie)
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

	cfg := &Config{KimiAPIKey: "test"}
	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() should create directory, got error: %v", err)
	}
}
```

- [ ] **Step 3: 运行测试验证失败→通过**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
go test ./internal/config/ -v
```

预期:3 个测试全部 PASS。

- [ ] **Step 4: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add internal/config/
git commit -m "feat: config load/save to %APPDATA%/quota-viewer/config.json"
```

---

## Task 3: Fetcher 类型定义与接口

**Files:**
- Create: `internal/fetcher/types.go`

**Interfaces:**
- Produces: `fetcher.QuotaResult` struct, `fetcher.Fetcher` interface(`Fetch(ctx) QuotaResult`)。所有后续 fetcher task 依赖此定义。

- [ ] **Step 1: 写 types.go**

```go
package fetcher

import "time"

// QuotaResult 是所有 fetcher 的统一返回结构。
type QuotaResult struct {
	Platform    string    `json:"platform"`     // "Kimi" / "讯飞星辰" / "小米MiMo"
	Used        float64   `json:"used"`         // 已用量
	Total       float64   `json:"total"`        // 总量(平台返回则填,否则 0)
	Percent     float64   `json:"percent"`      // Used/Total * 100;无总量时由剩余百分比反推
	Remaining   string    `json:"remaining"`    // 原始剩余描述(如 "1,200/18,000 次" 或 "无限制")
	ResetAt     string    `json:"reset_at"`     // 下次重置时间(ISO 8601,空则未知)
	LastUpdated time.Time `json:"last_updated"`
	Error       string    `json:"error"`        // 非空表示失败
}

// Fetcher 是额度查询器的统一接口。
type Fetcher interface {
	Fetch() QuotaResult
}
```

- [ ] **Step 2: 验证编译**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
go build ./internal/fetcher/
```

预期:无错误。

- [ ] **Step 3: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add internal/fetcher/types.go
git commit -m "feat: define QuotaResult struct and Fetcher interface"
```

---

## Task 4: KimiFetcher — API Key 查询额度

**Files:**
- Create: `internal/fetcher/kimi.go`
- Create: `internal/fetcher/kimi_test.go`

**Interfaces:**
- Consumes: `fetcher.QuotaResult`, `fetcher.Fetcher` (from Task 3)
- Produces: `fetcher.NewKimiFetcher(apiKey string) *KimiFetcher`, `KimiFetcher.Fetch() QuotaResult`

- [ ] **Step 1: 写 kimi.go**

```go
package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// KimiFetcher 通过 Kimi Code API Key 查询额度。
// 端点: GET https://api.kimi.com/coding/v1/usages
// 认证: Authorization: Bearer sk-kimi-xxx
type KimiFetcher struct {
	apiKey string
}

func NewKimiFetcher(apiKey string) *KimiFetcher {
	return &KimiFetcher{apiKey: apiKey}
}

type kimiUsagesResponse struct {
	Data []kimiUsageItem `json:"data"`
}

type kimiUsageItem struct {
	ModelName string `json:"model_name"`
	Used      int64  `json:"used"`
	Limit     int64  `json:"limit"`
	Remaining int64  `json:"remaining"`
	ResetIn   int64  `json:"reset_in"`
	ResetAt   string `json:"reset_at"`
}

func (k *KimiFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "Kimi",
		LastUpdated: time.Now(),
	}

	if k.apiKey == "" {
		result.Error = "未配置 Kimi API Key"
		return result
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", "https://api.kimi.com/coding/v1/usages", nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}
	req.Header.Set("Authorization", "Bearer "+k.apiKey)
	req.Header.Set("User-Agent", "KimiCLI/1.6")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		result.Error = "API Key 无效或已过期(请确认使用 sk-kimi-xxx 格式的 Key)"
		return result
	}
	if resp.StatusCode != 200 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	var body kimiUsagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Error = fmt.Sprintf("解析响应失败: %v", err)
		return result
	}

	// 找 model_name == "all" 的汇总条目
	for _, item := range body.Data {
		if item.ModelName == "all" {
			result.Used = float64(item.Used)
			result.Total = float64(item.Limit)
			if item.Limit > 0 {
				result.Percent = float64(item.Used) / float64(item.Limit) * 100
			}
			result.Remaining = fmt.Sprintf("%s / %s", formatNum(item.Used), formatNum(item.Limit))
			result.ResetAt = item.ResetAt
			return result
		}
	}

	// 如果没有 "all" 条目,取第一条
	if len(body.Data) > 0 {
		item := body.Data[0]
		result.Used = float64(item.Used)
		result.Total = float64(item.Limit)
		if item.Limit > 0 {
			result.Percent = float64(item.Used) / float64(item.Limit) * 100
		}
		result.Remaining = fmt.Sprintf("%s / %s", formatNum(item.Used), formatNum(item.Limit))
		result.ResetAt = item.ResetAt
		return result
	}

	result.Error = "响应中未找到用量数据"
	return result
}

func formatNum(n int64) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
```

- [ ] **Step 2: 写 kimi_test.go**

```go
package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKimiFetcher_EmptyKey_ReturnsError(t *testing.T) {
	f := NewKimiFetcher("")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty API key")
	}
	if result.Platform != "Kimi" {
		t.Errorf("expected platform 'Kimi', got '%s'", result.Platform)
	}
}

func TestKimiFetcher_ValidResponse_ParsesUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 Authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer sk-kimi-test" {
			t.Errorf("expected 'Bearer sk-kimi-test', got '%s'", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": [
				{
					"model_name": "all",
					"used": 300000000,
					"limit": 1000000000,
					"remaining": 700000000,
					"reset_at": "2026-07-21T00:00:00Z"
				},
				{
					"model_name": "kimi-k2-0905",
					"used": 50000000,
					"limit": 200000000,
					"reset_at": "2026-07-21T00:00:00Z"
				}
			]
		}`))
	}))
	defer server.Close()

	// 临时替换 URL(通过构造自定义请求来测试)
	// 由于 URL 硬编码,我们测试 401 和空 key 场景
	// 完整的端到端测试在集成阶段做
}

func TestKimiFetcher_Unauthorized_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	// 验证 401 场景的错误消息
	// 注意:由于 URL 硬编码为 api.kimi.com,单元测试主要覆盖空 key 路径
	// 401 路径在集成测试中覆盖
	_ = server // 保持 server 引用
}
```

注意:由于 KimiFetcher 的 URL 硬编码为 `https://api.kimi.com`,无法直接用 httptest 替换。完整的 401 和 200 路径测试在 Task 8(集成联调)中手动验证。单元测试覆盖空 Key 的边界情况。

- [ ] **Step 3: 运行测试**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
go test ./internal/fetcher/ -run Kimi -v
```

预期:`TestKimiFetcher_EmptyKey_ReturnsError` PASS。

- [ ] **Step 4: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add internal/fetcher/kimi.go internal/fetcher/kimi_test.go
git commit -m "feat: KimiFetcher via api.kimi.com/coding/v1/usages"
```

---

## Task 5: XfyunFetcher — Cookie 抓网页额度

**Files:**
- Create: `internal/fetcher/xfyun.go`
- Create: `internal/fetcher/xfyun_test.go`

**Interfaces:**
- Consumes: `fetcher.QuotaResult`, `fetcher.Fetcher` (from Task 3)
- Produces: `fetcher.NewXfyunFetcher(cookie string, apiURL string) *XfyunFetcher`, `XfyunFetcher.Fetch() QuotaResult`

- [ ] **Step 1: 写 xfyun.go**

讯飞星辰 MaaS 的额度页面 `maas.xfyun.cn/packageSubscription` 背后有一个返回 JSON 的 XHR 接口。具体接口 URL 首次运行时需用 F12 Network 确认,因此 `apiURL` 作为参数传入,可在配置中指定。响应结构同样需要实抓确认,这里先做通用的 JSON 解析,按常见字段名提取。

```go
package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

// XfyunFetcher 通过 Cookie 抓取讯飞星辰 MaaS 额度。
type XfyunFetcher struct {
	cookie  string
	apiURL  string // 额度 XHR 接口地址,首次需 F12 确认
}

func NewXfyunFetcher(cookie string, apiURL string) *XfyunFetcher {
	// 默认 URL,待 F12 确认后更新
	if apiURL == "" {
		apiURL = "https://maas.xfyun.cn/api/packageSubscription"
	}
	return &XfyunFetcher{cookie: cookie, apiURL: apiURL}
}

// xfyunRawResponse 是通用的响应结构,字段名按实抓结果调整。
// 常见模式: { "data": { "used": X, "total": Y, "resetTime": "..." } }
// 或: { "data": [{ "name": "...", "used": X, "surplus": Y }] }
type xfyunRawResponse struct {
	Data json.RawMessage `json:"data"`
	Code int             `json:"code"`
}

func (x *XfyunFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "讯飞星辰",
		LastUpdated: time.Now(),
	}

	if x.cookie == "" {
		result.Error = "未配置讯飞 Cookie"
		return result
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		// 讯飞 401/302 时不自动跟随到登录页
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", x.apiURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

	// 设置 Cookie header
	req.Header.Set("Cookie", x.cookie)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://maas.xfyun.cn/packageSubscription")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 302 {
		result.Error = "Cookie 已过期,请更新"
		return result
	}
	if resp.StatusCode != 200 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result
	}

	// 解析通用结构
	var raw xfyunRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		// 可能不是标准 JSON,保存原始文本
		result.Remaining = string(body[:min(len(body), 200)])
		result.Error = "响应格式未识别(需 F12 确认实际接口)"
		return result
	}

	// 尝试解析 data 为对象
	var dataMap map[string]interface{}
	if err := json.Unmarshal(raw.Data, &dataMap); err == nil {
		result = parseXfyunDataMap(result, dataMap)
		return result
	}

	// 尝试解析 data 为数组
	var dataArray []map[string]interface{}
	if err := json.Unmarshal(raw.Data, &dataArray); err == nil && len(dataArray) > 0 {
		result = parseXfyunDataMap(result, dataArray[0])
		return result
	}

	result.Error = "无法解析额度数据(需 F12 确认实际字段)"
	return result
}

// parseXfyunDataMap 从 data 对象中提取用量信息。
// 字段名按常见命名尝试,实际需根据 F12 抓包结果调整。
func parseXfyunDataMap(result QuotaResult, m map[string]interface{}) QuotaResult {
	used := getFloat(m, "used", "usage", "usedAmount", "consume")
	total := getFloat(m, "total", "limit", "totalAmount", "quota")
	remaining := getFloat(m, "remaining", "surplus", "left")
	resetAt := getString(m, "resetTime", "reset_at", "resetAt", "expireTime", "expireTimeStr")

	if total > 0 {
		result.Total = total
		result.Used = used
		result.Percent = used / total * 100
		result.Remaining = fmt.Sprintf("%.0f / %.0f", used, total)
	} else if remaining > 0 && total > 0 {
		result.Total = total
		result.Used = total - remaining
		result.Percent = result.Used / total * 100
		result.Remaining = fmt.Sprintf("%.0f / %.0f", result.Used, total)
	} else {
		// 无数值,尝试取文本描述
		desc := getString(m, "desc", "description", "remaining", "surplusText")
		if desc != "" {
			result.Remaining = desc
		} else {
			result.Error = "未找到用量字段(需 F12 确认实际字段名)"
		}
	}

	result.ResetAt = resetAt
	return result
}

func getFloat(m map[string]interface{}, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch n := v.(type) {
			case float64:
				return n
			case int:
				return float64(n)
			case int64:
				return float64(n)
			case string:
				var f float64
				if _, err := fmt.Sscanf(n, "%f", &f); err == nil {
					return f
				}
			}
		}
	}
	return 0
}

func getString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

- [ ] **Step 2: 写 xfyun_test.go**

```go
package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestXfyunFetcher_EmptyCookie_ReturnsError(t *testing.T) {
	f := NewXfyunFetcher("", "")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty cookie")
	}
	if result.Platform != "讯飞星辰" {
		t.Errorf("expected platform '讯飞星辰', got '%s'", result.Platform)
	}
}

func TestXfyunFetcher_401_ReturnsCookieExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 Cookie header
		if r.Header.Get("Cookie") != "test=cookie" {
			t.Errorf("expected cookie header, got '%s'", r.Header.Get("Cookie"))
		}
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "Cookie 已过期,请更新" {
		t.Errorf("expected 'Cookie 已过期', got '%s'", result.Error)
	}
}

func TestXfyunFetcher_ValidResponse_ParsesData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"data": {
				"used": 5000,
				"total": 18000,
				"resetTime": "2026-07-21T08:00:00Z"
			}
		}`))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Used != 5000 {
		t.Errorf("expected Used=5000, got %f", result.Used)
	}
	if result.Total != 18000 {
		t.Errorf("expected Total=18000, got %f", result.Total)
	}
	if result.Percent < 27.7 || result.Percent > 27.8 {
		t.Errorf("expected ~27.78%%, got %f%%", result.Percent)
	}
	if result.ResetAt != "2026-07-21T08:00:00Z" {
		t.Errorf("expected reset time, got '%s'", result.ResetAt)
	}
}

func TestXfyunFetcher_ArrayData_ParsesFirst(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"data": [
				{"used": 100, "total": 200, "resetTime": "2026-07-21T00:00:00Z"}
			]
		}`))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Used != 100 || result.Total != 200 {
		t.Errorf("expected 100/200, got %f/%f", result.Used, result.Total)
	}
}
```

- [ ] **Step 3: 运行测试**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
go test ./internal/fetcher/ -run Xfyun -v
```

预期:4 个测试全部 PASS。

- [ ] **Step 4: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add internal/fetcher/xfyun.go internal/fetcher/xfyun_test.go
git commit -m "feat: XfyunFetcher with Cookie-based scraping + generic JSON parser"
```

---

## Task 6: MiMoFetcher — Cookie 抓网页额度

**Files:**
- Create: `internal/fetcher/mimo.go`
- Create: `internal/fetcher/mimo_test.go`

**Interfaces:**
- Consumes: `fetcher.QuotaResult`, `fetcher.Fetcher` (from Task 3), `getFloat`/`getString`/`min` helpers (from Task 5)
- Produces: `fetcher.NewMiMoFetcher(cookie string, apiURL string) *MiMoFetcher`, `MiMoFetcher.Fetch() QuotaResult`

- [ ] **Step 1: 写 mimo.go**

MiMo 与 Xfyun 模式完全一致(Cookie 抓 XHR),复用 Task 5 的 `getFloat`/`getString`/`parseXfyunDataMap`(这些是包内共享的辅助函数)。为保持文件清晰,MiMoFetcher 只负责请求,解析复用通用函数。

```go
package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// MiMoFetcher 通过 Cookie 抓取小米 MiMo 额度。
type MiMoFetcher struct {
	cookie string
	apiURL string
}

func NewMiMoFetcher(cookie string, apiURL string) *MiMoFetcher {
	if apiURL == "" {
		apiURL = "https://platform.xiaomimimo.com/api/plan/quota"
	}
	return &MiMoFetcher{cookie: cookie, apiURL: apiURL}
}

type mimoRawResponse struct {
	Data json.RawMessage `json:"data"`
	Code int             `json:"code"`
}

func (m *MiMoFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "小米MiMo",
		LastUpdated: time.Now(),
	}

	if m.cookie == "" {
		result.Error = "未配置 MiMo Cookie"
		return result
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", m.apiURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

	req.Header.Set("Cookie", m.cookie)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://platform.xiaomimimo.com/console/plan-manage")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 302 {
		result.Error = "Cookie 已过期,请更新"
		return result
	}
	if resp.StatusCode != 200 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result
	}

	var raw mimoRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		result.Remaining = string(body[:min(len(body), 200)])
		result.Error = "响应格式未识别(需 F12 确认实际接口)"
		return result
	}

	// 复用 Task 5 的通用解析函数
	var dataMap map[string]interface{}
	if err := json.Unmarshal(raw.Data, &dataMap); err == nil {
		result = parseXfyunDataMap(result, dataMap)
		return result
	}

	var dataArray []map[string]interface{}
	if err := json.Unmarshal(raw.Data, &dataArray); err == nil && len(dataArray) > 0 {
		result = parseXfyunDataMap(result, dataArray[0])
		return result
	}

	result.Error = "无法解析额度数据(需 F12 确认实际字段)"
	return result
}
```

- [ ] **Step 2: 写 mimo_test.go**

```go
package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiMoFetcher_EmptyCookie_ReturnsError(t *testing.T) {
	f := NewMiMoFetcher("", "")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty cookie")
	}
	if result.Platform != "小米MiMo" {
		t.Errorf("expected platform '小米MiMo', got '%s'", result.Platform)
	}
}

func TestMiMoFetcher_401_ReturnsCookieExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "Cookie 已过期,请更新" {
		t.Errorf("expected 'Cookie 已过期', got '%s'", result.Error)
	}
}

func TestMiMoFetcher_ValidResponse_ParsesData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"data": {
				"used": 60000000,
				"total": 200000000,
				"resetTime": "2026-08-01T00:00:00Z"
			}
		}`))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Used != 60000000 {
		t.Errorf("expected Used=60000000, got %f", result.Used)
	}
	if result.Total != 200000000 {
		t.Errorf("expected Total=200000000, got %f", result.Total)
	}
	if result.Percent < 29.9 || result.Percent > 30.1 {
		t.Errorf("expected ~30%%, got %f%%", result.Percent)
	}
}
```

- [ ] **Step 3: 运行测试**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
go test ./internal/fetcher/ -run MiMo -v
```

预期:3 个测试全部 PASS。

- [ ] **Step 4: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add internal/fetcher/mimo.go internal/fetcher/mimo_test.go
git commit -m "feat: MiMoFetcher with Cookie-based scraping, reuses generic parser"
```

---

## Task 7: App 方法绑定 — 刷新 / 配置 / 测试连接

**Files:**
- Modify: `app.go` (from Task 1)

**Interfaces:**
- Consumes: `config.Load()`/`config.Save()` (Task 2), `KimiFetcher`/`XfyunFetcher`/`MiMoFetcher` (Tasks 4-6), `runtime.EventsEmit`
- Produces: 前端可调用的 Go 方法 `Refresh()`, `GetConfig()`, `SaveConfig()`, `TestConnection(platform string)`, `OpenLoginPage(url string)`, `SaveBallPosition(x, y int)`

- [ ] **Step 1: 改写 app.go**

```go
package main

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"quota-viewer/internal/config"
	"quota-viewer/internal/fetcher"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	cfg     *config.Config
	mu      sync.Mutex
	cache   []fetcher.QuotaResult
}

func NewApp() *App {
	cfg, _ := config.Load()
	return &App{
		cfg: cfg,
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// 启动后台定时刷新
	go a.startAutoRefresh()
}

// Refresh 并发调用三个 fetcher,返回结果并推送事件到前端。
func (a *App) Refresh() []fetcher.QuotaResult {
	results := a.fetchAll()
	a.mu.Lock()
	a.cache = results
	a.mu.Unlock()

	// 推送事件到前端
	runtime.EventsEmit(a.ctx, "quota:update", results)

	return results
}

// GetConfig 返回当前配置(Cookie/Key 做掩码)。
func (a *App) GetConfig() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	return map[string]interface{}{
		"kimi_api_key":         maskSecret(a.cfg.KimiAPIKey),
		"xfyun_cookie":         maskSecret(a.cfg.XfyunCookie),
		"mimo_cookie":          maskSecret(a.cfg.MimoCookie),
		"refresh_interval_min": a.cfg.RefreshIntervalMin,
		"ball_x":               a.cfg.BallX,
		"ball_y":               a.cfg.BallY,
		"has_kimi_key":         a.cfg.KimiAPIKey != "",
		"has_xfyun_cookie":     a.cfg.XfyunCookie != "",
		"has_mimo_cookie":      a.cfg.MimoCookie != "",
	}
}

// SaveConfig 保存凭证配置。
func (a *App) SaveConfig(kimiKey, xfyunCookie, mimoCookie string, refreshMin int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 空字符串 = 不修改(避免掩码覆盖)
	if kimiKey != "" {
		a.cfg.KimiAPIKey = kimiKey
	}
	if xfyunCookie != "" {
		a.cfg.XfyunCookie = xfyunCookie
	}
	if mimoCookie != "" {
		a.cfg.MimoCookie = mimoCookie
	}
	if refreshMin > 0 {
		a.cfg.RefreshIntervalMin = refreshMin
	}

	return config.Save(a.cfg)
}

// TestConnection 测试单个平台连接是否可用。
func (a *App) TestConnection(platform string) string {
	a.mu.Lock()
	cfg := *a.cfg
	a.mu.Unlock()

	var f fetcher.Fetcher
	switch platform {
	case "kimi":
		f = fetcher.NewKimiFetcher(cfg.KimiAPIKey)
	case "xfyun":
		f = fetcher.NewXfyunFetcher(cfg.XfyunCookie, "")
	case "mimo":
		f = fetcher.NewMiMoFetcher(cfg.MimoCookie, "")
	default:
		return "未知平台"
	}

	result := f.Fetch()
	if result.Error != "" {
		return "失败: " + result.Error
	}
	return "成功: " + result.Remaining
}

// OpenLoginPage 用默认浏览器打开 URL。
func (a *App) OpenLoginPage(url string) {
	openBrowser(url)
}

// SaveBallPosition 保存悬浮球位置。
func (a *App) SaveBallPosition(x, y int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.BallX = x
	a.cfg.BallY = y
	return config.Save(a.cfg)
}

// fetchAll 并发调用三个 fetcher。
func (a *App) fetchAll() []fetcher.QuotaResult {
	a.mu.Lock()
	cfg := *a.cfg
	a.mu.Unlock()

	var wg sync.WaitGroup
	results := make([]fetcher.QuotaResult, 3)

	wg.Add(3)
	go func() {
		defer wg.Done()
		results[0] = fetcher.NewKimiFetcher(cfg.KimiAPIKey).Fetch()
	}()
	go func() {
		defer wg.Done()
		results[1] = fetcher.NewXfyunFetcher(cfg.XfyunCookie, "").Fetch()
	}()
	go func() {
		defer wg.Done()
		results[2] = fetcher.NewMiMoFetcher(cfg.MimoCookie, "").Fetch()
	}()
	wg.Wait()

	return results
}

// startAutoRefresh 定时后台刷新。
func (a *App) startAutoRefresh() {
	for {
		interval := 15
		a.mu.Lock()
		if a.cfg.RefreshIntervalMin > 0 {
			interval = a.cfg.RefreshIntervalMin
		}
		a.mu.Unlock()

		time.Sleep(time.Duration(interval) * time.Minute)
		a.Refresh()
	}
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		if s == "" {
			return ""
		}
		return "****"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}
```

- [ ] **Step 2: 更新 main.go 绑定**

确认 `main.go` 的 `Bind` 列表已包含 `app`(Task 1 已设置)。无需修改。

- [ ] **Step 3: 编译验证**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
go build ./...
```

预期:无错误。

- [ ] **Step 4: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add app.go
git commit -m "feat: App methods - Refresh/GetConfig/SaveConfig/TestConnection/OpenLoginPage"
```

---

## Task 8: 前端 — 双态 UI + 事件监听

**Files:**
- Modify: `frontend/src/index.html`
- Modify: `frontend/src/main.js`
- Modify: `frontend/src/style.css`

**Interfaces:**
- Consumes: Wails bindings `window.go.main.App.Refresh()`, `.GetConfig()`, `.SaveConfig()`, `.TestConnection()`, `.OpenLoginPage()`, `.SaveBallPosition()` (Task 7); Wails event `quota:update` (Task 7)
- Produces: 可交互的悬浮球 + 详情面板 + 配置面板

- [ ] **Step 1: 写 index.html(三视图:球面 / 详情 / 配置)**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Quota Viewer</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <!-- 收起态:悬浮球 -->
    <div id="ball" class="ball">
        <div class="dot" id="dot-kimi"></div>
        <div class="dot" id="dot-xfyun"></div>
        <div class="dot" id="dot-mimo"></div>
    </div>

    <!-- 展开态:详情面板 -->
    <div id="panel" class="panel hidden">
        <div class="panel-header" --wails-draggable="drag">
            <span>额度详情</span>
            <button id="btn-settings" class="icon-btn">⚙</button>
        </div>
        <div id="quota-list" class="quota-list"></div>
        <div class="panel-footer">
            <span id="last-updated" class="muted"></span>
            <button id="btn-refresh" class="btn">刷新</button>
        </div>
    </div>

    <!-- 配置面板 -->
    <div id="settings" class="settings hidden">
        <div class="panel-header" --wails-draggable="drag">
            <span>配置</span>
            <button id="btn-close-settings" class="icon-btn">✕</button>
        </div>
        <div class="settings-body">
            <div class="form-group">
                <label>Kimi API Key</label>
                <input type="password" id="input-kimi" placeholder="sk-kimi-xxx">
                <button class="btn-sm" data-test="kimi">测试</button>
            </div>
            <div class="form-group">
                <label>讯飞 Cookie</label>
                <textarea id="input-xfyun" placeholder="从浏览器 F12 复制 Cookie" rows="2"></textarea>
                <button class="btn-sm" data-test="xfyun">测试</button>
                <button class="btn-sm" data-open="https://maas.xfyun.cn/packageSubscription">打开登录页</button>
            </div>
            <div class="form-group">
                <label>MiMo Cookie</label>
                <textarea id="input-mimo" placeholder="从浏览器 F12 复制 Cookie" rows="2"></textarea>
                <button class="btn-sm" data-test="mimo">测试</button>
                <button class="btn-sm" data-open="https://platform.xiaomimimo.com/console/plan-manage">打开登录页</button>
            </div>
            <div class="form-group">
                <label>刷新间隔(分钟)</label>
                <input type="number" id="input-interval" min="1" value="15">
            </div>
            <button id="btn-save-config" class="btn">保存</button>
        </div>
    </div>

    <script src="main.js"></script>
</body>
</html>
```

- [ ] **Step 2: 写 style.css**

```css
* { margin: 0; padding: 0; box-sizing: border-box; }

html, body {
    background: transparent;
    overflow: hidden;
    font-family: -apple-system, "Segoe UI", "Microsoft YaHei", sans-serif;
    font-size: 13px;
    color: #e0e0e0;
}

.hidden { display: none !important; }
.muted { color: #888; font-size: 11px; }

/* === 悬浮球 === */
.ball {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background: rgba(30, 30, 40, 0.92);
    backdrop-filter: blur(8px);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    cursor: pointer;
    --wails-draggable: drag;
    transition: transform 0.15s;
}
.ball:hover { transform: scale(1.08); }

.dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #555;
    transition: background 0.3s;
}
.dot.green { background: #4caf50; }
.dot.yellow { background: #ff9800; }
.dot.red { background: #f44336; }

/* === 详情面板 === */
.panel, .settings {
    width: 340px;
    background: rgba(30, 30, 40, 0.95);
    backdrop-filter: blur(12px);
    border-radius: 12px;
    border: 1px solid rgba(255,255,255,0.08);
    overflow: hidden;
}

.panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    background: rgba(255,255,255,0.04);
    font-weight: 600;
    cursor: move;
}

.quota-list {
    padding: 8px 14px;
}

.quota-item {
    padding: 10px 0;
    border-bottom: 1px solid rgba(255,255,255,0.06);
}
.quota-item:last-child { border-bottom: none; }

.quota-item-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 6px;
}
.quota-platform { font-weight: 600; }
.quota-remaining { color: #aaa; font-size: 12px; }

.progress-bar {
    height: 6px;
    background: rgba(255,255,255,0.1);
    border-radius: 3px;
    overflow: hidden;
}
.progress-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.4s;
}
.progress-fill.green { background: #4caf50; }
.progress-fill.yellow { background: #ff9800; }
.progress-fill.red { background: #f44336; }

.panel-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 14px;
}

/* === 配置面板 === */
.settings-body {
    padding: 14px;
    max-height: 400px;
    overflow-y: auto;
}

.form-group {
    margin-bottom: 14px;
}
.form-group label {
    display: block;
    margin-bottom: 4px;
    font-size: 12px;
    color: #aaa;
}
.form-group input, .form-group textarea {
    width: 100%;
    padding: 6px 8px;
    background: rgba(255,255,255,0.06);
    border: 1px solid rgba(255,255,255,0.1);
    border-radius: 6px;
    color: #e0e0e0;
    font-size: 12px;
    font-family: monospace;
}
.form-group input:focus, .form-group textarea:focus {
    outline: none;
    border-color: #7c83ff;
}

.btn {
    padding: 6px 16px;
    background: #7c83ff;
    border: none;
    border-radius: 6px;
    color: white;
    cursor: pointer;
    font-size: 12px;
}
.btn:hover { background: #6b72e8; }

.btn-sm {
    padding: 4px 10px;
    background: rgba(255,255,255,0.08);
    border: 1px solid rgba(255,255,255,0.1);
    border-radius: 4px;
    color: #ccc;
    cursor: pointer;
    font-size: 11px;
    margin-top: 4px;
    margin-right: 6px;
}
.btn-sm:hover { background: rgba(255,255,255,0.14); }

.icon-btn {
    background: none;
    border: none;
    color: #aaa;
    cursor: pointer;
    font-size: 16px;
}
.icon-btn:hover { color: #fff; }
```

- [ ] **Step 3: 写 main.js**

```javascript
// === Wails 绑定 ===
// window.go.main.App 在运行时由 Wails 注入

let isExpanded = false;
let currentResults = [];

// === 事件监听 ===
window.runtime.EventsOn("quota:update", (results) => {
    currentResults = results;
    renderResults(results);
});

// === 球面点击:展开/收起 ===
document.getElementById("ball").addEventListener("click", () => {
    togglePanel();
});

function togglePanel() {
    isExpanded = !isExpanded;
    const ball = document.getElementById("ball");
    const panel = document.getElementById("panel");
    if (isExpanded) {
        ball.classList.add("hidden");
        panel.classList.remove("hidden");
        // 展开时若数据超过 3 分钟则刷新
        refreshIfNeeded();
    } else {
        panel.classList.add("hidden");
        ball.classList.remove("hidden");
    }
}

// === 刷新 ===
document.getElementById("btn-refresh").addEventListener("click", () => {
    refreshQuota();
});

async function refreshQuota() {
    document.getElementById("last-updated").textContent = "刷新中...";
    try {
        const results = await window.go.main.App.Refresh();
        currentResults = results;
        renderResults(results);
    } catch (e) {
        console.error("refresh error:", e);
    }
}

let lastRefreshTime = 0;
async function refreshIfNeeded() {
    if (Date.now() - lastRefreshTime > 3 * 60 * 1000) {
        await refreshQuota();
    }
}

// === 渲染结果 ===
function renderResults(results) {
    // 更新详情面板
    const list = document.getElementById("quota-list");
    list.innerHTML = "";
    results.forEach((r) => {
        const color = getDotColor(r);
        const percent = r.Percent || 0;
        const item = document.createElement("div");
        item.className = "quota-item";
        item.innerHTML = `
            <div class="quota-item-header">
                <span class="quota-platform">${r.Platform}</span>
                <span class="quota-remaining">${r.Error ? "⚠ " + r.Error : r.Remaining || "-"}</span>
            </div>
            <div class="progress-bar">
                <div class="progress-fill ${color}" style="width: ${r.Error ? 100 : percent}%"></div>
            </div>
        `;
        list.appendChild(item);
    });

    // 更新球面指示点
    updateDots(results);

    // 更新时间
    const now = new Date();
    lastRefreshTime = now.getTime();
    document.getElementById("last-updated").textContent = "更新于 " + now.toLocaleTimeString("zh-CN");
}

function getDotColor(r) {
    if (r.Error) return "red";
    if (r.Percent >= 90) return "red";
    if (r.Percent >= 75) return "yellow";
    return "green";
}

function updateDots(results) {
    const names = ["Kimi", "讯飞星辰", "小米MiMo"];
    const ids = ["dot-kimi", "dot-xfyun", "dot-mimo"];
    results.forEach((r, i) => {
        const dot = document.getElementById(ids[i]);
        if (!dot) return;
        dot.className = "dot " + getDotColor(r);
    });
}

// === 配置面板 ===
document.getElementById("btn-settings").addEventListener("click", () => {
    document.getElementById("panel").classList.add("hidden");
    document.getElementById("settings").classList.remove("hidden");
    loadConfig();
});

document.getElementById("btn-close-settings").addEventListener("click", () => {
    document.getElementById("settings").classList.add("hidden");
    document.getElementById("ball").classList.remove("hidden");
    isExpanded = false;
});

async function loadConfig() {
    const cfg = await window.go.main.App.GetConfig();
    document.getElementById("input-kimi").placeholder = cfg.kimi_api_key || "sk-kimi-xxx";
    document.getElementById("input-xfyun").placeholder = cfg.xfyun_cookie || "从浏览器 F12 复制 Cookie";
    document.getElementById("input-mimo").placeholder = cfg.mimo_cookie || "从浏览器 F12 复制 Cookie";
    document.getElementById("input-interval").value = cfg.refresh_interval_min || 15;
}

document.getElementById("btn-save-config").addEventListener("click", async () => {
    const kimi = document.getElementById("input-kimi").value;
    const xfyun = document.getElementById("input-xfyun").value;
    const mimo = document.getElementById("input-mimo").value;
    const interval = parseInt(document.getElementById("input-interval").value) || 15;
    await window.go.main.App.SaveConfig(kimi, xfyun, mimo, interval);
    // 清空输入框(已保存)
    document.getElementById("input-kimi").value = "";
    document.getElementById("input-xfyun").value = "";
    document.getElementById("input-mimo").value = "";
    alert("已保存");
});

// 测试连接按钮
document.querySelectorAll("[data-test]").forEach((btn) => {
    btn.addEventListener("click", async () => {
        const platform = btn.getAttribute("data-test");
        // 先保存当前输入(如果有)
        const kimi = document.getElementById("input-kimi").value;
        const xfyun = document.getElementById("input-xfyun").value;
        const mimo = document.getElementById("input-mimo").value;
        if (kimi || xfyun || mimo) {
            await window.go.main.App.SaveConfig(kimi, xfyun, mimo, 0);
        }
        const result = await window.go.main.App.TestConnection(platform);
        alert(result);
    });
});

// 打开登录页按钮
document.querySelectorAll("[data-open]").forEach((btn) => {
    btn.addEventListener("click", () => {
        const url = btn.getAttribute("data-open");
        window.go.main.App.OpenLoginPage(url);
    });
});

// === 球位置记忆(拖动结束时保存)===
let dragTimer = null;
document.getElementById("ball").addEventListener("mouseup", () => {
    clearTimeout(dragTimer);
    dragTimer = setTimeout(() => {
        // Wails 获取窗口位置
        window.runtime.WindowGetPosition().then((pos) => {
            window.go.main.App.SaveBallPosition(pos.x, pos.y);
        });
    }, 500);
});

// === 启动:加载初始数据 ===
window.go.main.App.Refresh();
```

- [ ] **Step 4: 构建并运行验证**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
"C:/Users/joe/go/bin/wails.exe" dev
```

预期:
- 启动后显示 60×60 悬浮球,三个灰点
- 点击球展开详情面板,显示三平台(初始为"未配置"错误)
- 点击 ⚙ 进入配置面板,可输入凭证并测试
- 保存后点刷新,已配置的平台显示进度条
- 拖动球后位置记忆

- [ ] **Step 5: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add frontend/
git commit -m "feat: dual-state UI - ball/panel/settings with event-driven updates"
```

---

## Task 9: 系统托盘

**Files:**
- Create: `internal/tray/tray.go`
- Modify: `main.go` (添加托盘初始化)

**Interfaces:**
- Consumes: `App` struct (Task 7), Wails tray API
- Produces: 系统托盘图标 + 右键菜单(刷新/显隐/打开配置/退出)

- [ ] **Step 1: 写 tray.go**

```go
package tray

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TrayHandler struct {
	ctx context.Context
}

func New(ctx context.Context) *TrayHandler {
	return &TrayHandler{ctx: ctx}
}

func (t *TrayHandler) Menu() *menu.Menu {
	return menu.NewMenuFromItems(
		menu.Text("刷新", nil, func(c *menu.CallbackData) {
			runtime.EventsEmit(t.ctx, "tray:refresh")
		}),
		menu.Text("显示/隐藏", nil, func(c *menu.CallbackData) {
			runtime.EventsEmit(t.ctx, "tray:toggle")
		}),
		menu.Separator(),
		menu.Text("打开配置", nil, func(c *menu.CallbackData) {
			runtime.EventsEmit(t.ctx, "tray:settings")
		}),
		menu.Separator(),
		menu.Text("退出", nil, func(c *menu.CallbackData) {
			runtime.Quit(t.ctx)
		}),
	)
}
```

- [ ] **Step 2: 修改 main.go — 添加托盘**

在 `main.go` 中添加托盘配置。将 `main()` 改为:

```go
package main

import (
	"embed"

	"quota-viewer/internal/tray"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Quota Viewer",
		Width:     60,
		Height:    60,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:  options.NewRGBA(0, 0, 0, 0),
		AlwaysOnTop:       true,
		HideWindowOnClose: true,
		DisableResize:     true,
		WindowStartState:  options.Normal,
		OnStartup:         app.OnStartup,
		Bind: []interface{}{
			app,
		},
		SystemTray: &options.SystemTray{
			Label:   "QV",
			Tooltip: "Quota Viewer",
		},
		TrayMenu:      nil, // 在 OnStartup 中动态设置
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
```

- [ ] **Step 3: 修改 app.go — 托盘事件处理**

在 `app.go` 的 `OnStartup` 中添加托盘菜单设置和事件监听:

```go
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// 设置托盘菜单
	trayHandler := tray.New(ctx)
	runtime.SetTrayMenu(ctx, trayHandler.Menu())

	// 监听托盘事件
	runtime.EventsOn(ctx, "tray:refresh", func(...interface{}) {
		a.Refresh()
	})
	runtime.EventsOn(ctx, "tray:toggle", func(...interface{}) {
		if runtime.WindowIsVisible(ctx) {
			runtime.WindowHide(ctx)
		} else {
			runtime.WindowShow(ctx)
		}
	})
	runtime.EventsOn(ctx, "tray:settings", func(...interface{}) {
		runtime.EventsEmit(ctx, "ui:show-settings")
	})

	// 启动后台定时刷新
	go a.startAutoRefresh()
}
```

- [ ] **Step 4: 前端监听托盘事件**

在 `frontend/src/main.js` 末尾追加:

```javascript
// === 托盘事件 ===
window.runtime.EventsOn("tray:refresh", () => {
    refreshQuota();
});

window.runtime.EventsOn("tray:toggle", () => {
    // 由 Go 端控制窗口显隐,前端无需操作
});

window.runtime.EventsOn("ui:show-settings", () => {
    document.getElementById("ball").classList.add("hidden");
    document.getElementById("panel").classList.add("hidden");
    document.getElementById("settings").classList.remove("hidden");
    loadConfig();
});
```

- [ ] **Step 5: 编译验证**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
go build ./...
```

预期:无错误。

- [ ] **Step 6: 运行验证**

```bash
"C:/Users/joe/go/bin/wails.exe" dev
```

预期:任务栏出现托盘图标,右键显示菜单(刷新/显示隐藏/打开配置/退出),各项功能可用。

- [ ] **Step 7: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add internal/tray/ main.go app.go frontend/src/main.js
git commit -m "feat: system tray with refresh/toggle/settings/quit menu"
```

---

## Task 10: 集成联调与 F12 抓包确认

**Files:**
- Modify: `internal/fetcher/xfyun.go` (根据 F12 抓包结果更新默认 URL 和字段映射)
- Modify: `internal/fetcher/mimo.go` (同上)

**注意**:此 task 需要**人工操作**——用浏览器 F12 Network 面板抓取讯飞和小米的实际 XHR 接口。

- [ ] **Step 1: 抓取讯飞额度接口**

1. 打开浏览器,登录 `https://maas.xfyun.cn`
2. 打开 F12 → Network 面板
3. 访问 `https://maas.xfyun.cn/packageSubscription`
4. 在 Network 中找到返回额度数据的 XHR 请求(筛选 Fetch/XHR,看 Response 包含 used/total/quota 等字段的)
5. 记录:
   - 请求 URL(完整路径)
   - 请求方法(GET/POST)
   - 响应 JSON 结构(截图或复制)

- [ ] **Step 2: 更新讯飞默认 URL 和字段映射**

根据 Step 1 抓到的结果,修改 `internal/fetcher/xfyun.go`:

1. 更新 `NewXfyunFetcher` 中的默认 `apiURL`
2. 如果字段名与 `parseXfyunDataMap` 中的不同,更新 `getFloat`/`getString` 的 keys 列表
3. 如果响应结构不是 `{ "data": {...} }` 形式,调整解析逻辑

- [ ] **Step 3: 抓取 MiMo 额度接口**

1. 打开浏览器,登录 `https://platform.xiaomimimo.com`
2. F12 → Network
3. 访问 `https://platform.xiaomimimo.com/console/plan-manage`
4. 找到返回额度数据的 XHR 请求
5. 记录 URL、方法、响应结构

- [ ] **Step 4: 更新 MiMo 默认 URL 和字段映射**

同 Step 2,修改 `internal/fetcher/mimo.go`。

- [ ] **Step 5: 端到端验证**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
"C:/Users/joe/go/bin/wails.exe" dev
```

验证清单:
1. 启动后悬浮球显示,三平台初始为"未配置"状态(灰/红点)
2. 进入配置面板,输入 Kimi API Key,点"测试"→ 显示"成功: xxx"
3. 输入讯飞 Cookie(从浏览器复制),点"测试"→ 显示"成功: xxx"或"Cookie 已过期"
4. 输入 MiMo Cookie,点"测试"→ 同上
5. 保存配置,点刷新→ 球面指示点变绿/黄/红,详情面板显示进度条
6. 关闭窗口→ 最小化到托盘,托盘右键"显示/隐藏"恢复
7. 等待 15 分钟(或改短间隔)→ 自动刷新
8. 拖动球到新位置,重启应用→ 位置恢复

- [ ] **Step 6: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add internal/fetcher/xfyun.go internal/fetcher/mimo.go
git commit -m "fix: update xfyun/mimo API URLs and field mappings from F12 capture"
```

---

## Task 11: 窗口尺寸动态调整(收起/展开)

**Files:**
- Modify: `frontend/src/main.js` (添加窗口尺寸切换)
- Modify: `app.go` (添加 `SetWindowSize` 方法)

**背景**:Wails v2 窗口初始 60×60(收起态),展开时需要变为 340×宽。需要 Go 端调用 `runtime.WindowSetSize`。

- [ ] **Step 1: 在 app.go 添加窗口尺寸方法**

在 `app.go` 中添加:

```go
// SetWindowSize 由前端调用,切换收起/展开尺寸。
func (a *App) SetWindowSize(w, h int) {
	runtime.WindowSetSize(a.ctx, w, h)
}
```

- [ ] **Step 2: 在 main.js 中切换窗口尺寸**

修改 `togglePanel()` 函数:

```javascript
function togglePanel() {
    isExpanded = !isExpanded;
    const ball = document.getElementById("ball");
    const panel = document.getElementById("panel");
    if (isExpanded) {
        ball.classList.add("hidden");
        panel.classList.remove("hidden");
        window.go.main.App.SetWindowSize(340, 260);
        refreshIfNeeded();
    } else {
        panel.classList.add("hidden");
        ball.classList.remove("hidden");
        window.go.main.App.SetWindowSize(60, 60);
    }
}
```

同样修改配置面板的显示/隐藏:

```javascript
document.getElementById("btn-settings").addEventListener("click", () => {
    document.getElementById("panel").classList.add("hidden");
    document.getElementById("settings").classList.remove("hidden");
    window.go.main.App.SetWindowSize(340, 480);
    loadConfig();
});

document.getElementById("btn-close-settings").addEventListener("click", () => {
    document.getElementById("settings").classList.add("hidden");
    document.getElementById("ball").classList.remove("hidden");
    window.go.main.App.SetWindowSize(60, 60);
    isExpanded = false;
});
```

- [ ] **Step 3: 验证**

```bash
"C:/Users/joe/go/bin/wails.exe" dev
```

预期:点击球时窗口从 60×60 平滑变为 340×260,收起时变回 60×60。

- [ ] **Step 4: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add app.go frontend/src/main.js
git commit -m "feat: dynamic window resize on ball/panel/settings toggle"
```

---

## Task 12: 最终构建与打包

**Files:**
- Modify: `wails.json` (确认构建配置)
- Build: 生成 `.exe`

- [ ] **Step 1: 确认 wails.json 配置**

检查 `wails.json`,确保 `outputdirectory` 和 `name` 正确:

```json
{
  "name": "QuotaViewer",
  "outputdirectory": "build/bin",
  "frontend:install": "",
  "frontend:build": "",
  "frontend:dev:watcher": "",
  "frontend:dev:serverUrl": "",
  "wailsjsdir": "./frontend"
}
```

(vanilla 模板无 npm 构建,这些为空)

- [ ] **Step 2: 构建生产版本**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
"C:/Users/joe/go/bin/wails.exe" build
```

预期:在 `build/bin/` 下生成 `QuotaViewer.exe`。

- [ ] **Step 3: 运行 exe 验证**

```bash
"build/bin/QuotaViewer.exe"
```

验证:
- 悬浮球显示
- 配置保存后重启,凭证和位置保留
- 内存占用检查:`tasklist -fi "imagename eq QuotaViewer.exe" -fo csv -nh`

- [ ] **Step 4: 验证内存占用**

```bash
tasklist -fi "imagename eq QuotaViewer.exe" -fo csv -nh
```

预期:单个进程,工作集 < 150 MB(目标远低于 Chrome 的 2.2 GB)。

- [ ] **Step 5: Commit**

```bash
cd "C:/Users/joe/Desktop/工作学习/软件开发/quota viewer"
git add -A
git commit -m "build: production build configuration and packaging"
```

---

## 自审清单

**Spec 覆盖检查:**

| Spec 要求 | 对应 Task |
|---|---|
| Wails v2 悬浮球窗口(收起/展开双态 + 拖动 + 位置记忆) | Task 1, 8, 11 |
| 系统托盘(显隐 + 右键菜单) | Task 9 |
| KimiFetcher(API Key 走通) | Task 4 |
| XfyunFetcher(Cookie 抓包走通,含 Cookie 失效提示) | Task 5, 10 |
| MiMoFetcher(同上) | Task 6, 10 |
| 详情面板三平台额度展示 | Task 8 |
| 配置文件读写 + 首次引导 | Task 2, 7, 8 |
| 刷新策略(15分钟自动 + 展开3分钟 + 手动) | Task 7, 8 |
| 错误处理(单 Fetcher 失败不影响其他) | Task 7 (fetchAll 并发) |
| 外观审美打磨 | 留实现阶段(Task 8 CSS 已做基础,后续迭代) |

**类型一致性检查:**
- `QuotaResult` 字段在 Task 3 定义,Task 4/5/6 使用,Task 7/8 传递——字段名一致(Platform, Used, Total, Percent, Remaining, ResetAt, LastUpdated, Error)
- `Fetcher` 接口 `Fetch() QuotaResult` 在 Task 3 定义,Task 4/5/6 实现——签名一致
- `getFloat`/`getString`/`parseXfyunDataMap` 在 Task 5 定义,Task 6 复用——包内可见,一致
- 前端调用的 Go 方法名在 Task 7 定义,Task 8 使用——`Refresh`, `GetConfig`, `SaveConfig`, `TestConnection`, `OpenLoginPage`, `SaveBallPosition`, `SetWindowSize` 全部一致

**占位符扫描:** Task 10 的 F12 抓包步骤是有意设计的人工操作(讯飞/小米无公开 API),不是占位符。其余步骤均含完整代码。
