# Node and Element

Every widget in Gru is a **`Node`**. Concrete types embed **`Element`** and override the methods they need. The engine walks the tree each frame: **Update → Layout → Draw**.

**Package godoc:** `go doc github.com/ledocorp/gru/ui.Node`

---

## Node interface

`Node` is the contract all widgets implement (directly or via embedding):

| Method | Phase | Purpose |
|--------|-------|---------|
| `Update(dt float32)` | Input | Pointer/keyboard, animations, signal side effects |
| `Layout()` | Geometry | Recompute bounds when dirty; no-op if clean |
| `Draw()` | Render | Raylib draws inside `Bounds()` |
| `MarkDirty()` | — | Layout + draw needed; bubbles to root |
| `MarkDrawDirty()` | — | Draw only (hover, color); no layout pass |
| `IsDirty()` | — | Should `Layout()` run this frame? |
| `Bounds()` / `SetBounds(rect)` | Geometry | Screen-space rectangle |
| `AddChild` / `RemoveChild` / `Children()` | Tree | Hierarchy |
| `Parent()` / `ParentNode()` | Tree | Walk ancestors (`ParentNode` for `*Viewport`, `*Panel`, etc.) |
| `Hide()` / `Show()` / `IsHidden()` | Visibility | Skip update/layout/draw |
| `GetZIndex()` / `SetZIndex()` | Layering | Sibling draw and hit-test order |
| `GetFlexGrow()` / `SetFlexGrow()` | Layout | Flex participation in parent |
| `IsInteractive()` | Input | Participates in focus/hit testing |
| `UsesScissor()` | Draw | Opens scissor in `Draw()` (viewport coordination) |

Event helpers come from embedded **`EventEmitter`** (`Emit`, `On`, focus/blur events).

---

## Element base struct

Custom widgets embed `Element`:

```go
type Meter struct {
    ui.Element
    value float32
}

func NewMeter(id string) *Meter {
    return &Meter{Element: ui.NewElement(id, 0, 0, 120, 24)}
}

func (m *Meter) Update(dt float32) {
    // read input, mutate state
}

func (m *Meter) Layout() {
    m.ClearLayoutDirtyFlag() // required — idle FPS contract
}

func (m *Meter) Draw() {
    b := m.Bounds()
    // draw inside b
}
```

`NewElement(id, x, y, w, h)` sets initial bounds. Use **`SetStyle("button")`** / theme v2 presets for themed chrome.

---

## Dirty flags

Two levels:

| Flag | Set by | Triggers |
|------|--------|----------|
| `layoutDirty` | `MarkDirty()`, bounds/structure changes | `Layout()` + `Draw()` |
| `drawDirty` | `MarkDrawDirty()`, visual-only changes | `Draw()` only |

```go
btn.MarkDrawDirty() // hover highlight — skip Layout()
panel.MarkDirty()   // child added or size changed — run Layout()
```

`MarkDirty()` propagates upward so `doc.Root.IsDirty()` drives the layout gate in main:

```go
if doc.Root.IsDirty() {
    doc.Root.Layout()
}
```

**Invariant:** leaf widgets must clear `layoutDirty` in `Layout()` even when they have no children. See architecture [rendering.md](../architecture/rendering.md) for idle/wake interaction with `NeedsRedraw()`.

---

## Update / Layout / Draw

### Update

Runs **before** layout. Mutate reactive state, handle clicks, advance tweens. Do not assume final child bounds yet.

### Layout

Runs only when the subtree is layout-dirty. Containers run flex/grid; `Panel`/`Viewport` delegate to the same engine. Idempotent when `IsDirty()` is false.

### Draw

Always invoked when the frame paints (gated upstream by `doc.NeedsRedraw()` in SSAA apps). Clip carefully: `BeginScissorMode` / `EndScissorMode` are not stackable — `Viewport` re-applies clip after children that open their own scissor.

---

## Common node types

| Type | Role |
|------|------|
| `Container` | Flex row/column/grid layout box |
| `Viewport` | Scrollable, clipped region + scrollbar |
| `Panel` / `Card` | Titled surface (see [../widgets/surfaces.md](../widgets/surfaces.md)) |
| `Button`, `Label`, `TextInput`, … | Leaf controls — [../widgets/README.md](../widgets/README.md) |
| `WebViewPanel` | Layout slot for embedded web — [webview-api.md](webview-api.md) |

Add children with the parent’s `AddChild` so `ParentNode()` chains work (`Viewport.AddChild` sets parent to the viewport, not only the embedded container).

---

## Building a scene

Typical pattern:

```go
func Build(doc *ui.Document) {
    root := ui.NewContainer("page", 0, 0, 0, 0)
    root.LayoutType = ui.LayoutFlex
    root.FlexDirection = ui.FlexColumn
    root.SetFlexGrow(1)

    body := ui.NewViewport("body", 0, 0, 0, 0)
    body.SetFlexGrow(1)
    root.AddChild(body)

    doc.Add(root)
}
```

After borderless mount, `doc.Resize` / shell helpers size `root` to the content band below `ChromeTop()`.

---

## Next steps

- [document.md](document.md) — `Document` owner and main loop
- [signals.md](signals.md) — reactive fields on widgets
- [../architecture/layout.md](../architecture/layout.md) — flex, grid, viewport contracts
