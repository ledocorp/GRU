# Selection controls

**Purpose:** Toggle, checkbox, segmented control, rating, and related widgets capture boolean, enum, and scalar choices. Many integrate with `ListTile` trailing slots or `Form.AddField`.

**Related:** [list-tile.md](list-tile.md) · [forms.md](forms.md) · [buttons-signals.md](buttons-signals.md)

---

## Key types

### Toggle

```go
enabled := ui.NewSignal(true)
tog := ui.NewToggle("wifi", true, 0, 0, 52, 28)
// Value signal on widget — see godoc
tile.SetTrailing(tog) // auto ListTileSwitchOnly
```

| Constructor | `NewToggle(id string, initialValue bool, x, y, w, h float32) *Toggle` |

### Checkbox

```go
cb := ui.NewCheckbox("agree", false, 0, 0, 0, 0)

// Labelled variant:
cap := ui.NewCheckboxCaption("agree-cap", "I agree", agreeSignal, boxSize, fontSize, gap)
```

| Constructor | Notes |
|-------------|-------|
| `NewCheckbox(id, initialValue bool, x, y, w, h)` | Standalone box |
| `NewCheckboxCaption(id, text, value *Signal[bool], ...)` | Box + caption row |

### SegmentedControl

```go
sel := ui.NewSignal(0)
seg := ui.NewSegmentedControl("view", []string{"List", "Grid", "Map"}, sel, 0, 0, 0, 36)
```

| Constructor | `NewSegmentedControl(id string, options []string, selected *Signal[int], x, y, w, h float32)` |

### Rating

```go
stars := ui.NewSignal(float32(3))
rating := ui.NewRating("stars", stars, 5, 0, 0, 0, 0)
```

| Constructor | `NewRating(id string, value *Signal[float32], maxStars int, x, y, w, h float32)` |

### ComboBox (single selection from list)

Also documented under [forms.md](forms.md):

```go
country := ui.NewSignal("United States")
cb := ui.NewComboBox("country", countries, country, 0, 0, 0, 40)
```

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| Toggle switch | **Batch 22 · Toggle** |
| Checkbox | **Batch 23 · Checkbox** |
| Segmented control | **Batch 14 · SegmentedControl** |
| Star rating | **Batch 12 · Rating** |
| Combo box selection | **Batch 15 · ComboBox** |
| Toggle in ListTile | **Settings · Desktop (Go)** · **Batch 21 · ListTile** |

---

## Pitfalls

| Mistake | Fix |
|---------|-----|
| Row `OnClick` + trailing toggle | Use `ListTileSwitchOnly` ([list-tile.md](list-tile.md)) |
| Unbound segmented index | Share one `*Signal[int]` with summary labels via `Effect` |
| Checkbox without signal wiring | Use `NewCheckboxCaption` or bind value signal |
| Combo box options recreated every frame | Stable `[]string` slice in scene state |
| Rating max stars mismatch | Pass consistent `maxStars` and signal clamping |

---

## Composition notes

1. **Settings pages:** boolean prefs → `ListTile` + trailing `Toggle`; multi-option view mode → `SegmentedControl` in a toolbar row.
2. **Forms:** add toggles/checkboxes with `Form.AddField("Notify", toggleWidget)`.
3. **Signals:** all selection widgets notify on change — bind read-only labels with `NewEffect`.
4. **Theme v2:** button-like controls support `SetStyleVariant("button", "primary")` etc. — see [theme-v2.md](theme-v2.md).
5. **Accessibility / hit targets:** keep trailing toggles at ~52×28 or larger in list rows.
