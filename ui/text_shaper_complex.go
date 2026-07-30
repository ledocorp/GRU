package ui

import (
	"os"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
)

// Complex-script fixtures (Engine T1.5).
const (
	ComplexScriptArabicSample     = "مرحبا بالعالم"
	ComplexScriptDevanagariSample   = "नमस्ते दुनिया"
	ComplexScriptMixedSample      = "Hello — " + ComplexScriptArabicSample
)

var complexScriptFont rl.Font
var complexScriptFontReady bool

var (
	shapedDevanagari *font.Face
)

var devanagariFontCandidates = []string{
	"C:/Windows/Fonts/Nirmala.ttc",
	"C:/Windows/Fonts/Nirmala.ttf",
	"C:/Windows/Fonts/NirmalaB.ttf",
	"C:/Windows/Fonts/NirmalaS.ttf",
	"C:/Windows/Fonts/mangal.ttf",
	"assets/fonts/NotoSansDevanagari-Regular.ttf",
}

// shapedFontCollection maps dominant script → TTF face (Skia FontCollection pattern).
// Latin/Arabic use the UI stack from pickShapedFace; Devanagari uses Nirmala/Noto when present.
func shapedFaceForScript(sc language.Script, bold, italic, mono, preview bool) *font.Face {
	switch sc {
	case language.Devanagari:
		if shapedDevanagari != nil {
			return shapedDevanagari
		}
	}
	return pickShapedFace(bold, italic, mono, preview)
}

func initShapedScriptFaces() {
	if shapedDevanagari != nil {
		return
	}
	if p := firstExisting(devanagariFontCandidates...); p != "" {
		if f, err := loadShapedFace(p); err == nil {
			shapedDevanagari = f
		}
	}
}

func dominantShapedScript(runes []rune) language.Script {
	for _, r := range runes {
		if r <= ' ' || r == '\n' || r == '\t' {
			continue
		}
		sc := language.LookupScript(r)
		if sc != language.Common && sc != language.Inherited {
			return sc
		}
	}
	return language.Latin
}

func pickShapedFaceForText(text string, bold, italic, mono, preview bool) *font.Face {
	runes := []rune(text)
	mono = shapedEffectiveMono(mono, text)
	return shapedFaceForScript(dominantShapedScript(runes), bold, italic, mono, preview)
}

// shapedEffectiveMono keeps editor mono for Latin only. Mono faces (Fira Code) lack
// Arabic/Devanagari outlines — using them produces "?" tofu in Notepad.
func shapedEffectiveMono(mono bool, text string) bool {
	if mono && shapedTextUsesComplexScript(text) {
		return false
	}
	return mono
}

// shapedTextUsesComplexScript reports non-Latin text that should use HarfBuzz shaping (T1.5).
func shapedTextUsesComplexScript(text string) bool {
	for _, r := range text {
		if r <= ' ' || r == '\n' || r == '\t' {
			continue
		}
		sc := language.LookupScript(r)
		if sc != language.Latin && sc != language.Common && sc != language.Inherited {
			return true
		}
	}
	return false
}

// shapedTextParagraphDirection returns paragraph direction from the first strong
// character (UAX #9). Mixed strings like "Arabic sample: …" stay LTR; pure Arabic stays RTL.
func shapedTextParagraphDirection(runes []rune) di.Direction {
	for _, r := range runes {
		if r <= ' ' || r == '\n' || r == '\t' {
			continue
		}
		sc := language.LookupScript(r)
		switch sc {
		case language.Arabic, language.Hebrew:
			return di.DirectionRTL
		default:
			return di.DirectionLTR
		}
	}
	return di.DirectionLTR
}

// shapedTextDirection is an alias for paragraph direction (T1.8 bidi fix).
func shapedTextDirection(runes []rune) di.Direction {
	return shapedTextParagraphDirection(runes)
}

func shapedTextLanguage(runes []rune) language.Language {
	for _, r := range runes {
		sc := language.LookupScript(r)
		switch sc {
		case language.Arabic:
			return language.NewLanguage("ar")
		case language.Hebrew:
			return language.NewLanguage("he")
		case language.Devanagari:
			return language.NewLanguage("hi")
		}
	}
	return language.NewLanguage("en")
}

func complexScriptFixtureRunes() []rune {
	seen := make(map[rune]bool)
	var cp []rune
	for _, s := range []string{
		ComplexScriptArabicSample,
		ComplexScriptDevanagariSample,
		ComplexScriptMixedSample,
	} {
		for _, r := range s {
			if r < 32 {
				continue
			}
			if !seen[r] {
				seen[r] = true
				cp = append(cp, r)
			}
		}
	}
	// Latin atlas coverage for mixed fixture.
	for _, r := range sdfCodepoints() {
		if !seen[r] {
			seen[r] = true
			cp = append(cp, r)
		}
	}
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	return cp
}

// EnsureComplexScriptUIFont loads an SDF atlas with fixture codepoints (Segoe/Inter path).
func EnsureComplexScriptUIFont() {
	if complexScriptFontReady || !sdfReady {
		return
	}
	path := LoadedUIFontPath
	if path == "" {
		path = firstExisting(fontCandidates...)
	}
	if path == "" {
		return
	}
	fileData, err := readUIFontFile(path)
	if err != nil {
		return
	}
	cp := complexScriptFixtureRunes()
	cps := make([]int32, len(cp))
	for i, r := range cp {
		cps[i] = int32(r)
	}
	atlasSize := SDFFont.BaseSize
	if atlasSize <= 0 {
		atlasSize = 176
	}
	glyphs := rl.LoadFontData(fileData, atlasSize, cps, int32(len(cps)), int32(rl.FontSdf))
	if len(glyphs) == 0 {
		return
	}
	glyphRecs := make([]*rl.Rectangle, 1)
	atlas := rl.GenImageFontAtlas(glyphs, glyphRecs, atlasSize, 4, 0)
	if atlas.Width == 0 || glyphRecs[0] == nil {
		return
	}
	tex := rl.LoadTextureFromImage(&atlas)
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	rl.UnloadImage(&atlas)
	complexScriptFont = rl.Font{
		BaseSize:     atlasSize,
		CharsCount:   int32(len(glyphs)),
		CharsPadding: 4,
		Texture:      tex,
		Recs:         glyphRecs[0],
		Chars:        &glyphs[0],
	}
	complexScriptFontReady = true
}

func unloadComplexScriptUIFont() {
	if complexScriptFontReady {
		rl.UnloadFont(complexScriptFont)
		complexScriptFont = rl.Font{}
		complexScriptFontReady = false
	}
}

func readUIFontFile(path string) ([]byte, error) {
	if data, err := ReadAssetFile(path); err == nil {
		return data, nil
	}
	return os.ReadFile(path)
}
