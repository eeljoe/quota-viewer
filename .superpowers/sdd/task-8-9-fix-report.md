# Task 8 & Task 9 Review Fixes

## Status: DONE

All 6 fixes applied, verified, and committed.

## Files Changed

| File | Fix(es) |
|------|---------|
| `frontend/src/main.js` | Fix 1, Fix 3, Fix 4 |
| `frontend/src/index.html` | Fix 2 |
| `internal/tray/tray.go` | Fix 6 |
| `main.go` | Fix 5 (gofmt) |

## Fix Details

### Fix 1 — QuotaResult field name case mismatch (Critical)
Wails JSON-marshals `QuotaResult` using the lowercase JSON tags in `internal/fetcher/types.go` (`platform`, `used`, `total`, `percent`, `remaining`, `reset_at`, `error`). The frontend was reading PascalCase fields, all `undefined` at runtime.

Changed every PascalCase access in `frontend/src/main.js` to lowercase:
- `r.Platform` → `r.platform` (renderResults template)
- `r.Error` → `r.error` (template condition + progress width + getDotColor)
- `r.Remaining` → `r.remaining` (template)
- `r.Percent` → `r.percent` (percent var + getDotColor thresholds)
- `r.Used`, `r.Total`, `r.ResetAt` — not referenced in main.js; grep confirmed zero PascalCase occurrences remain.

### Fix 2 — Production build 404s on main.js (Important)
`frontend/src/index.html`: `<script src="main.js">` → `<script type="module" src="/main.js"></script>`. Vite now bundles the entry into `dist/`. Still vanilla JS, no dependency added.

### Fix 3 — Unguarded Wails calls (Important)
Wrapped four call sites in try/catch with `alert()` on error, matching the existing `refreshQuota()` pattern:
- `loadConfig()` — GetConfig
- `btn-save-config` click — SaveConfig
- `data-test` handler — SaveConfig + TestConnection
- `data-open` handler — OpenLoginPage

### Fix 4 — Missing frontend event listeners (Important)
Appended to `frontend/src/main.js`:
```javascript
window.runtime.EventsOn("tray:refresh", () => { refreshQuota(); });
window.runtime.EventsOn("ui:show-settings", () => { /* hide ball+panel, show settings, loadConfig */ });
```
`tray:toggle` is handled Go-side (window show/hide), no frontend listener needed.

### Fix 5 — gofmt compliance (Important)
Ran `gofmt -w main.go`. `gofmt -l main.go` returns no output (clean). `internal/tray/tray.go` (touched by Fix 6) is also gofmt-clean.

Note: `internal/fetcher/types.go` shows a pre-existing gofmt diff (struct comment alignment), but that file is outside this task's scope (Fix 5 is scoped to `main.go` only) and was left untouched.

### Fix 6 — Double-click toggle (Critical)
In `internal/tray/tray.go`, `onReady`, added:
```go
systray.SetOnDClick(func(menu systray.IMenu) {
    wailsruntime.EventsEmit(t.ctx, "tray:toggle")
})
```
Verified against the actual `github.com/energye/systray v1.0.3` API in the module cache: `SetOnDClick(fn func(menu IMenu))` is the correct exported signature (confirmed in `systray.go:100` and `example/main.go:54`).

## Verification

| Check | Command | Result |
|-------|---------|--------|
| Build | `go build ./...` | exit 0, clean |
| Tests | `go test ./internal/... -v` | all 14 tests PASS (config: 3, fetcher: 11) |
| gofmt | `gofmt -l main.go` | no output (clean) |
| Field grep | grep for `r.(Platform|Error|Remaining|Percent|Used|Total|ResetAt)` in main.js | no matches |

## Commit
See git log for the commit SHA and subject.
