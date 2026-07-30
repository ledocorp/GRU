# Forms & inputs

**Purpose:** `Form` lays out labelled fields; individual input widgets (`TextInput`, pickers, combo boxes, spin boxes) bind to `Signal` values for two-way reactive updates.

**Related:** [selection.md](selection.md) · [filters-toolbar.md](filters-toolbar.md) · [surfaces.md](surfaces.md)

---

## Key types

### Form container

```go
f := ui.NewForm("signup", 0, 0, 500, 0)
f.AddField("Username", ui.NewTextInput("user", "", 0, 0, 0, 0))
f.AddField("Volume", ui.NewSlider("vol", 0, 100, 50, 0, 0, 0, 0))
f.SetError("Username", "Required")
panel.AddChild(f)
```

| Constructor | `NewForm(id string, x, y, w, h float32) *Form` |
|-------------|--------------------------------------------------|

| Field / method | Notes |
|----------------|-------|
| `AddField(label string, widget Node)` | Pairs label + control |
| `Vertical bool` | Stacked label-above-control layout |
| `LabelW`, `RowH`, `Gap`, `FieldGap` | Layout tuning |
| `SetError(label, msg)` / `ClearError` | Inline validation styling |

Theme keys: **`form`**, **`form-label`**, **`form-error`**.

### TextInput

```go
name := ui.NewTextInput("name", "", 0, 0, 0, 0)
// Text held in widget signals — see godoc for Value / placeholder APIs
```

| Constructor | `NewTextInput(id, text string, x, y, w, h float32) *TextInput` |

### Date & time pickers

| Widget | Constructor |
|--------|-------------|
| `DatePicker` | `NewDatePicker(id string, value *Signal[time.Time], x, y, w, h float32)` |
| `DateRangePicker` | `NewDateRangePicker(id string, start, end *Signal[time.Time], x, y, w, h float32)` |

### Choice & numeric inputs

| Widget | Constructor |
|--------|-------------|
| `ComboBox` | `NewComboBox(id string, options []string, selected *Signal[string], x, y, w, h float32)` |
| `SpinBox` | `NewSpinBox(id string, value *Signal[float64], min, max, step float64, x, y, w, h float32)` |

### Boolean & range (also usable inside forms)

| Widget | See also |
|--------|----------|
| `Toggle`, `Checkbox` | [selection.md](selection.md) |
| `Slider` | [sliders-progress.md](sliders-progress.md) |

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| Full labelled form layout | **Form Demo** |
| Date picker | **Batch 3 · DatePicker** |
| Date range | **Batch 11 · DateRangePicker** · **Filters (Go)** |
| Combo box | **Batch 15 · ComboBox** · **Filters (Go)** |
| Spin box | **Batch 16 · SpinBox** |
| Text focus handling | **Theme v2 Foundation** (click-to-focus pattern) |

---

## Pitfalls

| Mistake | Fix |
|---------|-----|
| Form taller than viewport without scroll | Wrap form in scrollable `Panel` / `Viewport` |
| Hardcoded absolute Y positions per field | Let `Form.Layout` position rows |
| Mixing vertical and two-column fields inconsistently | Set `Form.Vertical` for narrow layouts |
| Forgetting to share `Signal` between summary labels and inputs | One signal per field, bind with `Effect` |
| Validation only in `OnClick` | Use `SetError` for inline field feedback |

---

## Composition notes

1. **Host forms in `Panel`** — Form Demo uses panel-per-section inside `MountAppPage`.
2. **Filter rows vs forms:** toolbar-style label + control rows (Filters demo) can mirror form layout with flex rows; dense settings use `Form.AddField`.
3. **Signals:** pickers and combo boxes take `*Signal[T]` — create signals once in scene state, pass to widgets and summary labels.
4. **Theme:** use `form-label` / `form-value` for helper copy outside the form widget itself.
5. **Focus:** demo scenes call `focusClickedTextInput(doc)` on click — copy when building text-heavy pages.
