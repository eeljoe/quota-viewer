# 旧 STATUS 会话档案（Legacy）

> 本页为旧 STATUS 会话档案，新会话无需阅读，仅审计/回溯用。会话记录由 status 系列 skill 的归档机制自动搬移至此，原样保留。

**When to read**: 审计历史决策、回溯早期实现细节时。

---

## 会话记录：2026-08-04 05:31

> **会话摘要**：初始化项目进度文档体系（/read-status 未找到文档 → /save-status 创建 STATUS.md + /wiki-init）
> **Git**：`e60f844` on `master`（dirty）

### 本次完成
- 扫描项目，确认无既有 STATUS.md / progress.md；收集 git 状态（dirty 15 项、无 stash、单 worktree）
- 识别工作区进行中任务：MiMo → OpenCode Go 抓取器替换（含 fitToScreen 窗口定位修复）；`go test ./...` 全绿
- 创建 STATUS.md（docs/ 下，完整模板）
- 初始化 docs/wiki/：11 个专题文件 + .covered-files 缓存

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| STATUS.md 放 docs/ | docs/ 已存在，与搜索范围一致 | 项目根目录 |
| 用完整模板 | 源文件 > 10 个，属大中型项目 | 精简模板 |

### 未完成 & 下一步
- 提交工作区改动；OpenCode Go 真实抓取冒烟（后经用户确认工作正常）

---

## 会话记录：2026-08-04 14:10

> **会话摘要**：通用化升级——动态 Provider 配置（1-3 个）、五平台注册表、恢复 MiMo、新增 DeepSeek、动态球格、开源准备
> **Git**：`e60f844` on `master`（dirty，大量未提交）

### 本次完成
- 确认 OpenCode Go 实际工作正常（修正会话 1 的"未冒烟验证"误记）
- fetcher 注册表 registry.go（ProviderDef/CredentialField/Build）+ QuotaResult 扩展（ID/Abbr/Kind，usage/balance 两种类型）
- 从 git 历史恢复 mimo.go + mimo_test.go；新增 deepseek.go + 测试（余额端点已联网核实）
- config v2：动态 Providers 列表，Load 时旧扁平格式自动迁移（含 mimo_cookie 保留、4 平台钳制）并回写
- app.go：SaveConfig([]ProviderInput) / GetConfig(全量元数据) / fetchAll 动态并发 / TestConnection 走注册表
- 前端：球格按结果数动态重建（1 格放大占满 / 2 格各半 / 3 格各 1/3）、配置面板按元数据动态生成、勾选 1-3 限制、事件委托
- 验证：go test 全绿；npm build / wails build 成功；exe 启动冒烟通过；git 历史凭证扫描 0 命中
- README 重写 + MIT LICENSE；wiki 8 个文件同步五平台模型

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 注册表 + 凭证字段元数据驱动前端 | 新增 Provider 零前端改动（README 已写扩展指南） | 前端硬编码五平台 |
| 配置固定 5 条 Providers + enabled 标志 | 结构稳定、迁移简单、顺序固定为注册表序 | 动态切片（顺序可自定义，复杂度高） |
| DeepSeek 用 Kind=balance 余额型 | 余额无"用量百分比"语义，前端恒绿 | 复用 percent（语义错误） |
| LICENSE 用 MIT | 最宽松、社区默认 | Apache-2.0 |
| 配置迁移在 Load 内自动回写 | 用户无感升级 | 手动迁移工具 |

### 未完成 & 下一步
- 用户冒烟（球格布局 / 勾选限制 / DeepSeek 余额）
- 提交全部改动（拆 commit 序列见上）；开源发布（remote/GitHub）

### 已知问题 & 注意事项
- mimo.go 状态冲突：会话前 staged 删除 + 本次恢复，提交前需 `git restore --staged`
- Wails CLI 2.13.0 vs go.mod 2.12.0 版本警告（可选升级）
- 展示顺序固定为注册表顺序（用户未要求自定义顺序）

### 关键上下文
- 新增 Provider 的完整路径：`internal/fetcher/` 新抓取器（Fetcher 接口 + baseURL 注入 + 测试）→ `registry.go` 注册（id/显示名/缩写/字段/Build）→ 前端自动适配
- Provider id 契约：kimi / xfyun / opencode-go / mimo / deepseek（config 存储、TestConnection、前端绑定共用）

---

## 会话记录：2026-08-04 15:00

> **会话摘要**：v1.0.0 开源发布——双语 README、DeepSeek 多币种修复、干净构建、GitHub Release
> **Git**：`85e03e9` on `master`（clean）

### 本次完成
- DeepSeek 多币种修复：用户截图显示 USD $0.00 + CNY ¥247.51，原代码取 `balance_infos[0]` 会错误显示 $0.00；改为自动遍历取首个非零余额币种（CNY→¥ / USD→$），新增多币种测试用例（上会话遗留）
- 用户确认升级效果"很完美"，7 个 commit 拆分提交（c28aeb7 ~ 99527d9）并推送到 master
- GitHub 仓库创建：`eeljoe/quota-viewer`（公开，MIT，gh CLI 创建 + push）
- 双语 README：英文 `README.md`（默认）+ 中文 `README.zh-CN.md`，顶部语言切换链接
- 干净构建 exe：清理旧产物后重新 `wails build`，二进制扫描 0 凭证命中
- 凭证安全全量排查：git 历史 0 命中、exe 二进制 0 命中、config.json 不在仓库内（`%APPDATA%/quota-viewer/`）、远端文件树扫描无敏感文件
- GitHub Release v1.0.0：tag v1.0.0 + exe 附件（10.78 MB）+ 中英双语 Release Notes

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| DeepSeek 取首个非零余额币种 | 用户真实场景 USD 0.00 + CNY 247.51，取 [0] 会显示 $0.00 | 取最后一个 / 显示全部 |
| 英文 README 为默认 | 开源项目国际化默认英文 | 中文默认 |
| Release v1.0.0 | 首个开源稳定版本，功能完整 | 0.x（不必要，已验证） |
| gh CLI 创建仓库 | 已认证，一条命令完成 create + push | 手动 GitHub Web 创建 |

### 新增/变更文件
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `internal/fetcher/deepseek.go` | 多币种自动选非零余额 + currencySymbol 函数 |
| 修改 | `internal/fetcher/deepseek_test.go` | 新增多币种测试用例 |
| 修改 | `README.md` | 重写为英文默认版（原中文移至 zh-CN） |
| 新增 | `README.zh-CN.md` | 简体中文版 README |
| 修改 | `docs/wiki/05-fetching-platforms.md` | DeepSeek 多币种描述同步 |
| 修改 | `docs/STATUS.md` | 本文件 |

> 本次变更（从 99527d9 到 85e03e9）：+200/-50 行，6 个文件

### 未完成 & 下一步
- 无明确待办——项目已交付开源
- 可选方向：Release 推广 / 新 Provider 扩展 / Wails 版本升级

### 关键上下文
- **GitHub 仓库**：https://github.com/eeljoe/quota-viewer（公开）
- **Release v1.0.0**：https://github.com/eeljoe/quota-viewer/releases/tag/v1.0.0（含 exe 下载）
- **凭证安全确认**：全量排查 git 历史 + exe 二进制 + 远端文件树，0 泄漏；用户真实配置在 `%APPDATA%/quota-viewer/config.json`（仓库外）
- **DeepSeek 余额型**：`Kind="balance"`，前端恒绿；`balance_infos` 数组取首个非零余额币种
- **新增 Provider 路径**：fetcher 实现 → registry.go 注册 → 前端自动适配（README 有英文指南）
- Wiki 指针状态：`docs/wiki/` 11 个文件，`.covered-files` 46 项，synced_commit `85e03e9`

---

## 会话记录：2026-09-19 01:16

> **会话摘要**：修复悬浮球状态灯两处问题（周窗口耗尽不告警、阈值 75/90 过晚）并发布 Release v1.1.1
> **Git**：`7d4adcc` on `master`（已推送；本会话 STATUS/ROADMAP 更新待提交）

### 本次完成
- **状态灯阈值调整**：75% 黄 / 90% 红 → **60% 黄 / 80% 红 / 100% 熄灭**（`main.js` getStatusColor 统一用量型与余额型预算条；`style.css` 新增 off 暗灰态：球格/圆点/进度条三处）
- **周窗口耗尽不告警修复**：`ollama.go` / `commandcode.go` 的 `Percent` 改为取 5h 与周窗口中更紧张的一个（原仅 5h 窗口驱动球色，周 100% + 5h 0% 仍闪绿灯；Command Code 同构顺手同修）。`Used`/`ResetAt` 仍记 5h 窗口
- 测试：新增 2 个回归用例（WeeklyExhausted，先红后绿复现症状），4 个既有 ollama 断言更新为 max 语义；`go test ./...` 全绿
- 交付：`wails build` 成功，桌面 `Quota Viewer.lnk` 指向 `build/bin/quota-viewer.exe`（01:10 新构建，dist 反查含新阈值）；杀旧实例（PID 22300）重启新实例，ollama 球格由绿变灰（周 100% 熄灭态）
- 发布：推送 master（`7d4adcc`）+ **Release v1.1.1**（附 exe 11.3MB；按用户要求 release notes 不提 bug 细节，只写"修复了一些简单的小问题"）

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| Percent = max(5h, 周) 在 fetcher 层取 | 周额度耗尽必须告警，进度条主展示更紧张窗口；前端零改动 | QuotaResult 加窗口数组前端取 max（过度设计） |
| 100% 熄灭用暗灰（text-3） | 语义="灯灭了"，与未加载灰一致，用户明确要"100%熄灭" | 红色闪烁 |
| Release notes 只写"修复小问题" | 用户明确不提 bug 细节 | 完整变更日志 |

### 新增/变更文件
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `internal/fetcher/ollama.go` / `commandcode.go` | Percent 取两窗口较大值 |
| 修改 | `internal/fetcher/ollama_test.go` / `commandcode_test.go` | +2 回归用例，4 断言更新 |
| 修改 | `frontend/src/main.js` / `style.css` | 阈值 60/80/100 + off 态 |
| 修改 | `frontend/dist/*` | wails build 重建产物（已随 fix 提交） |
| 新增 | `ROADMAP.md` | 本会话按 update-status 创建（此前缺失） |

> 本次变更：`7d4adcc`（+95/-24 行，10 个文件）

### 未完成 & 下一步
- 无明确待办；可选方向已沉淀到 `ROADMAP.md`（Next: Wails 版本升级对齐；Later: Release 推广 / 新 Provider / 反馈迭代）

### 已知问题 & 注意事项
- wiki 漂移（见下方推荐 Skill），待 `/wiki-update` 修复
- 前端行尾警告 dist/wailsjs LF→CRLF（原有，不影响构建）
- `build/bin/quota-viewer.exe~`（8/25 旧 exe 改名残留，运行中构建的副产物，可删）

### 推荐 Skill
- `/wiki-update` - 检测到 4 个 wiki 覆盖文件自 synced_commit（a59635f）后有代码变更：`internal/fetcher/ollama.go`、`internal/fetcher/commandcode.go`、`frontend/src/main.js`、`frontend/src/style.css`（05 页"主展示=5 小时窗口"描述已过时）

### 关键上下文
- **状态灯契约（新）**：≥60 黄 / ≥80 红 / ≥100 熄灭（off=暗灰）；ollama/commandcode 的 `Percent` = 5h 与周窗口较紧张者，`Used` 仍记 5h 窗口
- **Release v1.1.1**：https://github.com/eeljoe/quota-viewer/releases/tag/v1.1.1（Latest，附 quota-viewer.exe）
- 当前启用 Provider 含 `ollama`（用户在 8/25 后自行启用——本次"周 100% 仍绿灯"症状即来自它；其余勾选组合以 `%APPDATA%/quota-viewer/config.json` 为准）
- Wiki 指针状态：`docs/wiki/` 12 文件，`.covered-files` 49 项，synced_commit `a59635f`（漂移 4 文件待同步）
- 桌面 `Quota Viewer.lnk` → `build/bin/quota-viewer.exe`（即 wails build 产物，无需复制）

---

## 会话记录：2026-09-22 10:03

> **会话摘要**：修复 Kimi 周额度耗尽漏报（Percent 取 5h 与 7 天窗口较紧张者），推送 master 并发布 Release v1.1.2
> **Git**：`639e4f1` on `master`（已推送；v1.1.2 已发布；本会话 STATUS 更新待提交）
> **任务组**：额度告警修复（承接 2026-09-19 状态灯修复会话）
> **任务组状态**：已完成（修复已验证并发布 v1.1.2）

### 本次完成
- **Kimi 周窗口漏报修复**：用户给出两张截图对照——Kimi Code 官方页显示「5 小时用量 Code 0%、7 天用量 Code 100%」，而本应用 Kimi 仍是绿点 + `0 / 100 (5小时)`，即只读 `limits[0]`（5 小时窗口）、完全忽略周窗口（与 9/19 的 ollama/commandcode 同构漏报）
- **线上响应确认**（真实 Key 打 `GET api.kimi.com/coding/v1/usages`）：`usage{limit:100,used:100,resetTime:...}`（周）+ `limits[0].detail`（5h，remaining=100）+ 新增字段 `usages.limit_5h/limit_7d.used_ratio`（本次才发现的窗口比率字段）
- **修复**：`Percent` 取 5h 与 7 天窗口较紧张者（沿用 9/19 契约），`Used/Total/ResetAt` 仍记 5 小时窗口，`Remaining` 追加周用量 → 实测输出 `0 / 100 (5小时) · 周 100% 已用`，球灯按 ≥100% 熄灭为暗灰；新增 `usages.limit_5h/limit_7d` 比率解析作为 `details.limit` 缺失时的兜底
- 测试：3 个新用例（真实 payload 的周耗尽复现 `TestKimiFetcher_WeeklyExhausted_PercentAlerts`；仅比率响应解析；周低于 5h 时不得压低 Percent），先红后绿；`go test ./...` 全绿
- 交付：`wails build`（09:35 构建，CLI 不在 PATH，用 `C:\Users\joe\go\bin\wails.exe`）→ 杀旧实例（PID 31920）重启新实例（PID 16584）
- 发布：推送 master（`639e4f1`）+ **Release v1.1.2**（附 quota-viewer.exe，标题「用量统计小修复」，notes 按既有偏好只写"修复了一些简单的小问题"）

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| Percent = max(5h, 周) 同样适用于 Kimi | 与 9/19 ollama/commandcode 契约一致，周耗尽必须告警 | Kimi 单独用周窗口驱动（破坏一致性） |
| 周用量优先读 `usages.limit_7d.used_ratio`，回退 `usage.used/limit` | 官方新增的比率字段最直接；字符串对象保留兼容 | 只读 `usage` 字符串（旧字段可能下线） |
| `Remaining` 追加「· 周 X% 已用」而非替换 | 5h 的绝对值/总量信息仍有用，与 Ollama 行格式对齐 | 只显示周百分比 |
| 倒计时仍按 5 小时窗口 | 与 ollama 现有行为一致，避免本次扩大改动面 | ResetAt 取较紧张窗口（已列入可选项待用户决定） |

### 新增/变更文件
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `internal/fetcher/kimi.go` | usages/周窗口解析 + Percent 取两窗口较大值 |
| 修改 | `internal/fetcher/kimi_test.go` | +3 用例（周耗尽 / 仅比率 / 周低于 5h） |

> 本次变更：`639e4f1`（+168/-10 行，2 个文件）
> 变更基准：`git diff d6e7572..HEAD --stat`

### 未完成 & 下一步
- 无明确待办；可选方向：① Kimi 倒计时改取"驱动告警的窗口"的重置时间；② 官方页「总使用量 33.68%」对应的总额度字段（`totalQuota`/booster）未纳入监控，需要时可加
- 计划级事项见 `ROADMAP.md`（Next: Wails 版本升级对齐）

### 已知问题 & 注意事项
- **本机截图取证受限**：`PrintWindow` 抓 Wails 窗口只得到部分渲染、`BitBlt`(CAPTUREBLT) 与全屏 `CopyFromScreen` 抓不到悬浮球（WebView2/合成层），结论以 fetcher 线上实测为准
- **wails CLI 不在 PATH**：须用 `C:\Users\joe\go\bin\wails.exe build`
- `gofmt -l` 会列出仓库里几乎所有 Go 文件（全仓 CRLF 行尾），非本次改动引入，勿按此"修格式"
- wiki 漂移（见「推荐 Skill」），待 `/wiki-update` 修复
- `build/bin/quota-viewer.exe~`（8/25 旧 exe 改名残留，可删）

### 推荐 Skill
- `/wiki-update` - 检测到 5 个 wiki 覆盖文件自 synced_commit（a59635f）后有代码变更：`internal/fetcher/kimi.go`、`internal/fetcher/ollama.go`、`internal/fetcher/commandcode.go`、`frontend/src/main.js`、`frontend/src/style.css`（02/05 页的 Kimi 条目与"主展示=5 小时窗口"描述已过时）

### 关键上下文
- **窗口告警契约（完整版）**：`Percent = max(较紧张的窗口)` 对 Kimi / Ollama / Command Code 三家统一；`Used/Total/Remaining/ResetAt` 仍以 5 小时（或主）窗口为准；≥60 黄 / ≥80 红 / ≥100 熄灭
- **Kimi 响应结构**：`usage`（周：limit/used/remaining/resetTime 字符串）+ `limits[0].detail`（5h）+ `usages.limit_5h/limit_7d.used_ratio`（0-1 比率，2026-09 新增）；旧版 `{"data":[{model_name:"all"}]}` 仍兼容
- **Release v1.1.2**：https://github.com/eeljoe/quota-viewer/releases/tag/v1.1.2（Latest，附 quota-viewer.exe）
- 当前启用 Provider：kimi / ollama / command-code（其余在配置里关闭；以 `%APPDATA%/quota-viewer/config.json` 为准）
- Wiki 指针状态：`docs/wiki/` 12 文件，`.covered-files` 49 项，synced_commit `a59635f`（漂移 5 文件待同步）
- 桌面 `Quota Viewer.lnk` → `build/bin/quota-viewer.exe`（即 wails build 产物）


---

<!-- 第二批归档 2026-10-03：2026-08-04 ~ 2026-08-25 共 5 条 -->

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
- **球形方案回滚**（10/3 用户裁决）：region 方案三宗罪——圆缘锯齿明显（region 无抗锯齿）、圆外底色丑、展开面板空白（Wails 窗口尺寸状态被物理像素直调搅乱）。revert e96af01 + 恢复方圆角视觉；扩展模式（4+ 格网格）与面板高度自适应保留
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

---

## 会话记录：2026-10-03 00:15

> **会话摘要**：用户目验后否决球形方案（展开面板全白 / 圆外底色丑 / 圆缘锯齿）——窗口层 revert + 球体视觉恢复方圆角，保留 MiniMax 与扩展模式；全部提交已推送
> **Git**：`537daea` on `master`（全部已推送）
> **任务组**：MiniMax Provider 接入 + 展示形态改版
> **任务组状态**：已完成（球形否决回滚收尾；MiniMax + 扩展模式交付运行中）

### 本次完成
- （上会话遗留）**用户目验悬浮球——已完成，结论：否决**。用户发截图裁决：① 点击展开面板全白 ② 圆形背后底色丑 ③ 圆缘锯齿明显，明确「我的意见是回滚」
- **窗口层回滚**：`git revert 86a76c7` → `e96af01`（SetWindowRgn 裁圆 / 物理像素 SetWindowPos / Expand-Collapse 挂钩全部移除，main.go 回到原始配置）
- **球体视觉恢复**：style.css 回到方圆角（border-radius 16px + hairline 分隔），保留 4+ 格网格与面板高度自适应；main.js 去掉 cells-2-3 字号档 → `537daea`
- **重建交付**：`wails build` → 新实例 PID 23220，截图验证：方圆角球体、K/O/F/MX 四格、无锯齿无杂色、窗口 120x120 物理正常（撑宽未复现——属窗口创建时机偶发，非必现 bug）
- **补交 dist**：回滚后 `frontend/dist` 重建产物入库（新 hash，与交付 exe 一致）
- **留痕**：ROADMAP「已否决」记录球形方案与重试前提；STATUS 10-02 记录补回滚根因 → `5bc855f`；归档 8 月旧会话 5 条到 wiki 99 页（本文件 ≥400 行触发）

### 本次决策
| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 球形方案整体放弃，不做锯齿修补 | region 无抗锯齿是 Win32 API 本质限制；Wails 透明链路（WebviewIsTransparent/NOREDIRECTIONBITMAP/Mica/Backdrop）逐层试过全残影，继续投入产出比低 | 微调 region（无解）/ 继续透明链路（已试败） |
| 回滚只回外观，保留扩展模式网格与面板自适应 | 用户否决的是窗口外观三宗罪；4+ 格布局与 MiniMax 功能本身好用 | 连功能一起回退（误伤） |
| dist 重建产物入库 | 项目惯例：产物变更即提交、逐字节相同才回退；保持仓库 dist 与交付 exe 一致 | 回退产物（仓库与 exe 脱节） |

### 新增/变更文件
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 回退 | `main.go` / `workarea_windows.go` / `workarea_other.go` / `app.go` | 窗口层球形代码 revert（e96af01，与 b13a10a 前零差异） |
| 修改 | `frontend/src/style.css` / `main.js` | 球体恢复方圆角，保留网格（537daea） |
| 修改 | `frontend/dist/*` | 回滚后重建产物（本次补交） |
| 修改 | `ROADMAP.md` / `docs/STATUS.md` | 已否决留痕 + 回滚根因（5bc855f） |
| 归档 | `docs/wiki/99-appendix-legacy-status.md` | 第二批 5 条（08-04 ~ 08-25 12:13） |

> 本次变更（b13a10a..537daea）：+100/-38 行，4 个文件（含文档）

### 未完成 & 下一步
- wiki 三件套同步（05/02/07/09/00，`/wiki-update`）——持续中
- 计划级事项见 `ROADMAP.md`（Next: Wails 版本升级对齐）

### 已知问题 & 注意事项
- 球形改版已否决留痕 ROADMAP「已升级/已否决」；若未来重试，先解决两个硬前提：Wails OnStartup 与窗口不同线程（SetWindowSubclass 必失败）、物理像素直调窗口尺寸会搅乱 Wails 尺寸状态（展开面板空白的根源）
- 262px 撑宽为窗口创建时机偶发（回滚构建后未复现），无需处理
- MiniMax Key 曾在对话中明文出现（10-02 记录），介意可去 platform.minimaxi.com 轮换后更新应用配置

### 推荐 Skill
- `/wiki-update` - 检测到 14 个 wiki 覆盖文件自 synced_commit（cc01446）后有变更：`internal/fetcher/{minimax,minimax_test,registry,registry_test,factorydroid,factorydroid_test}.go`、`internal/config/{config,config_test}.go`、`app.go`、`frontend/src/{index.html,main.js,style.css}`、`frontend/wailsjs/go/main/App.{js,d.ts}`（主体为 MiniMax+扩展模式三件套；factorydroid 两文件的展示规则事实已在 3a9eec2 同步进 wiki 05，仅 meta 待推进）

### 关键上下文
- 球形回滚后 UI = v1.2.0 观感 + 扩展模式网格：当前启用 kimi / ollama / factory-droid / minimax（4 格 2x2，`extended_mode=true`）
- **球形失败三层教训**（已留痕，防止重蹈）：① Win32 SetWindowRgn 无抗锯齿，圆缘必锯齿；② Wails v2 + WebView2 透明链路逐层残影（实心底色/Mica 方块）；③ OnStartup 非窗口线程 → SetWindowSubclass 必失败；物理像素直调窗口尺寸会搅乱 Wails 尺寸状态机 → 展开面板全白
- 当前运行实例 PID 23220（回滚版）；桌面 `Quota Viewer.lnk` → `build/bin/quota-viewer.exe` 已是回滚构建
- Wiki 指针状态：`docs/wiki/` 12 文件 + 99 归档页（第二批 5 条），`.covered-files` 50 项，synced_commit `cc01446`（漂移 14 文件待 `/wiki-update`）

---

## 会话记录：2026-10-11 15:14

> **会话摘要**：新增第 10 家 Provider Open Design（浏览器会话 Cookie，5 小时/7 天双窗口）；全程走完整流水线——真机实探定端点、worker 实现、三路审查（运行验收门 VERIFIED / Spec 轴 PASS / Standards 轴 FAIL）、按标签返工 302 与文档欠账后重验 VERIFIED，构建交付并端到端跑通
> **Git**：`2d9031f` on `master`（全部已推送前状态；本会话记录提交后推送）
> **任务组**：Open Design Provider 接入
> **任务组状态**：已完成（功能端到端验证通过并交付运行）

### 本次完成

- **真机实探定端点**（用户提供 key 与浏览器截图，要求「自己上去查」）：`amr-link.open-design.ai` 数据面只有推理端点、**无任何额度接口**（LiteLLM 风格路径全 404）；控制面 `amr-api.open-design.ai` 的额度接口只认浏览器会话 Cookie，API key 一律 401。用真实浏览器会话实探拿到三个端点：`/api/v1/workspaces/current`（发现 workspaceId）、`/api/v1/workspaces/{id}/billing/coding-plan-usage`（5h/7d 双窗口 credits）、`/wallet/balance`（用户否决，未接）。官方开源仓库源码交叉确认额度模型（Vela 后端、窗口池记账、Go 档两窗口）
- **实现 `opendesign.go`**（commit `a74d370`）：两步 GET，窗口按 `durationSeconds` 识别（18000/604800，不靠数组下标）；credits 十进制字符串按 ParseFloat 算已用%；5h 驱动 Used/ResetAt/Remaining 首项，Percent 取两窗口更紧张者；注册表 + AllProviderIDs 双处同步；16 个测试函数（httptest，含倒序窗口、limit=0 除零、空 Cookie 不发请求、真实 id 拼路径等）
- **三路审查**：运行验收门 VERIFIED（12 发变异全杀、11 个边界探针全过、真实端点独立冒烟通过）；Spec 轴 PASS（7 条拍板决定逐条核对无违背）；Standards 轴 **FAIL**（5 条：302 未拦重定向属真偏离、文档速查表欠账、3 条注释/口径失真）
- **返工 r2**（commit `bb276a7` / `c306903` / `1881d69`）：`get()` 补 `CheckRedirect: ErrUseLastResponse` 并把 302 并入 Cookie 失效分支（照 4 个 Cookie 型先例）；补两条 302 用例（先红后绿，红症状与 finding 逐字一致）；`docs/ADDING_A_PROVIDER.md` 速查表补三家欠账 + 新增 MiniMax / Open Design 特殊说明节 + §1.3 Kind 口径与 `types.go` 对齐；两处失真注释修正
- **重验 VERIFIED**：3 发指定变异 + 4 发自选变异全被击杀；301/303/307/308 探针确认全部不跟随、无静默失败；**真实端点独立复验成功**（`percent=5.356`、`remaining="5小时 5.4% 已用 · 周 1.8% 已用"`）
- **端到端交付**：`wails build` → 应用配置启用 open-design 并写入 Cookie → 新实例 PID 30100；悬浮球出现 **OD** 格（绿色），展开面板显示「Open Design 5小时 5.6% 已用 · 周 1.9% 已用」，与其它 4 家同构；README 两家（中英）补齐 Factory Droid / MiniMax / Open Design 三行

### 本次决策

| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 走浏览器会话 Cookie 而非 API key | 实测 API key 读不到任何额度（数据面无额度端点、控制面对它一律 401）；与讯飞/MiMo/Ollama 同类 | 用 API key（不可行） |
| 只取 5 小时 + 7 天两个窗口，**不接**钱包余额 | 用户明确拍板「只要两个窗口」；额度用尽后才轮到钱包，平时是噪声 | 附带展示钱包 $1.50 |
| 只填 Cookie 一个输入框，workspaceId 自动发现 | 用户拍板；换工作区不用重填，与 Cookie 型先例一致 | 多一个手填 workspaceId 字段 |
| 注册第 10 家但扩展模式上限**仍是 9**（10 选 9） | 用户拍板；9 是前端 3x3 球格容量，加第 10 格要改网格布局与字号，收益低于风险 | 球改 5x2 网格支持 10 格 |
| 302 修复照做（尽管真实端点实测返 401） | 文档 §1.4 明文 + 4/4 Cookie 型先例一致；成本低，防的是平台改行为 | 不修（文档与先例不一致，留给后人踩） |
| 顺带补齐文档欠账（速查表三家、README 两家、两处注释） | 审查判定「provider 票把自己登记进指南是同一件事的收尾」；欠账已积三轮 | 只修 302，欠账留给后续会话 |

### 新增/变更文件

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/fetcher/opendesign.go` | Open Design 抓取器（两步 GET + 双窗口 + 302 拦截） |
| 新增 | `internal/fetcher/opendesign_test.go` | 16 个测试函数（httptest，含 302 不跟随断言） |
| 修改 | `internal/fetcher/registry.go` | 注册 open-design（第 10 家，Abbr=OD） |
| 修改 | `internal/fetcher/registry_test.go` | 注册表断言 9→10 |
| 修改 | `internal/config/config.go` | AllProviderIDs 追加 + 注释修正（上限 9 是球格容量） |
| 修改 | `internal/config/config_test.go` | 3 处断言适配 10 家（扩展模式改为「恰好 9 开 + 末位被钳」双断言，属加强） |
| 修改 | `docs/ADDING_A_PROVIDER.md` | 速查表补三家 + MiniMax/Open Design 特殊说明节 + §1.3 Kind 口径 |
| 修改 | `README.md` / `README.zh-CN.md` | 支持 Provider 表补齐三家 |
| 修改 | `docs/STATUS.md` | 本会话记录 |
| 修改 | `%APPDATA%/quota-viewer/config.json` | 启用 open-design + 写入 Cookie（应用外修改，已备份 `.bak-20261011`） |

> 本次变更（537daea..HEAD）：16 文件，+994/-236 行（含本会话前 b1fcb85 归档提交的 202 行）
> 代码变更（b1fcb85..HEAD）：9 文件，+715/-15 行

### 未完成 & 下一步

- wiki 同步：本会话已派 worker 同步十平台与行数基线（`/wiki-update` 口径），完成后待查
- 计划级事项见 `ROADMAP.md`（Next: Wails 版本升级对齐）

### 已知问题 & 注意事项

- **Open Design 额度只能靠浏览器会话 Cookie**：API key（sk-amr-）读不到额度；Cookie 是 better-auth 的 `__Secure-better-auth.session_token`，**HttpOnly**（`document.cookie` 拿不到），须从 DevTools → 网络 → 该请求 → 请求标头复制整条 `Cookie`。Cookie 失效实时测返回 **401**（不是 302），代码里 302 拦截是照文档与先例补的防御
- **第二步不能用 `current` 别名**：`/api/v1/workspaces/current/billing/coding-plan-usage` → `workspace_member_required`；必须先取真实 workspaceId（有测试守护）
- 该控制面是官方前端私有接口，无稳定性承诺；失效时对照 open-design.ai 的 web bundle 更新 `opendesign.go`（已入 ROADMAP Later）
- 取 Chrome Cookie 时踩到的坑：`Network/Cookies` SQLite 被浏览器独占锁，`sqlite3` 免锁模式、`.NET FileShare` 全开、`robocopy /b`（需备份权限）全部打不开 —— 直接读 DevTools 请求标头最省事
- 本会话真实 Cookie 只在临时探针与配置文件中出现，**未落进任何仓库文件**（已全仓扫描确认）；介意可改 open-design.ai 重新登录使其失效
- 两个验收实例都披露了同一工具陷阱：gate-kit 的变异脚本按**进程 cwd** 解析 `--file`，从主仓发起会把变异打到主仓文件。两次都即时还原，主工作区已用 `git hash-object` 逐文件核验与 HEAD 一致、无残留

### 推荐 Skill

- `/wiki-update` - 本会话已委派同步（十平台措辞、05 表补两行、02 模块行数、07 的 AllProviderIDs、00 的 wiki-meta 与 covered-files）

### 关键上下文

- **Open Design 端点契约**（真机实测 2026-10-11）：`GET https://amr-api.open-design.ai/api/v1/workspaces/current` → `workspaceId`（本机 `p619hvqak6gwawty2rfzl7mb`，tier=go）；`GET .../api/v1/workspaces/{id}/billing/coding-plan-usage` → `windows[]`，`durationSeconds` 18000=5h / 604800=7d，`usedCredits`/`limitCredits`/`remainingCredits` 是十进制字符串，计数是**已用**，`resetsAt` 为毫秒精度 RFC3339
- **额度模型**（官方开源仓库 nexu-io/open-design）：计费真相在 Vela 后端；Coding Plan 按窗口池记账（Go 档 = 5h + 7d 两窗口，Plus/Pro/Max = 7d 单窗口）；钱包 `balanceUsd` 与 plan 额度并行（plan 用尽转按量）
- 当前启用 Provider：kimi / ollama / factory-droid / minimax / open-design（5 家，`extended_mode=true`，上限 9）
- 运行实例 PID 30100（含 open-design 的构建）；桌面 `Quota Viewer.lnk` → `build/bin/quota-viewer.exe` 已是本次构建
- 本会话实现/返工提交：`a74d370`（实现）→ `bb276a7`（302 修复）→ `c306903`（文档）→ `1881d69`（注释）→ `2d9031f`（README）
- 记忆：`.agents/memory/pending/2026-10-11-open-design-amr-link-no-quota-api.md`（API key 死路 / Cookie 路径 / 三个端点 / 401 而非 302 / Cookie 库锁坑）
- Wiki 指针状态：`docs/wiki/` 12 文件 + 99 归档页（10 条），synced_commit `cc01446`（本次派 worker 同步中）

---

## 会话记录：2026-10-11 16:10

> **会话摘要**：用户目验后提两笔悬浮球格布局修正——5 格末行居中（五环式）、窄格里双字符缩写（MX/OD）不再粘连；均在渲染探针上量测验证后构建交付
> **Git**：`9760924` on `master`
> **任务组**：悬浮球格布局打磨
> **任务组状态**：已完成（两笔均已构建交付、运行中实例生效）

### 本次完成

- **5 格末行居中（commit `757f4fd`）**：扩展模式排布从 CSS Grid 改为 `flex-wrap` + `justify-content: center` —— 末行不满时整行自动居中。5 格由「3+2 左对齐」变为居中的五环式；**7 格（3+3+1）、8 格（3+3+2）的末行同样受益**（同类毛病一起修）。每格宽 `calc((100% - (cols-1)*1px)/cols - 0.5px)`，留 0.5px 余量防浮点误差把第三格挤到下一行
- **窄格双字符间距（commit `9760924`）**：根因是格子等宽仅 18.8px，而双字符缩写在 10px 下占 15.4px（两侧各剩 1.7px，相邻两格文字净空仅 4.4px，单字符是 13.7px）；3 格时 13px 字号下 "MX" 宽 20.1px **已溢出格子**。修法：JS 在窄格场景（3 格单行 / 5 格及以上三列）加 `narrow` class、给多字符缩写格加 `wide` class，CSS 把 `.narrow .wide` 降到 8.5px；宽格（1-2 格、4 格 2x2）字号不动
- **量测验证**（renderNode 探针，非目测）：把 1-9 格全量渲染，用 Range 量每对相邻文字的精确净空 —— 改后最紧一对（MX\|OD）由 4.4px 提到 6.3px；各格数最小净空：3 格 6.8px、5-9 格 6.3px、宽格 9.9-18.5px（宽格未被误伤）
- **交付**：`wails build` → 新实例 PID 35032；真机截图确认第一行 K O F、第二行 MX OD 居中且分得开

### 本次决策

| 决策 | 原因 | 备选方案 |
|------|------|----------|
| 末行居中用 flex-wrap 而非 Grid 子轨道 | 子轨道方案（每格跨 2 轨 + 显式定位）会在格子内部夹进多余 1px 间隙，且自动布局与显式定位混用行为不可预期 | CSS Grid + `grid-column: span` 显式定位（内部会多出分隔） |
| 双字符按字号降档，而非改格子宽度 | 格子必须等宽（球仅 60px，等比缩放会破网格）；字号是唯一可调量 | 双字符格加宽（破等宽）/ 缩写改单字符（丢失辨识度，且注册表 Abbr 是契约） |
| 3 格也纳入「窄格」 | 实测 3 格最糟：13px 下 MX 宽 20.1px > 格子 20px，是溢出而非拥挤 | 只处理 5+ 格（漏掉唯一真正溢出的场景） |
| 降到 8.5px 而非 9px 或 8px | 8.5px 净空 6.7px 已明显分开，且与单字符 10px 的视觉大小差最小；9px 仍偏挤、8px 偏小 | 9px（净空 5.9px，改善有限） |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `frontend/src/main.js` | updateBall：narrow/wide class 标记 + `--cols` |
| 修改 | `frontend/src/style.css` | 网格改 flex-wrap 居中 + `.narrow .wide` 降字号 |
| 修改 | `frontend/dist/*` | 重建产物（新 hash，与交付 exe 一致） |

> 本次变更（2d9031f..9760924）：19 文件，+260/-115 行（含 wiki 同步 12 文件的 131 行）
> 布局修复本体：前端 2 文件 +6 行净增，产物 3 文件

### 未完成 & 下一步

- 无遗留；计划级事项见 `ROADMAP.md`（Next: Wails 版本升级对齐）

### 已知问题 & 注意事项

- **球格布局的两个约束**（本次实测得来，改布局前必读）：① 格子必须等宽（`flex: 0 0 calc(...)`），球仅 60px 宽，任何等比缩放都会破网格；② 双字符缩写（MX/OD/Go）在窄格里只能靠降字号换净空，字号档位与 `--cols` 绑定
- 布局验证的有效手段：起本地静态服务 + Playwright 探针页（复刻 style.css 规则）渲染 1-9 格，用 `document.createRange()` 量文字**实际**边界（而非格子边界）——目测在 60px 尺度上不可靠
- 探针页留在 `.scratch/ballprobe/`（未入库，`.scratch/` 未跟踪）

### 推荐 Skill

- `/wiki-update` - 本次已委派同步（02/03/08 三页的球格渲染与前端行数）

### 关键上下文

- **格布局现状**：1-3 格单行；4 格 2x2；5-9 格 3 列换行 + **末行居中**；字号档：1 格 20px、2-3 格 13px、4-6 格 10px、7-9 格 9px、**窄格里的双字符 8.5px**
- 双字符缩写当前有：MiniMax=`MX`、Open Design=`OD`、OpenCode Go=`Go`（注册表 `registry.go` 的 Abbr 是契约，改动需同步 `frontend/src/main.js` 的字号假设）
- 运行实例 PID 35032（含两笔布局修复）；桌面 `Quota Viewer.lnk` → `build/bin/quota-viewer.exe` 已是本次构建
- 本会话布局提交：`757f4fd`（末行居中）→ `9760924`（双字符间距）
- Wiki 指针状态：`docs/wiki/` 12 文件 + 99 归档页（10 条），本次布局变更已派 worker 增量同步（02/03/08）

---

## 会话记录：2026-10-13 13:30

> **会话摘要**：给 quota viewer 加无界面 `--json` 探针（对端 handoff 交付）+ 修两家 Provider 的漂移——OpenCode Go 换 API Key 通道（旧 `auth` cookie 已废）、Factory Droid 在 Standard 顶格且 Core 接管时把展示主体切到 Core；两条 spec 各走完整流水线（逐票验收门 + 整体验收门全部 VERIFIED），并交付重建的 exe
> **Git**：`6d8cdee` on `master`（工作区 tracked 干净）
> **任务组**：Provider 漂移修复 + `--json` 探针交付
> **任务组状态**：已完成

### 本次完成

- **OpenCode Go 换通道**（票 01，`9ea337b`，门 VERIFIED）：opencode.ai 把 dashboard 页面换成了 console，旧 `auth` cookie 通道全废（真机取证：带旧 cookie 与匿名请求**完全同形**，都是 302 → `/console/login`）。改走官方 `GET https://opencode.ai/zen/go/v1/usage` + `Authorization: Bearer <API Key>`；registry 收成单字段 `api_key`（留空自动读 `~/.local/share/opencode/auth.json`），HTML/SSR/data-slot 解析整段删除；401 / 403(`EntitlementError`) / 其他 403 / 500 / 非 JSON 五档文案。门真机跑通 200/401/403 三路径并逐字段对拍。
- **Factory Droid Core 接管展示**（票 02，`6b980e4`+`c115324`+`e2b7356`，门 VERIFIED）：上游有 `overagePreference: "droidCore"` 显式信号（原先代码没解析）；Standard 顶格（`stdMax >= 100`）且超额走 Core 时，`Percent`/`Used`/`ResetAt`/`Remaining` 全部切到 Core 池（文案 `Core X.X% 已用（Standard 已限流） · 周 …`），重置时间取 Core 驱动窗口的 `secondsRemaining`。返工 r2 修掉平手 tie-break 写反（`>=`→`>`，平手应取先出现者）并把 ResetAt 断言从 ±60s 收紧成亚秒夹逼。门用 28 夹具矩阵 + 600 条随机对拍 + 14 发变异全杀。
- **加固与残留清理**（票 03，`35254f4`+`a3b09c5`+`6847da0`，门 VERIFIED）：`Fetch()` 显式声明 `Kind: KindUsage`、`User-Agent` 断言补牙、删掉 `internal/config` 里两个已被坐实无人读取的旧扁平格式死 key。
- **无界面 `--json` 探针**（json-mode spec，13 笔提交 `73ab272`…`6d8cdee`，整体门 VERIFIED）：新增 `cmd/quotaviewer`（console 子系统独立 main）+ `internal/probe`——一行合法 JSON（顶层 7 键、每条记录 8 键、`reset_at` 三条规则、失败记录五字段 `null`、极性不反转、`remaining_percent` 钳制）、退出码三分（0 含失败信封 / 2 usage / 1 写 stdout 失败）、顶层失败同时写 stderr。过程：r1 门 REJECTED（单测层真读真实凭证、真打上游）→ r2 VERIFIED → 整体门 Spec 轴又查出两处 spec 明写未达成（失败未写 stderr；`--json` 内参数失败未出信封）→ r3 补齐，整体门 VERIFIED。
- **交付**：`wails build`（13:18）出 `build/bin/quota-viewer.exe`，新实例 PID 45528 在跑，球面 Go 那格已由红转绿；CLI 探针 13:22 重建并复验 r3 六态；用户 2026-10-13 新订阅的 API Key 已写进 `%APPDATA%\quota-viewer\config.json`（旧的 `session_token`/`workspace_id` 一并清掉，改前备份 `config.json.bak-20261013-opencodekey`）。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| OpenCode Go 用 API Key，不走 console cookie | 官方 API 不随浏览器 session 轮换过期；console cookie 通道还要 `x-org-id` 且会过期 | 抓 `/console/api/go/status`（`__Host-console_session` + 微分字段） |
| 不给 `ResetsAt`/`Percent` 加类型容错 | 上游改字段类型就是漂移，本仓惯例是响亮报错；静默降级会把漂移藏起来 | `json.RawMessage` + 自定义 `UnmarshalJSON` |
| 接管判定 = `stdMax>=100 && coreHasData && overagePreference ∈ {"", "droidCore"}` | 上游有显式信号就不猜；`extraUsage` 时不把主体切过去 | 维持 `Percent = max(std, core)`（会把球面钉死在 100%） |
| `--json` 内参数失败出信封 + 退 0；不带 `--json` / 无参 → usage + 2 | spec 约束 8 字面（参数失败仍出同形 JSON 并退 0）+ 契约不适用于非 `--json` 调用 | 一律 usage + 2（r1 老口径，与约束 8 字面冲突） |
| 失败只写顶层 stderr、单家失败不写 | spec §二「诊断旁路」+ 约束 5（单家失败由记录里的 `error` 承载，写 stderr 只是噪声） | 每家失败都写 stderr |
| 探针四句归一扩成「空 `Kind` 先取注册表声明」 | 对真实注册表十家取值等价（5 家空 Kind 的声明全是 `usage`，deepseek 自设 `balance`），且让「归一顺序」变异可被杀 | 逐字照抄 `app.go` 四句（则顺序变异不可杀） |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|---|---|---|
| 新增 | `cmd/quotaviewer/main.go` | CLI 入口（console 子系统；逻辑在 `internal/probe`） |
| 新增 | `internal/probe/probe.go`、`report_test.go`、`probe_test.go` | JSON 契约、结果归一、`reset_at` 归一、失败信封、退出码 |
| 修改 | `internal/fetcher/opencode_go.go` + `_test.go` | 换 `/zen/go/v1/usage` + API Key（老 HTML 解析全删） |
| 修改 | `internal/fetcher/factorydroid.go` + `_test.go` | `overagePreference` 解析 + Core 接管时切展示主体 |
| 修改 | `internal/fetcher/registry.go`、`internal/config/config.go`、`docs/wiki/10-build-test-baseline.md` | 单字段 `api_key`、删死 key、构建基线补探针节 |

> 本次变更（代码基准 `01c1920`）：15 文件，+1863/−427，20 个提交
> 文档基准 `9760924` 起算：19 文件，+1944/−440

### 未完成 & 下一步

- **wiki 同步（6 页，`/wiki-update` 一次做完）**：`05`（opencode-go 改写成 API Key + `/zen/go/v1/usage`；droid 段补 Core 接管语义）、`07`（删旧字段清单与「有值的旧字段 → 凭证迁移」）、`02`（模块表补 `cmd/quotaviewer`/`internal/probe`）、`00`（目录速览 + `wiki-meta` + 测试条数）、`09`（120 条 → 现行口径）、`10`（条数漂移 + 契约落点）。
- **非阻断清理批次（待用户拍板，9 条见「关键上下文」）**：多数会作废对应票的 VERIFIED、需重验。
- **对端告知（仓外动作，本仓做不到）**：探针可用 + 偏离清单，要点见「关键上下文」。
- 计划级事项见 `ROADMAP.md`。

### 已知问题 & 注意事项

- **Gate Kit 的 `mutate` 子命令在本机不可信**：同一发变异经它报「存活」、手动施加后立刻 16 处红（两发复现，根因未定位）；返工 r3 也自曝过「变异脚本没把 `go test` 的 cwd 切进副本 ⇒ 11 发假绿」。**后续变异协议以手动施加 + sha1 核验为准。**
- **`APPDATA` 指向不存在的盘符时探针输出「成功形状」**（exit 0、`error=""`、各记录「未配置」）：`config.Load()` 在 `configPath()` 出错时吞错返回 `Default()`，而预检只能判空字符串。spec 只要求「判空即可」，故不阻断，已列入对端告知。
- `secondsRemaining` 若漂移成毫秒/超大值（门实测 `1e18`）→ `ResetAt` 会显示 1734 年。非阻断。
- `go test -race` 在本机跑不了（MinGW 缺 `crt2.o`）；`wails` 不在 PATH，用 `C:\Users\joe\go\bin\wails.exe`。
- 本机 Droid 的 Standard 周限与 Core 5h 双双顶格、mcode 通道 Token Plan 用量上限也未恢复 ⇒ 非 DSH 侧的子代理派发全线阵亡（当天实测）。

### 推荐 Skill

- `/wiki-update` - 检测到 wiki 覆盖文件自 `synced_commit`（`9760924`）后有代码变更：`internal/fetcher/{opencode_go,factorydroid,registry}.go`、`internal/config/config.go`，另有新增包 `cmd/quotaviewer/`、`internal/probe/`（README×2 与 `docs/ADDING_A_PROVIDER.md` 不属 wiki 覆盖面）

### 关键上下文

- **交付现状**：`build/bin/quota-viewer.exe` 13:18 构建、PID 45528 在跑；球面 K 绿 / **Go 绿** / O 黄(69.5%) / F·MX·OD 灰(100%)。CLI 探针 13:22 重建（`--json` → 退 0、单行、stderr 空）。
- **OpenCode Go 新通道**：`GET https://opencode.ai/zen/go/v1/usage` + `Authorization: Bearer oc_sk_…`（**前缀是下划线**）；响应 `usage.{rolling,weekly,monthly}.{status,percent,resetsAt}`，`percent` 是 0–100，**重置时间是 `resetsAt` 绝对时间**（社区文档写的 `resetInSec` 是旧通道字段，别照抄）。失败面：无效 key→401 `AuthError`；旧订阅 key→403 `EntitlementError`。若将来需要 cookie 通道：`GET https://opencode.ai/console/api/go/status`，cookie **`__Host-console_session`**（`st_` 开头）+ 头 `x-org-id`，响应是微分字段。
- **Factory Droid 口径**：`overagePreference: "droidCore"` 是上游显式信号；Standard 顶格 + 超额走 Core 时展示主体切 Core；重置时间一律用 `secondsRemaining`（`windowEnd` 与它同源、只差约 1s）。
- **待用户拍板的 9 条清理项**：① `registry_test.go` 补「声明键 == 实际读取键」与「敏感字段 `Type==password`」断言（门实测 `api_key`→`session_token`、`password`→`text` 两发变异全量套件仍绿）② `Save`/`Load` 剔除未声明 cred key + 用例（门实测该修法无测试）③ `oc_sk-` → `oc_sk_`（`registry.go:55` + README×2 + `ADDING_A_PROVIDER.md:341`，改它会作废票 01/03）④ `factorydroid.go:192` 与测试注释「windowEnd 是过期值」改正 + 补 `secondsRemaining ≤ 0` 夹具（改它作废票 02）⑤ wiki 3 页 ⑥ `factorydroid.go` 三处窗口枚举合并的归属（票 02/03 互相指、无人做）⑦ `ADDING_A_PROVIDER.md:131` 的「必须显式设 Kind」对另外 4 家仍不成立 ⑧ `opencode_go.go:189` 不填 `Used`（基线即如此、前端不消费）⑨ `35254f4` 用了 `test+fix:` 双类型提交。⑩ `internal/probe/probe.go:2` 的包注释把契约指向未入库的 `.scratch/json-mode/spec.md`（与 wiki 10 的指针成环、外读者无落点）⑪ `app.go:402-403` 注释写「结果顺序 = 注册表顺序」，实现是**配置顺序**（既有注释错，探针声称与它同构 ⇒ 会误导对账）⑫ `factorydroid.go:193-197` 脏 `secondsRemaining`（如 `1e18`）会让 `ResetAt` 显示 1734 年（上游漂移情景）。
- **对端告知报文要点**：二进制名 `quota-viewer-cli.exe`、开关名 `--json`（另接受 `-json`、`--json=1`）、构建命令 `go build -o build/bin/quota-viewer-cli.exe ./cmd/quotaviewer`；已知字面偏离——`internal/probe/probe.go` 的四句归一扩成「空 `Kind` 先取注册表声明」（对真实注册表取值等价）、`-json` 别名、`--json=1` 判为「未知参数 + 同时声明 JSON 模式」、退出码三态（0 含信封 / 2 usage / 1 写 stdout 失败）、单家失败不写 stderr、`APPDATA` 存在但路径不可用时退化为默认配置的「成功形状」。
- **未决疑问**：本次有一次观察到工作区里**未提交**的 `registry.go` 改动（`password`→`text`、label 简化成 `API Key`），非任何 worker 所为，随后被某次变异还原覆盖、现已不在工作区；是否要做待用户确认（若做，顺带修 `oc_sk-` 笔误）。
- Wiki 指针状态：`docs/wiki/` 12 文件（00–10 + 99），本轮需同步 6 页；99 归档页 10 条。

---

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

---

## 会话记录：2026-10-29 02:45

> **会话摘要**：按两段式把 Devin（第 11 家 Provider）落地——三张票逐票实现 + 逐票验收门（01/02/03 各返工一轮）+ 整体双轴审查 + 集成验收门 VERIFIED；wiki 与文档同步，界面目视仍待用户
> **Git**：`81d4224` on `master`（Devin 任务组 10 个提交，未发版）
> **任务组**：Devin Provider（承 2026-10-28 23:21 规划会话）
> **任务组状态**：实现与验收已完成；界面目视（AS-11b）与发版待用户

### 本次完成

- **三张票逐票落地，每票都过独立动态验收门**：
  - 01 手填 token 全链路（`69a0f3e`，6 文件 1069 行 ⇒ 超限票，先跑静态双轴）：`devin.go` 请求/解析/展示映射 + 注册表第 11 家 + `config.AllProviderIDs`；门 r1 判 REJECTED，拒因只在测试侧（并列取日 / null 容错 / 单侧缺百分比三条口径无断言，变异 M4/M15/M16 存活），返工 `ce19008` 补 4 个用例后 r2 VERIFIED。
  - 02 本机登录态自动读取（`9848208` + `acbf6d1`）：`state.vscdb` 按字节扫 token 与 `apiServerUrl`、server 三级兜底、默认值不回写字段；超限（553 行）先静态双轴；门 r1 REJECTED（11 处用例未隔离 `APPDATA` 且注释失实 + `devin.go` 注释与生产事实相反），返工 `24a006d` 后 r2 VERIFIED。
  - 03 文档与真机冒烟（`c031ca5` + `4adfe6c`）：两份 README 改 11 家、`ADDING_A_PROVIDER.md` 加 Devin 速查行与专节、ROADMAP 记 web 通道退路；真机直连取证（真实 `APPDATA`、凭证留空）十键齐、`percent=0`、`remaining=100`、`reset_at` 非空；门 r1 REJECTED（「多命中形态」被写成会退化为「未配置」，与代码不符），返工只改一句后 r2 VERIFIED。
- **整体验收门 VERIFIED**：全范围双轴静态审查（基线 `8e0a691`）+ 集成动态门；门独立重跑「前沿普查」（本仓无 `frontier.mjs`，按逐票 Status 取证：3/3 `resolved`）、全量 `go test ./...`、`wails build`（23.6s）、AS-1..AS-11a 逐条成立性、跨票接缝红队、三条红队预测亲手复验。
- **顺手修掉整体门挑出的两处硬违规**：`16fd6f0` 订正 `ADDING_A_PROVIDER.md` 的 400 口径（实现只透出 `HTTP 400`）、`dc127dc` 订正 `devin.go` 字段注释的 server 兜底语义（旧口径「留空 = 默认端点」已作废）。
- **wiki 同步**（`81d4224`）：00/01/02/03/05/07/08/09 八页 + `.covered-files`，残留检索「十家/十平台」清零，`wiki-meta` 推进到 `dc127dc`。

### 本次决策

| 决策 | 原因 | 备选方案 |
|---|---|---|
| 动态门 r1 的「单侧缺百分比」与「超额金额」两条由主会话补口径、不挂起 | 票面只是欠规定：前者按「不可见」比渲染成 100% 已用安全，后者票面晚于 spec 且出现两次 | 挂 `needs-human` 等用户裁决（会白停一轮） |
| 「超额金额」判为票面覆盖 spec | 票面是本票操作判据、晚于 spec、示例金额按夹具重算过；spec 已同步并留作废说明 | 改回「超额余额」（要动实现 + 测试 + 文档三处） |
| 首命中点（多命中取第一个）不返工 | 票面未规定多命中语义，失败模式安全（退化为「未配置」）；由票 03 真机冒烟兜住，文档已如实写清 | 改成扫到第一个带 payload 的命中（属新范围，需新票） |
| 票 02 的「手填 token 仍读整份 state.vscdb」只订正注释、不改行为 | 行为是 AS-8 server 兜底的直接后果；改行为会动已 VERIFIED 的实现 | 加「只缺 server 才读文件」优化（作废 VERIFIED、需重验） |
| 整体门挑出的 wiki 漂移与 ROADMAP 过期条目在收尾一次做掉 | 两者都是「文档写错事实」，越留下一个会话越贵 | 留给下个会话（带着错文档交付） |

### 关键文件（≤ 5 个）

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/fetcher/devin.go` | Devin 抓取器：GetUserStatus、剩余→已用、本机登录态、三级 server 兜底（484 行） |
| 新增 | `internal/fetcher/devin_test.go` | 44 个用例，含隔离 APPDATA 的本机凭据与请求形状断言（1177 行） |
| 修改 | `internal/fetcher/registry.go` + `internal/config/config.go` | 注册第 11 家 `devin`（DV / usage / app.devin.ai）+ `AllProviderIDs` |
| 修改 | `docs/ADDING_A_PROVIDER.md` + `README.md` + `README.zh-CN.md` | 11 家口径、Devin 专节、server 三级兜底例外、400 口径订正 |
| 新增 | `.scratch/devin-provider/{spec.md,issues/,verifier-probes/}` | 不出库：spec、三票与逐票裁决、三轮返工的攻击面与变异日志 |

> 本次变更：`8e0a691..81d4224` +1811/−39 行，18 个文件（其中 9 个是 wiki 与文档）
> 变更基准：`git diff --shortstat 8e0a691..81d4224`

### 未完成 & 下一步

- **AS-11b 界面目视**（唯一未验项，只能你本人做）：桌面实例在跑（devin 已启用、凭证留空、exe 已是含票 01+02 的重建产物），看球格 `DV` 与面板日/周百分比是否与 Devin Desktop 的 Usage 页一致
- 目视通过 → 发版 v1.5.0（11 家）；未通过 → 记下实际显示，按症状重开票
- 「对端告知」报文（必须带 probe `schema` 1→2）
- 非阻断清理批次：原有 12 条 + 本轮 6 条（`devin_test.go:835` 提示文案、`devin.go` 浮点兜底无测试、两份 README 未提「超上限 Provider 静默禁用」、`clampPercent` 寄居 `factorydroid.go`、两个 UnmarshalJSON 同形、`applyPlanStatus` 不读接收者）

### 已知问题 & 注意事项

- **`.scratch/` 被 `.gitignore` 忽略**：票面的 `Status` 改动、`## Verdicts` 裁决行、`## 实现说明` 都进不了 commit —— 这是本仓既有约定（内部过程产物不入公开仓），不是缺陷；跨会话可靠性靠本地文件本身
- **`wails build` 会把 `frontend/dist/*.css` 与 `frontend/wailsjs/go/main/App.{d.ts,js}` 置成 CRLF**，`git status` 变脏但 `git diff` 为空：每次构建后 `git checkout -- frontend/` 还原，别顺手提交
- **门用 Gate Kit 副本时会镜像宿主目录（junction）**：在副本里 `wails build` 也可能写穿到主工作区的 `frontend/`，收尾务必核一次 `git status`
- 无 git stash；仓库仍是 `main/master` 单线，无额外 worktree

### 推荐 Skill

- `/update-status` - 本会话记账（已执行）
- `/wiki-update` - 已执行（`81d4224`）；后续改 `devin.go` 后重跑
- `/code-review` - 整体门双轴已跑；后续大改再跑

### 关键上下文

- **三张票的裁决链**（新会话要追细节时按这个顺序读）：`.scratch/devin-provider/issues/01*.md` → `02*.md` → `03*.md` 的 `## Verdicts`；每票格式是 `<日期> r<N> <VERDICT> <HEAD> <一句话>`。本轮共 6 次门裁决：01 r1 REJECTED → r2 VERIFIED；02 r1 REJECTED → r2 VERIFIED；03 r1 REJECTED → r2 VERIFIED；整体门 VERIFIED。
- **返工全部是机械项，没有一次是实现缺陷**：01 缺测试断言、02 缺测试隔离 + 注释失实、03 文档一句话与代码不符。这说明本块的实现质量靠「门亲手跑变异」兜住，而不是靠实现者自检 —— 三次都是实现者自检全绿、门一打变异就露。
- **Devin 的真实字段形态已定档**（真机两次独立取证）：`dailyQuotaResetAtUnix`/`weeklyQuotaResetAtUnix` 是字符串；`planName` 在 `planInfo` 里而不在 `planStatus` 顶层；`hideDailyQuota`/`hideWeeklyQuota`/`overageBalanceMicros` 在本机**整个缺失**（所以「缺失 vs false」仍未证实，实现按「缺失 = 不隐藏」处理）；响应约 405 KB，带顶层 `planInfo` 与 `planStatus.planInfo` 两份（`devinInfo` 只在后者）。
- **探针契约没动**：仍是 schema 2、每记录十个键（无 `reset_label` / `remaining`）；本块所有真机判据只用了这十个键。
- Wiki 指针状态：`docs/wiki/` 12 个常规页 + 99 归档页（23 条），`.covered-files` 61 项，`synced_commit` = `dc127dc`。
