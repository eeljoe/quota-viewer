# 核心调用链

**When to read**: 追踪"刷新配额"从触发到渲染的完整路径，或修改刷新/事件逻辑时。

---

## 核心内容

### 刷新触发源（三条路径汇合到 `App.Refresh()`）

```
① 定时器   startAutoRefresh()          [app.go:447]  每 RefreshIntervalMin 分钟
② 托盘     tray:refresh 事件           [app.go:69]   托盘"刷新"菜单
③ 前端     手动刷新按钮 → Refresh()    [main.js]     直接调用绑定
```

### 刷新链路（一次完整周期）

```
App.Refresh()                          [app.go:101]
 └─ fetchAll()                         [app.go:404]
     └─ 遍历 cfg.Providers 中 enabled 的(普通最多 3 个 / 扩展最多 9 个,注册表顺序)
         ├─ goroutine1: registry["kimi"].Build(creds).Fetch()
         ├─ goroutine2: registry["xfyun"].Build(creds).Fetch()
         └─ goroutine3: registry["opencode-go"].Build(creds).Fetch()
     （sync.WaitGroup 并发,按启用顺序写 results[i])
     → 补 ID/Abbr,Kind 空则补 "usage";余额型 ApplyBudget(r, p.Budget) 换算消耗百分比 [budget.go]
 └─ 加锁写 a.cache
 └─ EventsEmit("quota:update", results)   → 前端
```

启用列表为空(理论上不会,前后端都钳制 ≥1)→ 空数组,球格为空。

### 前端渲染（main.js）

```
监听 window.runtime.EventsOn("quota:update")
 └─ updateBall(results)     按结果数动态重建球格:1-3 格单行(1 格放大占满/2 格各半/3 格各 1/3)
                            4 格 2×2;5-9 格 3 列网格(字号随 4-6 / 7-9 档位缩小)
                            颜色阈值:≥60% 黄 / ≥80% 红 / ≥100% 熄灭(off);余额型未设预算恒绿,
                            设了预算后按消耗百分比走同一套阈值
 └─ updatePanel(results)    展开面板进度条 + 千分位数字 + 重置时间 + 刷新倒计时
                            (面板高度随条目数伸缩:96 + n×66,封顶 640;见 main.js panelSize)
状态: ball(收起) ⇄ panel(展开)，SIZES.ball=60
```

### 配置保存链路

```
前端 保存按钮 → App.SaveConfig(providers, refreshMin, extendedMode)   [app.go:171]
  providers: [{id, enabled, creds:{字段:输入值}, budget}]
  → 后端钳制:0 个启用 → 全部启用;超过上限 → 保留前 N 个(N = 普通 3 / 扩展 9)
  → 写回 ExtendedMode
  → 空字符串 creds = 不修改(避免掩码覆盖)
  → 所有凭证值过 NormalizeCookieInput(非 PS 格式原样返回)
  → 按注册表顺序重排 → config.Save() 写 %APPDATA%/quota-viewer/config.json
```

### 连接测试链路

```
前端 [测试] → App.TestConnection(platform)   [app.go:260]
  platform ∈ 注册表 id(当前 10 个: kimi|xfyun|opencode-go|mimo|deepseek|ollama|command-code|factory-droid|minimax|open-design)
  → 按 id 查注册表 → Build(已存 creds).Fetch() → "成功: 剩余描述" 或 "失败: 错误"
```

### 窗口形态切换

```
展开: 前端 → ExpandWindow(w, h)     [app.go:306] 首展记录 savedX/Y，fitToScreen 定位
收起: 前端 → CollapseWindow()       [app.go:323] 恢复 savedX/Y（fitToScreen 转相对坐标）
```

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `app.go` | 全部编排逻辑（Refresh/fetchAll/SaveConfig/TestConnection） |
| `frontend/src/main.js` | 事件监听、视图状态机、渲染 |
| `internal/fetcher/*.go` | Fetch() 实现 + budget.go 预算换算 |
| `internal/tray/tray.go` | 托盘事件发射点 |

---

## Must NOT Change

- `quota:update` 负载 = `[]fetcher.QuotaResult`（JSON 数组），**顺序 = 启用 Provider 的注册表顺序**——前端球格与面板按数组顺序渲染，不依赖固定下标
- 定时刷新是后端驱动的（不依赖前端），前端只做展示
