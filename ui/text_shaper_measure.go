package ui

import (
	"bytes"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

// shapedFaceStack holds go-text/typesetting faces mirroring the SDF uiFontStack.
type shapedFaceStack struct {
	regular    *font.Face
	bold       *font.Face
	italic     *font.Face
	boldItalic *font.Face
	mono       *font.Face
	ready      bool
	monoReady  bool
}

var (
	shapedUI      shapedFaceStack
	shapedPreview shapedFaceStack
	hbShaper      shaping.HarfbuzzShaper

	// shapedMeasureRatio maps rounded pixel size → SDF/shaped width scale (T1.1 parity).
	shapedMeasureRatio             [96]float32
	shapedMeasureRatioReady        bool
	shapedPreviewMeasureRatio      [96]float32
	shapedPreviewMeasureRatioReady bool
)

var legacyUIFontBoldCandidates = []string{
	"C:/Windows/Fonts/segoeuib.ttf",
	"assets/fonts/Roboto-Bold.ttf",
}

var legacyUIFontItalicCandidates = []string{
	"C:/Windows/Fonts/segoeuii.ttf",
	"assets/fonts/Roboto-Italic.ttf",
}

var legacyUIFontBoldItalicCandidates = []string{
	"C:/Windows/Fonts/segoeuiz.ttf",
	"assets/fonts/Roboto-BoldItalic.ttf",
}

// InitShapedFonts loads TTF faces for shaped measure/draw (T1).
// Safe without raylib; call after InitSDFFont or standalone in tests.
func InitShapedFonts() bool {
	if shapedUI.ready {
		calibrateShapedMeasureRatio()
		return true
	}
	if useBundledUIFontStack {
		return initShapedBundledStack()
	}
	return initShapedLegacyStack()
}

func initShapedLegacyStack() bool {
	regPath := firstExisting(fontCandidates...)
	if regPath == "" {
		return false
	}
	reg, err := loadShapedFace(regPath)
	if err != nil {
		return false
	}
	shapedUI.regular = reg
	if f, err := loadShapedFace(firstExisting(legacyUIFontBoldCandidates...)); err == nil {
		shapedUI.bold = f
	}
	if f, err := loadShapedFace(firstExisting(legacyUIFontItalicCandidates...)); err == nil {
		shapedUI.italic = f
	}
	if f, err := loadShapedFace(firstExisting(legacyUIFontBoldItalicCandidates...)); err == nil {
		shapedUI.boldItalic = f
	}
	shapedUI.ready = true
	initShapedGlyphCache()
	EnsureComplexScriptUIFont()
	initShapedPreviewFaces()
	calibrateShapedMeasureRatio()
	return true
}

func initShapedBundledStack() bool {
	for _, set := range poppinsFaceCandidates() {
		regPath := firstExisting(set.regular)
		if regPath == "" {
			continue
		}
		reg, err := loadShapedFace(regPath)
		if err != nil {
			continue
		}
		shapedUI.regular = reg
		if f, err := loadShapedFace(firstExisting(set.bold)); err == nil {
			shapedUI.bold = f
		}
		if f, err := loadShapedFace(firstExisting(set.italic)); err == nil {
			shapedUI.italic = f
		}
		if f, err := loadShapedFace(firstExisting(set.boldItalic)); err == nil {
			shapedUI.boldItalic = f
		}
		shapedUI.ready = true
		break
	}
	if !shapedUI.ready {
		return false
	}
	initShapedGlyphCache()
	EnsureComplexScriptUIFont()
	shapedMeasureCacheClear()
	calibrateShapedMeasureRatio()
	if f, err := loadShapedFace(firstExisting(monoFontCandidates...)); err == nil {
		shapedUI.mono = f
		shapedUI.monoReady = true
	}
	initShapedPreviewFaces()
	return true
}

func shapedEnsureMono() {
	if shapedUI.monoReady || !shapedUI.ready {
		return
	}
	if f, err := loadShapedFace(firstExisting(monoFontCandidates...)); err == nil {
		shapedUI.mono = f
		shapedUI.monoReady = true
	}
}

// initShapedPreviewFaces loads Inter faces for the markdown preview lane (T1.4).
func initShapedPreviewFaces() bool {
	if shapedPreview.ready {
		return true
	}
	for _, set := range interFaceCandidates() {
		regPath := firstExisting(set.regular)
		if regPath == "" {
			continue
		}
		reg, err := loadShapedFace(regPath)
		if err != nil {
			continue
		}
		shapedPreview.regular = reg
		if f, err := loadShapedFace(firstExisting(set.bold)); err == nil {
			shapedPreview.bold = f
		}
		if f, err := loadShapedFace(firstExisting(set.italic)); err == nil {
			shapedPreview.italic = f
		}
		if f, err := loadShapedFace(firstExisting(set.boldItalic)); err == nil {
			shapedPreview.boldItalic = f
		}
		shapedPreview.ready = true
		return true
	}
	return false
}

func shapedEnsurePreview() {
	if shapedPreview.ready || !shapedUI.ready {
		return
	}
	initShapedPreviewFaces()
}

func loadShapedFace(path string) (*font.Face, error) {
	if path == "" {
		return nil, os.ErrNotExist
	}
	data, err := readUIFontFile(path)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".ttc") || strings.HasSuffix(lower, ".otc") {
		faces, err := font.ParseTTC(r)
		if err != nil {
			return nil, err
		}
		if len(faces) == 0 {
			return nil, os.ErrNotExist
		}
		return faces[0], nil
	}
	return font.ParseTTF(r)
}

// pickShapedFace mirrors pickUIFont for measure — legacy single-face stack ignores
// bold/italic flags (SDF bold shader is draw-only).
func pickShapedFace(bold, italic, mono, preview bool) *font.Face {
	if mono {
		shapedEnsureMono()
		if shapedUI.monoReady && shapedUI.mono != nil {
			return shapedUI.mono
		}
	}
	if preview {
		shapedEnsurePreview()
		if shapedPreview.ready {
			return pickShapedFaceStack(shapedPreview, bold, italic)
		}
	}
	if fonts.ready && shapedUI.ready {
		return pickShapedFaceStack(shapedUI, bold, italic)
	}
	if shapedUI.regular != nil {
		return shapedUI.regular
	}
	return nil
}

func pickShapedFaceStack(stack shapedFaceStack, bold, italic bool) *font.Face {
	switch {
	case bold && italic && stack.boldItalic != nil:
		return stack.boldItalic
	case bold && stack.bold != nil:
		return stack.bold
	case italic && stack.italic != nil:
		return stack.italic
	default:
		return stack.regular
	}
}

// shapedSizePx encodes a pixel font size for shaping.Input.Size (26.6 fixed point).
func shapedSizePx(fontSize float32) fixed.Int26_6 {
	return fixed.I(int(fontSize + 0.5))
}

// shapedLetterSpacingPx encodes raylib FontSpacing for AddLetterSpacing.
func shapedLetterSpacingPx(spacing float32) fixed.Int26_6 {
	return fixed.Int26_6(spacing * 64)
}

func shapedScriptFor(text []rune) language.Script {
	if len(text) == 0 {
		return language.Latin
	}
	return language.LookupScript(text[0])
}

func shapedMeasureRatioFor(fontSize float32, preview bool, text string) float32 {
	if shapedTextUsesComplexScript(text) {
		return 1
	}
	if preview {
		shapedEnsurePreviewCalib()
		if shapedPreviewMeasureRatioReady {
			idx := int(fontSize + 0.5)
			if idx >= 0 && idx < len(shapedPreviewMeasureRatio) && shapedPreviewMeasureRatio[idx] > 0 {
				return shapedPreviewMeasureRatio[idx]
			}
		}
	}
	if !shapedMeasureRatioReady {
		return 1
	}
	idx := int(fontSize + 0.5)
	if idx >= 0 && idx < len(shapedMeasureRatio) && shapedMeasureRatio[idx] > 0 {
		return shapedMeasureRatio[idx]
	}
	return 1
}

// shapedEnsurePreviewCalib loads preview SDF + shaped faces and calibrates Inter lane (T1.4).
func shapedEnsurePreviewCalib() {
	shapedEnsurePreview()
	if shapedPreviewMeasureRatioReady || !shapedPreview.ready {
		return
	}
	if sdfReady && SDFFont.BaseSize > 0 {
		EnsurePreviewUIFont(SDFFont.BaseSize)
	}
	calibrateShapedPreviewMeasureRatio()
}

// calibrateShapedPreviewMeasureRatio aligns shaped preview widths to Inter SDF atlas.
func calibrateShapedPreviewMeasureRatio() {
	if !previewFonts.ready || !shapedPreview.ready || previewFonts.regular.BaseSize <= 0 {
		return
	}
	const probe = "Hello world"
	sizes := []int{11, 12, 14, 15, 16, 17, 18, 19, 22, 26, 28, 32}
	font := previewFonts.regular
	for _, sz := range sizes {
		fs := float32(sz)
		sdfW := rl.MeasureTextEx(font, probe, fs, FontSpacing).X
		raw, ok := shapedMeasureTextFRaw(probe, fs, false, false, false, true)
		if ok && raw > 0 {
			shapedPreviewMeasureRatio[sz] = sdfW / raw
		}
	}
	shapedPreviewMeasureRatioReady = true
}

// calibrateShapedMeasureRatio aligns shaped widths to the active SDF atlas (T1.1).
func calibrateShapedMeasureRatio() {
	if !sdfReady || !shapedUI.ready || SDFFont.BaseSize <= 0 {
		return
	}
	const probe = "Hello world"
	sizes := []int{11, 12, 14, 15, 16, 17, 18, 19, 22, 26, 28, 32}
	for _, sz := range sizes {
		fs := float32(sz)
		sdfW := rl.MeasureTextEx(SDFFont, probe, fs, FontSpacing).X
		raw, ok := shapedMeasureTextFRaw(probe, fs, false, false, false, false)
		if ok && raw > 0 {
			shapedMeasureRatio[sz] = sdfW / raw
		}
	}
	shapedMeasureRatioReady = true
}

func unloadShapedFonts() {
	shapedUI = shapedFaceStack{}
	shapedPreview = shapedFaceStack{}
	shapedDevanagari = nil
	shapedMeasureRatioReady = false
	shapedPreviewMeasureRatioReady = false
	shapedMeasureCacheClear()
}

// shapedTextRun holds a shaped line used by measure and draw (T1.1+).
type shapedTextRun struct {
	out   shaping.Output
	runes []rune
}

func shapedShapeText(text string, fontSize float32, bold, italic, mono, preview bool) (shapedTextRun, bool) {
	if text == "" {
		return shapedTextRun{}, true
	}
	face := pickShapedFaceForText(text, bold, italic, mono, preview)
	if face == nil {
		return shapedTextRun{}, false
	}
	runes := []rune(text)
	out := hbShaper.Shape(shaping.Input{
		Text:      runes,
		RunStart:  0,
		RunEnd:    len(runes),
		Direction: shapedTextDirection(runes),
		Face:      face,
		Size:      shapedSizePx(fontSize),
		Script:    shapedScriptFor(runes),
		Language:  shapedTextLanguage(runes),
	})
	if FontSpacing > 0 && len(runes) > 1 {
		out.AddLetterSpacing(shapedLetterSpacingPx(FontSpacing), false, false)
	}
	return shapedTextRun{out: out, runes: runes}, true
}

// shapedMeasureTextFRaw returns uncalibrated shaped width in pixels.
func shapedMeasureTextFRaw(text string, fontSize float32, bold, italic, mono, preview bool) (float32, bool) {
	run, ok := shapedShapeText(text, fontSize, bold, italic, mono, preview)
	if !ok {
		return 0, false
	}
	if len(run.runes) == 0 {
		return 0, true
	}
	if shapedTextUsesComplexScript(text) {
		if w, ok := shapedRunVisualWidth(run, fontSize); ok {
			return w, true
		}
	}
	return float32(run.out.Advance.Ceil()), true
}

// shapedMeasureTextF returns shaped text width in pixels. ok is false when no face is loaded.
func shapedMeasureTextF(text string, fontSize float32, bold, italic, mono, preview bool) (float32, bool) {
	if shapedNeedsSDFFallbackMeasure(text) {
		return 0, false
	}
	// T1.7: preview lane stays on SDF metrics until shaped preview is parity-tested.
	if preview && !shapedTextUsesComplexScript(text) {
		return 0, false
	}
	// Demo session: Latin chrome stays on SDF measure; editor/Notepad keeps shapedLatinMeasureActive.
	if !shapedLatinMeasureActive && !shapedTextUsesComplexScript(text) {
		return 0, false
	}
	key := shapedMeasureCacheKey{
		text:     text,
		fontSize: uint16(fontSize*2 + 0.5),
		flags:    shapedMeasureCacheFlags(bold, italic, mono, preview),
	}
	if w, ok := shapedMeasureCacheLookup(key); ok {
		return w, true
	}
	raw, ok := shapedMeasureTextFRaw(text, fontSize, bold, italic, mono, preview)
	if !ok {
		return 0, false
	}
	w := raw * shapedMeasureRatioFor(fontSize, preview, text)
	shapedMeasureCacheStore(key, w)
	return w, true
}

// shapedNeedsSDFFallbackMeasure reports strings that must use the legacy SDF string path.
func shapedNeedsSDFFallbackMeasure(text string) bool {
	for _, r := range text {
		if r >= 0x2500 && r <= 0x257F {
			return true
		}
	}
	return false
}
