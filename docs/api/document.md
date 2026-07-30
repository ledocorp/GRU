# Document

`ui.Document` is the top-level owner of the widget tree. Every Gru app creates one after `rl.InitWindow`, builds children under `doc.Root`, then drives it each frame.

**Module:** `github.com/ledocorp/gru/ui`  
**Canonical sample:** `cmd/hello` + `samples/hello`

---

## NewDocument

```go
doc := ui.NewDocument(windowW, windowH)
```

`NewDocument` allocates:

- A root `Container` named `"root"` sized to the window
- `LayoutAbsolute` on the root so direct children keep explicit bounds from your `Build()` and from resize (flex on the root would overwrite fixed-width children after resize)

Attach your scene to the root:

```go
doc.Add(ui.NewContainer("page", 0, 0, float32(w), float32(h)))
// or assign doc.Root children directly in a Build(doc) helper
```

---

## Chrome insets

Borderless Gru apps draw a custom title bar above scene content. Reserve space with **`SetChromeTop`** before mount/resize:

```go
doc.SetChromeTop(ui.TitleBarHeight) // 48px
```

| API | Role |
|-----|------|
| `SetChromeTop(top float32)` | Y inset for `doc.Root`; content draws below the title bar |
| `SetChromeBottom(bottom float32)` | Reserve bottom band (e.g. launcher nav) |
| `ChromeTop()` | Read current top inset |

`Resize` and `SyncBorderlessLayout` subtract chrome from the client height when sizing the root. Scene widgets use screen-absolute bounds starting at `Y = ChromeTop()`.

Pair document chrome with overlay helpers (toasts/tooltips hit-testing):

```go
ui.SetOverlayChromeInsets(ui.TitleBarHeight, 0)
```

See [mount-chrome.md](mount-chrome.md) for the full borderless spine.

---

## Frame loop

Each frame, run callbacks first, then input, layout, draw:

```go
ui.SetActiveDocument(doc)
doc.DrainQueue()

doc.Root.Update(dt)
if doc.Root.IsDirty() {
    doc.Root.Layout()
}

if doc.NeedsRedraw() {
    ui.BeginSuperFrame(bg, borderless, windowW, windowH)
    doc.Root.Draw()
    titleBar.Draw() // if borderless
    ui.EndSuperFrame()
}
rl.BeginDrawing()
ui.BlitToScreenBorderless(windowW, windowH, borderless)
rl.EndDrawing()
```

`SetActiveDocument` lets overlay widgets (modal `TextInput`, etc.) sync keyboard handling with `doc.Focused`.

---

## DrainQueue

Background work must not call raylib or mutate widgets off-thread. Post back to the main goroutine:

```go
go func() {
    data, err := fetch(url)
    if err != nil { return }
    doc.QueueMain(func() {
        label.Text.Set(data.Title)
        doc.Root.MarkDirty()
    })
}()

// start of each frame, before Update:
doc.DrainQueue()
```

| API | Role |
|-----|------|
| `QueueMain(fn func())` | Enqueue callback; wakes idle loop if needed |
| `DrainQueue()` | Run all pending callbacks once |
| `DrainQueueCount()` | Same, returns count run |
| `TaskQueueLen()` | Queue depth (monitoring) |

The queue is bounded (128). A full queue drops callbacks with a trace warning — keep async work coarse-grained.

`Signal.SetAsync(doc, computeFn)` wraps the same pattern for reactive updates.

---

## Resize

When `rl.GetScreenWidth()` / `GetScreenHeight()` change, update the document:

```go
// Thin-host path (preferred with TitleBar):
ui.SyncBorderlessClientSize(doc, titleBar, windowW, windowH)

// Or direct:
doc.Resize(fullW, fullH)
```

`Resize(fullW, fullH)`:

- Computes content height = `fullH - ChromeTop - ChromeBottom`
- Sets root bounds to the content band
- Propagates size into page shells and flex-grow children via `applyShellFlexAndSyncRootLayout`

Call **`MountBorderlessDocument`** once after building the scene — not on every resize tick. Use **`ResizeBorderlessDocument`** / **`SyncBorderlessClientSize`** during drag-resize. See [mount-chrome.md](mount-chrome.md).

---

## Focus

Keyboard events route to the focused node:

```go
doc.SetFocus(textInput)
doc.SetFocus(nil) // clear
```

`SetFocus` emits `EventBlur` on the old node and `EventFocus` on the new one. Setting focus releases WebView keyboard capture when applicable.

Interactive widgets (`TextInput`, `Button`, etc.) typically call `SetFocus` on click; read [../architecture/input-focus.md](../architecture/input-focus.md) for overlay and WebView handoff.

---

## NeedsRedraw and GPU cache

Gru skips expensive draw passes when the tree is visually clean (especially with supersampling).

```go
if doc.NeedsRedraw() {
    ui.BeginSuperFrame(...)
    doc.Root.Draw()
    ui.EndSuperFrame()
}
```

`NeedsRedraw()` returns true when:

- Any subtree has pending **layout** dirty (`IsDirty`)
- Any subtree has pending **draw** dirty (hover, blink, color — no layout change)
- No cache texture is available (1× mode without `EnableUIRenderTexture`)

Optional 1× fallback cache:

```go
doc.EnableUIRenderTexture(true)
```

In SSAA mode, the supersampling target already acts as a frame cache — `NeedsRedraw` gates `BeginSuperFrame` instead of redrawing every frame on idle UI.

| API | Role |
|-----|------|
| `NeedsRedraw()` | Should this frame run `Draw()` into the cache? |
| `EnableUIRenderTexture(on bool)` | Allocate 1× `RenderTexture` when SSAA is off |
| `InvalidatePaint()` | Force full repaint (scene remount, theme flip) |
| `UnloadCache()` | Free document GPU texture (scene unload / shutdown) |

---

## Responsive breakpoint

Package-level reactive width tier:

```go
ui.ActiveBreakpoint.Set(ui.CurrentBreakpoint(float32(windowW)))
tier := ui.ActiveBreakpoint.Get() // BreakpointXS … BreakpointXL
```

Demos read this in effects to reflow grid column spans; optional for fixed layouts.

---

## Next steps

- [node-element.md](node-element.md) — `Node` lifecycle behind `doc.Root`
- [mount-chrome.md](mount-chrome.md) — borderless mount helpers
- [../GETTING_STARTED.md](../GETTING_STARTED.md) — end-to-end hello loop
