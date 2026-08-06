# Gru Licensing

**Status:** Active — MPL only  
**Product:** **Gru (GRU)** — Go UI toolkit  
**Organization:** [LedoCorp](https://ledocorp.org)  
**Author / maintainer:** Shawn Londono  

Gru platform source that we publish is licensed under the **Mozilla Public License 2.0** (root `LICENSE`).

## Ownership

| | |
|--|--|
| **Copyright** | Copyright (c) Shawn Londono and LedoCorp |
| **Publisher** | LedoCorp |
| **Primary maintainer** | Shawn Londono |

LedoCorp publishes Gru; Shawn Londono is the founder and primary maintainer. The public project name is **Gru (GRU)**. The historical label **GoRy** is deprecated and must not appear in public branding.

## What MPL covers (when published)

| Component | Included |
|-----------|----------|
| **`ui/`** | Full platform — all widgets and components |
| **Inspector** | Devtools in `ui/` |
| **`cmd/gru`** | App build / packaging CLI |
| **Curated demos** | Widget showcase / batch demos |
| **Sample apps** | `cmd/hello`, `cmd/webviewhello` (not the demo catalog) |

You may use, modify, and distribute MPL-covered files per the `LICENSE`. Modifications to MPL-covered **platform** files that you distribute remain under MPL 2.0.

## What is not in the public repo yet

**Gru Notepad** ships as a separate MPL showcase repo: [`shawnlondono/gru-notepad`](https://github.com/shawnlondono/gru-notepad)
(depends on this toolkit; not vendored here).

These stay in the private monorepo for now (still Gru-related; **not** a second license):

- Prism  
- Studio / testing harness / full demo catalog / benchmarks  
- Foundry (internal R&D)

## Commercial / dual license

**Not active** for this release. Dual-license / paid Showcase may come later; it is not part of the public tree.

## Third-party software

See root [`NOTICE`](../../NOTICE) and the credits table in the repository [`README.md`](../../README.md). Maintainer checklist: [THIRD_PARTY.md](THIRD_PARTY.md).

---

*LedoCorp / Shawn Londono — Gru (GRU) — MPL 2.0*
