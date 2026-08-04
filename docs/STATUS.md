# 📋 项目进度文档

> **⚠️ 新会话阅读指南（必读）**
>
> 本文档不长，按以下顺序阅读：
>
> 1. 阅读下方「上下文摘要（TL;DR）」快速了解项目状态
> 2. 阅读下方「项目概况」
> 3. 阅读最新一条「会话记录」章节，重点看「未完成 & 下一步」
> 4. 给用户一份简短进度汇报（3-5 句话）：当前在做什么 → 做到哪了 → 下一步是什么 → 有无阻塞项
> 5. **严禁在用户给出指示之前修改任何文件或代码**
>
> 📍 最新章节位置：`## 会话记录：2026-08-04 14:10`（文档末尾）
> 📊 累计会话数：2 次

---

## 最后更新

<!-- git-meta: {"last_commit": "2276ae8", "branch": "master", "dirty": false, "timestamp": "2026-08-04T14:30:00Z"} -->

- **日期**：2026-08-04 14:10
- **会话摘要**：通用化升级完成——动态 Provider 配置（1-3 个可勾选）、恢复 MiMo、新增 DeepSeek、动态球格、开源准备（README/LICENSE）；测试全绿、exe 构建并冒烟通过，**工作区改动未提交**

---

## 上下文摘要（TL;DR）

- 项目：Quota Viewer，桌面悬浮球 + AI 平台额度监控工具，Go + Wails v2.12.0 + 原生 HTML/CSS/JS（Vite）
- 当前阶段：**通用化升级代码完成**——五平台注册表（Kimi/讯飞/OpenCode Go/MiMo/DeepSeek）、动态配置（最多 3 最少 1）、旧配置自动迁移、动态球格；`go test ./...` 全绿，exe 启动冒烟通过
- **当前唯一目标**：用户冒烟验证 UI 效果 → 确认提交策略（拆 commit）
- 下一步：提交全部改动（建议按提交序列拆分）；真实凭据冒烟；开源发布前再核对 README
- 注意事项：工作区 dirty（会话前遗留 15 项 + 本次升级约 30 项变更），**提交前先与用户确认拆分策略**

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
- **新增 DeepSeek**：`deepseek.go` + 测试（`GET https://api.deepseek.com/user/balance`，余额型）
- **配置模型 v2**：`Config.Providers []ProviderConfig` 动态结构；旧扁平格式（含 mimo_cookie）Load 时自动迁移并回写；默认启用 kimi/xfyun/opencode-go；钳制最多 3 个
- **app.go 动态编排**：`SaveConfig([]ProviderInput, int)` 新契约、GetConfig 返回全部 Provider 元数据（fields/掩码 creds/login_url）、fetchAll 按启用列表 1-3 并发、TestConnection 走注册表
- **前端**：球格动态重建（1 个占满放大/2 个各半/3 个各 1/3）、配置面板按注册表元数据动态生成（勾选 1-3 限制 + 测试 + 打开登录页）、余额型恒绿
- **构建验证**：`go test ./...` 全绿；`npm run build` 成功；`wails build` 成功（绑定含 ProviderInput）；exe 启动冒烟通过（WebView2 正常，无崩溃）
- **开源准备**：README 重写（支持平台表/新增 Provider 指南/免责声明）、LICENSE（MIT）、git 历史凭证扫描 0 命中
- **wiki 同步**：00/01/02/03/05/07/08/09 已更新至五平台动态模型，`.covered-files` 重新生成（46 项）

### 🔄 进行中
- 工作区大量未提交变更（会话前遗留 15 项 + 本次升级改动），待用户确认拆分提交策略

### 📋 待办
- 用户真实凭据冒烟（勾选 1/2/3 个 Provider、球格布局、DeepSeek 余额显示）
- 提交全部改动（建议序列见「未完成 & 下一步」）
- 开源发布前：确认 LICENSE 名称/作者、检查 build/ 资源（appicon 等）、可选 GitHub 初始化

---

## 未完成 & 下一步

1. **用户冒烟**：运行 `build/bin/quota-viewer.exe` 验证 UI（球格 1/2/3 布局、配置面板勾选限制、各平台测试按钮、DeepSeek 余额）
2. **提交**（遵循小步提交 + Conventional Commits，建议序列）：
   - `feat: OpenCode Go 替换 MiMo 抓取器`（会话前遗留，已确认工作正常）
   - `fix: 修复多显示器下窗口定位坐标换算`（会话前遗留，fitToScreen）
   - `feat: 引入 fetcher 注册表并新增 DeepSeek 余额抓取`（registry + QuotaResult 扩展 + deepseek + mimo 恢复?——mimo 恢复属注册表引入，按实现顺序归入此或上一步，提交时再定）
   - `refactor: 配置模型改为动态 Provider 列表并支持旧配置迁移`
   - `refactor: app.go 动态编排 Provider 抓取与配置契约`
   - `feat: 前端动态球格与 Provider 配置面板`
   - `docs: 重写 README 并添加 MIT 许可证`
   - `docs: 同步 wiki 与 STATUS`
   - ⚠️ 注意：mimo.go 曾在 HEAD 被删除，本次恢复；会话前遗留的 mimo.go 删除已 staged，提交时需 `git restore --staged` 处理
3. **开源发布**：git remote 初始化（可选）、GitHub 仓库创建、发布说明

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
