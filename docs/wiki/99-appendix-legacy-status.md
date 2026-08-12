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
