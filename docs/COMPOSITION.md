# Composition (public)

**Audience:** app authors building on Gru  
**Shell helpers:** `examples/page_shell.go` (demo catalog) · samples use raw `ui` flex + Card  

Go-first composition rules for public apps. Widget catalog: [widgets/README.md](widgets/README.md). System layout contracts: [architecture/layout.md](architecture/layout.md).

---

## Golden rules

1. **Go-first** — compose widgets in Go; `.gru` is optional structure.  
2. **One job per layer** — shell → (optional viewport) → **Panel / Card** → flex children / widgets.  
3. **Surfaces host lists** — put dense `ListTile` stacks in a `Panel` or `Card`, not a bare `Container` with ad-hoc styles.  
4. **No new wrapper widgets** when `Panel`, `Card`, `Container`, and `ListTile` already cover the case.  
5. **Borrow widgets, not whole apps** — copy a control pattern from a demo; do not copy private product chrome.  
6. **Name the height mode** — intrinsic, fill (`FlexGrow`), or fixed — see [architecture/layout.md](architecture/layout.md).

---

## Shell matrix (pick one)

| Shell ID | Helper | Use when | Demo |
|----------|--------|----------|------|
| **CP-SHELL-PAGE** | `MountAppPage` / `MountFlexPageShell` | Scrollable demo / document pages | Form Demo, most batches |
| **CP-SHELL-APPSHELL** | `MountAppShellRoot` | Mobile-first AppBar + scroll / fill body | App Shell (Go), WebView Module |
| **CP-SHELL-DESKTOP** | `MountDesktopPageShell` | MenuBar + nav rail + main | Desktop Shell (Go) |
| **CP-SHELL-EDGE** | `MountEdgeToEdgeRoot` | Full-bleed utility chrome (menubar + status) | List Pane (Go), Settings · Desktop |
| **CP-SHELL-GRID** | `MountPageGrid` | Fixed 12-col grids | Responsive - Breakpoints - Grid |

**Default for new apps / samples:** flex column root + `Card` / `Panel` (`samples/hello`) or **CP-SHELL-PAGE** when using demo helpers.

More detail: [widgets/shells.md](widgets/shells.md).

---

## Surfaces

| Widget | Role |
|--------|------|
| `Panel` | Framed region, optional title / scroll features |
| `Card` | Section surface inside a page |
| `ListTile` | Row in a list (title, subtitle, trailing) |
| `Container` | Layout grouping — not a styled “fake panel” |

Prefer theme styles (`card`, `panel`, `list-tile`, `form-label`, `form-value`) over one-off colors.

| Pattern | Do | Don’t |
|---------|----|--------|
| Settings rows | `ListTile` + trailing `Toggle` / `ComboBox` inside `Panel` | Custom “SettingsRow” widget type |
| Master–detail | Edge shell + list surface + detail `Card` | Ad-hoc absolute stacks |
| Nested callouts | `Card` variants (`callout`, `code`) | Deep Card-in-Card without purpose |

See [widgets/surfaces.md](widgets/surfaces.md) · [widgets/list-tile.md](widgets/list-tile.md) · [widgets/list-pane.md](widgets/list-pane.md).

---

## Borrow matrix (public)

Copy **widgets and small control patterns** from allowlisted demos. Do **not** copy private app shells, stores, or product chrome.

| You want | Borrow from | Avoid |
|----------|-------------|--------|
| Page + cards | Card Nest (Go) | Inventing surface wrappers |
| Desktop rail | Desktop Shell (Go) | Private monorepo directory hubs |
| AppBar + FAB | App Shell (Go) | Private mobile product chrome |
| Desktop settings rows | Settings · Desktop (Go) | Mobile Settings demo as desktop recipe |
| List + detail | List Pane (Go) | Private note-app sidebars |
| Forms | Form Demo + field batches | One mega-form without Panels |
| Overlays | Batch 1 · Modal; Batch 7 · Toast | `Viewport` as modal root |
| WebView embed | WebView Module / Focus Handoff | Passthrough / OS mouse tricks — [WEBVIEW.md](WEBVIEW.md) |
| Signals starter | Counter Demo | Rebuilding the tree every frame |

---

## Modals & overlays

- Modal body = **flat flex** only — **never** a `Viewport` as the modal root.  
- Toasts / tooltips / context menus use overlay managers (after blit).  
- Modal / drawer over WebView → host occlusion (hide HWND), not click-through.

Demos: **Batch 1 · Tooltip / TabView / Modal**, **Batch 7 · Toast / Notification**.  
Guide: [widgets/overlays.md](widgets/overlays.md).

---

## Forbidden (public apps)

- Custom list/sidebar wrapper types instead of `Panel` / `Card`  
- `SetStyle("transparent")` on `ListTile` to “fix” nesting  
- Improvised theme keys when `panel` / `card` / `list-tile` suffice  
- Copying private product layouts wholesale  
- Live `WebViewPanel` inside a page-scroll Viewport for fill-height layouts  

---

## Learn by demo

| Goal | Scene |
|------|--------|
| Page + cards | Card Nest (Go) |
| Desktop rail | Desktop Shell (Go) |
| AppBar shell | App Shell (Go) |
| Settings (desktop) | Settings · Desktop (Go) |
| List sidebar | List Pane (Go) |
| Form fields | Form Demo |
| Signals + button | Counter Demo |
| WebView | WebView Module Demo / Focus Handoff |
| Minimal standalone | `go run ./cmd/hello` |
| WebView standalone | `go run -tags webview2 ./cmd/webviewhello` |

Full title list: [DEMO_INDEX.md](DEMO_INDEX.md). Widget map: [widgets/README.md](widgets/README.md).
