# Filters & toolbar rows

**Purpose:** Compact control rows for filtering reports and dashboards — combo box, date range, and color pickers arranged as label + control pairs.

---

## Key types

| Widget | Constructor | Role |
|--------|-------------|------|
| `ComboBox` | `NewComboBox(id, options, selected *Signal[string], x, y, w, h)` | Single choice / searchable country list |
| `DateRangePicker` | `NewDateRangePicker(id, start, end *Signal[time.Time], x, y, w, h)` | Inclusive date range |
| `ColorWell` | `NewColorWell(id, initial, swatches, x, y, w, h)` | Preset + custom accent color |

Also see [forms.md](forms.md) for the same widgets in form layout.

---

## Demo

See scene **Filters (Go)** — ComboBox, DateRangePicker, and ColorWell with a live summary bound to shared signals.

```bash
go run -tags grudemo .
# Tab to Filters (Go)
```

**Related:** [forms.md](forms.md) · [selection.md](selection.md) · [../DEMO_INDEX.md](../DEMO_INDEX.md)

---

## Tips

- Build toolbar rows as flex `Container` rows (`filterFieldRow` pattern in the demo) with fixed label width.
- One `Signal` per filter dimension; bind summary card text with `NewEffect`.
- `ColorWell` swatches are app-provided `[]rl.Color` — keep the palette stable in scene state.
