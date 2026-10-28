# API 契约（Wails 绑定）

**When to read**: 前端调用后端方法、修改绑定签名、排查 `window.go.main.App.*` 调用时。

---

## 核心内容

前端通过 `frontend/wailsjs/go/main/App.*` 调用（构建时自动生成，签名随 app.go 方法变化）。

### App 导出方法（Go 侧签名）

```go
// 刷新: 并发抓取全部监控对象(启用 Provider × 凭证组 × 套餐卡), 写缓存,
// emit "quota:update", 返回结果。此处不截断,截断在前端展示层
// (超上限时末格显示 "+N")。顺序 = 注册表顺序 → 凭证组顺序 → 套餐卡顺序
func (a *App) Refresh() []fetcher.QuotaResult

// 读配置: 全部 Provider 元数据 + 掩码凭证 + extended_mode
func (a *App) GetConfig() map[string]interface{}

// 存配置: 动态 Provider 列表(见下); 空字符串凭证 = 不修改;
// 监控对象数按 extendedMode 钳制(普通 ≤9 / 扩展 ≤16), 并写回 ExtendedMode
func (a *App) SaveConfig(providers []ProviderInput, refreshMin int, extendedMode bool) error

// 测试单平台连接: platform ∈ 注册表 id; credIndex 选第几组凭证(0 = 主凭证,
// 越界回落主凭证); 多套餐平台逐张校验后汇总回报;
// 返回 "成功: ..." 或 "失败: ..."
func (a *App) TestConnection(platform string, credIndex int) string

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
    ID      string             `json:"id"`
    Enabled bool               `json:"enabled"`
    Creds   map[string]string  `json:"creds"`  // 主凭证(第 1 组);字段 key 见注册表 Fields;空串 = 不修改
    Budget  float64            `json:"budget"`
    Extra   []config.CredGroup `json:"extra"`  // 第 2..N 组凭证;按 Key 对齐,未提交的组即删除
}

// config.CredGroup
type CredGroup struct {
    Key   string            `json:"key,omitempty"` // 稳定标识(前端生成、保存时原样带回)
    Creds map[string]string `json:"creds,omitempty"`
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
      "creds": {"api_key": "sk-...xxxx"},
      "extra": [{"key": "kimi#1", "creds": {"api_key": "sk-...yyyy"}}]
    },
    { "...": "每个注册 Provider 一条,顺序 = 注册表顺序" }
  ],
  "refresh_interval_min": 15,
  "ball_x": -1,
  "ball_y": -1,
  "extended_mode": false,
  "object_limit": 9,
  "object_count": 6
}
```

- `providers` 按注册表返回 10 条（kimi/xfyun/opencode-go/mimo/deepseek/ollama/command-code/factory-droid/minimax/open-design），`enabled` 标记展示
- `fields` 驱动前端动态渲染凭证输入框（password/text/textarea）
- `creds` 为掩码后值（前端放 placeholder）；`login_url` 空则不渲染"打开登录页"按钮
- `extra` 是附加凭证组(同平台多 Key / 多账号),各自报一个监控对象
- `object_limit` / `object_count` 是悬浮球对象上限与当前占用(前端展示提示用) 

### 事件契约（后端 → 前端）

| 事件 | 负载 | 触发 |
|---|---|---|
| `quota:update` | `[]QuotaResult`（一个监控对象一条；顺序 = 注册表 → 凭证组 → 套餐卡） | 每次 Refresh 成功 |
| `ui:show-settings` | 无 | 托盘"打开配置" |

### QuotaResult 关键字段

| 字段 | 说明 |
|---|---|
| `id` / `abbr` | provider id 与球格缩写（fetchAll 填充，前端零匹配） |
| `slot` | 监控对象稳定标识 `<provider>/<凭证组>[/p<套餐卡>]`（SettleIdentity 填） |
| `sub_label` | 展示后缀：套餐名 / `Key 2 · 视频赠送卡`；单对象平台为空 |
| `kind` | `"usage"`（默认）/ `"balance"`（余额型，按预算消耗百分比着色） |
| `balance` / `currency` | 余额型原始余额数值与货币代码（ApplyBudget 换算前的原始值） |

### 前端调用示例

```js
const cfg = await window.go.main.App.GetConfig();
await window.go.main.App.SaveConfig(providers, interval, extendedMode);
const result = await window.go.main.App.TestConnection("deepseek", 0); // 0 = 主凭证
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
- `TestConnection` 的 platform 字符串 = 注册表 id（当前 11 个：kimi/xfyun/opencode-go/mimo/deepseek/ollama/command-code/factory-droid/minimax/open-design/devin），是前后端契约
- GetConfig 的 providers 数组结构与 `extended_mode` 字段（前端 main.js 渲染依赖）
- SaveConfig 的钳制语义（0 个监控对象 → 全部启用;超过上限 → 按顺序保留到上限为止;上限由 extendedMode 决定:普通 9 / 扩展 16，计算口径是“启用 Provider 的凭证组数之和”,套餐卡数拿到才知、由前端展示层截断）
- `CredGroup.Key` 的对齐语义（删中间一行不得让后面的凭证值整体错位）
- `probe` 探针的记录与悬浮球同口径：一对象一条、含 `slot` / `sub_label`（`schema: 2`）
