# WebView on Windows (public guide)

Embed Edge WebView2 beside (or under) Gru’s OpenGL UI. Live host requires **`-tags webview2`** and the **WebView2 Evergreen Runtime**.

**Starters:** `go run -tags webview2 ./cmd/webviewhello` · `gru new myapp --webview`  
**Demos:** WebView Module Demo, WebView Focus Handoff — `go run -tags "grudemo,webview2" .`

---

## Two compositor layers

| Layer | Technology | Notes |
|-------|------------|-------|
| Gru UI | OpenGL SSAA blit | Drawn in the Raylib frame |
| WebView2 | Child `HWND` | Independent renderer; raised after `EndDrawing` |

The HWND sits **above** the OpenGL blit. Layout widgets do **not** scissor the live browser — only HWND bounds sync does.

---

## Layout rules

| Need | Do | Don’t |
|------|----|--------|
| Fill-height embed | App shell + flex-grow row; `WebViewPanel` as a **direct** flex child | Put live WebView inside page-scroll viewports |
| Full-client web (Wails-shaped) | `WebViewPanel` with `FillClient=true` below the title bar | Invent OS mouse / passthrough tricks |
| Scrollable forms page | Native widgets only (or no live WebView in the scroll band) | Nest WebView in `Panel` title + body stacks casually |

Clamp HWND top to the document chrome (`ChromeTop`) — never underlap the 48px title band.

---

## Title bar + WebView (gestures)

Title bar stays **native** (Raylib). While the cursor is in the title band, a drag is active, or a title click is pending, the host **defers HWND raise** but stays **visible** (hiding caused blank web).

| Window state | First press | Drag-out | Double-click |
|--------------|-------------|----------|--------------|
| Restored | Arm grab immediately + pending | Already dragging | Maximize |
| Maximized | Pending only (no immediate restore) | ≥4px → restore then drag | Native restore |

Apps must refresh chrome drag state every frame after `titleBar.Update` (`IsDragging() || IsTitleClickPending()`).

**Rejected:** `WS_EX_TRANSPARENT` while visible; TitleBar OS mouse polling over the WebView; raising every orphan Chrome child HWND each frame.

---

## FillClient edge resize

When `FillClient=true`, the flush HWND covers Raylib edge grips. The **page** owns S / E / W / SE / SW hit targets (via `assets/web/gru.js` when `windowChromeResize` is enabled) and posts `chrome.resize` on the bridge; Go applies window size/position like the title bar.

- **Scope:** FillClient only (`cmd/webviewhello` and apps from `gru new --webview`).  
- **Not for:** nested embeds / Focus Handoff side-by-side.  
- N / NE / NW remain title-bar native.

---

## Main-loop order (do not reorder casually)

```text
titleBar.Update → chrome drag / pending flags
Layout (when dirty)
SyncWebViewHosts(doc)     ← bounds + visibility
RouteScenePointerFocus    ← editor ↔ web clicks
Draw / EndDrawing
PresentWebViewHosts()     ← Show + bounds-scoped raise
```

On client resize: mark WebView hosts for force bounds sync; keep last COM bounds through transient 0×0 layout so the page does not go blank.

---

## Modals and overlays over WebView (occlusion)

The live WebView2 HWND sits **above** the OpenGL blit. A native modal drawn in Raylib cannot paint on top of that HWND. Gru therefore **hides** live hosts while a blocking overlay is open.

| Overlay | Occludes WebView? |
|---------|-------------------|
| Modal (`ShowModal`) | Yes |
| Drawer / bottom sheet | Yes |
| Context menu, toast, title bar | No (native layers beside the panel) |

**Engine rule:** `WebViewHostOccluded()` is true when a modal, drawer, or bottom sheet is visible. Host sync then calls `SetVisible(false)` on live hosts until the overlay closes. You do not call this yourself when using `ShowModal` / standard drawers.

**What you should do**

1. Use `ui.ShowModal(...)` (or the drawer APIs) for blocking UI over a scene that contains `WebViewPanel`.
2. Keep the modal body **flat flex** (no `Viewport` as modal root). Same composition rule as native-only apps.
3. Verify with **WebView Focus Handoff**: button “Open modal (hides web host)” — the web panel should disappear for the modal and return when closed.
4. Run with `-tags webview2` so the host is live; without the tag there is no HWND to occlude.

**What not to do**

| Don’t | Why |
|-------|-----|
| Invent click-through / `WS_EX_TRANSPARENT` so GL “shows through” the web | Blank or broken compositing |
| Manually hide the host every frame outside overlay APIs | Fights sync; easy to leave the host stuck hidden |
| Expect a Raylib-drawn dialog to appear *on top of* a visible HWND | Z-order makes that impossible without occlusion |

---

## Anti-patterns (do not reintroduce)

| Anti-pattern | Why it fails |
|--------------|--------------|
| `WS_EX_TRANSPARENT` / click-through while visible | Blank web (GL shows through) |
| Hide WebView during title-bar drag or edge resize | Blank until click / scene change |
| Raise all Chrome child HWNDs | Orphans cover the title bar |
| OS `GetAsyncKeyState` over WebView for chrome | DPI/coord bugs; fights the HWND |
| FillClient L/R/bottom HWND inset “for grips” | Visible GL frame around flush web |

---

## Debug

```bash
# PowerShell
$env:GRU_WEBVIEW_DEBUG=1   # legacy alias: GORY_WEBVIEW_DEBUG=1
go run -tags webview2 ./cmd/webviewhello
```

Shift+F11 toggles stderr traces when debug is on. F12 remains the widget inspector.

---

## Related

- [GETTING_STARTED.md](GETTING_STARTED.md) · [CLI.md](CLI.md) · [BUILD.md](BUILD.md)  
- [architecture/webview.md](architecture/webview.md) — HWND loop model  
- Demos: [DEMO_INDEX.md](DEMO_INDEX.md)
