# Data display

**Purpose:** Read-only and tabular presentation — labels, rich text, badges, data tables, and timeline-style compositions.

---

## Key types

| Widget | Constructor | Role |
|--------|-------------|------|
| `Label` | `NewLabel(id, text, x, y, w, h)` | Plain text line |
| `RichText` / `PlainText` | `NewPlainText(id, styleKey, text, ...)` | Themed text + signal bind helpers |
| `Badge` | `NewBadge(id, text, variant BadgeVariant, x, y, w, h)` | Status chip |
| `DataTable[T]` | `NewDataTable[T](id, cols, binding *ListBinding[T], x, y, w, h)` | Typed tabular data |

Timeline demo composes labels/cards in a vertical list — not a separate `Timeline` widget type.

---

## Demos

| Goal | Scene title (exact) |
|------|---------------------|
| Data table | **Batch 6 · DataTable** |
| Badge variants | **Batch 3 · Badge** |
| Timeline composition | **Timeline (Go)** |
| Theme + badge panels | **Theme v2 Foundation** |

**Related:** [surfaces.md](surfaces.md) · [search.md](search.md) · [../DEMO_INDEX.md](../DEMO_INDEX.md)

---

## Tips

- Host tables inside a `Panel` with scroll enabled when row count is unbounded.
- `ListBinding[T]` connects table rows to your domain slice — see Batch 6 source.
- Use `BadgeVariant` constants rather than ad-hoc colors for status chips.
