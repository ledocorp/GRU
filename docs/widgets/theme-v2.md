# Theme v2

**Purpose:** Theme v2 layers **component + variant** style resolution (`SetStyleVariant`) on top of the existing theme system — consistent button, badge, and control appearances without ad-hoc colors.

---

## Key API

```go
btn := ui.NewButton("save", "Save", 0, 0, 0, 36)
btn.SetStyleVariant("button", "primary")

badge := ui.NewBadge("status", "New", ui.BadgeInfo, 0, 0, 0, 0)
badge.SetStyleVariant("badge", "info")
```

| Pattern | Example keys |
|---------|--------------|
| Component + variant | `"button"` + `"primary"` / `"danger"` / `"default"` |
| Legacy style | `SetStyle("primary")` still works — demo shows side-by-side |

Surfaces use variants too: `SetStyleVariant("card", "default")`, `SetStyleVariant("panel", "default")`.

---

## Demo

See scene **Theme v2 Foundation** — button, badge, and control panels in a responsive grid.

**Related:** [surfaces.md](surfaces.md) · [buttons-signals.md](buttons-signals.md) · [../COMPOSITION.md](../COMPOSITION.md)

---

## Tips

- Prefer theme keys (`form-label`, `card`, `panel`, `list-tile`) over inline colors — golden rule in [README.md](README.md).
- When adding a new control variant, extend theme JSON/assets rather than per-widget overrides.
- Theme v2 demo uses `AutoHeight` panels in a 12-column grid — copy colspan pattern from the scene.
