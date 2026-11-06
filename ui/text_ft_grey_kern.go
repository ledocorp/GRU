package ui

import (
	"fmt"
	"math"
	"sync"

	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/shaping"
)

// Sticky greyscale draws from a per-rune atlas (no ligature substitution). HarfBuzz
// supplies GPOS kerning with liga/clig/calt off so advances stay 1 glyph ↔ 1 rune.
// This is the spacing half of “DWrite on Linux”: raster stays outline-AA; metric
// pairing matches what Inter/Poppins expect.

var (
	ftGreyKernOnce    sync.Once
	ftGreyKernReady   bool
	ftGreyKernLogOnce sync.Once
	stickyKernFaces   sync.Map // path → *font.Face (same TTF as outline bake)
)

func ensureFTGreyKern() bool {
	ftGreyKernOnce.Do(func() {
		ftGreyKernReady = InitShapedFonts()
		if ftGreyKernReady {
			ftGreyKernLogOnce.Do(func() {
				fmt.Println("Gru: ft-grey sticky kerning via HarfBuzz (liga off; GPOS kern)")
			})
		}
	})
	return ftGreyKernReady
}

// stickyKernFeatures: keep kern/GPOS; disable ligatures so rune atlas draw stays 1:1.
func stickyKernFeatures() []shaping.FontFeature {
	return []shaping.FontFeature{
		{Tag: ot.MustNewTag("liga"), Value: 0},
		{Tag: ot.MustNewTag("clig"), Value: 0},
		{Tag: ot.MustNewTag("calt"), Value: 0},
		{Tag: ot.MustNewTag("dlig"), Value: 0},
	}
}

func stickyKernFaceForPath(path string) *font.Face {
	if path == "" {
		return nil
	}
	if v, ok := stickyKernFaces.Load(path); ok {
		return v.(*font.Face)
	}
	f, err := loadShapedFace(path)
	if err != nil || f == nil {
		return nil
	}
	stickyKernFaces.Store(path, f)
	return f
}

// pickStickyKernFace picks the same TTF the sticky outline strike baked from.
// shapedUI.bold is Poppins-Bold (700); chrome bold strikes use Medium/SemiBold —
// GPOS on Bold vs Medium was a systematic spacing mismatch on Scenes / titles.
func pickStickyKernFace(sz int, bold, mono, preview bool) *font.Face {
	if mono {
		return pickShapedFace(bold, false, true, false)
	}
	if preview {
		if bold {
			if f := stickyKernFaceForPath(interFaceForStickySize(sz, true)); f != nil {
				return f
			}
		}
		return pickShapedFace(bold, false, false, true)
	}
	if bold {
		path := resolveFTGreyChromeBoldPathForSize(ftGreyRegularPath, sz)
		if f := stickyKernFaceForPath(path); f != nil {
			return f
		}
	}
	return pickShapedFace(bold, false, false, false)
}

// ftGreyKernAdvances returns per-rune pen advances (pixels) including kerning.
// Length equals rune count. ok=false → caller uses atlas advances.
func ftGreyKernAdvances(text string, fontSize float32, bold, mono, preview bool) ([]float32, bool) {
	if text == "" || !ensureFTGreyKern() {
		return nil, false
	}
	if shapedTextUsesComplexScript(text) || shapedNeedsSDFFallbackMeasure(text) {
		return nil, false
	}
	// Match sticky strike size so kern amounts sit on the same grid as atlas advances.
	sz := fontSize
	if PreferFTGreyStrikes() {
		if mono || preview {
			sz = QuantizeFTGreyAuxSize(fontSize)
		} else {
			sz = QuantizeFTGreyChromeSize(fontSize)
		}
	}
	face := pickStickyKernFace(int(sz+0.5), bold, mono, preview)
	if face == nil {
		return nil, false
	}
	runes := []rune(text)
	out := hbShaper.Shape(shaping.Input{
		Text:         runes,
		RunStart:     0,
		RunEnd:       len(runes),
		Direction:    shapedTextDirection(runes),
		Face:         face,
		Size:         shapedSizePx(sz),
		Script:       shapedScriptFor(runes),
		Language:     shapedTextLanguage(runes),
		FontFeatures: stickyKernFeatures(),
	})
	if len(out.Glyphs) == 0 {
		return nil, false
	}
	advs := make([]float32, len(runes))
	for i := range advs {
		advs[i] = -1 // unset
	}
	for _, g := range out.Glyphs {
		ci := g.ClusterIndex
		if ci < 0 || ci >= len(runes) {
			return nil, false
		}
		// Ligature slipped through: refuse shaped advances (atlas can't draw it).
		if g.RuneCount > 1 {
			return nil, false
		}
		px := float32(g.XAdvance) / 64
		if advs[ci] < 0 {
			advs[ci] = px
		} else {
			advs[ci] += px
		}
	}
	for i := range advs {
		if advs[i] < 0 {
			return nil, false
		}
		// Integer advances for FilterPoint. Fold FontSpacing into the round so
		// draw keeps an integer pen — fractional +0.3 then snapPresentPx(pen)
		// each glyph discarded tracking and fought bearings ("Cl ose", "th in").
		extra := float32(0)
		if FontSpacing != 0 && i < len(advs)-1 {
			extra = FontSpacing
		}
		advs[i] = float32(math.Round(float64(advs[i]+extra)))
	}
	return advs, true
}
