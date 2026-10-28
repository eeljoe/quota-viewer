# 新增 Provider 指南（Agent 专用）

> 本文档供 AI agent 阅读。目标：添加一个新的 AI 平台额度抓取器，完成后 `go test ./...` 全绿、前端自动适配、无需改动 UI 代码。

---

## 架构概述

项目采用**注册表驱动**架构。新增 Provider 只需两步代码 + 一步测试：

1. **实现 Fetcher** — 在 `internal/fetcher/` 新增一个 `.go` 文件
2. **注册** — 在 `internal/fetcher/registry.go` 的 `registry` 数组追加一条
3. **测试** — 同目录新增 `_test.go`，用 `httptest` 覆盖

前端配置面板和球格通过 `app.go` → `GetConfig()` 返回的注册表元数据动态渲染，**不需要任何前端改动**。

---

## Step 1：实现 Fetcher

在 `internal/fetcher/` 创建 `yourplatform.go`。参考现有实现（`kimi.go` 用量型 / `deepseek.go` 余额型）。

### 1.1 结构体与构造函数

```go
package fetcher

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// YourPlatformFetcher 通过 XXX 查询额度。
// 端点: GET https://api.example.com/quota
// 认证: Authorization: Bearer xxx  (或 Cookie 头)
type YourPlatformFetcher struct {
    apiKey string       // 或 cookie / token，取决于平台
    apiURL string       // 可为空，默认用线上端点（测试时注入 httptest server）；例外：devin 留空时先试本机读到的端点，见文末「Devin 特殊说明」
}

func NewYourPlatformFetcher(apiKey string) *YourPlatformFetcher {
    return &YourPlatformFetcher{apiKey: apiKey}
}
```

**关键点**：
- `apiURL` 字段必须保留且可为空——测试通过它注入 httptest server URL
- 构造函数只接收凭证字符串，`apiURL` 留空走默认端点 —— **唯一例外是 `devin`**：留空时先试本机读到的 `apiServerUrl`，读不到才回落到默认端点（见文末「Devin 特殊说明」）

### 1.2 实现 Fetch() 方法

先分清一下**这家平台一个凭证回来几个独立额度池**：

- **只有一个**（大多数平台）—— 实现 `Fetch() QuotaResult` 即可，下面的步骤照做。
- **有多个**（如 MiniMax 的主套餐 + 赠送卡、MiMo 的套餐额度 + 补偿额度）——
  额外实现 `FetchAll() []QuotaResult`，**每个额度池一条记录**（说人话：一个额度池 = 一个监控对象 = 一格球），
  并把套餐名写进 `SubLabel`、每卡的短缩写写进 `Abbr`（主卡留给 Provider 缩写即可）。
  只有一个额度池的平台**不要**实现 `FetchAll`——上层 `fetcher.FetchAll` 会自动包成单条。
  参考实现：`internal/fetcher/minimax.go`、`internal/fetcher/mimo.go`。

```go
func (f *YourPlatformFetcher) Fetch() QuotaResult {
    result := QuotaResult{
        Platform:    "YourPlatform",   // 展示名（与注册表 DisplayName 一致）
        Kind:        KindUsage,        // 用量型显式设 KindUsage；余额型改 KindBalance
        LastUpdated: time.Now(),
    }

    // 1. 凭证空检查
    if f.apiKey == "" {
        result.Error = "未配置 API Key"
        return result
    }

    // 2. 确定 URL（测试覆盖的关键）
    url := f.apiURL
    if url == "" {
        url = "https://api.example.com/quota"
    }

    // 3. 构建请求
    client := &http.Client{Timeout: 10 * time.Second}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        result.Error = fmt.Sprintf("创建请求失败: %v", err)
        return result
    }
    req.Header.Set("Authorization", "Bearer "+f.apiKey)
    req.Header.Set("Accept", "application/json")

    // 4. 发送请求
    resp, err := client.Do(req)
    if err != nil {
        result.Error = fmt.Sprintf("请求失败: %v", err)
        return result
    }
    defer resp.Body.Close()

    // 5. 状态码处理（401/403 → 凭证无效；非 200 → 通用错误）
    if resp.StatusCode == 401 || resp.StatusCode == 403 {
        result.Error = "API Key 无效或已过期"
        return result
    }
    if resp.StatusCode != 200 {
        result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
        return result
    }

    // 6. 解析 JSON 响应
    var body yourResponseStruct
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
        result.Error = fmt.Sprintf("解析响应失败: %v", err)
        return result
    }

    // 7. 填充 QuotaResult 字段
    //    用量型：填 Used / Total / Percent / Remaining / ResetAt
    //    余额型：只填 Remaining（如 "余额 ¥247.51 (CNY)"），Percent 留 0
    result.Used = body.Used
    result.Total = body.Total
    if body.Total > 0 {
        result.Percent = body.Used / body.Total * 100
    }
    result.Remaining = fmt.Sprintf("%s / %s", formatNum(body.Used), formatNum(body.Total))
    result.ResetAt = body.ResetAt

    return result
}
```

### 1.3 两种 Provider 类型

| 类型 | `Kind` 值 | 语义 | 填充字段 | 前端表现 |
|------|-----------|------|----------|----------|
| **用量型**（默认） | `KindUsage`（`"usage"`） | Percent = 已用百分比 | `Used` / `Total` / `Percent` / `Remaining` / `ResetAt` | 球格颜色随 Percent 变化（绿/黄/红） |
| **余额型** | `KindBalance` | 余额展示，无百分比 | 仅 `Remaining`（如 `余额 ¥247.51 (CNY)`） | 球格恒绿 |

用量型必须显式设 `Kind: KindUsage`（注册表条目的 `Kind` 字段同样填 `KindUsage`）——不要留零值 `""`，那是历史写法，与 `types.go` 的常量及 `minimax.go` 等先例不一致。余额型必须设 `Kind: KindBalance`。

### 1.4 Cookie 类凭证

如果平台用 Cookie 认证（如讯飞、MiMo），参考 `xfyun.go`：

- 凭证字段类型用 `textarea`（注册表里指定），用户可粘贴浏览器 "Copy as PowerShell" 格式
- `config/cookie.go` 会在保存时自动从 PowerShell 脚本中提取 Cookie 字符串
- Fetcher 里直接用 `req.Header.Set("Cookie", cookie)` 即可，不需要自己解析
- Cookie 过期时平台通常返回 302 跳转登录页，需禁止自动重定向：

```go
client := &http.Client{
    Timeout: 10 * time.Second,
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    },
}
// 然后把 302 当作 Cookie 过期处理
if resp.StatusCode == 401 || resp.StatusCode == 302 {
    result.Error = "Cookie 已过期，请更新"
    return result
}
```

### 1.5 formatNum 工具函数

`format.go` 中的 `formatNum(float64) string` 做千分位格式化，构建 `Remaining` 字符串时直接调用。

---

## Step 2：注册到 registry.go

打开 `internal/fetcher/registry.go`，在 `registry` 数组末尾追加：

```go
{
    ID:          "your-platform",          // 唯一 ID（小写中划线，用于 config 存储和前端绑定）
    DisplayName: "YourPlatform",           // 展示名（与 Fetch() 里的 Platform 字段一致）
    Abbr:        "Y",                      // 球格缩写（1-2 个字符）
    LoginURL:    "https://example.com",    // 登录页 URL（空字符串 = 不显示"打开登录页"按钮）
    Fields: []CredentialField{
        {Key: "api_key", Label: "API Key", Type: "password"},
        // 多个凭证字段示例：
        // {Key: "workspace_id", Label: "Workspace ID", Type: "text"},
        // {Key: "session_token", Label: "Session Token", Type: "password"},
        // {Key: "cookie", Label: "Cookie(浏览器 F12 复制)", Type: "textarea"},
    },
    Build: func(creds map[string]string) Fetcher {
        return NewYourPlatformFetcher(creds["api_key"])
        // 多字段：NewYourPlatformFetcher(creds["workspace_id"], creds["session_token"])
    },
},
```

### 字段说明

| 字段 | 说明 | 约束 |
|------|------|------|
| `ID` | Provider 唯一标识 | 小写中划线，全局唯一，不可变更（config 存储、TestConnection、前端共用） |
| `DisplayName` | 配置面板和 tooltip 展示名 | 与 `Fetch()` 返回的 `Platform` 字段一致 |
| `Abbr` | 球格内显示的缩写字母 | 1-2 字符（如 `K` / `讯` / `Go`） |
| `LoginURL` | 配置面板"打开登录页"按钮的 URL | 空字符串 = 不显示按钮 |
| `Fields` | 凭证字段定义，驱动前端动态生成输入框 | 每个字段有 `Key` / `Label` / `Type` |
| `Build` | 工厂函数，从 `map[string]string` 构造 Fetcher | Key 与 Fields 中的 `Key` 对应 |

### CredentialField.Type 取值

| Type | 前端渲染 | 用途 |
|------|----------|------|
| `"password"` | 密码输入框（掩码） | API Key、Token |
| `"text"` | 普通文本框 | Workspace ID 等非敏感 ID |
| `"textarea"` | 多行文本框 | Cookie（可粘贴整段 PowerShell 脚本） |

### 注册表顺序 = 展示顺序

`registry` 数组的顺序决定了配置面板中的 Provider 列表顺序和球格默认排列顺序。新 Provider 追加到数组末尾即可。

---

## Step 3：编写测试

在同目录创建 `yourplatform_test.go`。**全部使用 `net/http/httptest`**，不发起真实网络请求。

### 测试模板

```go
package fetcher

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

// 1. 凭证为空 → 返回 Error
func TestYourPlatform_EmptyKey_ReturnsError(t *testing.T) {
    f := NewYourPlatformFetcher("")
    result := f.Fetch()
    if result.Error == "" {
        t.Error("expected error for empty key")
    }
}

// 2. 正常响应 → 解析成功
func TestYourPlatform_OK_ParsesQuota(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 验证认证头
        if auth := r.Header.Get("Authorization"); auth != "Bearer sk-test" {
            t.Errorf("expected Bearer sk-test, got %s", auth)
        }
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"used": 5000, "total": 18000, "reset_at": "2026-12-31T23:59:59Z"}`))
    }))
    defer server.Close()

    f := NewYourPlatformFetcher("sk-test")
    f.apiURL = server.URL   // 关键：注入 httptest server URL
    result := f.Fetch()
    if result.Error != "" {
        t.Fatalf("unexpected error: %s", result.Error)
    }
    if result.Total != 18000 {
        t.Errorf("expected Total=18000, got %f", result.Total)
    }
    if !strings.Contains(result.Remaining, "5,000") {
        t.Errorf("expected formatted number in Remaining, got %s", result.Remaining)
    }
}

// 3. 401 → 凭证无效
func TestYourPlatform_401_ReturnsInvalidKey(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(401)
    }))
    defer server.Close()

    f := NewYourPlatformFetcher("sk-test")
    f.apiURL = server.URL
    result := f.Fetch()
    if !strings.Contains(result.Error, "无效") {
        t.Errorf("expected invalid key error, got %s", result.Error)
    }
}

// 4. 异常 JSON → 返回解析错误
func TestYourPlatform_BadJSON_ReturnsError(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`not json`))
    }))
    defer server.Close()

    f := NewYourPlatformFetcher("sk-test")
    f.apiURL = server.URL
    result := f.Fetch()
    if result.Error == "" {
        t.Error("expected error for bad JSON")
    }
}
```

### 测试要点

- **必须覆盖的路径**：凭证为空、正常响应、401/403、非 200、异常 JSON
- **httptest 模式**：`httptest.NewServer` 起假服务 → `f.apiURL = server.URL` 注入 → `defer server.Close()`
- **不要**发真实网络请求（不依赖外部服务可达性）
- 余额型额外覆盖：`is_available=false`、全部币种余额为 0、多币种选非零

---

## Step 4：验证

```bash
# 运行全部测试
go test ./...

# 构建
cd frontend && npm run build && cd ..
wails build
```

确认：
- `go test ./...` 全绿（包括 `registry_test.go` 的注册表完整性检查——它会校验新 Provider 的字段定义和 Build 可执行性）
- `npm run build` 成功（前端无需改动，但确认构建无碍）
- `wails build` 成功

---

## 完整检查清单

- [ ] `internal/fetcher/yourplatform.go` — Fetcher 结构体 + 构造函数 + `Fetch() QuotaResult`（一多度池平台加 `FetchAll() []QuotaResult`）
- [ ] `internal/fetcher/registry.go` — `registry` 数组追加新条目
- [ ] `internal/fetcher/yourplatform_test.go` — httptest 测试（空凭证 / 正常 / 401 / 异常 JSON）
- [ ] `go test ./...` 全绿
- [ ] `wails build` 成功

**不需要改动的文件**：
- `app.go` — `GetConfig()` / `SaveConfig()` / `fetchAll()` / `TestConnection()` 全部通过注册表动态工作
- `frontend/src/*` — 配置面板和球格根据注册表元数据自动渲染
- `internal/config/config.go` — `Providers` 列表动态存储任意 Provider ID

---

## 现有 Provider 速查

| id | 展示名 | 缩写 | 认证方式 | Kind | 参考文件 |
|----|--------|------|----------|------|----------|
| `kimi` | Kimi | K | API Key (Bearer) | usage | `kimi.go` |
| `xfyun` | 讯飞星辰 | 讯 | Cookie | usage | `xfyun.go` |
| `opencode-go` | OpenCode Go | Go | API Key (Bearer, `oc_sk-...`；留空自动读 `~/.local/share/opencode/auth.json`) | usage | `opencode_go.go` |
| `mimo` | 小米 MiMo | M | Cookie | usage | `mimo.go` |
| `deepseek` | DeepSeek | D | API Key (Bearer) | balance | `deepseek.go` |
| `ollama` | Ollama | O | Browser session Cookie | usage | `ollama.go` |
| `command-code` | Command Code | C | API Key (Bearer, `user_...`) | usage | `commandcode.go` |
| `factory-droid` | Factory Droid | F | API Key (Bearer, `fk-...`；留空自动读 `~/.factory/.env`) | usage | `factorydroid.go` |
| `minimax` | MiniMax | MX | Token Plan Key (`sk-cp-` 订阅 Key) | usage | `minimax.go` |
| `open-design` | Open Design | OD | Browser session Cookie | usage | `opendesign.go` |
| `devin` | Devin | DV | 本机 Devin Desktop 登录态（留空自动读）或手填 token | usage | `devin.go` |

### Ollama Cloud 特殊说明

Ollama Cloud 当前没有公开的 quota API。`ollama.go` 请求 `GET https://ollama.com/settings`，解析页面中的 `Session usage`（5 小时窗口）作为 `QuotaResult.Percent`，并将 `Weekly usage` 作为辅助信息写入 `Remaining`。重置时间从对应 HTML 元素的 `data-time` 属性读取；页面未登录、结构变化或 Cookie 失效时返回错误。注册表字段使用 `textarea`，用户可在 Ollama 设置页登录后复制包含 `wos-session` 或 `__Secure-session` 的 Cookie，或直接粘贴浏览器的 “Copy as PowerShell” 内容。

### Command Code 特殊说明

Command Code（commandcode.ai，AI 编码订阅）无公开额度查询 API。`commandcode.go` 逆向官方 CLI（npm 包 `command-code` 的 `dist/cli.mjs`），复用其私有路由：`GET https://api.commandcode.ai/alpha/whoami` 取 `orgId`，再 `GET /alpha/billing/credits`（团队用户带 `?orgId=`）取月度剩余 credits 与 5 小时/周滚动窗口。认证 `Authorization: Bearer <API Key>`；Key 在 Studio 生成（`user_...`），与官方 CLI 写入 `~/.commandcode/auth.json` 的 `apiKey` 相同——注册表 API Key 字段留空时自动读取该文件（用户已登录官方 CLI 则免填）。主展示 = 5 小时窗口用量（驱动球色），月度剩余与周窗口写入 `Remaining`，`ResetAt` 取 5 小时窗口的 `resetAt`（epoch 毫秒 → UTC ISO）。该接口为 CLI 内部路由，结构可能随 CLI 升级变化，失效时对照官方 cli.mjs 更新 `commandcode.go`。

### MiniMax 特殊说明

MiniMax 的 Token Plan 额度走开放平台私有路由 `GET {region}/v1/api/openplatform/coding_plan/remains`，认证 `Authorization: Bearer <Token Plan Key>`——必须是 `sk-cp-` 开头的订阅 Key（在 platform.minimaxi.com/console/plan 生成），按量计费的普通 API Key 读不到套餐额度。三个主机组成回退链，`minimax.go` 按 `api.minimaxi.com` → `www.minimaxi.com` → `api.minimax.io` 逐个尝试直到成功：实测（2026-10）`api.minimaxi.com` 对国内 Key 可用，`www.minimaxi.com` 偶发 TLS 握手超时，海外主机不认国内 Key（业务码 1004）；网络不通、HTTP 401/403/404、业务鉴权码都算该主机失败，全部失败时报第一个错误。响应是套餐卡数组 `model_remains[]`（general 主套餐在首位，视频/音乐等赠送卡在后），本实现取首位。**关键坑位**：字段是 `current_interval_remaining_percent` / `current_weekly_remaining_percent`，语义为**剩余百分比**，展示前必须反转为已用%（`100 - remaining`），照名字直用会让球色刚好反过来（minimax-status 1.2.5 的语义事故，本仓有回归测试守护）；部分套餐不返回绝对量（`current_*_total_count` 恒 0），此时退化为纯百分比展示。窗口语义与 Kimi/Ollama 一致：5 小时窗口驱动主展示与 `ResetAt`，`Percent` 取 5 小时与周窗口中更紧张者。该接口为 web 端私有路由，结构可能随官方升级变化，失效时对照 minimax-status 源码更新 `minimax.go`。

### Open Design 特殊说明

Open Design（open-design.ai，AMR Link 中转控制面）无公开额度 API，且没有可用的 API-key 通道：`sk-amr-` 形式的 API key 读不到任何额度，控制面对它一律返回 401，只能用**浏览器会话 Cookie**（注册表字段用 `textarea`，用户在 open-design.ai/cloud/dashboard 登录后按 F12 复制，或直接粘贴浏览器 “Copy as PowerShell” 内容，`config/cookie.go` 会自动提取）。`opendesign.go` 请求的是控制面 `amr-api.open-design.ai` 的 `/api/v1` 私有路由（不是数据面 `amr-link.open-design.ai` 的 `/v1`），两步 GET：先 `GET /api/v1/workspaces/current` 取 `workspaceId`，再 `GET /api/v1/workspaces/{workspaceId}/billing/coding-plan-usage` 取窗口用量。坑位：第二步**不能拿 `current` 当 id 用**（会返回 400 `workspace_member_required`），必须先取真实 id；窗口按 `durationSeconds` 识别而不按数组下标（18000 = 5 小时，604800 = 7 天，窗口集合归后端管，官方前端同样按 durationSeconds 匹配）；`usedCredits` / `limitCredits` / `remainingCredits` 是十进制字符串，按 `ParseFloat` 算百分比（credits 计数可能超出 float64 精确整数范围，不要转 int64）。5 小时窗口驱动 `Used` / `ResetAt`（`resetsAt` 官方给毫秒精度，规范化为秒精度 UTC），`Percent` 取 5 小时与 7 天窗口中更紧张者，两个窗口都写进 `Remaining`。Cookie 失效时控制面可能返 401 也可能 302 跳登录页，故 client 设了 `CheckRedirect: http.ErrUseLastResponse` 并把 302 并入 Cookie 失效分支——否则会跟随跳转拿到 200 登录页 HTML，报成「解析响应失败(响应结构可能已变化)」而丢掉「请重新登录」指引。该路由是官方前端的私有接口，无稳定性承诺，失效时对照 open-design.ai 的 web bundle 更新 `opendesign.go`。

### Devin 特殊说明

Devin 没有可用的公开额度 API：v2/v3 的 `/enterprise/consumption/*` 只对 Enterprise 管理员的个人 Key 开放，self-serve（Free / Pro / Max / Teams）一律 403，官方文档明确写了没有可以放开的开关。能读到日/周两个窗口的只有 Devin Desktop（Windsurf 内核）自己调用的席位管理接口：`POST {server}/exa.seat_management_pb.SeatManagementService/GetUserStatus` —— CONNECT 协议，请求头 `Content-Type: application/json` 与 `Connect-Protocol-Version: 1`，UA 用浏览器型固定串；body 只有 `metadata`，`metadata.api_key` 传 token，`ide_name` / `extension_name` 填 `windsurf`，`ide_version` / `extension_version` 填实测可用的 `1.126.0`（接口会校验版本号），`session_id` 任意占位。

响应在 `userStatus.planStatus` 里，用到的字段（类型为 2026-10-29 真机直连实测）：

| 字段 | 实测类型 | 语义 |
|---|---|---|
| `dailyQuotaRemainingPercent` | number | 日窗口**剩余**百分比 |
| `weeklyQuotaRemainingPercent` | number | 周窗口**剩余**百分比 |
| `dailyQuotaResetAtUnix` | **string**（protojson int64） | 日窗口重置时间，epoch 秒 |
| `weeklyQuotaResetAtUnix` | **string**（protojson int64） | 周窗口重置时间，epoch 秒 |
| `overageBalanceMicros` | **string** int64 微美元；可整个字段缺失（本机响应即缺失） | 超额余额 |
| `planInfo.hideDailyQuota` / `planInfo.hideWeeklyQuota` | bool；可整个字段缺失（本机套餐两者都缺失） | 为真则该窗口不可见 |
| `planInfo.planName` | string（在 `planInfo` 里，**不在** `planStatus` 顶层） | `Pro` / `Free` / `Max` …，只解析不展示 |

口径：响应给的是**剩余**百分比，展示前必须反转成已用（`100 − 剩余`，夹在 `[0,100]`），与官方前端 `tgY` / `tgK` 两个换算函数同口径；照名字直用会让球色刚好反过来。`Percent` 与 `ResetAt` 都由**最紧张的可见窗口**驱动，倒计时标明「日额度 / 周额度」，被隐藏的窗口既不参与百分比也不参与倒计时。`planInfo.hide*Quota` 为真、**或本次响应干脆没给该窗口的剩余百分比**，两种都按不可见处理（缺字段不能当 0% 已用，那会把额度渲染成全新）；两个窗口都不可见时报「当前套餐不提供可读的额度窗口」。套餐名不展示（卡片头部只有 210px，多一段会把最紧张的数字挤出可视区），`overageBalanceMicros > 0` 时追加到 `Remaining` 尾部。三个 int64 字段无论给字符串还是数字都要能解析，百分比给字符串也不该报错。

凭据：`api_key` 手填优先；留空时读本机 Devin Desktop 登录态 —— `%APPDATA%\Devin\User\globalStorage\state.vscdb` 里 `devin-session-token$<jwt>` 与 `"apiServerUrl":"…"` 都是**明文**，直接按字节扫，不引 SQLite 驱动、也不解析库结构（实测文件约 1.5 MB，毫秒级）。扫描是**按字节匹配、每种字面量只取第一个命中点**（`apiServerUrl` 亦然）：Devin Desktop 一旦改成加密存储（文件里没有明文）、把凭据字面量改名、或挪了 `state.vscdb` 位置，扫描就什么都找不到，两个值都退化成空串 —— 用户看到的是「未配置 Devin token」（错误文案里点了「登录 Devin Desktop 后留空即可」这个前提），不是「读取失败」，此时改走手填。**多命中形态不属于这一类**：`bytes.Index` 只认文件里第一个 `devin-session-token$`，后面的命中点一律不采用，所以 token 非空 —— 用户看到的是 401/403「Devin token 已失效，请重新登录 Devin Desktop 或更新 token」，或者（首个命中那个 token 还没作废时）拿到它对应的额度，未必是当前登录态那份。反过来，**第一个命中点若是裸前缀**（`devin-session-token$` 后面没有 `[A-Za-z0-9_.\-]` 的 payload），实现判为未命中、token 返回空串，后面即使还有完整命中也读不到 —— 同样落到「未配置 Devin token」。读文件要按 Go 的共享模式（`os.ReadFile` 即是）：Devin Desktop 运行时会持有该库，独占读会失败。

server 有三级兜底：调用方注入的 server → 本机读到的 `apiServerUrl` → 默认 `https://server.codeium.com`。**注意「`apiURL` 留空 = 默认端点」这条通用说明对 Devin 不成立**：`NewDevinFetcher` 不预填默认端点，留空会先试本机端点，读不到才回落默认值；默认值只允许出现在 URL 拼装那一处，提前写进构造期字段会让本机值永远失效。

错误映射：401/403 走专属文案「Devin token 已失效，请重新登录 Devin Desktop 或更新 token」；**400 没有专属分支**，与其它所有非 200 一样原样透出 `HTTP N`（用户看到的是 `HTTP 400`；该状态码的服务端语义是缺 `api_key`，只是实现不单独区分，500 等同样走这条兜底）；结构不符报「无法解析额度数据（响应结构可能已变化）」。该路由是 Devin Desktop 的私有接口，无稳定性承诺；失效时对照 Devin Desktop 的 `resources/app/node_modules/@exa/chat-client/index.js`（`GetUserStatus` / `PlanStatus` 类型定义）更新 `devin.go`。
