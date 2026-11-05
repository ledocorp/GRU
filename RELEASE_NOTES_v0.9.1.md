# v0.9.1 — export fix: x11util

Patch on **v0.9.0** (S1 ship default).

## Fix

- Export **`x11util`** (Linux X11 raise helper). `examples/appinstance` and `platform_linux` already imported it; the package was missing from the public module zip, which broke `go` resolution on Linux and `go mod tidy` for apps that pull `appinstance`.

## Module

```bash
go get github.com/ledocorp/gru@v0.9.1
```

No present/text behavior change vs v0.9.0.
