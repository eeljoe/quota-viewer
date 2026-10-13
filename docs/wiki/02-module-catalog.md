# 模块目录

**When to read**: 快速定位某个功能在哪个文件、某个模块多大时。

---

## 核心内容

行数基于 2026-10-13(`6d8cdee`)工作区状态，口径 = `(Get-Content <文件> | Measure-Object -Line).Lines`（**非空行数，空行不计**；与 `(Get-Content <文件>).Count` 的总行数不同）。本页此前混用「总行数」与「非空行数」两种口径，本次统一为上述命令，因此个别**代码未变**的文件数字也会随之变化。

**2026-10-11 16:11 复核**（`2d9031f` → `9760924`）：本轮只有前端两处代码变更，按同一口径（`(Get-Content <文件> | Measure-Object -Line).Lines`）重数——`frontend/src/main.js` 377→382、`frontend/src/style.css` 391→402；`frontend/src/index.html`（59）、`frontend/vite.config.js`（8）以及全部 Go 文件未变（`git diff --name-only 2d9031f..HEAD` 无 `.go` 条目）。

| 模块 | 文件 | 行数 | 职责 |
|---|---|---|---|
| 应用编排 | `app.go` | 427 | 启动/刷新/动态配置/窗口定位/事件转发 |
| 入口 | `main.go` | 34 | Wails 运行时选项、embed 前端产物 |
| Win32 辅助 | `workarea_windows.go` | 112 | 工作区查询、DPI、工具窗样式、最小宽度子类 |
| Win32 桩 | `workarea_other.go` | 10 | 非 Windows 编译桩 |
| 配置模型 | `internal/config/config.go` | 212 | 动态 Provider 列表、Load/Save、旧格式迁移（opencode-go 的两个死 key 已删）、新 Provider 追加、超限启用钳制(普通 3 / 扩展 9) |
| Cookie 解析 | `internal/config/cookie.go` | 43 | PowerShell Cookie 粘贴 → 请求头格式 |
| 配置测试 | `internal/config/config_test.go` | 400 | Load 默认值、Save 往返(含 Budget)、旧格式迁移用例、清单同步守护、超限钳制(3/9)用例 |
| 统一类型 | `internal/fetcher/types.go` | 29 | QuotaResult(含 ID/Abbr/Kind/Balance/Currency)+ Fetcher 接口 |
| 预算换算 | `internal/fetcher/budget.go` | 24 | 余额型预算 → 消耗百分比(ApplyBudget,默认预算 300) |
| Provider 注册表 | `internal/fetcher/registry.go` | 164 | 十平台元数据 + Build 工厂 |
| 数字格式化 | `internal/fetcher/format.go` | 22 | 千分位 formatNum |
| Kimi 抓取器 | `internal/fetcher/kimi.go` | 239 | Kimi 开放平台额度 API；Percent 取 5h/7d 窗口较紧张者 |
| Kimi 测试 | `internal/fetcher/kimi_test.go` | 204 | httptest 用例 |
| 讯飞抓取器 | `internal/fetcher/xfyun.go` | 114 | 讯飞星辰额度 API |
| 讯飞测试 | `internal/fetcher/xfyun_test.go` | 124 | httptest 用例 |
| OpenCode 抓取器 | `internal/fetcher/opencode_go.go` | 204 | OpenCode Go 官方 JSON API `GET /zen/go/v1/usage`（Bearer `oc_sk_` Key；留空读 `~/.local/share/opencode/auth.json`），取最紧张窗口；老的 Dashboard 页面/Cookie 通道已废 |
| OpenCode 测试 | `internal/fetcher/opencode_go_test.go` | 358 | httptest 用例（响应形状/选窗 tiebreak/五档错误文案/凭证回退面） |
| MiMo 抓取器 | `internal/fetcher/mimo.go` | 103 | 小米 MiMo 额度 API(恢复) |
| MiMo 测试 | `internal/fetcher/mimo_test.go` | 124 | httptest 用例(恢复) |
| DeepSeek 抓取器 | `internal/fetcher/deepseek.go` | 110 | DeepSeek 账户余额(余额型) |
| DeepSeek 测试 | `internal/fetcher/deepseek_test.go` | 183 | httptest + ApplyBudget 用例 |
| Ollama 抓取器 | `internal/fetcher/ollama.go` | 210 | Ollama Cloud settings HTML，5 小时 Session 主窗口；慢响应 30s 超时 + 重试一次 |
| Ollama 测试 | `internal/fetcher/ollama_test.go` | 382 | httptest 用例（HTML/登录失效/0%/回退解析） |
| Command Code 抓取器 | `internal/fetcher/commandcode.go` | 177 | Command Code 私有 CLI 路由，5h/周窗口 + 月度 credits |
| Command Code 测试 | `internal/fetcher/commandcode_test.go` | 223 | httptest 用例 |
| Factory Droid 抓取器 | `internal/fetcher/factorydroid.go` | 360 | Factory 私有 billing 路由，5h/周/月窗口 + Core 池 + Extra 余额；`overagePreference` 决定 Standard 顶格后超额走哪一池，接管时展示主体切 Core；Key 留空读 ~/.factory/.env |
| Factory Droid 测试 | `internal/fetcher/factorydroid_test.go` | 535 | httptest 用例（双账单模型/.env 发现/告警钳制/Core 接管与平手取先者/ResetAt 亚秒夹逼） |
| MiniMax 抓取器 | `internal/fetcher/minimax.go` | 201 | MiniMax Token Plan 私有路由，三主机回退链，5h/周双窗口(剩余% 反转为已用%) |
| MiniMax 测试 | `internal/fetcher/minimax_test.go` | 226 | httptest 用例（域名回退/百分比语义守护/周无限制） |
| Open Design 抓取器 | `internal/fetcher/opendesign.go` | 207 | 控制面两步 GET(workspace 发现 + coding-plan-usage)，Cookie 认证，302 并入失效分支 |
| Open Design 测试 | `internal/fetcher/opendesign_test.go` | 369 | httptest 用例（两步流程/窗口按 durationSeconds 识别/302 与鉴权失效） |
| 注册表测试 | `internal/fetcher/registry_test.go` | 59 | 注册表完整性（离线） |
| CLI 入口 | `cmd/quotaviewer/main.go` | 10 | 无界面探针的独立 main（console 子系统，不 import wails）；逻辑全在 `internal/probe` |
| 探针契约 | `internal/probe/probe.go` | 254 | `--json` 的 JSON 契约：顶层 7 键 + 记录 8 键、结果归一（四句 + `def.Kind` 兜底）、`reset_at` 三条规则、失败信封、退出码三分（0/2/1） |
| 探针测试（契约/入口） | `internal/probe/report_test.go` | 614 | 契约逐字比对、失败信封、退出码矩阵（含 stderr 与写失败） |
| 探针测试（归一） | `internal/probe/probe_test.go` | 67 | `reset_at` 三条规则（亚秒→UTC 秒、无时区串按 +08:00、非未来→null） |
| 系统托盘 | `internal/tray/tray.go` | 85 | 托盘菜单 + 事件转发 + LockOSThread |
| 托盘图标 | `internal/tray/assets/` | - | icon.ico / icon.png（embed） |
| 前端 HTML | `frontend/src/index.html` | 59 | 悬浮球/面板/动态配置面板结构 |
| 前端逻辑 | `frontend/src/main.js` | 382 | 视图状态机、动态球格(列数 `--cols` / 窄格 `narrow` 标记)、配置渲染、刷新倒计时 |
| 前端样式 | `frontend/src/style.css` | 402 | 设计令牌 + 深色玻璃质感 + Provider 卡片 + 球格布局(等宽换行 / 末行居中 / 窄格字号档) |
| Vite 配置 | `frontend/vite.config.js` | 8 | 构建配置 |
| 绑定 | `frontend/wailsjs/` | - | Wails 生成的前端绑定(构建生成) |

---

## 关键文件

| 文件 | 职责 |
|---|---|
| 全部见上表 | |

---

## Must NOT Change

- `workarea_other.go` 只做桩，不得实现真实逻辑（非 Windows 不维护）
- `frontend/wailsjs/` 与 `frontend/dist/` 是生成物，手工修改会被下次构建覆盖
