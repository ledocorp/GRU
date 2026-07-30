# Surfaces (Panel & Card)

**Purpose:** `Panel` and `Card` are framed surface widgets that host page content — especially stacks of `ListTile` rows, form sections, and nested callouts. They replace ad-hoc styled `Container` wrappers.

**Related:** [../COMPOSITION.md](../COMPOSITION.md) · [list-tile.md](list-tile.md) · [shells.md](shells.md)

---

## Key types

Both types are facades over `SurfaceShell` with different header modes.

| Widget | Constructor | Header mode | Typical role |
|--------|-------------|-------------|--------------|
| `Panel` | `ui.NewPanel(id, title, x, y, w, h)` | Title bar band | Settings sections, batch demo hosts, scrollable lists |
| `Card` | `ui.NewCard(id, title, x, y, w, h)` | Inset title | Nested sections, callouts, detail panes |

### Common API

```go
p := ui.NewPanel("settings", "Settings", 0, 0, 0, 0)
p.SetFlexGrow(1)
p.Gap = 12
p.AutoHeight = true          // shrink-wrap content in grids
p.AddChild(widget)

p.EnableCollapse(true)       // optional collapse behavior
p.SetHeaderMode(ui.HeaderModeGlass) // title band variants
```

| Method / field | Notes |
|----------------|-------|
| `AddChild(child Node)` | Attaches to body or internal scroll host |
| `EnableCollapse(expanded bool)` | Returns `*CollapseBehavior` |
| `SetHeaderMode(mode SurfaceHeaderMode)` | `TitleBar`, `Inset`, `Glass`, `None`, … |
| `Gap`, `AutoHeight`, `TitleHeight` | Layout tuning |
| `SetColSpan(bp, cols)` | When parent is a page grid |
| `ApplyPanelBodyTextColor()` / `ApplyCardBodyTextColor()` | Sync label typography to surface chrome |

Theme keys: **`panel`** and **`card`** (via `SetStyleVariant("panel", "default")` etc.).

### PanelFeatures (scroll, move, resize)

`Panel` and `Card` attach `PanelFeaturesBehavior` by default. Enable vertical scroll on the body when content may exceed visible height — see `ui/panel_features.go` and batch demos that host long field lists.

---

## When to use which

| Situation | Prefer |
|-----------|--------|
| Primary page section with a strong title band | `Panel` |
| Nested block inside another surface | `Card` |
| Settings-style list of rows | `Panel` + `ListTile` children |
| Detail pane beside a sidebar | `Card` in the main column |
| Full-width demo section in a grid | `Panel` with `AutoHeight = true` |

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| Nested cards and callouts | **Card Nest (Go)** |
| Panel-per-section batch layout | **Form Demo** |
| Theme variants on surfaces | **Theme v2 Foundation** |
| List tiles inside a panel | **Batch 21 · ListTile** |
| Grid + panel colspan | **Responsive - Breakpoints - Grid** |
| Minimal Card usage | `go run ./cmd/hello` |

---

## Pitfalls

| Mistake | Fix |
|---------|-----|
| Dense `ListTile` stack in a bare `Container` | Wrap in `Panel` or `Card` |
| Fixed height panels in responsive grids | Set `AutoHeight = true` or explicit min height |
| One-off background colors on every node | Use theme keys `panel` / `card` |
| `SetStyle("transparent")` on `ListTile` to fix clipping | Fix parent surface bounds / scroll |
| Putting live `WebViewPanel` inside Panel body stacks casually | Follow [../WEBVIEW.md](../WEBVIEW.md) layout rules |

---

## Composition notes

1. **One job per layer:** shell → grid (optional) → **Panel / Card** → flex rows / widgets.
2. **Lists belong in surfaces** — this is the most important public composition rule ([../COMPOSITION.md](../COMPOSITION.md)).
3. **Grid placement:** panels in responsive demos use `SetColSpan` across breakpoints; default span is full row (12 cols).
4. **Typography:** direct `Label` / `RichText` children inherit body text color via `Apply*BodyTextColor`.
5. **Do not invent** `SettingsPanel`, `SidebarCard`, or similar wrappers — compose with `Panel` / `Card` + layout containers.
