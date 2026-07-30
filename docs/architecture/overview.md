# Overview

Gru is a **retained-mode** Go UI engine: widgets live in a `Node` tree across frames. Each frame is strictly **Update → Layout → Draw**. Dirty flags decide when layout and redraw run. The `Document` owns the root, chrome insets, focus, and the main-thread work queue.

Go-first: compose in Go. Optional `.gru` / DocumentSpec compiles to the same widgets — there is no second layout engine.

---

## Retained mode

| Idea | Meaning |
|------|---------|
| Persist | Widgets are not rebuilt every frame |
| Mutate | Change properties / signals; mark dirty |
| Engine work | Layout only when dirty; redraw only when needed; otherwise blit the last SSAA frame |

Philosophy:

- Explicit frame phases — a phase must not call into another
- Reactive by default — mutable UI state is typically `Signal[T]`
- Minimal work — idle frames blit the cache
- Composition — widgets embed `Element` and implement `Node`

---

## Core types

### `Node`

`Update(dt)`, `Layout()`, `Draw()`, `MarkDirty()`, `IsDirty()`, bounds, children, events, visibility, `ZIndex`, `FlexGrow`.

### `Element`

Embedded base: parent link, dirty flags, cache fields, event handlers.

**Contract:** `parent` is typed as `Node` (not `*Container`). Dirty must bubble through `Container`, `Viewport`, and any wrapper. Asserting `*Container` can silently drop dirty under a `Viewport` (scroll vs hit-test bugs).

### `Document`

Owns root `Container`, logical size, focus, chrome insets, optional UI render-texture, `QueueMain` / `DrainQueue`.

```go
doc := ui.NewDocument(w, h) // root uses LayoutAbsolute
doc.SetChromeTop(ui.TitleBarHeight)
// build scene under doc.Root …
doc.DrainQueue()
doc.Root.Update(dt)
if doc.Root.IsDirty() {
    doc.Root.Layout()
}
```

Authoritative public loop: copy [`cmd/hello`](../../cmd/hello).

---

## Frame loop

```text
Drain wake / sample input
doc.DrainQueue()
titleBar.Update (+ SyncBorderlessInputFrame)
doc.Root.Update(dt)
if dirty → Layout
if NeedsRedraw → BeginSuperFrame → Draw → EndSuperFrame
else → blit previous SSAA
BeginDrawing → BlitToScreen* → overlays → EndDrawing
[WebView: SyncWebViewHosts after Layout; PresentWebViewHosts after EndDrawing]
```

| Rule | Why |
|------|-----|
| Update always | Mutations visible to every widget before layout |
| Layout only if dirty | Avoids full tree work on idle frames |
| Draw into SSAA only if `NeedsRedraw()` | Clean frames blit the cache |
| Never Layout from Draw | Mid-paint geometry breaks hit-testing |

---

## Dirty flags (two systems)

| Flag | Meaning | Set by | Cleared by |
|------|---------|--------|------------|
| Layout dirty | Needs relayout | `MarkDirty()` | `Layout()` (overrides must clear) |
| Draw dirty | Needs redraw | `MarkDrawDirty()`, construction | `Draw()` (overrides must clear) |

- `MarkDirty()` sets self dirty (and cache dirty if cached) and bubbles to the root via `parent.MarkDirty()`.
- `NeedsRedraw()` walks the visible tree: any layout or draw dirty forces a full SSAA redraw; otherwise blit-only.
- Hidden subtrees do not pin redraw.
- Prefer interaction/animation **overlays** for transient hover chrome — do not `MarkDrawDirty` for every hover.
- During layout geometry moves use layout-only setters (`layoutSetBounds`), not `SetBounds` (avoids dirty every frame).

---

## Document API (essentials)

| API | Role |
|-----|------|
| `NewDocument(w, h)` | Root is **absolute** so resize does not flex-scatter direct children |
| `SetChromeTop` / `ChromeTop` / `SetChromeBottom` | Content band below title / above reserved bottom |
| `Resize` | Shell-aware resize; clamps width ≥ `MinClientWidth` (**480**) |
| `SyncBorderlessLayout` / `ForceFullLayout` | Fit content under chrome |
| `InvalidatePaint` | After scene remount — never keep a stale SSAA blit |
| `SetFocus` / `Focused` | Single focus; no-op if already focused |
| `QueueMain` / `DrainQueue` | Goroutine → main thread before Update |
| `SetActiveDocument` | Per-frame for overlay managers |

Scene footers (`StatusBar`, etc.) live **in the tree** — they are not `SetChromeBottom` unless you intentionally reserve OS chrome.

---

## Reactivity (short)

- `Signal[T]` — mutable value with subscribers  
- `Effect` — auto-tracking side effects  
- `Binding[T]` / `ListBinding[T]` — bind widgets to data  

Layout-affecting changes should dirty layout; pure hover chrome should use overlays (see [rendering.md](rendering.md)).

Demo: **Counter Demo** (`go run -tags grudemo .`).

---

## Do / Don’t

**Do**

- Mutate retained widgets; let dirty drive work  
- `DrainQueue` before Update  
- Gate Layout on `IsDirty()`; gate SSAA on `NeedsRedraw()`  
- Copy `cmd/hello` for the public spine  

**Don’t**

- Call Layout from Draw (or Draw from Update)  
- Type-assert parent to `*Container` for dirty propagation  
- Rebuild the whole tree every frame “to be safe”  
- Skip `InvalidatePaint` after scene remount  

---

## Next

- [window-chrome.md](window-chrome.md) — TitleBar, DWM, mount contract  
- [layout.md](layout.md) — flex, Viewport, height modes  
- [GETTING_STARTED.md](../GETTING_STARTED.md) · [api/README.md](../api/README.md)
