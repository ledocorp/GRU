# Shells & page chrome

**Purpose:** Shell helpers define how your app fills the document content band below the native title bar. Pick one shell recipe per scene or app — do not mix mount patterns casually.

**Related:** [../COMPOSITION.md](../COMPOSITION.md) · [surfaces.md](surfaces.md) · [../GETTING_STARTED.md](../GETTING_STARTED.md)

---

## Shell matrix

| Shell ID | Helper | Use when |
|----------|--------|----------|
| **CP-SHELL-PAGE** | `MountAppPage` / `MountFlexPageShell` | Scrollable demo or document pages (most batch scenes) |
| **CP-SHELL-APPSHELL** | `MountAppShellRoot` | Mobile-first AppBar + scroll body |
| **CP-SHELL-DESKTOP** | `MountDesktopPageShell` | MenuBar + nav rail + main workspace |
| **CP-SHELL-EDGE** | `MountEdgeToEdgeRoot` | Full-bleed flex chrome (menubar + status rows) |
| **CP-SHELL-GRID** | `MountPageGrid` | Fixed 12-column responsive grids |

Mount helpers live in `examples/page_shell.go` (demo catalog). Standalone samples such as **`cmd/hello`** use a raw flex column on `doc.Root` instead — that is the CP-SHELL-PAGE *spirit* without importing `examples`.

---

## Key types & constructors

### Page scroll shell

```go
page := examples.MountAppPage(doc, "my-page", "Title", "Subtitle")
// page.Body  — scroll viewport; add grids, Panel, Card here
// page.Frame — full-client flex wrapper (resize contract)
```

| Helper | Returns | Notes |
|--------|---------|-------|
| `MountAppPage(doc, id, title, subtitle)` | `AppPage` | Header omitted when `title == ""` |
| `MountFlexPageShell(doc, id)` | `(*Container, *Viewport)` | Page scroll without title header |
| `MountSceneHeader(vp, id, title, subtitle)` | `*Header` | Legacy: header inside an existing viewport |

### App shell (mobile)

```go
mount := examples.MountAppShellRoot(doc, "appshell")
vp := examples.NewAppShellScrollViewport("scroll")
mount.Shell.AddChild(appBar)
mount.Shell.AddChild(vp) // flex-grow 1
```

| Type | Constructor | Role |
|------|-------------|------|
| `AppBar` | `ui.NewAppBar(id, title, x, y, w, h)` | Top app bar row |
| `Viewport` | `examples.NewAppShellScrollViewport(id)` | `page-scroll` body between pinned rows |

### Desktop shell

```go
shell, workspace := examples.MountDesktopPageShell(doc, "desktop")
// Add MenuBar / StatusBar as direct shell children (full width)
// Add NavigationRail + main column to workspace (flex row)
```

| Type | Constructor | Role |
|------|-------------|------|
| `MenuBar` | `ui.NewMenuBar(id, menus, x, y, w, h)` | Top menu strip |
| `NavigationRail` | `ui.NewNavigationRail(id, items, selected, x, y, w, h)` | Left nav rail |
| `StatusBar` | `ui.NewStatusBar(id, x, y, w, h)` | Bottom status row |

### Edge-to-edge shell

```go
root := examples.MountEdgeToEdgeRoot(doc, "edge", false) // false = column
// Add MenuBar, body rows, StatusBar as direct children
```

Use for utility layouts that need menubar + status without the desktop rail recipe.

### Responsive grid shell

```go
grid := examples.MountPageGrid(doc, "grid")
// or: grid := ui.NewPageGrid(id, width, height)
child.SetColSpan(ui.BreakpointMD, 6)
```

| Type | Constructor | Role |
|------|-------------|------|
| `PageGrid` | `ui.NewPageGrid(id, width, height)` | 12-column breakpoint-aware grid |

### Finish pass

Call `examples.FinishShellMount(doc)` once after `Build` + `Document.Resize` in the demo harness. Scenes should not call it from `Build` directly.

---

## Demo links

| Goal | Scene title (exact) |
|------|---------------------|
| Desktop menubar + rail | **Desktop Shell (Go)** |
| AppBar + scroll body | **App Shell (Go)** |
| Desktop settings rows | **Settings · Desktop (Go)** |
| Mobile settings in App Shell | **Settings (Go)** |
| Edge shell + split list | **List Pane (Go)** |
| Breakpoint grid reflow | **Responsive - Breakpoints - Grid** |
| Minimal standalone (no examples import) | `go run ./cmd/hello` |

---

## Pitfalls

| Mistake | Fix |
|---------|-----|
| Calling `FinishShellMount` inside `Build` | Let the host call it after resize |
| Nesting a second full-page viewport inside `MountAppPage.Body` | Add content directly to `page.Body` |
| Putting MenuBar beside the rail instead of above workspace | MenuBar is a **shell** child, full width |
| Using `MountAppShellRoot` for desktop rail layouts | Use `MountDesktopPageShell` |
| Hardcoded absolute positions for page body | Flex column/row + `SetFlexGrow(1)` |
| Inventing a new shell wrapper type | Extend an existing mount helper or copy `cmd/hello` flex root |

---

## Composition notes

1. **Window chrome is outside the tree** — the native `TitleBar` is owned by `Document`; shell helpers only manage `doc.Root` children.
2. **Scroll ownership** — CP-SHELL-PAGE and App Shell scroll viewports use the `page-scroll` style; pinned chrome (AppBar, MenuBar) stays outside the scroll band.
3. **One job per layer:** shell → (optional grid) → **Panel / Card** → widgets. See [surfaces.md](surfaces.md).
4. **Resize contract** — shells size to `doc.Width` / `doc.Height` on mount; rely on `Document.Resize` + `FinishShellMount` for stable intrinsic heights.
5. For WebView-heavy apps, prefer app-shell flex rows with `WebViewPanel` as a direct flex child — see [webview.md](webview.md) and [../WEBVIEW.md](../WEBVIEW.md).
