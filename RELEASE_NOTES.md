# Gru toolkit v0.9.x — cumulative notes (draft toward v0.9.2)

Rollup of the **0.9** public tip: S1 ship default, export fix, then Linux + Android S1 experience merges.

**Last tag:** `v0.9.1` · **main tip:** platform S1 merges (2026-11-10). Tag **v0.9.2** when cut.

---

## Bold looks different on Linux vs Windows — expected

Not a Windows “bold broken” bug. Post-merge Windows smoke shows chrome Bold on **Segoe UI Bold** + DirectWrite ClearType (`segoeuib.ttf`, `bold=8` strikes).

| | Windows | Linux / Android |
|--|---------|-----------------|
| Raster | DirectWrite ClearType sticky | Outline → fx AA greyscale sticky |
| Chrome Bold | `segoeuib` (Segoe Bold) | Poppins Medium (≤18px) / SemiBold |
| Present | SSAA 1× + monitor DPI for atlases | SSAA 1×; Linux DisplayScale=1 (1:1) |

Linux avoids Poppins-Bold 700 under greyscale AA (read as muddy). Details: [`docs/architecture/linux-s1-text.md`](docs/architecture/linux-s1-text.md).

---

## v0.9.0 — S1 ship default

- Default **SSAA 1×** present (not 2×). Sticky ClearType/LCD Latin, glyph floor, coverage AA.
- Optional `GRU_SSAA` / `GRU_RENDER_QUALITY` for testing.
- fx emit-C kernels ship for default CGO builds (`-tags gru_nofx` / `GRU_FX_*=0` to opt out).
- Public dogfood: **`grudemo`**. EqualFill row-slice fix; export hygiene.

```bash
go get github.com/ledocorp/gru@v0.9.0
```

Showcase: [gru-notepad v0.9.0](https://github.com/shawnlondono/gru-notepad) (pins a later toolkit patch).

---

## v0.9.1 — x11util export

- Add missing **`x11util`** package so Linux `appinstance` raise resolves in the module zip.

```bash
go get github.com/ledocorp/gru@v0.9.1
```

---

## After v0.9.1 on main — Linux + Android S1 (candidate v0.9.2)

Merged with `--no-ff` (easy revert). Undo anchors: `pre-platform-s1-merge`, `pre-android-s1-merge`.

**Linux:** 1:1 present; outline sticky text; soft-bold Poppins; Wayland/X11 chrome helpers; desktop packaging/docs.

**Android:** NativeActivity density/IME/SAF/insets; GLES sticky text; skip desktop-only shaders/SSAA FBOs; packaging under `packaging/android/`.

**Windows checks after merge:** `go build` hello + grudemo, `go test ./ui/`, hello S1 smoke (DWrite + segoeuib) — green.

---

## Module (until v0.9.2 is tagged)

```bash
go get github.com/ledocorp/gru@v0.9.1
# or @main for platform S1 tip
go run ./cmd/hello
go run -tags grudemo .
```

---

*LedoCorp / Shawn Londono — GPLv3*
