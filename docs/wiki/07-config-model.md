# 配置模型

**When to read**: 修改配置结构、持久化、迁移、Cookie 粘贴解析时。

---

## 核心内容

### Config 结构与 JSON 映射（动态 Provider 列表）

```go
type Config struct {                       // internal/config/config.go
    Providers          []ProviderConfig `json:"providers"` // 有序;顺序 = 展示顺序
    RefreshIntervalMin int              `json:"refresh_interval_min"`
    BallX              int              `json:"ball_x"`
    BallY              int              `json:"ball_y"`
    ExtendedMode       bool             `json:"extended_mode,omitempty"` // 扩展模式:监控对象上限 16(默认 false = 9)
}

type ProviderConfig struct {
    ID      string            `json:"id"`
    Enabled bool              `json:"enabled"`
    Creds   map[string]string `json:"creds,omitempty"` // 主凭证(第 1 组);key 见注册表 Fields
    Budget  float64           `json:"budget,omitempty"` // 余额型 Provider 预算总量(0 = 未设)
    Extra   []CredGroup       `json:"extra,omitempty"` // 第 2..N 组凭证(同平台多 Key / 多账号)
}

// CredGroup 是一组额外凭证。Key 是稳定标识(前端生成、保存时原样带回):
// 保存时按它对齐到已有组,避免删中间一行导致后面的凭证值整体错位。
type CredGroup struct {
    Key   string            `json:"key,omitempty"`
    Creds map[string]string `json:"creds,omitempty"`
}

// CredSets() 返回全部凭证组(主凭证在前),全空的附加组被跳过
// (用户点了"添加一行"却没填内容时,不该凭空多出一格报错的监控对象)。
func (p ProviderConfig) CredSets() []map[string]string
```

示例：

```json
{
  "providers": [
    {"id": "kimi", "enabled": true, "creds": {"api_key": "sk-kimi-xxx"}},
    {"id": "xfyun", "enabled": true},
    {"id": "opencode-go", "enabled": true},
    {"id": "mimo", "enabled": false},
    {"id": "deepseek", "enabled": false, "budget": 300},
    {"id": "ollama", "enabled": false},
    {"id": "command-code", "enabled": false},
    {"id": "factory-droid", "enabled": true},
    {"id": "minimax", "enabled": false},
    {"id": "open-design", "enabled": false}
  ],
  "refresh_interval_min": 15,
  "ball_x": -1,
  "ball_y": -1
}
```

### 常量

```go
var AllProviderIDs    = []string{"kimi", "xfyun", "opencode-go", "mimo", "deepseek", "ollama", "command-code", "factory-droid", "minimax", "open-design"}
var DefaultProviderIDs = []string{"kimi", "xfyun", "opencode-go"} // 默认启用前三个
// 上限按**监控对象**计(不再按 Provider 计):
// 一个监控对象 = 一组凭证 × 一张套餐卡。60x60 的悬浮球里:
//   普通模式 9 个 = 3x3 网格,每格约 20px
//   扩展模式 16 个 = 4x4 网格,每格约 15px(再密 CJK/两字缩写就糊成一团)
const (
    maxObjectsNormal   = 9
    maxObjectsExtended = 16
)
func MaxObjects(extended bool) int   // 扩展 16 / 普通 9
func CountObjects(providers []ProviderConfig) int // 启用 Provider 的凭证组数之和
func clampEnabled(providers []ProviderConfig, limit int) // 超限的 Provider 整家关闭(不切半组)
```

> 套餐卡数拿到才知(要打过上游),所以这里的上限按「凭证组」计;拿到套餐卡后的超限由前端展示层以 `+N` 格呈现。

> 清单与 fetcher 注册表的同步由 `config_test.go` 的 `TestAllProviderIDs_MatchesFetcherRegistry` 守护——注册表新增 Provider 时两处必须同步改。
> **上限计的是监控对象而不是 Provider 数**：一家挂 3 把 Key 就占 3 格。注册表已有 10 家，但它们并不会同时启用——真正的很现是先到 9 / 16 格的那一家。

### 旧格式迁移（config v1 → v2，Load 时自动）

旧扁平字段 `kimi_api_key` / `xfyun_cookie` / `mimo_cookie`：

- 检测：JSON 中 `providers` 数组非空 → 新格式；否则按旧格式解析
- 有值的旧字段 → 对应 Provider `enabled=true` + 凭证迁移（mimo_cookie 也会迁移保留）
- **`opencode_go_workspace_id` / `opencode_go_session_token` 自 2026-10-13 起不再迁移**：该家凭证已换成单字段 `api_key`（走官方 `/zen/go/v1/usage`），旧的 `auth` Cookie 通道已废、无可迁价值；这两个旧 key 与对应结构体字段已删除，迁移时被忽略（该家仍按 `DefaultProviderIDs` 默认启用，只是 creds 为空）
- **已知取舍**：已落盘的 v2 配置里若残留注册表未声明的 cred key（例如升级前存下的那两个旧 opencode-go key），`Load`/`Save` 会原样读入回写、不会自动剔除；设置面板保存只更新注册表声明的字段。含旧敏感凭证的配置需手工清理（清理批次见 `ROADMAP.md`）
- 旧字段全空 → 默认启用前三个
- 迁移后立即回写新格式（`_ = Save(cfg)`，失败静默）
- 超过对象上限时钳制：按注册表顺序保留到上限为止，其余 Provider 整家关闭但凭证保留（旧格式无扩展模式，恒用普通上限 9 个对象）
- 已有 v2 配置缺少新 Provider 时，`ensureKnownProviders` 在 Load 时追加默认关闭的条目并回写，保留用户原有启用状态与凭证
- Load 末尾钳制：监控对象数 > `MaxObjects(cfg.ExtendedMode)`（普通 9 / 扩展 16）时按顺序保留到上限为止、其余 Provider 整家关闭（凭证与凭证组不动）——手工编辑配置/异常状态不可能再让面板渲染出不可能的状态（2026-09-30 教训）

### 存储位置

- 路径：`%APPDATA%/quota-viewer/config.json`（`configDir()` 读 APPDATA 环境变量）
- 文件不存在 → 返回 `Default()` 配置不报错；目录不存在 → Save 时 `MkdirAll`
- 默认值：`RefreshIntervalMin=15`，`BallX/BallY=-1`（-1 = 未记录位置，启动不移动窗口），`ExtendedMode=false`（普通模式，上限 9 个监控对象）

### 凭证保护

- `App.GetConfig()` 返回前对凭证做 `maskSecret` 掩码（前 4 + 后 4，长度 ≤8 显示 `****`）
- `App.SaveConfig()` 空字符串凭证 = 不修改（防止掩码回写覆盖真实值）；Budget 每次原样写入（`budget<=0` 时 fetcher 用默认预算 300）
- 配置 JSON 明文落盘（本工具定位为本地单用户工具）

### PowerShell Cookie 解析（cookie.go）

Chrome/Edge "复制为 PowerShell" 输出形如：
```
$session.Cookies.Add((New-Object System.Net.Cookie("name", "value", "/", "domain")))
```
`NormalizeCookieInput`：
1. 不含 `System.Net.Cookie` → 原样返回（已是 Cookie 头格式）
2. 正则 `psCookieLineRe` 提取全部 name/value 对 → 去重 → 拼成 `k1=v1; k2=v2` 请求头
3. `unescapePSString` 还原 PowerShell 反引号转义（`` `` ``→`` ` ``，`` `" ``→`"`，`n/r/t）

> SaveConfig 对**所有**凭证值过 NormalizeCookieInput：非 Cookie 字段（api_key 等）不含 `System.Net.Cookie` 原样返回，无害。

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `internal/config/config.go` | Config/ProviderConfig/CredGroup 结构、Default、Load（含迁移 + 超限钳制）、Save、AllProviderIDs/MaxObjects/CredSets |
| `internal/config/cookie.go` | NormalizeCookieInput + PS 转义还原 |
| `internal/config/config_test.go` | 默认值、往返、旧格式迁移（含 mimo_cookie）、清单同步守护、按凭证组数钳制（普通 9 / 扩展 16）、多凭证组往返与空组跳过用例 |
| `app.go` | GetConfig 掩码 + `extra` + `object_limit`/`object_count`、SaveConfig 钳制(按对象数 9/16)、凭证组按 Key 对齐与空串语义 |

---

## Must NOT Change

- `providers` 结构变更 = 破坏性变更；迁移逻辑在 config.Load 内,改结构必须同步迁移
- `SaveConfig` 的空凭证"不修改"语义——前端掩码回显依赖它
- Cookie 输入规范化必须始终经过 `NormalizeCookieInput`（xfyun 与 mimo 都走它）
- 展示上限：普通 9 / 扩展 16 个监控对象，下限 1。config Load(含迁移)与 app.SaveConfig 按 `ExtendedMode` 钳制;AllProviderIDs 与 fetcher 注册表必须同步(有守护测试)
- 旧配置零迁移必要：`Extra` 为空时数据模型与旧版完全一致(主凭证语义未变)
- 凭证组的主凭证位置不可移：前端对第 1 组无删除按钮，后端也不会自动抬它到前面
