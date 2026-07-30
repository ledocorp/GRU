# WebView embed

**Purpose:** `WebViewPanel` reserves layout space for an embedded web surface (Edge WebView2 on Windows). Gru draws native OpenGL UI in the same window; the browser is a separate child `HWND` composited above the GL blit.

**Full guide:** [../WEBVIEW.md](../WEBVIEW.md) — read that document for layout rules, title-bar gestures, main-loop order, and anti-patterns.

**Related:** [shells.md](shells.md) · [../architecture/webview.md](../architecture/webview.md) · [../GETTING_STARTED.md](../GETTING_STARTED.md)

---

## Key type

```go
wv := ui.NewWebViewPanel("panel", "https://example.com", 0, 0, 0, 400)
wv.SetFlexGrow(1)
wv.OnMessage = func(payload string) { /* JS bridge */ }
shell.AddChild(wv)

// Full-client web app (Wails-shaped):
wv.FillClient = true
```

| Constructor | `NewWebViewPanel(id, url string, x, y, w, h float32) *WebViewPanel` |

| Field / method | Role |
|----------------|------|
| `URL`, `HTML` | Navigation source |
| `FillClient bool` | Edge-to-edge below title bar |
| `HostStatus *Signal[string]` | Stub vs live host state |
| `OnMessage func(string)` | JS → Go bridge |
| `PostMessage(json string)` | Go → JS envelope |
| `BridgeCapabilities()` | Flags exposed to `gru.js` |

Without `-tags webview2`, the panel draws a **placeholder frame** so layout works on all platforms.

---

## Build & run

```bash
go run -tags webview2 ./cmd/webviewhello     # standalone FillClient sample
go run -tags "grudemo,webview2" .            # live WebView demo scenes
go run ./cmd/gru new myapp --webview         # scaffold
```

Requires **WebView2 Evergreen Runtime** on Windows.

---

## Main-loop integration (summary)

See [../WEBVIEW.md](../WEBVIEW.md) for the full sequence:

```text
titleBar.Update
Layout (when dirty)
SyncWebViewHosts(doc)
RouteScenePointerFocus(doc)
Draw / EndDrawing
PresentWebViewHosts()
```

Apps with title bar + web must refresh chrome drag flags every frame after `titleBar.Update` (`IsDragging() || IsTitleClickPending()`).

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| Nested WebView module panel | **WebView Module Demo** |
| Editor ↔ web focus routing | **WebView Focus Handoff** |
| Minimal FillClient shell | `go run -tags webview2 ./cmd/webviewhello` |

---

## Pitfalls (short list)

| Anti-pattern | Why |
|--------------|-----|
| WebView inside page-scroll viewports | HWND does not scroll with GL scissor |
| Hiding WebView during title-bar drag | Blank web until click |
| `WS_EX_TRANSPARENT` passthrough | GL shows through |
| Nesting casually in Panel title+body stacks | Use direct flex child or FillClient recipe |

Full table: [../WEBVIEW.md](../WEBVIEW.md#anti-patterns-do-not-reintroduce).

---

## Composition notes

1. **Two compositors** — layout widgets do not clip the live browser; only HWND bounds sync does.
2. **FillClient vs nested:** `FillClient=true` for web-first apps; nested panel embeds for module-style UI (**WebView Module Demo**).
3. **Focus handoff:** **WebView Focus Handoff** shows pointer routing between native text editor and web — copy `RouteScenePointerFocus` patterns from the demo host.
4. **Shell choice:** app-shell or flex column with `WebViewPanel` as **`SetFlexGrow(1)`** direct child.
5. **Debug:** `$env:GORY_WEBVIEW_DEBUG=1` + Shift+F11 — [../WEBVIEW.md](../WEBVIEW.md#debug).
