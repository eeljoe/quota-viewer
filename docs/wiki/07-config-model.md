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
}

type ProviderConfig struct {
    ID      string            `json:"id"`
    Enabled bool              `json:"enabled"`
    Creds   map[string]string `json:"creds,omitempty"` // 凭证 key 见注册表 Fields
    Budget  float64           `json:"budget,omitempty"` // 余额型 Provider 预算总量(0 = 未设)
}
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
    {"id": "ollama", "enabled": false}
  ],
  "refresh_interval_min": 15,
  "ball_x": -1,
  "ball_y": -1
}
```

### 常量

```go
var AllProviderIDs    = []string{"kimi", "xfyun", "opencode-go", "mimo", "deepseek", "ollama"}
var DefaultProviderIDs = []string{"kimi", "xfyun", "opencode-go"} // 默认启用前三个
```

### 旧格式迁移（config v1 → v2，Load 时自动）

旧扁平字段 `kimi_api_key` / `xfyun_cookie` / `mimo_cookie` / `opencode_go_workspace_id` / `opencode_go_session_token`：

- 检测：JSON 中 `providers` 数组非空 → 新格式；否则按旧格式解析
- 有值的旧字段 → 对应 Provider `enabled=true` + 凭证迁移（mimo_cookie 也会迁移保留）
- 旧字段全空 → 默认启用前三个
- 迁移后立即回写新格式（`_ = Save(cfg)`，失败静默）
- 4 个以上有值 → 钳制：按注册表顺序保留前 3 个 enabled，其余凭证保留但关闭
- 已有 v2 配置缺少新 Provider 时，`ensureKnownProviders` 在 Load 时追加默认关闭的条目并回写，保留用户原有启用状态与凭证

### 存储位置

- 路径：`%APPDATA%/quota-viewer/config.json`（`configDir()` 读 APPDATA 环境变量）
- 文件不存在 → 返回 `Default()` 配置不报错；目录不存在 → Save 时 `MkdirAll`
- 默认值：`RefreshIntervalMin=15`，`BallX/BallY=-1`（-1 = 未记录位置，启动不移动窗口）

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
| `internal/config/config.go` | Config/ProviderConfig 结构、Default、Load（含迁移）、Save |
| `internal/config/cookie.go` | NormalizeCookieInput + PS 转义还原 |
| `internal/config/config_test.go` | 默认值、往返、旧格式迁移（含 mimo_cookie）用例 |
| `app.go` | GetConfig 掩码、SaveConfig 钳制与空串语义 |

---

## Must NOT Change

- `providers` 结构变更 = 破坏性变更；迁移逻辑在 config.Load 内,改结构必须同步迁移
- `SaveConfig` 的空凭证"不修改"语义——前端掩码回显依赖它
- Cookie 输入规范化必须始终经过 `NormalizeCookieInput`（xfyun 与 mimo 都走它）
- 展示上限 3 / 下限 1:config 迁移与 app.SaveConfig 双处钳制
