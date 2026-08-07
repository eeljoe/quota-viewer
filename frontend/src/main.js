// === Wails 绑定 ===
// window.go.main.App 在运行时由 Wails 注入

// 各视图窗口尺寸,与 Go 侧 ballSize 常量保持一致
const SIZES = {
    ball: [60, 60],
    panel: [340, 310],
    settings: [340, 480],
};

let currentView = "ball"; // ball | panel | settings
let currentResults = [];
let providerCards = []; // [{id, enabled: checkbox, fields: [{key, input}]}]

// === 视图切换(统一入口,负责窗口尺寸与屏幕内定位) ===
function setView(view) {
    currentView = view;
    document.getElementById("ball").classList.toggle("hidden", view !== "ball");
    document.getElementById("panel").classList.toggle("hidden", view !== "panel");
    document.getElementById("settings").classList.toggle("hidden", view !== "settings");

    if (view === "ball") {
        window.go.main.App.CollapseWindow();
    } else {
        const [w, h] = SIZES[view];
        window.go.main.App.ExpandWindow(w, h);
    }
}

// === 事件监听 ===
window.runtime.EventsOn("quota:update", (results) => {
    currentResults = results;
    renderResults(results);
});

// === 球面点击:展开/收起 ===
document.getElementById("ball").addEventListener("click", () => {
    if (currentView !== "ball") return;
    setView("panel");
    refreshIfNeeded(); // 展开时若数据超过 3 分钟则刷新
});

// 收起按钮
document.getElementById("btn-collapse").addEventListener("click", () => {
    setView("ball");
});

// === 刷新 ===
document.getElementById("btn-refresh").addEventListener("click", () => {
    refreshQuota();
});

async function refreshQuota() {
    const btn = document.getElementById("btn-refresh");
    btn.disabled = true;
    btn.classList.add("spinning");
    document.getElementById("last-updated").textContent = "刷新中...";
    try {
        const results = await window.go.main.App.Refresh();
        currentResults = results;
        renderResults(results);
    } catch (e) {
        console.error("refresh error:", e);
        toast("刷新失败: " + e, "error");
    } finally {
        btn.disabled = false;
        btn.classList.remove("spinning");
    }
}

let lastRefreshTime = 0;
async function refreshIfNeeded() {
    if (Date.now() - lastRefreshTime > 3 * 60 * 1000) {
        await refreshQuota();
    }
}

// === 渲染结果 ===
function renderResults(results) {
    // 更新详情面板
    const list = document.getElementById("quota-list");
    list.innerHTML = "";
    results.forEach((r, idx) => {
        const color = getStatusColor(r);
        const percent = r.percent || 0;
        const item = document.createElement("div");
        item.className = "quota-item" + (r.error ? " error" : "");
        item.style.animationDelay = idx * 45 + "ms";
        const resetHtml = (r.reset_at && !r.error)
            ? `<div class="quota-reset" data-reset-at="${r.reset_at}">${formatCountdown(r.reset_at)}</div>`
            : "";
        item.innerHTML = `
            <div class="quota-item-header">
                <span class="quota-platform"><i class="status-dot ${color}"></i>${r.platform}</span>
                <span class="quota-remaining">${r.error || r.remaining || "-"}</span>
            </div>
            <div class="progress-bar">
                <div class="progress-fill ${color}" style="width: ${r.error ? 100 : percent}%"></div>
            </div>
            ${resetHtml}
        `;
        list.appendChild(item);
    });

    // 更新球面格子
    updateBall(results);

    // 更新时间
    const now = new Date();
    lastRefreshTime = now.getTime();
    document.getElementById("last-updated").textContent = "更新于 " + now.toLocaleTimeString("zh-CN");
}

// === 倒计时 ===
function formatCountdown(isoStr) {
    const target = new Date(isoStr);
    if (isNaN(target.getTime())) return "";
    const diff = target - Date.now();
    if (diff <= 0) return "已过期";
    const hours = Math.floor(diff / 3600000);
    const mins = Math.floor((diff % 3600000) / 60000);
    if (hours > 0) return `距下次刷新: ${hours}时${mins}分`;
    return `距下次刷新: ${mins}分`;
}

function updateCountdowns() {
    document.querySelectorAll(".quota-reset").forEach((el) => {
        const resetAt = el.getAttribute("data-reset-at");
        const text = formatCountdown(resetAt);
        if (text) el.textContent = text;
    });
}

setInterval(updateCountdowns, 30000);

function getStatusColor(r) {
    if (r.error) return "red";
    // 余额型(如 DeepSeek):设了预算后按消耗百分比走颜色,未设预算恒绿
    if (r.kind === "balance" && r.percent > 0) {
        if (r.percent >= 90) return "red";
        if (r.percent >= 75) return "yellow";
        return "green";
    }
    if (r.kind === "balance") return "green";
    if (r.percent >= 90) return "red";
    if (r.percent >= 75) return "yellow";
    return "green";
}

// 球面格子 = 启用的 Provider 数量(1-3),flex 均分(1 个占满 / 2 个各半 / 3 个各 1/3);
// 格字颜色=状态;悬停 tooltip 显示各平台明细
function updateBall(results) {
    const ball = document.getElementById("ball");
    ball.querySelectorAll(".ball-cell").forEach((c) => c.remove());

    results.forEach((r) => {
        const cell = document.createElement("span");
        cell.className = "ball-cell " + getStatusColor(r);
        cell.textContent = r.abbr || r.platform.slice(0, 1);
        ball.appendChild(cell);
    });

    // 单格时放大字母占满整个球
    ball.classList.toggle("single-cell", results.length === 1);

    ball.title = results
        .map((r) => r.platform + ": " + (r.error || r.remaining || "未知"))
        .join("\n");
}

// === 配置面板 ===
document.getElementById("btn-settings").addEventListener("click", () => {
    setView("settings");
    loadConfig();
});

document.getElementById("btn-close-settings").addEventListener("click", () => {
    setView("ball");
});

async function loadConfig() {
    try {
        const cfg = await window.go.main.App.GetConfig();
        renderProviderList(cfg.providers || []);
        document.getElementById("input-interval").value = cfg.refresh_interval_min || 15;
    } catch (e) {
        console.error("loadConfig error:", e);
        toast("加载配置失败: " + e, "error");
    }
}

// 收集当前 UI 上的 Provider 状态(保存/测试共用)
function collectProviders() {
    return providerCards.map((c) => {
        const creds = {};
        c.fields.forEach((f) => {
            creds[f.key] = f.input.value;
        });
        const budget = c.budget ? parseFloat(c.budget.value) || 0 : 0;
        return { id: c.id, enabled: c.enabled.checked, creds, budget };
    });
}

// 渲染 Provider 卡片列表(勾选 + 凭证字段动态生成)
function renderProviderList(providers) {
    const container = document.getElementById("provider-list");
    container.innerHTML = "";
    providerCards = [];

    providers.forEach((p) => {
        const card = document.createElement("div");
        card.className = "provider-card" + (p.enabled ? "" : " disabled");
        card.dataset.id = p.id;

        // 头部:勾选框 + 名称 + 动作按钮
        const head = document.createElement("div");
        head.className = "provider-head";

        const toggle = document.createElement("label");
        toggle.className = "provider-toggle";
        const cb = document.createElement("input");
        cb.type = "checkbox";
        cb.className = "provider-check";
        cb.checked = !!p.enabled;
        const name = document.createElement("span");
        name.className = "provider-name";
        name.textContent = p.name;
        toggle.append(cb, name);

        const actions = document.createElement("div");
        actions.className = "provider-actions";
        const testBtn = document.createElement("button");
        testBtn.className = "btn-sm";
        testBtn.dataset.test = p.id;
        testBtn.textContent = "测试";
        actions.appendChild(testBtn);
        if (p.login_url) {
            const loginBtn = document.createElement("button");
            loginBtn.className = "btn-sm";
            loginBtn.dataset.open = p.login_url;
            loginBtn.textContent = "打开登录页";
            actions.appendChild(loginBtn);
        }

        head.append(toggle, actions);
        card.appendChild(head);

        // 凭证字段(按注册表元数据生成,placeholder 显示掩码值)
        const fieldsWrap = document.createElement("div");
        fieldsWrap.className = "provider-fields";
        const fields = [];
        (p.fields || []).forEach((f) => {
            const group = document.createElement("div");
            group.className = "form-group";
            const label = document.createElement("label");
            label.textContent = f.label;
            const input = document.createElement(f.type === "textarea" ? "textarea" : "input");
            if (f.type === "password") input.type = "password";
            if (f.type === "text") input.type = "text";
            if (f.type === "textarea") input.rows = 2;
            input.placeholder = (p.creds && p.creds[f.key]) || "";
            group.append(label, input);
            fieldsWrap.appendChild(group);
            fields.push({ key: f.key, input });
        });
        card.appendChild(fieldsWrap);

        // 余额型 Provider 增加预算输入框
        let budgetInput = null;
        if (p.kind === "balance") {
            const budgetGroup = document.createElement("div");
            budgetGroup.className = "form-group";
            const budgetLabel = document.createElement("label");
            budgetLabel.textContent = "预算(用于进度条计算)";
            budgetInput = document.createElement("input");
            budgetInput.type = "number";
            budgetInput.min = "0";
            budgetInput.step = "0.01";
            budgetInput.placeholder = "设为 0 则不计算进度条";
            if (p.budget > 0) budgetInput.value = p.budget;
            budgetGroup.append(budgetLabel, budgetInput);
            fieldsWrap.appendChild(budgetGroup);
        }

        // 勾选限制:最多 3 个、最少 1 个
        cb.addEventListener("change", () => {
            const enabledCount = providerCards.filter((c) => c.enabled.checked).length;
            if (enabledCount > 3) {
                cb.checked = false;
                toast("最多展示 3 个 Provider", "error");
                return;
            }
            if (enabledCount < 1) {
                cb.checked = true;
                toast("至少保留 1 个 Provider", "error");
                return;
            }
            card.classList.toggle("disabled", !cb.checked);
        });

        container.appendChild(card);
        providerCards.push({ id: p.id, enabled: cb, fields, budget: budgetInput });
    });
}

document.getElementById("btn-save-config").addEventListener("click", async () => {
    const providers = collectProviders();
    const interval = parseInt(document.getElementById("input-interval").value) || 15;
    try {
        await window.go.main.App.SaveConfig(providers, interval);
        // 清空输入框(已保存)
        providerCards.forEach((c) => c.fields.forEach((f) => {
            f.input.value = "";
        }));
        toast("已保存", "success");
        await loadConfig(); // 重新拉取(placeholder 显示新掩码)
        await refreshQuota(); // 立即刷新展示
    } catch (e) {
        console.error("saveConfig error:", e);
        toast("保存配置失败: " + e, "error");
    }
});

// 测试/打开登录页按钮(事件委托,兼容动态生成的按钮)
document.addEventListener("click", async (e) => {
    const testBtn = e.target.closest("[data-test]");
    if (testBtn) {
        const platform = testBtn.getAttribute("data-test");
        // 先保存当前输入(刷新间隔不动),再测试
        try {
            await window.go.main.App.SaveConfig(collectProviders(), 0);
            const result = await window.go.main.App.TestConnection(platform);
            toast(result, result.startsWith("成功") ? "success" : "error");
        } catch (err) {
            console.error("testConnection error:", err);
            toast("测试连接失败: " + err, "error");
        }
        return;
    }
    const openBtn = e.target.closest("[data-open]");
    if (openBtn) {
        const url = openBtn.getAttribute("data-open");
        try {
            window.go.main.App.OpenLoginPage(url);
        } catch (err) {
            console.error("openLoginPage error:", err);
            toast("打开登录页失败: " + err, "error");
        }
    }
});

// === Toast(替代 alert) ===
let toastTimer = null;
function toast(msg, type) {
    const el = document.getElementById("toast");
    el.textContent = msg;
    el.className = "toast show " + (type || "");
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => {
        el.classList.remove("show");
    }, 2500);
}

// === 球位置记忆(拖动结束时保存)===
let dragTimer = null;
document.getElementById("ball").addEventListener("mouseup", () => {
    clearTimeout(dragTimer);
    dragTimer = setTimeout(() => {
        // Wails 获取窗口位置
        window.runtime.WindowGetPosition().then((pos) => {
            window.go.main.App.SaveBallPosition(pos.x, pos.y);
        });
    }, 500);
});

// === 启动:加载初始数据 ===
window.go.main.App.Refresh();

// === 托盘事件 ===
window.runtime.EventsOn("tray:refresh", () => {
    refreshQuota();
});

window.runtime.EventsOn("ui:show-settings", () => {
    setView("settings");
    loadConfig();
});
