package ui

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"strings"
	"sync"

	"github.com/fogleman/gg"
	gotext "github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Linux S1 greyscale sticky — Windows uses DirectWrite ClearType; here we bake
// the same sticky ladder from TrueType outlines via fx coverage AA (gg
// fallback). Avoids golang/freetype bitmap dropouts in thin joins (R/r) without
// SkMaskGamma / BoostRGBAAlpha speckles.
//
// Sticky outlines bake at GlyphSupersampleFloor (typically 4×) then box-filter
// down to the strike px — same idea as shaped glyphs. Baking at 1× left stems
// between pixel centers (cut t/m, white joins, jagged curves) under FilterPoint.

var ftGreyOutlineOnce sync.Once

// packFTGreyAtlasOutline packs a greyscale sticky atlas from outline rasters.
// atlasPad is the gap between glyphs in the atlas (2 chrome / 4 aux).
func packFTGreyAtlasOutline(ttfData []byte, pixelSize, atlasPad int) (*image.RGBA, map[rune]ftGreyGlyph, int, bool) {
	return packFTGreyAtlasOutlineVars(ttfData, pixelSize, atlasPad, nil)
}

// packFTGreyAtlasOutlineVars is [packFTGreyAtlasOutline] with optional VF axes
// (e.g. Inter opsz,wght — bold preview when only the variable file is shipped).
func packFTGreyAtlasOutlineVars(ttfData []byte, pixelSize, atlasPad int, vars []gotext.Variation) (*image.RGBA, map[rune]ftGreyGlyph, int, bool) {
	if len(ttfData) == 0 || pixelSize < 11 {
		return nil, nil, 0, false
	}
	face, err := gotext.ParseTTF(bytes.NewReader(ttfData))
	if err != nil || face == nil {
		return nil, nil, 0, false
	}
	if len(vars) > 0 {
		face.SetVariations(vars)
	}
	if atlasPad < 2 {
		atlasPad = 2
	}
	atlasW, atlasH := 1024, 1024
	if atlasPad < 4 && pixelSize <= 28 {
		atlasW, atlasH = 512, 512
	} else if atlasPad < 4 && pixelSize >= 72 {
		atlasW, atlasH = 2048, 2048
	}
	atlas, glyphs, ink, ok := packFTGreyOutlineInto(face, pixelSize, atlasPad, atlasW, atlasH)
	if ok {
		return atlas, glyphs, ink, true
	}
	if atlasPad >= 4 || pixelSize <= 28 {
		return packFTGreyOutlineInto(face, pixelSize, atlasPad, 2048, 2048)
	}
	return nil, nil, ink, false
}

// interVariableBoldVariations matches interFaceForStickySize weight policy when
// only Inter-VariableFont_opsz,wght.ttf is available (no static Medium/SemiBold).
func interVariableBoldVariations(pixelSize int) []gotext.Variation {
	wght := float32(600) // SemiBold
	opsz := float32(24)
	switch {
	case pixelSize <= 18:
		wght = 500 // Medium — opens counters at small sticky sizes
		opsz = 18
	case pixelSize <= 20:
		opsz = 18
	case pixelSize <= 26:
		opsz = 24
	default:
		opsz = 28
	}
	return []gotext.Variation{
		{Tag: ot.MustNewTag("wght"), Value: wght},
		{Tag: ot.MustNewTag("opsz"), Value: opsz},
	}
}

func ttfPathLooksVariable(path string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "variablefont") || strings.Contains(p, "-vf.") || strings.Contains(p, "variable")
}

// outlineVarsForAuxPath returns VF axes for bold Inter when static cuts are absent.
func outlineVarsForAuxPath(fontPath string, bold bool, pixelSize int) []gotext.Variation {
	if !bold || !ttfPathLooksVariable(fontPath) {
		return nil
	}
	// Inter opsz,wght only — other VFs keep default instance.
	if !strings.Contains(strings.ToLower(fontPath), "inter") {
		return nil
	}
	return interVariableBoldVariations(pixelSize)
}

var interVFBoldLogOnce sync.Once

func logInterVFBoldOnce(fontPath string, bold bool, pixelSize int, vars []gotext.Variation) {
	if !bold || len(vars) == 0 {
		return
	}
	interVFBoldLogOnce.Do(func() {
		wght, opsz := float32(0), float32(0)
		for _, v := range vars {
			switch v.Tag {
			case ot.MustNewTag("wght"):
				wght = v.Value
			case ot.MustNewTag("opsz"):
				opsz = v.Value
			}
		}
		fmt.Printf("Gru: Inter VF bold strikes — wght=%.0f opsz=%.0f (no static SemiBold; path=%s @%dpx)\n",
			wght, opsz, fontPath, pixelSize)
	})
}

func packFTGreyOutlineInto(face *gotext.Face, pixelSize, atlasPad, atlasW, atlasH int) (*image.RGBA, map[rune]ftGreyGlyph, int, bool) {
	glyphs := make(map[rune]ftGreyGlyph, 256)
	atlas := image.NewRGBA(image.Rect(0, 0, atlasW, atlasH))
	x, y, rowH := atlasPad, atlasPad, 0
	ink := 0
	truncated := false
	upem := float32(face.Upem())
	if upem < 1 {
		upem = 1000
	}
	pxScale := float32(pixelSize) / upem

	for _, r := range sdfCodepoints() {
		gid, ok := face.NominalGlyph(r)
		if !ok {
			continue
		}
		// Integer advances keep the pen on-pixel so FilterPoint greyscale stems
		// match (Hello "ll", button "tt"). Fractional advances were the uneven-letter bug.
		adv := float32(math.Round(float64(face.HorizontalAdvance(gid) * pxScale)))
		outline, hasOutline := shapedGlyphOutline(face, gid)
		ext, hasExt := face.GlyphExtents(gid)
		if !hasOutline || !hasExt || len(outline.Segments) == 0 {
			glyphs[r] = ftGreyGlyph{advance: adv}
			continue
		}
		img, bearingX, bearingY, okR := rasterFTGreyOutlineGlyph(face, outline, ext, float32(pixelSize))
		if !okR || img == nil {
			glyphs[r] = ftGreyGlyph{advance: adv}
			continue
		}
		b := img.Bounds()
		gw, gh := b.Dx(), b.Dy()
		if gw <= 0 || gh <= 0 {
			glyphs[r] = ftGreyGlyph{advance: adv}
			continue
		}
		if x+gw+atlasPad > atlasW {
			x = atlasPad
			y += rowH + atlasPad
			rowH = 0
		}
		if y+gh+atlasPad > atlasH {
			truncated = true
			break
		}
		dst := image.Rect(x, y, x+gw, y+gh)
		n := blitFTGreyImage(atlas, dst, img)
		if n == 0 {
			glyphs[r] = ftGreyGlyph{advance: adv}
			continue
		}
		glyphs[r] = ftGreyGlyph{
			src:      rl.NewRectangle(float32(x), float32(y), float32(gw), float32(gh)),
			bearingX: bearingX,
			bearingY: bearingY,
			advance:  adv,
			hasInk:   true,
		}
		ink += n
		x += gw + atlasPad
		if gh > rowH {
			rowH = gh
		}
	}
	if truncated || ink < 32 || len(glyphs) < 32 {
		return nil, nil, ink, false
	}
	if atlasPad >= 4 && !auxAtlasHasASCII(glyphs) {
		return nil, nil, ink, false
	}
	return atlas, glyphs, ink, true
}

// stickyOutlineBakeSSAA is the CPU supersample for sticky outline strikes.
// Matches shaped-glyph floor; atlas still stores 1× after box downsample.
func stickyOutlineBakeSSAA() int {
	s := int(math.Round(float64(EffectiveGlyphSupersample())))
	if s < 2 {
		s = 2
	}
	if s > 4 {
		s = 4
	}
	return s
}

func rasterFTGreyOutlineGlyph(face *gotext.Face, outline gotext.GlyphOutline, ext gotext.GlyphExtents, sizePx float32) (image.Image, float32, float32, bool) {
	ssaa := stickyOutlineBakeSSAA()
	rasterPx := sizePx * float32(ssaa)
	// Keep ~3px of AA fringe after downsample (pad is in high-res pixels).
	pad := 3.0 * float64(ssaa)
	scale, w, h, ox, oy := shapedGlyphRasterLayoutPad(outline, ext, rasterPx, face.Upem(), pad)
	if w < 1 || h < 1 {
		return nil, 0, 0, false
	}
	// Grid lock: the pen origin (ox) and baseline (oy) must land on a strike-pixel
	// boundary of the hi-res image, and w/h must be whole strike pixels. Then the
	// bearings below are exact integers and every glyph shares one baseline row
	// ("1" and "0" in "100"), with no per-glyph Round() drift and no fractional
	// shift from the box downsample. Outline sub-pixel position stays in coverage.
	s := float64(ssaa)
	lockOx := math.Ceil(ox/s) * s
	lockOy := math.Ceil(oy/s) * s
	w += int(math.Ceil(lockOx - ox))
	h += int(math.Ceil(lockOy - oy))
	w = (w + ssaa - 1) / ssaa * ssaa
	h = (h + ssaa - 1) / ssaa * ssaa
	ox, oy = lockOx, lockOy
	img, ok := tryRasterShapedGlyphFX(outline, scale, w, h, ox, oy)
	if !ok {
		gc := gg.NewContext(w, h)
		gc.SetRGBA(0, 0, 0, 0)
		gc.Clear()
		gc.SetRGBA(1, 1, 1, 1)
		drawGlyphOutlineGG(gc, outline, scale, ox, oy)
		img = gc.Image()
	}
	if img == nil || !shapedGlyphCacheHasInk(img) {
		return nil, 0, 0, false
	}
	down := boxDownsampleGlyphCoverage(img, ssaa)
	if down == nil || !shapedGlyphCacheHasInk(down) {
		return nil, 0, 0, false
	}
	// Bearings in strike px: exact integers because ox/oy were grid-locked above.
	bx := float32(-ox / s)
	by := float32(-oy / s)
	return down, bx, by, true
}

// boxDownsampleGlyphCoverage area-averages coverage alpha by factor×factor.
// RGB stays white; output is 1× sticky atlas cells.
func boxDownsampleGlyphCoverage(src image.Image, factor int) *image.RGBA {
	if factor < 2 {
		b := src.Bounds()
		out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		blitFTGreyImage(out, out.Bounds(), src)
		return out
	}
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dw := (sw + factor - 1) / factor
	dh := (sh + factor - 1) / factor
	if dw < 1 || dh < 1 {
		return nil
	}
	out := image.NewRGBA(image.Rect(0, 0, dw, dh))
	area := factor * factor
	for dy := 0; dy < dh; dy++ {
		for dx := 0; dx < dw; dx++ {
			sum := 0
			n := 0
			for oy := 0; oy < factor; oy++ {
				sy := sb.Min.Y + dy*factor + oy
				if sy >= sb.Max.Y {
					break
				}
				for ox := 0; ox < factor; ox++ {
					sx := sb.Min.X + dx*factor + ox
					if sx >= sb.Max.X {
						break
					}
					_, _, _, a := src.At(sx, sy).RGBA()
					sum += int(a >> 8)
					n++
				}
			}
			if n == 0 {
				continue
			}
			// Prefer full-factor area; edge tiles use actual sample count.
			den := area
			if n < area {
				den = n
			}
			aa := sum / den
			if aa <= 0 {
				continue
			}
			if aa > 255 {
				aa = 255
			}
			i := out.PixOffset(dx, dy)
			out.Pix[i+0] = 255
			out.Pix[i+1] = 255
			out.Pix[i+2] = 255
			out.Pix[i+3] = uint8(aa)
		}
	}
	return out
}

func blitFTGreyImage(atlas *image.RGBA, dst image.Rectangle, src image.Image) int {
	ink := 0
	sb := src.Bounds()
	for py := dst.Min.Y; py < dst.Max.Y; py++ {
		for px := dst.Min.X; px < dst.Max.X; px++ {
			_, _, _, a := src.At(sb.Min.X+(px-dst.Min.X), sb.Min.Y+(py-dst.Min.Y)).RGBA()
			if a == 0 {
				continue
			}
			aa := uint8(a >> 8)
			i := atlas.PixOffset(px, py)
			atlas.Pix[i+0] = 255
			atlas.Pix[i+1] = 255
			atlas.Pix[i+2] = 255
			atlas.Pix[i+3] = aa
			ink++
		}
	}
	return ink
}

func logFTGreyOutlineBake() {
	ftGreyOutlineOnce.Do(func() {
		backend := "gg"
		if fxGlyphBackendActive() {
			backend = "fx"
		}
		fmt.Printf("Gru: ft-grey sticky outline AA via %s @%d× bake→1× (Windows DWrite analogue on Linux)\n",
			backend, stickyOutlineBakeSSAA())
	})
}
