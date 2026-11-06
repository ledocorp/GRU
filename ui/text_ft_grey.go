package ui

import (
	"fmt"
	"image"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/image/font"
	xfixed "golang.org/x/image/math/fixed"
)

// T1.13 — FreeType (pure-Go) greyscale Latin chrome.
// Same Face advances for measure and draw — safe for RichText wrap (unlike outline Latin).
//
// Policy: on by default. Opt out with GRU_TEXT_GREY=0.
//
// S1 (GRU_TEXT_STRIKES=0 opts out): sticky exact-px strikes + optional LCD @ SSAA 1×.
// No bake on draw/measure (startup prewarm only) — stops resize/tab hitch.
// Gamma/contrast + optional LCD RGB strikes when PreferFTLCDChrome (auto @ SSAA 1×).
// See docs/TEXT_NO_SSAA_RESEARCH.md §4.

type ftGreyGlyph struct {
	src      rl.Rectangle
	bearingX float32
	bearingY float32
	advance  float32
	hasInk   bool
}

type ftGreyPhase struct {
	origin float32
	tex    rl.Texture2D
	glyphs map[rune]ftGreyGlyph
}

type ftGreyStrike struct {
	basePx float32
	ascent float32
	// Phase 0 (origin 0) — always present; ftGreyPick returns these.
	tex    rl.Texture2D
	glyphs map[rune]ftGreyGlyph
	// Extra ClearType X-bins (Skia-class). Rem-body 24 holds 0.25/0.5/0.75.
	phases []ftGreyPhase
	sticky bool
	lcd    bool // RGB subpixel atlas; draw via LCD shader
}

var (
	ftGreyReady          bool
	ftGreyTex            rl.Texture2D
	ftGreyBasePx         float32
	ftGreyAscentPx       float32
	ftGreyGlyphs         map[rune]ftGreyGlyph
	ftGreyWarn           sync.Once
	ftGreyTTFData        []byte
	ftGreyTTF            *truetype.Font
	ftGreyRegularPath    string
	ftGreyBoldPath       string
	ftGreyBoldTTFData    []byte
	ftGreyBoldTTF        *truetype.Font
	ftGreyStrikes        map[int]*ftGreyStrike
	ftGreyBoldStrikes    map[int]*ftGreyStrike // Poppins Medium/SemiBold (or Segoe Bold)
	ftGreyPlainStrikes   map[int]*ftGreyStrike // grey-only sticky (transparent hosts)
	ftGreyStrikeLRU      []int
	ftGreyStrikeSSAAWarn sync.Once
)

const (
	ftGreyOpticalScale   float32 = 0.93 // Path A only
	ftGreyStrikeMax              = 8
	ftGreyStrikeScaleMin float32 = 0.85
	ftGreyStrikeScaleMax float32 = 1.20 // was 1.15 — headers 28→32 need slight stretch budget
	// LCD ClearType atlases must draw 1:1 — stretch/shrink looks "shaky" vs browser.
	ftGreyLCDExactEps float32 = 0.03
)

// PreferFTGreyStrikes enables S1 hybrid exact-px strikes. On by default; opt out
// with GRU_TEXT_STRIKES=0 (Path A single-atlas fallback).
func PreferFTGreyStrikes() bool {
	v := strings.ToLower(strings.TrimSpace(EnvOr("GRU_TEXT_STRIKES", "GORY_TEXT_STRIKES")))
	if v == "0" || v == "off" || v == "false" || v == "no" {
		return false
	}
	if v == "" {
		return true
	}
	return v == "1" || v == "on" || v == "true" || v == "yes"
}

// Sticky chrome sizes only — no runtime bake (Impeller: avoid atlas churn on interaction).
// S1 ship floor: docs/S1_SHIP_FLOOR.md — do not bake from ftGreyPick.
var ftGreyStrikeStickySizes = []int{14, 16, 18, 20, 22, 24, 28, 32}

// DWrite sticky sizes = client/framebuffer px (same space as rem × token).
// Do not × DisplayScale — layout is already physical with FlagWindowHighdpi off;
// scaling sticky + EffectiveFontSize by DPI made chrome ~25% too large (2026-08-16).
// Covers captions → body → headers within 8 atlases. CSS-compare dogfood: see Glyph Quality.
var ftGreyStrikeStickyDWrite = []int{14, 16, 18, 20, 24, 28, 32, 40}

// ftGreyPreferredBodyPx is theme token 18 × rem in client px (before quantize).
func ftGreyPreferredBodyPx() float32 {
	rem := RootFontSize
	if rem < TypeScaleMinRoot {
		rem = TypeScaleMinRoot
	}
	return 18 * (rem / TypeScaleReference)
}

// ftGreySubXEnabled reports whether the rem-body strike bakes 4 subpixel-X
// phases. Default off: DWrite hints stems onto the pixel grid at origin 0, and
// the 0.25/0.5/0.75 bins land stems across two columns. Mixing bins inside one
// run therefore reads heavier and softer than the same string at a size that
// only ever uses origin 0 — the Demo Directory body looked bold while smaller
// menu/caption chrome stayed crisp, because only the rem-body strike is 4-phase.
//
// Opt in with GRU_TEXT_SUBX=4 and A/B on the Glyph Quality ladder: the rem-body
// row must not read heavier than its neighbours before this becomes default.
func ftGreySubXEnabled() bool {
	return EnvOr("GRU_TEXT_SUBX", "GORY_TEXT_SUBX") == "4"
}

// ftGreySubXOrigins lists DWrite baseline_origin_x bins for a sticky size.
// 4-phase on rem-body sticky only (WS: caption sizes stay single-phase LCD).
func ftGreySubXOrigins(sz int) []float32 {
	if !ftGreySubXEnabled() || !PreferFTLCDChrome() || !dwriteLCDAvailable() {
		return []float32{0}
	}
	ladder := stickyStrikeSizes()
	if len(ladder) == 0 {
		return []float32{0}
	}
	body := ftGreyPreferredBodyPx()
	best := ladder[0]
	bestDist := absFloat32(float32(best) - body)
	for _, s := range ladder[1:] {
		d := absFloat32(float32(s) - body)
		if d < bestDist {
			bestDist = d
			best = s
		}
	}
	if sz == best {
		return []float32{0, 0.25, 0.5, 0.75}
	}
	return []float32{0}
}

func stickyStrikeSizes() []int {
	var base []int
	if PreferFTLCDChrome() && dwriteLCDAvailable() {
		base = ftGreyStrikeStickyDWrite
	} else {
		base = ftGreyStrikeStickySizes
	}
	// Identity client-px ladder (Linux DisplayScale=1; Windows densify via T1.9 only).
	out := make([]int, len(base))
	copy(out, base)
	return out
}

// mergeStickyDeviceKeys adds absolute device-px keys into the DPI-scaled ladder
// until max size-slots are filled (caption coverage without exceeding WS budget).
func mergeStickyDeviceKeys(ladder, extra []int, max int) []int {
	if max <= 0 {
		return ladder
	}
	seen := make(map[int]struct{}, len(ladder)+len(extra))
	out := make([]int, 0, max)
	for _, s := range ladder {
		if len(out) >= max {
			break
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	for _, e := range extra {
		if len(out) >= max {
			break
		}
		if e < 11 || e > 72 {
			continue
		}
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		out = append(out, e)
	}
	sort.Ints(out)
	return out
}

// scaleStickyStrikeSizes maps 100%-DPI sticky keys to device pixels (T1.9).
// Without this, S1 at DisplayScale=1.25 baked 24px strikes into a denser
// framebuffer — ClearType looked no sharper than soft Path A (user dogfood).
func scaleStickyStrikeSizes(base []int, dpi float32) []int {
	if len(base) == 0 {
		return nil
	}
	if dpi < 1.02 {
		out := make([]int, len(base))
		copy(out, base)
		return out
	}
	out := make([]int, 0, len(base))
	seen := map[int]struct{}{}
	for _, sz := range base {
		v := int(float32(sz)*dpi + 0.5)
		if v < 11 {
			v = 11
		}
		if v > 72 {
			v = 72
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// ftGreyStrikeWantsSubX is true when size bakes extra X-phase atlases beyond 0.
func ftGreyStrikeWantsSubX(sz int) bool {
	return len(ftGreySubXOrigins(sz)) > 1
}

// QuantizeFTGreyChromeSize snaps a rem-scaled size onto the S1 sticky ladder so
// draw uses scale≈1.0 (exact strikes). Larger titles already look crisp; small
// chrome was soft because rem left fractional sizes (e.g. 19.25) off the ladder.
// Path A (strikes off): returns fs unchanged — do not alter frozen rem sizing.
//
// S1 hard snap: chrome below the ladder floors to the smallest sticky (never
// Path A soft-scale for captions). Within the scale band, always return sticky.
func QuantizeFTGreyChromeSize(fs float32) float32 {
	if !PreferFTGreyStrikes() {
		return fs
	}
	if fs < minRenderPx {
		fs = minRenderPx
	}
	rounded := float32(int(fs + 0.5))
	if rounded < minRenderPx {
		rounded = minRenderPx
	}
	ladder := stickyStrikeSizes()
	if len(ladder) == 0 {
		return rounded
	}
	lo, hi := ladder[0], ladder[len(ladder)-1]
	if fs <= float32(lo) {
		return float32(lo)
	}
	best := lo
	bestDist := absFloat32(fs - float32(best))
	for _, sz := range ladder[1:] {
		d := absFloat32(fs - float32(sz))
		if d < bestDist {
			bestDist = d
			best = sz
		}
	}
	scale := ftGreyStrikeScale(fs, float32(best))
	if scale >= ftGreyStrikeScaleMin && scale <= ftGreyStrikeScaleMax {
		return float32(best)
	}
	// Near ladder ends: still prefer sticky over soft Path A (titles → hi).
	if bestDist <= 8 && (best == lo || best == hi || fs <= float32(hi)*1.15) {
		return float32(best)
	}
	return rounded
}

// S1RemReleaseShouldRelayout reports whether rem freeze release must ForceFullLayout.
// Guarded by TestS1ResizeRemReleaseContract / docs/S1_SHIP_FLOOR.md.
func S1RemReleaseShouldRelayout(prevRem, nextRem float32) bool {
	return prevRem != nextRem
}

func ftGreyScale(fontSize, basePx float32) float32 {
	if basePx <= 0 {
		return 1
	}
	return (fontSize / basePx) * ftGreyOpticalScale
}

// ftGreyStrikeScale is 1:1 device mapping (no Path A optical).
func ftGreyStrikeScale(fontSize, basePx float32) float32 {
	if basePx <= 0 {
		return 1
	}
	return fontSize / basePx
}

// PreferFTGreyChrome reports Latin chrome should use FreeType greyscale (not SDF).
func PreferFTGreyChrome() bool {
	v := strings.ToLower(strings.TrimSpace(EnvOr("GRU_TEXT_GREY", "GORY_TEXT_GREY")))
	switch v {
	case "0", "off", "false", "no", "sdf":
		return false
	case "1", "on", "true", "yes", "grey", "gray":
		return true
	default:
		return true
	}
}

// InitFTGreyFont builds the fallback greyscale atlas + enables exact-size strikes.
func InitFTGreyFont(pixelSize int32) bool {
	unloadFTGreyFont()
	if pixelSize < 12 {
		pixelSize = 12
	}
	path := LoadedUIFontPath
	if path == "" {
		path = firstExisting(resolvedFontCandidates()...)
	}
	if path == "" {
		return false
	}
	data, err := ReadAssetFile(path)
	if err != nil {
		return false
	}
	ttf, err := truetype.Parse(data)
	if err != nil {
		return false
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              float64(pixelSize),
		DPI:               72,
		Hinting:           font.HintingFull,
		GlyphCacheEntries: 8,
	})
	ascent := float32(face.Metrics().Ascent.Ceil())
	atlas, glyphs, ink, ok := packFTGreyAtlas(face, pixelSize)
	_ = face.Close()
	if !ok {
		ftGreyWarn.Do(func() {
			fmt.Printf("Gru: ft-grey atlas failed (ink=%d) — staying on SDF\n", ink)
		})
		return false
	}
	tex := GoImageToTexture(atlas)
	if tex.ID == 0 {
		return false
	}
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	ftGreyGlyphs = glyphs
	ftGreyTex = tex
	ftGreyBasePx = float32(pixelSize)
	ftGreyAscentPx = ascent
	ftGreyReady = true // must be set before prewarm — ensureFTGreyStrike gates on ready
	// Path A atlas always retained. S1 keeps TTF + sticky strike set on top.
	if PreferFTGreyStrikes() {
		ftGreyTTFData = data
		ftGreyTTF = ttf
		ftGreyStrikes = make(map[int]*ftGreyStrike)
		ftGreyBoldStrikes = make(map[int]*ftGreyStrike)
		ftGreyStrikeLRU = ftGreyStrikeLRU[:0]
		ftGreyRegularPath = path
		loadFTGreyChromeBoldFace(path)
		ftGreyStrikeSSAAWarn.Do(func() {
			if EffectiveSupersamplingScale() > 1.01 {
				fmt.Println("Gru: ft-grey+strikes — dogfood with GRU_SSAA=1 (scene SSAA>1 fights exact strikes)")
			} else if PreferFTLCDChrome() && dwriteLCDAvailable() {
				fmt.Println("Gru: ft-grey+strikes+dwrite-lcd hybrid @ SSAA 1× (ClearType+subX; Path A fallback)")
			} else if PreferFTLCDChrome() {
				fmt.Println("Gru: ft-grey+strikes+lcd hybrid @ SSAA 1× (lite LCD; Path A fallback)")
			} else {
				fmt.Println("Gru: ft-grey+strikes hybrid @ SSAA 1× (sticky exact; Path A fallback)")
			}
		})
		if PreferFTLCDChrome() {
			_ = ensureFTLCDShader()
		}
		prewarmFTGreyStickyStrikes()
		fmt.Printf("Gru: ft-grey sticky strikes baked=%d bold=%d ladder=%v\n",
			len(ftGreyStrikes), len(ftGreyBoldStrikes), stickyStrikeSizes())
	} else {
		ftGreyTTFData = nil
		ftGreyTTF = nil
		ftGreyBoldTTFData = nil
		ftGreyBoldTTF = nil
		ftGreyBoldPath = ""
		ftGreyStrikes = nil
		ftGreyBoldStrikes = nil
		ftGreyStrikeLRU = nil
	}
	return true
}

func chromeRegularIsPoppins(regularPath string) bool {
	b := strings.ToLower(baseName(regularPath))
	return strings.Contains(b, "poppins")
}

func poppinsChromeMediumCandidates() []string {
	return []string{
		"assets/fonts/poppins/Poppins-Medium.ttf",
		"assets/fonts/Poppins/static/Poppins-Medium.ttf",
		"assets/fonts/Poppins-Medium.ttf",
	}
}

func poppinsChromeSemiBoldCandidates() []string {
	return []string{
		"assets/fonts/poppins/Poppins-SemiBold.ttf",
		"assets/fonts/Poppins/static/Poppins-SemiBold.ttf",
		"assets/fonts/Poppins-SemiBold.ttf",
	}
}

// resolveFTGreyChromeBoldPath picks a chrome "Bold" face.
// Poppins Bold (700) reads much heavier than Segoe Bold under greyscale outline AA —
// that made Studio look "deep bold" once chrome Bold stopped being a no-op. Match Inter
// preview policy: Medium ≤18px, else SemiBold, else Bold. Windows Segoe stays segoeuib.
func resolveFTGreyChromeBoldPath(regularPath string) string {
	return resolveFTGreyChromeBoldPathForSize(regularPath, 20)
}

func resolveFTGreyChromeBoldPathForSize(regularPath string, sz int) string {
	if chromeRegularIsPoppins(regularPath) || (regularPath == "" && firstExisting("assets/fonts/poppins/Poppins-Regular.ttf") != "") {
		if sz <= 18 {
			if p := firstExisting(poppinsChromeMediumCandidates()...); p != "" {
				return p
			}
		}
		if p := firstExisting(poppinsChromeSemiBoldCandidates()...); p != "" {
			return p
		}
	}
	for _, set := range poppinsFaceCandidates() {
		if set.regular != "" && (regularPath == set.regular || baseName(regularPath) == baseName(set.regular)) {
			if p := firstExisting(set.bold); p != "" {
				return p
			}
		}
	}
	return firstExisting(
		"C:/Windows/Fonts/segoeuib.ttf",
		"assets/fonts/poppins/Poppins-Bold.ttf",
		"assets/fonts/Poppins/static/Poppins-Bold.ttf",
		"assets/fonts/poppins/Poppins-SemiBold.ttf",
		"assets/fonts/Poppins-Bold.ttf",
	)
}

func loadFTGreyChromeBoldFace(regularPath string) {
	ftGreyBoldPath = ""
	ftGreyBoldTTFData = nil
	ftGreyBoldTTF = nil
	boldPath := resolveFTGreyChromeBoldPath(regularPath)
	if boldPath == "" || boldPath == regularPath {
		return
	}
	bdata, err := ReadAssetFile(boldPath)
	if err != nil {
		return
	}
	bttf, err := truetype.Parse(bdata)
	if err != nil {
		return
	}
	ftGreyBoldPath = boldPath
	ftGreyBoldTTFData = bdata
	ftGreyBoldTTF = bttf
	fmt.Printf("Gru: ft-grey chrome soft-bold %s (≤18 Medium / else SemiBold; Bold only as fallback)\n",
		baseName(boldPath))
}

func packFTGreyAtlas(face font.Face, pixelSize int32) (*image.RGBA, map[rune]ftGreyGlyph, int, bool) {
	return packFTGreyAtlasSized(face, pixelSize, 0)
}

// packFTGreyAuxAtlas packs Inter/Fira sticky greyscale with pad=4 and ≥1024²
// (512² truncation left garbage under descenders / ghost digits).
func packFTGreyAuxAtlas(face font.Face, pixelSize int32) (*image.RGBA, map[rune]ftGreyGlyph, int, bool) {
	atlas, glyphs, ink, ok := packFTGreyAtlasSized(face, pixelSize, 4)
	if ok && auxAtlasHasASCII(glyphs) {
		return atlas, glyphs, ink, true
	}
	return packFTGreyAtlasForced(face, pixelSize, 4, 2048)
}

func auxAtlasHasASCII(glyphs map[rune]ftGreyGlyph) bool {
	if glyphs == nil {
		return false
	}
	for r := rune(33); r <= 126; r++ {
		if _, ok := glyphs[r]; !ok {
			return false
		}
	}
	return true
}

func packFTGreyAtlasForced(face font.Face, pixelSize int32, pad, dim int) (*image.RGBA, map[rune]ftGreyGlyph, int, bool) {
	glyphs := make(map[rune]ftGreyGlyph, 256)
	atlas := image.NewRGBA(image.Rect(0, 0, dim, dim))
	if pad < 2 {
		pad = 2
	}
	x, y, rowH := pad, pad, 0
	ink := 0
	truncated := false
	for _, r := range sdfCodepoints() {
		dr, mask, mp, adv, ok := face.Glyph(xfixed.Point26_6{}, r)
		if !ok {
			continue
		}
		advPx := fixed26ToPx(adv)
		gw, gh := dr.Dx(), dr.Dy()
		if mask == nil || gw <= 0 || gh <= 0 {
			glyphs[r] = ftGreyGlyph{advance: advPx}
			continue
		}
		if x+gw+pad > dim {
			x = pad
			y += rowH + pad
			rowH = 0
		}
		if y+gh+pad > dim {
			truncated = true
			break
		}
		dst := image.Rect(x, y, x+gw, y+gh)
		n := blitFTGreyMask(atlas, dst, mask, mp)
		if n == 0 {
			glyphs[r] = ftGreyGlyph{advance: advPx}
			continue
		}
		glyphs[r] = ftGreyGlyph{
			src:      rl.NewRectangle(float32(x), float32(y), float32(gw), float32(gh)),
			bearingX: float32(dr.Min.X),
			bearingY: float32(dr.Min.Y),
			advance:  advPx,
			hasInk:   true,
		}
		ink += n
		x += gw + pad
		if gh > rowH {
			rowH = gh
		}
	}
	_ = pixelSize
	if truncated || ink < 32 || len(glyphs) < 32 || !auxAtlasHasASCII(glyphs) {
		return nil, nil, ink, false
	}
	return atlas, glyphs, ink, true
}

func packFTGreyAtlasSized(face font.Face, pixelSize int32, pad int) (*image.RGBA, map[rune]ftGreyGlyph, int, bool) {
	glyphs := make(map[rune]ftGreyGlyph, 256)
	// WS: prefer 1024². Aux (pad>=4) must not auto-jump to 2048² (~16MB/tex) —
	// Inter/Fira sticky ladders were blowing Studio to ~150MB+. Truncation retries
	// via packFTGreyAuxAtlas → packFTGreyAtlasForced(..., 2048).
	atlasW, atlasH := 1024, 1024
	if pad < 4 {
		if pixelSize >= 72 {
			atlasW, atlasH = 2048, 2048
		}
		if pixelSize <= 28 {
			atlasW, atlasH = 512, 512
		}
	}
	if pad < 2 {
		pad = 2
	}
	atlas := image.NewRGBA(image.Rect(0, 0, atlasW, atlasH))
	x, y, rowH := pad, pad, 0
	ink := 0
	truncated := false
	for _, r := range sdfCodepoints() {
		dr, mask, mp, adv, ok := face.Glyph(xfixed.Point26_6{}, r)
		if !ok {
			continue
		}
		advPx := fixed26ToPx(adv)
		gw, gh := dr.Dx(), dr.Dy()
		if mask == nil || gw <= 0 || gh <= 0 {
			glyphs[r] = ftGreyGlyph{advance: advPx}
			continue
		}
		if x+gw+pad > atlasW {
			x = pad
			y += rowH + pad
			rowH = 0
		}
		if y+gh+pad > atlasH {
			truncated = true
			break
		}
		dst := image.Rect(x, y, x+gw, y+gh)
		n := blitFTGreyMask(atlas, dst, mask, mp)
		if n == 0 {
			glyphs[r] = ftGreyGlyph{advance: advPx}
			continue
		}
		glyphs[r] = ftGreyGlyph{
			src:      rl.NewRectangle(float32(x), float32(y), float32(gw), float32(gh)),
			bearingX: float32(dr.Min.X),
			bearingY: float32(dr.Min.Y),
			advance:  advPx,
			hasInk:   true,
		}
		ink += n
		x += gw + pad
		if gh > rowH {
			rowH = gh
		}
	}
	_ = pixelSize
	if truncated || ink < 32 || len(glyphs) < 32 {
		return nil, nil, ink, false
	}
	return atlas, glyphs, ink, true
}

func blitFTGreyMask(atlas *image.RGBA, dst image.Rectangle, mask image.Image, mp image.Point) int {
	ink := 0
	for py := dst.Min.Y; py < dst.Max.Y; py++ {
		for px := dst.Min.X; px < dst.Max.X; px++ {
			_, _, _, a := mask.At(mp.X+(px-dst.Min.X), mp.Y+(py-dst.Min.Y)).RGBA()
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

// applyFTGreyGammaContrast — legacy helper; S1 strikes use enhanceS1StrikeAtlas (SkMaskGamma-lite).
func applyFTGreyGammaContrast(atlas *image.RGBA, gamma, contrast float32) {
	if atlas == nil {
		return
	}
	_ = gamma
	table := buildMaskGammaLUT(maskGammaSrcDark, contrast, maskGammaDeviceDefault)
	applyMaskGammaLUTAlpha(atlas, table)
}

func snapFTGreyStrikeSize(fontSize float32) int {
	if fontSize < 11 {
		fontSize = 11
	}
	if fontSize > 72 {
		fontSize = 72
	}
	sz := int(fontSize + 0.5)
	if sz < 11 {
		sz = 11
	}
	if sz > 72 {
		sz = 72
	}
	return sz
}

func absFloat32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// ftGreyAscentFromGlyphs expands Face ascent to cover actual ink tops (DWrite
// ClearType bitmaps often extend above FT Metrics.Ascent — clips "Demo Directory").
func ftGreyAscentFromGlyphs(glyphs map[rune]ftGreyGlyph, fallback float32) float32 {
	a := fallback
	if a < 1 {
		a = 1
	}
	for _, g := range glyphs {
		if !g.hasInk {
			continue
		}
		// bearingY is typically ≤0 (above baseline). Top extent = -bearingY.
		top := -g.bearingY
		if top > a {
			a = top
		}
	}
	return a
}

// ftGreyDescentFromGlyphs is ink below the baseline (descenders), in strike px.
func ftGreyDescentFromGlyphs(glyphs map[rune]ftGreyGlyph) float32 {
	var d float32
	for _, g := range glyphs {
		if !g.hasInk || g.src.Height <= 0 {
			continue
		}
		bottom := g.bearingY + g.src.Height
		if bottom > d {
			d = bottom
		}
	}
	return d
}

func prewarmFTGreyStickyStrikes() {
	for _, sz := range stickyStrikeSizes() {
		if len(ftGreyStrikes) >= ftGreyStrikeMax {
			break
		}
		_ = ensureFTGreyStrike(sz, true)
	}
	if ftGreyBoldTTF == nil {
		return
	}
	for _, sz := range stickyStrikeSizes() {
		if len(ftGreyBoldStrikes) >= ftGreyStrikeMax {
			break
		}
		_ = ensureFTGreyBoldStrike(sz, true)
	}
}

// ftGreySubpixelPhase picks a quarter-pixel ClearType X-bin (Skia-class).
// Requires fractional penX — do not Round(pen) before this.
// 0:[0,0.25) 1:[0.25,0.5) 2:[0.5,0.75) 3:[0.75,1)
func ftGreySubpixelPhase(penX float32) int {
	f := penX - float32(math.Floor(float64(penX)))
	i := int(f * 4)
	if i < 0 {
		return 0
	}
	if i > 3 {
		return 3
	}
	return i
}

// ftGreySubpixelHalf is kept for older tests; maps mid-half to phase 2 (0.5).
func ftGreySubpixelHalf(penX float32) bool {
	return ftGreySubpixelPhase(penX) == 2 || ftGreySubpixelPhase(penX) == 1
}

// fixed26ToPx keeps fractional advances (Skia natural / DWrite natural).
// Rounding advances to ints kills subpixel phase selection mid-string.
func fixed26ToPx(a xfixed.Int26_6) float32 {
	return float32(a) / 64
}

func touchFTGreyStrike(sz int) {
	for i, v := range ftGreyStrikeLRU {
		if v == sz {
			ftGreyStrikeLRU = append(ftGreyStrikeLRU[:i], ftGreyStrikeLRU[i+1:]...)
			break
		}
	}
	ftGreyStrikeLRU = append(ftGreyStrikeLRU, sz)
}

func evictFTGreyStrikeIfNeeded() {
	// Never evict sticky; if only sticky remain at max, refuse (caller keeps Path A).
	scanned := 0
	for len(ftGreyStrikes) >= ftGreyStrikeMax && len(ftGreyStrikeLRU) > 0 && scanned < len(ftGreyStrikeLRU) {
		old := ftGreyStrikeLRU[0]
		ftGreyStrikeLRU = ftGreyStrikeLRU[1:]
		scanned++
		s, ok := ftGreyStrikes[old]
		if !ok {
			continue
		}
		if s.sticky {
			ftGreyStrikeLRU = append(ftGreyStrikeLRU, old)
			continue
		}
		if s.tex.ID != 0 {
			rl.UnloadTexture(s.tex)
		}
		for i := range s.phases {
			if s.phases[i].tex.ID != 0 {
				rl.UnloadTexture(s.phases[i].tex)
			}
		}
		delete(ftGreyStrikes, old)
		return
	}
}

// EnsurePlainStickyStrike bakes an exact-px grey (non-LCD) sticky for layered transparent hosts.
func EnsurePlainStickyStrike(px int) {
	if px < 11 {
		px = 11
	}
	ensurePlainStickyStrike(px)
}

func ensurePlainStickyStrike(sz int) *ftGreyStrike {
	if !PreferFTGreyStrikes() || !ftGreyReady || ftGreyTTF == nil {
		return nil
	}
	if ftGreyPlainStrikes == nil {
		ftGreyPlainStrikes = make(map[int]*ftGreyStrike)
	}
	if s, ok := ftGreyPlainStrikes[sz]; ok {
		return s
	}
	return bakeFTGreyStrike(sz, true, true, false)
}

// ensureFTGreyStrike bakes at init/prewarm only. Draw/measure never call this for new sizes.
func ensureFTGreyStrike(sz int, sticky bool) *ftGreyStrike {
	if !PreferFTGreyStrikes() || !ftGreyReady || ftGreyTTF == nil {
		return nil
	}
	if s, ok := ftGreyStrikes[sz]; ok {
		if sticky {
			s.sticky = true
		}
		touchFTGreyStrike(sz)
		return s
	}
	if len(ftGreyStrikes) >= ftGreyStrikeMax {
		evictFTGreyStrikeIfNeeded()
		if len(ftGreyStrikes) >= ftGreyStrikeMax {
			return nil
		}
	}
	return bakeFTGreyStrike(sz, sticky, false, false)
}

func ensureFTGreyBoldStrike(sz int, sticky bool) *ftGreyStrike {
	if !PreferFTGreyStrikes() || !ftGreyReady || ftGreyBoldTTF == nil {
		return nil
	}
	if ftGreyBoldStrikes == nil {
		ftGreyBoldStrikes = make(map[int]*ftGreyStrike)
	}
	if s, ok := ftGreyBoldStrikes[sz]; ok {
		if sticky {
			s.sticky = true
		}
		return s
	}
	if len(ftGreyBoldStrikes) >= ftGreyStrikeMax {
		return nil
	}
	return bakeFTGreyStrike(sz, sticky, false, true)
}

func bakeFTGreyStrike(sz int, sticky, plain, bold bool) *ftGreyStrike {
	ttf := ftGreyTTF
	ttfData := ftGreyTTFData
	fontPath := ""
	if bold {
		// Size-aware Poppins weight (Medium/SemiBold); Segoe stays Bold via default path.
		boldPath := resolveFTGreyChromeBoldPathForSize(ftGreyRegularPath, sz)
		if boldPath == "" {
			return nil
		}
		if boldPath == ftGreyBoldPath && ftGreyBoldTTF != nil {
			ttf = ftGreyBoldTTF
			ttfData = ftGreyBoldTTFData
			fontPath = ftGreyBoldPath
		} else {
			bdata, err := ReadAssetFile(boldPath)
			if err != nil {
				if ftGreyBoldTTF == nil {
					return nil
				}
				ttf = ftGreyBoldTTF
				ttfData = ftGreyBoldTTFData
				fontPath = ftGreyBoldPath
			} else if bttf, err2 := truetype.Parse(bdata); err2 != nil {
				if ftGreyBoldTTF == nil {
					return nil
				}
				ttf = ftGreyBoldTTF
				ttfData = ftGreyBoldTTFData
				fontPath = ftGreyBoldPath
			} else {
				ttf = bttf
				ttfData = bdata
				fontPath = boldPath
			}
		}
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              float64(sz),
		DPI:               72,
		// Vertical = FreeType LIGHT / Skia kSlight — Full causes stem-width jumps
		// across sticky sizes (uneven chrome). DWrite path ignores this Face option.
		Hinting:           font.HintingVertical,
		GlyphCacheEntries: 8,
	})
	ascent := float32(face.Metrics().Ascent.Ceil())
	useLCD := !plain && PreferFTLCDChrome()
	fromDWrite := false
	fromOutline := false
	var (
		atlas  *image.RGBA
		glyphs map[rune]ftGreyGlyph
		ok     bool
	)
	var extraPhases []ftGreyPhase
	if useLCD {
		// Prefer real DirectWrite ClearType (T1.14b); fall back to fake neighbor LCD.
		// North star: 4-phase rem-body within 8-atlas WS budget.
		origins := ftGreySubXOrigins(sz)
		if atlas, glyphs, ok = packDWriteLCDStrikeAtlas(int32(sz), origins[0], fontPath, false); ok {
			fromDWrite = true
			for _, ox := range origins[1:] {
				a, g, okP := packDWriteLCDStrikeAtlas(int32(sz), ox, fontPath, false)
				if !okP || a == nil || g == nil {
					continue
				}
				enhanceS1DWriteLCDAtlas(a)
				texP := GoImageToTexture(a)
				if texP.ID == 0 {
					continue
				}
				rl.SetTextureFilter(texP, rl.FilterPoint)
				extraPhases = append(extraPhases, ftGreyPhase{origin: ox, tex: texP, glyphs: g})
			}
		} else {
			atlas, glyphs, ok = packFTLCDStrikeAtlas(face, int32(sz))
		}
	} else {
		var ink int
		// Linux: outline AA (fx/gg) — Windows DWrite analogue. No gamma/fx boost.
		if atlas, glyphs, ink, ok = packFTGreyAtlasOutline(ttfData, sz, 2); ok {
			fromOutline = true
			logFTGreyOutlineBake()
		} else {
			atlas, glyphs, ink, ok = packFTGreyAtlas(face, int32(sz))
		}
		_ = ink
	}
	_ = face.Close()
	if !ok {
		return nil
	}
	if fromDWrite {
		enhanceS1DWriteLCDAtlas(atlas)
	} else if useLCD {
		enhanceS1StrikeAtlas(atlas, true)
	}
	// Greyscale: leave masks alone (no SkMaskGamma / BoostRGBAAlpha).
	tex := GoImageToTexture(atlas)
	if tex.ID == 0 {
		return nil
	}
	// LCD/DWrite + outline-AA grey: point. Legacy FT bitmap grey: bilinear.
	if useLCD || fromDWrite || fromOutline {
		rl.SetTextureFilter(tex, rl.FilterPoint)
	} else {
		rl.SetTextureFilter(tex, rl.FilterBilinear)
	}
	ascent = ftGreyAscentFromGlyphs(glyphs, ascent)
	for i := range extraPhases {
		ascent = ftGreyAscentFromGlyphs(extraPhases[i].glyphs, ascent)
	}
	s := &ftGreyStrike{
		basePx: float32(sz),
		ascent: ascent,
		tex:    tex,
		glyphs: glyphs,
		phases: extraPhases,
		sticky: sticky,
		lcd:    useLCD,
	}
	if plain {
		if ftGreyPlainStrikes == nil {
			ftGreyPlainStrikes = make(map[int]*ftGreyStrike)
		}
		ftGreyPlainStrikes[sz] = s
		return s
	}
	if bold {
		if ftGreyBoldStrikes == nil {
			ftGreyBoldStrikes = make(map[int]*ftGreyStrike)
		}
		ftGreyBoldStrikes[sz] = s
		return s
	}
	ftGreyStrikes[sz] = s
	touchFTGreyStrike(sz)
	return s
}

// packFTLCDStrikeAtlas builds an exact-size LCD RGB atlas using ftGreyGlyph layout.
func packFTLCDStrikeAtlas(face font.Face, pixelSize int32) (*image.RGBA, map[rune]ftGreyGlyph, bool) {
	lcdAtlas, lcdGlyphs, _, ok := packFTLCDAtlas(face, pixelSize)
	if !ok {
		return nil, nil, false
	}
	applyFTLCDGammaContrast(lcdAtlas, 1.45, 0.18)
	out := make(map[rune]ftGreyGlyph, len(lcdGlyphs))
	for r, g := range lcdGlyphs {
		out[r] = ftGreyGlyph{
			src:      g.src,
			bearingX: g.bearingX,
			bearingY: g.bearingY,
			advance:  g.advance,
			hasInk:   g.hasInk,
		}
	}
	return lcdAtlas, out, true
}

func unloadFTGreyFont() {
	if ftGreyTex.ID != 0 {
		rl.UnloadTexture(ftGreyTex)
		ftGreyTex = rl.Texture2D{}
	}
	for _, s := range ftGreyStrikes {
		if s == nil {
			continue
		}
		if s.tex.ID != 0 {
			rl.UnloadTexture(s.tex)
		}
		for i := range s.phases {
			if s.phases[i].tex.ID != 0 {
				rl.UnloadTexture(s.phases[i].tex)
			}
		}
	}
	ftGreyStrikes = nil
	for _, s := range ftGreyBoldStrikes {
		if s == nil {
			continue
		}
		if s.tex.ID != 0 {
			rl.UnloadTexture(s.tex)
		}
		for i := range s.phases {
			if s.phases[i].tex.ID != 0 {
				rl.UnloadTexture(s.phases[i].tex)
			}
		}
	}
	ftGreyBoldStrikes = nil
	for _, s := range ftGreyPlainStrikes {
		if s != nil && s.tex.ID != 0 {
			rl.UnloadTexture(s.tex)
		}
	}
	ftGreyPlainStrikes = nil
	ftGreyStrikeLRU = nil
	ftGreyTTFData = nil
	ftGreyTTF = nil
	ftGreyBoldTTFData = nil
	ftGreyBoldTTF = nil
	ftGreyBoldPath = ""
	ftGreyRegularPath = ""
	ftGreyGlyphs = nil
	ftGreyReady = false
	ftGreyBasePx = 0
	ftGreyAscentPx = 0
	unloadFTGreyAuxFaces()
}

func ftGreyActive(mono, preview bool) bool {
	if !PreferFTGreyChrome() {
		return false
	}
	switch ftFaceFromFlags(mono, preview) {
	case ftFaceMono:
		return ftMonoFamily.ready
	case ftFacePreview:
		return ftPreviewFamily.ready
	default:
		return ftGreyReady
	}
}

// ftGreyPick returns glyphs for draw/measure. S1 looks up sticky strikes only —
// never bakes on the hot path (that caused resize/tab lag).
// Prefer exact snap; else nearest sticky within scale band (fixes rem-scaled
// segmented 14→~19 falling through to soft Path A).
func ftGreyPick(fontSize float32) (glyphs map[rune]ftGreyGlyph, tex rl.Texture2D, basePx, ascent float32, exactStrike, lcd bool) {
	if PreferFTGreyStrikes() && ftGreyStrikes != nil && len(ftGreyStrikes) > 0 {
		if s, ok := ftGreyPickNearestStrike(fontSize); ok {
			touchFTGreyStrike(int(s.basePx + 0.5))
			return s.glyphs, s.tex, s.basePx, s.ascent, true, s.lcd
		}
	}
	return ftGreyGlyphs, ftGreyTex, ftGreyBasePx, ftGreyAscentPx, false, false
}

// ftGreyResolveScale returns the draw/measure scale for a picked atlas.
// Sticky strikes (LCD and greyscale) always render 1:1 — stretching
// (16→16.8 FontDensity, 32→30) with FilterPoint looks wavy/jagged.
func ftGreyResolveScale(fontSize, basePx float32, exact, lcd bool) float32 {
	if !exact {
		return ftGreyScale(fontSize, basePx)
	}
	_ = lcd
	_ = fontSize
	_ = basePx
	return 1
}

// ftGreyPickNearestStrike finds sticky strike with scale closest to 1 within
// [ftGreyStrikeScaleMin, ftGreyStrikeScaleMax]. LCD prefers near-exact matches
// (off-by-2px stretch is the shaky band in Glyph Quality).
func ftGreyPickNearestStrike(fontSize float32) (*ftGreyStrike, bool) {
	return ftGreyPickNearestStrikeFiltered(fontSize, false)
}

func ftGreyPickNearestGreyStrike(fontSize float32) (*ftGreyStrike, bool) {
	if ftGreyPlainStrikes != nil {
		px := int(fontSize + 0.5)
		if s, ok := ftGreyPlainStrikes[px]; ok {
			return s, true
		}
	}
	return ftGreyPickNearestStrikeFiltered(fontSize, true)
}

func ftGreyPickNearestStrikeFiltered(fontSize float32, greyOnly bool) (*ftGreyStrike, bool) {
	if fontSize < 11 {
		fontSize = 11
	}
	var best *ftGreyStrike
	bestDist := float32(999)
	for sz, s := range ftGreyStrikes {
		if s == nil || s.basePx <= 0 {
			continue
		}
		if greyOnly && s.lcd {
			continue
		}
		scale := ftGreyStrikeScale(fontSize, float32(sz))
		if scale < ftGreyStrikeScaleMin || scale > ftGreyStrikeScaleMax {
			continue
		}
		if s.lcd && absFloat32(scale-1) > ftGreyLCDExactEps {
			continue
		}
		dist := absFloat32(scale - 1)
		if dist < bestDist {
			bestDist = dist
			best = s
		}
	}
	if best == nil {
		return nil, false
	}
	return best, true
}

// ftGreyPickForDraw selects glyphs for draw/measure; transparent hosts skip LCD atlases.
func ftGreyPickForDraw(fontSize float32) (glyphs map[rune]ftGreyGlyph, tex rl.Texture2D, basePx, ascent float32, exact, lcd bool) {
	return ftGreyPickForDrawWeight(fontSize, false)
}

func ftGreyPickForDrawWeight(fontSize float32, bold bool) (glyphs map[rune]ftGreyGlyph, tex rl.Texture2D, basePx, ascent float32, exact, lcd bool) {
	if bold && PreferFTGreyStrikes() && len(ftGreyBoldStrikes) > 0 {
		if s, ok := ftGreyPickNearestFrom(ftGreyBoldStrikes, fontSize, textSkipLCD > 0); ok {
			return s.glyphs, s.tex, s.basePx, s.ascent, true, s.lcd && textSkipLCD == 0
		}
	}
	if textSkipLCD > 0 {
		if PreferFTGreyStrikes() && ftGreyStrikes != nil {
			if s, ok := ftGreyPickNearestGreyStrike(fontSize); ok {
				touchFTGreyStrike(int(s.basePx + 0.5))
				return s.glyphs, s.tex, s.basePx, s.ascent, true, false
			}
		}
		return ftGreyGlyphs, ftGreyTex, ftGreyBasePx, ftGreyAscentPx, false, false
	}
	return ftGreyPick(fontSize)
}

// ftGreyPickNearestFrom finds the closest sticky strike in strikes within the scale band.
func ftGreyPickNearestFrom(strikes map[int]*ftGreyStrike, fontSize float32, greyOnly bool) (*ftGreyStrike, bool) {
	if strikes == nil || len(strikes) == 0 || fontSize < 1 {
		return nil, false
	}
	var best *ftGreyStrike
	bestDist := float32(1e9)
	for _, s := range strikes {
		if s == nil || s.basePx < 1 {
			continue
		}
		if greyOnly && s.lcd {
			continue
		}
		scale := ftGreyStrikeScale(fontSize, s.basePx)
		if scale < ftGreyStrikeScaleMin || scale > ftGreyStrikeScaleMax {
			continue
		}
		dist := absFloat32(scale - 1)
		if dist < bestDist {
			bestDist = dist
			best = s
		}
	}
	if best == nil {
		return nil, false
	}
	return best, true
}

func measureTextFTGrey(text string, fontSize float32) float32 {
	return measureTextFTGreyWeight(text, fontSize, false)
}

func measureTextFTGreyBold(text string, fontSize float32) float32 {
	return measureTextFTGreyWeight(text, fontSize, true)
}

func measureTextFTGreyWeight(text string, fontSize float32, bold bool) float32 {
	if !ftGreyReady {
		return 0
	}
	glyphs, _, basePx, _, exact, lcd := ftGreyPickForDrawWeight(fontSize, bold)
	if basePx <= 0 || glyphs == nil {
		return 0
	}
	scale := ftGreyResolveScale(fontSize, basePx, exact, lcd)
	if exact {
		if advs, ok := ftGreyKernAdvances(text, fontSize, bold, false, false); ok {
			var w float32
			for _, a := range advs {
				w += a * scale
			}
			return w
		}
	}
	var w float32
	n := 0
	for _, r := range text {
		g, ok := glyphs[r]
		if !ok {
			if sp, ok := glyphs[' ']; ok {
				w += sp.advance * scale
			} else if exact {
				w += basePx * 0.5
			} else {
				w += fontSize * 0.5 * ftGreyOpticalScale
			}
			continue
		}
		w += g.advance * scale
		n++
	}
	if FontSpacing != 0 && n > 1 {
		w += FontSpacing * float32(n-1)
	}
	return w
}

// snapPresentPx snaps sticky glyph coords to the layout pixel grid (layout==FB).
func snapPresentPx(v float32) float32 {
	return float32(math.Round(float64(v)))
}

func drawTextFTGreyBold(text string, x, y, fontSize float32, color rl.Color) {
	drawTextFTGreyWeight(text, x, y, fontSize, color, true)
}

func drawTextFTGrey(text string, x, y, fontSize float32, color rl.Color) {
	drawTextFTGreyWeight(text, x, y, fontSize, color, false)
}

func drawTextFTGreyWeight(text string, x, y, fontSize float32, color rl.Color, bold bool) {
	if !ftGreyReady {
		return
	}
	glyphs, tex, basePx, ascentPx, exact, lcd := ftGreyPickForDrawWeight(fontSize, bold)
	if basePx <= 0 || tex.ID == 0 || glyphs == nil {
		return
	}
	scale := ftGreyResolveScale(fontSize, basePx, exact, lcd)
	ascent := ascentPx * scale
	// Snap baseline + start pen once. Greyscale FilterPoint uses integer outline
	// advances/bearings and integer kern advances (FontSpacing folded in). Do NOT
	// snap pen+bearing or re-snap pen each glyph — that caused pair gaps
	// ("Close", "thin", "That") and discarded fractional tracking. No dstY snap.
	baseline := y + ascent
	penX := x
	if exact {
		baseline = snapPresentPx(baseline)
	}
	var strike *ftGreyStrike
	if exact && PreferFTGreyStrikes() {
		if bold {
			if s, ok := ftGreyPickNearestFrom(ftGreyBoldStrikes, fontSize, textSkipLCD > 0); ok {
				strike = s
			}
		} else if textSkipLCD > 0 {
			if s, ok := ftGreyPickNearestGreyStrike(fontSize); ok {
				strike = s
			}
		} else if s, ok := ftGreyPickNearestStrike(fontSize); ok {
			strike = s
		}
	}
	hasPhases := strike != nil && len(strike.phases) > 0 && textSkipLCD == 0
	if exact && (!lcd || !hasPhases) {
		penX = float32(math.Round(float64(penX)))
	}
	useLCD := lcd && exact && ftLCDShader.ID != 0
	var kernAdvs []float32
	kernOK := false
	if exact && !useLCD {
		kernAdvs, kernOK = ftGreyKernAdvances(text, fontSize, bold, false, false)
	}
	if useLCD {
		bg := lcdBgForText(color)
		rl.BeginShaderMode(ftLCDShader)
		if ftLCDBgLoc >= 0 {
			rl.SetShaderValue(ftLCDShader, ftLCDBgLoc, []float32{bg.X, bg.Y, bg.Z}, rl.ShaderUniformVec3)
		}
		bindFTLCDBlendParamsForSize(basePx, 1)
	}
	runeI := 0
	nRunes := 0
	for range text {
		nRunes++
	}
	for _, r := range text {
		gMap, gTex := glyphs, tex
		if hasPhases {
			ph := ftGreySubpixelPhase(penX)
			// phases holds extras beyond origin 0: index 0 → 0.25, 1 → 0.5, 2 → 0.75
			if ph > 0 && ph-1 < len(strike.phases) {
				gMap = strike.phases[ph-1].glyphs
				gTex = strike.phases[ph-1].tex
			}
		}
		g, ok := gMap[r]
		if !ok {
			// Advances from primary phase (identical Face metrics).
			if kernOK && runeI < len(kernAdvs) {
				penX += kernAdvs[runeI] * scale
			} else if sp, ok := glyphs[' ']; ok {
				penX += sp.advance * scale
			} else if exact {
				penX += basePx * 0.5
			} else {
				penX += fontSize * 0.5 * ftGreyOpticalScale
			}
			if !kernOK && FontSpacing != 0 && runeI < nRunes-1 {
				penX += float32(math.Round(float64(FontSpacing)))
			}
			runeI++
			continue
		}
		if g.hasInk && g.src.Width > 0 && g.src.Height > 0 {
			dstX := penX + g.bearingX*scale
			dstY := baseline + g.bearingY*scale
			dstW := g.src.Width * scale
			dstH := g.src.Height * scale
			dst := rl.NewRectangle(dstX, dstY, dstW, dstH)
			rl.DrawTexturePro(gTex, g.src, dst, rl.NewVector2(0, 0), 0, color)
		}
		if kernOK && runeI < len(kernAdvs) {
			penX += kernAdvs[runeI] * scale
		} else if pg, ok := glyphs[r]; ok {
			adv := pg.advance * scale
			if FontSpacing != 0 && runeI < nRunes-1 {
				adv = float32(math.Round(float64(adv + FontSpacing)))
			}
			penX += adv
		} else {
			adv := g.advance * scale
			if FontSpacing != 0 && runeI < nRunes-1 {
				adv = float32(math.Round(float64(adv + FontSpacing)))
			}
			penX += adv
		}
		runeI++
	}
	if useLCD {
		rl.EndShaderMode()
	}
}

func ftGreyInkHeight(fontSize float32) float32 {
	return ftGreyInkHeightKind(fontSize, ftFaceChrome, false)
}

func ftGreyInkHeightKind(fontSize float32, kind ftFaceKind, bold bool) float32 {
	if kind != ftFaceChrome {
		fam := ftAuxFamily(kind)
		_, _, basePx, ascentPx, exact, lcd := ftGreyAuxPick(fam, fontSize, bold)
		if basePx <= 0 {
			return fontSize * ftGreyOpticalScale
		}
		scale := ftGreyResolveScale(fontSize, basePx, exact, lcd)
		h := ascentPx * scale
		if h <= 0 {
			if exact {
				return basePx
			}
			return fontSize * ftGreyOpticalScale
		}
		return h
	}
	if !ftGreyReady {
		return fontSize
	}
	_, _, basePx, ascentPx, exact, lcd := ftGreyPickForDrawWeight(fontSize, bold)
	if basePx <= 0 {
		return fontSize * ftGreyOpticalScale
	}
	scale := ftGreyResolveScale(fontSize, basePx, exact, lcd)
	h := ascentPx * scale
	if h <= 0 {
		if exact {
			return basePx
		}
		return fontSize * ftGreyOpticalScale
	}
	return h
}

// ftGreyLineBoxHeightKind is ascent+descent for Label AutoHeight (avoids scissor clip).
func ftGreyLineBoxHeightKind(fontSize float32, kind ftFaceKind, bold bool) float32 {
	if kind == ftFaceChrome {
		if !ftGreyReady {
			return fontSize
		}
		glyphs, _, basePx, ascentPx, exact, lcd := ftGreyPickForDrawWeight(fontSize, bold)
		if basePx <= 0 {
			return fontSize
		}
		scale := ftGreyResolveScale(fontSize, basePx, exact, lcd)
		ascent := ascentPx * scale
		descent := ftGreyDescentFromGlyphs(glyphs) * scale
		h := ascent + descent
		if h < fontSize {
			return fontSize
		}
		return h
	}
	fam := ftAuxFamily(kind)
	glyphs, _, basePx, ascentPx, exact, lcd := ftGreyAuxPick(fam, fontSize, bold)
	if basePx <= 0 || glyphs == nil {
		return fontSize * ftGreyOpticalScale
	}
	scale := ftGreyResolveScale(fontSize, basePx, exact, lcd)
	ascent := ascentPx * scale
	descent := ftGreyDescentFromGlyphs(glyphs) * scale
	h := ascent + descent + 2 // 1px slack top+bottom for AA fringe / FilterPoint
	if h < fontSize {
		return fontSize + 2
	}
	return h
}

// DrawTextAtDevicePx draws Latin chrome at exact framebuffer pixels (no rem/DPI).
// For Glyph Quality dogfood — compare strike sizes 1:1.
func DrawTextAtDevicePx(text string, x, y, px float32, color rl.Color) {
	if px < minRenderPx {
		px = minRenderPx
	}
	drawTextF(text, x, y, px, color, false, false, false, false)
}

// MeasureTextAtDevicePx is the measure twin of [DrawTextAtDevicePx].
func MeasureTextAtDevicePx(text string, px float32) float32 {
	if px < minRenderPx {
		px = minRenderPx
	}
	return measureTextF(text, px, false, false, false, false)
}

// LatinStrikeProbe reports whether fontSize (device px) hits a sticky strike.
func LatinStrikeProbe(fontSize float32) (hit bool, basePx float32, lcd bool, phases int) {
	if !ftGreyReady || !PreferFTGreyStrikes() {
		return false, 0, false, 0
	}
	s, ok := ftGreyPickNearestStrike(fontSize)
	if !ok || s == nil {
		return false, 0, false, 0
	}
	scale := ftGreyStrikeScale(fontSize, s.basePx)
	if s.lcd && absFloat32(scale-1) > ftGreyLCDExactEps {
		return false, s.basePx, s.lcd, 1 + len(s.phases)
	}
	if scale < ftGreyStrikeScaleMin || scale > ftGreyStrikeScaleMax {
		return false, s.basePx, s.lcd, 1 + len(s.phases)
	}
	return true, s.basePx, s.lcd, 1 + len(s.phases)
}

// LatinStickyStrikeSizes returns the current DPI-scaled sticky ladder (dogfood / probes).
func LatinStickyStrikeSizes() []int {
	return stickyStrikeSizes()
}

// LatinCSSToDevicePx maps browser CSS px → client px for Glyph Quality compare
// rows (css × OS DPR, quantized). Product chrome uses rem×token client px only —
// not this helper (×DPI there inflated widgets).
func LatinCSSToDevicePx(css float32) float32 {
	if css < minRenderPx {
		css = minRenderPx
	}
	return QuantizeFTGreyChromeSize(css * EffectiveGlyphDisplayScale())
}

// LatinDeviceToCSSPx estimates browser CSS px for a client size (÷ DisplayScale).
func LatinDeviceToCSSPx(device float32) float32 {
	d := EffectiveGlyphDisplayScale()
	if d < 1.02 {
		return device
	}
	return device / d
}

// GlyphQualityHUD is a one-line status for the Glyph Quality dogfood scene.
func GlyphQualityHUD() string {
	ladder := stickyStrikeSizes()
	return fmt.Sprintf("backend=%s  dpi=%.2f  ssaa=%.2f  rem=%.0f  sticky=%v  quality=%s",
		UIFontBackend(), EffectiveGlyphDisplayScale(), EffectiveSupersamplingScale(),
		RootFontSize, ladder, RenderQualityString())
}
