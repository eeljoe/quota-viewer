# Quota Viewer

**[English](README.md)** | [简体中文](README.zh-CN.md)

A desktop floating ball + AI platform quota monitoring tool: displays API quota / balance remaining for various AI platforms in real time.

![Floating Ball](docs/screenshots/preview-1.png)
![Expanded Panel](docs/screenshots/preview-2.png)

**10 platforms**: [Kimi](https://platform.kimi.com) · [Xfyun](https://maas.xfyun.cn) · [OpenCode Go](https://opencode.ai/zen) · [Xiaomi MiMo](https://platform.xiaomimimo.com) · [DeepSeek](https://www.deepseek.com/platform) · [Ollama](https://ollama.com/settings) · [Command Code](https://commandcode.ai) · [Factory Droid](https://factory.ai) · [MiniMax](https://platform.minimaxi.com) · [Open Design](https://open-design.ai)

> The registry supports 10 platforms, but the floating ball shows **at most 9 cells at once** (3 normally, 9 in extended mode).

- **Tech Stack**: Go 1.24 + Wails v2.12.0 + Vanilla HTML/CSS/JS (Vite bundling)
- **Platform**: Windows 10+ (WebView2 Runtime)
- **Footprint**: ~10.9 MB single executable, ~60 MB working set while running

## Features

- **Floating Ball**: 60×60 glassmorphic square showing enabled providers (cell color = status: green / yellow / red), hover tooltip for per-platform details
- **Adaptive Cells**: one to three cells in a single row (1 fills the whole ball, 2 split it in half, 3 in thirds); in extended mode 4 cells use a 2×2 grid and 5-9 cells wrap in 3 columns
- **Expandable Panel**: Progress bars + remaining quota details (thousands-separated numbers), auto-refresh (default 15 min), refresh countdown timer (e.g. Kimi 5-hour window reset)
- **Budget Progress Bar**: Balance-type providers (e.g. DeepSeek) support a configurable budget — progress bar shows consumption as `(budget - balance) / budget`, default budget ¥300
- **Config Panel**: Check which providers to display (**up to 3 normally, up to 9 with extended mode on; at least 1**), each provider has its own credential inputs (API Key / Cookie), supports pasting browser "Copy as PowerShell" format directly
- **System Tray**: Refresh / Show-Hide / Open Config / Quit; click the tray icon to bring the floating ball to the front (show/hide lives in the tray menu; double-click no longer toggles it)
- **Window Positioning**: Expand without going off-screen (smart corner flipping, multi-monitor / DPI scaling adaptive), collapse precisely back to original position
- Close to tray, does not exit the process

## Supported Providers

| Provider | Credential | Type | Displays |
|----------|------------|------|----------|
| [Kimi](https://platform.kimi.com) | API Key (`sk-kimi-xxx`) | usage | 5-hour window / weekly quota usage |
| [Xfyun](https://maas.xfyun.cn) (讯飞星辰) | Cookie | usage | Package total usage |
| [OpenCode Go](https://opencode.ai/zen) | API Key (`oc_sk-...`, or auto-read `~/.local/share/opencode/auth.json` if empty) | usage | Rolling window quota percentage |
| [Xiaomi MiMo](https://platform.xiaomimimo.com) | Cookie | usage | Package token usage |
| [DeepSeek](https://www.deepseek.com/platform) | API Key (`sk-...`) | balance | Account balance (balance type) |
| [Ollama](https://ollama.com/settings) | Browser session Cookie | usage | Cloud 5-hour Session usage / weekly usage (settings page) |
| [Command Code](https://commandcode.ai) | API Key (`user_...`, or auto-read `~/.commandcode/auth.json` if empty) | usage | 5-hour / weekly rolling window usage + remaining credits |
| [Factory Droid](https://factory.ai) | API Key (`fk-...`, or auto-read `~/.factory/.env` if empty) | usage | 5-hour / weekly / monthly windows + Core pool usage |
| [MiniMax](https://platform.minimaxi.com) | Subscription Key (`sk-cp-...`) | usage | 5-hour / weekly window usage (Token Plan) |
| [Open Design](https://open-design.ai) | Browser session Cookie | usage | Design Plan 5-hour / 7-day quota usage |

> Tip: For Cookie-based providers, you can paste the entire "Copy as PowerShell" script from your browser's F12 → Network tab, and it will be automatically parsed into a Cookie header on save.

## Command-Line Probe

Besides the GUI there is a headless probe, handy for scripts, status bars and external tooling:

```bash
go build -o build/bin/quota-viewer-cli.exe ./cmd/quotaviewer
```

- `quota-viewer-cli.exe --json` — prints **one line of JSON** with every platform's quota to stdout and exits `0`
- no argument, or an unknown argument — prints usage to stderr, writes nothing to stdout, exits `2`

## Monitoring Open Design

[Open Design](https://open-design.ai) is a design-agent workspace; its quota comes from a **Design Plan** subscription, exposed as a **5-hour window and a 7-day window**.

Quota Viewer reads those windows from Open Design's control plane, and this is the one provider that **cannot** use an API key: keys of the form `sk-amr-` get a `401` from the control plane, so a **browser session Cookie is required**.

To get it: sign in at `open-design.ai/cloud/dashboard`, open F12 → Network, copy the request as a Cookie (or use "Copy as PowerShell" and paste the whole script), then paste it into the Open Design card in the config panel. When the session expires, sign in again and replace the Cookie.

## Development

```bash
# Install dependencies
go mod tidy
cd frontend && npm install && cd ..

# Development mode (hot reload)
wails dev

# Build
wails build
```

Output: `build/bin/quota-viewer.exe`

## Project Structure

```
├── app.go               # Wails app logic (window positioning, config, refresh orchestration)
├── main.go              # Entry point + Wails runtime options (frameless / AlwaysOnTop)
├── workarea_windows.go  # Win32 helpers (workarea query, DPI, min window width subclass)
├── workarea_other.go    # Non-Windows stub
├── cmd/
│   └── quotaviewer/     # Headless CLI probe entry point (--json)
├── internal/
│   ├── config/          # Config persistence (dynamic provider list + PowerShell cookie parsing)
│   ├── fetcher/         # Provider registry + per-platform fetchers
│   ├── probe/           # CLI probe implementation (shared fetcher orchestration)
│   └── tray/            # System tray (energye/systray)
└── frontend/
    ├── src/
    │   ├── index.html   # Floating ball + detail panel + config panel + Toast
    │   ├── main.js      # View state machine / dynamic ball cells / data rendering / interactions
    │   └── style.css    # Design tokens + dark glassmorphic theme
    └── dist/            # Vite build output (embedded into exe via go:embed)
```

## Adding a New Provider

Quota Viewer uses a registry-driven architecture — adding a new provider requires only a fetcher implementation + a registry entry, with **zero frontend changes**.

If you want your AI agent to add a new provider, copy and paste this to it:

> Read `docs/ADDING_A_PROVIDER.md` and follow it to add a new provider for the platform `<name>`. The endpoint is `<URL>`, authentication is `<method>`, and it should display `<what to show>`.

For manual steps, see [`docs/ADDING_A_PROVIDER.md`](docs/ADDING_A_PROVIDER.md).

## Disclaimer

This is an **unofficial** tool, not affiliated with any platform. Platform web structures and APIs may change at any time, causing fetchers to fail. Credentials are stored only in the local config file (`%APPDATA%\quota-viewer\config.json`) and are never uploaded to any server. Use at your own risk.

## License

[MIT](LICENSE)
