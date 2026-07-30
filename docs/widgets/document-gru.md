# Document / `.gru` pages

**Purpose:** Optional **DocumentSpec** pages loaded from `.gru` files — declarative structure compiled into `ui` nodes via `BuildContext`. Go remains authoritative for shell mount and hot-reload wiring.

---

## Key types & flow

```go
ctx := ui.NewBuildContext()
ctx.LinkHandler = func(link string) { /* in-app nav */ }
ctx.Actions["myAction"] = func() { /* button actions */ }

compiled, err := gruLoader.Compile() // see examples GRUPageReloader
gruLoader.MountShell(doc, id, title, subtitle, compiled)
```

| Concept | Role |
|---------|------|
| `BuildContext` | Link handlers, named actions, build options |
| DocumentSpec | Compiled `.gru` page definition |
| `GRUPageReloader` | Examples helper: compile + mount + poll hot reload |

Pages live under `pages/` in the repo; run demos from repo root so paths resolve.

---

## Demo

See scene **Gallery (.gru)** — loads `pages/gallery.gru` as a DocumentSpec control gallery with hot-reload.

**Related:** [shells.md](shells.md) · [../GETTING_STARTED.md](../GETTING_STARTED.md) · [../COMPOSITION.md](../COMPOSITION.md)

---

## Tips

- **Go-first** for app logic — use `.gru` for document-shaped layout when it helps, not as a full app replacement.
- Wire `PreserveControls: true` when hot-reloading gallery-style pages (demo pattern).
- On load error, the gallery scene falls back to a Go-built error `Card` — copy that guard for shipped apps.
