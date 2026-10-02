# 📋 项目进度文档

> **⚠️ 新会话阅读指南（必读）**
>
> 本文档较长，**不要通读全文**。按以下顺序阅读：
>
> 1. 阅读下方「上下文摘要（TL;DR）」快速了解项目状态
> 2. 阅读下方「项目概况」了解项目基本信息
> 3. 如果下方有「知识库」区块且状态为已初始化，读取知识库获取项目结构和函数映射
> 4. 阅读项目根目录 ROADMAP.md 的「Now」栏（滚动计划，了解接下来做什么）
> 5. **直接跳到最新一条「会话记录：2026-09-22 10:03」章节**（用标题搜索定位，不要假设在文件末尾）
>    - 重点看「未完成 & 下一步」和「关键上下文」
> 6. 如果最新章节引用了更早的内容，再按需回溯
> 7. 给用户一份简短进度汇报（每段 3-5 行）：**当前进度**（当前在做什么 → 做到哪了 → 下一步是什么 → 有无阻塞项）+ **上个会话做了什么**（最新会话章节的摘要、完成与遗留）
> 8. **严禁在用户给出指示之前修改任何文件或代码**
> 9. 汇报后等待用户指示，不要主动开始执行任务
>
> 📍 最新章节位置：`## 会话记录：2026-10-02 22:57`（搜索定位，可能不在文件末尾）
> 🔖 对应 commit：`b13a10a` on `master`（已推送）
> 📊 累计会话数：主文档 8 条，另有 5 条已归档（docs/wiki/99-appendix-legacy-status.md）

---

## 最后更新

<!-- git-meta: {"last_commit": "b13a10a", "branch": "master", "dirty": false, "timestamp": "2026-10-02T23:00:00+08:00"} -->

- **日期**：2026-10-02 22:57
- **会话摘要**：新增 MiniMax 额度监控（第 9 家 Provider，Token Plan sk-cp- Key）+ 扩展模式（悬浮球 4-9 格网格）+ 球形悬浮球改版；config.json 中途被误写坏、经内存写回完整恢复；推送 b13a10a

---

## 上下文摘要（TL;DR）

- 项目：Quota Viewer，桌面悬浮球 + AI 平台额度监控工具，Go + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite）
- 当前阶段：**v1.2.0 已发布** + master 新增三件套（未发版）：九平台监控（新增 MiniMax）、扩展模式（悬浮球最多 9 格，默认 3）、球形外观
- 状态灯契约 60% 黄 / 80% 红 / 100% 熄灭；usage 型一律「最紧张窗口」驱动 Percent（Kimi/ollama/commandcode/factory-droid/minimax 一致）
- 下一步：无明确待办，计划见 `ROADMAP.md`（Next: Wails 版本升级对齐）
- 注意事项：wiki 待 `/wiki-update` 同步本次三件套（05/02/07/09/00）；无阻塞项

---

## 项目概况

- **项目名称**：Quota Viewer
- **技术栈**：Go 1.24 + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite 打包）
- **项目根目录**：`C:/Users/joe/Desktop/工作学习/软件开发/quota viewer`
- **平台**：Windows 10+（WebView2 运行时）
- **功能**：悬浮球（1-9 格动态，颜色=状态，球形）+ 展开面板（进度条 + 剩余量明细）+ 配置面板（勾选 Provider + 各平台凭证 + 扩展模式开关）+ 系统托盘 + 关闭到托盘
- **支持 Provider**：Kimi（API Key）、讯飞星辰（Cookie）、OpenCode Go（Workspace ID + Token）、小米 MiMo（Cookie）、DeepSeek（API Key，余额型 + 预算进度条）、Ollama（Cookie，Cloud 5 小时/周用量）、Command Code（API Key，5 小时/周窗口 + 剩余 credits）、Factory Droid（API Key，双账单模型 + Core 池）、MiniMax（Token Plan 订阅 Key，5 小时/周窗口）

## 当前分支与最近提交

- **分支**：master（与远程同步）
- **HEAD**：`b13a10a`（全部已推送）
- **最近提交**：
  - `b13a10a` - feat: MiniMax 额度监控(第 9 家)+ 扩展模式 4-9 格 + 球形悬浮球
  - `edce48b` - docs: 收尾 v1.2.0 发布记录,归档额度告警修复组会话
  - `3a9eec2` - docs: 记录 Factory 计费顺序与 Core 展示规则
- **Release**：v1.2.0（2026-09-30，Latest，附 exe）← v1.1.2 ← v1.1.1 ← v1.1.0 ← v1.0.0（三件套未发版）

---

## 当前任务进度

### ✅ 已完成（截至 2026-09-30）
- **八平台监控**：Kimi / 讯飞星辰 / OpenCode Go / MiMo / DeepSeek / Ollama / Command Code / Factory Droid——fetcher 注册表驱动，新增平台前端零改动
- **状态灯契约**（9/19）：60% 黄 / 80% 红 / 100% 熄灭（off 暗灰）；ollama/commandcode 的 Percent 取 5h 与周窗口较紧张者；发布 v1.1.1
- **DeepSeek 预算条**（8/7）：QuotaResult Balance/Currency + `budget.go` ApplyBudget + ProviderConfig.Budget + 前端预算输入/色阈值/刷新倒计时
- **Ollama Provider**（8/13）：`ollama.go` HTML 解析 5 小时 Session 主窗口 + 周用量；14 个 httptest 用例；config 自动补全新 Provider（默认关闭）；真实账号冒烟通过（8/13 01:33）
- **Command Code Provider**（8/25）：`commandcode.go` 逆向官方 CLI 私有 `/alpha/*` API（whoami + billing/credits），Bearer API Key（留空自动读 `~/.commandcode/auth.json`）；8 个 httptest 用例；真实账号冒烟通过
- **Factory Droid Provider**（9/30）：`factorydroid.go` 走官方 web 同源 `/api/billing/limits`（5h/周/月 + Core 池 + Extra 余额，旧账单模型兜底），Key 留空自动读 `~/.factory/.env`；9 个 httptest 用例；真实冒烟通过；发布 v1.2.0
- **MiniMax Provider + 扩展模式 + 球形悬浮球**（10/2）：`minimax.go` 走 `coding_plan/remains`（sk-cp- 订阅 Key，三主机域名链，剩余% 反转已用%）；扩展模式启用上限 3→9（配置面板开关）；悬浮球改圆形 + 4+ 格网格布局；9 个 httptest 用例；真实冒烟通过
- **开源发布**：GitHub eeljoe/quota-viewer，Release v1.0.0 / v1.1.0 / v1.1.1 / v1.1.2 / v1.2.0（均附 exe）

### 🔄 进行中
- 无

### 📋 待办
- wiki 同步本次三件套（05 端点细节 / 02 模块表 / 07 钳制契约+扩展模式 / 09 测试 / 00 元数据），`/wiki-update` 顺手刷
- 计划级事项见 `ROADMAP.md`

---

## 未完成 & 下一步

1. 用户目验新悬浮球（4 格 2x2 网格 + 球形观感，代码侧无法替代目验）
2. wiki 三件套同步（见待办）
3. 计划级事项见 `ROADMAP.md`（Next: Wails 版本升级对齐）

---

## 已知问题与注意事项

- Wails CLI 版本 v2.13.0 > go.mod 的 v2.12.0（构建警告，不阻塞；已入 ROADMAP Next）
- 前端行尾警告：dist/wailsjs 文件 LF→CRLF，git 会提示但不影响构建
- `build/bin/quota-viewer.exe~` 是运行中构建产生的旧 exe 改名残留，可删
- Ollama 依赖 ollama.com/settings 页面 HTML 结构（无公开 quota API），页面改版 → "页面结构可能已变化"错误，需更新 `ollama.go` 解析
- Command Code 依赖官方 CLI 私有 `/alpha/*` API（无公开额度 API），CLI 升级可能改结构 → 需对照官方 cli.mjs 更新 `commandcode.go`；API Key 留空自动读 `~/.commandcode/auth.json`（与本机官方 CLI 复用同一凭证，Key 变更会同时影响两者）
- 余额型（DeepSeek）经 ApplyBudget 按预算换算消耗百分比（默认预算 300）；取首个非零余额币种（CNY→¥ / USD→$），全部为 0 才报错

---

## 会话2文件变更记录（历史，已全部提交）

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/fetcher/registry.go` | Provider 注册表（升级核心） |
| 新增 | `internal/fetcher/registry_test.go` | 注册表完整性测试 |
| 新增 | `internal/fetcher/deepseek.go` | DeepSeek 余额抓取器 |
| 新增 | `internal/fetcher/deepseek_test.go` | 余额抓取测试 |
| 新增 | `internal/fetcher/mimo.go` | 从 git 历史恢复的 MiMo 抓取器（会话前被 staged 删除） |
| 新增 | `internal/fetcher/mimo_test.go` | MiMo 测试恢复 |
| 新增 | `internal/fetcher/opencode_go.go` / `_test.go` | 会话前遗留（OpenCode Go 替换） |
| 修改 | `internal/fetcher/types.go` | QuotaResult 扩展 ID/Abbr/Kind + Kind 常量 |
| 修改 | `internal/config/config.go` | 动态 Provider 模型 + 旧格式迁移 |
| 修改 | `internal/config/config_test.go` | 迁移/往返/默认用例重写 |
| 修改 | `app.go` | 动态编排 + fitToScreen（会话前遗留）+ OpenCode 同步（会话前遗留） |
| 修改 | `frontend/src/index.html` | 动态球格容器 + 动态配置面板 |
| 修改 | `frontend/src/main.js` | 动态球格/配置渲染/事件委托 |
| 修改 | `frontend/src/style.css` | Provider 卡片 + single-cell 样式 |
| 修改 | `frontend/dist/*`、`frontend/wailsjs/*` | 重建产物与绑定（含 ProviderInput） |
| 修改 | `README.md` | 重写（开源版） |
| 新增 | `LICENSE` | MIT 许可证 |
| 新增 | `docs/STATUS.md` | 本文件（会话 1 创建） |
| 新增 | `docs/wiki/00~10` + `.covered-files` | wiki 初始化（会话 1）并同步（本次） |

---

> 📦 历史会话已归档：docs/wiki/99-appendix-legacy-status.md

## Wiki

- **位置**：`docs/wiki/`（已初始化 2026-08-04，2026-08-13 同步至 cc82c71）
- **文件数**：12 个（00-agent-rules ~ 10-build-test-baseline + 99 归档页已启用）
- **入口**：先读 `docs/wiki/00-agent-rules.md`（含索引与 wiki-meta 同步状态）
- **缓存**：`docs/wiki/.covered-files`（49 项）
- **未同步文件**：04-window-positioning、06-systray、10-build-test-baseline（近期未改动相关代码）

---

## 关键文件索引

- `app.go` - Wails 主应用（窗口定位 fitToScreen、动态 Provider 编排、配置）
- `main.go` - 入口 + Wails 运行时配置（frameless / AlwaysOnTop）
- `workarea_windows.go` / `workarea_other.go` - Win32 工作区查询辅助（非 Windows 桩）
- `internal/fetcher/registry.go` - **Provider 注册表（新增 Provider 的唯一入口）**
- `internal/fetcher/types.go` - Fetcher 接口 / QuotaResult / Kind 常量
- `internal/fetcher/{kimi,xfyun,opencode_go,mimo,deepseek,ollama}.go` + `budget.go` - 六平台抓取器（均有测试）+ 余额型预算换算
- `internal/config/config.go` + `cookie.go` - 动态 Provider 配置 + 旧格式迁移 + PowerShell Cookie 解析
- `internal/tray/tray.go` - 系统托盘（systray.Run + LockOSThread）
- `frontend/src/index.html` / `main.js` / `style.css` - 动态球格 + 动态配置面板 UI
- `frontend/wailsjs/` - Wails v2 前端绑定（构建生成）
- `docs/wiki/05-fetching-platforms.md` - 新增 Provider 指南详情
- `docs/ADDING_A_PROVIDER.md` - **Agent 专用**新增 Provider 完整指南（fetcher 实现 + 注册 + 测试 + 检查清单）

---

## 会话记录：2026-08-04 23:02

> **会话摘要**：新增 Agent 专用 Provider 添加指南文档，双语 README 更新指向该文档并附 agent 指令模板
> **Git**：`b62e6d5` on `master`（clean，已推送）

### 本次完成
- 创建 `docs/ADDING_A_PROVIDER.md`：面向 AI agent 的新增 Provider 完整指南，涵盖架构概述、Fetcher 实现（用量型/余额型/Cookie 类）、registry 注册、httptest 测试模板、验证步骤、检查清单、现有 Provider 速查表
- 更新 `README.md`（英文）"Adding a New Provider" 部分：改为引导用户将文档路径 + 平台信息复制给 agent，附可复制指令模板
- 更新 `README.zh-CN.md`（中文）"如何新增一个 Provider" 部分：同上，中文版指令模板
- 提交 `b62e6d5` 并推送到远程

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 文档放在 `docs/` 而非 `docs/wiki/` | wiki 面向 agent 日常查阅，本文档是用户主动提供给 agent 的专题指南，独立文件更合适 | 放 wiki 06 或新编号 |
| README 中附可复制 agent 指令模板 | 用户只需填空 `<平台名>` `<URL>` `<认证方式>` `<展示内容>` 即可让 agent 自主完成 | 仅放文档链接，agent 自己读 |
| 文档用中文撰写 | 项目为中文开发者项目，代码注释也全中文，保持一致 | 英文（但与代码注释风格不符） |

### 新增/变更文件
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `docs/ADDING_A_PROVIDER.md` | Agent 专用新增 Provider 完整指南（343 行） |
| 修改 | `README.md` | "Adding a New Provider" 改为 agent 指令模板 + 文档链接 |
| 修改 | `README.zh-CN.md` | "如何新增一个 Provider" 改为 agent 指令模板 + 文档链接 |
| 修改 | `docs/STATUS.md` | 本文件 |

> 本次变更（从 85e03e9 到 b62e6d5）：+432/-39 行，4 个文件

### 未完成 & 下一步
- 无明确待办——项目已交付开源
- 可选方向：Release 推广 / 新 Provider 扩展 / Wails 版本升级 / 用户反馈迭代

### 关键上下文
- **Agent 指南文档**：`docs/ADDING_A_PROVIDER.md`——用户想让 agent 新增 Provider 时，在 README 中复制指令模板填空即可
- **指令模板格式**（中文）：`阅读 docs/ADDING_A_PROVIDER.md，按照文档为 <平台名> 新增一个 Provider。端点是 <URL>，认证方式是 <方式>，展示内容是 <展示什么>`
- **GitHub 仓库**：https://github.com/eeljoe/quota-viewer（公开，已推送至 b62e6d5）
- Wiki 指针状态：`docs/wiki/` 11 个文件，`.covered-files` 46 项，synced_commit `85e03e9`（本次未改动 wiki）

---

## 会话记录：2026-08-07 16:37（补记）

> **会话摘要**：DeepSeek 余额预算进度条 + 展开面板刷新倒计时 + README 效果截图（该会话当时未写入 STATUS，本次补记）
> **Git**：`39aeaed` on `master`（clean）

### 本次完成
- `QuotaResult` 新增 `Balance/Currency` 字段，deepseek.go 填充原始余额与货币代码
- 新增 `budget.go` ApplyBudget：余额型按预算换算消耗百分比（默认预算 300，余额超预算钳制为 0）
- `ProviderDef` 新增 `Kind` 字段（usage/balance）；`ProviderConfig` 新增 `Budget` 字段
- app.go fetchAll 传递 budget 并调用 ApplyBudget；配置面板对余额型显示预算输入框
- 前端：展开面板进度条下方显示刷新倒计时（ResetAt 相对 now）；余额型状态色按消耗百分比走 yellow/red 阈值
- 双语 README 更新新功能说明与效果截图（`docs/screenshots/preview-1/2.png`）
- 测试：TestApplyBudget_* 5 个用例 + config Budget 往返

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 余额型引入"预算"换算消耗百分比 | 余额无用量语义，用户想看到"花了多少" | 余额恒绿（原方案） |
| 默认预算 300 元 | 用户未设时的合理默认 | 无默认/强制填写 |
| Kind 字段放 ProviderDef | 注册表一处标注，前端按类型渲染 | 各 fetcher 自报 |

> 本次变更：2 个 commit（+213/-20 行，20 个文件，含截图二进制）

### 关键上下文
- budget 语义：已消耗 = 预算 - 余额；`Percent = 已消耗/预算*100`；Remaining 显示 `余额 / 预算`
- 根目录遗留 `09314663d2975b947bfa75fcf46e4769.png` = preview-2.png 的副本（未跟踪）

---

## 会话记录：2026-08-13 00:55

> **会话摘要**：新增 Ollama Cloud 额度监控 Provider（六平台收尾）+ 补齐 wiki 预算模块同步
> **Git**：`cc82c71` on `master`（clean，本地领先远程 2 commit）

### 本次完成
- 接手上一会话遗留的 Ollama 半成品（ollama.go + 14 个测试 + registry 注册 + config 迁移），验证后提交
- Ollama 抓取 `https://ollama.com/settings` 页面 HTML，解析 Session（5 小时）主窗口 + Weekly 周用量百分比（无公开 quota API，issue #15132）
- config `ensureKnownProviders`：已有 v2 配置自动追加新 Provider（默认关闭，保留用户选择），测试覆盖
- 注册表测试改为空凭证离线模式（避免 Build 测试发真实网络请求）
- 验证：`go test -count=1 ./...` 全绿 / `npm run build` / `wails build` 成功；dist 重建产物与 HEAD 逐字节一致（仅文件名 hash 变化）已回退，不产生噪音提交
- wiki 补同步：budget.go 模块、Balance/Currency、Kind/Budget 契约、ollama 平台行，`.covered-files` 重建（49 项）

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| Ollama 用 HTML 解析而非 API | Ollama 无公开 quota API，只能解析 server-rendered 页面 | 放弃该平台 |
| 主展示 5 小时 Session 窗口 | 用户最关心当前会话额度（驱动球色），周用量写入 Remaining 辅助 | 周窗口为主 |
| 回退 dist/wailsjs 重建产物 | 内容与 HEAD 逐字节一致，仅文件名 hash 变化，提交纯噪音 | 提交新 hash 产物 |
| 补记 8/7 会话 + 补齐 wiki | STATUS/wiki 与代码脱节，违反文档一致性 | 只记录本次会话 |

### 新增/变更文件
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/fetcher/ollama.go` | Ollama Cloud settings HTML 抓取（206 行） |
| 新增 | `internal/fetcher/ollama_test.go` | 14 个 httptest 用例（320 行） |
| 修改 | `internal/fetcher/registry.go` / `registry_test.go` | ollama 注册 + 测试改离线空凭证 |
| 修改 | `internal/config/config.go` / `config_test.go` | AllProviderIDs + ensureKnownProviders |
| 修改 | `README.md` / `README.zh-CN.md` | 支持平台表 +Ollama |
| 修改 | `docs/ADDING_A_PROVIDER.md` | 速查表 +Ollama 特殊说明 |
| 修改 | `docs/wiki/00~09` + `.covered-files` | 预算模块补同步 + ollama 条目 |

> 本次变更：2 个 commit（+685/-57 行，26 个文件）

### 未完成 & 下一步
1. **Ollama 真实账号冒烟**——用户在配置面板粘贴 Ollama Cookie → 测试连接 → 球格展示 5 小时用量
2. **推送 master 到远程**——本地领先 2 个提交（a59635f / cc82c71），确认后 `git push`

### 已知问题 & 注意事项
- Ollama 依赖 ollama.com/settings 页面结构，页面改版 → "页面结构可能已变化"错误，需更新解析
- Cookie 失效判定：302 跳转 / 登录页 HTML 特征（form + "Sign in to Ollama"）
- 根目录 `09314663d2975b947bfa75fcf46e4769.png` 为 preview-2.png 未跟踪副本，可删除（未删，避免误删用户文件）

### 关键上下文
- Ollama 注册表字段：`textarea` Cookie，登录 URL `https://ollama.com/settings`；用户在设置页登录后粘贴含 wos-session / __Secure-session 的 Cookie（支持 "Copy as PowerShell" 整段粘贴）
- 新增 Provider 完整路径见 `docs/ADDING_A_PROVIDER.md`（注册表驱动，前端零改动）
- Wiki 指针状态：`docs/wiki/` 11 文件，synced_commit `cc82c71`，`.covered-files` 49 项
- GitHub 仓库：https://github.com/eeljoe/quota-viewer（公开，远程落后本地 2 commit）

---

## 会话记录：2026-08-13 01:33

> **会话摘要**：Ollama 真实账号冒烟通过——用户按方法 A 粘贴 PowerShell Cookie 整段，测试连接成功，球格正常展示
> **Git**：`9927e2e` on `master`（clean）

### 本次完成
- （上会话遗留）**Ollama 真实账号冒烟——已完成**：用户在配置面板粘贴浏览器 "Copy as PowerShell" 整段脚本（含 aid + __Secure-session），测试连接成功，球格展示 5 小时 Session 用量——六平台全部实测可用
- 确认构建与桌面快捷方式：`build/bin/quota-viewer.exe` 为 00:37 新构建（二进制含 ollama）；桌面 `Quota Viewer.lnk` 直接指向该 exe，无需复制
- 归档旧会话：主文档 ≥400 行触发阈值，3 条最早会话（05:31/14:10/15:00）原样移入 `docs/wiki/99-appendix-legacy-status.md`

### 未完成 & 下一步
1. **推送 master 到远程**——本地领先 3 个提交（a59635f / cc82c71 / 9927e2e），确认后 `git push`
2. 可选：根目录截图副本 PNG 删除（待用户确认）
3. 可选方向：Release 推广 / 新 Provider 扩展 / 用户反馈迭代

### 关键上下文
- Ollama Cookie 可行格式：浏览器 F12 → Network 找 settings 文档请求（不是 api/v1）→ 复制为 PowerShell → 整段粘贴，程序自动提取 `System.Net.Cookie` 的 name/value
- 无公开 quota API，抓的是 settings 页面 HTML；页面改版需更新 `ollama.go` 解析
- Wiki 指针状态：`docs/wiki/` 12 文件（新增 99 归档页），`.covered-files` 49 项
- GitHub 仓库：https://github.com/eeljoe/quota-viewer（公开，远程落后本地 3 commit，待推送）

---

## 会话记录：2026-08-25 12:13

> **会话摘要**：新增 Command Code 额度监控 Provider（第七平台）——逆向官方 CLI 私有 /alpha API，实现 + 8 个 httptest 用例 + 真实账号冒烟通过；顺手加固注册表空凭证测试隔离用户目录
> **Git**：`d149d3d`（feat）on `master`（本会话；docs 提交见本次）

### 本次完成
- 调研：从 `command-code` npm 包（v1.32.2）`dist/cli.mjs` 逆向出官方 CLI 的额度接口——`GET https://api.commandcode.ai/alpha/whoami`（取 orgId）+ `GET /alpha/billing/credits`（月度剩余 credits + 5 小时/周滚动窗口），认证 `Authorization: Bearer <API Key>`（`user_...`，Studio 生成，与 `~/.commandcode/auth.json` 的 apiKey 同源）
- 新增 `internal/fetcher/commandcode.go`：主展示 = 5 小时窗口用量（驱动球色，与 Ollama 同构），月末剩余与周窗口写入 `Remaining`，`ResetAt` 取 5h 窗口 `resetAt`（epoch ms → UTC ISO）；API Key 留空自动读 `~/.commandcode/auth.json`（用户已登录官方 CLI 免填）
- 注册到 registry（`command-code` / C / KindUsage）+ config `AllProviderIDs` 追加（ensureKnownProviders 自动补全新 Provider，默认关闭）
- 测试：`commandcode_test.go` 8 个 httptest 用例（覆盖空 key、auth.json 自动读取、401、500、正常解析、org 用户带 orgId、结构缺失）；`registry_test.go` 改为 6→7 个并隔离 USERPROFILE（防止空凭证测试读到真实用户凭证/网络）
- 验证：`go test -count=1 ./...` 全绿 / `go vet` 干净 / `wails build` 成功（2m13s，dist/wailsjs 重建产物与 HEAD 逐字节一致仅 hash 变，按惯例回退）；真实环境冒烟通过——`5小时 0.73/3.00 已用 · 周 0.73/6.00 · 余额 $9.27`（individual-go 计划真实数据）
- README 双语平台表 / `docs/ADDING_A_PROVIDER.md` 速查表 + Command Code 特殊说明 / wiki 05 平台表与测试清单同步

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 复用官方 CLI 私有 `/alpha/*` 端点 | Command Code 无公开额度查询 API，CLI 是其唯一数据源 | 抓 Studio 页面 HTML（更脆） / 放弃该平台 |
| 主展示用 5 小时窗口 | 与 Ollama 同构，短窗口耗尽即被限流，球色告警最有价值 | 月度 credits 消耗为主 |
| API Key 留空自动读 `~/.commandcode/auth.json` | 用户本机已登录官方 CLI，复用同一凭证免填 | 仅手动填 Key |
| 注册表测试隔离 USERPROFILE | 空凭证 Build 测试不能读真实凭证或发真实网络 | 单 Provider 特判 |

### 新增/变更文件
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/fetcher/commandcode.go` | Command Code 额度抓取器（/alpha 私有 API） |
| 新增 | `internal/fetcher/commandcode_test.go` | 8 个 httptest 用例 |
| 修改 | `internal/fetcher/registry.go` / `registry_test.go` | command-code 注册 + 测试 6→7 & 隔离 HOME |
| 修改 | `internal/config/config.go` / `config_test.go` | AllProviderIDs +command-code |
| 修改 | `README.md` / `README.zh-CN.md` | 支持平台表 +Command Code |
| 修改 | `docs/ADDING_A_PROVIDER.md` | 速查表 + Command Code 特殊说明 |
| 修改 | `docs/wiki/05-fetching-platforms.md` | 平台表 + 测试清单 + 关键文件 +Command Code |
| 修改 | `docs/STATUS.md` | 本文件 |

> 本次变更（未提交）：约 +430/-30 行，11 个文件

### 未完成 & 下一步
1. **提交并推送**——按项目习惯拆 feat（代码/测试/README/ADDING_A_PROVIDER/wiki）+ docs（STATUS）两个 commit，然后 `git push`（本地领先远程 4 commit）
2. 可选：根目录截图副本 PNG 删除（待用户确认）
3. 可选方向：Release 推广 / 新 Provider 扩展 / 用户反馈迭代

### 关键上下文
- Command Code 端点（逆向官方 cli.mjs，`dist/bundled` 常量 `or/sr/ir/lr`）：`https://api.commandcode.ai/alpha/{whoami,billing/credits,billing/subscriptions,usage/summary}`；credits 响应含 `credits.monthlyCredits/purchasedCredits/freeCredits` + `windowLimits.fiveHour/weekly{used,cap,resetAt(ms)}`
- CLI 认证：优先 `COMMANDCODE_API_KEY` 环境变量，其次 `~/.commandcode/auth.json` 的 `apiKey`；authorization 头为 `Bearer <key>`
- 用户账号现状：individual-go 计划（月度 credits、5h cap $3 / 周 $6），个人账号 org=null（credits 端不带 orgId 参数）
- Wiki 指针状态：`docs/wiki/` 12 文件，`.covered-files` 49 项
- GitHub 仓库：https://github.com/eeljoe/quota-viewer（公开，远程落后本地 4 commit，待推送）

### 会话内补记（修复，2026-08-25 13:0x）
- **OpenCode Go「未找到有效配额窗口」修复**：OpenCode 页面把 `usagePercent` 改成浮点（如 `3.1`），`opencode_go.go` 用 `strconv.Atoi` 解析失败 → 全被当 0 → 误报无有效窗口（真实数据 rolling 3.1 / weekly 43 / monthly 38.3）。改为 `ParseFloat` + `windowInfo.usagePercent` 改 `float64`，新增 2 个浮点用例；真实冒烟 Percent=43.5、`周窗口 · 已用 43.5% · 剩余 56.5%`
- **Command Code Remaining 精简**：原 `5小时 x/x 已用 · 周… · 余额…` 三连串在 340px 面板 210px 区域被省略号截断成乱串、且「5小时 x/x」与进度条重复。去掉 5h 段，Remaining 改为 `周 $1.26/6.00 · 余额 $8.74`（实测）
- `go test ./...` 全绿、`wails build` 成功（产物回退惯例）；wiki 05 同步浮点说明

---

## 会话记录：2026-08-25 23:11

> **会话摘要**：用户重启验证通过——OpenCode Go 浮点修复生效（周窗口 43.5%）、Command Code 剩余文本可读；提交推送与 PNG 清理收尾；截图计划取消
> **Git**：`3f1df7b` on `master`（dirty：本会话 STATUS 更新待提交）

### 本次完成
- （上会话遗留）**重启验证通过**：新 exe 打开后 opencode-go 重新勾选，球格正常——Go 显示周窗口 43.5%、Command Code 显示 5 小时窗口进度 + `周 $1.26/6.00 · 余额 $8.74`，不再截断堆叠
- （上会话遗留）提交推送完成：`231657c`（fix）+ `3f1df7b`（docs）已推送 GitHub（`912c06a..3f1df7b`）
- （上会话遗留）根目录截图副本 PNG 已删除（用户确认）
- 截图计划取消（用户明确"截图就算了"）

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 不截图记录 UI 效果 | 用户明确不需要，功能验证以实际使用为准 | 存 preview-3.png 到 docs/screenshots |

### 未完成 & 下一步
- 无明确待办；可选方向：Release 推广 / 新 Provider 扩展 / 用户反馈迭代 / Wails 版本升级（CLI 2.13.0 vs go.mod 2.12.0）

### 关键上下文
- **当前启用 Provider（球格 3 上限）**：`kimi` / `opencode-go` / `command-code`（ollama 在配置里保持 enabled=false——用户勾选 opencode-go 时被 3 上限钳出让位，凭证都还在，随时可换）
- OpenCode Go 页面 `usagePercent` 已是浮点格式，解析逻辑用 ParseFloat（旧整数用例仍兼容）
- Command Code Remaining 契约：只含辅助信息（周窗口 + 余额），5 小时窗口由进度条主展示
- Wiki 指针状态：`docs/wiki/` 12 文件，`.covered-files` 49 项
- GitHub 仓库：https://github.com/eeljoe/quota-viewer（已同步至 3f1df7b）

> 📦 历史会话已归档（任务组「额度告警修复」2026-09-19 / 2026-09-22 两条）：docs/wiki/99-appendix-legacy-status.md
---

## 会话记录：2026-09-30 22:10

> **会话摘要**：新增 Factory Droid 额度监控（第 8 家 Provider）+ 诊断「保存多次才生效」并修复两个潜在缺陷 + 面板改展示 5小时/周；e2e 确认后推送全部提交并发布 Release v1.2.0
> **Git**：`3a9eec2` on `master`（全部已推送；Release v1.2.0 已发布）
> **任务组**：Factory Droid Provider 接入
> **任务组状态**：已完成（功能已验证并交付运行）

### 本次完成
- **调研**：官方 Analytics API 不适合悬浮球（个人端点 Enterprise 限定、数据滞后一天、是消耗量非剩余额度）；改走与官方 web 端同源的私有路由，字段经 token-monitor（PR #685，对照官方 CLI 与 web bundle）与 CodexBar 双重交叉验证
- **合规检查**（用户重点关切）：个人版条款无「禁查自身用量 / 禁自动化轮询」条款；官方 `fk-` API Key 只读自己数据，非 OAuth 凭证、非共享账号，与 Claude Code 当年封号场景不同；CodexBar 生产运行数月、全网无封号案例。结论：低风险灰色用法；该端点漂移过一次，失效需对照官方更新（与 Command Code 同级维护风险）
- **实现**：`internal/fetcher/factorydroid.go` — 优先 token-rate limits 模型（`/api/billing/limits`：5h/周/月窗口 + Core 池 + Extra 预付余额），旧账单模型兜底（`/api/organization/subscription/usage`）；API Key 留空自动读 `~/.factory/.env`（Droid CLI 同款，支持 export 前缀/引号/行尾注释）；先红后绿 9 用例，`go test ./...` 全绿
- **契约对齐**：Percent = max(全部窗口，含 Core) 延续窗口告警契约；Used/Total/ResetAt 以 5h 主窗口为准（Total=100/Used=5h%，沿 Ollama 先例）；Core 全 0 与 Extra $0 不产生展示噪音
- **真实冒烟**：`GET api.factory.ai/api/billing/limits` 返回结构与实现逐字段吻合（5h 1% / 周 2% / 月 1%，`usesTokenRateLimitsBilling=true`）
- **交付**：杀旧实例（PID 29844）→ `wails build`（16.9s，CLI 用 `C:\Users\joe\go\bin\wails.exe`）→ config 启用 factory-droid（key 走 .env 自动发现）→ 重启新实例（PID 31888）
- **「保存多次才生效」诊断**（用户报疑似缓存 bug）：非缓存——手工改配置造成 4 个同时启用，SaveConfig 的「≤3 静默钳制」把排在最后的 factory-droid 悄悄关掉，用户反复保存才收敛；顺带抓到真 bug：`config.AllProviderIDs` 漏登记 factory-droid（7≠8，全新安装/自动补全路径都不带它）。修复：AllProviderIDs 补齐 + 新增跨包同步守护测试 + Load 时钳制超限启用（配置面板永远不会再出现「勾了 4 个」的不可能状态），先红后绿，全量测试通过，已重建交付（PID 32216）
- **wiki 同步**（`/wiki-update`）：5 文件——02 模块表基线刷到 cc01446（补 Command Code 两行 + Factory Droid 两行）、05 增 Factory Droid 端点/解析细节与 Kimi/Ollama 行为注记、07 增清单同步契约与 Load 钳制说明、09 测试分类更新、00 元数据与「八平台」措辞；覆盖缓存 49→50 项，漂移清零
- **e2e 反馈修复（三轮）**：① 面板 Remaining 只显示周/月（漏 5h 主窗口）→ 补 5h 首段；② 用户仍不通过——与 Kimi/Ollama 的两段式不一致且面板截断 → 常驻只显示 5小时/周，月度仅在驱动告警时追加；③ 澄清计费顺序后 Core 改为「一有消耗即显示」（用户常驻 GLM-5.3-Flash 等 Core 池模型，Standard 烧满后 Core 仓接管，那是他最需要看到的时刻）。全程先红后绿，重建交付（PID 20988）
- **发布完成**（e2e 确认通过后收尾）：推送 6 个提交（`7961a87..3a9eec2`）→ `gh release upload --clobber` 换上新 exe（quota-viewer.exe 11.36MB，含 Core 展示修正）→ publish **v1.2.0**（https://github.com/eeljoe/quota-viewer/releases/tag/v1.2.0，notes 沿用「修复小问题 + 新增 Droid 视图」措辞偏好）

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 用私有 `/api/billing/limits` 而非官方 Analytics API | 后者个人端点 Enterprise 限定 + 数据滞后一天 + 是消耗量非剩余额度 | 等 Factory 官方开放 |
| Key 存 `~/.factory/.env` 而非只存应用配置 | Droid CLI 同样自动读取，一处配置两处生效，应用配置留空即可 | 只存应用配置（重复管理） |
| Percent 含 Core 池 | Core 是独立计费窗口，耗尽同样影响可用性，与「最紧张窗口」契约一致 | 只算 standard（漏报 Core 耗尽） |
| ResetAt 仍取 5h 窗口 | 与 9/22 Kimi 会话决策一致（倒计时跟主窗口） | 取驱动告警窗口的重置时间（仍是待定可选项） |
| 超限钳制放 Load 而非只在 SaveConfig | 配置面板从 GetConfig 渲染，Load 不钳制就会出现「勾了 4 个」的不可能状态，保存时静默被砍（本次用户踩坑的直接原因） | 保存时报错提示（改动更大，前端也要跟上） |
| 月度不常驻展示、仅超 5h/周 时出现 | 用户明确要求与 Kimi/Ollama 两段式一致（且面板会截断）；但月度耗尽驱动告警必须保留（9/19、9/22 长窗口漏报同类坑） | 月度完全不算入 Percent（退回长窗口漏报坑） |
| Core 仓一有消耗即展示 | 计费顺序是「Core 模型先烧 Standard，Core 独立仓只在 Standard 限流后接管」——Core 有读数正是用户进入回退状态的信号（该用户常驻 flash 模型） | 仅驱动告警时显示（回退场景反而看不见） |

### 新增/变更文件
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/fetcher/factorydroid.go` | Factory Droid 抓取器（双账单模型 + .env 自动发现） |
| 新增 | `internal/fetcher/factorydroid_test.go` | 9 个 httptest 用例 |
| 修改 | `internal/fetcher/registry.go` | 注册 factory-droid（缩写 F，LoginURL=api-keys 页） |
| 修改 | `internal/fetcher/registry_test.go` | 注册表断言 7→8 |
| 修改 | `internal/config/config.go` | AllProviderIDs 补 factory-droid + Load 钳制超限启用 |
| 修改 | `internal/config/config_test.go` | +2 用例（清单同步守护 / Load 钳制）+ 夹具补第 8 家 |
| 修改 | `docs/wiki/`（02/05/07/09/00 共 5 文件） | Factory Droid 条目 + 行数基线 + 钳制契约说明 |
| 修改 | `docs/STATUS.md` | 本会话记录 |
| 修改 | `%APPDATA%/quota-viewer/config.json` | 启用 factory-droid（应用外修改） |
| 新增 | `~/.factory/.env` | FACTORY_API_KEY（应用外，不入库） |

### 未完成 & 下一步
- 可选：观察几天，Factory 私有端点漂移时对照官方 web bundle 更新 `factorydroid.go`
- 计划级事项见 `ROADMAP.md`（Next: Wails 版本升级对齐）

### 已知问题 & 注意事项
- Factory 私有路由无稳定性承诺（曾漂移过一次）；Key 失效时报错指向 app.factory.ai/settings/api-keys 重新生成
- Key 曾在对话中明文出现过一次，介意可去 api-keys 页轮换——轮换后只需更新 `~/.factory/.env`，应用配置无需改动
- `~/.factory/.env` 现在被 Droid CLI 与 Quota Viewer 共用，注意不要把它提交进任何仓库
- `frontend/wailsjs/go/models.ts` 显示 M 但 diff 为空（纯行尾噪音）

### 推荐 Skill
- `/wiki-update` - 检测到 2 个 wiki 覆盖文件自 synced_commit（cc01446）后有代码变更：`internal/fetcher/factorydroid.go`、`internal/fetcher/factorydroid_test.go`——但 Core 展示规则的事实已在 `3a9eec2` 同步进 wiki 05，仅 wiki-meta synced_commit 待推进

### 关键上下文
- **Factory 端点**：`GET https://api.factory.ai/api/billing/limits`（Bearer fk- key；请求头 `x-factory-client: web-app` + `Origin`/`Referer: https://app.factory.ai`）；`usesTokenRateLimitsBilling=true` 时读 `limits.standard/core.{fiveHour,weekly,monthly}`（均只有 `usedPercent`/`secondsRemaining`，无绝对值）+ `extraUsageBalanceCents`；否则回退 `GET /api/organization/subscription/usage?useCache=true`（standard/premium：`userTokens`/`totalAllowance`/`usedRatio`；`usedRatio` 有恒 0 脏数据，绝对值可信时优先）
- **Factory 计费结构**（官方 pricing/individuals + models 文档）：Individual 套餐三滚动窗口 5h/7d/30d，三个都有余量才能发请求；Droid Core = 开源权重模型池（GLM/DeepSeek/Qwen/MiniMax/Kimi K 等，倍率 0.06×~1.2×），**Core 模型用量先烧 Standard 窗口**，独立 Core 窗口只在 Standard 限流后接管（`overagePreference` 控制回退到 Core 还是 Extra 预付）；商业模型倍率 0.2×~12×；窗口绝对数值官方未公布
- **Key 来源**：app.factory.ai/settings/api-keys 生成；本机已写入 `~/.factory/.env`（fetcher 留空自动读）
- **参考实现**：token-monitor PR #685（`src/shared/providers/factory/limits.js`）、CodexBar `docs/factory.md`
- 当前启用 Provider：kimi / ollama / command-code / factory-droid（以 `%APPDATA%/quota-viewer/config.json` 为准）
- Wiki 指针状态：`docs/wiki/` 12 文件，`.covered-files` 50 项，synced_commit `cc01446`（漂移已清零，2026-09-30 同步）
- **Release v1.2.0**：https://github.com/eeljoe/quota-viewer/releases/tag/v1.2.0（Latest，附 quota-viewer.exe 11.36MB）
- 桌面 `Quota Viewer.lnk` → `build/bin/quota-viewer.exe`（本会话已重建）

---

## 会话记录：2026-10-02 22:57

> **会话摘要**：新增 MiniMax 额度监控（第 9 家 Provider，Token Plan sk-cp- Key）+ 扩展模式（悬浮球 4-9 格网格）+ 球形悬浮球改版；config.json 中途被误写坏、经应用内存写回完整恢复（零丢失）；推送 `b13a10a`
> **Git**：`b13a10a` on `master`（已推送）
> **任务组**：MiniMax Provider 接入 + 展示形态改版
> **任务组状态**：已完成（冒烟通过，新实例运行中，待用户目验）

### 本次完成
- **调研**：MiniMax Token Plan（M Plan）私有端点 `GET /v1/api/openplatform/coding_plan/remains`（Bearer sk-cp- 订阅 Key + Referer 官方 web 同款）；字段 `current_*_remaining_percent` 是**剩余%**，部分套餐绝对量 count 恒 0；与 minimax-status（JochenYang）CLI 源码、token_manager 调研交叉验证
- **实现 `minimax.go`**：剩余% 反转为已用%（语义反转事故写回归守护测试）；5h + 周双窗口取最紧张者（窗口告警契约对齐 Kimi/Ollama/Factory）；三主机域名链 `api.minimaxi.com → www.minimaxi.com → api.minimax.io`（实测 www 偶发 TLS 握手超时，海外不认国内 Key 返回业务 1004）；注册表第 9 家，缩写 MX，LoginURL 指向套餐页
- **测试**：9 个 httptest 用例先红后绿（语义反转守护 / 周无限制判定 / Total=0 退化 / 域名链回退 / 1004 提示 / 空套餐卡等），`go test ./...` 全绿
- **扩展模式**：`Config.ExtendedMode` + `EnabledLimit`（普通 3 / 扩展 9），Load / SaveConfig / 旧格式迁移三处钳制统一走一个函数；配置面板新增「扩展模式」开关；悬浮球 4+ 格走网格（4=2x2，5-6=3 列两行，7-9=3x3）；详情面板高度随条目伸缩（封顶 640，列表内部滚动）
- **球形改版**：`border-radius: 50%` + 径向高光 + 下缘内阴影营造球体；格子从竖分隔条改为软芯片（1px gap，去 hairline 分隔线）；单格放大占满；2-3 格条带、字号随数量分档
- **球形窗口级修复**（用户报「为什么还是方形」后）：根因两层——① overlapped 窗口系统最小宽度把 60px 球窗钳到 262 物理宽，WebView 实心底色露出方形残影；② Wails OnStartup 与窗口不同线程，`SetWindowSubclass` 必失败（comctl32 限制），历史的最小宽度子类修复从未生效过。修复：`SetWindowRgn` 椭圆 region 窗口级裁圆（圆外点击穿透）+ Collapse 时物理像素 `SetWindowPos` 兜底规整尺寸 + region 切换挂 Expand/Collapse 流程
- **交付**：MiniMax key 写入应用配置并启用 + `extended_mode=true`；真实 key 冒烟（5h 25.0% 已用 · 周 3.0% 已用，与 curl 原始响应逐字段一致，ResetAt 4.2h 后）；杀旧实例（PID 9548）→ `wails build`（30s）→ 新实例 PID 32608；推送 `b13a10a`

### 事故与恢复（如实记录）
- 修 config.json 时 PowerShell 脚本失误（先 `Remove` 属性再引用它），把 `%APPDATA%\quota-viewer\config.json` 覆盖成 133 字节残骸，8 家凭证一度全部丢失
- **恢复路径**：运行中的应用（PID 9548）内存仍持有完整配置 → 用户点击悬浮球触发 `mouseup → SaveBallPosition → config.Save` 整份写回，3147 字节完整恢复，**零丢失**
- **教训**：改应用外配置文件前先确认应用运行状态（运行中实例是热备份）；PowerShell 对 PSObject Remove 属性后原引用即失效

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 取 `model_remains[0]` 主套餐卡 | 与官方 CLI 生态一致；视频/音乐等赠送卡（3/3 之类）不该进主告警 | 遍历全卡取最紧张（赠送卡会污染告警） |
| 三主机域名链逐个回退 | www 实测偶发超时、海外不认国内 Key；单域名故障即全挂 | 只用 api.minimaxi.com（回退能力弱） |
| Total=0 退化为纯百分比展示 | 实测订阅档 count 恒 0，绝对量不可信 | 硬显示 0/0（NaN/误导） |
| 扩展模式默认关、上限 9 | 保持默认极简；9=当前注册表容量 | 默认全开 / 无上限（网格布局无界） |
| 球形用 SetWindowRgn 裁圆而非 WebView 透明 | WebView 透明链路（WebviewIsTransparent/NOREDIRECTIONBITMAP/Mica）逐层试过都有残影，region 从窗口级裁圆最彻底，还白送圆外点击穿透 | WindowIsTranslucent+BackdropType=None（Mica 方块仍露） |

### 新增/变更文件
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/fetcher/minimax.go` | MiniMax 抓取器（域名链 + 剩余%反转 + 百分比退化） |
| 新增 | `internal/fetcher/minimax_test.go` | 9 个 httptest 用例 |
| 修改 | `internal/fetcher/registry.go` + `_test.go` | 注册 minimax（8→9，断言同步） |
| 修改 | `internal/config/config.go` + `_test.go` | ExtendedMode + EnabledLimit/clampEnabled 收敛 + 扩展钳制用例 |
| 修改 | `app.go` | SaveConfig 增 extendedMode 参数；fetchAll 上限动态；GetConfig 回传 extended_mode |
| 修改 | `frontend/src/{index.html,main.js,style.css}` | 扩展开关 + 网格球格 + 球形样式 + 面板伸缩 |
| 修改 | `frontend/wailsjs/go/main/App.{js,d.ts}` | SaveConfig 三参绑定（手改，与生成物一致） |
| 修改 | `frontend/dist/*` | 重建产物 |
| 修改 | `%APPDATA%/quota-viewer/config.json` | minimax 启用 + extended_mode（应用外修改） |

### 未完成 & 下一步
- 用户目验悬浮球（4 格 2x2 网格 + 球形观感）
- wiki 三件套同步（05/02/07/09/00），`/wiki-update`
- 计划级事项见 `ROADMAP.md`（Next: Wails 版本升级对齐）

### 已知问题 & 注意事项
- MiniMax sk-cp- Key 在对话中明文出现过，介意可去 platform.minimaxi.com/console/plan 轮换，轮换后更新应用配置即可（无自动发现机制，需手工填）
- MiniMax 私有端点无稳定性承诺；漂移时对照 minimax-status 源码或官方 web bundle 更新 `minimax.go`（已入 ROADMAP Later）
- `~/.minimax-config.json`（minimax-status CLI）与本应用凭证相互独立，不共享
- 前端行尾警告（dist/wailsjs LF→CRLF）依旧，不影响构建

### 关键上下文
- **MiniMax 端点**：`GET https://api.minimaxi.com/v1/api/openplatform/coding_plan/remains`（Bearer sk-cp-；Referer `https://platform.minimaxi.com/`）；成功响应 `model_remains[]` + `base_resp.status_code=0`；鉴权失败 1004/2049（HTTP 200 也可能携带）
- **字段语义**：`current_interval_remaining_percent` / `current_weekly_remaining_percent` 是剩余%（0-100）；`remains_time` / `weekly_remains_time` 毫秒级重置倒计时；`current_interval_total_count` 部分套餐恒 0；「无周限」判定 = 周总额 0 且无周剩余%（minimax-status 同款）
- **展示对齐**：Remaining = `5小时 X% 已用 · 周 Y% 已用`（无周限→`周无限制`）；Percent = max(5h 已用, 周 已用)；ResetAt 取 5h 窗口倒计时
- 当前启用 Provider：kimi / ollama / factory-droid / minimax（4 个，`extended_mode=true`）
- **Wails 窗口线程事实**（本次实测）：OnStartup 与窗口不同线程 → `SetWindowSubclass` 必失败（ret=0），凡是依赖子类的窗口修复都无效；窗口操作要么走 Wails runtime，要么拿 hwnd 直调 Win32
- 桌面 `Quota Viewer.lnk` → `build/bin/quota-viewer.exe`（本会话已重建，PID 31076）
