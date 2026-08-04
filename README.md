# Quota Viewer

**[English](README.md)** | [简体中文](README.zh-CN.md)

A desktop floating ball + AI platform quota monitoring tool: displays API quota / balance remaining for various AI platforms in real time.

- **Tech Stack**: Go + Wails v2.12.0 + Vanilla HTML/CSS/JS (Vite bundling)
- **Platform**: Windows 10+ (WebView2 Runtime)

## Features

- **Floating Ball**: 60×60 glassmorphic square, 1-3 cells showing enabled providers (cell color = status: green / yellow / red), hover tooltip for per-platform details
- **Adaptive Cells**: Enable 1 provider → letter fills the entire ball; 2 → each takes 1/2; 3 → each takes 1/3
- **Expandable Panel**: Progress bars + remaining quota details (thousands-separated numbers), auto-refresh (default 15 min)
- **Config Panel**: Check which providers to display (**max 3, min 1**), each provider has its own credential inputs (API Key / Cookie), supports pasting browser "Copy as PowerShell" format directly
- **System Tray**: Refresh / Show-Hide / Open Config / Quit; double-click tray icon to toggle visibility
- **Window Positioning**: Expand without going off-screen (smart corner flipping, multi-monitor / DPI scaling adaptive), collapse precisely back to original position
- Close to tray, does not exit the process

## Supported Providers

| Provider | Credential | Displays |
|----------|------------|----------|
| Kimi | API Key (`sk-kimi-xxx`) | 5-hour window / weekly quota usage |
| Xfyun (讯飞星辰) | Cookie | Package total usage |
| OpenCode Go | Workspace ID + Session Token | Rolling window quota percentage |
| Xiaomi MiMo | Cookie | Package token usage |
| DeepSeek | API Key (`sk-...`) | Account balance (balance type) |

> Tip: For Cookie-based providers, you can paste the entire "Copy as PowerShell" script from your browser's F12 → Network tab, and it will be automatically parsed into a Cookie header on save.

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
├── internal/
│   ├── config/          # Config persistence (dynamic provider list + PowerShell cookie parsing)
│   ├── fetcher/         # Provider registry + per-platform fetchers
│   └── tray/            # System tray (energye/systray)
└── frontend/
    ├── src/
    │   ├── index.html   # Floating ball + detail panel + config panel + Toast
    │   ├── main.js      # View state machine / dynamic ball cells / data rendering / interactions
    │   └── style.css    # Design tokens + dark glassmorphic theme
    └── dist/            # Vite build output (embedded into exe via go:embed)
```

## Adding a New Provider

1. Add a new fetcher in `internal/fetcher/` (implement the `Fetcher` interface, `Fetch() QuotaResult`), with httptest tests
2. Register it in `internal/fetcher/registry.go`: ID, display name, ball abbreviation, credential field definitions, `Build` factory
3. The frontend config panel and ball cells adapt automatically — no UI code changes needed

## Disclaimer

This is an **unofficial** tool, not affiliated with any platform. Platform web structures and APIs may change at any time, causing fetchers to fail. Credentials are stored only in the local config file (`%APPDATA%/quota-viewer/config.json`) and are never uploaded to any server. Use at your own risk.

## License

[MIT](LICENSE)
