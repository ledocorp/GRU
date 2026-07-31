# WebView architecture

Gru’s OpenGL UI and Edge WebView2 are **two compositor layers**. This chapter is the mental model and loop contract. Production rules, gestures, FillClient resize, and anti-patterns: **[../WEBVIEW.md](../WEBVIEW.md)**.

Requires **`-tags webview2`** and the WebView2 Evergreen Runtime on Windows.

---

## Two layers

| Layer | Technology | Drawn when |
|-------|------------|------------|
| Gru UI | OpenGL SSAA blit | Through `EndDrawing` |
| WebView2 | Child `HWND` | Independent renderer; raised in `PresentWebViewHosts` |

The HWND sits **above** the GL blit. `Panel` / `Viewport` scissor **do not** clip the live browser — only HWND bounds sync (`visibleHostRect` → COM `SetBounds`) does.

Public types: `WebViewPanel` (`URL` / `HTML`, `FillClient`, messages), host status helpers, `WebViewHostSupported()`.

---

## Layout contract

| Need | Do | Don’t |
|------|----|--------|
| Fill-height embed | App shell + flex-grow row; `WebViewPanel` as **direct** flex child | Live WebView inside page-scroll |
| Full-client shell | `FillClient=true` below the title bar | OS passthrough / click-through tricks |
| Scrollable forms | Native widgets (or web outside the scroll band) | Nest WebView in titled Panel + body Viewport casually |

`visibleHostRect()`: panel bounds → clamp top to `Document.ChromeTop()` → intersect ancestor clips → host bounds. During force-sync / transient 0×0 layout: **keep last COM bounds** — do not Hide (blank web).

---

## Main-loop order (do not reorder)

```text
titleBar.Update → chrome drag / pending flags
Layout (when dirty)
SyncWebViewHosts(doc)      ← bounds + visibility + pointer policy
RouteScenePointerFocus(doc)
Draw / EndDrawing
PresentWebViewHosts()      ← Show + bounds-scoped raise + parent HWND pump
RouteScenePointerFocusAfterPresent (if handoff pending)
```

- Sync **after** Layout; Present **after** EndDrawing  
- Pump the **parent window HWND only** — never the global queue (`hwnd=0`)

---

## Title bar with a live WebView

While the cursor is in the title band, a drag is active, or a title click is pending: **defer HWND raise**, keep the host **visible** (hiding caused blank web).

- Demote stray Chrome/WebView children that are not near the panel bounds  
- Raise only near-matching HWNDs whose top ≥ `ChromeTop`  
- Same grab + double-click table as [window-chrome.md](window-chrome.md)  
- Edge resize: stay visible; live bounds sync + `MarkWebViewHostsResize`

Hosts must set `SetChromeTitleBarDragging(IsDragging() || IsTitleClickPending())` every frame after TitleBar update.

---

## FillClient `chrome.resize`

When `FillClient=true`, the flush HWND covers Raylib edge grips. The **page** owns S / E / W / SE / SW hit targets (`assets/web/gru.js` when window-chrome resize is enabled) and posts `chrome.resize` on the bridge; Go applies window size/position like the TitleBar.

- **Scope:** FillClient only (`cmd/webviewhello`, `gru new --webview`)  
- **Not for:** nested Focus Handoff embeds  
- N / NE / NW remain TitleBar-native  
- Do not inset the HWND L/R/bottom “for grips” (visible GL frame)

---

## Idle / frame budget

Web paint is independent of Gru redraw. Alive hosts use a WebView idle FPS floor; `SampleChromeHoverWake` keeps the title/web client snappy without global mouse wake.

Native modals under a full-client HWND cannot paint above the browser. Use engine occlusion (`WebViewHostOccluded` via `ShowModal` / drawer / bottom sheet) so hosts hide until the overlay closes. Howto: [../WEBVIEW.md](../WEBVIEW.md#modals-and-overlays-over-webview-occlusion). Demo: WebView Focus Handoff.

---

## Anti-patterns (summary)

| Rejected | Why |
|----------|-----|
| `WS_EX_TRANSPARENT` while visible | Blank web (GL shows through) |
| Hide WebView during title drag or edge resize | Blank until click / scene change |
| Raise all Chrome child HWNDs | Orphans cover the title bar |
| OS mouse polling over WebView for chrome | DPI/coord bugs; fights HWND |
| FillClient HWND insets for grips | GL “frame” around flush web |
| Re-navigate on every resize | Tears content; keep content synced |

Full table: [../WEBVIEW.md](../WEBVIEW.md).

---

## Samples

```bash
go run -tags webview2 ./cmd/webviewhello
go run -tags "grudemo,webview2" .
```

grudemo: **WebView Module Demo**, **WebView Focus Handoff**.  
Widgets map: [../widgets/webview.md](../widgets/webview.md).

Next: [input-focus.md](input-focus.md)
