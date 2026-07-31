# Getting started with Gru

Build a small desktop window with `ui`, then explore the curated demos.

**Module:** `github.com/ledocorp/gru`  
**Shell recipe:** CP-SHELL-PAGE spirit (flex column + Card) — see the sample.

---

## 1. Run the sample apps

From the **repository root** (folder with `go.mod` and `assets/`):

```bash
go run ./cmd/hello
go run -tags webview2 ./cmd/webviewhello   # Windows + WebView2 runtime
```

| Path | Role |
|------|------|
| `cmd/hello` + `samples/hello` | Native chrome + Button |
| `cmd/webviewhello` + `samples/webviewhello` | FillClient WebView shell |

Scaffolds:

```bash
go run ./cmd/gru new myapp
go run ./cmd/gru new myapp --webview
```

See [CLI.md](CLI.md) for build and package commands. WebView details: [WEBVIEW.md](WEBVIEW.md).

`cmd/hello` is a **correctness / composition sample**, not a claim of minimal RAM or EXE size.

No build tags on Windows for hello. Linux: `go run -tags x11 ./cmd/hello`. Live WebView host is Windows `-tags webview2` for now.

---

## 2. Use Gru as a module

```bash
go get github.com/ledocorp/gru@latest
```

```go
import "github.com/ledocorp/gru/ui"
```

Confirm the GitHub org/repo slug before relying on `@latest` (LedoCorp public remote).

Working directory (or packaged layout) must resolve `assets/fonts/` (Remix + UI faces).

---

## 3. Minimal loop

Authoritative template: godoc on `ui.NewDocument` / `ui/document.go`.

1. `rl.InitWindow` (often borderless + `FlagWindowHidden` until first frame)  
2. `ui.InitDisplayAwareAtlases` + `ui.InitSupersampling`  
3. `doc := ui.NewDocument(w, h)` → build children under `doc.Root`  
4. Each frame: `DrainQueue` → `Update` → `Layout` if dirty → `BeginSuperFrame` / `Draw` / `BlitToScreen*`  

Copy `cmd/hello` rather than inventing a new loop.

---

## 4. Composition (public)

See [COMPOSITION.md](COMPOSITION.md) for the shell matrix and Panel/Card rules. Widget map: [widgets/README.md](widgets/README.md).

| Prefer | Avoid |
|--------|--------|
| `Panel` / `Card` + `ListTile` | Custom list/sidebar wrapper widgets |
| Flex column/row containers | Hardcoded absolute stacks for page body |
| Theme styles (`form-label`, `card`, …) | Ad-hoc colors on every node |

---

## 5. Curated demos

```bash
go run -tags grudemo .
```

- **Tab** — next scene  
- **F12** — Inspector  

Allowlisted titles: [DEMO_INDEX.md](DEMO_INDEX.md) (40 scenes).

Optional WebView demos: `go run -tags "grudemo,webview2" .` (Windows + WebView2 runtime).

---

## 6. CLI

```bash
go run ./cmd/gru build hello            # → dist/HelloGru.exe
go run ./cmd/gru build webviewhello -webview2
go run ./cmd/gru build demo             # → dist/GruDemo.exe (-tags grudemo)
go run ./cmd/gru build demo -webview2   # live WebView2
go run ./cmd/gru package hello windows  # zip + staged assets/fonts
```

Full reference: [CLI.md](CLI.md). Prefer explicit product names (`hello` / `webviewhello` / `demo`).

---

## Related

- [docs home](README.md) — reading order  
- [PUBLIC_README](../README.md) — product overview + credits (export root)  
- [BUILD.md](BUILD.md) — tags, smoke checklist  
- [architecture/README.md](architecture/README.md) — system chapters  
- [widgets/README.md](widgets/README.md) · [api/README.md](api/README.md)  
- Root `LICENSE` (MPL 2.0) + `NOTICE`
