## Summary

**0.8.0** is the **GPLv3 + BDFL** public line for Gru (GRU). LedoCorp-published toolkit sources are now under the **GNU General Public License v3.0**. Official repositories are **maintainer-led (BDFL)** — forks under GPLv3 are welcome; pull requests are not merged on the official trees.

This tag is primarily a **license and governance release**. Toolkit behavior matches the prior public **v0.7.2** tree unless you already pulled intermediate commits on `main`.

## What's new

- **License:** **GNU GPLv3** — root `LICENSE`, `NOTICE`, and `docs/licensing/*`
- **Governance:** BDFL model in `CONTRIBUTING.md` and `docs/licensing/Governance.md`
- **VERSION:** root `VERSION` file set to **0.8.0**
- **README:** public overview, credits, and install pin updated for **0.8.0**
- **Showcase app:** [Gru Notepad v0.8.0](https://github.com/shawnlondono/gru-notepad/releases/tag/v0.8.0) (separate repo; also GPLv3)

Commercial dual-licensing (GPLv3 **or** a paid proprietary-use grant) may come later — **not** part of this release.

## Install

```bash
go get github.com/ledocorp/gru@v0.8.0
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

- **Site:** [ledocorp.org/gru](https://ledocorp.org/gru/) · [docs](https://ledocorp.org/gru/docs/) · [licensing](https://ledocorp.org/gru/docs/licensing/)
- **Third-party:** see root `NOTICE`

Primary target: **Windows**. Basic Linux samples exist; see docs for platform notes.

---

*LedoCorp / Shawn Londono — Gru (GRU) — GPLv3 · BDFL*
