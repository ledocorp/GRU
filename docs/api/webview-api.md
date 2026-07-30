# WebView API

`WebViewPanel` is a **`Node`** that reserves layout space for an embedded web surface. On Windows with the `webview2` build tag, a live WebView2 HWND compositor fills that region; elsewhere a placeholder frame keeps layout and demos runnable.

**Production guide:** [../WEBVIEW.md](../WEBVIEW.md)  
**Sample:** `cmd/webviewhello` + `samples/webviewhello`  
**Architecture:** [../architecture/webview.md](../architecture/webview.md)

---

## Build tags

| Target | Command |
|--------|---------|
| Stub host (all platforms) | `go run ./cmd/webviewhello` |
| Live WebView2 (Windows) | `go run -tags webview2 ./cmd/webviewhello` |
| grudemo WebView scenes | `go run -tags "grudemo,webview2" .` |

Requires WebView2 runtime on Windows for the live host. The stub draws a bordered placeholder so CI and Linux builds still compile.

Scaffold:

```bash
go run ./cmd/gru new myapp --webview
```

---

## Create a panel

```go
wv := ui.NewWebViewPanel("main-web", "https://example.com", 0, 0, 0, 400)
wv.OnMessage = func(payload string) {
    // JSON bridge from page
}
parent.AddChild(wv)
```

| Field | Role |
|-------|------|
| `URL` / `HTML` | Navigation source |
| `HostStatus` | `*Signal[string]` — stub vs live host state |
| `OnMessage` | Page → Go bridge callback |
| `FillClient` | Edge-to-edge below title bar (Wails-shaped shells) |
| `AllowHTTP` / `AllowFileBundle` | Policy overrides per panel |

Pass `w=0, h=0` for flex sizing in flex parents; set `PreferredWidth` when using explicit width.

---

## Main loop sync

WebView is a **separate HWND** — not drawn inside `doc.Root.Draw()` like raylib widgets. Each frame after layout:

```go
ui.SyncWebViewHosts(doc)
```

Walks the tree, positions the OS host to each panel’s bounds, and handles resize/z-order. Call after bounds are final for the frame (post-`Layout`). On Windows live builds, pair with borderless chrome sync from [mount-chrome.md](mount-chrome.md).

Fill-client layouts also use:

```go
ui.SyncBorderlessChromeFrame(titleBar)
// … titleBar.Update, input frame …
```

See [../WEBVIEW.md](../WEBVIEW.md) for drag deferral, focus handoff, and resize pitfalls.

---

## Bridge (Go ↔ page)

```go
// Go → page (gru.bridge v1 envelope)
_ = wv.EmitBridge("toast", map[string]string{"message": "Saved"})

// Go → page (raw JSON)
wv.PostMessage(`{"type":"ping"}`)

// Page → Go
wv.OnMessage = func(payload string) { /* parse JSON */ }

caps := wv.BridgeCapabilities() // exposed to JS as gru.capabilities
```

Load **`assets/web/gru.js`** in your page for the bridge protocol. FillClient panels may expose `windowChromeResize` in capabilities.

---

## Focus and input

WebView keyboard focus is OS-managed on the HWND. `doc.SetFocus` on a native widget releases web keyboard capture. Demos **WebView Focus Handoff** (grudemo) show `TextInput` ↔ web coordination.

Pointer hit-testing runs through the retained tree; live hosts receive input when the cursor is over synced bounds. Do not assume raylib draw order equals HWND z-order — follow the sync order in WEBVIEW.md.

---

## Policy

Global defaults come from `ui.WebViewHostPolicyCurrent()`. Per-panel overrides:

```go
devTools := true
wv.DevToolsEnabled = &devTools
```

Production apps should tighten HTTP/file/devtools flags. Details in [../WEBVIEW.md](../WEBVIEW.md).

---

## Placeholder vs live

When the stub host is active, `Draw()` renders a panel frame so flex layout and grudemo scenes work without WebView2. `HostStatus` reports the current mode — useful for status labels in samples.

---

## Next steps

- [../WEBVIEW.md](../WEBVIEW.md) — authoritative host rules and checklist
- [mount-chrome.md](mount-chrome.md) — `FillClient` + title bar spine
- [../widgets/webview.md](../widgets/webview.md) — grudemo WebView scenes
- [document.md](document.md) — `SetPlatformWindowHandle` when embedding in custom platform loops
