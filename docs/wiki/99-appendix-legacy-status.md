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

