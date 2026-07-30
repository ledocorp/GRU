# Layout

Composition is a stack from window → shell → surface → widgets. Height modes and Viewport vs Container are the usual sources of demo-shaped hacks — name the mode per node and pick the right scroll owner.

---

## Composition stack

```text
Document.Root          LayoutAbsolute (window-sized)
└── Page / app shell   flex column or row (window inset)
      └── Grid (opt)   12-col; row height policy
            └── Viewport (opt)  scroll + clip in assigned rect
                  └── Panel / Card
                        └── flex / controls
```

- Dropdowns / modals / overlays draw above and must **not** change sibling layout extent.
- `SetChromeTop` / `SetChromeBottom` reserve Document bands. In-tree footers (`StatusBar`) are children, not automatic chrome bottom unless you set it.

---

## Container vs Viewport

| | `Container` | `Viewport` |
|---|-------------|------------|
| Scroll | No | Yes (wheel) |
| Scissor | Optional `ClipChildren` | Always |
| Use | Toolbars, forms, grouping | Scroll panels, page body |

- Always use `viewport.AddChild` (sets parent correctly). Do not attach via a bare `Element.AddChild` path that skips Viewport parenting — clip / hit-test break.
- raylib scissor is **not stackable**. Viewport re-applies scissor after each child Draw.
- Open dropdown lists use an **overlay** group drawn after scissor ends.
- Wheel: one owner per gesture (`PrepareWheelScroll`); nested viewports can absorb at limits.

---

## Three height modes

| Mode | Typical setup | Meaning |
|------|---------------|---------|
| **Intrinsic** | `h=0`, auto-height, `FlexGrow=0` | Shrink-wrap at assigned width |
| **Fill** | `FlexGrow > 0` (often `h=0`) | Take remaining main-axis space |
| **Fixed** | `h > 0`, auto-height off | Exact band; clamp + scissor |

Rules:

1. Measure width-first.
2. Internal probe heights must never stick as the final `bounds.Height`.
3. Intrinsic wins when content grows; fill wins on an assigned band.
4. Overlay children are excluded from layout extent measurement.

---

## Flex & grid

- Layout modes: flex, grid, responsive, absolute, none.
- `FlexGrow` shares remaining space. Buttons with `w=0` shrink-wrap preferred width — they do not full-bleed unless you intend it.
- Grid row sizing: shrink-wrap by default; **`GridRowSizingEqualFill`** is an explicit opt-in.
- Breakpoints: xs &lt; 480 … xl ≥ 1280. Below `MinClientWidth`, grids force the tightest column span.

---

## Public shell recipes

| Need | Helper / pattern | Demo |
|------|------------------|------|
| Scrollable page | `MountAppPage` — one page-scroll Viewport | Form Demo, batches |
| App shell | `MountAppShellRoot` — pinned header + flex body | App Shell (Go), WebView Module |
| Desktop | `MountDesktopPageShell` — rail \| main | Desktop Shell (Go) |
| Edge utility | `MountEdgeToEdgeRoot` — menubar + status | List Pane, Settings · Desktop |
| Equal-fill splits | Equal-fill grid opt-in | Responsive / batch grids |

Do not `SetStyle("transparent")` on the main page-scroll Viewport — padding/gutter math depends on the style.

Default for new apps: flex column + `Card` / `Panel` (`cmd/hello`) or **CP-SHELL-PAGE** when using demo helpers. See [../COMPOSITION.md](../COMPOSITION.md).

---

## Resize contract (summary)

On client size change: Document resize → fit root children → relayout **shell** containers → sync flex layout.

- Mount the main Viewport at **0×0 + FlexGrow(1)** — never bake window height into the Viewport at construction.
- Same-size “resize” at min width still relayouts shells without unloading the whole cache.
- `ForceFullLayout` is for chrome toggle / recovery — not the hot drag path.

---

## Panel / Card body

The shell receives `(W, H)`; the body gets one flex pass; children are clamped/scissored to the padded body ∩ ancestor viewports. Auto-height finalizes from the subtree — probe heights must not remain.

---

## WebView layout (teaser)

Live HWNDs are **not** clipped by Viewport scissor — only COM bounds sync. Clamp top to `ChromeTop`. For fill-height embeds, put `WebViewPanel` as a **direct** flex child of a grow row (`MountAppShellRoot`), not inside a titled Panel + page-scroll. Full-client: `FillClient=true`. Details: [webview.md](webview.md).

---

## Do / Don’t

**Do**

- Name the height mode per node  
- Prefer Viewport for scroll; Container + clip for non-scroll clips  
- Keep page-scroll padding engine-owned  

**Don’t**

- Nest live WebView in page-scroll for fill layouts  
- Infer equal-fill from “all heights are zero”  
- Use `SetBounds` during Layout for internal moves  
- Invent underlap cameras / corner hacks for chrome  

---

## Samples

grudemo: **Desktop Shell (Go)**, **App Shell (Go)**, **List Pane (Go)**, **Responsive - Breakpoints - Grid**, **Card Nest (Go)**, **Form Demo**.

Next: [rendering.md](rendering.md) · [../widgets/shells.md](../widgets/shells.md)
