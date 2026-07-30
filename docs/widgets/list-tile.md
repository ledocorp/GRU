# ListTile

**Purpose:** `ListTile` is a settings-style list row with optional leading/trailing slots, title, subtitle, and row-level click handling. It is the standard building block for settings pages, master lists, and navigation rows.

**Related:** [surfaces.md](surfaces.md) · [list-pane.md](list-pane.md) · [selection.md](selection.md)

---

## Key type

```go
tile := ui.NewListTile("wifi", "Wi-Fi", "Home network", 0, 0, 0, 0)
tile.SetTrailing(ui.NewToggle("wifi-tog", true, 0, 0, 52, 28))
tile.SetLeading(badgeOrIcon)
tile.OnClick = func() { /* navigate */ }
listContainer.AddChild(tile)
```

| Constructor | Signature |
|-------------|-----------|
| `NewListTile` | `NewListTile(id, title, subtitle string, x, y, w, h float32) *ListTile` |

Pass **`w=0, h=0`** for flex width and intrinsic height (`AutoHeight`).

### Row modes

| Mode | Constant | Behavior |
|------|----------|----------|
| Navigation | `ListTileNavigation` | Full-row hover, click, `OnClick` |
| Switch-only | `ListTileSwitchOnly` | Row body inert; only trailing `Toggle` receives input |

```go
tile.SetRowMode(ui.ListTileSwitchOnly)
```

Setting a trailing `*Toggle` via `SetTrailing` automatically switches to `ListTileSwitchOnly`.

### Slots & state

| API | Role |
|-----|------|
| `SetLeading(n Node)` | Left slot (badge, icon button, …) |
| `SetTrailing(n Node)` | Right slot (toggle, chevron label, value) |
| `SetRowMode(mode ListTileRowMode)` | Navigation vs switch-only |
| `OnClick func()` | Row tap handler (navigation mode) |
| `Selected bool` | Visual selection state |
| `Dense bool` | Shorter row height (48 vs 56 px) |
| `Children() []Node` | Exposes slots for hit-testing |

Theme key: **`list-tile`**.

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| ListTile variants in a panel | **Batch 21 · ListTile** |
| Desktop settings rows | **Settings · Desktop (Go)** |
| Mobile settings in App Shell | **Settings (Go)** |
| Master list in sidebar | **List Pane (Go)** |
| Toggle in trailing slot | **Batch 22 · Toggle** (also used inside tiles) |

---

## Pitfalls

| Mistake | Fix |
|---------|-----|
| `SetStyle("transparent")` on `ListTile` | Never — adjust parent `Panel` / `Card` |
| `OnClick` on switch-only rows | Use `ListTileSwitchOnly`; clicks go to the toggle |
| Long subtitles painting past row bounds | Truncate in app code (see list-pane demo subtitles) |
| Bare `Container` list without a surface | Wrap list in `Panel` — see [surfaces.md](surfaces.md) |
| Custom `SettingsRow` widget type | Compose `ListTile` + slots |

---

## Composition notes

1. **Always host lists in a surface** — `Panel` or `Card` body, or a `ListPane` scroll column ([list-pane.md](list-pane.md)).
2. **Switch rows:** trailing `Toggle` + automatic `ListTileSwitchOnly`; do not also wire row `OnClick`.
3. **Navigation rows:** optional trailing chevron as `Label` or `IconButton`; wire `OnClick` for drill-down.
4. **Selection:** update `Selected` on the active tile; rebuild or mark dirty when the selection model changes.
5. **Inspector:** slotted children appear under the tile node via `Children()` — useful when debugging hit targets.
