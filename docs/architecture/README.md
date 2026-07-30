# Architecture (public)

How Gru’s core hangs together — **split chapters**, not one megafile.

Private monorepo maintainers keep a full internal `ARCHITECTURE.md`. What ships here is rewritten for public apps (no private product trees).

---

## Reading order

| # | Chapter | What you learn |
|---|---------|----------------|
| 1 | [overview.md](overview.md) | Retained mode, frame loop, dirty flags, Document |
| 2 | [window-chrome.md](window-chrome.md) | Undecorated window, TitleBar, grab + double-click, DWM |
| 3 | [layout.md](layout.md) | Chrome insets, flex, Viewport vs fill-height |
| 4 | [rendering.md](rendering.md) | SSAA, eager overlay RT, idle / wake |
| 5 | [webview.md](webview.md) | HWND vs GL — full guide: [../WEBVIEW.md](../WEBVIEW.md) |
| 6 | [input-focus.md](input-focus.md) | Pointer routing, overlays, WebView handoff |

---

## Design pillars

1. **Retained mode** — mutate widgets; the engine layouts and draws when dirty.  
2. **Document chrome** — title bar / insets are first-class; content lives below `ChromeTop`.  
3. **Eager quality** — supersampling and overlay targets initialize up front (no deferred soft-modal footguns).  
4. **WebView honesty** — separate HWND compositor; no passthrough tricks.

---

## Related

- [../GETTING_STARTED.md](../GETTING_STARTED.md) · [../CLI.md](../CLI.md) · [../BUILD.md](../BUILD.md)  
- [../COMPOSITION.md](../COMPOSITION.md) · [../widgets/README.md](../widgets/README.md) · [../api/README.md](../api/README.md)  
- [../DEMO_INDEX.md](../DEMO_INDEX.md) · [../WEBVIEW.md](../WEBVIEW.md) · [../README.md](../README.md)
