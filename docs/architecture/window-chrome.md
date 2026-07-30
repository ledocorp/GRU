# Window chrome

Gru’s default desktop host is an **undecorated, resizable** window with a custom **`TitleBar`**, edge resize grips, and (on Windows 11+) **DWM rounded corners**. Thin hosts must follow the mount/resize APIs in `ui/app_mount.go` — do not freestyle a DIY caption strip.

Capability order: window manager → TitleBar → Document chrome → render → widgets. Do not strip chrome “for lean.”

---

## Default window flags

- `FlagWindowUndecorated` + `FlagWindowResizable` + `FlagMsaa4xHint` + `FlagVsyncHint`
- High-DPI layout flag is typically omitted so layout pixels match the framebuffer; quality comes from **2× SSAA**
- `SetWindowMinSize(MinClientWidth, …)` — **`MinClientWidth` is 480** (`ui/window_policy.go`)
- Paint: clear + `DrawBorderlessWindowFill` (rounded page fill) + alpha blit of the UI supersample

Content lives below the **48px** title band via `doc.SetChromeTop(ui.TitleBarHeight)` and `SyncBorderlessLayout`.

---

## TitleBar

```go
tb := ui.NewTitleBar(title, style, onClose, onMin, onMax)
tb.HandleResize = true // borderless edge/corner grips
```

| Constant / API | Role |
|----------------|------|
| `TitleBarHeight` (48) | Chrome band height |
| `Update` / `Draw` | Every frame |
| `ApplyResizeCursor` | Edge hit cursors |
| `IsDragging` / `IsTitleClickPending` / `IsResizing` | Chrome gesture state |
| `BorderlessRoundedChrome` | Drawn vs DWM coordination |

Windows: minimize / maximize / close via HWND. Double-click maximize/restore uses the native non-client path with a ShowWindow fallback. Non-Windows uses Raylib + timer double-click.

---

## Grab + double-click (contract)

Both paths are required. Immediate `beginDrag` on every press broke double-click; pending-only on restored windows broke grab.

| Window state | First press | Drag-out (≥4px) | Double-click |
|--------------|-------------|-----------------|--------------|
| **Restored** | Arm grab immediately (**no** restore) + pending | Already dragging | Maximize |
| **Maximized** | Pending only (**no** begin-drag) | ≥4px → restore then drag | Native restore |

Every frame after `titleBar.Update`, hosts must refresh:

```go
ui.SetChromeTitleBarDragging(tb.IsDragging() || tb.IsTitleClickPending())
```

`SyncBorderlessInputFrame` does this for thin hosts that call it. With a live WebView, the same flags drive **defer HWND raise** while the title band is active — see [webview.md](webview.md) and [../WEBVIEW.md](../WEBVIEW.md).

---

## DWM rounded corners (Windows)

- After window init and chrome toggles: `ApplyNativeBorderlessRoundedCorners` (`DWMWA_WINDOW_CORNER_PREFERENCE`)
- DWM rounds the **window silhouette**; the drawn fill still paints the client
- Each borderless frame: `SyncBorderlessChromeFrame(titleBar)` so native + drawn chrome stay aligned
- Skipping DWM → silhouette vs fill mismatch / corner artifacts

---

## Mount / resize contract

### Once per Build / remount

1. `doc.SetChromeTop(TitleBarHeight)`
2. Build the scene under `doc.Root`
3. `MountBorderlessDocument(doc, w, h)` — Resize + SyncBorderlessLayout + settle passes
4. Enable UI render texture / SSAA as in `cmd/hello`
5. `SetOverlayChromeInsets(TitleBarHeight, bottom)` so overlays respect the title band

### On client size change

```text
SyncBorderlessClientSize(doc, titleBar, w, h)
  → ResizeWindowTextures + display-aware SSAA + type scale
  → ResizeBorderlessDocument + titleBar.SetSize
```

Never blit an old supersample into a new client size. Do **not** call full mount/settle on every drag tick.

### Every borderless frame

1. `titleBar.Update`
2. `SyncBorderlessChromeFrame`
3. `SyncBorderlessInputFrame` (before `Root.Update`)
4. Draw: superframe → scene → `titleBar.Draw` → blit borderless

API cheat sheet: `MountBorderlessDocument`, `SyncBorderlessClientSize`, `ResizeBorderlessDocument`, `ResizeWindowTextures`, `SyncBorderlessChromeFrame`, `SyncBorderlessInputFrame`, `ApplyNativeBorderlessRoundedCorners`, `SetOverlayChromeInsets`.

---

## Do / Don’t

**Do**

- Copy `cmd/hello` / `cmd/webviewhello` chrome wiring  
- Keep the title bar **native Raylib** even when WebView is present  
- Set `HandleResize = true` on borderless apps  

**Don’t**

- Replace TitleBar with a DIY strip  
- Freestyle TitleBar gesture logic to “fix” lean hosts  
- Skip native DWM rounding on Windows  
- Mount/settle on every resize tick  
- Poll OS mouse (`GetAsyncKeyState`) over a WebView for chrome  

---

## Samples

- `go run ./cmd/hello`  
- `go run -tags webview2 ./cmd/webviewhello`  
- grudemo: **Desktop Shell (Go)**, **App Shell (Go)**  

Next: [layout.md](layout.md) · [input-focus.md](input-focus.md)
