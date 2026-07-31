# Gru documentation

**Module:** `github.com/ledocorp/gru`  
**Samples:** `cmd/hello`, `cmd/webviewhello`, and the curated demo catalog (`go run -tags grudemo .`).

This folder is the public docs home for Gru — the same paths ship in the GitHub repo under `docs/`. Start here, then follow the reading order below.

---

## Reading order

### 0 — Cheatsheet (fingertips)

| Doc | What you get |
|-----|----------------|
| [CHEATSHEET.md](CHEATSHEET.md) | Commands, intent index, capability matrices, Status gap tracker |
| [LAYOUT.md](LAYOUT.md) | Height modes, scroll owner, author checklist |

Use this when you already know the kind of UI you need. Guides below teach; the cheatsheet routes.

### 1 — Build and run

| Doc | What you get |
|-----|----------------|
| [GETTING_STARTED.md](GETTING_STARTED.md) | Install, run `hello` / `webviewhello`, minimal frame loop, module import |
| [CLI.md](CLI.md) | `gru new`, `gru build`, `gru package`, build tags |
| [BUILD.md](BUILD.md) | Platform tags, assets/fonts, release layout |

Copy **`cmd/hello`** for a correct native loop before inventing your own.

### 2 — Compose scenes

| Doc | What you get |
|-----|----------------|
| [COMPOSITION.md](COMPOSITION.md) | Shell recipes, Panel/Card surfaces, borrow matrix for public demos |
| [WEBVIEW.md](WEBVIEW.md) | WebView2 build tags, HWND compositing, FillClient shells |

### 3 — Deep dives

| Section | What you get |
|---------|----------------|
| [architecture/](architecture/README.md) | System contracts — retained mode, chrome, layout, rendering, input |
| [widgets/](widgets/README.md) | Control catalog mapped to grudemo scenes |
| [api/](api/README.md) | Narrative API guide + links to `go doc` / pkgsite |

Architecture chapters explain *why*; API pages explain *how to call*; widgets tie types to runnable demos.

### 4 — Explore and legal

| Doc | What you get |
|-----|----------------|
| [DEMO_INDEX.md](DEMO_INDEX.md) | All 40 grudemo scenes — titles, widgets, Tab navigation |
| [licensing/](licensing/README.md) | MPL 2.0, contributing, security, third-party credits |

---

## Quick commands

```bash
go run ./cmd/hello
go run -tags webview2 ./cmd/webviewhello    # Windows + WebView2 runtime
go run -tags grudemo .                       # scene catalog
go run ./cmd/gru new myapp
```

Package reference: `go doc github.com/ledocorp/gru/ui` or [pkgsite](https://pkg.go.dev/github.com/ledocorp/gru/ui).

---

## Related (repo root)

- [`README.md`](../README.md) — project overview (export root)
- [`NOTICE`](../NOTICE) — copyright and attribution
