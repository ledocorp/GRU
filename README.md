# Gru

**Version:** [0.7.0](https://github.com/ledocorp/GRU/releases/tag/v0.7.0)

**Gru (GRU)** is a Go-first UI toolkit (retained-mode, Raylib-backed) for building desktop applications.

**License:** [Mozilla Public License 2.0](LICENSE)  
**Third-party notices:** [NOTICE](NOTICE)  
**Copyright:** Shawn Londono and LedoCorp  
**Publisher:** [LedoCorp](https://ledocorp.org) · **Maintainer:** Shawn Londono

---

## What’s included

- Full **`ui/`** platform — widgets, layout, text, inspector  
- Curated **demo app** (`-tags grudemo`) — widget / composition demos, Tab to switch, **F12** Inspector  
- Optional **WebView2** demos (`-tags webview2`) on Windows  
- **`cmd/gru`** — scaffold, build, and package apps (`gru new`, `gru build`, `gru package`)  
- **Sample apps** — `cmd/hello` (native) and `cmd/webviewhello` (FillClient WebView shell)

## What’s not in this repository

- Gru Notepad / Gru Notes (private for now)  
- Prism (paused / private)  
- Full Studio testing harness and private catalog  
- Foundry R&D (internal)  

Commercial dual-licensing may come later; it is **not** part of this release.

---

## Module

```bash
go get github.com/ledocorp/gru@v0.7.0
```

```go
import "github.com/ledocorp/gru/ui"
```

Use `@latest` once you are comfortable tracking the newest tag.

## Quick start (Windows)

From the **repository root** (folder with `go.mod`, `assets/`, `pages/`):

```bash
go run ./cmd/hello
go run -tags webview2 ./cmd/webviewhello
go run -tags grudemo .
go run -tags "grudemo,webview2" .
```

Scaffold a new app:

```bash
go run ./cmd/gru new myapp
go run ./cmd/gru new myapp --webview
```

| Guide | Path |
|-------|------|
| Docs home | [docs/README.md](docs/README.md) |
| Getting started | [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md) |
| CLI | [docs/CLI.md](docs/CLI.md) |
| Composition | [docs/COMPOSITION.md](docs/COMPOSITION.md) |
| WebView | [docs/WEBVIEW.md](docs/WEBVIEW.md) |
| Architecture | [docs/architecture/README.md](docs/architecture/README.md) |
| Widgets | [docs/widgets/README.md](docs/widgets/README.md) |
| API | [docs/api/README.md](docs/api/README.md) |
| Demo index | [docs/DEMO_INDEX.md](docs/DEMO_INDEX.md) |
| Build & smoke | [docs/BUILD.md](docs/BUILD.md) |

| Key | Action |
|-----|--------|
| **Tab** | Next demo scene |
| **F12** | Inspector |
| **F11** | Perf overlay (optional) |

**Pull requests are not accepted** on this repository — see [CONTRIBUTING.md](CONTRIBUTING.md).

Primary target today: **Windows**.

---

## Open-source credits

Gru depends on these projects. Please respect their licenses. Full texts and notices: **[NOTICE](NOTICE)**.

| Project | Role | License (summary) | Link |
|---------|------|-------------------|------|
| **[Raylib](https://www.raylib.com/)** (via [raylib-go](https://github.com/gen2brain/raylib-go)) | Windowing & rendering | Zlib | https://www.raylib.com/ |
| **[Remix Icon](https://remixicon.com/)** | Icon font / atlas | Remix Icon License v1.0 | https://github.com/Remix-Design/RemixIcon |
| **Inter**, **Poppins**, **Fira Code**, **DM Mono** | Bundled UI / mono fonts | SIL OFL 1.1 | See `assets/fonts/*/OFL.txt` |
| **[go-text/typesetting](https://github.com/go-text/typesetting)** | Text shaping | BSD-3-Clause (+ MIT HarfBuzz subtree) | https://github.com/go-text/typesetting |
| **[goldmark](https://github.com/yuin/goldmark)** | Markdown | MIT | https://github.com/yuin/goldmark |
| **[chroma](https://github.com/alecthomas/chroma)** | Syntax highlighting | MIT | https://github.com/alecthomas/chroma |
| **[go-latex](https://codeberg.org/go-latex/latex)** | LaTeX helpers | BSD-3-Clause | https://codeberg.org/go-latex/latex |
| **[gospell](https://github.com/client9/gospell)** | Spell-check | MIT | https://github.com/client9/gospell |
| **[gg](https://github.com/fogleman/gg)** / [sbinet/gg](https://git.sr.ht/~sbinet/gg) | Imaging helpers | MIT | — |
| **[fsnotify](https://github.com/fsnotify/fsnotify)** | File watching | BSD-3-Clause | https://github.com/fsnotify/fsnotify |
| **[zenity](https://github.com/ncruces/zenity)** | Native dialogs | MIT | https://github.com/ncruces/zenity |
| **[modernc.org/sqlite](https://gitlab.com/cznic/sqlite)** (+ libc/memory/mathutil) | SQLite | BSD-3-Clause | https://gitlab.com/cznic/sqlite |
| **[golang.org/x/image](https://pkg.go.dev/golang.org/x/image)** / **x/net** / **x/sys** / **x/text** | Go extended libs | BSD-3-Clause | https://pkg.go.dev |
| **[golang/freetype](https://github.com/golang/freetype)** | Font raster (via x/image stack) | BSD-style | https://github.com/golang/freetype |
| **[go-webview2](https://github.com/wailsapp/go-webview2)** (vendored) | Windows WebView host | MIT / ISC (loader) | Vendored under `internal/go-webview2` |
| **Microsoft Edge WebView2** | Embedded web runtime (Windows) | Microsoft redistribution terms | https://developer.microsoft.com/microsoft-edge/webview2 |


---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

Report security issues per [SECURITY.md](SECURITY.md).

---

*Gru (GRU) — Copyright Shawn Londono and LedoCorp — MPL 2.0*
