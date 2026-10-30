# 📋 项目进度文档

> **⚠️ 新会话阅读指南（必读）**
>
> 本文档较长，**不要通读全文**。按以下顺序阅读：
>
> 1. 阅读下方「上下文摘要（TL;DR）」快速了解项目状态
> 2. 阅读下方「项目概况」了解项目基本信息
> 3. 读「会话索引」拿全历史地图（一行一次会话）；索引没点名的内容一律不读
> 4. 阅读项目根目录 ROADMAP.md 的「Now」栏（只读 Now；Next/Later 属按需层，任务点名才读）
>    - 注：`ROADMAP.md` 自 2026-10-14 起**仅在本地存在、不入公开仓库**（与 `.scratch/`、`.agents/` 同）；若你拿到的是 clone，此文件缺席属正常，计划内容可临时向维护者索取
> 5. **直接跳到最新一条「会话记录：2026-10-30 21:05」章节**（用标题搜索定位，不要假设在文件末尾）
>    - 重点看「未完成 & 下一步」和「关键上下文」
> 6. 更早内容只经「会话索引」按需回溯（📦 行的章节在 `docs/wiki/99-appendix-legacy-status.md`）；「当前任务进度」「关键文件索引」等常驻节默认不读
> 7. 给用户一份简短进度汇报（每段 3-5 行，大白话、可读性优先）：**当前进度**（现在在干什么 → 干到哪了 → 下一步是什么 → 有没有卡点）+ **上个会话做了什么**（最新会话章节的摘要、完成与遗留）
> 8. **严禁在用户给出指示之前修改任何文件或代码**
> 9. 汇报后等待用户指示，不要主动开始执行任务
>
> 📍 最新章节位置：`## 会话记录：2026-10-30 21:05`（搜索定位，可能不在文件末尾）
> 🔖 对应 commit：`aca413f` on `master`（tag `v1.5.0` 已发布并标 Latest）
> 📊 累计会话数：主文档 3 条，另有 25 条已归档（docs/wiki/99-appendix-legacy-status.md）

---

## 最后更新

<!-- git-meta: {"last_commit": "aca413f", "branch": "master", "dirty": false, "timestamp": "2026-10-30T21:05:00+08:00"} -->

- **日期**：2026-10-30 21:05
- **会话摘要**：发布 **v1.5.0**（十一家平台，新增 Devin；含 Factory 球灯修复）——本地重建 exe、打 tag 推送、建 Release 并核验上传件 SHA256 与本地一致；两份 README 换新截图并订正体积与 Factory 展示描述。

---

## 上下文摘要（TL;DR）

- 项目：Quota Viewer，桌面悬浮球 + AI 平台额度监控工具，Go + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite）
- 当前阶段：**v1.5.0 已发布并标 Latest**（`aca413f`，tag `v1.5.0`，11 家平台含 Devin，含 Factory 球灯修复）；桌面实例已重启跑新产物。唯一未做的核对仍是 Devin 界面目视（球格 `DV` 与 Devin Desktop Usage 页逐项比对）
- 十一家 Provider 除 DeepSeek（余额型）外均为用量型；状态灯契约 60% 黄 / 80% 红 / 100% 熄灭；usage 型一律「最紧张窗口」驱动 Percent；**一个监控对象 = 一组凭证 × 一张套餐卡**，悬浮球上限按对象数计（普通 9 / 扩展 16，超出显示 `+N`；注意对象数超上限时 `clampEnabled` 会把排在后面的 Provider 静默禁用）
- Open Design 只能靠浏览器会话 Cookie 读额度（API key `sk-amr-` 一律 401），读法已写进 README 与 `docs/ADDING_A_PROVIDER.md`
- 下一步：目视确认 Devin 后本轮迭代即闭合；「对端告知」报文待发（必须带上 probe `schema` 1→2 的契约变更）；非阻断清理批次待你拍板（原有 12 条 + Devin 组 6 条，见本地 `ROADMAP.md`「Now」/「Later」）
- 注意事项：**`wails build` 的产物字节与构建目录绑定**（未加 `-trimpath` 时 Go 把源码绝对路径编进二进制）⇒ 核验交付 exe 必须在同目录重建；**公开仓库自 2026-10-14 起只含产品代码与对外文档**——`.scratch/`、`.agents/`、`.video_agent/`、`.superpowers/`、`docs/superpowers/`、`AGENTS.md`、`ROADMAP.md` 全部已进 `.gitignore`，提交仍应精确 `git add` 而非 `-A`（ignore 只管未跟踪文件，已跟踪文件不受保护）



---

## 项目概况

- **项目名称**：Quota Viewer
- **技术栈**：Go 1.24 + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite 打包）
- **项目根目录**：`C:/Users/joe/Desktop/工作学习/软件开发/quota viewer`
- **平台**：Windows 10+（WebView2 运行时）
- **功能**：悬浮球（1-9 格动态，颜色=状态，方圆角）+ 展开面板（进度条 + 剩余量明细）+ 配置面板（勾选 Provider + 各平台凭证 + 扩展模式开关）+ 系统托盘 + 关闭到托盘
- **支持 Provider**：Kimi（API Key）、讯飞星辰（Cookie）、OpenCode Go（API Key，留空读 opencode CLI 凭证）、小米 MiMo（Cookie）、DeepSeek（API Key，余额型 + 预算进度条）、Ollama（Cookie，Cloud 5 小时/周用量）、Command Code（API Key，5 小时/周窗口 + 剩余 credits）、Factory Droid（API Key，双账单模型 + Core 池）、MiniMax（Token Plan 订阅 Key，5 小时/周窗口）、Open Design（浏览器会话 Cookie，Design Plan 5 小时/7 天窗口）、Devin（手填 token 或留空自动读本机 Devin Desktop 登录态，日/周窗口）
- **同屏上限**：注册表 11 家，悬浮球同屏最多 9 格（普通模式 3 格 / 扩展模式 9 格，`config.go:35-36`）；对象数超上限时 `clampEnabled` 按注册表顺序禁用靠后的 Provider（第 10、11 家是 `open-design` 与 `devin`，凭证保留）

## 当前分支与最近提交

- **分支**：master
- **HEAD**：`81d4224`
- **最近提交**（Devin 任务组，按时间倒序）：
  - `81d4224` - docs: wiki 同步第 11 家 Provider Devin（十家 → 十一家）
  - `dc127dc` - docs: 订正 devin.go 字段注释的 server 兜底语义
  - `16fd6f0` - docs: 订正 Devin 400 错误码的文档口径
  - `4adfe6c` - docs: 订正 Devin 多命中形态的退化说明
  - `c031ca5` - docs: 补 Devin Provider 文档与平台清单（第 11 家）
  - `24a006d` - test: 补 Devin 测试的 APPDATA 隔离并订正过宽注释
  - `acbf6d1` - test: 修正 Devin 手填 token 优先用例，使其真正走本机读取
  - `9848208` - feat: Devin 凭证留空自动读本机 Desktop 登录态
  - `ce19008` - test: 补 Devin 三条口径断言（并列取日 / null 容错 / 单侧缺百分比）
  - `69a0f3e` - feat: 新增 Devin Provider 手填 token 全链路
- **Release**：v1.4.1（2026-10-25，Latest，附 exe）← v1.4.0（2026-10-19）← v1.3.0 ← v1.2.0 ← v1.1.2 ← v1.1.1 ← v1.1.0 ← v1.0.0

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

- **Devin Provider（第 11 家）**（10/29，`69a0f3e`…`dc127dc`）：`devin.go` 走 Devin Desktop 的 `SeatManagementService/GetUserStatus`（CONNECT POST，`metadata.api_key`），日/周**剩余**百分比反转成已用、单球由最紧张窗口驱动；凭证手填优先、留空自动读 `%APPDATA%\Devin\User\globalStorage\state.vscdb`（按字节扫明文 token 与 `apiServerUrl`，不引 SQLite 驱动）；server 三级兜底（注入 → 本机 → 默认）；注册表第 11 家 + `config.AllProviderIDs`；文档与 wiki 同步、两份 README 改 11 家；真机直连取证通过（`percent=0`、`remaining=100`、`reset_at` 非空）

### 🔄 进行中
- 无

### 📋 待办
- **Devin 界面目视（AS-11b，唯一未验项）**：桌面实例已在跑（devin 已启用、凭证留空），看球格是否出现 `DV`、面板日/周数字与 Devin Desktop 的 Usage 页一致
- **发版 v1.5.0**（目视通过后）：11 家平台，`Remaining` 按「超额金额」口径
- 「对端告知」报文待发给 subagent-fallback（必须带上 probe `schema` 1→2）
- 非阻断清理批次待用户拍板：原有 12 条 + 本轮新增 6 条（见 ROADMAP「Now」/「Later」）
- 计划级事项见 `ROADMAP.md`

---

## 未完成 & 下一步

1. **Devin 界面目视（AS-11b）**：你本人在桌面上看一眼即可；未验前 Devin 不算交付完成（门与三张票都如实留了这条）
2. **发版 v1.5.0**：目视通过后 `wails build`（构建前先杀实例）→ 打 tag → Release 附 exe；release notes 按惯例只提「新增第 11 家平台 Devin」与可见行为，不写内部返工细节
3. **「对端告知」报文**：回 `subagent-fallback` 说明探针就绪并给出二进制名/开关名/构建命令，**必须写明 `schema` 已从 1 升到 2**（新增 `slot` / `sub_label` 两个键，纯增量），否则对端按旧契约解析会对不上
4. 非阻断清理批次（原有 12 条 + 本轮 6 条）待你拍板；改代码的那几条会作废对应票的 VERIFIED、需重验
5. Wails 版本升级对齐（go.mod v2.12.0 vs CLI v2.13.0，构建时仍打印该告警）
6. 前端渲染逻辑的可视化验证缺口仍在（见 2026-10-19 会话「关键上下文」）

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
- **Devin 实现侧的两个已知取舍（2026-10-29）**：① 本机凭据扫描**按字节匹配、每种字面量只取第一个命中点**（`bytes.Index`，无循环）⇒ 加密存储 / 字面量改名 / 挪了 `state.vscdb` 位置 / 首命中是裸前缀时会静默退化为「未配置 Devin token」，而**多命中**不退化（token 非空，可能拿到旧 token 对应额度或 403）；② 生产构造点 `registry.go` 只传 `api_key`、`serverURL` 恒空 ⇒ **每次 Fetch 都会整份读 `state.vscdb`（约 1.5 MB，毫秒级）**，手填 token 只省「取令牌」不省读文件
- **配置层钳制（11 家之后更容易踩）**：对象数超上限时 `clampEnabled`（`config.go:103-116`）按注册表顺序**静默禁用**靠后的 Provider（普通模式 9 格 ⇒ 第 10、11 家 `open-design` 与 `devin` 同时被关，凭证保留）；前端 `+N` 只表示同家多套餐卡超出，不表示 Provider 被禁用

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

- **位置**：`docs/wiki/`（已初始化 2026-08-04；最近同步 **2026-10-29**，锚点 `dc127dc`）
- **文件数**：13 个（00-agent-rules ~ 10-build-test-baseline + 99 归档页 + `.covered-files` 缓存）
- **入口**：先读 `docs/wiki/00-agent-rules.md`（含索引与 wiki-meta 同步状态）
- **缓存**：`docs/wiki/.covered-files`（61 项；2026-10-29 新增 `internal/fetcher/devin.go` / `devin_test.go`）
- **最近一次同步**：2026-10-29 第 11 家 Provider Devin 落地，同步 00 / 01 / 02 / 03 / 05 / 07 / 08 / 09 八页（平台表 + Devin 五个小节 + 模块行数 + 覆盖文件）
- **99 归档页**：23 条会话（最近一次归档 2026-10-29，Factory 展示口径组整组搬入；Devin 组仍在主文档）

---

## 关键文件索引

- `app.go` - Wails 主应用（窗口定位 fitToScreen、监控对象编排 `fetchAll`、配置）
- `main.go` - 入口 + Wails 运行时配置（frameless / AlwaysOnTop）
- `workarea_windows.go` / `workarea_other.go` - Win32 工作区查询辅助（非 Windows 桩）
- `internal/fetcher/registry.go` - **Provider 注册表（新增 Provider 的唯一入口）**
- `internal/fetcher/types.go` - Fetcher / MultiFetcher 接口、QuotaResult（含 Slot / SubLabel）、Kind 常量
- `internal/fetcher/abbr.go` - `SettleIdentity`（缩写/后缀/slot 归一）+ 缩写派生（app.go 与 probe 共用）
- `internal/fetcher/{kimi,xfyun,opencode_go,mimo,deepseek,ollama,commandcode,factorydroid,minimax,opendesign,devin}.go` + `budget.go` - 十一家平台抓取器（均有测试）+ 余额型预算换算
- `internal/config/config.go` + `cookie.go` - 动态 Provider 配置 + 多凭证组（`CredGroup`）+ 旧格式迁移 + PowerShell Cookie 解析
- `internal/tray/tray.go` - 系统托盘（systray.Run + LockOSThread）
- `frontend/src/index.html` / `main.js` / `style.css` - 动态球格（含 `+N` 溢出格）+ 多组凭证配置面板
- `frontend/wailsjs/` - Wails v2 前端绑定（构建生成）
- `.scratch/multi-key/probe/{ball-render,config-render}-probe.js` - 前端渲染探针（**不入库**；改前端 DOM 后必跑）
- `docs/wiki/05-fetching-platforms.md` - 新增 Provider 指南详情
- `docs/ADDING_A_PROVIDER.md` - **Agent 专用**新增 Provider 完整指南（fetcher 实现 + 注册 + 测试 + 检查清单）

---

## 会话索引

> 一行一次会话；索引没点名的内容一律不读。`位置` 为 📦 的章节已移入 `docs/wiki/99-appendix-legacy-status.md`。

| # | 时间 | 组 | 摘要 | 位置 |
|---|---|---|---|---|
| 1 | 08-25 23:11 | Command Code / 重启验证 | 重启验证通过：OpenCode Go 浮点修复生效、Command Code 剩余文本可读，提交推送收尾 | 本文 |
| 2 | 10-25 18:45 | Factory 窗口展示口径 | 修 Factory 最紧张月额度看不到 + 倒计时不对，重建 exe 并截图确认 | 📦 |
| 3 | 10-25 19:18 | Factory 窗口展示口径 | 发布 v1.4.1 并按任务组归档旧会话 | 📦 |
| 4 | 10-28 23:21 | Devin Provider | 规划第 11 家：查清私有额度通道、实测打通，出 spec + 3 张票并三轴审票 | 📦 |
| 5 | 10-29 02:45 | Devin Provider | 三张票逐票实现与验收（含三轮返工），整体门 VERIFIED；wiki 与文档同步 | 📦 |
| 6 | 10-30 17:45 | Factory Droid 球色口径 | 修 F 灯在 Standard 刷新后仍灰：Core 不再并入球色，真机与截图确认转绿 | 本文 |
| 7 | 10-30 21:05 | v1.5.0 发版 | 发 v1.5.0（十一家含 Devin，含球灯修复）+ README 换新截图并订正体积 | 本文 |

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

> 📦 任务组「Factory Droid 窗口展示口径」（2026-10-25 18:45 / 19:18 两条）已归档：docs/wiki/99-appendix-legacy-status.md

---

> 📦 任务组「Devin Provider」（2026-10-28 23:21 / 2026-10-29 02:45 两条）已归档：docs/wiki/99-appendix-legacy-status.md

---

## 会话记录：2026-10-30 17:45

> **会话摘要**：修 Factory Droid 的球灯——Standard 额度刷新后 F 灯仍灰，根因是非接管路径把 Core 池并进了 `Percent`（Core 未恢复就一直判成耗尽）；改为 Standard 未封顶时 Core 完全不参与，真机复测与球面截图确认转绿
> **Git**：`0ed9651` on `master`（本组 4 个提交，未发版）
> **任务组**：Factory Droid 球色口径（单会话轻量路线，无 spec 拆票）

### 本次完成

- **根因定位**：`parseTokenRateLimits` 的非接管路径有一句 `if coreMax > result.Percent { result.Percent = coreMax }`（旧 `factorydroid.go:222-224`），把 Core 池并进球色。真机现场 standard 5h/周/月 = 0/24/0、core = 0/75/**100**（11-11 才恢复）⇒ `Percent` 被顶成 100 ⇒ 前端 `percent >= 100` 判熄灭灰。同一 Core 段还排到 `Remaining` 首位，把倒计时带去 11-11 的 Core 月窗（探针 `reset_at` 实测就是这个值）。
- **口径依据**：Factory 官方计费是「Standard 先烧，三个滚动窗口都有余量才能发请求；Droid Core 是 Standard 跑完之后才生效的第二管血，有独立窗口」⇒ Standard 未封顶时 Core 的读数不构成任何约束。2026-09-30 的「Core 仓一有消耗即展示」是同一个语义错误的另一面，本次一并作废。
- **实现**（`2e43531`）：非接管路径删掉上面那句与「Core 一有消耗即展示」整块；`Percent` 只由 Standard 三窗口决定。Standard 封顶后的 Core 接管逻辑**一字不动**（2026-10-13「球看着满了、其实 Core 还有量」的语义仍在）。
- **测试**：新增 10-30 真机夹具回归用例 + 两条「Core 健康态必须隐藏」用例；改写两条前提已作废的 Core 用例；AC 表 AC3/AC5/AC6 期望同步到新口径。变异实测：加回 `coreMax` 顶高 → 3 处红；`>= 100` 改 `> 99` → AC5 红；退回「一有消耗即展示」→ 6 处红。
- **独立动态验收 VERIFIED**（`27c331c` 固定点）：AS-1…AS-5 逐条亲手取证、全量 204 用例绿、8 发变异 7 杀、真机 CLI `percent` 与 Standard 一致且 `reset_at` 落在 Standard 窗口。门挑出的唯一 finding 是测试盲区——接管路径 `ResetLabel` 的 `Core ` 前缀没有断言（变异 M6 存活），已用 `63db78c` 补断言并复验该变异被杀，产品代码未动。
- **交付**：`wails build` 重建（18.9s）→ 旧实例（PID 39228）结束、新实例（PID 28996）启动 → 桌面 `.lnk` 确认指向新产物 → 球面截图确认 **F 格已从暗灰变成绿色**。
- **wiki 同步**（`27c331c` + `8695cc5` + `0ed9651`）：05 的 droid 段改掉 `Percent = max(stdMax, coreMax)` 与「Core 一有消耗即显示」，02 两条行数与说明刷新，`synced_commit` 推进到 `63db78c`。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| Standard 健康时 Core **完全不显示** | 用户 2026-10-30 明确选择；面板保持「当前在烧哪管血」的单一叙事，Core 什么时候恢复去 Factory 页面看 | Core 退到 `Remaining` 尾部作信息位（被否） |
| 接管路径（Standard 封顶）一字不动 | 2026-10-13 事故的守护仍在生效，本次现场不涉及该分支 | 顺手一起改（会作废已 VERIFIED 的那块） |
| 作废「Core 一有消耗即展示」 | 它就是同一个语义错误：让 Standard 健康时的展示顺序与倒计时被第二管血带偏 | 保留展示但只放到尾部（与用户选择冲突） |
| 门挑出的 `ResetLabel` 盲区当场补断言 | 机械可修、不动产品代码、不推翻 VERIFIED | 留给清理批次（会在接管场景静默错标窗口名） |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `internal/fetcher/factorydroid.go` | 非接管路径不再并入 Core；注释重写「两管血的分界」（465 行） |
| 修改 | `internal/fetcher/factorydroid_test.go` | 10-30 真机夹具回归 + 两条健康态隐藏用例 + 接管 `ResetLabel` 前缀断言（674 行） |
| 修改 | `docs/wiki/05-fetching-platforms.md` + `02-module-catalog.md` + `00-agent-rules.md` | droid 段口径、模块行数、同步点 |
| 新增 | `.scratch/factory-core-ball-color/` | 不出库：spec、票与裁决、门的变异脚本 |

> 本次变更：`fa1a732..0ed9651` +108/−52 行（另有 2 次 wiki 提交），7 个文件
> 变更基准：`git diff --stat fa1a732..HEAD`

### 未完成 & 下一步

- **发版**：本次修复已进 `build/bin/quota-viewer.exe` 并在桌面实例生效，但**没有打 tag、没有发 Release**——等你决定是单发 v1.4.2 还是并入 v1.5.0
- **Devin 界面目视（AS-11b）**（持续中）：桌面实例刚重启过（PID 28996），球格 `DV` 现在就能看，面板日/周百分比与 Devin Desktop 的 Usage 页对比
- **「对端告知」报文**（持续中）：必须带 probe `schema` 1→2
- **非阻断清理批次**（持续中）：原有 12 条 + Devin 组 6 条，见 `ROADMAP.md`「Now」

### 已知问题 & 注意事项

- **Factory 两管血的展示边界**（本次定档）：`limits.core` 的三窗口**只在 Standard 顶格时**参与任何展示；Standard 没顶满时 Core 怎么满都不影响球色与文字。若日后上游改了「Core 先烧」的顺序，这条要重审。
- `wails build` 后 `frontend/` 三个生成物会变脏（CRLF），本次已 `git checkout -- frontend/` 还原，未提交
- 无 git stash；仓库仍是 `master` 单线，无额外 worktree

### 推荐 Skill

- `/wiki-update` - 本次已手动同步（Factory 一条事实）；后续大改 Factory 后重跑
- `/update-status` - 本会话记账（已执行）

### 关键上下文

- **Factory 检测逻辑的分叉点**（新会话要改这块先看这里）：`parseTokenRateLimits` 里 `takesOver = stdMax >= 100 && coreHasData && (overagePreference ∈ {"", "droidCore"})`。`takesOver` 为真 ⇒ 主体切 Core（Percent/Used/ResetAt/Remaining 全跟 Core 最紧张窗口，Core 段钉首位）；为假 ⇒ **只看 Standard**，Core 一个字段都不读。
- **2026-10-30 真机数据**（`realFactoryLimitsJSON20261030` 夹具逐字存档）：standard 0/24/0、core 0/75/100、`overagePreference=droidCore`、`extraUsageBalanceCents=0`。旧代码在此数据下 `Percent=100`（错）、新代码 `Percent=24`（对，球绿）。
- **验收证据链**：`.scratch/factory-core-ball-color/spec.md`（AS-1…AS-5）→ `issues/01-*.md` 的 `## Verdicts`（r1 VERIFIED 27c331c）→ `verifier-probes/mutate.mjs`（门的 8 发变异脚本）。
- **探针契约没动**：仍是 schema 2、每记录十个键。
- Wiki 指针状态：`docs/wiki/` 12 个常规页 + 99 归档页（23 条），`synced_commit` = `63db78c`。

---

## 会话记录：2026-10-30 21:05

> **会话摘要**：发布 **v1.5.0**（十一家平台，新增 Devin；含 Factory 球灯修复）——本地重建 exe、打 tag 推送、建 Release 并核验上传件字节一致；README 两份换新截图并同步体积口径
> **Git**：`aca413f` on `master`（tag `v1.5.0`，已推送）
> **任务组**：v1.5.0 发版（承 2026-10-30 17:45 的 Factory 球色口径会话）

### 本次完成

- **发版 v1.5.0**（用户拍板：不单发 v1.4.2，直接并入含 Devin 的 v1.5.0）：
  - 在 `aca413f` 本地 `wails build` 重建交付 exe（28.5s，11,684,352 字节），构建后 `git checkout -- frontend/` 还原 CRLF 脏文件
  - `git push origin master`（`fa1a732..aca413f`，7 个提交）→ 打 `v1.5.0` 附注 tag 并推送 → `gh release create v1.5.0` 附 exe，已标 Latest：https://github.com/eeljoe/quota-viewer/releases/tag/v1.5.0
  - **交付件核验**：`gh release download` 拉回后与本地 `Get-FileHash -Algorithm SHA256` 比对，`D9BF14C4…C715` 两边一致、字节数相同
  - release notes 按既有偏好写：只提新增第 11 家 Devin 与 Factory 球灯的可见行为，末尾一句「修复了一些简单的小问题」，不写内部返工细节
- **README 两份换新截图**（`aca413f`）：用户提供两张实机图（悬浮球 124×118、展开面板 640×1218）替换旧图（原来只覆盖 6 家，现在是含 Devin 在内的 8 个监控对象 / 11 家平台）；图片下方补一句图注说明格色与倒计时含义，中英同步。
- **顺带订正三处失实文案**：体积从「~10.9 MB / ~60 MB」改成实测「~11.1 MB / ~66 MB」（`WorkingSet64` 实测 65.6 MB）；Factory 一行从「+ Core 池用量」改成「Standard 用满后改由 Core 池接管展示」（与本次口径修正一致）。
- **桌面实例**：旧实例结束后重启（PID 20540），`.lnk` 确认指向新产物（`matches fresh build: True`）；探针 `--json` 复核 8 个对象全绿、factory-droid `percent=2`（Standard 5h 窗口）、`reset_at` 落 Standard 窗口。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| 不单发 v1.4.2，直接发 v1.5.0 | 用户拍板「1.5.0, 推」；Devin 已过整体门只差界面目视，两笔改动合一次 release 省一次往返 | 先发 v1.4.2 单含 Factory 修复，Devin 留到下次 |
| 界面目视不作为发版前置 | 用户已看面板截图并直接要求发版（截图本身即目视证据：`DV` 格在、日/周 0% 合理） | 等用户再看一眼 Devin Desktop 的 Usage 页再发 |
| 顺手改体积文案 | 旧数字已与实测差 0.2 MB / 6 MB，release 前是唯一低成本窗口 | 留到清理批次（会让 README 继续说错数字） |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `README.md` + `README.zh-CN.md` | 新截图引用 + 图注 + 体积口径 + Factory 展示描述 |
| 修改 | `docs/screenshots/preview-1.png`（124×118）+ `preview-2.png`（640×1218） | 用户提供的两张实机截图，覆盖旧图 |
| 新增 | `.scratch/factory-core-ball-color/release-notes.md` | 不出库：v1.5.0 release notes 原文 |
| — | `build/bin/quota-viewer.exe` | 交付件（不入库），11,684,352 字节，SHA256 `D9BF14C4…C715` |

> 本次变更：`0ed9651..aca413f` +83/−13 行，5 个文件（其中 2 个是二进制截图）
> 变更基准：`git diff --shortstat 0ed9651..HEAD`

### 未完成 & 下一步

- **Devin 界面目视（AS-11b）**（持续中）：已发版但仍未与 Devin Desktop 的 Usage 页逐项对过；球格 `DV` 与面板「日 0% 已用 · 周 0% 已用」看起来正常，若发现数字不符再重开票
- **「对端告知」报文**（持续中）：必须带 probe `schema` 1→2
- **非阻断清理批次**（持续中）：原有 12 条 + Devin 组 6 条，见 `ROADMAP.md`「Now」
- Wails 版本升级对齐（go.mod v2.12.0 vs CLI v2.13.0）与前端渲染探针纳管仍挂在 ROADMAP

### 已知问题 & 注意事项

- **`gh release view --json isLatest` 在本机会报「Unknown JSON field」**，看 Latest 标记要用 `gh release list`；交付件校验用 `gh release download -p <asset> -D <dir>` 后本地算 SHA256 比对
- release notes 里「十一家平台」与 README 平台表必须同时改，否则门面又出现「写十家但注册表十一家」的旧毛病
- 无 git stash；仓库仍是 `master` 单线，无额外 worktree

### 推荐 Skill

- `/read-status` - 新会话接入（本次已收尾，可直接用）
- `/update-status` - 本会话记账（已执行）

### 关键上下文

- **当前线上版本**：**v1.5.0 是 Latest**（2026-10-30 发布，https://github.com/eeljoe/quota-viewer/releases/tag/v1.5.0 ，tag 指向 `aca413f`），覆盖的可见变化 = 新增第 11 家 Devin + Factory 球灯只跟 Standard 走 + 两份 README 新截图。
- **发布流程定式**（下次照抄）：`aca413f` 这类文档提交先推 → 本地 `wails build`（先杀实例）→ `git checkout -- frontend/` → `git tag -a vX.Y.Z` → `git push origin vX.Y.Z` → `gh release create vX.Y.Z build/bin/quota-viewer.exe --title "..." --notes-file ...` → `gh release download` 回拉比 SHA256 → 重启实例并核 `.lnk`。
- **截图资产**：`docs/screenshots/preview-1.png`（悬浮球）、`preview-2.png`（展开面板），两份 README 共用；换图时两张 README 的引用路径不用改，但要同步图注与「覆盖几家平台」的措辞。
- 探针契约没动：仍是 schema 2、每记录十个键。
- Wiki 指针状态：`docs/wiki/` 12 个常规页 + 99 归档页（23 条），`synced_commit` = `63db78c`；本次只改 README 与截图，均不在 wiki「关键文件」表覆盖面内 → 无漂移。
