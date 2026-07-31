# Gru cheatsheet (v0.7.1)

Fingertip lookup for the public toolkit **as shipped in v0.7.1**. Guides teach; this page answers *what exists* and *where to jump*.

**Module:** `github.com/ledocorp/gru` · **Package:** `ui`  
**Site:** [ledocorp.org/gru/docs/cheatsheet](https://www.ledocorp.org/gru/docs/cheatsheet/)  
**Deeper:** [GETTING_STARTED](GETTING_STARTED.md) · [COMPOSITION](COMPOSITION.md) · [widgets/](widgets/README.md) · [api/](api/README.md) · [DEMO_INDEX](DEMO_INDEX.md)

**Status legend**

| Status | Meaning |
|--------|---------|
| `ok` | Shipped with a usable demo and/or doc pointer |
| `demo-thin` | Type exists; demo coverage is weak or indirect |
| `doc-thin` | Type/demo exist; narrative doc is thin |
| `api-thin` | Usable in code; godoc / api narrative thin |
| `missing` | Need acknowledged; not in this cut |
| `private` | Exists in private products; not public |

Scan Status ≠ `ok` before a release. That list *is* the gap backlog.

---

## 1. Commands

```bash
go get github.com/ledocorp/gru@v0.7.1
go run ./cmd/hello
go run -tags webview2 ./cmd/webviewhello
go run -tags grudemo .
go run ./cmd/gru new myapp
```

| Tag / flag | When |
|------------|------|
| `grudemo` | Curated 40-scene catalog |
| `webview2` | Live WebView2 host (Windows + Runtime) |
| `x11` | Linux hello path |
| `gru new --webview` | Scaffold FillClient WebView app |
| `gru build demo` | Package the demo launcher |

CLI details: [CLI.md](CLI.md).

---

## 2. Frame loop

```text
DrainQueue
Update (title bar + Root)
Layout if Root dirty
Draw into SSAA if NeedsRedraw, else blit cache
Present + overlays
[WebView: sync hosts after Layout; present after EndDrawing]
```

Copy [`cmd/hello`](../cmd/hello). Do not invent a loop from scratch.  
Contracts: [architecture/overview.md](architecture/overview.md).

---

## 3. Intent index (I need…)

| Intent | Use | See | Status |
|--------|-----|-----|--------|
| Minimal window + button | `cmd/hello` · `Card` + `Button` + `Signal` | hello sample | ok |
| Scaffold an app | `gru new myapp` | [CLI.md](CLI.md) | ok |
| HTML pane in chrome | `cmd/webviewhello` · FillClient WebView | webviewhello · WebView demos | ok |
| Settings list row | `ListTile` + trailing control in `Panel`/`Card` | Settings · Desktop (Go) | ok |
| Mobile-style settings | App Shell + `ListTile` | Settings (Go) | ok |
| Master / detail list | Edge shell + list surface + detail | List Pane (Go) | ok |
| Desktop rail layout | `CP-SHELL-DESKTOP` / `MountDesktopPageShell` | Desktop Shell (Go) | ok |
| AppBar + scroll body | `CP-SHELL-APPSHELL` / `MountAppShellRoot` | App Shell (Go) | ok |
| Scrollable page + cards | `CP-SHELL-PAGE` / flex + `Card` | Form Demo · Card Nest | ok |
| 12-col grid page | `CP-SHELL-GRID` / `MountPageGrid` | Responsive - Breakpoints - Grid | ok |
| Nested cards / callouts | `Card` variants | Card Nest (Go) | ok |
| Form fields | `Form` · `TextInput` · related inputs | Form Demo | ok |
| Date pick | `DatePicker` | Batch 3 · DatePicker | ok |
| Date range | `DateRangePicker` | Batch 11 · DateRangePicker | ok |
| Combo / select | `ComboBox` | Batch 15 · ComboBox | ok |
| Numeric spin | `SpinBox` | Batch 16 · SpinBox | ok |
| Toggle / checkbox | `Toggle` · `Checkbox` | Batch 22 / 23 | ok |
| Segmented control | `SegmentedControl` | Batch 14 | ok |
| Rating stars | `Rating` | Batch 12 | ok |
| Slider / progress | `Slider` · `ProgressBar` | Batch 24 / 25 | ok |
| Search field | `SearchBar` | Batch 2 · SearchBar | ok |
| Filter toolbar row | filters helpers · chips | Filters (Go) | ok |
| Modal dialog | flat flex modal body (**not** `Viewport` root) | Batch 1 · Modal | ok |
| Toast / notification | toast APIs | Batch 7 · Toast | ok |
| Tooltip / tabs | `Tooltip` · `TabView` | Batch 1 | ok |
| Accordion | `Accordion` | Batch 3b | ok |
| Stepper | `Stepper` | Batch 4 | ok |
| Pagination | `Pagination` | Batch 13 | ok |
| Data table | `DataTable` | Batch 6 | ok |
| Badge / timeline | `Badge` · timeline composition | Batch 3 · Timeline | ok |
| Theme variants | theme v2 keys / `SetStyleVariant` | Theme v2 Foundation | ok |
| Load `.gru` page | DocumentSpec / Gallery | Gallery (.gru) | ok |
| Reactive counter | `Signal` + `Button` | Counter Demo | ok |
| Inspect widget tree | F12 Inspector | any grudemo | ok |
| Embed live WebView | `-tags webview2` · `WebViewPanel` | WebView Module · Focus Handoff | ok |
| Modal over WebView | `ShowModal` (engine occludes HWND) | Focus Handoff · [WEBVIEW.md](WEBVIEW.md) | ok |
| Breadcrumb path | `Breadcrumbs` | Batch 17 · Breadcrumbs | ok |
| Color well / picker | `ColorWell` · `ColorPicker` | Batch 18 · Batch 26 | ok |
| Dropdown select | `Dropdown` | Batch 19 · Dropdown | ok |
| Radio group | `RadioGroup` | Batch 20 · RadioGroup | ok |
| Height / scroll rules | intrinsic · fill · fixed | [LAYOUT.md](LAYOUT.md) | ok |

---

## 4. Capability matrices

### Shells

| Symbol | Demo | Doc | Status |
|--------|------|-----|--------|
| `CP-SHELL-PAGE` / `MountAppPage` | Form Demo | [widgets/shells.md](widgets/shells.md) · [COMPOSITION.md](COMPOSITION.md) | ok |
| `CP-SHELL-APPSHELL` / `MountAppShellRoot` | App Shell (Go) · Settings (Go) | shells.md | ok |
| `CP-SHELL-DESKTOP` / `MountDesktopPageShell` | Desktop Shell (Go) | shells.md | ok |
| `CP-SHELL-EDGE` / `MountEdgeToEdgeRoot` | List Pane · Settings · Desktop | shells.md | ok |
| `CP-SHELL-GRID` / `MountPageGrid` | Responsive - Breakpoints - Grid | shells.md | ok |

### Surfaces

| Symbol | Demo | Doc | Status |
|--------|------|-----|--------|
| `Panel` | List Pane · Settings · Desktop | [widgets/surfaces.md](widgets/surfaces.md) | ok |
| `Card` | Card Nest · hello | surfaces.md | ok |
| `ListTile` | Batch 21 · Settings rows | [widgets/list-tile.md](widgets/list-tile.md) | ok |
| `Container` | (layout only) | architecture/layout.md | ok |
| List pane helper | List Pane (Go) | [widgets/list-pane.md](widgets/list-pane.md) | ok |
| `SplitView` | List Pane (indirect) | list-pane.md | demo-thin |

### Controls

| Symbol | Demo | Doc | Status |
|--------|------|-----|--------|
| `Button` · `IconButton` | Counter · hello | [buttons-signals.md](widgets/buttons-signals.md) | ok |
| `Signal` · `Effect` | Counter Demo | [api/signals.md](api/signals.md) | ok |
| `Form` · `TextInput` | Form Demo | [forms.md](widgets/forms.md) | ok |
| `DatePicker` | Batch 3 · DatePicker | forms.md | ok |
| `DateRangePicker` | Batch 11 | forms.md · filters-toolbar.md | ok |
| `ComboBox` | Batch 15 · Filters | forms.md | ok |
| `SpinBox` | Batch 16 | forms.md | ok |
| `Toggle` | Batch 22 | [selection.md](widgets/selection.md) | ok |
| `Checkbox` | Batch 23 | selection.md | ok |
| `SegmentedControl` | Batch 14 | selection.md | ok |
| `Rating` | Batch 12 | selection.md | ok |
| `Slider` | Batch 24 | [sliders-progress.md](widgets/sliders-progress.md) | ok |
| `ProgressBar` | Batch 25 | sliders-progress.md | ok |
| `SearchBar` | Batch 2 | [search.md](widgets/search.md) | ok |
| `ShowModal` / modal body | Batch 1 | [overlays.md](widgets/overlays.md) | ok |
| `ShowToast` / notifications | Batch 7 | overlays.md | ok |
| `Tooltip` · `TabView` | Batch 1 | overlays.md | ok |
| `Accordion` | Batch 3b | [navigation-widgets.md](widgets/navigation-widgets.md) | ok |
| `Stepper` | Batch 4 | navigation-widgets.md | ok |
| `Pagination` | Batch 13 | navigation-widgets.md | ok |
| `NavigationRail` · `AppBar` | Desktop / App Shell | navigation-widgets.md · shells.md | ok |
| `DataTable` | Batch 6 | [data-display.md](widgets/data-display.md) | ok |
| `Badge` · `Label` · `RichText` | Batch 3 · Timeline | data-display.md | ok |
| `ColorWell` | Batch 18 · ColorWell | filters-toolbar.md | ok |
| `ColorPicker` | Batch 26 · ColorPicker | forms.md · filters-toolbar.md | ok |
| `Dropdown` | Batch 19 · Dropdown | forms.md | ok |
| `RadioGroup` | Batch 20 · RadioGroup | selection.md | ok |
| `Breadcrumbs` | Batch 17 · Breadcrumbs | navigation-widgets.md | ok |
| Theme v2 / `SetStyleVariant` | Theme v2 Foundation | [theme-v2.md](widgets/theme-v2.md) | ok |
| DocumentSpec / `.gru` | Gallery (.gru) | [document-gru.md](widgets/document-gru.md) | ok |

### Core API

| Symbol | Demo | Doc | Status |
|--------|------|-----|--------|
| `Document` / `NewDocument` | hello | [api/document.md](api/document.md) | ok |
| `Node` / `Element` / dirty | any | [api/node-element.md](api/node-element.md) · architecture/overview | ok |
| Chrome / `TitleBar` / borderless mount | hello · shells | [api/mount-chrome.md](api/mount-chrome.md) | ok |
| `DrainQueue` / `QueueMain` | hello | document.md | ok |
| `NeedsRedraw` / SSAA blit | (engine) | architecture/rendering.md | ok |
| Focus routing | Focus Handoff · overlays | architecture/input-focus.md | ok |
| Viewport scroll | many shells | architecture/layout.md | ok |

### WebView

| Symbol | Demo | Doc | Status |
|--------|------|-----|--------|
| `-tags webview2` | webviewhello · WebView demos | [WEBVIEW.md](WEBVIEW.md) | ok |
| FillClient shell | `cmd/webviewhello` | WEBVIEW.md · [api/webview-api.md](api/webview-api.md) | ok |
| `WebViewPanel` | WebView Module Demo | widgets/webview.md | ok |
| Focus handoff | WebView Focus Handoff | WEBVIEW.md | ok |
| Host occlusion over modals | Focus Handoff · `ShowModal` | WEBVIEW.md (occlusion howto) | ok |
| Linux / non-Windows live host | — | — | missing |

### Tooling

| Symbol | Demo | Doc | Status |
|--------|------|-----|--------|
| `gru new` / `build` / `package` | CLI | [CLI.md](CLI.md) | ok |
| F12 Inspector | grudemo | GETTING_STARTED | ok |
| Demo Index scene | Demo Index | [DEMO_INDEX.md](DEMO_INDEX.md) | ok |
| pkgsite / `go doc` | — | [api/README.md](api/README.md) | ok |
| CI link / doc drift checker | — | — | missing |

### Private (not this package)

| Symbol | Status |
|--------|--------|
| Gru Notepad / Notes | private |
| Prism | private |
| Studio harness / full catalog | private |
| Foundry | private |

---

## 5. Anti-patterns

1. Do not invent a frame loop; copy `cmd/hello`.
2. Do not use a `Viewport` as a modal root; modal body is flat flex.
3. Do not invent sidebar/list wrapper widgets; use `Panel`/`Card` + `ListTile`.
4. Do not call `Layout` from `Draw`.
5. Do not `SetStyle("transparent")` on `ListTile` to fake nesting.
6. Do not assume WebView click-through / passthrough; occlude the HWND when overlays need hits.
7. Do not forget `assets/fonts/` must resolve at runtime.
8. Do not mix shell recipes casually; pick one shell ID per scene.
9. Do not treat `@latest` as pinned; prefer `@v0.7.1` until you mean to float.
10. Do not document private apps as public capabilities.

---

## 6. Keys and env

| Key / env | Action |
|-----------|--------|
| **Tab** | Next grudemo scene |
| **F12** | Widget Inspector |
| **F11** | Perf overlay (optional) |
| Footer **Scenes** | Open Demo Index |
| **Shift+F11** | WebView stderr traces (when debug on) |
| `GRU_WEBVIEW_DEBUG=1` | WebView debug (alias `GORY_WEBVIEW_DEBUG`) |

---

## Maintainer notes

Before each public cut:

1. Every user-facing `ui` type appears once in a matrix.
2. Every grudemo scene title appears in a Demo column (or DEMO_INDEX link covers the rest).
3. No `ok` without a demo title and/or doc path.
4. Private products stay `private`, never `ok`.
5. Prefer fixing `demo-thin` / `doc-thin` over adding aspirational `missing` rows.

### Triage (closed in v0.7.1)

Closed in this release:

| Item | Was | Now |
|------|-----|-----|
| WebView modal occlusion howto | doc-thin | ok (`WEBVIEW.md`) |
| ColorWell / ColorPicker / Dropdown / RadioGroup / Breadcrumbs demos | deferred / demo-thin | ok (public catalog) |
| Slim layout contracts | internal-only | ok (`LAYOUT.md` + architecture/layout) |

Still accepted later (not blockers):

| Item | Status | Note |
|------|--------|------|
| `SplitView` standalone demo | demo-thin | Covered inside List Pane |
| FilePicker dedicated public scene | deferred | Demo exists; keep private until path UX reviewed |
| Non-Windows live WebView | missing | Windows-only host for now |
| Automated cheatsheet/CI drift check | missing | Nice-to-have |

---

*Gru (GRU) · v0.7.1 · MPL 2.0 · LedoCorp*
