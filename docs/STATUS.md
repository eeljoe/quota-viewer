# 📋 项目进度文档

> **⚠️ 新会话阅读指南（必读）**
>
> 本文档较长，**不要通读全文**。按以下顺序阅读：
>
> 1. 阅读下方「上下文摘要（TL;DR）」快速了解项目状态
> 2. 阅读下方「项目概况」了解项目基本信息
> 3. 如果下方有「知识库」区块且状态为已初始化，读取知识库获取项目结构和函数映射
> 4. **直接跳到最新一条「会话记录：2026-08-13 01:33」章节**（用标题搜索定位，不要假设在文件末尾）
>    - 重点看「未完成 & 下一步」和「关键上下文」
> 5. 如果最新章节引用了更早的内容，再按需回溯
> 6. 给用户一份简短进度汇报（3-5句话）：当前在做什么 → 做到哪了 → 下一步是什么 → 有无阻塞项
> 7. **严禁在用户给出指示之前修改任何文件或代码**
> 8. 汇报后等待用户指示，不要主动开始执行任务
>
> 📍 最新章节位置：`## 会话记录：2026-08-13 01:33`（搜索定位，可能不在文件末尾）
> 🔖 对应 commit：`9927e2e` on `master`
> 📊 累计会话数：主文档 4 条，另有 3 条已归档（docs/wiki/99-appendix-legacy-status.md）

---

## 最后更新

<!-- git-meta: {"last_commit": "9927e2e", "branch": "master", "dirty": false, "timestamp": "2026-08-13T01:33:00Z"} -->

- **日期**：2026-08-13 01:33
- **会话摘要**：Ollama 真实账号冒烟通过（PowerShell Cookie 整段粘贴 → 测试成功 → 球格展示 5 小时用量）；触发旧会话归档

---

## 上下文摘要（TL;DR）

- 项目：Quota Viewer，桌面悬浮球 + AI 平台额度监控工具，Go + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite）
- 当前阶段：**v1.0.0 已开源 + 持续迭代**——六平台监控全部实测可用（Ollama 冒烟通过）；工作区干净
- **当前唯一目标**：无明确待办
- 下一步：推送 master（本地领先远程 3 commit）；可选——根目录 PNG 副本清理 / 新 Provider 扩展
- 注意事项：根目录遗留一张未跟踪截图副本 PNG（preview-2.png 的副本，可删除）；无阻塞项

---

## 项目概况

- **项目名称**：Quota Viewer
- **技术栈**：Go 1.24 + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite 打包）
- **项目根目录**：`C:/Users/joe/Desktop/工作学习/软件开发/quota viewer`
- **平台**：Windows 10+（WebView2 运行时）
- **功能**：悬浮球（1-3 格动态，颜色=状态）+ 展开面板（进度条 + 剩余量明细）+ 配置面板（勾选 Provider + 各平台凭证）+ 系统托盘 + 关闭到托盘
- **支持 Provider**：Kimi（API Key）、讯飞星辰（Cookie）、OpenCode Go（Workspace ID + Token）、小米 MiMo（Cookie）、DeepSeek（API Key，余额型 + 预算进度条）、Ollama（Cookie，Cloud 5 小时/周用量）

## 当前分支与最近提交

- **分支**：master
- **HEAD**：`9927e2e`（工作区干净，仅遗留 1 张未跟踪截图 PNG）
- **远程**：origin/master 落后本地 3 个提交（a59635f / cc82c71 / 9927e2e 未推送）
- **最近提交**：
  - `9927e2e` - docs: 更新 STATUS 追加会话5补记与会话6(Ollama)
  - `cc82c71` - docs: 同步 wiki 补齐预算换算模块与余额契约
  - `a59635f` - feat: 新增 Ollama Cloud 额度监控 Provider
  - `39aeaed` - feat: DeepSeek 余额预算进度条 + 展开面板刷新倒计时
  - `22185dc` - docs: README 新增效果截图

---

## 当前任务进度

### ✅ 已完成（截至 2026-08-13）
- **六平台监控**：Kimi / 讯飞星辰 / OpenCode Go / MiMo / DeepSeek / Ollama——fetcher 注册表驱动，新增平台前端零改动
- **DeepSeek 预算条**（8/7）：QuotaResult Balance/Currency + `budget.go` ApplyBudget + ProviderConfig.Budget + 前端预算输入/色阈值/刷新倒计时
- **Ollama Provider**（8/13）：`ollama.go` HTML 解析 5 小时 Session 主窗口 + 周用量；14 个 httptest 用例；config 自动补全新 Provider（默认关闭）；真实账号冒烟通过（8/13 01:33）
- **v1.0.0 开源发布**：GitHub eeljoe/quota-viewer + Release v1.0.0（exe 附件）

### 🔄 进行中
- 无

### 📋 待办
- 推送 master（本地领先远程 3 commit）
- 根目录截图副本 PNG 清理（待用户确认）

---

## 未完成 & 下一步

1. **推送 master 到远程**——本地领先 3 个提交（a59635f / cc82c71 / 9927e2e），确认后 `git push`
2. 可选：根目录截图副本 PNG 删除（用户确认后）
3. 可选方向：Release 推广 / 新 Provider 扩展 / 用户反馈迭代 / Wails 版本升级（CLI 2.13.0 vs go.mod 2.12.0）

---

## 已知问题与注意事项

- Wails CLI 版本 v2.13.0 > go.mod 的 v2.12.0（构建警告，不阻塞；可选升级）
- 前端行尾警告：dist/wailsjs 文件 LF→CRLF，git 会提示但不影响构建
- 根目录 `09314663d2975b947bfa75fcf46e4769.png` 是 `docs/screenshots/preview-2.png` 的未跟踪副本，可删除（未删，避免误删用户文件）
- Ollama 依赖 ollama.com/settings 页面 HTML 结构（无公开 quota API），页面改版 → "页面结构可能已变化"错误，需更新 `ollama.go` 解析
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
