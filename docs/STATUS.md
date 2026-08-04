# 📋 项目进度文档

> **⚠️ 新会话阅读指南（必读）**
>
> 本文档较长，**不要通读全文**。按以下顺序阅读：
>
> 1. 阅读下方「上下文摘要（TL;DR）」快速了解项目状态
> 2. 阅读下方「项目概况」了解项目基本信息
> 3. 如果下方有「知识库」区块且状态为已初始化，读取知识库获取项目结构和函数映射
> 4. **直接跳到最新一条「会话记录：2026-08-04 14:10」章节**（用标题搜索定位，不要假设在文件末尾）
>    - 重点看「未完成 & 下一步」和「关键上下文」
> 5. 如果最新章节引用了更早的内容，再按需回溯
> 6. 给用户一份简短进度汇报（3-5句话）：当前在做什么 → 做到哪了 → 下一步是什么 → 有无阻塞项
> 7. **严禁在用户给出指示之前修改任何文件或代码**
> 8. 汇报后等待用户指示，不要主动开始执行任务
>
> 📍 最新章节位置：`## 会话记录：2026-08-04 23:02`（搜索定位，可能不在文件末尾）
> 🔖 对应 commit：`b62e6d5` on `master`
> 📊 累计会话数：4 次

---

## 最后更新

<!-- git-meta: {"last_commit": "b62e6d5", "branch": "master", "dirty": false, "timestamp": "2026-08-04T23:02:00Z"} -->

- **日期**：2026-08-04 23:02
- **会话摘要**：新增 Agent 专用 Provider 添加指南文档（`docs/ADDING_A_PROVIDER.md`），双语 README 更新指向该文档

---

## 上下文摘要（TL;DR）

- 项目：Quota Viewer，桌面悬浮球 + AI 平台额度监控工具，Go + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite）
- 当前阶段：**v1.0.0 已开源发布**——GitHub 仓库 eeljoe/quota-viewer + Release v1.0.0（含 exe 下载）
- **当前唯一目标**：无明确待办；项目已交付，等待用户后续指示
- 下一步：可选——Release 推广 / 新 Provider 扩展 / 用户反馈迭代
- 注意事项：工作区干净（全部已提交并推送）；无阻塞项

---

## 项目概况

- **项目名称**：Quota Viewer
- **技术栈**：Go 1.24 + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite 打包）
- **项目根目录**：`C:/Users/joe/Desktop/工作学习/软件开发/quota viewer`
- **平台**：Windows 10+（WebView2 运行时）
- **功能**：悬浮球（1-3 格动态，颜色=状态）+ 展开面板（进度条 + 剩余量明细）+ 配置面板（勾选 Provider + 各平台凭证）+ 系统托盘 + 关闭到托盘
- **支持 Provider**：Kimi（API Key）、讯飞星辰（Cookie）、OpenCode Go（Workspace ID + Token）、小米 MiMo（Cookie）、DeepSeek（API Key，余额型）

## 当前分支与最近提交

- **分支**：master
- **HEAD**：`e60f844`（工作区大量未提交改动）
- **最近提交**：
  - `e60f844` - fix: LockOSThread for tray message pump, replace RunWithExternalLoop with systray.Run
  - `447e0cc` - chore: 回滚后重新构建前端产物与 wailsjs 绑定
  - `a6b5e65` - Revert "fix: 移除 WS_EX_TOOLWINDOW,使窗口回到任务栏和 Alt+Tab"
  - `6fcdb28` - fix: 移除 WS_EX_TOOLWINDOW,使窗口回到任务栏和 Alt+Tab
  - `c5d8eeb` - docs: 更新 README 为真实项目文档

---

## 当前任务进度

### ✅ 已完成（本次升级会话）
- **fetcher 注册表**：`registry.go`（ProviderDef + 凭证字段定义 + Build 工厂，5 个 Provider 固定顺序）；`QuotaResult` 扩展 `ID/Abbr/Kind`（usage/balance）
- **恢复 MiMo**：`mimo.go` + `mimo_test.go`（从 git 历史恢复，httptest 全过）
- **新增 DeepSeek**：`deepseek.go` + 测试（`GET https://api.deepseek.com/user/balance`，余额型，多币种自动选非零）
- **配置模型 v2**：`Config.Providers []ProviderConfig` 动态结构；旧扁平格式（含 mimo_cookie）Load 时自动迁移并回写；默认启用 kimi/xfyun/opencode-go；钳制最多 3 个
- **app.go 动态编排**：`SaveConfig([]ProviderInput, int)` 新契约、GetConfig 返回全部 Provider 元数据（fields/掩码 creds/login_url）、fetchAll 按启用列表 1-3 并发、TestConnection 走注册表
- **前端**：球格动态重建（1 个占满放大/2 个各半/3 个各 1/3）、配置面板按注册表元数据动态生成（勾选 1-3 限制 + 测试 + 打开登录页）、余额型恒绿
- **构建验证**：`go test ./...` 全绿；`npm run build` 成功；`wails build` 成功；exe 启动冒烟通过
- **开源发布**：7 commit 拆分提交并推送 → GitHub 仓库 eeljoe/quota-viewer 创建 → 双语 README（英文默认 + 简体中文）→ Release v1.0.0（含 exe 附件）
- **凭证安全**：git 历史全量扫描 0 命中；exe 二进制扫描 0 命中；config.json 在 `%APPDATA%` 不在仓库
- **wiki 同步**：00/01/02/03/05/07/08/09 已更新至五平台动态模型，`.covered-files` 重新生成（46 项）

### 🔄 进行中
- 无

### 📋 待办
- 无明确待办；项目已交付开源

---

## 未完成 & 下一步

1. **无明确待办**——项目已交付开源（GitHub 仓库 + Release v1.0.0）
2. 可选方向：Release 推广 / 新 Provider 扩展 / 用户反馈迭代 / Wails 版本升级（CLI 2.13.0 vs go.mod 2.12.0）

---

## 已知问题与注意事项

- **提交复杂性**：工作区混合「会话前遗留（OpenCode 替换 + fitToScreen，其中 mimo.go 删除已 staged）」与「本次升级（mimo.go 恢复等）」；`git add -p` 或按文件粒度拆分，mimo.go 删除/恢复需先 `git restore --staged internal/fetcher/mimo.go`
- Wails CLI 版本 v2.13.0 > go.mod 的 v2.12.0（构建警告，不阻塞；可选升级）
- 前端行尾警告：dist/wailsjs 文件 LF→CRLF，git 会提示但不影响构建
- 配置迁移已实现自动回写；用户旧配置（kimi/xfyun/opencode-go）会在首次启动时迁移——已由测试覆盖，但用户真实配置值得冒烟确认
- 余额型（DeepSeek）无"用量百分比"，恒绿显示；取首个非零余额币种（CNY→¥ / USD→$），全部为 0 才报错

---

## 本次会话文件变更（含会话前遗留，未提交）

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

## Wiki

- **位置**：`docs/wiki/`（已初始化 2026-08-04，本次升级后同步）
- **文件数**：11 个（00-agent-rules ~ 10-build-test-baseline，99 归档页预留）
- **入口**：先读 `docs/wiki/00-agent-rules.md`（含索引与 wiki-meta 同步状态）
- **缓存**：`docs/wiki/.covered-files`（46 项）
- **未同步文件**：04-window-positioning、06-systray、10-build-test-baseline（本次未改动相关代码）

---

## 关键文件索引

- `app.go` - Wails 主应用（窗口定位 fitToScreen、动态 Provider 编排、配置）
- `main.go` - 入口 + Wails 运行时配置（frameless / AlwaysOnTop）
- `workarea_windows.go` / `workarea_other.go` - Win32 工作区查询辅助（非 Windows 桩）
- `internal/fetcher/registry.go` - **Provider 注册表（新增 Provider 的唯一入口）**
- `internal/fetcher/types.go` - Fetcher 接口 / QuotaResult / Kind 常量
- `internal/fetcher/{kimi,xfyun,opencode_go,mimo,deepseek}.go` - 五平台抓取器（均有测试）
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
