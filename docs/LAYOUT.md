# Layout contracts (public)

Short author rules for height, scroll, and the composition stack. Full detail: [architecture/layout.md](architecture/layout.md).

---

## One job per layer

```text
Document.Root → shell → (grid) → (Viewport) → Panel/Card → widgets
```

Overlays (modals, dropdowns) draw above the tree and must not change sibling layout extent.

---

## Name the height mode

| Mode | Meaning |
|------|---------|
| **Intrinsic** | Shrink-wrap content at assigned width |
| **Fill** | Consume remaining main-axis space (`FlexGrow`) |
| **Fixed** | Exact height band |

Pick one per node. Measure width-first. Never leave probe/measure heights as final bounds.

---

## Scroll owner

| Need | Use |
|------|-----|
| Page scrolls as a document | One page-scroll `Viewport` in the shell body |
| Pane fills remaining height | `FlexGrow` fill child (optional nested Viewport) |
| Modal | Flat flex body only — **not** a Viewport root |

---

## Related

- [COMPOSITION.md](COMPOSITION.md) · [CHEATSHEET.md](CHEATSHEET.md) · [architecture/layout.md](architecture/layout.md) · [GETTING_STARTED.md](GETTING_STARTED.md)
