# List pane (master–detail)

**Purpose:** The list-pane pattern combines a flat, scrollable sidebar of `ListTile` rows with a detail region — typically via `SplitView`. The demo catalog provides `NewListPane` in `examples/list_pane_shell.go` as a reusable sidebar builder.

**Related:** [list-tile.md](list-tile.md) · [surfaces.md](surfaces.md) · [shells.md](shells.md)

---

## Key types

### ListPane (examples helper)

```go
pane := examples.NewListPane(doc, examples.ListPaneOptions{
    ID:             "docs",
    Title:          "Documents",
    PreferredWidth: 260,
    MinWidth:       180,
    ShowCollapse:   true,
    OnCollapse:     func() { /* hide sidebar */ },
})
// Add rows: pane.List.AddChild(ui.NewListTile(...))
```

| Field / type | Role |
|--------------|------|
| `ListPane.Root` | Borderless flex column (`list-pane` style) |
| `ListPane.List` | Flex column of `ListTile` rows |
| `ListPane.Scroll` | Flush viewport; scrollbar at split edge |
| `ListPaneOptions` | ID, title, widths, collapse button |

This helper is **not** a separate `ui` package type — it composes `Container`, `Viewport`, `Label`, and optional `IconButton`.

### SplitView

```go
split := ui.NewSplitView("main-split", ui.SplitHorizontal, leftPane, rightDetail, 0, 0, 0, 0)
split.SetFlexGrow(1)
```

| Constructor | Signature |
|-------------|-----------|
| `NewSplitView` | `NewSplitView(id string, dir SplitDirection, first, second Node, x, y, w, h float32) *SplitView` |

Use `SplitHorizontal` for sidebar | detail. Drag the divider to resize; respect `MinWidth` on the pane root.

### Supporting surfaces

Detail content typically lives in a `Card` or `Panel` in the second split pane — see [surfaces.md](surfaces.md).

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| Full list-pane + split + detail | **List Pane (Go)** |
| ListTile row behavior | **Batch 21 · ListTile** |
| Edge shell hosting split layout | **List Pane (Go)** (uses `MountEdgeToEdgeRoot`) |

---

## Pitfalls

| Mistake | Fix |
|---------|-----|
| Inventing a `Sidebar` or `MasterList` widget | Use `NewListPane` or copy its structure |
| Rebuilding the whole shell on row select | Update detail labels/cards; keep list tiles stable |
| List outside a scroll viewport | Use `ListPane.Scroll` or Panel scroll features |
| Transparent list-tile hacks | Fix pane `list-pane` / `list-pane-header` theme |
| Ignoring `MinWidth` / `PreferredWidth` | Set via `ListPaneOptions` for usable split drag |

---

## Composition notes

1. **Shell:** **List Pane (Go)** uses **CP-SHELL-EDGE** (`MountEdgeToEdgeRoot`) — menubar/status optional; split fills body.
2. **Row model:** each document/item → one `ListTile` with `OnClick` updating selection + detail pane.
3. **Subtitles:** truncate long paths in app code before assigning `ListTile.Subtitle`.
4. **Collapse:** optional `IconButton` in pane header; wire `OnCollapse` to hide the split first pane or zero its width.
5. **No new list wrappers** — the public rule is `Panel` / `Card` + `ListTile`; `ListPane` is an examples-level composition helper, not a license to add app-specific sidebar types.
