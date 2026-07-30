# Mount and borderless chrome

Gru desktop apps use an **undecorated** raylib window plus a drawn **`TitleBar`**. Package `ui` exposes mount helpers so thin hosts (`cmd/hello`, `cmd/webviewhello`) match the Studio spine without copying a large main loop.

**Module:** `github.com/ledocorp/gru/ui`  
**Sample:** `cmd/hello/main.go`

---

## One-time mount

After you build the scene under `doc.Root`:

```go
doc.SetChromeTop(ui.TitleBarHeight)
hello.Build(doc)
ui.MountBorderlessDocument(doc, windowW, windowH)
```

`MountBorderlessDocument`:

1. Wires OS title-bar hooks (`WireBorderlessTitleBarOS`)
2. Ensures `ChromeTop` (defaults to `TitleBarHeight` if unset)
3. `Resize` + `SyncBorderlessLayout` — root bounds, shell flex, content band
4. `SettleDocumentMount` — up to five width-first layout passes for auto-height text
5. `InvalidatePaint` — force first frame draw

Call **once** after scene build or remount — **not** on every resize drag tick.

---

## Per-frame chrome sync

Borderless main loop (abbreviated from `cmd/hello`):

```go
titleBar := ui.NewTitleBar("Hello Gru", ui.TitleBarStyleDark,
    onClose, onMinimize, onMaximize)
titleBar.HandleResize = true
titleBar.SetSize(windowW, windowH)

ui.SetOverlayChromeInsets(ui.TitleBarHeight, 0)
ui.ApplyNativeBorderlessRoundedCorners(true)

for !shouldClose && !rl.WindowShouldClose() {
    ui.SetActiveDocument(doc)
    doc.DrainQueue()
    dt := rl.GetFrameTime()

    nw, nh := int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())
    if nw != windowW || nh != windowH {
        windowW, windowH = nw, nh
        ui.SyncBorderlessClientSize(doc, titleBar, windowW, windowH)
    }

    titleBar.Update(windowW, windowH)
    ui.SyncBorderlessChromeFrame(titleBar)
    titleBar.ApplyResizeCursor(windowW, windowH)
    ui.SyncBorderlessInputFrame(titleBar, doc.Root, dt)

    doc.Root.Update(dt)
    if doc.Root.IsDirty() {
        doc.Root.Layout()
    }
    // … NeedsRedraw → Draw → Blit …
}
```

---

## SyncBorderlessClientSize

Call whenever client width/height change. Never blit an old supersampling target to a new window size.

```go
ui.SyncBorderlessClientSize(doc, titleBar, fullW, fullH)
```

This mirrors the Studio resize path:

| Step | Effect |
|------|--------|
| `ResizeWindowTextures` | SSAA render targets match client |
| `ApplyDisplayAwareSupersampling` | Scale factor; may `doc.UnloadCache()` |
| `RefreshTypeScaleFromWindow` | Font/icon scale |
| `Toasts` / `Tooltips` window size | Overlay layout |
| `ResizeBorderlessDocument` | `doc.Resize` + `ForceFullLayout` + `InvalidatePaint` |
| `titleBar.SetSize` | Title bar + resize grips |

---

## SyncBorderlessChromeFrame

Once per frame after `titleBar.Update`:

```go
ui.SyncBorderlessChromeFrame(titleBar)
```

- Binds fill-client WebView chrome to the title bar when used
- Applies native rounded corners from `titleBar.BorderlessRoundedChrome()`
- Updates package-level rounded-chrome state for drawn window shape

---

## SyncBorderlessInputFrame

Call **after** `titleBar.Update`, **before** `doc.Root.Update`:

```go
ui.SyncBorderlessInputFrame(titleBar, doc.Root, dt)
```

Sets chrome drag / window-move flags (defers WebView raise during title drag), wheel suppress band at `TitleBarHeight`, then:

- `PrepareWheelScroll(root)` — route wheel to scroll viewports
- `ProcessSwitchListTilePointers(root, dt)` — switch rows on pointer pass

Required for correct pointer routing in borderless mode; see [../architecture/input-focus.md](../architecture/input-focus.md).

---

## TitleBar

```go
tb := ui.NewTitleBar(title, ui.TitleBarStyleDark, onClose, onMinimize, onMaximize)
tb.HandleResize = true   // edge/corner resize grips
tb.ShowAppIcon = true    // optional lettermark
tb.Draw()                // after doc.Root.Draw inside BeginSuperFrame
```

Constants:

| Name | Value | Role |
|------|-------|------|
| `TitleBarHeight` | 48 | Match `SetChromeTop` |
| `TitleBarShadowGap` | 6 | Gap below shadow — avoid opaque AppBar overlap |
| `MinClientWidth` | (package) | Minimum width during resize |

Styles: `TitleBarStyleDark`, `TitleBarStyleLight`, `TitleBarStyleAccent`.

Draw order: scene (`doc.Root.Draw`) then `titleBar.Draw` inside the same super-frame so shadows composite correctly.

---

## Overlay insets

Document chrome (`SetChromeTop`) shifts **scene content**. Overlay chrome insets shift **toasts, tooltips, and wheel suppress** separately:

```go
ui.SetOverlayChromeInsets(top, bottom)
```

Use the same top value as `doc.SetChromeTop` in standard borderless apps. Fill-client WebView layouts may use different pairing — see [../WEBVIEW.md](../WEBVIEW.md).

---

## Resize without remount

| API | When |
|-----|------|
| `ResizeBorderlessDocument(doc, w, h)` | Client size changed; no settle passes |
| `SyncBorderlessClientSize(doc, tb, w, h)` | Preferred — textures + overlays + document |
| `MountBorderlessDocument(doc, w, h)` | Scene first attached or rebuilt |

---

## Next steps

- [document.md](document.md) — `Resize`, `NeedsRedraw`, focus
- [../architecture/window-chrome.md](../architecture/window-chrome.md) — drag, grips, double-click maximize
- [webview-api.md](webview-api.md) — FillClient + `SyncWebViewHosts`
