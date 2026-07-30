# Input & focus

Pointer and keyboard reach **scene widgets**, **overlay managers**, or a **WebView HWND**. Routing is explicit — especially TextEditor ↔ WebView — and must not rely on click-through / passthrough tricks.

---

## Events & focus basics

- Event kinds include Click, Focus, Blur, KeyPress, MouseMove (`On` / `Emit`; bubble via parents).
- `Document.SetFocus(node)` blurs the previous focus and focuses the new node (nil clears; no-op if unchanged).
- Text fields gate keyboard on the focused flag from Focus/Blur.
- Hit testing walks interactive nodes with siblings ordered **high Z → low** (matches paint).

---

## Pointer preparation (main loop)

Typical borderless host order before `Root.Update`:

1. Prepare pointer input  
2. `SyncBorderlessInputFrame` — chrome drag flags, wheel suppress band under the title bar, wheel owner, ListTile switch-row pass  
3. Scene / overlay update  

Title band and chrome drag: Raylib owns the caption; WebView **defers raise** while pending/dragging (see [webview.md](webview.md)).

When a modal, drawer, bottom sheet, context menu, command palette, or notification center is open, scene pointer should be blocked (`OverlayBlocksSceneInput` / equivalent host gates).

---

## Overlays vs scene

| Layer | Input |
|-------|--------|
| Open overlay managers | Receive clicks; scene underneath should not |
| Modal / drawer / bottom sheet over WebView | `WebViewHostOccluded` → **hide** HWND (not passthrough) |
| Context menu | Does **not** necessarily occlude WebView — document app behavior carefully |
| Dropdown open list | Overlay path escapes Viewport scissor |

High `ZIndex` inside a Viewport does **not** escape clip — use the overlay path for open lists.

---

## WebView focus handoff

| Event | Behavior |
|-------|----------|
| Click `WebViewPanel` | Clear Document focus; give keyboard to the host |
| Click native `TextEditor` | `SetFocus(editor)`; release web keyboard |
| Click elsewhere | Release web keyboard focus |
| Modal over web | Hide via occlusion |
| Escape / blur paths | Host blur |

Call `RouteScenePointerFocus` after `SyncWebViewHosts` (and after present when a handoff is pending). Side-by-side layouts: keep HWND bounds from overlapping the editor — route by bounds, **never** `WS_EX_TRANSPARENT` while visible.

Demo: **WebView Focus Handoff**.

---

## Wheel / scroll ownership

Nested vs page gesture windows choose one scroll owner per gesture. Horizontal viewports should not steal vertical page wheel. Controls like Button/Slider typically do not own wheel — the underlying Viewport does.

---

## Chrome vs content (with WebView)

- Title gestures: [window-chrome.md](window-chrome.md) + [../WEBVIEW.md](../WEBVIEW.md)  
- FillClient edges: page JS grips → bridge `chrome.resize`  
- Rejected: TitleBar OS mouse polling over WebView  

Every frame:

```go
ui.SetChromeTitleBarDragging(tb.IsDragging() || tb.IsTitleClickPending())
```

---

## Do / Don’t

**Do**

- Route TextEditor ↔ WebView focus explicitly  
- Gate scene picks when overlays block  
- Keep chrome drag / pending flags fresh every frame  

**Don’t**

- Enable `WS_EX_TRANSPARENT` so Raylib can “see through” the HWND  
- Duplicate WebView pointer policy beyond title/overlay gates  
- Assume high ZIndex escapes Viewport clip  

---

## Samples

- **WebView Focus Handoff**, **WebView Module Demo**  
- `cmd/webviewhello`  
- **Batch 1 · Tooltip / TabView / Modal**, **Counter Demo**  

← [Architecture index](README.md)
