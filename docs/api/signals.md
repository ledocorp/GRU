# Signals, effects, and bindings

Gru uses lightweight reactive **`Signal[T]`** values for UI state. Widgets expose mutable fields as signals (`Button` text, `Slider` value). Application code reads and writes them; **`Effect`** callbacks re-run when dependencies change.

**Package:** `github.com/ledocorp/gru/ui`  
**Demo:** grudemo scene **Counter Demo** (`examples/counter_demo.go`) — Tab to scene 12 in `go run -tags grudemo .`

Thread safety: all signal access must happen on the **main goroutine** (the raylib thread).

---

## Signal basics

```go
count := ui.NewSignal(0)

count.Set(1)           // notify subscribers if value changed
v := count.Get()       // read current value

count.SetDebounce(5)   // notify subscribers every 5th Set (slider drag throttling)
```

`Set` uses deep equality — assigning the same value is a no-op (avoids ping-pong between mirrored signals).

Subscribe manually when you do not use effects:

```go
count.Subscribe(func() {
    fmt.Println("count changed")
})
```

---

## Effect — automatic dependencies

`NewEffect` runs a function immediately and tracks every `Signal.Get()` inside it. When any dependency changes, the effect re-runs.

```go
count := ui.NewSignal(0)
label := ui.NewLabel("lbl", "Count: 0", 0, 0, 200, 24)

ui.NewEffect(func() {
    label.Text.Set(fmt.Sprintf("Count: %d", count.Get()))
})

// Later in a button handler:
count.Set(count.Get() + 1) // label updates same frame
```

During `Get()`, the signal registers the effect’s run function as a subscriber (SolidJS/MobX-style tracking). Duplicate subscriptions are deduplicated by function pointer.

There is no `Cancel` on `Effect` today — effects live for the scene lifetime.

---

## Binding — typed app state

`Binding[T]` wraps a signal with a clean scalar API:

```go
name := ui.NewBinding("Ada")
name.Set("Grace")
s := name.Get()

name.Subscribe(func() { /* react */ })
```

Use **`Binding`** for single values; use **`ListBinding[T]`** for virtual lists (items + selected index signals with `SubscribeItems` / `SubscribeSelection`).

---

## Async updates

Never touch signals from background goroutines directly. Use the document queue:

```go
count.SetAsync(doc, func() int {
    return expensiveCompute()
})
```

Or post manually:

```go
doc.QueueMain(func() { count.Set(42) })
```

Always **`doc.DrainQueue()`** at frame start before `Update`.

---

## Counter Demo walkthrough

The public **Counter Demo** scene shares one counter signal across panels:

```go
counter := ui.NewSignal(0)
autoReset := ui.NewSignal(false)

// Display label driven by effect:
countLbl, countDisplay := FlexCopyPair("counter-count", "form-value", "Count: 0")
ui.NewEffect(func() {
    countDisplay.Set(fmt.Sprintf("Count: %d", counter.Get()))
})

// Button mutates signal:
incBtn.OnClick = func() {
    next := counter.Get() + 1
    if autoReset.Get() && next > 100 {
        next = 0
    }
    counter.Set(next)
}
```

Other panels in the same scene attach **derived** effects — progress bar fill from `counter`, stats text, toggle bound to `autoReset` — without manual listener wiring.

Patterns shown:

| Pattern | Example in demo |
|---------|-----------------|
| Shared app signal | `counter` passed into multiple panel builders |
| Effect → label text | `countDisplay.Set(fmt.Sprintf(...))` |
| Effect → progress | Progress bar value from clamped counter |
| Signal → control state | `Toggle` bound to `autoReset` |
| Imperative feedback | `Tween` bounce on button click (not signal-driven) |

Run it:

```bash
go run -tags grudemo .
# Tab until "Counter Demo" (scene 12 in DEMO_INDEX)
```

Widget-oriented notes: [../widgets/buttons-signals.md](../widgets/buttons-signals.md).

---

## Widget signal fields

Many widgets already expose signals:

```go
btn := ui.NewButton("inc", "+", 0, 0, 48, 36)
btn.Text.Set("Add") // Button.Text is *Signal[string]

slider := ui.NewSlider("vol", 0, 0, 200, 24, 0, 100, 50)
slider.Value.Set(75)
```

Prefer effects to sync **between** widgets; mutate signals in `OnClick` / `Update` handlers for user actions.

---

## Next steps

- [node-element.md](node-element.md) — when to `MarkDirty` after signal-driven UI changes
- [document.md](document.md) — `DrainQueue` and main-loop ordering
- [../DEMO_INDEX.md](../DEMO_INDEX.md) — find Counter Demo and related scenes
