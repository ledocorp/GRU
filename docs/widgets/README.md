# Widgets (public encyclopedia)

Gru’s interactive widgets live in package **`ui`**. Learn them from **curated demos** (`go run -tags grudemo .`) and the standalone samples **`cmd/hello`** and **`cmd/webviewhello`**.

**Related:** [../GETTING_STARTED.md](../GETTING_STARTED.md) · [../COMPOSITION.md](../COMPOSITION.md) · [../DEMO_INDEX.md](../DEMO_INDEX.md) · [../architecture/README.md](../architecture/README.md)

---

## How to explore

```bash
go run ./cmd/hello                              # minimal Button + Card sample
go run -tags grudemo .                          # 40-scene catalog (Tab = next scene)
go run -tags "grudemo,webview2" .               # live WebView demos (Windows)
```

| Key | Action |
|-----|--------|
| **Tab** | Next scene |
| **F12** | Widget Inspector |
| Footer **Scenes** | Open **Demo Index** |

---

## Scene → widget families (40 scenes)

Titles match `Scene.Title()` exactly — see [../DEMO_INDEX.md](../DEMO_INDEX.md).

| # | Scene title | Primary widget families | Doc page |
|---|-------------|-------------------------|----------|
| 1 | Demo Index | — (scene picker) | — |
| 2 | Desktop Shell (Go) | [Shells](shells.md) | [shells.md](shells.md) |
| 3 | App Shell (Go) | [Shells](shells.md) | [shells.md](shells.md) |
| 4 | Settings · Desktop (Go) | [Shells](shells.md), [List tile](list-tile.md) | [list-tile.md](list-tile.md) |
| 5 | List Pane (Go) | [List pane](list-pane.md), [List tile](list-tile.md), [Surfaces](surfaces.md) | [list-pane.md](list-pane.md) |
| 6 | Filters (Go) | [Filters / toolbar](filters-toolbar.md) | [filters-toolbar.md](filters-toolbar.md) |
| 7 | Form Demo | [Forms](forms.md) | [forms.md](forms.md) |
| 8 | Card Nest (Go) | [Surfaces](surfaces.md) | [surfaces.md](surfaces.md) |
| 9 | Responsive - Breakpoints - Grid | [Shells](shells.md), [Surfaces](surfaces.md) | [shells.md](shells.md) |
| 10 | Settings (Go) | [Shells](shells.md), [List tile](list-tile.md) | [list-tile.md](list-tile.md) |
| 11 | Timeline (Go) | [Data display](data-display.md) | [data-display.md](data-display.md) |
| 12 | Counter Demo | [Buttons & signals](buttons-signals.md) | [buttons-signals.md](buttons-signals.md) |
| 13 | Theme v2 Foundation | [Theme v2](theme-v2.md) | [theme-v2.md](theme-v2.md) |
| 14 | Gallery (.gru) | [Document / .gru](document-gru.md) | [document-gru.md](document-gru.md) |
| 15 | Batch 1 · Tooltip / TabView / Modal | [Overlays](overlays.md) | [overlays.md](overlays.md) |
| 16 | Batch 2 · SearchBar | [Search](search.md) | [search.md](search.md) |
| 17 | Batch 3 · Badge | [Data display](data-display.md) | [data-display.md](data-display.md) |
| 18 | Batch 3b · Accordion | [Navigation widgets](navigation-widgets.md) | [navigation-widgets.md](navigation-widgets.md) |
| 19 | Batch 3 · DatePicker | [Forms](forms.md) | [forms.md](forms.md) |
| 20 | Batch 4 · Stepper | [Navigation widgets](navigation-widgets.md) | [navigation-widgets.md](navigation-widgets.md) |
| 21 | Batch 6 · DataTable | [Data display](data-display.md) | [data-display.md](data-display.md) |
| 22 | Batch 7 · Toast / Notification | [Overlays](overlays.md) | [overlays.md](overlays.md) |
| 23 | Batch 11 · DateRangePicker | [Forms](forms.md), [Filters / toolbar](filters-toolbar.md) | [forms.md](forms.md) |
| 24 | Batch 12 · Rating | [Selection](selection.md) | [selection.md](selection.md) |
| 25 | Batch 13 · Pagination | [Navigation widgets](navigation-widgets.md) | [navigation-widgets.md](navigation-widgets.md) |
| 26 | Batch 14 · SegmentedControl | [Selection](selection.md) | [selection.md](selection.md) |
| 27 | Batch 15 · ComboBox | [Forms](forms.md), [Selection](selection.md) | [forms.md](forms.md) |
| 28 | Batch 16 · SpinBox | [Forms](forms.md) | [forms.md](forms.md) |
| 29 | Batch 17 · Breadcrumbs | [Navigation widgets](navigation-widgets.md) | [navigation-widgets.md](navigation-widgets.md) |
| 30 | Batch 18 · ColorWell | [Filters / toolbar](filters-toolbar.md) | [filters-toolbar.md](filters-toolbar.md) |
| 31 | Batch 19 · Dropdown | [Forms](forms.md), [Selection](selection.md) | [forms.md](forms.md) |
| 32 | Batch 20 · RadioGroup | [Selection](selection.md) | [selection.md](selection.md) |
| 33 | Batch 21 · ListTile | [List tile](list-tile.md), [Surfaces](surfaces.md) | [list-tile.md](list-tile.md) |
| 34 | Batch 22 · Toggle | [Selection](selection.md), [List tile](list-tile.md) | [selection.md](selection.md) |
| 35 | Batch 23 · Checkbox | [Selection](selection.md) | [selection.md](selection.md) |
| 36 | Batch 24 · Slider | [Sliders & progress](sliders-progress.md) | [sliders-progress.md](sliders-progress.md) |
| 37 | Batch 25 · ProgressBar | [Sliders & progress](sliders-progress.md) | [sliders-progress.md](sliders-progress.md) |
| 38 | Batch 26 · ColorPicker | [Forms](forms.md), [Filters / toolbar](filters-toolbar.md) | [forms.md](forms.md) |
| 39 | WebView Module Demo | [WebView](webview.md) | [webview.md](webview.md) |
| 40 | WebView Focus Handoff | [WebView](webview.md) | [webview.md](webview.md) |

---

## Family index

| Family | Page | Key types |
|--------|------|-----------|
| Shells & page chrome | [shells.md](shells.md) | `MountAppPage`, `MountAppShellRoot`, `MountDesktopPageShell`, `MountEdgeToEdgeRoot`, `MountPageGrid` |
| Surfaces | [surfaces.md](surfaces.md) | `Panel`, `Card` |
| List rows | [list-tile.md](list-tile.md) | `ListTile` |
| Master–detail sidebar | [list-pane.md](list-pane.md) | `NewListPane` (examples helper), `SplitView` |
| Buttons & reactive state | [buttons-signals.md](buttons-signals.md) | `Button`, `IconButton`, `Signal`, `Effect` |
| Forms & inputs | [forms.md](forms.md) | `Form`, `TextInput`, `DatePicker`, `ComboBox`, `SpinBox` |
| Overlays | [overlays.md](overlays.md) | `ShowModal`, `ShowToast`, `TabView`, `Tooltip` |
| Selection controls | [selection.md](selection.md) | `Toggle`, `Checkbox`, `SegmentedControl`, `Rating` |
| Sliders & progress | [sliders-progress.md](sliders-progress.md) | `Slider`, `ProgressBar` |
| WebView embed | [webview.md](webview.md) | `WebViewPanel` — full guide: [../WEBVIEW.md](../WEBVIEW.md) |
| Search | [search.md](search.md) | `SearchBar` |
| Navigation widgets | [navigation-widgets.md](navigation-widgets.md) | `NavigationRail`, `AppBar`, `Stepper`, `Pagination`, `Accordion` |
| Data display | [data-display.md](data-display.md) | `DataTable`, `Badge`, `Label`, `RichText` |
| Filters & toolbar rows | [filters-toolbar.md](filters-toolbar.md) | `ComboBox`, `DateRangePicker`, `ColorWell` |
| Theme v2 | [theme-v2.md](theme-v2.md) | `SetStyleVariant`, component keys |
| Document / `.gru` | [document-gru.md](document-gru.md) | `BuildContext`, DocumentSpec |

---

## Golden rules

1. **Surfaces host lists** — put dense `ListTile` stacks inside a `Panel` or `Card`, not a bare styled `Container`.
2. **No new wrapper widgets** — use `Panel` / `Card` + `ListTile` instead of inventing sidebar or list types.
3. **Borrow controls, not whole apps** — copy a widget pattern from a demo scene; do not copy private product chrome.
4. **One shell per scene** — pick one mount helper from [shells.md](shells.md) and [../COMPOSITION.md](../COMPOSITION.md); do not mix shell recipes casually.
5. **Theme keys over ad-hoc colors** — prefer `card`, `panel`, `list-tile`, `form-label`, `form-value`, `button`, `primary`.
6. **Never `SetStyle("transparent")` on `ListTile`** to “fix” nesting — adjust the parent surface instead.
7. **Modals are flat flex bodies** — do not use a `Viewport` as the modal root ([overlays.md](overlays.md)).
8. **WebView is a separate compositor** — read [../WEBVIEW.md](../WEBVIEW.md) before embedding live browser content.

---

## Standalone samples (outside the 40-scene list)

| Sample | Command | What it shows |
|--------|---------|---------------|
| Hello Gru | `go run ./cmd/hello` | Flex root + `Card` + `Button` + `Signal` |
| WebView hello | `go run -tags webview2 ./cmd/webviewhello` | `FillClient` WebView shell |

Scaffold new apps: `go run ./cmd/gru new myapp` or `go run ./cmd/gru new myapp --webview` — see [../GETTING_STARTED.md](../GETTING_STARTED.md).

---

## API reference

For field-level godoc: `go doc github.com/ledocorp/gru/ui` or your IDE’s symbol search on `ui`.
