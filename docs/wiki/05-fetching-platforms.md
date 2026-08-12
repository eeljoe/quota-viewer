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

- 每个 fetcher 的 `baseURL`/`apiURL` 可重写（构造时传空用默认）——测试通过该参数注入 httptest server
- OpenCode Go 抓取的是 Dashboard 页面（SSR hydration + data-slot 双模式解析）
- DeepSeek 是余额型：`Kind="balance"`，响应 `{"is_available":bool,"balance_infos":[{"currency","total_balance",...}]}`；取**第一条非零余额**的币种（如 USD $0.00 + CNY ¥247.51 → 显示 `余额 ¥247.51 (CNY)`）；`is_available=false` 或全部余额为 0 → Error；展示值经 ApplyBudget 按用户预算换算为消耗百分比（默认预算 300）
- Ollama Cloud 无公开 quota API（issue #15132），抓取 `ollama.com/settings` 页 HTML 解析 Session(5 小时)与 Weekly 用量百分比（详见 ADDING_A_PROVIDER.md 特殊说明）
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
- `kimi_test.go` / `xfyun_test.go` / `opencode_go_test.go` / `mimo_test.go` / `deepseek_test.go` / `ollama_test.go` 覆盖成功/失败/异常响应路径
- `registry_test.go` 校验注册表完整性（6 个、顺序、字段定义、Build 可执行；空凭证保持离线）

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `internal/fetcher/types.go` | QuotaResult + Fetcher 接口 + Kind 常量（契约核心） |
| `internal/fetcher/registry.go` | ProviderDef + 注册表（新增 Provider 的唯一入口） |
| `internal/fetcher/kimi.go` / `xfyun.go` / `opencode_go.go` / `mimo.go` / `deepseek.go` / `ollama.go` | 各平台实现 |
| `internal/fetcher/format.go` | 千分位格式化 |
| `internal/fetcher/budget.go` | 余额型预算 → 消耗百分比换算(ApplyBudget) |

---

## Must NOT Change

- `QuotaResult` JSON 字段名（前端渲染契约）
- Fetcher 接口签名 `Fetch() QuotaResult`
- Provider id 与注册表顺序（config 存储、TestConnection、前端绑定共用契约）
- fetchAll 的 results 顺序 = 启用 Provider 的注册表顺序（前端球格顺序绑定）
