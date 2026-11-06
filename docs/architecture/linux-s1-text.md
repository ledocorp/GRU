# Linux S1 text — outline AA (Windows DWrite analogue)

**Status:** polished Linux sticky greyscale (Studio + Notepad share `ui`). Ready to land as the **linux-experience** PR once reviewed.  
**North star on Windows:** DirectWrite `CLEARTYPE_NATURAL` sticky strikes + LCD shader + `FilterPoint`.  
**Linux analogue:** TrueType **outline → fx 4× coverage AA** (gg fallback) → box-downsample → sticky ladder; greyscale blit; `FilterPoint`; **no** SkMaskGamma / `BoostRGBAAlpha`.

Related present policy: [`LINUX_S1_RETARGET_REVIEW.md`](../../LINUX_S1_RETARGET_REVIEW.md) (1:1 client=FB, no Cosmic DPI densify).

---

## What “good” looks like

- Sticky exact-px strikes (`ladder=[14 16 18 20 22 24 28 32]`, aux may include 40).
- Solid joins on hard glyphs (`R` / `r`) — no light holes in stems.
- Even baseline across digits (`100` — `1` and `0` share one line) and words (`Unsaved`).
- Pair spacing coherent (`Close`, `thin`, `That`) — no `Cl ose` / `th in` / `T hat` gaps from draw-time snap fights.
- Chrome “Bold” reads lean (Poppins **Medium ≤18** / **SemiBold**, not crushed 700).
- Startup log lines:

```text
Gru: ft-grey sticky outline AA via fx @4× bake→1× (Windows DWrite analogue on Linux)
Gru: ft-grey chrome soft-bold Poppins-SemiBold.ttf (≤18 Medium / else SemiBold; Bold only as fallback)
Gru: ft-grey sticky strikes baked=8 bold=8 ladder=[14 16 18 20 22 24 28 32]
Gru: ft-grey sticky kerning via HarfBuzz (liga off; GPOS kern)
```

Build / dogfood:

```bash
# Studio (demo)
cd /path/to/GRU
GRU_SSAA=1 go build -tags grudemo -o /tmp/gru-studio-s1 .
GRU_SSAA=1 /tmp/gru-studio-s1

# Notepad (replace → ../GRU)
cd /path/to/gru-notepad
GRU_SSAA=1 go build -o /tmp/grunotepad-s1 ./cmd/grunotepad
GRU_SSAA=1 /tmp/grunotepad-s1
```

Both hosts import the same toolkit; fixing `ui/text_ft_grey*.go` fixes both. Rebuild **both** after every text change.

---

## Pipeline comparison

| Stage | Windows (default S1) | Linux (default S1) |
|-------|----------------------|--------------------|
| Raster | DWrite `CLEARTYPE_NATURAL` RGB (± subX phases) | Outline → **fx** at **glyph SSAA** (4× bake) → box-downsample to strike px (gg if `GRU_FX_GLYPH=0`) |
| Atlas | RGB ClearType, A=255 | Greyscale white RGB + coverage alpha (**1×** cells) |
| Post-bake | `enhanceS1DWriteLCDAtlas` = **no-op** | **None** (no gamma / no fx boost) |
| Filter | `FilterPoint` | `FilterPoint` (outline); bilinear only if FT bitmap fallback |
| Draw | LCD shader, linear-light mix, backdrop stack | `DrawTexturePro`; snap **baseline + start pen once**; `dst = pen + bearing` (**no** per-glyph `snap(pen+bearing)`); integer kern advances |
| Bearings | DWrite metrics | **Grid-locked** ox/oy before raster so bearings are exact integers — shared baseline (`TestFTGreyOutlineBaselineLock`) |
| Hinting | DWrite natural | Outlines unhinted; **geometric SSAA + downsample** stands in for grid-fit |
| Scale / DPI | Monitor DPI densify (T1.9) where applicable | `DisplayScale=1`; present 1:1 (see Linux present review) |
| Chrome bold | `segoeuib` (Segoe Bold) | Poppins **Medium ≤18** / **SemiBold** / Bold fallback |
| Kern face | shaped stack | **`pickStickyKernFace`** — same TTF as outline bake (not shapedUI.bold 700 while drawing Medium) |

## Do we need DirectWrite on Linux?

**Role yes, API no.** Windows sharpness comes from DWrite (hinted + ClearType LCD). Linux cannot call DWrite; the analogues are:

| Need | Linux approach | Status |
|------|----------------|--------|
| Solid joins (no `R` holes) | Outline → fx SSAA → 1× atlas | Shipped |
| Shared baseline (digits / Unsaved) | Grid-lock ox/oy; exact integer bearings | Shipped |
| Pair spacing without snap fights | Integer pen + HarfBuzz GPOS; no `snap(pen+bearing)` | Shipped |
| Real `**bold**` on Inter VF | `wght`/`opsz` variations | Shipped |
| Lean chrome Bold | Medium/SemiBold ladder | Shipped |
| ClearType-class edge crispness | FreeType LCD + existing LCD shader | **Not default** — lite LCD was soft/fringed; greyscale SSAA is intentional |

So: **not** “port DWrite”; **do** keep outline grey + kerning + baseline lock. Optional future A/B: real FreeType `FT_RENDER_MODE_LCD` (not neighbor-RGB lite) behind `GRU_TEXT_LCD`.

**Do not enable on Linux without a new A/B:**

- `GRU_TEXT_LCD=1` lite neighbor-RGB LCD — soft/fringed at 1:1 without DWrite.
- `tryFxBoostStrikeAlpha` / `BoostRGBAAlpha` on sticky grey — speckles / ghost stems.
- Aggressive morphological “bridge repair” across the whole atlas — irregular weight, waviness (reverted).
- Per-glyph `dstX = snap(penX + bearingX)` — pair gaps (`Close` → `Cl ose`).
- Fractional `FontSpacing` added then snap-killed — tracking discarded or fighting bearings.
- Baking Medium/SemiBold while HarfBuzz shapes Poppins-Bold — systematic kern mismatch.

---

## Code map

| Path | Role |
|------|------|
| `ui/text_ft_grey_outline.go` | `packFTGreyAtlasOutline` / `rasterFTGreyOutlineGlyph` — outline → fx/gg → **grid-lock** → box downsample → atlas |
| `ui/text_ft_grey.go` | Chrome sticky bake/draw; soft-bold resolve; `drawTextFTGreyWeight` |
| `ui/text_ft_grey_aux.go` | Fira / Inter sticky; same outline pack (`pad=4`) |
| `ui/text_ft_grey_kern.go` | HarfBuzz GPOS kern (liga off); `pickStickyKernFace`; FontSpacing folded into integer advances |
| `ui/text_glyph_fx.go` | `tryRasterShapedGlyphFX` / `packGlyphOutline` |
| `ui/text_glyph_cache.go` | `shapedGlyphRasterLayoutPad` |
| `fxlib/glyph_raster.fx` | `raster_outline`; `boost_rgba_alpha` (**must not** call on sticky grey) |
| `ui/text_dwrite_lcd_windows.go` | Windows ClearType bake (unchanged by Linux polish) |
| `ui/font_stack.go` | `interFaceForStickySize` Medium/SemiBold optical cuts |
| `ui/style.go` | `panel-title` density 1.0; `form-value` MinFontSize/NoCaptionBold |
| `examples/batch_caption.go` | Hints use **`form-label`**, not `form-value` |

**Tests:** `TestPackFTGreyAtlasOutlineHasSolidR`, `TestFTGreyOutlineBaselineLock` (`ui/text_ft_grey_test.go`).

---

## Full decision log (polish campaign)

This is the history of the Linux font / present polish (scale → holes → outline → bold → kern → snap → baseline). **Do not re-open settled branches** without screenshots and a new A/B.

### Phase A — Present / scale (settled — leave alone)

Cosmic / Pop!_OS fractional scale made early Linux builds look soft when Gru treated framebuffer ≠ client, stacked DPI into rem, densified Camera zoom, or layout-at-FB size.

Failed “add scale” ideas (rejected by product direction): SSAA floor, presentBacking Camera zoom, glyph densify-for-stretch.

**Final Linux present policy:** force `SCALE_FRAMEBUFFER` off; `ClientLayoutSize` = screen size only; `DisplayScale=1`; SSAA 1×; MSAA off on S1 Linux hosts. Cursor/Brave densify for DPI; Gru deliberately stays **1:1**. Softness that remains after 100% compositor scale is outside the app. See `LINUX_S1_RETARGET_REVIEW.md`.

User direction: *“I don't care about scale… Once the fonts are fixed…”*

### Phase B — Greyscale sticky default

Without DWrite, auto LCD (`PreferFTLCDChrome`) is off. Lite FreeType LCD looked soft/fringed. Path: sticky greyscale strikes + Path A fallback.

### Phase C — Softness and holes — failed experiments

| Attempt | Symptom |
|---------|---------|
| SkMaskGamma / midtone densify on grey | White holes / uneven joins (`r`) |
| `BoostRGBAAlpha` (fx contrast) | Bowl speckles, ghost stems |
| `FilterPoint` on FT bitmap grey | Magnified 1px join dropouts as white gaps |
| `FilterBilinear` + raw FT masks | Better overall grey, but `R`/`r` still punched |
| Atlas-wide **bridge repair** | Briefly targeted `R`; **made Notepad much worse** — **fully reverted** |

Root cause of `R` holes: **golang/freetype** `Face.Glyph` bitmaps leave near-zero coverage in thin joins. Not a present/scale bug.

### Phase D — Working raster — outline AA via fx (+ sticky bake SSAA)

Windows never hits the FT bitmap path for sticky chrome: DWrite supplies coverage. Linux fills the same *role* with the fx glyph raster:

1. Parse sticky TTF with go-text (`font.ParseTTF`).
2. For each rune: outline + extents → layout at **`sizePx × GlyphSupersample`** (default 4×).
3. **Grid-lock** ox/oy to strike-pixel boundaries; pad w/h to multiples of SSAA.
4. Raster with `fxglyph.RasterOutline` (gg fallback).
5. **Box-downsample** to 1× strike px.
6. Bearings = exact `-ox/ssaa`, `-oy/ssaa` (integers; shared baseline).
7. Advances from `HorizontalAdvance` (integer-rounded).
8. Upload **`FilterPoint`**; no post-bake gamma/boost.

Bake SSAA is atlas CPU only — **not** present/DPI scale. Cosmic at 100% on 4K confirmed compositor upscale was not the font bug.

### Phase E — Studio “soft” vs Notepad — Bold was a no-op

Same outline path in both apps. Studio looked softer because chrome **`Style.Bold` was ignored**: `drawTextFTGreyKind(chrome, bold)` always drew Regular Poppins while Notepad `**bold**` used real Inter/Fira bold strikes.

**False memory check:** users remembered leaner Studio chrome — that was **Regular drawn as “bold”** (bug), not a better Bold face.

### Phase F — Deep bold regression → soft-bold ladder

Wiring Bold for real with **Poppins-Bold (700)** + greyscale outline AA looked “deep bold” / muddy vs:

- Prior (buggy) Regular no-op on Linux  
- Segoe Bold ClearType on Windows (reads leaner)

**Fix:** size-aware chrome bold like Inter preview:

| Sticky px | Poppins face |
|-----------|----------------|
| ≤18 (e.g. Scenes ~14) | **Medium** |
| else | **SemiBold** |
| missing files | Bold fallback |

Assets: `assets/fonts/poppins/Poppins-Medium.ttf`, `Poppins-SemiBold.ttf`.  
Also: `panel-title` `FontDensity` **1.05 → 1.0** (ClearType optical nudge fattened greyscale).  
`batchCaption` **`form-value` → `form-label`** (hints were near-black 18pt readout style vs muted caption).  
`form-value` gained `MinFontSize` / `NoCaptionBold` for strike parity.

### Phase G — Inter VF preview bold

Linux trees often ship only `Inter-VariableFont_opsz,wght.ttf`. Without axes, regular and bold baked the same default instance → `**capability showcase**` “almost bold.”

**Fix:** `interVariableBoldVariations` — `wght=500` (≤18) / `600` + matching `opsz`.

### Phase H — Kerning and FontSpacing

Sticky grey draws a **per-rune atlas** (no liga). HarfBuzz supplies **GPOS kern** with liga/clig/calt off (`stickyKernFeatures`).

**Regression:** rounding `FontSpacing` (0.3) to **0** crushed Studio chrome after kern landed. Restored tracking by folding spacing into advances correctly (see Phase I).

**Face mismatch bug:** outline bake used Medium/SemiBold; HarfBuzz used `shapedUI.bold` → **Poppins-Bold**. GPOS pairs disagreed with drawn glyphs → Scenes/title gaps.

**Fix:** `pickStickyKernFace` loads the **same path** as `resolveFTGreyChromeBoldPathForSize` / `interFaceForStickySize`.

### Phase I — Conflicting transforms (the “doubling” bug)

Symptoms: `Close` → `Cl ose`, `thin` → `th in`, `That` → `T hat`; headers worse than small captions; user correctly suspected **too many conflicting things on the same text**.

What stacked:

1. HarfBuzz advances (rounded)  
2. Fractional `FontSpacing = 0.3`  
3. **`dstX = snapPresentPx(penX + bearingX)`** independently of the pen  

Measure used the pen; pixels used a different snap → ±1px pair gaps that looked like broken kerning.

**Fix (one positioning model):**

- Snap **baseline** and **start pen** once.  
- Kern advances: `Round(hb_adv + FontSpacing)` for non-last glyphs (integer pen stays integer).  
- Place at `dstX = penX + bearingX` — **never** re-snap `pen+bearing`.  
- Do **not** snap pen every glyph after adding fractional spacing (that discarded tracking).

### Phase J — Floating digits / wavy baseline

Symptoms: in `100`, zeros float above `1`; `Unsaved` letters sit on uneven lines; tall `l` exaggerated.

**Cause:** each glyph independently `Round(-oy/ssaa)` after a fractional baseline row in the hi-res bake (+ possible box-downsample shift when w/h not multiple of SSAA).

**Fix:** grid-lock ox/oy to strike-pixel boundaries before raster; pad dimensions to SSAA multiples; bearings exact integers. Guard: `TestFTGreyOutlineBaselineLock`.

**Tall `l`:** mostly **typeface design** (Inter/Poppins ascenders). Bearing lock removes the exaggeration; remaining height vs `e`/`c` is expected.

### Phase K — Cosmic Edit vs Gru (why Cosmic still “wins” on edges)

Side-by-side: Cosmic Edit (`cosmic-text` + FontConfig / FreeType, typically Noto/interface fonts) vs Gru mono (**Fira Code**) or chrome (**Poppins**) greyscale outline.

Cosmic looks leaner because it is a **different stack** (hinted FreeType, often LCD, system metrics) — not because Gru “forgot a font.” Gru’s closest match is the **markdown preview** Inter lane, not the mono source pane.

---

## PR note — Windows bold utilization (open follow-up)

Linux polish exposed that **chrome Bold must actually resolve to a bold face and match the kern face**. Windows was not the primary dogfood surface for this campaign; **verify before calling Windows “done”:**

| Check | Why |
|-------|-----|
| Does S1 chrome `Style.Bold` hit `segoeuib` / DWrite bold strikes, or still Regular? | Linux historically ignored chrome Bold entirely |
| Does HarfBuzz / shaped measure use the **same** bold face as DWrite bake? | Linux Medium-bake + Bold-kern mismatch |
| Segoe Bold at 14–16px vs Poppins Medium policy | Windows may still look leaner (ClearType); that can be correct — confirm it is not Regular |
| `FontDensity` 1.05 on titles | Dropped on Linux greyscale; Windows ClearType may still want a nudge — A/B before copying 1.0 blindly |
| `batchCaption` / `form-value` misuse | Theme bug is cross-platform; hints should stay `form-label` |

Capture Windows screenshots of Scenes, panel titles, Counter Demo, and Notepad `**bold**` in the linux-experience PR description. Treat “Windows bold path parity” as an explicit checklist item, not assumed green.

---

## If text regresses again

1. Log still says `sticky outline AA via fx @4× bake→1×` (or `via gg`). Missing / no `@N×` → outline pack failed → FT fallback or old binary.  
2. `bold=` count &gt; 0 and soft-bold log shows Medium/SemiBold — else Regular or deep 700.  
3. Kern log present; `pickStickyKernFace` still matches bake path.  
4. Do **not** reintroduce: bridge repair, sticky grey `BoostRGBAAlpha`, `snap(pen+bearing)`, lite LCD default, DPI densify for blur.  
5. Digits float again → check grid-lock in `rasterFTGreyOutlineGlyph` + `TestFTGreyOutlineBaselineLock`.  
6. Pair gaps again → check draw uses integer pen only; FontSpacing folded into Round, not added then snapped away.  
7. Rebuild **Studio and Notepad** after `ui/` text changes.

---

## LLM / maintainer prompt

```text
Read docs/architecture/linux-s1-text.md and LINUX_S1_RETARGET_REVIEW.md end-to-end.

Linux sticky Latin = outline→fx AA greyscale with grid-locked bearings
  (text_ft_grey_outline.go), HarfBuzz kern with matching bake face
  (text_ft_grey_kern.go), integer pen draw without snap(pen+bearing),
  chrome soft-bold Medium/SemiBold — NOT FT bitmaps, lite LCD,
  BoostRGBAAlpha, bridge-repair, DPI densify, or Poppins-Bold 700 for UI.

Windows sticky Latin = DWrite ClearType. Keep paths split.
Before merging linux-experience: verify Windows chrome Bold actually uses
segoeuib / DWrite bold strikes and matching shaped face (see PR note).
```

---

## What we deliberately did not change

- Windows DWrite ClearType path and LCD shader (aside from verifying Bold — open).  
- Linux 1:1 present / no compositor densify.  
- Sticky ladder sizes and RichText measure/draw contracts.  
- Opt-out: `GRU_FX_GLYPH=0` → gg outline; `GRU_TEXT_LCD=1` for A/B only.
