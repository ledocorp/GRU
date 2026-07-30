# Sliders & progress

**Purpose:** `Slider` captures continuous numeric input along a track; `ProgressBar` displays determinate or animated progress. Both integrate with `Signal[float32]` for reactive UI.

**Related:** [buttons-signals.md](buttons-signals.md) · [forms.md](forms.md)

---

## Key types

### Slider

```go
vol := ui.NewSlider("volume", 0, 100, 50, 0, 0, 0, 0)
// Value on widget — bind labels with Effect or Form.AddField
```

| Constructor | `NewSlider(id string, minVal, maxVal, initialVal float32, x, y, w, h float32) *Slider` |

| Notes |
|-------|
| High-frequency drag updates — use `Signal.SetDebounced` on subscribers if layout is expensive |
| Usable inside `Form.AddField` |
| Typical height: intrinsic or ~36 px in form rows |

### ProgressBar

```go
bar := ui.NewProgressBar("upload", 0.35, 0, 0, 0, 0)
bar.Value.Set(0.72) // reactive 0..1
```

| Constructor | `NewProgressBar(id string, initialValue float32, x, y, w, h float32) *ProgressBar` |

Counter Demo ties progress to the counter signal and tweens — good reference for animated feedback.

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| Slider control | **Batch 24 · Slider** |
| Progress bar | **Batch 25 · ProgressBar** |
| Progress + signals + tweens | **Counter Demo** |
| Slider in labelled form | **Form Demo** |

---

## Pitfalls

| Mistake | Fix |
|---------|-----|
| Subscribing to slider with heavy layout every frame | Debounce subscriber or bind label only |
| Progress outside 0..1 | Clamp in app logic before `Value.Set` |
| Zero-width slider in non-flex parent | `SetFlexGrow(1)` in row or explicit width |
| Using `ProgressBar` as input | Use `Slider` for interactive ranges |

---

## Composition notes

1. **Form rows:** `f.AddField("Volume", slider)` — Form allocates row height (`RowH` default 36).
2. **Live labels:** `NewEffect` formatting `slider.Value.Get()` as text beside the control.
3. **Counter Demo pattern:** progress bar + tween on button click demonstrates non-linear feedback without blocking the UI thread.
4. **Batch panels:** slider/progress batch scenes use `MountAppPage` + `Panel` hosts — copy panel-per-widget layout.
5. **Theme:** progress fill and track use theme keys on the widget; avoid custom draw colors.
