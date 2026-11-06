package ui

import (
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	xfixed "golang.org/x/image/math/fixed"
)

func TestFTGreyParseAndRaster(t *testing.T) {
	path := firstExisting(fontCandidates...)
	if path == "" {
		t.Skip("no UI font")
	}
	data, err := ReadAssetFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := truetype.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:    32,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	defer face.Close()
	dr, mask, origin, adv, ok := face.Glyph(xfixed.Point26_6{}, 'A')
	if !ok || mask == nil {
		t.Fatal("Glyph A failed")
	}
	if dr.Empty() || adv <= 0 {
		t.Fatalf("empty glyph dr=%v adv=%v origin=%v", dr, adv, origin)
	}
}

func TestPackFTGreyAtlasHasInk(t *testing.T) {
	path := firstExisting(fontCandidates...)
	if path == "" {
		t.Skip("no UI font")
	}
	data, err := ReadAssetFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := truetype.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              48,
		DPI:               72,
		Hinting:           font.HintingFull,
		GlyphCacheEntries: 8,
	})
	defer face.Close()
	atlas, glyphs, ink, ok := packFTGreyAtlas(face, 48)
	if !ok {
		t.Fatalf("pack failed ink=%d glyphs=%d", ink, len(glyphs))
	}
	if ink < 100 {
		t.Fatalf("expected substantial ink, got %d", ink)
	}
	g, ok := glyphs['A']
	if !ok || !g.hasInk || g.src.Width < 2 || g.src.Height < 2 {
		t.Fatalf("glyph A missing or empty: %+v", g)
	}
	// Sample atlas center of A for non-zero alpha.
	cx := int(g.src.X + g.src.Width/2)
	cy := int(g.src.Y + g.src.Height/2)
	_, _, _, a := atlas.At(cx, cy).RGBA()
	if a == 0 {
		// Center of some glyphs can be a hole; scan the rect.
		found := false
		for y := int(g.src.Y); y < int(g.src.Y+g.src.Height); y++ {
			for x := int(g.src.X); x < int(g.src.X+g.src.Width); x++ {
				_, _, _, aa := atlas.At(x, y).RGBA()
				if aa > 0 {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Fatal("glyph A atlas region has zero alpha")
		}
	}
}

func TestPreferFTGreyChromeDefaultOn(t *testing.T) {
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_GREY", "")
		t.Setenv("GORY_TEXT_GREY", "")
	})
	t.Setenv("GRU_TEXT_GREY", "")
	t.Setenv("GORY_TEXT_GREY", "")
	if !PreferFTGreyChrome() {
		t.Fatal("default should prefer ft-grey")
	}
	t.Setenv("GRU_TEXT_GREY", "0")
	if PreferFTGreyChrome() {
		t.Fatal("GRU_TEXT_GREY=0 must force SDF")
	}
	t.Setenv("GRU_TEXT_GREY", "1")
	if !PreferFTGreyChrome() {
		t.Fatal("env=1 enables")
	}
}

// Path A ship floor — opt-out via GRU_TEXT_STRIKES=0. Unset env ⇒ S1 strikes on.
func TestFTGreyS1StrikesDefault(t *testing.T) {
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_STRIKES", "")
		t.Setenv("GORY_TEXT_STRIKES", "")
	})
	t.Setenv("GRU_TEXT_STRIKES", "")
	t.Setenv("GORY_TEXT_STRIKES", "")
	if !PreferFTGreyStrikes() {
		t.Fatal("S1: unset env must enable strikes (default-on)")
	}
	t.Setenv("GRU_TEXT_STRIKES", "0")
	if PreferFTGreyStrikes() {
		t.Fatal("Path A: GRU_TEXT_STRIKES=0 must opt out")
	}
	t.Setenv("GRU_TEXT_STRIKES", "1")
	if !PreferFTGreyStrikes() {
		t.Fatal("S1: GRU_TEXT_STRIKES=1 must still enable")
	}
	if FTGreyAtlasBasePx != 48 {
		t.Fatalf("Path A: FTGreyAtlasBasePx=%d want 48", FTGreyAtlasBasePx)
	}
}

// S1 ship floor — rem release must not always force layout (docs/S1_SHIP_FLOOR.md).
func TestS1ResizeRemReleaseContract(t *testing.T) {
	t.Cleanup(func() {
		SetTypeScaleResizeHold(false)
		RootFontSize = TypeScaleReference
	})
	RootFontSize = 24
	SetTypeScaleResizeHold(true)
	RefreshTypeScaleFromWindow(2000, 1000) // would set 25 if not held
	if RootFontSize != 24 {
		t.Fatalf("hold must freeze rem, got %v", RootFontSize)
	}
	SetTypeScaleResizeHold(false)
	prev := RootFontSize
	RefreshTypeScaleFromWindow(2000, 1000)
	if !S1RemReleaseShouldRelayout(prev, RootFontSize) && prev == RootFontSize {
		// ok — helper agrees no relayout
	}
	if S1RemReleaseShouldRelayout(24, 24) {
		t.Fatal("identical rem must not request relayout")
	}
	if !S1RemReleaseShouldRelayout(24, 25) {
		t.Fatal("rem change must request relayout")
	}
}

// S1 ship floor — pick/measure must never create strikes (no hot-path bake).
func TestS1StrikeHotPathNoBake(t *testing.T) {
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_STRIKES", "")
		ftGreyReady = false
		ftGreyStrikes = nil
		ftGreyGlyphs = nil
		ftGreyBasePx = 0
	})
	t.Setenv("GRU_TEXT_STRIKES", "1")
	ftGreyReady = true
	ftGreyBasePx = 48
	ftGreyAscentPx = 40
	ftGreyGlyphs = map[rune]ftGreyGlyph{'A': {advance: 10, hasInk: false}}
	ftGreyStrikes = map[int]*ftGreyStrike{}
	before := len(ftGreyStrikes)
	_, _, _, _, exact, _ := ftGreyPick(18)
	if exact {
		t.Fatal("empty strike map must fall back to Path A (exact=false)")
	}
	if len(ftGreyStrikes) != before {
		t.Fatal("ftGreyPick must not bake strikes")
	}
}

func TestApplyFTGreyGammaContrast(t *testing.T) {
	atlas := image.NewRGBA(image.Rect(0, 0, 1, 1))
	atlas.SetRGBA(0, 0, color.RGBA{255, 255, 255, 128})
	before := atlas.Pix[3]
	applyFTGreyGammaContrast(atlas, 1.45, 0.55)
	if atlas.Pix[3] == before {
		t.Fatalf("LUT should remap mid alpha (before=%d after=%d)", before, atlas.Pix[3])
	}
}

func TestFTGreySubpixelPhase(t *testing.T) {
	if ftGreySubpixelPhase(10.0) != 0 {
		t.Fatal("0.0 → phase 0")
	}
	if ftGreySubpixelPhase(10.1) != 0 {
		t.Fatal("0.1 → phase 0")
	}
	if ftGreySubpixelPhase(10.25) != 1 {
		t.Fatal("0.25 → phase 1")
	}
	if ftGreySubpixelPhase(10.4) != 1 {
		t.Fatal("0.4 → phase 1")
	}
	if ftGreySubpixelPhase(10.5) != 2 {
		t.Fatal("0.5 → phase 2")
	}
	if ftGreySubpixelPhase(10.7) != 2 {
		t.Fatal("0.7 → phase 2")
	}
	if ftGreySubpixelPhase(10.75) != 3 {
		t.Fatal("0.75 → phase 3")
	}
	if ftGreySubpixelPhase(10.9) != 3 {
		t.Fatal("0.9 → phase 3")
	}
}

func TestStickyStrikeSizesDWriteBudget(t *testing.T) {
	wantSizes := []int{14, 16, 18, 20, 24, 28, 32, 40}
	if len(ftGreyStrikeStickyDWrite) != len(wantSizes) {
		t.Fatalf("sticky base=%v want %v", ftGreyStrikeStickyDWrite, wantSizes)
	}
	for i, sz := range wantSizes {
		if ftGreyStrikeStickyDWrite[i] != sz {
			t.Fatalf("sticky[%d]=%d want %d", i, ftGreyStrikeStickyDWrite[i], sz)
		}
	}
	t.Setenv("GRU_TEXT_STRIKES", "1")
	t.Setenv("GRU_SSAA", "1")
	t.Setenv("GRU_TEXT_LCD", "1")
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_STRIKES", "")
		t.Setenv("GRU_SSAA", "")
		t.Setenv("GRU_TEXT_LCD", "")
	})
	// Without live DWrite, stickyStrikeSizes may use greyscale ladder — still ≤8.
	got := stickyStrikeSizes()
	if len(got) == 0 || len(got) > ftGreyStrikeMax {
		t.Fatalf("stickyStrikeSizes len=%d max %d got %v", len(got), ftGreyStrikeMax, got)
	}
}

func TestScaleStickyStrikeSizesDPI(t *testing.T) {
	got := scaleStickyStrikeSizes([]int{14, 16, 20, 24, 32, 40}, 1)
	if len(got) != 6 || got[0] != 14 || got[5] != 40 {
		t.Fatalf("dpi1 = %v", got)
	}
	got = scaleStickyStrikeSizes([]int{14, 16, 20, 24, 32, 40}, 1.25)
	want := []int{18, 20, 25, 30, 40, 50}
	if len(got) != len(want) {
		t.Fatalf("dpi1.25 len=%v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("dpi1.25[%d]=%d want %d (full %v)", i, got[i], want[i], got)
		}
	}
}

func TestQuantizeFTGreyChromeCaptionFloor(t *testing.T) {
	prevDPI := DisplayScale
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_STRIKES", "")
		t.Setenv("GORY_TEXT_STRIKES", "")
		DisplayScale = prevDPI
	})
	t.Setenv("GRU_TEXT_STRIKES", "1")
	DisplayScale = 1.25
	// Client-px ladder includes 14 — floor stays 14 (not inflated by DPI).
	q := QuantizeFTGreyChromeSize(14)
	if q != 14 {
		t.Fatalf("caption 14 → %v want 14 client px", q)
	}
	q16 := QuantizeFTGreyChromeSize(16)
	if q16 != 16 {
		t.Fatalf("caption 16 → %v want 16", q16)
	}
}

func TestEffectiveFontSizeS1NoDPIMultiply(t *testing.T) {
	prevRoot := RootFontSize
	prevDPI := DisplayScale
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_STRIKES", "")
		t.Setenv("GORY_TEXT_STRIKES", "")
		t.Setenv("GRU_SSAA", "")
		RootFontSize = prevRoot
		DisplayScale = prevDPI
	})
	t.Setenv("GRU_TEXT_STRIKES", "1")
	t.Setenv("GRU_SSAA", "1")
	RootFontSize = 16
	DisplayScale = 1.25
	st := GetThemeStyle("list-tile")
	fs := EffectiveFontSize(st)
	// token 18 × rem 1.0 = 18 — must NOT become 18×1.25=22.5
	if fs < 16 || fs > 20 {
		t.Fatalf("list-tile S1 @ dpi1.25 = %v want ~18 (no DPI multiply)", fs)
	}
}

func TestEnhanceS1DWriteLCDAtlasLight(t *testing.T) {
	atlas := image.NewRGBA(image.Rect(0, 0, 1, 1))
	atlas.SetRGBA(0, 0, color.RGBA{128, 140, 120, 255})
	before := [3]uint8{atlas.Pix[0], atlas.Pix[1], atlas.Pix[2]}
	enhanceS1DWriteLCDAtlas(atlas)
	// DWrite atlases must stay untouched (no bake contrast).
	if atlas.Pix[0] != before[0] || atlas.Pix[1] != before[1] || atlas.Pix[2] != before[2] {
		t.Fatal("DWrite enhance must be a no-op to preserve thin crossbars")
	}
}

func TestQuantizeFTGreyChromeSize(t *testing.T) {
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_STRIKES", "")
		t.Setenv("GORY_TEXT_STRIKES", "")
	})
	// Path A is explicit opt-out — empty env means S1 strikes ON.
	t.Setenv("GRU_TEXT_STRIKES", "0")
	t.Setenv("GORY_TEXT_STRIKES", "")
	if q := QuantizeFTGreyChromeSize(19.25); q != 19.25 {
		t.Fatalf("Path A must not quantize, got %v", q)
	}
	t.Setenv("GRU_TEXT_STRIKES", "1")
	q := QuantizeFTGreyChromeSize(19.25)
	if q != 18 && q != 20 {
		t.Fatalf("S1 quantize 19.25 → %v want 18 or 20", q)
	}
	q2 := QuantizeFTGreyChromeSize(24.75)
	if q2 != 24 && q2 != 20 && q2 != 32 {
		t.Fatalf("S1 quantize 24.75 → %v want sticky neighbor", q2)
	}
}

func TestChromeFontSizeSnapsUnderS1(t *testing.T) {
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_STRIKES", "")
		t.Setenv("GORY_TEXT_STRIKES", "")
	})
	t.Setenv("GRU_TEXT_STRIKES", "1")
	// Calendar used DrawText(13) without snap — muddy vs form-label drawTextS.
	got := chromeFontSize(13)
	ladder := stickyStrikeSizes()
	onLadder := false
	for _, sz := range ladder {
		if got == float32(sz) {
			onLadder = true
			break
		}
	}
	if !onLadder {
		t.Fatalf("chromeFontSize(13)=%v not on sticky ladder %v", got, ladder)
	}
}

func TestFTGreyPickNearestStrike(t *testing.T) {
	t.Cleanup(func() {
		ftGreyStrikes = nil
	})
	ftGreyStrikes = map[int]*ftGreyStrike{
		18: {basePx: 18, ascent: 14, glyphs: map[rune]ftGreyGlyph{}},
		20: {basePx: 20, ascent: 16, glyphs: map[rune]ftGreyGlyph{}},
		22: {basePx: 22, ascent: 18, glyphs: map[rune]ftGreyGlyph{}},
	}
	// Segmented token 14 @ rem 22 → ~19.25 — must hit nearest sticky, not Path A.
	s, ok := ftGreyPickNearestStrike(19.25)
	if !ok || s == nil {
		t.Fatal("expected nearest strike for 19.25")
	}
	if s.basePx != 18 && s.basePx != 20 {
		t.Fatalf("nearest for 19.25 = %v want 18 or 20", s.basePx)
	}
	s2, ok := ftGreyPickNearestStrike(21)
	if !ok || s2 == nil {
		t.Fatal("expected nearest for 21")
	}
	if s2.basePx != 20 && s2.basePx != 22 {
		t.Fatalf("nearest for 21 = %v want 20 or 22", s2.basePx)
	}
	if _, ok := ftGreyPickNearestStrike(40); ok {
		t.Fatal("40px must miss scale band when only 18–22 sticky")
	}
}

func TestTextEditorRemLandsOnSticky(t *testing.T) {
	prevRoot := RootFontSize
	prevBase := BaseSupersamplingScale
	prevDPI := DisplayScale
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_STRIKES", "")
		t.Setenv("GORY_TEXT_STRIKES", "")
		t.Setenv("GRU_SSAA", "")
		t.Setenv("GORY_SSAA", "")
		t.Setenv("GRU_TEXT_LCD", "")
		RootFontSize = prevRoot
		BaseSupersamplingScale = prevBase
		DisplayScale = prevDPI
	})
	t.Setenv("GRU_TEXT_STRIKES", "1")
	t.Setenv("GRU_SSAA", "1")
	t.Setenv("GRU_TEXT_LCD", "")
	RootFontSize = 24
	BaseSupersamplingScale = 1
	DisplayScale = 1
	st := GetThemeStyle("text-editor")
	fs := EffectiveFontSize(st)
	// 18 × 1.5 rem = 27 → sticky neighbor 28 (client px).
	if fs != 28 && fs != 24 && fs != 32 {
		t.Fatalf("text-editor at rem24 dpi1 = %v want ~28", fs)
	}
	DisplayScale = 1.25
	fs = EffectiveFontSize(st)
	// DPI must not change EffectiveFontSize (client-px contract).
	if fs != 28 && fs != 24 && fs != 32 {
		t.Fatalf("text-editor at rem24 dpi1.25 = %v want same ~28 (no DPI multiply)", fs)
	}
}

func TestFTGreyAscentFromGlyphs(t *testing.T) {
	glyphs := map[rune]ftGreyGlyph{
		'D': {bearingY: -34, hasInk: true},
		'x': {bearingY: -20, hasInk: true},
	}
	a := ftGreyAscentFromGlyphs(glyphs, 28)
	if a < 34 {
		t.Fatalf("ascent=%v want >=34 from D bearing", a)
	}
}

func TestFTGreyPreferredBodyFloorsRem(t *testing.T) {
	prevRoot := RootFontSize
	prevDPI := DisplayScale
	t.Cleanup(func() {
		RootFontSize = prevRoot
		DisplayScale = prevDPI
	})
	DisplayScale = 1.25
	RootFontSize = 16 // pre-atlas-init default
	body := ftGreyPreferredBodyPx()
	// Floor rem→TypeScaleMinRoot (18): 18*(18/16) = 20.25 client px (no DPI).
	want := 18 * (TypeScaleMinRoot / TypeScaleReference)
	if body < want-0.5 || body > want+0.5 {
		t.Fatalf("preferred body with rem16 floor = %v want ~%v", body, want)
	}
}

func TestFTGreyLCDRejectsShakyScale(t *testing.T) {
	t.Cleanup(func() { ftGreyStrikes = nil })
	ftGreyStrikes = map[int]*ftGreyStrike{
		30: {basePx: 30, ascent: 24, glyphs: map[rune]ftGreyGlyph{}, lcd: true},
	}
	if _, ok := ftGreyPickNearestStrike(32); ok {
		t.Fatal("32→30 LCD scale must miss (shaky stretch band)")
	}
	s, ok := ftGreyPickNearestStrike(30)
	if !ok || s == nil || s.basePx != 30 {
		t.Fatal("exact 30 LCD must hit")
	}
}

func TestSnapFTGreyStrikeSizeExact(t *testing.T) {
	if g := snapFTGreyStrikeSize(18); g != 18 {
		t.Fatalf("snap(18)=%d want 18", g)
	}
	if g := snapFTGreyStrikeSize(24.4); g != 24 {
		t.Fatalf("snap(24.4)=%d want 24", g)
	}
	if g := snapFTGreyStrikeSize(24.6); g != 25 {
		t.Fatalf("snap(24.6)=%d want 25", g)
	}
	sz := snapFTGreyStrikeSize(16)
	scale := ftGreyStrikeScale(16, float32(sz))
	if absFloat32(scale-1) > 0.04 {
		t.Fatalf("strike scale for 16px = %v want ~1", scale)
	}
	// Path A optical must stay on single-atlas helper.
	pa := ftGreyScale(16, 48)
	want := (float32(16) / 48) * ftGreyOpticalScale
	if absFloat32(pa-want) > 0.001 {
		t.Fatalf("Path A scale drifted: %v want %v", pa, want)
	}
}

// TestFTGreyOutlineBaselineLock guards the "100:" float bug: outline bakes must
// have integer bearings and digits must share one baseline (bottom of "1" and
// "0" differ only by the round overshoot of "0", never by a whole-pixel drift).
func TestFTGreyOutlineBaselineLock(t *testing.T) {
	data, err := os.ReadFile("../assets/fonts/poppins/Poppins-Regular.ttf")
	if err != nil {
		t.Skip("no font")
	}
	for _, px := range []int{12, 13, 14, 15, 16, 18, 22} {
		_, glyphs, _, ok := packFTGreyAtlasOutline(data, px, 2)
		if !ok {
			t.Fatalf("pack failed @%d", px)
		}
		for r, g := range glyphs {
			if !g.hasInk {
				continue
			}
			if g.bearingX != float32(int(g.bearingX)) || g.bearingY != float32(int(g.bearingY)) {
				t.Fatalf("@%d %q fractional bearing (%v,%v)", px, r, g.bearingX, g.bearingY)
			}
		}
		one, zero := glyphs['1'], glyphs['0']
		bottom1 := one.bearingY + one.src.Height
		bottom0 := zero.bearingY + zero.src.Height
		d := bottom1 - bottom0
		if d < -1 || d > 1 {
			t.Fatalf("@%d digit baselines drift: '1' bottom=%v '0' bottom=%v", px, bottom1, bottom0)
		}
	}
}
