# 模块目录

**When to read**: 快速定位某个功能在哪个文件、某个模块多大时。

---

## 核心内容

行数基于 2026-09-30(cc01446)工作区状态。

| 模块 | 文件 | 行数 | 职责 |
|---|---|---|---|
| 应用编排 | `app.go` | 474 | 启动/刷新/动态配置/窗口定位/事件转发 |
| 入口 | `main.go` | 39 | Wails 运行时选项、embed 前端产物 |
| Win32 辅助 | `workarea_windows.go` | 131 | 工作区查询、DPI、工具窗样式、最小宽度子类 |
| Win32 桩 | `workarea_other.go` | 13 | 非 Windows 编译桩 |
| 配置模型 | `internal/config/config.go` | 210 | 动态 Provider 列表、Load/Save、旧格式迁移、新 Provider 追加、超限启用钳制 |
| Cookie 解析 | `internal/config/cookie.go` | 47 | PowerShell Cookie 粘贴 → 请求头格式 |
| 配置测试 | `internal/config/config_test.go` | 316 | Load 默认值、Save 往返(含 Budget)、旧格式迁移用例、清单同步守护、超限钳制用例 |
| 统一类型 | `internal/fetcher/types.go` | 33 | QuotaResult(含 ID/Abbr/Kind/Balance/Currency)+ Fetcher 接口 |
| 预算换算 | `internal/fetcher/budget.go` | 27 | 余额型预算 → 消耗百分比(ApplyBudget,默认预算 300) |
| Provider 注册表 | `internal/fetcher/registry.go` | 139 | 八平台元数据 + Build 工厂 |
| 数字格式化 | `internal/fetcher/format.go` | 24 | 千分位 formatNum |
| Kimi 抓取器 | `internal/fetcher/kimi.go` | 239 | Kimi 开放平台额度 API；Percent 取 5h/7d 窗口较紧张者 |
| Kimi 测试 | `internal/fetcher/kimi_test.go` | 204 | httptest 用例 |
| 讯飞抓取器 | `internal/fetcher/xfyun.go` | 134 | 讯飞星辰额度 API |
| 讯飞测试 | `internal/fetcher/xfyun_test.go` | 136 | httptest 用例 |
| OpenCode 抓取器 | `internal/fetcher/opencode_go.go` | 233 | OpenCode Go 套餐滚动额度 |
| OpenCode 测试 | `internal/fetcher/opencode_go_test.go` | 317 | httptest 用例 |
| MiMo 抓取器 | `internal/fetcher/mimo.go` | 117 | 小米 MiMo 额度 API(恢复) |
| MiMo 测试 | `internal/fetcher/mimo_test.go` | 136 | httptest 用例(恢复) |
| DeepSeek 抓取器 | `internal/fetcher/deepseek.go` | 127 | DeepSeek 账户余额(余额型) |
| DeepSeek 测试 | `internal/fetcher/deepseek_test.go` | 200 | httptest + ApplyBudget 用例 |
| Ollama 抓取器 | `internal/fetcher/ollama.go` | 210 | Ollama Cloud settings HTML，5 小时 Session 主窗口；慢响应 30s 超时 + 重试一次 |
| Ollama 测试 | `internal/fetcher/ollama_test.go` | 382 | httptest 用例（HTML/登录失效/0%/回退解析） |
| Command Code 抓取器 | `internal/fetcher/commandcode.go` | 177 | Command Code 私有 CLI 路由，5h/周窗口 + 月度 credits |
| Command Code 测试 | `internal/fetcher/commandcode_test.go` | 223 | httptest 用例 |
| Factory Droid 抓取器 | `internal/fetcher/factorydroid.go` | 307 | Factory 私有 billing 路由，5h/周/月窗口 + Core 池 + Extra 余额；Key 留空读 ~/.factory/.env |
| Factory Droid 测试 | `internal/fetcher/factorydroid_test.go` | 231 | httptest 用例（双账单模型/.env 发现/告警钳制） |
| 注册表测试 | `internal/fetcher/registry_test.go` | 63 | 注册表完整性（离线） |
| 系统托盘 | `internal/tray/tray.go` | 104 | 托盘菜单 + 事件转发 + LockOSThread |
| 托盘图标 | `internal/tray/assets/` | - | icon.ico / icon.png（embed） |
| 前端 HTML | `frontend/src/index.html` | 57 | 悬浮球/面板/动态配置面板结构 |
| 前端逻辑 | `frontend/src/main.js` | 342 | 视图状态机、动态球格、配置渲染、刷新倒计时 |
| 前端样式 | `frontend/src/style.css` | 370 | 设计令牌 + 深色玻璃质感 + Provider 卡片 |
| Vite 配置 | `frontend/vite.config.js` | 9 | 构建配置 |
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
