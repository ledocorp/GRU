# Navigation widgets

**Purpose:** Chrome and flow widgets that move users between sections — rails, app bars, steppers, pagination, and accordions.

---

## Key types

| Widget | Constructor | Role |
|--------|-------------|------|
| `NavigationRail` | `NewNavigationRail(id, items, selected *Signal[int], x, y, w, h)` | Desktop left rail |
| `AppBar` | `NewAppBar(id, title, x, y, w, h)` | Mobile / app-shell top bar |
| `Header` | `NewHeader(id, title, subtitle, x, y, w, h)` | Page title block (scrolls with CP-SHELL-PAGE) |
| `Stepper` | `NewStepper(id, steps []StepItem, x, y, w, h)` | Multi-step wizard indicator |
| `Pagination` | `NewPagination(id, totalPages, current *Signal[int], x, y, w, h)` | Paged content control |
| `Accordion` | `NewAccordion(id, title, x, y, w, h)` | Collapsible sections |

Shell placement: [shells.md](shells.md).

---

## Demos

| Widget | Scene title (exact) |
|--------|---------------------|
| Navigation rail + desktop shell | **Desktop Shell (Go)** |
| App bar | **App Shell (Go)** |
| Stepper | **Batch 4 · Stepper** |
| Pagination | **Batch 13 · Pagination** |
| Accordion | **Batch 3b · Accordion** |

**Related:** [../DEMO_INDEX.md](../DEMO_INDEX.md) · [../COMPOSITION.md](../COMPOSITION.md)

---

## Tips

- `NavigationRail` shares item type with bottom nav (`BottomNavItem`) — see Desktop Shell demo wiring.
- Steppers and accordions usually live inside `Panel` hosts on scroll pages.
- Pagination needs a shared `*Signal[int]` for current page and bound content region.
