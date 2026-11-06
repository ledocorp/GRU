package ui

import (
	"fmt"
	"image"
	"math"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/image/font"
)

// S1 aux faces — Fira Code (Mono) and Inter (PreviewFont) get the same sticky
// ClearType path as chrome Segoe, with independent strike budgets (max 8 each).

type ftFaceKind uint8

const (
	ftFaceChrome ftFaceKind = iota
	ftFaceMono
	ftFacePreview
)

type ftGreyAuxFamily struct {
	ready       bool
	path        string
	ttf         *truetype.Font
	ttfData     []byte
	boldPath    string
	boldTTF     *truetype.Font
	boldData    []byte
	extraTTF    map[string]*truetype.Font
	extraData   map[string][]byte
	tex         rl.Texture2D
	basePx      float32
	ascentPx    float32
	glyphs      map[rune]ftGreyGlyph
	strikes     map[int]*ftGreyStrike
	boldStrikes map[int]*ftGreyStrike
	lru         []int
}

var (
	ftMonoFamily    ftGreyAuxFamily
	ftPreviewFamily ftGreyAuxFamily
	ftAuxInitOnce   sync.Once
)

// Aux sticky sizes mirror chrome when strikes are on; fallback ladder if unset.
var ftGreyAuxStickySizes = []int{14, 16, 18, 20, 24, 28, 32, 40}

func QuantizeFTGreyAuxSize(fs float32) float32 {
	// Same snap policy as chrome — Fira/Inter must hit exact sticky keys for 1:1 LCD.
	return QuantizeFTGreyChromeSize(fs)
}

func ftFaceFromFlags(mono, preview bool) ftFaceKind {
	switch {
	case mono:
		return ftFaceMono
	case preview:
		return ftFacePreview
	default:
		return ftFaceChrome
	}
}

func ftAuxFamily(kind ftFaceKind) *ftGreyAuxFamily {
	switch kind {
	case ftFaceMono:
		return &ftMonoFamily
	case ftFacePreview:
		return &ftPreviewFamily
	default:
		return nil
	}
}

func resolveFTGreyAuxPath(kind ftFaceKind) string {
	switch kind {
	case ftFaceMono:
		return firstExisting(monoFontCandidates...)
	case ftFacePreview:
		for _, set := range interFaceCandidates() {
			if p := firstExisting(set.regular); p != "" {
				return p
			}
		}
		// Tree ships variable Inter only (no static optical cuts).
		return firstExisting(interVariableCandidates()...)
	default:
		return ""
	}
}

func resolveFTGreyAuxBoldPath(kind ftFaceKind) string {
	if kind != ftFacePreview {
		return ""
	}
	// Prefer SemiBold for HUD / default bold strike — Bold fills counters at UI sizes.
	for _, pt := range []int{24, 28, 18} {
		for _, cand := range interSemiBoldCandidates(pt) {
			if p := firstExisting(cand); p != "" {
				return p
			}
		}
	}
	for _, set := range interFaceCandidates() {
		if p := firstExisting(set.bold); p != "" {
			return p
		}
	}
	// Variable Inter: same file for bold (weight axis); better than (none).
	return firstExisting(interVariableCandidates()...)
}

// InitFTGreyAuxFaces loads Fira (Mono) + Inter (PreviewFont) onto the same S1
// sticky ClearType path as chrome Segoe — separate strike maps, same bake/draw.
func InitFTGreyAuxFaces() {
	if !PreferFTGreyChrome() || !ftGreyReady {
		return
	}
	initOneFTGreyAux(&ftMonoFamily, ftFaceMono, "mono/Fira")
	initOneFTGreyAux(&ftPreviewFamily, ftFacePreview, "preview/Inter")
	ftAuxInitOnce.Do(func() {
		aux := "ft-grey"
		switch {
		case PreferAuxFaceLCD():
			aux = "lcd"
		case PreferAuxFaceDWriteGrey():
			aux = "dwrite-grey"
		}
		gruLog("Gru: S1 faces — chrome=%s mono=%s preview=%s bold=%s aux=%s",
			baseName(LoadedUIFontPath),
			baseName(ftMonoFamily.path),
			baseName(ftPreviewFamily.path),
			baseName(ftPreviewFamily.boldPath),
			aux,
		)
	})
}

func baseName(path string) string {
	if path == "" {
		return "(none)"
	}
	path = strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// S1FaceHUD reports which TTFs back chrome / mono / preview for dogfood demos.
func S1FaceHUD() string {
	mode := "ft-grey"
	switch {
	case PreferAuxFaceLCD():
		mode = "lcd"
	case PreferAuxFaceDWriteGrey():
		mode = "dwrite-grey"
	}
	return fmt.Sprintf("chrome=%s · mono=%s · preview=%s · previewBold=%s · aux=%s",
		baseName(LoadedUIFontPath),
		baseName(ftMonoFamily.path),
		baseName(ftPreviewFamily.path),
		baseName(ftPreviewFamily.boldPath),
		mode,
	)
}

func initOneFTGreyAux(fam *ftGreyAuxFamily, kind ftFaceKind, label string) {
	unloadFTGreyAux(fam)
	path := resolveFTGreyAuxPath(kind)
	if path == "" {
		return
	}
	data, err := ReadAssetFile(path)
	if err != nil {
		return
	}
	ttf, err := truetype.Parse(data)
	if err != nil {
		return
	}
	pixelSize := int32(48)
	if PreferFTGreyStrikes() {
		pixelSize = 24
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              float64(pixelSize),
		DPI:               72,
		Hinting:           font.HintingNone,
		GlyphCacheEntries: 8,
	})
	ascent := float32(face.Metrics().Ascent.Ceil())
	atlas, glyphs, ink, ok := packFTGreyAtlas(face, pixelSize)
	_ = face.Close()
	if !ok {
		gruLog("Gru: S1 aux %s atlas failed (ink=%d)", label, ink)
		return
	}
	tex := GoImageToTexture(atlas)
	if tex.ID == 0 {
		return
	}
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	fam.path = path
	fam.ttfData = data
	fam.ttf = ttf
	fam.glyphs = glyphs
	fam.tex = tex
	fam.basePx = float32(pixelSize)
	fam.ascentPx = ascent
	fam.ready = true
	if boldPath := resolveFTGreyAuxBoldPath(kind); boldPath != "" {
		if bdata, err := ReadAssetFile(boldPath); err == nil {
			if bttf, err := truetype.Parse(bdata); err == nil {
				fam.boldPath = boldPath
				fam.boldData = bdata
				fam.boldTTF = bttf
			}
		}
	}
	if PreferFTGreyStrikes() {
		fam.strikes = make(map[int]*ftGreyStrike)
		fam.boldStrikes = make(map[int]*ftGreyStrike)
		fam.lru = fam.lru[:0]
		// Full sticky LCD ladder (Font Faces 16–32, Notepad mono, markdown).
		// Cutting this to {16,20,24,28} made 18/32 miss → soft Path A greyscale
		// (2026-08-21e quality regress). WS stays bounded by 1024²-first bake
		// (2048 only on truncate) — not by starving sticky coverage.
		for _, sz := range ftGreyAuxStickySizes {
			if len(fam.strikes) >= ftGreyStrikeMax {
				break
			}
			_ = bakeFTGreyAuxStrike(fam, sz, true, false)
		}
		// Preview Inter always has a bold cut via interFaceForStickySize; Fira only
		// when boldTTF loaded. Prewarm the same ladder so bold rows stay LCD.
		if fam.boldTTF != nil || fam == &ftPreviewFamily {
			for _, sz := range ftGreyAuxStickySizes {
				if len(fam.boldStrikes) >= ftGreyStrikeMax {
					break
				}
				_ = bakeFTGreyAuxStrike(fam, sz, true, true)
			}
		}
	}
}

func unloadFTGreyAux(fam *ftGreyAuxFamily) {
	if fam == nil {
		return
	}
	if fam.tex.ID != 0 {
		rl.UnloadTexture(fam.tex)
	}
	for _, s := range fam.strikes {
		unloadAuxStrike(s)
	}
	for _, s := range fam.boldStrikes {
		unloadAuxStrike(s)
	}
	*fam = ftGreyAuxFamily{}
}

func unloadAuxStrike(s *ftGreyStrike) {
	if s == nil {
		return
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

// auxTTFData returns the TTF bytes for outline sticky bake (path may be an Inter cut).
func (fam *ftGreyAuxFamily) auxTTFData(path string, bold bool) []byte {
	if fam == nil {
		return nil
	}
	if path != "" && fam.extraData != nil {
		if d := fam.extraData[path]; len(d) > 0 {
			return d
		}
	}
	if bold && len(fam.boldData) > 0 {
		return fam.boldData
	}
	return fam.ttfData
}

func (fam *ftGreyAuxFamily) ttfAt(path string) (*truetype.Font, string) {
	if fam == nil || path == "" {
		return nil, ""
	}
	if path == fam.path && fam.ttf != nil {
		return fam.ttf, fam.path
	}
	if path == fam.boldPath && fam.boldTTF != nil {
		return fam.boldTTF, fam.boldPath
	}
	if fam.extraTTF != nil {
		if f, ok := fam.extraTTF[path]; ok && f != nil {
			return f, path
		}
	}
	data, err := ReadAssetFile(path)
	if err != nil {
		return fam.ttf, fam.path
	}
	ttf, err := truetype.Parse(data)
	if err != nil {
		return fam.ttf, fam.path
	}
	if fam.extraTTF == nil {
		fam.extraTTF = make(map[string]*truetype.Font)
		fam.extraData = make(map[string][]byte)
	}
	fam.extraTTF[path] = ttf
	fam.extraData[path] = data
	return ttf, path
}

func bakeFTGreyAuxStrike(fam *ftGreyAuxFamily, sz int, sticky, bold bool) *ftGreyStrike {
	if fam == nil || !fam.ready {
		return nil
	}
	ttf := fam.ttf
	fontPath := fam.path
	dest := fam.strikes
	if fam == &ftPreviewFamily {
		fontPath = interFaceForStickySize(sz, bold)
		ttf, fontPath = fam.ttfAt(fontPath)
		if bold {
			if fam.boldStrikes == nil {
				fam.boldStrikes = make(map[int]*ftGreyStrike)
			}
			dest = fam.boldStrikes
		}
	} else if bold {
		if fam.boldTTF == nil {
			return nil
		}
		ttf = fam.boldTTF
		fontPath = fam.boldPath
		if fam.boldStrikes == nil {
			fam.boldStrikes = make(map[int]*ftGreyStrike)
		}
		dest = fam.boldStrikes
	}
	if ttf == nil {
		return nil
	}
	if s, ok := dest[sz]; ok {
		s.sticky = s.sticky || sticky
		return s
	}
	if len(dest) >= ftGreyStrikeMax {
		return nil
	}
	// Bundled Inter/Fira: ClearType+LCD by default (same recipe as Segoe chrome).
	// Greyscale only when GRU_AUX_LCD=0.
	useLCD := PreferFTLCDChrome() && PreferAuxFaceLCD()
	useDWriteGrey := !useLCD && PreferAuxFaceDWriteGrey()
	hint := font.HintingNone
	if useLCD {
		hint = font.HintingVertical
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              float64(sz),
		DPI:               72,
		Hinting:           hint,
		GlyphCacheEntries: 8,
	})
	ascent := float32(face.Metrics().Ascent.Ceil())
	fromDWrite := false
	fromOutline := false
	var (
		atlas  *image.RGBA
		glyphs map[rune]ftGreyGlyph
		ok     bool
	)
	if useLCD {
		// Soft aux: NATURAL_SYMMETRIC at all sticky sizes (matches the strong 24–32
		// look). NATURAL-only ≤20 was why small Font Faces lagged the big sizes.
		// Chrome Segoe stays NATURAL (symmetric=false in pack chrome path).
		sym := PreferAuxFaceLCDSoft()
		if atlas, glyphs, ok = packDWriteLCDStrikeAtlas(int32(sz), 0, fontPath, sym); ok {
			fromDWrite = true
		} else {
			atlas, glyphs, ok = packFTLCDStrikeAtlas(face, int32(sz))
		}
	} else if useDWriteGrey {
		if atlas, glyphs, ok = packDWriteGreyStrikeAtlas(int32(sz), fontPath); ok {
			fromDWrite = true
		} else {
			var ink int
			data := fam.auxTTFData(fontPath, bold)
			vars := outlineVarsForAuxPath(fontPath, bold, sz)
			if atlas, glyphs, ink, ok = packFTGreyAtlasOutlineVars(data, sz, 4, vars); ok {
				fromOutline = true
				logFTGreyOutlineBake()
				logInterVFBoldOnce(fontPath, bold, sz, vars)
			} else {
				atlas, glyphs, ink, ok = packFTGreyAuxAtlas(face, int32(sz))
			}
			_ = ink
		}
	} else {
		var ink int
		data := fam.auxTTFData(fontPath, bold)
		vars := outlineVarsForAuxPath(fontPath, bold, sz)
		if atlas, glyphs, ink, ok = packFTGreyAtlasOutlineVars(data, sz, 4, vars); ok {
			fromOutline = true
			logFTGreyOutlineBake()
			logInterVFBoldOnce(fontPath, bold, sz, vars)
		} else {
			atlas, glyphs, ink, ok = packFTGreyAuxAtlas(face, int32(sz))
		}
		_ = ink
	}
	_ = face.Close()
	if !ok {
		return nil
	}
	if fromDWrite && useLCD {
		enhanceS1DWriteLCDAtlas(atlas)
	} else if useLCD {
		enhanceS1StrikeAtlas(atlas, true)
	}
	// Greyscale: leave masks alone (no gamma/fx boost).
	tex := GoImageToTexture(atlas)
	if tex.ID == 0 {
		return nil
	}
	// LCD / outline-AA: point. Legacy FT bitmap grey: bilinear.
	if useLCD || fromOutline {
		rl.SetTextureFilter(tex, rl.FilterPoint)
	} else {
		rl.SetTextureFilter(tex, rl.FilterBilinear)
	}
	ascent = ftGreyAscentFromGlyphs(glyphs, ascent)
	s := &ftGreyStrike{
		basePx: float32(sz),
		ascent: ascent,
		tex:    tex,
		glyphs: glyphs,
		sticky: sticky,
		lcd:    useLCD,
	}
	dest[sz] = s
	fam.lru = append(fam.lru, sz)
	return s
}

func ftGreyAuxPick(fam *ftGreyAuxFamily, fontSize float32, bold bool) (glyphs map[rune]ftGreyGlyph, tex rl.Texture2D, basePx, ascent float32, exact, lcd bool) {
	if fam == nil || !fam.ready {
		return nil, rl.Texture2D{}, 0, 0, false, false
	}
	strikes := fam.strikes
	if bold && fam.boldStrikes != nil && len(fam.boldStrikes) > 0 {
		strikes = fam.boldStrikes
	}
	if PreferFTGreyStrikes() && strikes != nil && len(strikes) > 0 {
		var best *ftGreyStrike
		bestDist := float32(1e9)
		for _, s := range strikes {
			if s == nil || s.basePx <= 0 {
				continue
			}
			scale := fontSize / s.basePx
			if scale < ftGreyStrikeScaleMin || scale > ftGreyStrikeScaleMax {
				continue
			}
			if s.lcd {
				d := absFloat32(fontSize - s.basePx)
				if d > s.basePx*ftGreyLCDExactEps && d > 0.51 {
					continue
				}
			}
			d := absFloat32(fontSize - s.basePx)
			if d < bestDist {
				bestDist = d
				best = s
			}
		}
		if best != nil {
			return best.glyphs, best.tex, best.basePx, best.ascent, true, best.lcd
		}
	}
	return fam.glyphs, fam.tex, fam.basePx, fam.ascentPx, false, false
}

func unloadFTGreyAuxFaces() {
	unloadFTGreyAux(&ftMonoFamily)
	unloadFTGreyAux(&ftPreviewFamily)
}

func measureTextFTGreyKind(text string, fontSize float32, kind ftFaceKind, bold bool) float32 {
	if kind == ftFaceChrome {
		if bold {
			return measureTextFTGreyBold(text, fontSize)
		}
		return measureTextFTGrey(text, fontSize)
	}
	fam := ftAuxFamily(kind)
	glyphs, _, basePx, _, exact, lcd := ftGreyAuxPick(fam, fontSize, bold)
	if basePx <= 0 || glyphs == nil {
		return 0
	}
	scale := ftGreyResolveScale(fontSize, basePx, exact, lcd)
	if exact {
		mono, preview := kind == ftFaceMono, kind == ftFacePreview
		if advs, ok := ftGreyKernAdvances(text, fontSize, bold, mono, preview); ok {
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

func drawTextFTGreyKind(text string, x, y, fontSize float32, color rl.Color, kind ftFaceKind, bold bool) {
	if kind == ftFaceChrome {
		if bold {
			drawTextFTGreyBold(text, x, y, fontSize, color)
			return
		}
		drawTextFTGrey(text, x, y, fontSize, color)
		return
	}
	fam := ftAuxFamily(kind)
	glyphs, tex, basePx, ascentPx, exact, lcd := ftGreyAuxPick(fam, fontSize, bold)
	if basePx <= 0 || tex.ID == 0 || glyphs == nil {
		return
	}
	scale := ftGreyResolveScale(fontSize, basePx, exact, lcd)
	ascent := ascentPx * scale
	baseline := y + ascent
	penX := x
	useLCD := lcd && exact && ftLCDShader.ID != 0
	if exact {
		baseline = snapPresentPx(baseline)
		penX = snapPresentPx(penX)
	}
	mono, preview := kind == ftFaceMono, kind == ftFacePreview
	var kernAdvs []float32
	kernOK := false
	if exact && !useLCD {
		kernAdvs, kernOK = ftGreyKernAdvances(text, fontSize, bold, mono, preview)
	}
	if useLCD {
		bg := lcdBgForText(color)
		rl.BeginShaderMode(ftLCDShader)
		if ftLCDBgLoc >= 0 {
			rl.SetShaderValue(ftLCDShader, ftLCDBgLoc, []float32{bg.X, bg.Y, bg.Z}, rl.ShaderUniformVec3)
		}
		bindFTLCDBlendParamsForSizeOpt(basePx, 1, PreferAuxFaceLCDSoft(), false, false, false)
	}
	runeI := 0
	nRunes := 0
	for range text {
		nRunes++
	}
	for _, r := range text {
		g, ok := glyphs[r]
		if !ok {
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
			rl.DrawTexturePro(tex, g.src, dst, rl.NewVector2(0, 0), 0, color)
		}
		if kernOK && runeI < len(kernAdvs) {
			penX += kernAdvs[runeI] * scale
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
