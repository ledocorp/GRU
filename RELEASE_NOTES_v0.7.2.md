## Summary

Toolkit cut focused on the **Gru Notepad** public showcase path: document the separate app repo, fix the Linux `x11util` package gap that blocked consumers using `examples/appinstance`, and bump the public pin to **0.7.2**.

## What's new since v0.7.1

- **Showcase app:** [Gru Notepad](https://github.com/shawnlondono/gru-notepad) (`require`s this module; not vendored here)
- Site: [ledocorp.org/gru/notepad](https://ledocorp.org/gru/notepad) · clips on [X](https://x.com/ledocorp/status/2085492726074626160)
- **`x11util`** package shipped (Linux raise/focus helper used by `examples/appinstance`) so module consumers tidy/build cleanly
- README / licensing / cheatsheet: Notepad is a public showcase repo, not “private for now”

## Install

```bash
go get github.com/ledocorp/gru@v0.7.2
```

```bash
go run ./cmd/hello
go run -tags grudemo .
go run -tags webview2 ./cmd/webviewhello
```

Showcase app (separate clone):

```bash
git clone https://github.com/shawnlondono/gru-notepad
cd gru-notepad
go run ./cmd/grunotepad
```

See the [README](https://github.com/ledocorp/GRU#readme) and [Getting Started](https://github.com/ledocorp/GRU/blob/v0.7.2/docs/GETTING_STARTED.md).

Primary target: **Windows**.
