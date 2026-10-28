# 模块目录

**When to read**: 快速定位某个功能在哪个文件、某个模块多大时。

---

## 核心内容

行数基于 2026-10-13(`6d8cdee`)工作区状态，口径 = `(Get-Content <文件> | Measure-Object -Line).Lines`（**非空行数，空行不计**；与 `(Get-Content <文件>).Count` 的总行数不同）。本页此前混用「总行数」与「非空行数」两种口径，本次统一为上述命令，因此个别**代码未变**的文件数字也会随之变化。

**2026-10-11 16:11 复核**（`2d9031f` → `9760924`）：本轮只有前端两处代码变更，按同一口径（`(Get-Content <文件> | Measure-Object -Line).Lines`）重数——`frontend/src/main.js` 377→382、`frontend/src/style.css` 391→402；`frontend/src/index.html`（59）、`frontend/vite.config.js`（8）以及全部 Go 文件未变（`git diff --name-only 2d9031f..HEAD` 无 `.go` 条目）。

**2026-10-14 01:40 复核**（`6d8cdee` → `b5537c2`）：托盘单击唤出改动只碰 Go 与 README，按同一口径重数——`app.go` 427→**435**、`internal/tray/tray.go` 85→**87**；其余 Go 文件行数未变（`internal/probe/probe.go` 254、`internal/fetcher/registry.go` 164 实测复核）。

**2026-10-18 21:00 复核**（同 URL 多 Key / 多套餐任务组）：重数并新增两行（`internal/fetcher/abbr.go` / `abbr_test.go`）——`app.go` 435→**529**（编排层改为「Provider × 凭证组 × 套餐卡」）、`types.go` 新增 `MultiFetcher`、`probe.go` 记录由 8 键增至 10 键；前端 `main.js` 382→**512**（凭证组增删 + 球格 9/16 布局 + `+N` 溢出格）、`style.css` 402→**440**、`index.html` 59→**60**。

**2026-10-29 02:23 复核**（`dc127dc`，第 11 家 Provider Devin 落地）：按同一口径重数并新增两行——`internal/fetcher/registry.go` 164→**177**（新增 `devin` 条目，注册表 11 家、`devin` 末位）、新增 `internal/fetcher/devin.go` **446** 与 `internal/fetcher/devin_test.go` **1070**；其余行未重数。

| 模块 | 文件 | 行数 | 职责 |
|---|---|---|---|
| 应用编排 | `app.go` | 529 | 启动/刷新/动态配置/窗口定位/事件转发（含 `tray:activate` 唤出到前台）；`fetchAll` 按「Provider × 凭证组 × 套餐卡」展开监控对象 |
| 入口 | `main.go` | 34 | Wails 运行时选项、embed 前端产物 |
| Win32 辅助 | `workarea_windows.go` | 112 | 工作区查询、DPI、工具窗样式、最小宽度子类 |
| Win32 桩 | `workarea_other.go` | 10 | 非 Windows 编译桩 |
| 配置模型 | `internal/config/config.go` | 212 | 动态 Provider 列表、Load/Save、旧格式迁移（opencode-go 的两个死 key 已删）、新 Provider 追加、超限启用钳制(普通 3 / 扩展 9) |
| Cookie 解析 | `internal/config/cookie.go` | 43 | PowerShell Cookie 粘贴 → 请求头格式 |
| 配置测试 | `internal/config/config_test.go` | 485 | Load 默认值、Save 往返(含 Budget)、旧格式迁移用例、清单同步守护、按凭证组数钳制(9/16)、多凭证组往返与空组跳过用例 |
| 统一类型 | `internal/fetcher/types.go` | 53 | QuotaResult(含 ID/Abbr/Kind/Balance/Currency/Slot/SubLabel)+ Fetcher 接口 + 可选 `MultiFetcher` 与 `FetchAll` 入口 |
| 监控对象身份 | `internal/fetcher/abbr.go` | 127 | 球格缩写派生(不重复)、`planAbbrSetter`、`SettleIdentity`(缩写/后缀/slot 归一,app.go 与 probe 共用) |
| 身份测试 | `internal/fetcher/abbr_test.go` | 129 | SettleIdentity 五种组合 + 缩写派生/去重 + `FetchAll` 兜底用例 |
| 预算换算 | `internal/fetcher/budget.go` | 24 | 余额型预算 → 消耗百分比(ApplyBudget,默认预算 300) |
| Provider 注册表 | `internal/fetcher/registry.go` | 177 | 十一家平台元数据 + Build 工厂 |
| 数字格式化 | `internal/fetcher/format.go` | 22 | 千分位 formatNum |
| Kimi 抓取器 | `internal/fetcher/kimi.go` | 239 | Kimi 开放平台额度 API；Percent 取 5h/7d 窗口较紧张者 |
| Kimi 测试 | `internal/fetcher/kimi_test.go` | 204 | httptest 用例 |
| 讯飞抓取器 | `internal/fetcher/xfyun.go` | 114 | 讯飞星辰额度 API |
| 讯飞测试 | `internal/fetcher/xfyun_test.go` | 124 | httptest 用例 |
| OpenCode 抓取器 | `internal/fetcher/opencode_go.go` | 204 | OpenCode Go 官方 JSON API `GET /zen/go/v1/usage`（Bearer `oc_sk_` Key；留空读 `~/.local/share/opencode/auth.json`），取最紧张窗口；老的 Dashboard 页面/Cookie 通道已废 |
| OpenCode 测试 | `internal/fetcher/opencode_go_test.go` | 358 | httptest 用例（响应形状/选窗 tiebreak/五档错误文案/凭证回退面） |
| MiMo 抓取器 | `internal/fetcher/mimo.go` | 135 | 小米 MiMo 额度 API(恢复);每个有 limit 的额度桶(套餐/补偿)一个监控对象 |
| MiMo 测试 | `internal/fetcher/mimo_test.go` | 124 | httptest 用例(恢复) |
| DeepSeek 抓取器 | `internal/fetcher/deepseek.go` | 110 | DeepSeek 账户余额(余额型) |
| DeepSeek 测试 | `internal/fetcher/deepseek_test.go` | 183 | httptest + ApplyBudget 用例 |
| Ollama 抓取器 | `internal/fetcher/ollama.go` | 210 | Ollama Cloud settings HTML，5 小时 Session 主窗口；慢响应 30s 超时 + 重试一次 |
| Ollama 测试 | `internal/fetcher/ollama_test.go` | 382 | httptest 用例（HTML/登录失效/0%/回退解析） |
| Command Code 抓取器 | `internal/fetcher/commandcode.go` | 177 | Command Code 私有 CLI 路由，5h/周窗口 + 月度 credits |
| Command Code 测试 | `internal/fetcher/commandcode_test.go` | 223 | httptest 用例 |
| Factory Droid 抓取器 | `internal/fetcher/factorydroid.go` | 473 | Factory 私有 billing 路由，5h/周/月窗口 + Core 池 + Extra 余额；`overagePreference` 决定 Standard 顶格后超额走哪一池，接管时展示主体切 Core；`Remaining` 按已用百分比降序（最紧张的打头），`ResetAt`/`ResetLabel` 跟排头窗口走；Key 留空读 ~/.factory/.env |
| Factory Droid 测试 | `internal/fetcher/factorydroid_test.go` | 612 | httptest 用例（双账单模型/.env 发现/告警钳制/Core 接管与平手取先者/ResetAt 亚秒夹逼/最紧张窗口打头与倒计时标签） |
| MiniMax 抓取器 | `internal/fetcher/minimax.go` | 220 | MiniMax Token Plan 私有路由，三主机回退链，5h/周双窗口(剩余% 反转为已用%);`model_remains[]` 每张套餐卡一个监控对象 |
| MiniMax 测试 | `internal/fetcher/minimax_test.go` | 226 | httptest 用例（域名回退/百分比语义守护/周无限制） |
| Open Design 抓取器 | `internal/fetcher/opendesign.go` | 207 | 控制面两步 GET(workspace 发现 + coding-plan-usage)，Cookie 认证，302 并入失效分支 |
| Open Design 测试 | `internal/fetcher/opendesign_test.go` | 369 | httptest 用例（两步流程/窗口按 durationSeconds 识别/302 与鉴权失效） |
| Devin 抓取器 | `internal/fetcher/devin.go` | 446 | Devin Desktop 私有席位接口 `GetUserStatus`（CONNECT POST，`metadata.api_key`）：无公开额度 API，v2/v3 consumption 仅 Enterprise 管理员；token 手填优先、留空读本机 `%APPDATA%\Devin\User\globalStorage\state.vscdb` 登录态；响应是**剩余**百分比 → `clampPercent(100-剩余)` 反转；日/周可见窗口里已用较大者驱动 `Percent`/`ResetAt`/`ResetLabel`（并列取日），`Remaining` 按已用降序 + 超额金额 |
| Devin 测试 | `internal/fetcher/devin_test.go` | 1070 | httptest 用例（本机登录态扫描与体积守卫/server 三级兜底/字符串型 int64/剩余百分比反转/隐藏窗口与缺失窗口/驱动窗口与 ResetLabel/401-403 与 HTTP N 文案） |
| 注册表测试 | `internal/fetcher/registry_test.go` | 59 | 注册表完整性（离线） |
| CLI 入口 | `cmd/quotaviewer/main.go` | 10 | 无界面探针的独立 main（console 子系统，不 import wails）；逻辑全在 `internal/probe` |
| 探针契约 | `internal/probe/probe.go` | 302 | `--json` 的 JSON 契约（`schema: 2`）：顶层 7 键 + **记录 10 键**(新增 `slot`/`sub_label`)、一监控对象一条记录、结果归一（四句 + `def.Kind` 兜底）、`reset_at` 三条规则、失败信封、退出码三分（0/2/1） |
| 探针测试（契约/入口） | `internal/probe/report_test.go` | 614 | 契约逐字比对、失败信封、退出码矩阵（含 stderr 与写失败） |
| 探针测试（归一） | `internal/probe/probe_test.go` | 67 | `reset_at` 三条规则（亚秒→UTC 秒、无时区串按 +08:00、非未来→null） |
| 系统托盘 | `internal/tray/tray.go` | 87 | 托盘菜单 + 事件转发（单击图标发 `tray:activate`）× LockOSThread |
| 托盘图标 | `internal/tray/assets/` | - | icon.ico / icon.png（embed） |
| 前端 HTML | `frontend/src/index.html` | 60 | 悬浮球/面板/动态配置面板结构（含对象上限提示位） |
| 前端逻辑 | `frontend/src/main.js` | 512 | 视图状态机、动态球格(列数 `--cols` / 窄格 `narrow` / 溢出格 `+N`)、多组凭证配置渲染(增删与稳定 Key)、对象上限提示、刷新倒计时 |
| 前端样式 | `frontend/src/style.css` | 440 | 设计令牌 + 深色玻璃质感 + Provider 卡片 + 凭证组块 + 球格布局(等宽换行 / 末行居中 / 窄格字号档 / 10-16 格档) |
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
