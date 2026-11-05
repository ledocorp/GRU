package ui

import (
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// uiFontStack holds Poppins UI faces + Fira Code for code (SDF atlases).
type uiFontStack struct {
	regular    rl.Font
	bold       rl.Font
	italic     rl.Font
	boldItalic rl.Font
	mono       rl.Font
	glyphs     [][]rl.GlyphInfo
	ready      bool
	monoReady  bool
}

// useBundledUIFontStack loads Poppins/Inter TTF cuts when true. Default false → Segoe UI.
const useBundledUIFontStack = false

var fonts uiFontStack

// previewFonts holds Inter faces for the markdown preview lane (UI chrome stays on Segoe).
var previewFonts uiFontStack

type uiFacePaths struct {
	regular    string
	bold       string
	italic     string
	boldItalic string
}

func poppinsFaceCandidates() []uiFacePaths {
	// Static Poppins cuts only (no variable font — heavy default weight in SDF).
	return []uiFacePaths{
		{
			"assets/fonts/poppins/Poppins-Regular.ttf",
			"assets/fonts/poppins/Poppins-Bold.ttf",
			"assets/fonts/poppins/Poppins-Italic.ttf",
			"assets/fonts/poppins/Poppins-BoldItalic.ttf",
		},
		{
			"assets/fonts/Poppins/static/Poppins-Regular.ttf",
			"assets/fonts/Poppins/static/Poppins-Bold.ttf",
			"assets/fonts/Poppins/static/Poppins-Italic.ttf",
			"assets/fonts/Poppins/static/Poppins-BoldItalic.ttf",
		},
		{
			"assets/fonts/Poppins-Regular.ttf",
			"assets/fonts/Poppins-Bold.ttf",
			"assets/fonts/Poppins-Italic.ttf",
			"assets/fonts/Poppins-BoldItalic.ttf",
		},
	}
}

var monoFontCandidates = []string{
	"assets/fonts/firacode/static/FiraCode-Regular.ttf",
	"assets/fonts/FiraCode/static/FiraCode-Regular.ttf",
	"assets/fonts/firacode/FiraCode-VariableFont_wght.ttf",
	"assets/fonts/firacode/static/FiraCode-Medium.ttf",
	"assets/fonts/FiraCode/static/FiraCode-Medium.ttf",
}

func interFaceCandidates() []uiFacePaths {
	// Default resolve order for family init / HUD (24pt mid ladder).
	// Sticky strikes pick size-matched optical cuts via interFaceForStickySize.
	out := append([]uiFacePaths{}, interOpticalCandidates(24)...)
	out = append(out, interOpticalCandidates(28)...)
	out = append(out, interOpticalCandidates(18)...)
	out = append(out, []uiFacePaths{{
		"assets/fonts/Inter/Inter-Regular.ttf",
		"assets/fonts/Inter/Inter-Bold.ttf",
		"assets/fonts/Inter/Inter-Italic.ttf",
		"assets/fonts/Inter/Inter-BoldItalic.ttf",
	}}...)
	return out
}

func interOpticalCandidates(pt int) []uiFacePaths {
	tag := fmt.Sprintf("Inter_%dpt", pt)
	return []uiFacePaths{
		{
			fmt.Sprintf("assets/fonts/Inter/static/%s-Regular.ttf", tag),
			fmt.Sprintf("assets/fonts/Inter/static/%s-Bold.ttf", tag),
			fmt.Sprintf("assets/fonts/Inter/static/%s-Italic.ttf", tag),
			fmt.Sprintf("assets/fonts/Inter/static/%s-BoldItalic.ttf", tag),
		},
		{
			fmt.Sprintf("assets/fonts/inter/static/%s-Regular.ttf", tag),
			fmt.Sprintf("assets/fonts/inter/static/%s-Bold.ttf", tag),
			fmt.Sprintf("assets/fonts/inter/static/%s-Italic.ttf", tag),
			fmt.Sprintf("assets/fonts/inter/static/%s-BoldItalic.ttf", tag),
		},
	}
}

// interMediumCandidates — small sticky bold prefers Medium so counters stay open.
func interMediumCandidates(pt int) []string {
	tag := fmt.Sprintf("Inter_%dpt", pt)
	return []string{
		fmt.Sprintf("assets/fonts/Inter/static/%s-Medium.ttf", tag),
		fmt.Sprintf("assets/fonts/inter/static/%s-Medium.ttf", tag),
	}
}

// interSemiBoldCandidates: UI "bold" prefers SemiBold — Inter Bold fills counters
// at 16–20px and looks uneven next to Regular on Font Faces.
func interSemiBoldCandidates(pt int) []string {
	tag := fmt.Sprintf("Inter_%dpt", pt)
	return []string{
		fmt.Sprintf("assets/fonts/Inter/static/%s-SemiBold.ttf", tag),
		fmt.Sprintf("assets/fonts/inter/static/%s-SemiBold.ttf", tag),
	}
}

// interFaceForStickySize picks an Inter optical cut for one sticky px size.
// Baking Inter_24pt at 16px (or 18pt at 32px) is why preview looked crunchy vs Segoe.
func interFaceForStickySize(sz int, bold bool) string {
	pt := 24
	switch {
	case sz <= 20:
		pt = 18
	case sz <= 26:
		pt = 24
	default:
		pt = 28
	}
	if bold {
		// ≤18: Medium opens counters (SemiBold still clogs at soft LCD).
		if sz <= 18 {
			for _, cand := range interMediumCandidates(pt) {
				if p := firstExisting(cand); p != "" {
					return p
				}
			}
		}
		for _, cand := range interSemiBoldCandidates(pt) {
			if p := firstExisting(cand); p != "" {
				return p
			}
		}
	}
	for _, set := range interOpticalCandidates(pt) {
		cand := set.regular
		if bold {
			cand = set.bold
		}
		if p := firstExisting(cand); p != "" {
			return p
		}
	}
	// Fall back across other opticals / generic Inter.
	for _, set := range interFaceCandidates() {
		cand := set.regular
		if bold {
			cand = set.bold
		}
		if p := firstExisting(cand); p != "" {
			return p
		}
	}
	return ""
}

func uiFontFileExists(path string) bool {
	if path == "" {
		return false
	}
	if AndroidGLES() {
		_, err := ReadAssetFile(path)
		return err == nil
	}
	_, err := os.Stat(path)
	return err == nil
}

func firstExisting(paths ...string) string {
	for _, p := range paths {
		if uiFontFileExists(p) {
			return p
		}
	}
	return ""
}

func loadSingleSDFFont(path string, atlasSize int32) (rl.Font, []rl.GlyphInfo, bool) {
	fileData, err := ReadAssetFile(path)
	if err != nil {
		return rl.Font{}, nil, false
	}
	cp := sdfCodepoints()
	cps := make([]int32, len(cp))
	for i, r := range cp {
		cps[i] = int32(r)
	}
	glyphs := rl.LoadFontData(fileData, atlasSize, cps, int32(len(cps)), int32(rl.FontSdf))
	if len(glyphs) == 0 {
		return rl.Font{}, nil, false
	}
	glyphRecs := make([]*rl.Rectangle, 1)
	atlas := rl.GenImageFontAtlas(glyphs, glyphRecs, atlasSize, 4, 0)
	if atlas.Width == 0 || glyphRecs[0] == nil {
		return rl.Font{}, nil, false
	}
	tex := rl.LoadTextureFromImage(&atlas)
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	rl.UnloadImage(&atlas)
	font := rl.Font{
		BaseSize:     atlasSize,
		CharsCount:   int32(len(glyphs)),
		CharsPadding: 4,
		Texture:      tex,
		Recs:         glyphRecs[0],
		Chars:        &glyphs[0],
	}
	return font, glyphs, true
}

func unloadFont(f rl.Font) {
	if f.BaseSize > 0 {
		rl.UnloadFont(f)
	}
}

// initUIFontStack loads Poppins (regular/bold/italic/bold-italic) and Fira Code SDF atlases.
func initUIFontStack(atlasSize int32) bool {
	for _, set := range poppinsFaceCandidates() {
		regPath := firstExisting(set.regular)
		if regPath == "" {
			continue
		}
		reg, g0, ok := loadSingleSDFFont(regPath, atlasSize)
		if !ok {
			continue
		}
		fonts.regular = reg
		fonts.glyphs = append(fonts.glyphs, g0)

		if p := firstExisting(set.bold); p != "" {
			if f, g, ok := loadSingleSDFFont(p, atlasSize); ok {
				fonts.bold = f
				fonts.glyphs = append(fonts.glyphs, g)
			}
		}
		if p := firstExisting(set.italic); p != "" {
			if f, g, ok := loadSingleSDFFont(p, atlasSize); ok {
				fonts.italic = f
				fonts.glyphs = append(fonts.glyphs, g)
			}
		}
		if p := firstExisting(set.boldItalic); p != "" {
			if f, g, ok := loadSingleSDFFont(p, atlasSize); ok {
				fonts.boldItalic = f
				fonts.glyphs = append(fonts.glyphs, g)
			}
		}
		LoadedUIFontPath = regPath
		break
	}
	if fonts.regular.BaseSize == 0 {
		return false
	}
	if p := firstExisting(monoFontCandidates...); p != "" {
		if f, g, ok := loadSingleSDFFont(p, atlasSize); ok {
			fonts.mono = f
			fonts.monoReady = true
			fonts.glyphs = append(fonts.glyphs, g)
		}
	}
	fonts.ready = true
	SDFFont = fonts.regular
	return true
}

func unloadUIFontStack() {
	unloadFont(fonts.regular)
	unloadFont(fonts.bold)
	unloadFont(fonts.italic)
	unloadFont(fonts.boldItalic)
	unloadFont(fonts.mono)
	fonts = uiFontStack{}
	unloadPreviewUIFontStack()
}

func unloadPreviewUIFontStack() {
	unloadFont(previewFonts.regular)
	unloadFont(previewFonts.bold)
	unloadFont(previewFonts.italic)
	unloadFont(previewFonts.boldItalic)
	previewFonts = uiFontStack{}
}

// initUIAuxFonts loads Fira Code even when the main UI uses legacy Segoe (editor + code).
func initUIAuxFonts(atlasSize int32) {
	if fonts.monoReady {
		return
	}
	if p := firstExisting(monoFontCandidates...); p != "" {
		if f, g, ok := loadSingleSDFFont(p, atlasSize); ok {
			fonts.mono = f
			fonts.monoReady = true
			fonts.glyphs = append(fonts.glyphs, g)
		}
	}
}

// initPreviewInterStack loads Inter for markdown preview RichText only.
func initPreviewInterStack(atlasSize int32) bool {
	if previewFonts.ready {
		return true
	}
	for _, set := range interFaceCandidates() {
		regPath := firstExisting(set.regular)
		if regPath == "" {
			continue
		}
		reg, g0, ok := loadSingleSDFFont(regPath, atlasSize)
		if !ok {
			continue
		}
		previewFonts.regular = reg
		previewFonts.glyphs = append(previewFonts.glyphs, g0)
		if p := firstExisting(set.bold); p != "" {
			if f, g, ok := loadSingleSDFFont(p, atlasSize); ok {
				previewFonts.bold = f
				previewFonts.glyphs = append(previewFonts.glyphs, g)
			}
		}
		if p := firstExisting(set.italic); p != "" {
			if f, g, ok := loadSingleSDFFont(p, atlasSize); ok {
				previewFonts.italic = f
				previewFonts.glyphs = append(previewFonts.glyphs, g)
			}
		}
		if p := firstExisting(set.boldItalic); p != "" {
			if f, g, ok := loadSingleSDFFont(p, atlasSize); ok {
				previewFonts.boldItalic = f
				previewFonts.glyphs = append(previewFonts.glyphs, g)
			}
		}
		previewFonts.ready = true
		return true
	}
	return false
}

// EnsureUIFontAux loads Fira Code on first mono draw (editor / code).
func EnsureUIFontAux(atlasSize int32) {
	if atlasSize <= 0 {
		atlasSize = 176
	}
	initUIAuxFonts(atlasSize)
}

// EnsurePreviewUIFont loads Inter on first markdown preview draw.
func EnsurePreviewUIFont(atlasSize int32) {
	if atlasSize <= 0 {
		atlasSize = 176
	}
	initPreviewInterStack(atlasSize)
}

func pickUIFont(bold, italic, mono, preview bool) rl.Font {
	if mono {
		if !fonts.monoReady {
			EnsureUIFontAux(SDFFont.BaseSize)
		}
		if fonts.monoReady {
			return fonts.mono
		}
	}
	if preview {
		if !previewFonts.ready {
			EnsurePreviewUIFont(SDFFont.BaseSize)
		}
		if previewFonts.ready {
			return pickFaceStack(previewFonts, bold, italic)
		}
	}
	if fonts.ready {
		return pickFaceStack(fonts, bold, italic)
	}
	return SDFFont
}

func pickFaceStack(stack uiFontStack, bold, italic bool) rl.Font {
	switch {
	case bold && italic && stack.boldItalic.BaseSize > 0:
		return stack.boldItalic
	case bold && stack.bold.BaseSize > 0:
		return stack.bold
	case italic && stack.italic.BaseSize > 0:
		return stack.italic
	default:
		return stack.regular
	}
}

func uiFontUsesBoldShader(bold, italic, mono, preview bool) bool {
	if mono && fonts.monoReady {
		return false
	}
	stack := fonts
	if preview && previewFonts.ready {
		stack = previewFonts
	} else if !fonts.ready {
		return bold
	}
	if bold && italic && stack.boldItalic.BaseSize > 0 {
		return false
	}
	if bold && stack.bold.BaseSize > 0 {
		return false
	}
	if italic && stack.italic.BaseSize > 0 {
		return false
	}
	return bold
}
