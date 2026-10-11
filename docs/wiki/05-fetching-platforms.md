# 平台抓取器与注册表

**When to read**: 修改或新增某个平台的额度抓取逻辑时。

---

## 核心内容

### 统一接口

```go
const (
    KindUsage   = "usage"   // 用量型(默认):Percent = 已用百分比
    KindBalance = "balance" // 余额型(DeepSeek):Percent 无意义,Remaining 展示余额
)

type QuotaResult struct {            // internal/fetcher/types.go
    Platform    string    `json:"platform"`    // 展示名("Kimi" / "讯飞星辰" / ...)
    ID          string    `json:"id"`          // provider id(注册表)
    Abbr        string    `json:"abbr"`        // 球格缩写
    Kind        string    `json:"kind"`        // "usage"(默认) / "balance"
    Used        float64   `json:"used"`
    Total       float64   `json:"total"`       // 平台返回则填，否则 0
    Percent     float64   `json:"percent"`     // Used/Total*100; 无总量时由剩余百分比反推
    Balance     float64   `json:"balance"`     // 余额型原始余额数值(Kind=balance 时有效)
    Currency    string    `json:"currency"`    // 余额货币代码(如 "CNY" / "USD")
    Remaining   string    `json:"remaining"`   // 原始剩余描述（如 "1,200/18,000 次"）
    ResetAt     string    `json:"reset_at"`    // 下次重置时间 ISO8601，空则未知
    LastUpdated time.Time `json:"last_updated"`
    Error       string    `json:"error"`       // 非空表示失败
}

type Fetcher interface {
    Fetch() QuotaResult
}
```

### Provider 注册表（registry.go）

```go
type CredentialField struct {
    Key   string `json:"key"`   // 凭证 key
    Label string `json:"label"` // 显示名
    Type  string `json:"type"`  // "password" | "text" | "textarea"
}

type ProviderDef struct {
    ID          string
    DisplayName string
    Abbr        string // 球格缩写
    Kind        string // KindUsage | KindBalance(注册表标注,驱动前端渲染类型)
    LoginURL    string // 打开登录页按钮 URL(空 = 不显示)
    Fields      []CredentialField
    Build       func(creds map[string]string) Fetcher
}

func GetAll() []ProviderDef            // 固定顺序 = 推荐展示顺序
func Get(id string) (ProviderDef, bool)
```

### 平台实现

| id | 展示名 | 缩写 | 凭证字段 | 端点/认证 |
|---|---|---|---|---|
| `kimi` | Kimi | K | api_key (password) | Kimi 开放平台额度 API, Bearer |
| `xfyun` | 讯飞星辰 | 讯 | cookie (textarea) | 讯飞星辰 MaaS API, Cookie 头 |
| `opencode-go` | OpenCode Go | Go | workspace_id (text) + session_token (password) | `https://opencode.ai/workspace/{wsID}/go`, auth Cookie |
| `mimo` | 小米 MiMo | M | cookie (textarea) | `https://platform.xiaomimimo.com/api/v1/tokenPlan/usage`, Cookie + Referer |
| `deepseek` | DeepSeek | D | api_key (password) | `https://api.deepseek.com/user/balance`, Bearer |
| `ollama` | Ollama | O | cookie (textarea) | `https://ollama.com/settings`, Cookie 头,HTML 解析(无公开 quota API) |
| `command-code` | Command Code | C | api_key (password, 留空自动读 `~/.commandcode/auth.json`) | `https://api.commandcode.ai/alpha/*`, Bearer(CLI 私有路由) |
| `factory-droid` | Factory Droid | F | api_key (password, 留空自动读 `~/.factory/.env` 的 `FACTORY_API_KEY`) | `https://api.factory.ai/api/billing/limits`, Bearer + `x-factory-client: web-app`(web 私有路由) |
| `minimax` | MiniMax | MX | api_key (password, 必须是 `sk-cp-` 开头的 Token Plan 订阅 Key) | `https://api.minimaxi.com/v1/api/openplatform/coding_plan/remains`, Bearer + Referer(私有路由;三主机回退链 api.minimaxi.com → www.minimaxi.com → api.minimax.io) |
| `open-design` | Open Design | OD | cookie (textarea) | `https://amr-api.open-design.ai/api/v1/workspaces/current` + `/api/v1/workspaces/{id}/billing/coding-plan-usage`, Cookie 头(两步 GET;401/403/302 均按 Cookie 失效) |

- 每个 fetcher 的 `baseURL`/`apiURL` 可重写（构造时传空用默认）——测试通过该参数注入 httptest server
- OpenCode Go 抓取的是 Dashboard 页面（SSR hydration + data-slot 双模式解析）；`usagePercent` 为浮点（如 `3.1`），解析用 ParseFloat
- DeepSeek 是余额型：`Kind="balance"`，响应 `{"is_available":bool,"balance_infos":[{"currency","total_balance",...}]}`；取**第一条非零余额**的币种（如 USD $0.00 + CNY ¥247.51 → 显示 `余额 ¥247.51 (CNY)`）；`is_available=false` 或全部余额为 0 → Error；展示值经 ApplyBudget 按用户预算换算为消耗百分比（默认预算 300）
- Ollama Cloud 无公开 quota API（issue #15132），抓取 `ollama.com/settings` 页 HTML 解析 Session(5 小时)与 Weekly 用量百分比（详见 ADDING_A_PROVIDER.md 特殊说明）
- Command Code 无公开额度 API，`commandcode.go` 复用官方 CLI 的私有路由 `/alpha/whoami` + `/alpha/billing/credits`（Bearer 认证）；主展示 = 5 小时窗口用量，周窗口与剩余 credits 写入 `Remaining`，`ResetAt` 取 5 小时窗口 `resetAt`（epoch ms → UTC ISO）。结构随 CLI 升级可能变化，失效时对照官方 cli.mjs 更新
- Kimi 的 `Percent` 取 5h 与 7 天窗口较紧张者（周耗尽必须告警）；周用量优先读 `usages.limit_7d.used_ratio`，回退 `usage.used/limit`；`Used/Total/ResetAt` 仍记 5 小时窗口
- Ollama 抓取对慢响应放宽：30s 超时（ollama.com 偶发 >10s 慢响应）+ 传输层错误重试一次
- Factory Droid 无公开额度查询 API（官方 Analytics API 个人端点 Enterprise 限定且滞后一天），`factorydroid.go` 走官方 web 端同源私有路由：优先 token-rate limits 模型 `GET /api/billing/limits`（`usesTokenRateLimitsBilling=true` 时读 `limits.standard/core.{fiveHour,weekly,monthly}`，均只有 `usedPercent`/`secondsRemaining` 无绝对值；`extraUsageBalanceCents` 为预付 Extra Usage 余额），否则回退旧账单模型 `GET /api/organization/subscription/usage?useCache=true`（`usedRatio` 有恒 0 脏数据，绝对值可信时优先）。`Percent = max(全部窗口，含 Core)`；`Used/Total/ResetAt` 以 5h 主窗口为准（Total=100/Used=5h%，沿 Ollama 先例）；`Remaining` 常驻 5小时/周 两段（与 Kimi/Ollama 展示一致），月度仅在超过 5h/周 时追加，Core 仓一有消耗即显示，Extra 有余款才显示。计费顺序注意：Core 池模型的用量**先烧 Standard 窗口**（按模型倍率，如 GLM-5.3-Flash 0.06×），`limits.core` 独立窗口只在 Standard 顶满限流后接管（overage 回退，仅开源权重模型可走）。Key 在 app.factory.ai/settings/api-keys 生成（`fk-` 前缀），留空自动读 `~/.factory/.env`（支持 export 前缀/引号/行尾注释，与 Droid CLI 共用）。结构随官方升级可能变化（曾漂移过一次），失效时对照官方 web bundle 更新
- MiniMax 走开放平台私有路由 `GET {region}/v1/api/openplatform/coding_plan/remains`（Bearer + 官方 web 端同款 Referer），只认 `sk-cp-` 订阅 Key（按量计费 Key 读不到套餐额度）。三个主机组成回退链 `api.minimaxi.com` → `www.minimaxi.com` → `api.minimax.io`，逐个尝试到成功（网络不通/401/403/404/业务鉴权码都算该主机失败）；响应是套餐卡数组 `model_remains[]`，取首位主套餐。**坑位**：`current_interval_remaining_percent` / `current_weekly_remaining_percent` 是**剩余百分比**，展示前必须反转为已用%（`100 - remaining`），照名字直用会让球色刚好反过来（minimax-status 1.2.5 的语义事故，有回归测试守护）；部分套餐不返回绝对量（`current_*_total_count` 恒 0）→ 退化为纯百分比展示，周总额 0 且无周百分比 → 标「周无限制」。窗口语义与 Kimi/Ollama 一致：5 小时窗口驱动 `Used`/`ResetAt`（`remains_time` 毫秒倒计时），`Percent` 取两窗口更紧张者
- Open Design 无公开额度 API、也无 API-key 通道（`sk-amr-` Key 一律 401），只能用浏览器会话 Cookie。`opendesign.go` 打控制面 `amr-api.open-design.ai` 的 `/api/v1`（不是数据面 `amr-link.open-design.ai` 的 `/v1`），两步 GET：先 `/api/v1/workspaces/current` 取 `workspaceId`，再 `/api/v1/workspaces/{id}/billing/coding-plan-usage` 取窗口用量。坑位：第二步不能拿 `current` 当 id（400 `workspace_member_required`）；窗口按 `durationSeconds` 识别而不按下标（18000 = 5 小时、604800 = 7 天）；`usedCredits`/`limitCredits` 是十进制字符串，用 `strconv.ParseFloat` 算百分比（credits 可能超出 float64 精确整数范围，不转 int64）。5 小时窗口驱动 `Used`/`ResetAt`（`resetsAt` 官方毫秒精度，规范化为秒精度 UTC），`Percent` 取两窗口更紧张者，两窗口都写进 `Remaining`。Cookie 失效可能是 401 也可能是 302 跳登录页，故 client 设 `CheckRedirect: http.ErrUseLastResponse` 并把 302 并入失效分支（否则跟随跳转拿到 200 登录页 HTML，报成「解析响应失败」而丢掉「请重新登录」指引）
- `format.go` 的 `formatNum` 做千分位展示格式化（仅内部使用）

### 预算换算（budget.go）

余额型 Provider 没有"用量百分比"语义，`fetchAll` 在 Fetch 后调用 `ApplyBudget(r, budget)` 换算：

- `budget <= 0` → 用 `defaultBudget=300`
- 已消耗 = 预算 - 余额（余额超预算钳制为 0）；`Percent = 已消耗 / 预算 * 100`
- 覆盖 `Total/Used/Percent/Remaining`（如 `¥247.51 / ¥500.00 (预算)`）
- 非余额型 / 余额为负 / 结果有 Error → 直接返回不改动

### 失败语义

- 任何平台出错（网络/鉴权/解析）→ `QuotaResult.Error` 非空，其余字段尽力填充
- 调用方（app.go）不 panic：`fetchAll` 并发收集，前端按 Error 显示"失败"状态
- 未配置凭证 → 各 fetcher 自行返回带 Error 的结果（不 panic）

### 测试模式

全部用 `net/http/httptest` 起假服务，`baseURL` 指向假服务：
- `kimi_test.go` / `xfyun_test.go` / `opencode_go_test.go` / `mimo_test.go` / `deepseek_test.go` / `ollama_test.go` / `commandcode_test.go` / `factorydroid_test.go` / `minimax_test.go` / `opendesign_test.go` 覆盖成功/失败/异常响应路径
- `registry_test.go` 校验注册表完整性（10 个、顺序、字段定义、Build 可执行；空凭证保持离线）

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `internal/fetcher/types.go` | QuotaResult + Fetcher 接口 + Kind 常量（契约核心） |
| `internal/fetcher/registry.go` | ProviderDef + 注册表（新增 Provider 的唯一入口） |
| `internal/fetcher/kimi.go` / `xfyun.go` / `opencode_go.go` / `mimo.go` / `deepseek.go` / `ollama.go` / `commandcode.go` / `factorydroid.go` / `minimax.go` / `opendesign.go` | 各平台实现 |
| `internal/fetcher/format.go` | 千分位格式化 |
| `internal/fetcher/budget.go` | 余额型预算 → 消耗百分比换算(ApplyBudget) |

---

## Must NOT Change

- `QuotaResult` JSON 字段名（前端渲染契约）
- Fetcher 接口签名 `Fetch() QuotaResult`
- Provider id 与注册表顺序（config 存储、TestConnection、前端绑定共用契约）
- fetchAll 的 results 顺序 = 启用 Provider 的注册表顺序（前端球格顺序绑定）
