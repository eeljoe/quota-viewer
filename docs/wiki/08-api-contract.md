# API 契约（Wails 绑定）

**When to read**: 前端调用后端方法、修改绑定签名、排查 `window.go.main.App.*` 调用时。

---

## 核心内容

前端通过 `frontend/wailsjs/go/main/App.*` 调用（构建时自动生成，签名随 app.go 方法变化）。

### App 导出方法（Go 侧签名）

```go
// 刷新: 并发抓取启用的 Provider(1-3), 写缓存, emit "quota:update", 返回结果
func (a *App) Refresh() []fetcher.QuotaResult

// 读配置: 全部 Provider 元数据 + 掩码凭证
func (a *App) GetConfig() map[string]interface{}

// 存配置: 动态 Provider 列表(见下); 空字符串凭证 = 不修改; 数量钳制 1-3
func (a *App) SaveConfig(providers []ProviderInput, refreshMin int) error

// 测试单平台连接: platform ∈ 注册表 id; 返回 "成功: ..." 或 "失败: ..."
func (a *App) TestConnection(platform string) string

// 用默认浏览器打开 URL(打开登录页)
func (a *App) OpenLoginPage(url string)

// 保存悬浮球位置(收起时前端记录)
func (a *App) SaveBallPosition(x, y int) error

// 展开面板: w/h 为逻辑像素; 内部做 fitToScreen 定位
func (a *App) ExpandWindow(w, h int)

// 收起为悬浮球并还原位置
func (a *App) CollapseWindow()
```

### 绑定参数类型

```go
// main.ProviderInput(Wails 生成 models 类型)
type ProviderInput struct {
    ID      string            `json:"id"`
    Enabled bool              `json:"enabled"`
    Creds   map[string]string `json:"creds"` // 字段 key 见注册表 Fields;空串 = 不修改
    Budget  float64           `json:"budget"`
}
```

### GetConfig 返回结构（前端契约）

```json
{
  "providers": [
    {
      "id": "kimi",
      "name": "Kimi",
      "abbr": "K",
      "kind": "usage",
      "enabled": true,
      "budget": 0,
      "login_url": "",
      "fields": [{"key": "api_key", "label": "API Key", "type": "password"}],
      "creds": {"api_key": "sk-...xxxx"}
    },
    { "...": "每个注册 Provider 一条,顺序 = 注册表顺序" }
  ],
  "refresh_interval_min": 15,
  "ball_x": -1,
  "ball_y": -1
}
```

- `providers` 按注册表返回 6 条（kimi/xfyun/opencode-go/mimo/deepseek/ollama），`enabled` 标记展示
- `fields` 驱动前端动态渲染凭证输入框（password/text/textarea）
- `creds` 为掩码后值（前端放 placeholder）；`login_url` 空则不渲染"打开登录页"按钮

### 事件契约（后端 → 前端）

| 事件 | 负载 | 触发 |
|---|---|---|
| `quota:update` | `[]QuotaResult`（顺序 = 启用 Provider 注册表顺序） | 每次 Refresh 成功 |
| `ui:show-settings` | 无 | 托盘"打开配置" |

### QuotaResult 关键字段

| 字段 | 说明 |
|---|---|
| `id` / `abbr` | provider id 与球格缩写（fetchAll 填充，前端零匹配） |
| `kind` | `"usage"`（默认）/ `"balance"`（余额型，按预算消耗百分比着色） |
| `balance` / `currency` | 余额型原始余额数值与货币代码（ApplyBudget 换算前的原始值） |

### 前端调用示例

```js
const cfg = await window.go.main.App.GetConfig();
await window.go.main.App.SaveConfig(providers, interval);
const result = await window.go.main.App.TestConnection("deepseek");
window.runtime.EventsOn("quota:update", (results) => render(results));
```

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `app.go` | 方法实现（签名源头） |
| `frontend/wailsjs/go/main/App.d.ts` / `App.js` | 生成的前端绑定 |
| `frontend/wailsjs/go/models.ts` | ProviderInput 等绑定类型 |
| `frontend/src/main.js` | 调用方 |

---

## Must NOT Change

- 改 Go 方法签名后必须重新生成 wailsjs 绑定（`wails build` 自动），否则前端调用断裂
- `TestConnection` 的 platform 字符串 = 注册表 id（"kimi"/"xfyun"/"opencode-go"/"mimo"/"deepseek"/"ollama"），是前后端契约
- GetConfig 的 providers 数组结构（前端 main.js 渲染依赖）
- SaveConfig 的钳制语义（0 个 → 全部启用;>3 → 保留前 3）
