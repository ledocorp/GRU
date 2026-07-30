# Buttons & signals

**Purpose:** `Button` and `IconButton` handle primary user actions. `Signal[T]` and `Effect` provide fine-grained reactive state so labels and controls update without manual rebuild loops.

**Related:** [../GETTING_STARTED.md](../GETTING_STARTED.md) · [sliders-progress.md](sliders-progress.md) · [forms.md](forms.md)

---

## Key types

### Button

```go
btn := ui.NewButton("save", "Save", 0, 0, 0, 36)
btn.SetStyle("primary") // or "button", theme variants via SetStyleVariant
btn.OnClick = func() { /* action */ }
row.AddChild(btn)
```

| Constructor | Signature |
|-------------|-----------|
| `NewButton` | `NewButton(id, text string, x, y, w, h float32) *Button` |

| Field | Notes |
|-------|-------|
| `Text *Signal[string]` | Reactive label |
| `OnClick func()` | Fires on mouse press |
| `Scale float32` | Animation hook (tween on click) |
| `ToggleBinding *Signal[bool]` | Optional toolbar toggle tint |

Pass **`h=0`** for intrinsic height; pass **`w=0`** in flex rows with sibling sizing.

### IconButton

```go
ib := ui.NewIconButton("refresh", "Refresh", "", 0, 0, 32, 32)
ib.SetPhosphorIcon(ui.PhosphorArrowClockwise, ui.PhosphorRegular)
ib.OnClick = func() { /* ... */ }
```

Used in app bars, list-pane collapse, and toolbar rows. See `ui/button.go` for `NewIconButton`.

### Signal & Effect

```go
count := ui.NewSignal(0)
label := ui.NewPlainText("status", "form-label", "0", 0, 0, 0, 0)
ui.BindPlainText(label, count) // or manual Effect:

ui.NewEffect(func() {
    label.SetText(fmt.Sprintf("Count: %d", count.Get()))
})

count.Set(count.Get() + 1) // subscribers run synchronously
```

| API | Role |
|-----|------|
| `NewSignal[T](initial T) *Signal[T]` | Observable value |
| `Get() T` / `Set(value T)` | Read / write (main thread only) |
| `Subscribe(fn func())` | Manual subscription |
| `NewEffect(fn func()) *Effect` | Auto-tracks signals read inside `fn` |
| `SetDebounced(value, n)` | Notify every N sets (sliders, search) |

Many widgets expose `Signal` fields (`Button.Text`, `Slider.Value`, `Toggle.Value`, …).

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| Signals, effects, progress, tweens | **Counter Demo** (default starter scene) |
| Primary / secondary buttons | **Counter Demo** · **Theme v2 Foundation** |
| Modal action buttons | **Batch 1 · Tooltip / TabView / Modal** |
| Minimal Button + Signal sample | `go run ./cmd/hello` |

---

## Pitfalls

| Mistake | Fix |
|---------|-----|
| `NewButton(..., w, 0)` with fixed width but zero height in non-flex parent | Use flex row or set explicit height (36 is typical) |
| Updating labels manually every frame | Bind with `Signal` + `Effect` or `BindPlainText` |
| Cross-goroutine `Signal.Set` | Signals are **main-thread only** |
| Ping-pong loops between mirrored signals | `Set` no-ops when value unchanged — design one source of truth |
| Heavy work inside `Effect` callbacks | Debounce or move work out of reactive path |

---

## Composition notes

1. **Starter scene:** **Counter Demo** is the default `grudemo` entry — best first read for signals + buttons together.
2. **Hello sample:** `samples/hello` uses `Signal` + `Button` inside a `Card` without the demo harness — copy for standalone apps.
3. **Button rows:** flex `Container` with `FlexDirection = FlexRow`, `Gap = 8`, intrinsic-height buttons.
4. **Primary actions:** first button in a row often uses `SetStyle("primary")`; destructive actions use theme variant `danger` (Theme v2 demo).
5. **Progress feedback:** Counter Demo ties `ProgressBar` to the same signal model — see [sliders-progress.md](sliders-progress.md).
