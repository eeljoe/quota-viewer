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
> 5. **直接跳到最新一条「会话记录：2026-10-28 23:21」章节**（用标题搜索定位，不要假设在文件末尾）
>    - 重点看「未完成 & 下一步」和「关键上下文」
> 6. 如果最新章节引用了更早的内容，再按需回溯
> 7. 给用户一份简短进度汇报（每段 3-5 行）：**当前进度**（当前在做什么 → 做到哪了 → 下一步是什么 → 有无阻塞项）+ **上个会话做了什么**（最新会话章节的摘要、完成与遗留）
> 8. **严禁在用户给出指示之前修改任何文件或代码**
> 9. 汇报后等待用户指示，不要主动开始执行任务
>
> 📍 最新章节位置：`## 会话记录：2026-10-28 23:21`（搜索定位，可能不在文件末尾）
> 🔖 对应 commit：`3f73a08` on `master`（v1.4.1 已发布并标 Latest；本会话只动本地 spec 与文档）
> 📊 累计会话数：主文档 4 条，另有 21 条已归档（docs/wiki/99-appendix-legacy-status.md）

---

## 最后更新

<!-- git-meta: {"last_commit": "3f73a08", "branch": "master", "dirty": false, "timestamp": "2026-10-28T23:21:00+08:00"} -->

- **日期**：2026-10-28 23:21
- **会话摘要**：规划 **Devin（第 11 家 Provider）**——查清可用额度通道（Devin 没有公开额度 API）并用本机 Devin Desktop 登录态实测打通，产出 `.scratch/devin-provider/`（spec + 3 张票），完成 Fidelity / Repo-fit / Soundness 三轴审票并按结论重切票面；本会话未动产品代码。

---

## 上下文摘要（TL;DR）

- 项目：Quota Viewer，桌面悬浮球 + AI 平台额度监控工具，Go + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite）
- 当前阶段：**v1.4.1 已发布（2026-10-25，Latest，附 exe）**，本地 HEAD `3f73a08`；当前在手 = 新增 **Devin（第 11 家）** 的 spec 与票（`.scratch/devin-provider/`，已过三轴审票），下一会话 `/implement devin-provider` 落地
- 十家 Provider 除 DeepSeek（余额型）外均为用量型；状态灯契约 60% 黄 / 80% 红 / 100% 熄灭；usage 型一律「最紧张窗口」驱动 Percent；**一个监控对象 = 一组凭证 × 一张套餐卡**，悬浮球上限按对象数计（普通 9 / 扩展 16，超出显示 `+N`）
- Open Design 只能靠浏览器会话 Cookie 读额度（API key `sk-amr-` 一律 401），读法已写进 README 与 `docs/ADDING_A_PROVIDER.md`
- 下一步：下一会话 `/implement devin-provider`（spec + 3 张票已过三轴审票）；「对端告知」报文待发（必须带上 probe `schema` 1→2 的契约变更）；非阻断清理批次 12 条与 ROADMAP Now 栏超条（4 条 > 上限 3 条）待用户拍板（见本地 `ROADMAP.md`「Now」，该文件已不入仓库）
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
- **Devin Provider（第 11 家）**：spec 与 3 张票已出并过三轴审票（`.scratch/devin-provider/`），下一会话 `/implement devin-provider`
- 「对端告知」报文待发给 subagent-fallback（必须带上 probe `schema` 1→2）
- 非阻断清理批次 12 条待用户拍板（清单见本文件「已知问题」与 ROADMAP「Now」）
- 计划级事项见 `ROADMAP.md`

---

## 未完成 & 下一步

1. **下一会话 `/implement devin-provider`**（首要）：spec 与 3 张票已过三轴审票；按票序 01（手填 token 全链路，超限票先跑静态双轴）→ 02（本机登录态自动读取）→ 03（文档与真机冒烟），整体门照 `.scratch/devin-provider/review-leads.md` 的三条预测亲手复验
2. **「对端告知」报文**：回 `subagent-fallback` 说明探针就绪并给出二进制名/开关名/构建命令，**必须写明 `schema` 已从 1 升到 2**（记录新增 `slot` / `sub_label` 两个键，纯增量，按名取键的旧消费方不受影响），否则对端按旧契约解析会对不上
3. 非阻断清理批次 12 条待用户拍板（改代码的几条会作废对应票的 VERIFIED、需重验）
4. Wails 版本升级对齐（Next）：go.mod v2.12.0 vs CLI v2.13.0，构建时仍打印该告警
5. 前端渲染逻辑的可视化验证缺口：见最新会话「关键上下文」——改前端 DOM 结构前先确认探针覆盖到，否则「构建过 = 没事」不成立

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
- **Devin 额度通道（2026-10-28 实测）**：没有公开额度 API（v2/v3 consumption 仅 Enterprise 管理员可用，self-serve 一律 403）；可用的是 Devin Desktop 的 `exa.seat_management_pb.SeatManagementService/GetUserStatus`（CONNECT POST，`metadata.api_key`，`ide_version`/`extension_version` 用 `1.126.0`）—— 返回日/周**剩余**百分比与 epoch 秒重置时间（protojson 把 int64 写成字符串），展示前必须反转成已用；本机登录态在 `%APPDATA%\Devin\User\globalStorage\state.vscdb` 里明文（`devin-session-token$…` + `apiServerUrl`）；403 = token 失效、400 = 缺 `api_key`。探针报文（schema 2）只有十个键，没有 `reset_label` 与 `remaining` 文本，真机判据只能看 `percent` / `remaining_percent` / `reset_at`

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


---

## 会话记录：2026-10-25 18:45

> **会话摘要**：用户现场报告「Factory Droid 最紧张的月额度看不到、倒计时也不对」——单会话轻量路线修好展示口径（最紧张窗口打头 + 倒计时标明所属窗口），重建 exe、重启实例，截图确认
> **Git**：`284a497` on `master`（实现提交；随后 `19c29fe` 同步构建产物、文档跟进，master 领先 origin 3 个提交）
> **任务组**：Factory Droid 窗口展示口径（无 spec 拆票，单会话轻量路线）
> **任务组状态**：已完成（已构建并重启实例；未切版本号、未发 release、未推送）

### 本次完成

- **定位到两处根因**。① 面板头部那行（`.quota-remaining`）只有 210px 宽 + `text-overflow: ellipsis`，而 `Remaining` 按「5小时 → 周 → 月」固定顺序拼串，月百分比排第三段必然被截掉 —— 可球色与进度条是由月窗口（最紧张）驱动的，于是出现「条是红的、数字看不见」；② 进度条下的倒计时取的是 5 小时窗口的 `secondsRemaining`，跟驱动告警的窗口不是同一个。
- **展示口径改成「按紧张度排序 + 倒计时跟排头」**。`Remaining` 各段按已用百分比降序拼接，最紧张的窗口打头；Core 接管时 Core 段仍固定首位（保住 10-13 那条「球看着满了、其实 Core 还有量」的语义），`Extra` 恒居末位。新增 `QuotaResult.ResetLabel`（`json:"reset_label,omitempty"`），`ResetAt` 改取排头那段窗口的重置时间并标明窗口名。
- **前端倒计时带窗口名与天粒度**。`formatCountdown(iso, label)` 渲染成「距月额度刷新: 4天21时」，≥24 小时按天计（此前月窗口只能显示「288时41分」）；`reset_label` 经 `data-reset-label` 落位，30 秒一次的定时刷新会读回，不会被抹成通用措辞。
- **真机验证与交付**：`wails build` 重建 23s；旧实例（PID 28328）结束后重启（PID 22364）；桌面 `Quota Viewer.lnk` 仍指向 `build/bin/quota-viewer.exe`（11,662,336 字节，18:42:40）。截图确认 Factory 行现为「月 98.0% 已用 · Core 45.0% 已用 · 5小…」，倒计时「距月额度刷新: 4天21时」。
- **验收**：一个静态双轴审查（Standards/Spec，无 blocking）+ 一个独立动态验收门（VERIFIED：`go vet`/`go test ./...` 全绿、既有探针 7/7、自写探针 10/10、4000 组属性夹具不变量成立、三条指定变异全被杀且还原复绿）。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| 排头窗口缺 `secondsRemaining` 时倒计时整行不显示 | 用户 2026-10-19 拍板；与既有「驱动窗口缺秒数就留空、不借别的窗口」口径一致 | 退一步显示 5 小时窗口的倒计时并标明来源 |
| 只改 Factory，不动 Kimi / Ollama 等兄弟 provider | 用户只报了这一家；其余平台倒计时仍锚 5 小时窗口（ROADMAP Later 的 Kimi 打磨项仍在） | 顺手把「倒计时跟驱动窗口」推广到全部 usage 型 provider |
| 单会话轻量路线（1 个静态审查 + 1 个动态验收） | 改动 4 文件 +207/−60，不触认证/schema/并发/构建配置 | 严格规划 → 出票 → 逐票 worker |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|---|---|---|
| 修改 | `internal/fetcher/factorydroid.go` | `parseTokenRateLimits` 改为「段 + 紧张度排序 + 排头驱动倒计时」；`factoryDrivingWindow` 同时返回窗口名 |
| 修改 | `internal/fetcher/types.go` | 新增 `ResetLabel`（`json:"reset_label,omitempty"`） |
| 修改 | `internal/fetcher/factorydroid_test.go` | 新增 `TestFactoryDroid_TightestWindowLeadsDisplay`；AC 表顺序/倒计时期望同步到新口径 |
| 修改 | `frontend/src/main.js` | 倒计时带窗口名与天粒度；`data-reset-label` 写入与读回 |
| 新增 | `.scratch/factory-tightest-window/` | 本票 spec + 票 + 倒计时探针（不入库） |

> 本次变更：+207/−60 行，4 个源码文件（另有构建产物 `frontend/dist` 与 `wailsjs` 绑定随 `19c29fe` 入库）
> 变更基准：`git diff 211047c..284a497 --stat`

### 未完成 & 下一步

1. **要不要发 v1.4.1 并推送**（新）：master 领先 origin 3 个提交，未打 tag、未发 release —— 等用户定。
2. **对端告知报文**（持续中）。
3. **非阻断清理批次 12 条**（持续中）。
4. **Kimi 倒计时口径**（持续中，ROADMAP Later）：「倒计时取正在驱动告警的窗口」这次只在 Factory 落地，Kimi 仍固定 5h。

### 已知问题 & 注意事项

- **GUI 像素级截断没测**：无头环境只能按 CSS 规则 + 字宽估算。排头段一定落在可见区内，但三段全展开（约 245px > 210px）时第二段尾部仍可能出现省略号 —— 那不是告警来源，可接受。
- **探针契约**：`internal/probe` 的 `schema` 仍为 2、记录 10 键形状未变；`reset_at` 现在随最紧张窗口漂移（此前固定 5h），`reset_label` 未进报文。消费方若要区分窗口，才需要 schema 3 纯增量补键。

### 关键上下文

- Factory 私有路由 `/api/billing/limits` 的三窗口只有 `usedPercent` + `secondsRemaining`，没有绝对值；`Percent` = 全部窗口（含 Core 池）最大值，球色阈值 60% 黄 / 80% 红 / 100% 熄灭。
- 本次固定点：SPEC_BASELINE `211047c` → 实现 `284a497`。静态审查 1 个实例（无 blocking），动态验收 1 个独立实例（VERIFIED）。前端探针在 `.scratch/factory-tightest-window/probe/countdown-probe.js`（7/7）。
- 环境：Go 不在 PATH（`C:\Users\joe\sdk\go1.26.5\bin`），wails CLI 在 `C:\Users\joe\go\bin`；实例在跑时 `wails build` 会把旧 exe 改名成 `quota-viewer.exe~` 留在 `build/bin`（本次构建前先杀了实例，故没有残留）。
- Wiki 指针状态：`docs/wiki/` 同步 00 / 02 / 05（Factory bullet 的排序与倒计时口径、契约清单补 `reset_label`、模块清册行数与说明），`synced_commit` 刷到 `284a497`；文件数仍 11。

---

## 会话记录：2026-10-25 19:18

> **会话摘要**：发布 **v1.4.1**（Factory 展示口径修复）——打 tag、推 master、建 release 并上传 exe；随后按任务组归档旧会话
> **Git**：`326f962` on `master`（已推 origin；tag `v1.4.1` 已推）
> **任务组**：Factory Droid 窗口展示口径（承 2026-10-25 18:45 那条）
> **任务组状态**：已完成（已发版 v1.4.1）

### 本次完成

- **v1.4.1 发布**：`gh release create --draft` → 单独 `gh release upload` 上传 `build/bin/quota-viewer.exe`（11,662,336 字节）→ `gh release edit --draft=false --latest` 转正。沿用「先建草稿、再单独传产物、最后转正」的做法绕开上传超时。GitHub 返回的资源 digest `sha256:0e65bc85…a4eae0` 与本地 `Get-FileHash` 逐字节一致；v1.4.1 已是 **Latest**，v1.4.0 退为普通版本。
- **推送**：`master` 三个提交（`284a497` 修复 + `19c29fe` 构建产物 + `326f962` 文档）推上 origin，注释 tag `v1.4.1` 一并推远端。
- **发布说明**只写可见行为变化（额度行按紧张度排序、倒计时跟最紧张的窗口并标明窗口、超过一天按天显示），末尾一句「修复了一些简单的小问题」—— 按用户的发布口径不写 bug 细节（含这次「月额度被截断」的病灶）。
- **会话归档**（Step 8）：主文档追加本章后 413 行 ≥ 400 → 把已关闭的「同 URL 多 Key / 多套餐」任务组（2026-10-18 21:00 + 2026-10-19 00:50 两条）整组原样搬进 `docs/wiki/99-appendix-legacy-status.md`，主文档回落到 304 行，归档页 19 → 21 条。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| 发 v1.4.1 而不是复用 v1.4.0 | 展示口径变了、用户看得见，patch 语义合适 | 直接改 v1.4.0 的 release notes |
| 发布说明不写 bug 细节 | 用户长期口径 | 把「月额度被截断」写进 notes |
| 归档 Multi-Key 组、保留 Factory 组 | 该组最新会话标了「任务组状态：已完成」且已合并发版；Factory 组是最新活跃组，永不归档 | 什么都不归档（主文档继续涨） |

### 未完成 & 下一步

1. **对端告知报文**（持续中）。
2. **非阻断清理批次 12 条**（持续中）。
3. **Kimi 倒计时口径**（持续中，ROADMAP Later）。
4. **ROADMAP Now 栏超条**（新）：4 条超过 3 条上限，等用户定要不要把「前端渲染探针纳入常规验证」降到 Next。

### 已知问题 & 注意事项

- `gh release view --json isLatest` 在本机 gh 版本会报「Unknown JSON field」，改用 `gh release list` 看 Latest 标记。
- 发布产物校验用：`gh api repos/eeljoe/quota-viewer/releases/tags/v1.4.1 --jq '.assets[] | "\(.name) | \(.size) | \(.digest)"'`，与本地 `Get-FileHash -Algorithm SHA256` 对比。

### 关键上下文

- 当前线上：**v1.4.1 是 Latest**，release 页 `https://github.com/eeljoe/quota-viewer/releases/tag/v1.4.1`；发布提交 `326f962` 已在 origin/master，本章的文档提交随后跟进。
- Wiki 漂移检测：`synced_commit` = `284a497`，自它以来只有文档与构建产物变更（`frontend/dist/*`、`frontend/wailsjs/go/models.ts`），都不在 wiki「关键文件」表覆盖的文件里 → 无漂移。
- 归档页现有 21 条会话（新增 Multi-Key 组的 2 条）；主文档保留最近 2 条（Factory 组）+ 无关闭信号的 2026-08-25。

---

## 会话记录：2026-10-28 23:21

> **会话摘要**：规划 Devin（第 11 家 Provider）——查清额度通道并用本机登录态实测打通，产出 spec 与 3 张票，三轴审票后按结论重切票面；未动产品代码
> **Git**：`3f73a08` on `master`（本会话只产出本地 spec 与文档；`.scratch/` 与 `ROADMAP.md` 本就忽略）
> **任务组**：Devin Provider（新开组）
> **任务组状态**：规划完成，待实现

### 本次完成

- **通道勘察（实测，不是推测）**：Devin 没有可用的公开额度 API —— v2/v3 的 `/enterprise/consumption/*` 只对 Enterprise 管理员个人 Key 开放，self-serve 一律 403（官方文档写明「no per-account toggle for support to enable」）。可用通道是 Devin Desktop（Windsurf 内核）的 `exa.seat_management_pb.SeatManagementService/GetUserStatus`：CONNECT 协议 POST，`metadata.api_key` 带本机登录态 token，一次返回日/周两个窗口的**剩余**百分比、各自重置 epoch（protojson 把 int64 写成字符串）、`hideDailyQuota`/`hideWeeklyQuota`、`overageBalanceMicros` 与套餐名。
- **本机实测**：用本机登录态（`%APPDATA%\Devin\User\globalStorage\state.vscdb` 里明文 `devin-session-token$…` 与 `apiServerUrl=https://server.codeium.com`）打通 —— HTTP 200、套餐 Pro（`TEAMS_TIER_DEVIN_PRO`、`BILLING_STRATEGY_QUOTA`）、日/周剩余均 100%、日重置 `1793260800`、周重置 `1793520000`；错误面同样实测：403 = token 失效、400 = 缺 `api_key`。
- **产出（本地、不入库）**：`.scratch/devin-provider/spec.md`（17 条 Acceptance Statements + 接口契约）、3 张票（01 手填 token 全链路 / 02 本机登录态自动读取 / 03 文档与真机冒烟）、`review-leads.md`（三条红队预测）。
- **三轴审票**：Fidelity 5 条、Repo-fit 12 条、Soundness 8 条；其中「原票 05 要探针输出 `reset_label` / `remaining`，而探针契约只有十个键」被两个轴独立命中。按结论把原「请求 / 展示 / 凭据」三张水平票重切成三张能力票，并修掉 server 兜底顺序、错误路径必须保留 `Platform`、README 两处平台数、`config_test` 钳制断言、`registry_test` 函数名与 `APPDATA` 隔离等全部机械问题。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| 改走两段式，不在本会话实现 | 同类 provider 先例落地约 400-450 行、6 个文件，超轻量路线 300 行 / 3 文件上限 | 硬塞进轻量预算（验收强度达不到严格口径） |
| 凭证留空自动读本机 Devin Desktop 登录态 | 与 factory-droid / command-code 同形，用户不必手抄 token | 只支持手填 token |
| 一颗球、最紧张窗口驱动 | 与 Ollama / Factory / MiniMax 同一套告警口径，只占一格 | 日/周各占一格 |
| 不引 SQLite 驱动，按字节扫 `state.vscdb` | 只取两个字段，引驱动等于为一个取值加大依赖 | 引入纯 Go SQLite 驱动 |
| 票按能力切而不是按层切 | 审票判定层切的三张各自产不出可演示增量 | 保留原 5 张水平票（复核后否掉） |

### 未完成 & 下一步

1. 下一会话 `/implement devin-provider`：按票序 01（超限票，先静态双轴）→ 02 → 03；整体门照 `review-leads.md` 的三条预测亲手复验。
2. **AS-11b 界面目验必须由用户本人做**：球格出现 `DV`、面板数字与 Devin Desktop 的 Usage 页一致。
3. 三条旧尾巴不动：对端告知报文、非阻断清理 12 条、ROADMAP Now 栏超条。

### 已知问题 & 注意事项

- Devin 的通道与语义已写进上文「已知问题与注意事项」区（剩余百分比要反转、int64 是字符串、探针没有 `reset_label` / `remaining`）。
- 本会话用到的 token 只在内存与探针命令里出现过，**没有落进任何仓库文件**；spec 与票在 `.scratch/`（gitignore）。

### 关键上下文

- 产品级事实全部来自真实请求：`server.codeium.com` 的 `GetUserStatus` 在 `ide_version`/`extension_version = 1.126.0`、UA 为浏览器型时返回 200；缺 `api_key` 返回 400 `invalid_argument`，坏 token 返回 403 `permission_denied`。
- 本机 Devin Desktop 版本 `1.126.0`（`state.vscdb` 的 `devin.desktopStartup.lastLaunchedBuild`），与请求体里的版本号一致。
- Wiki 本轮未动（无代码变更）；实现落地后再 `/wiki-update` 增量 00 / 02 / 05 / 07 / 08。
