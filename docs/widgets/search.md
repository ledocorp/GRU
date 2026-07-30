# Search

**Purpose:** `SearchBar` is a single-line search field with optional clear affordance, reactive text via `Signal`, and debounce-friendly updates for filtering lists or tables.

---

## Key type

```go
q := ui.NewSearchBar("find", "Search…", 0, 0, 0, 40)
q.SetFlexGrow(1)
// Text signal on widget — filter list models in Subscribe/Effect
```

| Constructor | `NewSearchBar(id, placeholder string, x, y, w, h float32) *SearchBar` |

---

## Demo

See scene **Batch 2 · SearchBar** in the demo catalog (`go run -tags grudemo .`, Tab to scene).

**Related:** [forms.md](forms.md) · [data-display.md](data-display.md) · [../DEMO_INDEX.md](../DEMO_INDEX.md)

---

## Tips

- Debounce expensive filters with `Signal.SetDebounced` on the query signal subscriber side.
- Place search bars in a toolbar flex row above a `Panel` list or `DataTable`.
- Share one search signal between the bar and a summary label with `NewEffect`.
