// === Wails 绑定 ===
// window.go.main.App 在运行时由 Wails 注入

// 各视图窗口尺寸,与 Go 侧 ballSize 常量保持一致
const SIZES = {
    ball: [60, 60],
    panel: [340, 260],
    settings: [340, 480],
};

let currentView = "ball"; // ball | panel | settings
let currentResults = [];

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
        item.innerHTML = `
            <div class="quota-item-header">
                <span class="quota-platform"><i class="status-dot ${color}"></i>${r.platform}</span>
                <span class="quota-remaining">${r.error || r.remaining || "-"}</span>
            </div>
            <div class="progress-bar">
                <div class="progress-fill ${color}" style="width: ${r.error ? 100 : percent}%"></div>
            </div>
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

function getStatusColor(r) {
    if (r.error) return "red";
    if (r.percent >= 90) return "red";
    if (r.percent >= 75) return "yellow";
    return "green";
}

const BALL_CELL_IDS = ["cell-kimi", "cell-xfyun", "cell-mimo"];

// 球面三格对应三个平台,格字颜色=状态;悬停 tooltip 显示各平台明细
function updateBall(results) {
    results.forEach((r, i) => {
        if (i > 2) return;
        const cell = document.getElementById(BALL_CELL_IDS[i]);
        if (cell) cell.className = "ball-cell " + getStatusColor(r);
    });
    document.getElementById("ball").title = results
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
        document.getElementById("input-kimi").placeholder = cfg.kimi_api_key || "sk-kimi-xxx";
        document.getElementById("input-xfyun").placeholder = cfg.xfyun_cookie || "从浏览器 F12 复制 Cookie";
        document.getElementById("input-mimo").placeholder = cfg.mimo_cookie || "从浏览器 F12 复制 Cookie";
        document.getElementById("input-interval").value = cfg.refresh_interval_min || 15;
    } catch (e) {
        console.error("loadConfig error:", e);
        toast("加载配置失败: " + e, "error");
    }
}

document.getElementById("btn-save-config").addEventListener("click", async () => {
    const kimi = document.getElementById("input-kimi").value;
    const xfyun = document.getElementById("input-xfyun").value;
    const mimo = document.getElementById("input-mimo").value;
    const interval = parseInt(document.getElementById("input-interval").value) || 15;
    try {
        await window.go.main.App.SaveConfig(kimi, xfyun, mimo, interval);
        // 清空输入框(已保存)
        document.getElementById("input-kimi").value = "";
        document.getElementById("input-xfyun").value = "";
        document.getElementById("input-mimo").value = "";
        toast("已保存", "success");
    } catch (e) {
        console.error("saveConfig error:", e);
        toast("保存配置失败: " + e, "error");
    }
});

// 测试连接按钮
document.querySelectorAll("[data-test]").forEach((btn) => {
    btn.addEventListener("click", async () => {
        const platform = btn.getAttribute("data-test");
        // 先保存当前输入(如果有)
        const kimi = document.getElementById("input-kimi").value;
        const xfyun = document.getElementById("input-xfyun").value;
        const mimo = document.getElementById("input-mimo").value;
        try {
            if (kimi || xfyun || mimo) {
                await window.go.main.App.SaveConfig(kimi, xfyun, mimo, 0);
            }
            const result = await window.go.main.App.TestConnection(platform);
            toast(result, result.startsWith("成功") ? "success" : "error");
        } catch (e) {
            console.error("testConnection error:", e);
            toast("测试连接失败: " + e, "error");
        }
    });
});

// 打开登录页按钮
document.querySelectorAll("[data-open]").forEach((btn) => {
    btn.addEventListener("click", () => {
        const url = btn.getAttribute("data-open");
        try {
            window.go.main.App.OpenLoginPage(url);
        } catch (e) {
            console.error("openLoginPage error:", e);
            toast("打开登录页失败: " + e, "error");
        }
    });
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
