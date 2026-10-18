# 📋 项目进度文档

> **⚠️ 新会话阅读指南（必读）**
>
> 本文档较长，**不要通读全文**。按以下顺序阅读：
>
> 1. 阅读下方「上下文摘要（TL;DR）」快速了解项目状态
> 2. 阅读下方「项目概况」了解项目基本信息
> 3. 如果下方有「知识库」区块且状态为已初始化，读取知识库获取项目结构和函数映射
> 4. 阅读项目根目录 ROADMAP.md 的「Now」栏（滚动计划，了解接下来做什么）
>    - 注：`ROADMAP.md` 自 2026-10-14 起**仅在本地存在、不入公开仓库**（与 `.scratch/`、`.agents/` 同）；若你拿到的是 clone，此文件缺席属正常，计划内容可临时向维护者索取
> 5. **直接跳到最新一条「会话记录：2026-10-19 00:50」章节**（用标题搜索定位，不要假设在文件末尾）
>    - 重点看「未完成 & 下一步」和「关键上下文」
> 6. 如果最新章节引用了更早的内容，再按需回溯
> 7. 给用户一份简短进度汇报（每段 3-5 行）：**当前进度**（当前在做什么 → 做到哪了 → 下一步是什么 → 有无阻塞项）+ **上个会话做了什么**（最新会话章节的摘要、完成与遗留）
> 8. **严禁在用户给出指示之前修改任何文件或代码**
> 9. 汇报后等待用户指示，不要主动开始执行任务
>
> 📍 最新章节位置：`## 会话记录：2026-10-19 00:50`（搜索定位，可能不在文件末尾）
> 🔖 对应 commit：`62f81ff` on `master`（已与 origin/master 同步；tag `v1.4.0`）
> 📊 累计会话数：主文档 3 条，另有 19 条已归档（docs/wiki/99-appendix-legacy-status.md）

---

## 最后更新

<!-- git-meta: {"last_commit": "62f81ff", "branch": "master", "dirty": false, "timestamp": "2026-10-19T00:50:00+08:00"} -->

- **日期**：2026-10-19 00:50
- **会话摘要**：会话归档（`/read-status` 触发的 Step 2c 机械动作）——把 6 个已关闭任务组的会话章节（2026-09-30 Factory Droid 接入 / 2026-10-02 + 10-03 MiniMax 接入与展示形态改版 / 2026-10-11 15:14 Open Design 接入 / 16:10 悬浮球格布局打磨 / 10-13 Provider 漂移修复与 `--json` 探针）原样搬进 `docs/wiki/99-appendix-legacy-status.md`，主文档 754 → 366 行（低于 400 行阈值）。主文档保留最近 3 条 + 无关闭信号的 2026-08-25。**零产品代码改动。**

---

## 上下文摘要（TL;DR）

- 项目：Quota Viewer，桌面悬浮球 + AI 平台额度监控工具，Go + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite）
- 当前阶段：**v1.4.0 已发布（2026-10-19，附 exe）**，master 与 origin/master 同步在 `62f81ff`；本轮增量 = 多 Key / 多套餐的交付与收尾（代码 + 文档均已推送）
- 十家 Provider 除 DeepSeek（余额型）外均为用量型；状态灯契约 60% 黄 / 80% 红 / 100% 熄灭；usage 型一律「最紧张窗口」驱动 Percent；**一个监控对象 = 一组凭证 × 一张套餐卡**，悬浮球上限按对象数计（普通 9 / 扩展 16，超出显示 `+N`）
- Open Design 只能靠浏览器会话 Cookie 读额度（API key `sk-amr-` 一律 401），读法已写进 README 与 `docs/ADDING_A_PROVIDER.md`
- 下一步：「对端告知」报文待发（必须带上 probe `schema` 1→2 的契约变更）；非阻断清理批次 12 条待用户拍板（见本地 `ROADMAP.md`「Now」，该文件已不入仓库）；STATUS 已再次按任务组整组归档至 314 行（低于阈值）
- 注意事项：**`wails build` 的产物字节与构建目录绑定**（未加 `-trimpath` 时 Go 把源码绝对路径编进二进制）⇒ 核验交付 exe 必须在同目录重建；**公开仓库自 2026-10-14 起只含产品代码与对外文档**——`.scratch/`、`.agents/`、`.video_agent/`、`.superpowers/`、`docs/superpowers/`、`AGENTS.md`、`ROADMAP.md` 全部已进 `.gitignore`，提交仍应精确 `git add` 而非 `-A`（ignore 只管未跟踪文件，已跟踪文件不受保护）



---

## 项目概况

- **项目名称**：Quota Viewer
- **技术栈**：Go 1.24 + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite 打包）
- **项目根目录**：`C:/Users/joe/Desktop/工作学习/软件开发/quota viewer`
- **平台**：Windows 10+（WebView2 运行时）
- **功能**：悬浮球（1-9 格动态，颜色=状态，方圆角）+ 展开面板（进度条 + 剩余量明细）+ 配置面板（勾选 Provider + 各平台凭证 + 扩展模式开关）+ 系统托盘 + 关闭到托盘
- **支持 Provider**：Kimi（API Key）、讯飞星辰（Cookie）、OpenCode Go（API Key，留空读 opencode CLI 凭证）、小米 MiMo（Cookie）、DeepSeek（API Key，余额型 + 预算进度条）、Ollama（Cookie，Cloud 5 小时/周用量）、Command Code（API Key，5 小时/周窗口 + 剩余 credits）、Factory Droid（API Key，双账单模型 + Core 池）、MiniMax（Token Plan 订阅 Key，5 小时/周窗口）、Open Design（浏览器会话 Cookie，Design Plan 5 小时/7 天窗口）
- **同屏上限**：注册表 10 家，悬浮球同屏最多 9 格（普通模式 3 格 / 扩展模式 9 格，`config.go:35-36`）；此为用户拍板的「10 选 9」决策

## 当前分支与最近提交

- **分支**：master
- **HEAD**：`62f81ff`
- **最近提交**（2026-10-14 历史重写后的现行 hash；此前记录的 `d6c09a4`/`b5537c2` 等已作废，仅存于备份 bundle）：
  - `62f81ff` - feat: 多 Key / 多套餐各占一格（监控对象粒度 + 悬浮球上限改按对象计）
  - `38b370f` - docs: 记录仓库收口会话（内部产物移出版本控制 + 历史重写）
  - `91dacb0` - chore: 内部过程产物与计划文档移出版本控制，补全 .gitignore 忽略规则
  - `ec0f8c7` - docs: 追加 2026-10-14 门面刷新会话记录与 ROADMAP 对账
  - `18a0a5e` - docs: README 补齐十家平台清单与官网链接，新增命令行探针与 Open Design 章节
- **Release**：v1.4.0（2026-10-19，Latest，附 exe，sha256 `05a18782…`）← v1.3.0 ← v1.2.0 ← v1.1.2 ← v1.1.1 ← v1.1.0 ← v1.0.0

---

## 当前任务进度

### ✅ 已完成（截至 2026-09-30）
- **八平台监控**：Kimi / 讯飞星辰 / OpenCode Go / MiMo / DeepSeek / Ollama / Command Code / Factory Droid——fetcher 注册表驱动，新增平台前端零改动
- **状态灯契约**（9/19）：60% 黄 / 80% 红 / 100% 熄灭（off 暗灰）；ollama/commandcode 的 Percent 取 5h 与周窗口较紧张者；发布 v1.1.1
- **DeepSeek 预算条**（8/7）：QuotaResult Balance/Currency + `budget.go` ApplyBudget + ProviderConfig.Budget + 前端预算输入/色阈值/刷新倒计时
- **Ollama Provider**（8/13）：`ollama.go` HTML 解析 5 小时 Session 主窗口 + 周用量；14 个 httptest 用例；config 自动补全新 Provider（默认关闭）；真实账号冒烟通过（8/13 01:33）
- **Command Code Provider**（8/25）：`commandcode.go` 逆向官方 CLI 私有 `/alpha/*` API（whoami + billing/credits），Bearer API Key（留空自动读 `~/.commandcode/auth.json`）；8 个 httptest 用例；真实账号冒烟通过
- **Factory Droid Provider**（9/30）：`factorydroid.go` 走官方 web 同源 `/api/billing/limits`（5h/周/月 + Core 池 + Extra 余额，旧账单模型兜底），Key 留空自动读 `~/.factory/.env`；9 个 httptest 用例；真实冒烟通过；发布 v1.2.0
- **MiniMax Provider + 扩展模式**（10/2）：`minimax.go` 走 `coding_plan/remains`（sk-cp- 订阅 Key，三主机域名链，剩余% 反转已用%）；扩展模式启用上限 3→9（配置面板开关）+ 4+ 格网格布局；9 个 httptest 用例；真实冒烟通过（球形改版 10/3 经目验否决，已回滚）
- **开源发布**：GitHub eeljoe/quota-viewer，Release v1.0.0 / v1.1.0 / v1.1.1 / v1.1.2 / v1.2.0（均附 exe）
- **多 Key / 多套餐 → 一监控对象一格**（10/19，v1.4.0）：监控对象 = 一组凭证 × 一张套餐卡；`ProviderConfig.Extra` 多凭证组（稳定 Key 对齐）+ 可选接口 `MultiFetcher`/`FetchAll`；`SettleIdentity` 统一归一 slot/sub_label/abbr（app.go 与 probe 同口径）；MiniMax / MiMo 转多对象（此前 `model_remains[0]`、首个 `items` 之外的额度被静默丢弃）；悬浮球上限由 Provider 数改为对象数（普通 9 / 扩展 16，超出显示 `+N`）；配置面板凭证组增删；probe `schema` 1→2

### 🔄 进行中
- 无

### 📋 待办
- 「对端告知」报文待发给 subagent-fallback（必须带上 probe `schema` 1→2）
- 非阻断清理批次 12 条待用户拍板（清单见本文件「已知问题」与 ROADMAP「Now」）
- 计划级事项见 `ROADMAP.md`

---

## 未完成 & 下一步

1. **「对端告知」报文**（首要）：回 `subagent-fallback` 说明探针就绪并给出二进制名/开关名/构建命令，**必须写明 `schema` 已从 1 升到 2**（记录新增 `slot` / `sub_label` 两个键，纯增量，按名取键的旧消费方不受影响），否则对端按旧契约解析会对不上
2. 非阻断清理批次 12 条待用户拍板（改代码的几条会作废对应票的 VERIFIED、需重验）
3. Wails 版本升级对齐（Next）：go.mod v2.12.0 vs CLI v2.13.0，构建时仍打印该告警
4. 前端渲染逻辑的可视化验证缺口：见最新会话「关键上下文」——改前端 DOM 结构前先确认探针覆盖到，否则「构建过 = 没事」不成立

---

## 已知问题与注意事项

- Wails CLI 版本 v2.13.0 > go.mod 的 v2.12.0（构建警告，不阻塞；已入 ROADMAP Next）
- 前端行尾警告：dist/wailsjs 文件 LF→CRLF，git 会提示但不影响构建
- `build/bin/quota-viewer.exe~` 是运行中构建产生的旧 exe 改名残留，可删
- Ollama 依赖 ollama.com/settings 页面 HTML 结构（无公开 quota API），页面改版 → "页面结构可能已变化"错误，需更新 `ollama.go` 解析
- Command Code 依赖官方 CLI 私有 `/alpha/*` API（无公开额度 API），CLI 升级可能改结构 → 需对照官方 cli.mjs 更新 `commandcode.go`；API Key 留空自动读 `~/.commandcode/auth.json`（与本机官方 CLI 复用同一凭证，Key 变更会同时影响两者）
- 余额型（DeepSeek）经 ApplyBudget 按预算换算消耗百分比（默认预算 300）；取首个非零余额币种（CNY→¥ / USD→$），全部为 0 才报错
- **构建可复现性口径**：`wails build` 的产物字节与**构建目录**绑定（未加 `-trimpath` 时 Go 嵌入源码绝对路径）——`worktree`/副本里重建必然与主工作区交付件 hash 不同；核验"交付 exe 出自这个 commit"必须在**同一目录**重建比对，或用 `wails build -trimpath`（换了 flag 就是另一份字节，发版中途别换）
- **托盘库语义（`energye/systray@v1.0.3` Windows 后端）**：真双击会依次触发 `onClick(0x202) → onDClick(0x203) → onClick(0x202)`；`onClick == nil` 时库对单击什么都不做；注册 `onRClick` 会顶掉默认右键弹菜单；图标标识 = 类 `SystrayClass` + `uID=100` + 回调消息 `0x401`（`Shell_NotifyIconGetRect` 可跨进程取图标物理矩形）
- **注入点击类探针的硬坑**：本机主屏 2880x1800 @200%，未声明 DPI 感知的进程自报 1440x900 ⇒ 注入前必须 `SetThreadDpiAwarenessContext(-4)`，否则坐标 2 倍错位、点击落空；PowerShell 的 `$PID` 是只读变量，进程参数只能叫别的名字（本项目用 `-ProcId`）

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

- **位置**：`docs/wiki/`（已初始化 2026-08-04；最近同步 2026-10-18，锚点 `62f81ff`）
- **文件数**：13 个（00-agent-rules ~ 10-build-test-baseline + 99 归档页 + `.covered-files` 缓存）
- **入口**：先读 `docs/wiki/00-agent-rules.md`（含索引与 wiki-meta 同步状态）
- **缓存**：`docs/wiki/.covered-files`（59 项；本轮新增 `abbr.go` / `abbr_test.go` 两项覆盖文件）
- **最近一次同步**：2026-10-18 多 Key / 多套餐任务组改了 02 / 05 / 07 / 08 四页；2026-10-19 收尾无代码改动，wiki 无漂移
- **99 归档页**：19 条会话（最近一次归档 2026-10-19，托盘单击唤出 / 门面刷新 / 仓库收口三个已关闭任务组）

---

## 关键文件索引

- `app.go` - Wails 主应用（窗口定位 fitToScreen、监控对象编排 `fetchAll`、配置）
- `main.go` - 入口 + Wails 运行时配置（frameless / AlwaysOnTop）
- `workarea_windows.go` / `workarea_other.go` - Win32 工作区查询辅助（非 Windows 桩）
- `internal/fetcher/registry.go` - **Provider 注册表（新增 Provider 的唯一入口）**
- `internal/fetcher/types.go` - Fetcher / MultiFetcher 接口、QuotaResult（含 Slot / SubLabel）、Kind 常量
- `internal/fetcher/abbr.go` - `SettleIdentity`（缩写/后缀/slot 归一）+ 缩写派生（app.go 与 probe 共用）
- `internal/fetcher/{kimi,xfyun,opencode_go,mimo,deepseek,ollama,commandcode,factorydroid,minimax,opendesign}.go` + `budget.go` - 十平台抓取器（均有测试）+ 余额型预算换算
- `internal/config/config.go` + `cookie.go` - 动态 Provider 配置 + 多凭证组（`CredGroup`）+ 旧格式迁移 + PowerShell Cookie 解析
- `internal/tray/tray.go` - 系统托盘（systray.Run + LockOSThread）
- `frontend/src/index.html` / `main.js` / `style.css` - 动态球格（含 `+N` 溢出格）+ 多组凭证配置面板
- `frontend/wailsjs/` - Wails v2 前端绑定（构建生成）
- `.scratch/multi-key/probe/{ball-render,config-render}-probe.js` - 前端渲染探针（**不入库**；改前端 DOM 后必跑）
- `docs/wiki/05-fetching-platforms.md` - 新增 Provider 指南详情
- `docs/ADDING_A_PROVIDER.md` - **Agent 专用**新增 Provider 完整指南（fetcher 实现 + 注册 + 测试 + 检查清单）

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

> 📦 历史会话已归档：docs/wiki/99-appendix-legacy-status.md

> 📦 历史会话已归档：docs/wiki/99-appendix-legacy-status.md

## 会话记录：2026-10-18 21:00

> **会话摘要**：同 URL 多 Key / 多套餐——用户拍杼「一格一监控对象」（每个 Key / 每张套餐卡各一格）。新增 `CredGroup` 多凭证组与 `MultiFetcher` 可选接口，上限从「Provider 数」改为「监控对象数」（普通 9 / 扩展 16），超出显示 `+N`；`probe --json` 升到 `schema: 2`。分支 `feat/multi-key-multi-plan`，未提交、未推、未发 release。
> **Git**：分支 `feat/multi-key-multi-plan`（从 `38b370f` 分出），工作区有未提交改动
> **任务组**：同 URL 多 Key / 多套餐（无 spec，用户直接指令）
> **任务组状态**：代码与文档已落地（待用户目验球格 4×4 档观感）

### 本次完成

- **问题证据**（改代码前先确认）：「多套餐」**今天就在静默丢数据**——`minimax.go:122` 只取 `model_remains[0]`（注释自己写着「视频/音乐等赠送卡在后」），`mimo.go:98` 只取首个有 limit 的 item（`compensation_total_token` 补偿额度被忽略，测试夹具里就有这个 item）；「多 Key」**根本无法表达**——`Creds` 是 `map[string]string`，一个字段只能一个值。
- **数据模型**：`ProviderConfig.Extra []CredGroup`（第 2..N 组凭证，`Key` 是前端生成、保存时原样带回的稳定标识——**删中间一行不得让后面的凭证值整体错位**）；`CredSets()` 跳过全空附加组（加了行没填不该凭空多出一格报错的监控对象）；**旧配置零迁移**（`Extra` 为空时数据模型与旧版一致）。
- **多套餐接口**：`fetcher.MultiFetcher` 可选接口 + `fetcher.FetchAll(f)` 统一入口（未实现的自动包成单元素；违约返回空切片也退回单结果，避免上层拿到零监控对象而静默不显示这家）；`minimax.go` / `mimo.go` 已转多对象，全局失败仍返回单条错误结果。Kimi 的 `limits[]` / `data[]` 是**同一账号的窗口/模型维度**，不是独立套餐，**故不拆**。
- **身份归一**（`internal/fetcher/abbr.go` 新文件）：`SettleIdentity(r, def, credNo, planIdx, planCount, credCount)` 统一填 `slot` / `sub_label` / `abbr`——**app.go 与 probe 两处必须同口径**（否则球面与 `--json` 对不上）；单对象时清空后缀并沿用 Provider 缩写（保持旧观感），多对象时主对象留 Provider 缩写、套餐卡按卡名派生（CJK 首字 / ASCII 两字）、第 2..N 把 Key 用「首字母 + 序号」（K1 / K2）；`planAbbrSetter` 保证同平台内缩写不重复。
- **引擎重编排**：`app.go fetchAll` 由「Provider × 凭证组× 套餐卡」展开，并发后按「注册表 → 凭证组 → 套餐卡」顺序展平（用 `perJob [][]QuotaResult` 而不是直接 append，避免并发顺序乱）；**抓取层不截断**，截断留给展示层，展开面板永远列全部对象。
- **上限与布局重拍（用户拍杼的那一份）**：`config.MaxObjects` 普通 **9** / 扩展 **16**（上限按凭证组数计，套餐卡数拿到才知）；球格布局 1-3 单行 / 4 → 2×2 / 5-9 → 3 列 / 10-16 → 4 列，新增 `cells-10-16` 字号档；**超上限不静默丢弃**——末格变 `+N`，tooltip 写清还有多少个、去哪里看（本仓同屏 9 格上限过去的前车之鉴：门面写「十家」不说上限，用户勾满十家看到第十格无声消失）。上限取 9 / 16 的依据：悬浮球只有 **60×60px**，9 格 = 每格约 20px，16 格 = 约 15px，5×5 每格 12px 会把「讯」与两字缩写磨成一团。
- **前端**：配置面板每家可「+ 再加一组凭证」且附加组可删（主组无删除按钮，避免误删主凭证），每组带独立「测试」（带 `cred_index`）；面板行展示后缀（套餐名 / `Key 2 · 视频赠送卡`）；对象占用提示与扩展模式切换即重排球格。后端字段返回到前端的 HTML 前做了 `escapeHtml`（新增的套餐名是上游文本）。
- **probe 同步**：`--json` 记录粒度从「一家一条」改为「一对象一条」，`schema` 1 → **2**，新增 `slot` / `sub_label` 两个键（纯增量，按名取键的旧消费方不受影响）；`enabled_limit` 现在是对象上限。**这条连带着「对端告知」那条待办的报文——发出去前要带上 schema 变更。**
- **验证**：`go test ./...` 全绿（config/fetcher/probe 三包，含新增的多套餐 / 多凭证组 / 身份归一用例）；`wails build` 通过且绑定重生成（`TestConnection(platform, credIndex)`）；**真实凭证端到端**：`quota-viewer-cli.exe --json` 跑真实配置出 6 条对象（slot 均填正、单对象平台 sub_label 为空、`schema=2 limit=16`），用临时配置（MiniMax 真 Key + 一把无效 Key）实测出 2 条独立对象 `minimax/0 Key 1 pct=13` / `minimax/1 Key 2（报鉴误密码 1004）`。
- **渲染探针**（`.scratch/multi-key/probe/{ball-render,config-render}-probe.js`，未入库）：把 `main.js` 里**真实的** `updateBall` / `maxEnabled` / `getStatusColor` 源码抽出来跑在最小 DOM stub 上，断言 1-20 个对象下的格子数 / 列数 / 字号档 / 溢出格 / tooltip；**变异体（清空溢出格斆字）被杀死** —— 证明断言有牙。
- **文档**：README 两版同步（上限口径、多凭证、`schema: 2` 记录形态示例）；wiki 05 / 07 / 08 / 02 同步（`.covered-files` +2）；`docs/ADDING_A_PROVIDER.md` 补「这家平台有几个额度池」的分流；ROADMAP 换库。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| 用户拍杼「一格一对象」后直接实现，不再汇报 | 用户已在三个候选里明确选它；报表式形式“当中的紧迫额度最高”不适合本仓「一眼看完全部平台状态灯」的设定 | 聚合取最紧张力 / 配置时手选 |
| 上限从 3/9 改 9/16，按**对象**而非 Provider 计 | 一家挂 3 把 Key 就占 3 格，按 Provider 计会让真正的用户永远触顶 | 保留 3/9 并强制单 Key |
| 超上限用 `+N` 而不是静默丢弃 | 本仓名庪的前车之鉴（≤3 静默钳制慢慢关掉最后一个 provider）；门面文档也因此补了 AS10 | 直接只展前 N 格 |
| `CredGroup.Key` 用前端生成的稳定标识对齐 | 按位置合并时，删掉中间一行会把后面的凭证值整体错位到别的 Key 上 | 按位置合并（会有真 bug） |
| Kimi 不拆多对象 | `limits[]` / `data[]` 是同一账号的**窗口 / 模型维度**，不是独立套餐；拆会把一个 provider 变成几个无意义的格 | 按模型名拆格 |
| `probe` 记录加键并升 `schema` | 球面与 `--json` 不能对不上（本仓一直强调同口径）；加键是纯增量，旧消费方按名取键不受影响 | 保持 8 键、只在多对象时多输出行（但同 slot 无法区分） |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|---|---|---|
| 新增 | `internal/fetcher/abbr.go` | `SettleIdentity` + 缩写派生（app.go 与 probe 共用的一处口径） |
| 修改 | `internal/config/config.go` | `CredGroup` + `CredSets()` + `MaxObjects`(9/16) + `CountObjects` |
| 修改 | `app.go` | `fetchAll` 展开监控对象；`GetConfig` 返回 `extra`/`object_limit`/`object_count`；`SaveConfig` 按 Key 对齐凭证组；`TestConnection(platform, credIndex)` |
| 修改 | `frontend/src/{main.js,style.css,index.html}` | 球格 9/16 布局 + `+N` 溢出格；凭证组增删；面板行后缀 |
| 修收 | `internal/probe/probe.go` | `schema: 2`，记录 10 键，一对象一条 |

### 用户目验回弦（同日 23:30）与修复

- **用户截图直接暴露一个我没看过的症状**：配置面板里所有 Provider 都只有一行标题，凭证输入框与「+ 再加一组凭证」按钮**一个都没渲染**。
- **根因（我的代码）**：重写 `renderProviderList` 时把原有的 `card.appendChild(fieldsWrap)` 丢了 —— `fieldsWrap` 建好、被 `buildGroup` 填满，却从未挂到卡片上。之前只探了 `updateBall`（球格渲染），没有探配置面板的渲染函数，所以没拦下。
- **顺带拿到的第二个 bug**（新探针写完才暴露）：`groupFilled` 只看 `input.value`，而 `GetConfig` 回显的掩码在 **`placeholder`** 上 —— 已保存的第二组凭证被当成空组，面板下方「占用 N / 上限」提示偏小（后端 `CredSets()` 认存下来的凭证，两边口径不一致）。已改为同时看 `value` 与 `placeholder`。
- **为什么之前没探到**：本仓写前端时始终只拿 Go 测试、真实凭证探针、以及一个针对 `updateBall` 的渲染探针做验证 —— **配置面板的 DOM 结构从未被自动验证过**。现补上 `.scratch/multi-key/probe/config-render-probe.js`（抽真实 `renderProviderList` 源码跑 DOM stub，断言卡片挂载、凭证组数、删除按钮仅在附加组、每组测试按钮的 cred 下标、对象计数与提示文案），并带变异体（删掉 `card.appendChild(fieldsWrap)` 后断言变红）。

### 未完成 & 下一步

1. **待用户目验（重新）**：配置面板加第二组凭证 → 看「+ 再加一组凭证」按钮是否出现、加完是否多出一格、面板是否标出「Key 2」；并看 4×4 / 5-9 格档的字号观感是否过小。
2. **未提交未推**：分支 `feat/multi-key-multi-plan` 上是一批工作区改动（包含本会话的 STATUS 归档与本轮文档同步），等目验过后一趟单 commit。
3. **旧布局兼容性**：那些只记录 3/9 的历史会话章节已归档到 99 \9875（历史不改）；当前 README / wiki / ROADMAP 已全部改伺 9/16 口径。
4. **未发 release**：格子与展示形态都变了，发版前应由用户目验过后决定是否负责说明。

### 关键上下文

- **交付态**：`build/bin/quota-viewer.exe`（sha1 `40AF97C1F146BFC869E3EB483D582E023B53B06F`，11.11 MB，23:33:28；含两个前端修复），桌面 `Quota Viewer.lnk` 指向该文件（已核对 TargetPath + mtime），实例已重启运行。
- **环境笔记**：Go 不在 PATH，在 `C:\Users\joe\sdk\go1.26.5\bin\go.exe`（GOROOT），wails CLI 在 `C:\Users\joe\go\bin\wails.exe`；两者都需先加 PATH 才能跑。
- **Wiki 指针状态**：`docs/wiki/` 同步 02 / 05 / 07 / 08（`.covered-files` 57 → 59，新增 `abbr.go` / `abbr_test.go`）。
---

## 会话记录：2026-10-19 00:50

> **会话摘要**：多 Key / 多套餐任务组收尾——用户目验通过并拍板，单 commit 推 master、发 **v1.4.0**（附 exe）、并按任务组归档旧会话
> **Git**：`62f81ff` on `master`（已与 origin/master 同步；tag `v1.4.0`）
> **任务组**：同 URL 多 Key / 多套餐（无 spec，用户直接指令）
> **任务组状态**：已完成（已合并 master 并发市 v1.4.0）

### 本次完成

- **用户目验通过**。会话末尾下令推 GitHub、发新 Release、`/update-status` 收尾。上一条会话留下的四项未完成全部落地：多 Key 配置流程目验通过、已提交推送、已发 release、旧布局口径已在 README / wiki / ROADMAP 全部改为 9/16。
- **单 commit 交付**（`62f81ff`，33 文件 +1999/-759）：按本仓惯例直接快进 master（历史线性、无合并提交），推 `origin/master`。精确暂存不用 `git add -A`（项目惯例）。
- **v1.4.0 发布**：tag 已推，Release 标题「多 Key / 多套餐各占一格」，notes 只写可见能力变化（多 Key / 多套餐 / 球容量与排布 / `+N` 溢出 / `schema` 1→2），末尾一句「修复了一些简单的小问题」——**按用户的发布口径，不写 bug 细节**（包括会话中那个前端渲染缺陷）。
- **资源与可回滚性核对**：上传的 `quota-viewer.exe`（11,653,632 字节）GitHub 返回 sha256 `05a1878268edc3f6021aef0bb77d722729147371a603e148d33a8570c330ff6a`，与本地 `build/bin/quota-viewer.exe`（sha1 `40AF97C1F146BFC869E3EB483D582E023B53B06F`）逐字节一致；v1.4.0 已标记 **Latest**，v1.3.0 退为普通版本。
- **会话归档**（update-status Step 7）：将 3 个已关闭任务组原样搬入归档页——托盘单击唤出 / 门面刷新 / 仓库收口（2026-10-14 三条）；主文档 430 → 263 行（低于 400 阈值），归档页现共 19 条。保留 2026-08-25（无关闭信号，永不归档）与本任务组两条。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| 一质体化为 **v1.4.0** 而不是 v1.3.1 | 格子与展示形态都变了（对象上限 3/9 → 9/16、多凭证组、新增缩写派生），patch 语义不合 | 发 v1.3.1 小版 |
| 发布 notes 不写 bug 细节 | 用户长期口径：release 只写可见行为变化 + 一句「修复了一些简单的小问题」 | 把前端渲染缺陷写进 notes |
| 归档时把 2026-10-14 三条一并搬走 | 主文档 431 行超 400 阈值；三条均标「任务组状态：已完成」且无分支依赖 | 保留全部供新会话回溯（会进入无界增长） |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|---|---|---|
| 提交 | `62f81ff` | 本任务组全部代码 + 文档（33 文件） |
| 发布产物 | `build/bin/quota-viewer.exe` | v1.4.0 附件；sha256 `05a1878268edc3f6021aef0bb77d722729147371a603e148d33a8570c330ff6a` |
| 产物探针（未入库） | `.scratch/multi-key/probe/{ball-render,config-render}-probe.js` | 球格与配置面板的渲染探针，各自杀死一个变异体 |

### 未完成 & 下一步

1. **对端告知报斈**（持续中）：回 `subagent-fallback` 那边的探针可用性报文必须带上 **`schema` 已从 1 升到 2**（记录新增 `slot` / `sub_label`，按名取键的旧消费方不受影响），否则对端按旧契约解析会对不上。
2. **非阻断清理批次 12 条**（持续中）：仍待用户拍板，本轮未动。
3. **Wails 版本对齐**（Next）：go.mod v2.12.0 vs CLI v2.13.0，本轮构建仍有该告警，未阻塞。

### 关键上下文

- **交付态**与 v1.3.0 时一致的口径：`wails build` 产物字节与构建目录绑定（未加 `-trimpath` 时 Go 把源码绝对路径编进二进制）。本轮 exe 在主工作区构建，桌靯 `Quota Viewer.lnk` 指向它。
- **`gh release create` 上传 exe 易超时**：本轮第一次调用超 600s 被切，只创出了**空草稿**（无资源）。应对：先 `gh release create --draft` 快速建稿，再单独 `gh release upload` 传产物，最后 `gh release edit --draft=false`。本轮超时的个空草稿需须自己检查并发布（`gh release view <tag> --json isDraft,assets`）。
- **前端可视验证的缺口**：本轮用戸截图暴露了「配置面板凭证区整块没渲染」的自伤缺陷——Go 测试与球格渲染探针都绿了却都绕开了这段 DOM。**以后动前端渲染逻辑，先想一想有没有探针能覆到它**，不要默认「构建过就没事」。
- **环境笔记**：Go 不在 PATH，在 `C:\Users\joe\sdk\go1.26.5\bin\go.exe`；wails CLI 在 `C:\Users\joe\go\bin\wails.exe`；两者都需先加 PATH。
- **Wiki 指针状态**：`docs/wiki/` 本轮同步 02 / 05 / 07 / 08（`.covered-files` 57 → 59）；本次收尾无代码改动，wiki 无漂移。
