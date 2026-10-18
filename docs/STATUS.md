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
> 5. **直接跳到最新一条「会话记录：2026-10-14 16:05」章节**（用标题搜索定位，不要假设在文件末尾）
>    - 重点看「未完成 & 下一步」和「关键上下文」
> 6. 如果最新章节引用了更早的内容，再按需回溯
> 7. 给用户一份简短进度汇报（每段 3-5 行）：**当前进度**（当前在做什么 → 做到哪了 → 下一步是什么 → 有无阻塞项）+ **上个会话做了什么**（最新会话章节的摘要、完成与遗留）
> 8. **严禁在用户给出指示之前修改任何文件或代码**
> 9. 汇报后等待用户指示，不要主动开始执行任务
>
> 📍 最新章节位置：`## 会话记录：2026-10-18 21:00`（搜索定位，可能不在文件末尾）
> 🔖 对应 commit：`38b370f` on `master`（分支 `feat/multi-key-multi-plan` 从它分出；本轮改动均未提交）
> 📊 累计会话数：主文档 5 条，另有 16 条已归档（docs/wiki/99-appendix-legacy-status.md）

---

## 最后更新

<!-- git-meta: {"last_commit": "38b370f", "branch": "feat/multi-key-multi-plan", "dirty": true, "timestamp": "2026-10-18T21:00:00+08:00"} -->

- **日期**：2026-10-18 19:10
- **会话摘要**：会话归档（`/read-status` 触发的 Step 2c 机械动作）——把 6 个已关闭任务组的会话章节（2026-09-30 Factory Droid 接入 / 2026-10-02 + 10-03 MiniMax 接入与展示形态改版 / 2026-10-11 15:14 Open Design 接入 / 16:10 悬浮球格布局打磨 / 10-13 Provider 漂移修复与 `--json` 探针）原样搬进 `docs/wiki/99-appendix-legacy-status.md`，主文档 754 → 366 行（低于 400 行阈值）。主文档保留最近 3 条 + 无关闭信号的 2026-08-25。**零产品代码改动。**

---

## 上下文摘要（TL;DR）

- 项目：Quota Viewer，桌面悬浮球 + AI 平台额度监控工具，Go + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite）
- 当前阶段：**v1.3.0 已发布（2026-10-14，附 exe）+ 门面已刷新 + 仓库已收口**，master 与 origin/master 同步在 `38b370f`（历史已重写，旧 hash 全部作废）；本轮增量 = 会话归档（纯文档）
- 十家 Provider 除 DeepSeek（余额型）外均为用量型；状态灯契约 60% 黄 / 80% 红 / 100% 熄灭；usage 型一律「最紧张窗口」驱动 Percent；**注册表 10 家但悬浮球同屏上限 9 格（普通 3 / 扩展 9）**——门面文档已显式写明
- Open Design 只能靠浏览器会话 Cookie 读额度（API key `sk-amr-` 一律 401），读法已写进 README 与 `docs/ADDING_A_PROVIDER.md`
- 下一步：非阻断清理批次 12 条待用户拍板（见本地 `ROADMAP.md`「Now」，该文件已不入仓库）；STATUS 已按任务组整组归档至 366 行（低于阈值），本轮不再重复归档
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
- **HEAD**：`38b370f`
- **最近提交**（2026-10-14 历史重写后的现行 hash；此前记录的 `d6c09a4`/`b5537c2` 等已作废，仅存于备份 bundle）：
  - `38b370f` - docs: 记录仓库收口会话（内部产物移出版本控制 + 历史重写）
  - `91dacb0` - chore: 内部过程产物与计划文档移出版本控制，补全 .gitignore 忽略规则
  - `ec0f8c7` - docs: 追加 2026-10-14 门面刷新会话记录与 ROADMAP 对账
  - `18a0a5e` - docs: README 补齐十家平台清单与官网链接，新增命令行探针与 Open Design 章节
  - `2f07a2b` - docs: STATUS 补记托盘菜单退出项已实测闭环(4/4)与实例重启
- **Release**：v1.3.0（2026-10-14，Latest，附 exe）← v1.2.0 ← v1.1.2 ← v1.1.1 ← v1.1.0 ← v1.0.0（本轮不发新 release）

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

### 🔄 进行中
- 无

### 📋 待办
- 非阻断清理批次 12 条待用户拍板（清单见本文件「已知问题」与 ROADMAP「Now」）
- 同 URL 多 Key / 多套餐的处理方式待定（本轮新议题，见分支 `feat/multi-key-multi-plan`）
- 计划级事项见 `ROADMAP.md`

---

## 未完成 & 下一步

1. 非阻断清理批次 12 条待用户拍板（改代码的几条会作废对应票的 VERIFIED、需重验）
2. **同 URL 多 Key / 多套餐**：现模型是「1 provider = 1 凭证 = 1 格」，同平台第二个 Key 无法配置、同一 Key 的第 2+ 张套餐卡被静默丢弃（`minimax.go` 只取 `model_remains[0]`、`mimo.go` 只取首个 `items`）——处理方式待拍板
3. 计划级事项见 `ROADMAP.md`（Now: 清理批次 + 对端告知；Next: Wails 版本升级对齐）

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

- **位置**：`docs/wiki/`（已初始化 2026-08-04；最近同步 2026-10-14，锚点 `b5537c2`——该 hash 已随历史重写作废，仅存于备份 bundle）
- **文件数**：13 个（00-agent-rules ~ 10-build-test-baseline + 99 归档页 + `.covered-files` 缓存）
- **入口**：先读 `docs/wiki/00-agent-rules.md`（含索引与 wiki-meta 同步状态）
- **缓存**：`docs/wiki/.covered-files`（59 项；本轮新增 `abbr.go` / `abbr_test.go` 两项覆盖文件）
- **本轮同步**：无代码改动，wiki 无漂移；仅 99 归档页新增 6 条会话（现共 16 条）

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

## 会话记录：2026-10-14 01:40

> **会话摘要**：托盘图标单击唤出窗口到前台（撤销双击切换）——三轴审票 → 1 票实现 → 逐票验收门 VERIFIED → 全范围静态双轴；同批把这 25 个提交推到 GitHub 并发 **v1.3.0**（附 exe）
> **Git**：`b5537c2` on `master`（v1.3.0 的代码点；其后仅文档提交）
> **任务组**：托盘单击唤出（`.scratch/tray-activate/`）
> **任务组状态**：已完成

### 本次完成

- **需求与决策**（用户拍板）：多会话时球可能被盖住/已隐藏，而托盘图标单击没反应 ⇒ ① 单击 = 唤出到前台（显示 + 激活 + 重申置顶），幂等；② 撤销双击切换（显示/隐藏留在右键菜单）；③ 范围只做唤出 + 重申置顶；④ 改完推 GitHub 并发 v1.3.0。
- **seam 与实现**（`3530e59`，4 文件 +17/−7）：`internal/tray/tray.go` 注册 `SetOnClick` → 发新事件 `tray:activate`（删掉 `SetOnDClick`）；`app.go` 新增 handler（`visible.Store(true)` → `WindowShow` → `WindowSetAlwaysOnTop(true)`）；README 两版同步。**不需要自写 Win32 抢焦点**——出票前亲核 Wails v2.12 的 `WindowShow` 内部已含 `SetForegroundWindow` + `SetFocus`（`windows/frontend.go:958-991`）。
- **逐票验收门 VERIFIED**（独立实例、真机证据）：① 单击唤出（`hidden` 红→绿；基线 `89a04fa` 干净重建对照 `visible=false`）② 被盖住到前台（`before_fg=explorer(89212)` → `after_fg=球窗(64956, wailsWindow)`）③ 幂等（间隔 650ms 两次单击仍可见且前台）④ 双击不隐藏（真双击对拍 + 50ms 采样 0 次翻转；基线同注入 t=358ms 隐藏）⑤ AC5a/5b/6/7/8/9 逐条亲验。变异 A（注释掉 `WindowShow`）令场景变红 ⇒ 绿灯非恒真。
- **全范围静态双轴**（整体门）：规范轴 0 blocking（无 Must NOT Change 违规、精确暂存、凭证扫描零命中、提交内 blob 与 gofmt 输出逐字节一致；并确认"托盘无 Go 测试"属本仓既有惯例）；规格轴 0 blocking（seam 逐字落地、AC1–AC9 有落点、0 越界、0 少做）。
- **文档/注释订正**（`b5537c2`，纯注释 + README，**重建 exe 字节未变** `98173463…`）：托盘注释把"删双击"的理由改成实测机制（真双击 = `onClick→onDClick→onClick`，净效果是唤出→隐藏→再唤出、球闪一下）；包注释订正事件转发去向；README 两版补明"双击不再切换"。
- **交付**：推 master（25 个提交）+ tag `v1.3.0` + GitHub Release（附 `quota-viewer.exe`；notes 只写可见能力、不提 bug 细节）。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| 单击 = 唤出、双击撤销 | 实测保留双击的净效果是"唤出→隐藏→再唤出"（闪一下）而非净隐藏，且语义歧义 | 保留双击（多一次无意义的状态翻转） |
| 不自写 Win32 抢焦点 | Wails `WindowShow` 已含 `SetForegroundWindow` + `SetFocus` | 自写 `AttachThreadInput` 兜底（本票不需要） |
| 探针不入库 | 本仓 `.scratch/` 与 `.agents/` 都不在 git（公开仓库，内部笔记不发布） | 移进 `probes/` 入库 |
| 集成轴中止 | 该轴在这张 17 行的小票上反复起实例、耗时失控；它的三个目标（exe 溯源 / AC5c 退出 / 集成面无副作用）已由逐票门 + 出票方复核覆盖 | 跑完为止 |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|---|---|---|
| 修改 | `internal/tray/tray.go` | `SetOnClick` → `tray:activate`；删 `SetOnDClick`；包注释订正 |
| 修改 | `app.go` | 新增 `tray:activate` handler（唤出三步） |
| 修改 | `README.md` / `README.zh-CN.md` | 托盘行为描述（单击唤出、双击不再切换） |
| 新增 | `.scratch/tray-activate/{spec.md, issues/01-tray-click-activate.md, review-leads.md}` | spec / 票 / 审票产物（未入库） |
| 新增 | `.scratch/tray-activate/verifier-probes/` | worker 与验收门各自的探针与日志（未入库） |

### 未完成 & 下一步

- **v1.3.0 已发布**；ROADMAP「Now」仍是清理批次 12 条 + 对端告知。
- 逐票门三条非阻断 findings 的处置：exe 字节溯源 = **已证伪**（构建目录差异，同目录重建逐字节一致）；`dclick` 探针场景零牙齿 = 门已给出有牙齿的手法（50ms 采样抓翻转），探针不入库故无仓内遗留；AC5c「退出」项 = 门刻意未点，出票方已实测 `WM_COMMAND id=6` 有效（窗口消失、进程退出，右键菜单四项 **4/4** 闭环），随后重启实例（pid 31704）。
- 整体门的集成轴被中止（见决策），是否补跑由下一会话定。

### 已知问题 & 注意事项

- 见上方「已知问题与注意事项」新增的三条：构建字节与构建目录绑定、托盘库点击语义（双击三段/`onClick==nil`/`onRClick` 顶菜单）、注入点击探针的 DPI 与参数名硬坑。
- STATUS 已超归档阈值（≥400 行），本轮未归档。

### 关键上下文

- **交付态**：`build/bin/quota-viewer.exe` sha1 `98173463E712B14F648D157CE602FDA81D650947`（主工作区重建可逐字节复现），桌面 `Quota Viewer.lnk` 指向该文件。
- **v1.3.0 覆盖的可见变化**（自 v1.2.0）：MiniMax 与 Open Design 两家新平台、扩展模式最多 9 格、无界面 `--json` 探针 `quota-viewer-cli.exe`、OpenCode Go 官方 API Key 通道、Factory Droid 在 Standard 限流时展示 Core 池、托盘单击唤出到前台（双击不再切换）。
- **探针与日志**：worker 探针 `.scratch/tray-activate/verifier-probes/tray-activate.ps1`（`-ProcId` / `-Scenario hidden|covered|idempotent|dclick`）；验收门探针 `.scratch/tray-activate/verifier-probes/gate-r1/`（含 `dclick-fine.ps1` 真双击采样手法与库侧插桩日志）。
- Wiki 指针状态：`docs/wiki/` 12 文件（00–10 + 99），本轮同步 00/02/03/04/06 五页，锚点刷到 `b5537c2`；99 归档页 10 条。

---

## 会话记录：2026-10-14 16:05

> **会话摘要**：门面刷新——GitHub About 从五家平台改到十家 + 两份 README 补齐十家平台清单、官网链接、命令行探针与 Open Design 专节。单会话轻量路线，1 次审票 + 1 次动态验收门（VERIFIED），零产品代码改动，已推 master
> **Git**：`d6c09a4` on `master`（已与 origin/master 同步）
> **任务组**：门面刷新（`.scratch/readme-promo-20261014/`，未入库）
> **任务组状态**：已完成

### 本次完成

- **GitHub About / topics**：描述从「五家平台 + 中英混排」改为**纯英文 270 字符**，十家平台英文渲染名逐个列出，并含 `quota`、`~11 MB exe`、`~60 MB RAM`、`tray`、`JSON CLI probe`；追加 9 个 topics（`open-design`/`minimax`/`deepseek`/`ollama`/`kimi`/`opencode`/`ai-quota`/`quota-monitor`/`windows`），原有 7 个全保留，共 16 个；`homepageUrl` 保持为空。原描述与回滚方法已记进票面。
- **两份 README（`d6c09a4`，2 文件 +88/−34）**：① 顶部新增十家平台一行清单（平台名带官网链接）；② 紧跟一行写明**同屏上限 9 格**（普通 3 / 扩展 9）；③ 修正过时表述（英文 `1-3 cells`/`max 3, min 1`、中文 `1-3 格`/`最多 3 个` → 普通上限 3 / 扩展上限 9，并补 1-3 单行、4 格 2×2、5-9 格 3 列布局）；④ 平台表格平台名改为官网链接并新增「类型」列（usage / balance，DeepSeek 余额型一眼可辨）；⑤ 新增「命令行探针」小节（产物名、构建命令、`--json` 一行 JSON、无参退出码 2）；⑥ 新增「Open Design 怎么监控」小节（Design Plan 5 小时 / 7 天窗口、必须用浏览器会话 Cookie 并解释 `sk-amr-` key 401、从 `open-design.ai/cloud/dashboard` 取 Cookie）；⑦ 项目结构树补 `cmd/quotaviewer` 与 `internal/probe`；⑧ 体积改用实测值（exe 10.86 MB / 工作集 59.6 MB），删掉旧 About 里未经核实的 `~50MB vs Electron 200MB+` 对比。
- **审票（compact，1 个 verifier 顺序跑三轴）**：4 条 blocking + 若干 important，全部为票面/spec 的机械缺陷，主会话按纪律直接修正并留痕：① AS1「纯英文」与「逐字列十家平台名」自相矛盾（讯飞星辰/小米 MiMo 是中文名）→ 钉死英文渲染口径；② AS3 判据只有中文 pattern，对英文 README 恒零命中（**假绿**）→ 两份各配一组；③ AS8 白名单含 tracker，与「内部笔记不入公开仓」冲突 → 删去；④ 删无源 Electron 对比。**并采纳其指出的最危险失败点**，新增 AS10（同屏 9 格口径）：注册表 10 家但上限 9 格，若不写明，用户勾满十家会看到第十格无声消失。
- **动态验收门 VERIFIED**（独立实例、亲跑证据）：十条 AS 全过；60 条机械断言 0 失败；**14 个变异体全部被杀死**（文档侧 7 个在副本上做、每例 sha1 还原校验通过；元数据侧 7 个纯内存字符串，未写远端），证明判据有牙而非恒真；AS4 由**真实编译运行的二进制**证明（无参 `exit=2` + stdout 0 字节 + stderr usage；`--json` → `exit=0` 单行合法 JSON）；AS5 两份各 10 链接、逐个相同、11 个绝对 URL 全 200；AS8 diff 只含两份 README；AS9 v1.3.0 仍 Latest、exe 时间戳未变。
- **事实源**：本次改动引用的每个数字/端点都回源到代码行号（格数上限 `config.go:35-36`、刷新默认 15 分钟 `config.go:91`、CLI 退出码 `probe.go:206-259`、Open Design 必须 Cookie `opendesign.go:22`/`:222`、体积为本次实测）。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| 走单会话轻量路线 | 纯文档 + 仓元数据、零代码、用户明确批准；省掉静态 code-review 轴（无代码可审） | 完整规划/执行路线（为 2 份 README 出票派 worker，明显过重） |
| About 用纯英文 | GitHub 描述在搜索与列表页展示，英文覆盖更广；双语会挤掉十家平台名 | 保持中英双语（350 字符内塞十家会砍掉功能描述） |
| 新增 AS10「同屏上限 9 格」 | 门面把「十家」推到最高可见度，却不说同屏上限，用户勾满十家会看到第十格无声消失——本仓已有同类前车之鉴（≤3 静默钳制曾悄悄关掉最后一个 provider） | 只写「支持十家」（验收照样全绿，但用户会踩坑） |
| 不发新 release | 纯文档改动，v1.3.0 的能力描述已经正确，只是写得不够全 | 发 v1.3.1（无代码变化，属噪音） |
| `.gitignore` 不改 | 是否忽略 `.scratch/`/`.agents/` 是范围外决策，涉「内部笔记不发布」从纪律变机制 | 顺手加 ignore（超出本票范围，留给用户拍板） |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|---|---|---|
| 修改 | `README.md` | 十家平台清单 + 官网链接 + 命令行探针 + Open Design 专节 |
| 修改 | `README.zh-CN.md` | 与英文版逐节同构 |
| 修改（仓外） | GitHub 仓库 description / topics | 经 `gh repo edit` 写入，无仓内文件 |
| 新增 | `.scratch/readme-promo-20261014/{spec.md, issues/01, issues/02}` | spec / 票 / 裁决（未入库） |
| 新增 | `.scratch/readme-promo-20261014/verifier-probes/` | 两个门的断言与变异脚本（未入库） |

### 未完成 & 下一步

1. 非阻断清理批次 12 条仍待用户拍板（ROADMAP「Now」）；对端告知报文属仓外动作，仍挂着。
2. 本次验收门给 5 条非阻断 lead（不构成任何 AS 失败），已记进票 02 的 `## Verdicts`：AS7 阈值子项在本 README 上恒真（两份 0 个百分号 token）；探针小节可补 `--json=1` 走 JSON 信封退出 0 的说明；「默认预算 ¥300」写死币种而代码支持 USD；About 写 RAM、README 写 working set，同一数字两种口径；`Windows 10+` 支持基线在本仓查不到出处。
3. **`Windows 10+` 出处未定**——门指出仓内无证据（只有 `go-webview2` 依赖，依赖存在 ≠ 支持基线），该行是本次 diff 的未改动上下文行；需联网核 Wails 官方平台要求后再定措辞。
4. STATUS 已超归档阈值（600+ 行），本轮未归档。

### 已知问题 & 注意事项

- **`.scratch/`、`.agents/`、`AGENTS.md`、`.video_agent/` 未跟踪且未被 `.gitignore` 覆盖**（`.gitignore` 只有 `build/bin`、`node_modules`）⇒ 提交只能精确 `git add`；漏写「禁 `git add -A`」的票会留真实事故面。是否加 ignore **待用户拍板**。
- 其余见上方「已知问题与注意事项」（构建字节与目录绑定、托盘库点击语义、注入点击探针的 DPI 硬坑）。

### 关键上下文

- **交付态**：`d6c09a4` 已推 origin/master；v1.3.0 仍是 Latest；`build/bin/quota-viewer.exe` sha1 `98173463E712B14F648D157CE602FDA81D650947` 未变（本轮未重建）。
- **门给出的判据锚点**：README.md sha1 `C1CFADF6B7AE601BB31F22A766FDA9A69AA576C1`、README.zh-CN.md sha1 `2FC23B0DCFD97C8DFD9B5CCC1B114E3D61E7F95C`；此 hash 之后若再改 README 或仓元数据，本次验收即失效需重验。
- **事实源**：给门面文档写数字前先钉一张带 `文件:行号` 的表（本会话 spec 的「事实源」段即模板）。
- Wiki 指针状态：`docs/wiki/` 12 文件（00–10 + 99），本轮零代码改动，wiki 无漂移；99 归档页 10 条。

---

## 会话记录：2026-10-14 17:15

> **会话摘要**：仓库收口——内部过程产物移出版本控制（补全 `.gitignore`、`ROADMAP.md` 与 `.superpowers/` 取消跟踪），并重写全历史清除已推送的内部文档；强推 master、重建 6 个标签。**代码与对外文档零变化**
> **Git**：`91dacb0` on `master`（历史已重写，旧 hash 作废）
> **任务组**：仓库收口（无 spec，用户直接指令）
> **任务组状态**：已完成

### 本次完成

- **起因**：2026-10-14 门面刷新会话的验收门指出 `.scratch/`、`.agents/`、`AGENTS.md`、`.video_agent/` 处于「未跟踪但无 ignore 兜底」状态——`git add -A` 会把内部笔记扫进公开仓库。用户随后拍板加 ignore，并要求把 `ROADMAP.md`、`.superpowers/sdd/` 一并收口。
- **`.gitignore` 补全**：新增忽略 `.scratch/`、`.agents/`、`.video_agent/`、`.superpowers/`、`docs/superpowers/`、`.opencode/`、`out/`、`AGENTS.md`、`ROADMAP.md`（原有 `build/bin`、`node_modules` 保留）。
- **取消跟踪**：`git rm --cached ROADMAP.md` 与 `.superpowers/`（55 个文件），**磁盘文件全部保留**（本地照旧读写，只是不再随仓库发布）。
- **历史重写**（`python -m git_filter_repo --invert-paths --force`）：清除 `.superpowers/`、`docs/superpowers/`、`ROADMAP.md` 三条路径的全历史痕迹。重写前先在备份镜像上演练确认，再在真仓库执行。**提交数 119 → 118**，6 个标签全部重建指向，代码与文档完好（89 个已跟踪文件、`internal/` 34、`frontend/` 17、`docs/wiki/` 13）。
- **凭证扫描（重写前）**：遍历全部 commit 扫 `sk-`/`fk-`/`oc_sk-`/`sk-cp-`/`sk-amr-`/`user_`/`Bearer`/JWT 等特征串，**命中 0**——历史中从未提交过凭证；同目录的 `decrypt_cookies.py`、`extracted_cookies.json` 等也从未入过 git。
- **安全网**：重写前做了三重备份（`git bundle --all` 0.62 MB、`--mirror` 完整克隆、内部文件与 ROADMAP/AGENTS 副本），存于 `C:\Users\joe\quota-viewer-git-backup-20261014\`。
- **副作用修补**：公开的 `docs/STATUS.md` 有 20 处引用 `ROADMAP.md`（该文件已不在仓库），在阅读指南处加注说明「该文件仅本地存在」。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| `.gitignore` 保留公开跟踪 | clone 者需要忽略规则，否则本地会看到一堆噪音；它是机制而非内部笔记 | 连 `.gitignore` 也撤（用户候选之一，未选） |
| `ROADMAP.md` 撤出公开仓库 | 滚动计划属内部思路，对公开工具仓意义不大；用户选定 | 保留公开（原状） |
| 清历史而非仅停止跟踪 | 用户明确选「连历史一起清除」，`.superpowers/sdd/task-8-9-fix-report.md` 确曾推到 GitHub | 只停止跟踪（更安全，但旧提交仍可翻到） |
| 重写前做三重备份 + 备份上演练 | 历史重写不可逆、影响 3 个 fork 与全部 tag | 直接改（不可接受） |
| 未把 `docs/STATUS.md` 一并撤下 | 用户未选该项；它是项目历史与交接面，且 `docs/wiki/` 与之互引 | 一并撤下（用户候选之一，未选） |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|---|---|---|
| 修改 | `.gitignore` | 补 9 条内部路径忽略规则 |
| 修改 | `docs/STATUS.md` | 头部/TL;DR 更新 + ROADMAP 缺席加注 + 本会话记录 |
| 取消跟踪（保留磁盘） | `ROADMAP.md` | 仍在本地，24 行 |
| 取消跟踪（保留磁盘） | `.superpowers/`（55 文件） | Claude Code superpowers 插件产物，2026-07-17 开发期留下 |
| 未跟踪（新增 ignore） | `.scratch/`、`.agents/`、`.video_agent/`、`AGENTS.md`、`.opencode/`、`out/` | 全部仅本地 |

### 未完成 & 下一步

1. **3 个 fork 已失联**（`mr-dadong`、`itsharex`、`szxufan`）：历史重写后它们的 base 不存在，需 fork 主自行同步或删除重建。已接受此代价。
2. **旧对象可能在 GitHub 侧缓存一段时间**：强推后旧 commit 不再可达，但 GitHub 的 GC 是异步的；如需立即彻底清除可联系 GitHub Support。
3. 非阻断清理批次 12 条、门面刷新遗留 3 条小项仍待办（见本地 `ROADMAP.md`）。
4. STATUS 已超归档阈值（695 行），本轮未归档。

### 已知问题 & 注意事项

- **历史重写的长期影响**：任何在 2026-10-14 17:15 之前 clone 过本仓的副本，其本地历史与远端已分叉；协作者需 `git fetch && git reset --hard origin/master` 或重新 clone。
- 备份目录 `C:\Users\joe\quota-viewer-git-backup-20261014\` 含完整旧历史与内部文件副本，确认无误后可自行删除。
- `.superpowers/` 的来源：本机 `~/.claude/plugins/marketplaces/superpowers-marketplace` 装的 Claude Code 插件，在 2026-07-17 开发首日写入 `.superpowers/sdd/`（含 20 个 Python 探针脚本、8 个 diff、25 个 md 报告）。该目录与 `docs/superpowers/`（早期 spec/plan）现已全部排除在版本控制外。

### 关键上下文

- **重写工具**：`python -m git_filter_repo`（模块在 `C:\Users\joe\AppData\Roaming\Python\Python313\site-packages`，`git filter-repo` 命令本身不在 PATH）；它默认会移除 `origin` 远端，需事后 `git remote add` 接回——本次已接回 `git@github.com:eeljoe/quota-viewer.git`。
- **重写后 hash 全部作废**：此前 STATUS/wiki 里记录的任何 commit hash（如 `d6c09a4`、`b5537c2`）在本地与远端都已不存在，仅存在于备份 bundle 中。
- Wiki 指针状态：`docs/wiki/` 13 文件（00–10 + 99 + `.covered-files`），本轮零代码改动，wiki 无漂移。

---

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
