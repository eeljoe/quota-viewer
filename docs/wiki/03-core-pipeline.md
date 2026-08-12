# 核心调用链

**When to read**: 追踪"刷新配额"从触发到渲染的完整路径，或修改刷新/事件逻辑时。

---

## 核心内容

### 刷新触发源（三条路径汇合到 `App.Refresh()`）

```
① 定时器   startAutoRefresh()          [app.go:330]  每 RefreshIntervalMin 分钟
② 托盘     tray:refresh 事件           [app.go:69]   托盘"刷新"菜单
③ 前端     手动刷新按钮 → Refresh()    [main.js]     直接调用绑定
```

### 刷新链路（一次完整周期）

```
App.Refresh()                          [app.go:101]
 └─ fetchAll()                         [app.go:303]
     └─ 遍历 cfg.Providers 中 enabled 的(最多 3 个,注册表顺序)
         ├─ goroutine1: registry["kimi"].Build(creds).Fetch()
         ├─ goroutine2: registry["xfyun"].Build(creds).Fetch()
         └─ goroutine3: registry["opencode-go"].Build(creds).Fetch()
     （sync.WaitGroup 并发,按启用顺序写 results[i])
     → 每个结果补 ID/Abbr,Kind 默认 "usage"
 └─ 加锁写 a.cache
 └─ EventsEmit("quota:update", results)   → 前端
```

启用列表为空(理论上不会,前后端都钳制 ≥1)→ 空数组,球格为空。

### 前端渲染（main.js）

```
监听 window.runtime.EventsOn("quota:update")
 └─ updateBall(results)     按结果数动态重建球格(1 格占满/2 格各半/3 格各 1/3)
                            颜色: green / yellow / red;余额型(kind=balance)有余额即绿
 └─ updatePanel(results)    展开面板进度条 + 千分位数字 + 重置时间
状态: ball(收起) ⇄ panel(展开)，SIZES.ball=60
```

### 配置保存链路

```
前端 保存按钮 → App.SaveConfig(providers, refreshMin)   [app.go]
  providers: [{id, enabled, creds:{字段:输入值}}]
  → 后端钳制:0 个启用 → 全部启用;>3 → 保留前 3
  → 空字符串 creds = 不修改(避免掩码覆盖)
  → 所有凭证值过 NormalizeCookieInput(非 PS 格式原样返回)
  → 按注册表顺序重排 → config.Save() 写 %APPDATA%/quota-viewer/config.json
```

### 连接测试链路

```
前端 [测试] → App.TestConnection(platform)   [app.go]
  platform ∈ 注册表 id: "kimi"|"xfyun"|"opencode-go"|"mimo"|"deepseek"|"ollama"
  → 按 id 查注册表 → Build(已存 creds).Fetch() → "成功: 剩余描述" 或 "失败: 错误"
```

### 窗口形态切换

```
展开: 前端 → ExpandWindow(w, h)     [app.go:206] 首展记录 savedX/Y，fitToScreen 定位
收起: 前端 → CollapseWindow()       [app.go:223] 恢复 savedX/Y（fitToScreen 转相对坐标）
```

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `app.go` | 全部编排逻辑（Refresh/fetchAll/SaveConfig/TestConnection） |
| `frontend/src/main.js` | 事件监听、视图状态机、渲染 |
| `internal/fetcher/*.go` | Fetch() 实现 |
| `internal/tray/tray.go` | 托盘事件发射点 |

---

## Must NOT Change

- `quota:update` 负载 = `[]fetcher.QuotaResult`（JSON 数组），**顺序 = 启用 Provider 的注册表顺序**——前端球格与面板按数组顺序渲染，不依赖固定下标
- 定时刷新是后端驱动的（不依赖前端），前端只做展示
