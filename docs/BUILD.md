# Public demo — build & smoke

**Audience:** public GitHub consumers.

## Build

From the **repository root** (the folder that contains `go.mod`, `assets/`, and `pages/`):

```bash
go run ./cmd/hello
go run -tags webview2 ./cmd/webviewhello
go run -tags grudemo .
go run -tags "grudemo,webview2" .
```

Produce a binary with the CLI:

```bash
go run ./cmd/gru build hello -o dist/HelloGru.exe
go run ./cmd/gru build webviewhello -webview2 -o dist/WebViewHello.exe
go run ./cmd/gru build demo -o dist/GruDemo.exe
go run ./cmd/gru build demo -webview2 -o dist/GruDemo.exe
```

Or with `go build` directly:

```bash
mkdir -p dist
go build -tags grudemo -o dist/GruDemo.exe .
```

**Smoke (automated):**

```bash
go test ./examples/ -run TestPublicAllowlist -count=1
```

All 40 allowlisted scenes must register and `Build` without panic.

| Tag | Effect |
|-----|--------|
| `grudemo` | Curated demo allowlist, F12 Inspector, no Raylib DrawFPS corner |
| `webview2` | Live WebView2 host (Windows; requires WebView2 Runtime) |

**If the window never opens:** ensure you started from the repo root and that `assets/` / `pages/` are present.

**If Windows blocks the `.exe`:** prefer `go run` above, or adjust Defender / Smart App Control for your build output folder.

WebView assets: `assets/web/`. Gallery: `pages/gallery.gru`.  
WebView behavior: [WEBVIEW.md](WEBVIEW.md). CLI details: [CLI.md](CLI.md).

## Keys

| Key | Action |
|-----|--------|
| **Tab** | Next scene |
| **F12** | Inspector (widget tree + FPS) |
| **F11** | Perf strip (optional) |
| **F9** | Borderless chrome toggle (desktop) |

## Smoke checklist

```bash
go test ./examples/ -run TestPublicAllowlist -count=1
go run -tags grudemo .
# optional:
go run -tags "grudemo,webview2" .
go run -tags webview2 ./cmd/webviewhello
```

Agent-run smoke (2026-07-29):

- [x] `TestPublicAllowlist` green
- [x] **grudemo** starts on **Counter Demo** (`grudemo: 35 public scenes (start=Counter Demo)`)
- [x] Window presents (`presenting first frame…`); ran ~2 min — Gallery + Demo Index visited; clean exit path (Windows `0xc000041d` on close is a known Raylib/GLFW teardown quirk, not a start panic)
- [x] **webviewhello** starts with live WebView2 (`Environment created successfully`; FillClient host loop banner)
- [x] Title-bar grab + double-click + FillClient edge resize — already signed off in the WebView completion session (not re-blocking the docs cut)

Remaining only if you want a fresh human feel-pass before the public push (optional):

- [ ] Footer **Scenes** / Tab / F12 feel-check
- [ ] Gallery Surface capabilities scroll feel-check
