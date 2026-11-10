# Gru toolkit v0.9.x — cumulative release notes

Maintainer rollup for the **0.9** train (tip sync → platform S1).  
Public GitHub cut should use a scrubbed subset (no private product names).  
Tagged points so far: **v0.9.0**, **v0.9.1**; platform merges on `main`; tagged **v0.9.2** (2026-11-10).

---

## Cross-platform text / bold (read this first)

**Linux demo scenes looking “different bold” than Windows is expected — not a Windows regression.**

| | Windows S1 | Linux / Android S1 |
|--|------------|--------------------|
| Sticky Latin | DirectWrite **ClearType** (`CLEARTYPE_NATURAL`) | Outline → fx 4× AA → greyscale sticky |
| Chrome “Bold” | **`segoeuib.ttf`** (Segoe UI Bold) | Poppins **Medium ≤18** / **SemiBold** (Bold 700 only as fallback) |
| Why | ClearType + Segoe Bold already reads lean | Poppins-Bold 700 under greyscale AA looked “deep/muddy”; soft-bold ladder matches Inter preview policy |
| Present | SSAA **1×**; DisplayScale follows monitor DPI | SSAA **1×**; Linux `DisplayScale=1` (1:1 client=FB) |

**Post-merge Windows smoke (2026-11-10)** confirmed bold is live, not Regular:

```text
Gru: ft-grey chrome soft-bold segoeuib.ttf (≤18 Medium / else SemiBold; Bold only as fallback)
Gru: ft-grey+strikes+dwrite-lcd hybrid @ SSAA 1× (ClearType+subX; Path A fallback)
Gru: ft-grey sticky strikes baked=8 bold=8 ladder=[14 16 18 20 24 28 32 40]
```

`bold=8` + `segoeuib` = chrome Bold strikes baked. Demo host chrome not “breaking” after merge aligns with this.

Deep dive: `docs/architecture/linux-s1-text.md` (public tree after platform merge).

---

## v0.9.0 — S1 ship default (2026-11-05)

**Tag:** `v0.9.0` · tip sync from living monorepo present/text path.

- Ship default **SSAA 1×** present (not 2× downscale). Quality from MSAA, sticky ClearType/LCD Latin, high-res glyph raster, coverage AA.
- Optional `GRU_SSAA=1.5|2` / `GRU_RENDER_QUALITY=…` for testing.
- **fx emit-C** kernels in public tree (`gen/fxlib`, `third_party/fx/emit-go-abi`); opt out `-tags gru_nofx` / `GRU_FX_*=0`.
- Export hygiene: private product blank-imports excluded; EqualFill grid row-slice fix.
- Docs rewritten for 1× present; public dogfood host **`grudemo`**.
- Author: Shawn Londono `<shawnlondono@gmail.com>` / @shawnlondono.

**Paired showcase:** [gru-notepad v0.9.0](https://github.com/shawnlondono/gru-notepad) pins toolkit **v0.9.1** (see below).

---

## v0.9.1 — export fix: x11util (2026-11-05)

**Tag:** `v0.9.1` · patch on 0.9.0.

- Export **`x11util`** (Linux X11 raise). Was imported by `examples/appinstance` / `platform_linux` but missing from the module zip → broke Linux `go` resolution and Notepad `go mod tidy`.
- No present/text behavior change vs 0.9.0.

---

## v0.9.2 — Platform S1 merges (2026-11-10)

**Commits (merge --no-ff, Shawn only, no Cursor trailer):**

| Step | Merge | Undo anchor |
|------|-------|-------------|
| Pre-merge tip | `4302cf1` (v0.9.1) | tag/branch `pre-platform-s1-merge` / `main-pre-platform-s1` |
| Linux S1 | `9195aba` ← `linux-experience-s1` | tag/branch `pre-android-s1-merge` / `main-pre-android-s1` |
| Android S1 | `b8a763f` ← `android-experience-s1` (contains Linux) | current `main` |

### Linux S1 (`linux-experience-s1`)

- 1:1 present: client size = screen; force off framebuffer scale densify; `DisplayScale=1` on Linux.
- Sticky text: outline → fx AA greyscale; baseline grid-lock; HarfBuzz GPOS kern (matching bake face).
- Chrome soft-bold: Poppins Medium/SemiBold assets + resolve path.
- Wayland/X11 helpers: interactive move/resize, GLFW platform hooks, window position stubs.
- Packaging: `.desktop`, icons, install script; docs `linux-s1-text.md`, `LINUX_S1_RETARGET_REVIEW.md`.
- Shared but gated: `titlebar` Wayland paths; `FontDensity` panel-title 1.05→1.0 (greyscale fattening); `batch_caption` → `form-label`.
- `dwrite_lcd.c` → `dwrite_lcd_windows.c` (rename only).

### Android S1 (`android-experience-s1`)

- NativeActivity path: density / `Dp`, IME + SAF file pick, insets, OpenBrowser, WebView sample packaging.
- GLES: skip desktop GLSL 330; no SSAA FBO; scissor skip until verified; sticky outline text + soft-bold (shared with Linux).
- Exclude Android from Linux Wayland/X11 CGO so APK links.
- `type_scale`: phone rem via `AndroidGLES()` only; desktop buckets unchanged.
- Packaging under `packaging/android/`, scripts under `scripts/android/`, docs `ANDROID_*.md`.

### Windows verification after merges

- `go build ./cmd/hello`, `go build -tags grudemo .`, `go test ./ui/` — green after Linux and again after Android.
- Hello smoke: S1 defaults, DWrite LCD hybrid, **segoeuib** soft-bold log, `bold=8`.

### Undo (if needed)

```powershell
# In the public GRU checkout:
git revert -m 1 b8a763f   # drop Android
git revert -m 1 9195aba   # drop Linux
git push origin main
```

Maintainer undo notes live in the private tip-sync staging doc (`PLATFORM_S1_MERGE_UNDO`).

---

## Still open / follow-ups (not blocking v0.9.2 tag)

- [x] Tag **v0.9.2** + GitHub Release (2026-11-10).
- Public **gru-notepad**: Linux host APIs need toolkit tip that exports them (now on `main`); re-pin / rebuild against post-merge tip; strip Cursor co-author on new commits only.
- Optional A/B: Windows panel-title `FontDensity` 1.05 vs 1.0 under ClearType (Linux needed 1.0).
- `gru-experiments` seed still Tip P3.

---

## Module

```bash
go get github.com/ledocorp/gru@v0.9.2
go run ./cmd/hello
go run -tags grudemo .
```

---

*LedoCorp / Shawn Londono — Gru toolkit 0.9.x*
