# API documentation

Gru’s interactive API lives in package **`github.com/ledocorp/gru/ui`**. This folder is a **narrative companion** to generated reference docs — not a replacement for reading source godoc.

---

## Generated reference

Use the Go toolchain against your module checkout or pkg.go.dev:

```bash
# Local — browse in terminal
go doc github.com/ledocorp/gru/ui
go doc github.com/ledocorp/gru/ui.Document
go doc github.com/ledocorp/gru/ui.NewDocument

# Local — HTML site (optional)
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite -http :8080
# open http://localhost:8080/github.com/ledocorp/gru/ui
```

Online: [pkg.go.dev/github.com/ledocorp/gru/ui](https://pkg.go.dev/github.com/ledocorp/gru/ui)

Godoc on `ui/node.go` and `ui/document.go` includes LLM-oriented prompt templates for the frame loop and custom widgets.

---

## Narrative pages (this folder)

| Page | Topics |
|------|--------|
| [document.md](document.md) | `NewDocument`, chrome insets, `DrainQueue`, `Resize`, focus, `NeedsRedraw` |
| [node-element.md](node-element.md) | `Node` interface, `Element`, `MarkDirty`, `Update` / `Layout` / `Draw` |
| [mount-chrome.md](mount-chrome.md) | `MountBorderlessDocument`, `SyncBorderlessClientSize`, `SyncBorderlessInputFrame`, `TitleBar`, overlay insets |
| [signals.md](signals.md) | `Signal`, `Effect`, `Binding`, Counter Demo walkthrough |
| [webview-api.md](webview-api.md) | `WebViewPanel`, build tags, main-loop sync |

---

## Where else to look

| Need | Doc |
|------|-----|
| Minimal working loop | [../GETTING_STARTED.md](../GETTING_STARTED.md) — copy `cmd/hello` |
| System contracts (dirty flags, SSAA, focus) | [../architecture/README.md](../architecture/README.md) |
| Widget types ↔ demos | [../widgets/README.md](../widgets/README.md) |
| Shell recipes (Panel, page grid) | [../COMPOSITION.md](../COMPOSITION.md) |
| WebView production rules | [../WEBVIEW.md](../WEBVIEW.md) |

**Rule of thumb:** start with [GETTING_STARTED](../GETTING_STARTED.md) and `cmd/hello`, use **architecture** when behavior surprises you, use **widgets** when picking a control, use **api/** when wiring a specific type.
