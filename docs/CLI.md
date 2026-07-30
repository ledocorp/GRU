# Gru CLI (`cmd/gru`)

Public packaging and scaffold tool. Prefer this over ad-hoc scripts.

```bash
go run ./cmd/gru <command> …
# after install:
gru <command> …
```

---

## new — scaffold an app

```bash
go run ./cmd/gru new myapp
go run ./cmd/gru new myapp --webview
```

| Flag | Template |
|------|----------|
| (default) | Native sample based on `cmd/hello` + `samples/hello` |
| `--webview` | FillClient WebView shell based on `cmd/webviewhello` + `samples/webviewhello` |

Creates a small app directory you can `go run` from the repo (or move into your own module). Requires WebView2 Runtime for `--webview` on Windows.

---

## build — compile a product

```bash
go run ./cmd/gru build hello [-o path]
go run ./cmd/gru build webviewhello -webview2 [-o path]
go run ./cmd/gru build demo [-webview2] [-o path]
```

| Product | What it builds |
|---------|----------------|
| `hello` | Native sample (`./cmd/hello`) |
| `webviewhello` | WebView sample (`./cmd/webviewhello`); use `-webview2` for live host |
| `demo` | Curated demo launcher (`-tags grudemo`; optional `-webview2`) |

Optional platform args: `windows` / `linux` (see `gru build -h`). Default output lands under `dist/`.

---

## package — zip for distribution

```bash
go run ./cmd/gru package hello windows
go run ./cmd/gru package webviewhello windows
go run ./cmd/gru package demo windows
```

Stages fonts/assets needed at runtime into a release zip. Prefer packaging after a successful `build` of the same product.

---

## Related

- [GETTING_STARTED.md](GETTING_STARTED.md) — first window  
- [BUILD.md](BUILD.md) — tags and smoke  
- [WEBVIEW.md](WEBVIEW.md) — WebView host contract (FillClient)
