// Package ui (continued) — Phosphor icon font backend.
package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type phosphorFontFace struct {
	font     rl.Font
	glyphs   map[string]glyphEntry
	loadedCP map[int32]struct{}
	ready    bool
	failed   bool
}

type glyphEntry struct {
	primary   rune
	secondary rune
}

type selectionFile struct {
	Icons []selectionIcon `json:"icons"`
}

type selectionIcon struct {
	Properties selectionProps `json:"properties"`
}

type selectionProps struct {
	Name  string `json:"name"`
	Code  int    `json:"code"`
	Codes []int  `json:"codes"`
}

// WarmGlyphNames are loaded into the icon-font atlas first (sharp, small set).
// Additional icons trigger a one-time atlas rebuild when first drawn.
var WarmGlyphNames = []string{
	PhosphorCaretLeft,
	PhosphorCaretRight,
	PhosphorDotsThree,
	PhosphorDotsThreeVertical,
	PhosphorHouse,
	PhosphorMagnifyingGlass,
	PhosphorBell,
	PhosphorGear,
	PhosphorUser,
	PhosphorPlus,
	PhosphorX,
	PhosphorXCircle,
	PhosphorMinus,
	PhosphorSquare,
	PhosphorResize,
	PhosphorCopy,
	PhosphorCaretUp,
	PhosphorCaretCircleDown,
	PhosphorCaretCircleUp,
	PhosphorTray,
	PhosphorList,
	PhosphorFunnel,
	PhosphorEnvelope,
	PhosphorStar,
	PhosphorCheck,
	PhosphorTable,
	PhosphorCalendar,
	PhosphorPencilSimple,
	PhosphorTrash,
	PhosphorArrowDropDown,
	PhosphorArrowDropUp,
	PhosphorCodeView,
}

func parsePhosphorGlyphs(path string) (map[string]glyphEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var sf selectionFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, err
	}
	glyphs := make(map[string]glyphEntry, len(sf.Icons))
	for _, ic := range sf.Icons {
		name := ic.Properties.Name
		if name == "" {
			continue
		}
		ge := glyphEntry{primary: rune(ic.Properties.Code)}
		if len(ic.Properties.Codes) >= 2 {
			ge.primary = rune(ic.Properties.Codes[0])
			ge.secondary = rune(ic.Properties.Codes[1])
		}
		glyphs[name] = ge
	}
	return glyphs, nil
}

// parsePhosphorSelection parses selection.json and returns both:
//   - glyphs: glyph name -> primary/secondary rune codes
//   - cps: a flattened, de-duplicated codepoint list (primary + secondary)
//
// This is used by unit tests to validate the selection.json format.
func parsePhosphorSelection(path string) (map[string]glyphEntry, []int32, error) {
	glyphs, err := parsePhosphorGlyphs(path)
	if err != nil {
		return nil, nil, err
	}
	entries := make([]glyphEntry, 0, len(glyphs))
	for _, ge := range glyphs {
		entries = append(entries, ge)
	}
	cps := codepointsForEntries(entries...)
	return glyphs, cps, nil
}

func codepointsForEntries(entries ...glyphEntry) []int32 {
	seen := make(map[int32]struct{})
	var cps []int32
	add := func(r rune) {
		if r <= 0 {
			return
		}
		cp := int32(r)
		// Skip ASCII control chars — they are not icon glyphs and make LoadFontEx warn.
		if cp < 32 {
			return
		}
		if _, ok := seen[cp]; ok {
			return
		}
		seen[cp] = struct{}{}
		cps = append(cps, cp)
	}
	for _, ge := range entries {
		add(ge.primary)
		add(ge.secondary)
	}
	return cps
}

func codepointsForNames(glyphs map[string]glyphEntry, weight PhosphorWeight, names []string) []int32 {
	var entries []glyphEntry
	for _, name := range names {
		key := phosphorSelectionName(name, weight)
		if ge, ok := glyphs[key]; ok {
			entries = append(entries, ge)
		}
	}
	return codepointsForEntries(entries...)
}

func (r *PhosphorRegistry) fontDir(weight PhosphorWeight) string {
	return filepath.Join(r.Root, "Fonts", string(weight))
}

func phosphorFontFiles(weight PhosphorWeight) []string {
	switch weight {
	case PhosphorBold:
		return []string{"Phosphor-Bold.ttf", "Phosphor-Bold.woff2", "Phosphor-Bold.woff"}
	case PhosphorFill:
		return []string{"Phosphor-Fill.ttf", "Phosphor-Fill.woff2", "Phosphor-Fill.woff"}
	case PhosphorLight:
		return []string{"Phosphor-Light.ttf", "Phosphor-Light.woff2", "Phosphor-Light.woff"}
	case PhosphorThin:
		return []string{"Phosphor-Thin.ttf", "Phosphor-Thin.woff2", "Phosphor-Thin.woff"}
	case PhosphorDuotone:
		return []string{"Phosphor-Duotone.ttf", "Phosphor-Duotone.woff2", "Phosphor-Duotone.woff"}
	default:
		return []string{"Phosphor.ttf", "Phosphor.woff2", "Phosphor.woff"}
	}
}

func (r *PhosphorRegistry) fontTTFPath(weight PhosphorWeight) string {
	dir := r.fontDir(weight)
	for _, name := range phosphorFontFiles(weight) {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func (face *phosphorFontFace) mergeCodepoints(cps []int32) []int32 {
	for _, cp := range cps {
		face.loadedCP[cp] = struct{}{}
	}
	out := make([]int32, 0, len(face.loadedCP))
	for cp := range face.loadedCP {
		out = append(out, cp)
	}
	return out
}

func filterIconFontCodepoints(cps []int32) []int32 {
	out := make([]int32, 0, len(cps))
	for _, cp := range cps {
		// Control chars are not glyphs; raylib warns and atlas packing can fail.
		if cp < 32 || cp > 0x10FFFF {
			continue
		}
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (r *PhosphorRegistry) loadFontFace(face *phosphorFontFace, weight PhosphorWeight, cps []int32) bool {
	ttf := r.fontTTFPath(weight)
	cps = filterIconFontCodepoints(cps)
	if ttf == "" || len(cps) == 0 {
		return false
	}
	atlas := r.atlasSize
	if atlas < 256 {
		atlas = 256
	}
	prev := face.font
	prevReady := face.ready
	font := rl.LoadFontEx(ttf, atlas, cps, int32(len(cps)))
	if font.Texture.ID == 0 {
		face.failed = true
		if prevReady && prev.Texture.ID != 0 {
			face.font = prev
			face.ready = true
		}
		return false
	}
	if prevReady && prev.Texture.ID != 0 {
		rl.UnloadFont(prev)
	}
	rl.GenTextureMipmaps(&font.Texture)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	face.font = font
	face.ready = true
	face.failed = false
	return true
}

func (r *PhosphorRegistry) ensureFont(weight PhosphorWeight) *phosphorFontFace {
	if weight == "" {
		weight = PhosphorRegular
	}
	r.fontMu.Lock()
	if r.fonts == nil {
		r.fonts = make(map[PhosphorWeight]*phosphorFontFace)
	}
	if f, ok := r.fonts[weight]; ok {
		r.fontMu.Unlock()
		return f
	}
	face := &phosphorFontFace{
		glyphs:   make(map[string]glyphEntry),
		loadedCP: make(map[int32]struct{}),
	}
	r.fonts[weight] = face
	r.fontMu.Unlock()

	sel := filepath.Join(r.fontDir(weight), "selection.json")
	glyphs, err := parsePhosphorGlyphs(sel)
	if err != nil {
		face.failed = true
		return face
	}
	face.glyphs = glyphs
	if r.fontTTFPath(weight) == "" {
		return face
	}
	cps := codepointsForNames(glyphs, weight, WarmGlyphNames)
	face.mergeCodepoints(cps)
	all := make([]int32, 0, len(face.loadedCP))
	for cp := range face.loadedCP {
		all = append(all, cp)
	}
	r.loadFontFace(face, weight, all)
	return face
}

func phosphorSelectionName(name string, weight PhosphorWeight) string {
	if weight == "" || weight == PhosphorRegular {
		return name
	}
	suffix := "-" + string(weight)
	if len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix {
		return name
	}
	return name + suffix
}

// FontFaceReady reports whether the icon-font atlas for weight loaded successfully.
func (r *PhosphorRegistry) FontFaceReady(weight PhosphorWeight) bool {
	if weight == "" {
		weight = PhosphorRegular
	}
	if r.fontTTFPath(weight) == "" {
		return false
	}
	r.fontMu.Lock()
	face := r.fonts[weight]
	r.fontMu.Unlock()
	if face != nil && face.ready {
		return true
	}
	face = r.ensureFont(weight)
	return face != nil && face.ready
}

// FontTTFPath returns the on-disk icon font for weight, or "" when missing.
func (r *PhosphorRegistry) FontTTFPath(weight PhosphorWeight) string {
	return r.fontTTFPath(weight)
}

// IconFontSummary returns active icon backend for F6 / stderr diagnostics.
func (r *PhosphorRegistry) IconFontSummary() string {
	if remixIcons.ready {
		return remixIconSummary()
	}
	if remixIcons.failed {
		return "png/svg (remix failed)"
	}
	return "png/svg (remix off)"
}

func phosphorFontDrawSize(dst rl.Rectangle) float32 {
	size := dst.Height
	if dst.Width > 0 && dst.Width < size {
		size = dst.Width
	}
	size = float32(int32(size + 0.5))
	if size < minRenderPx {
		size = minRenderPx
	}
	return size
}

func iconFontGlyphReadyInSet(font rl.Font, cp rune, loaded map[int32]struct{}) bool {
	if font.Texture.ID == 0 || cp <= 0 {
		return false
	}
	rec := rl.GetGlyphAtlasRec(font, int32(cp))
	if rec.Width <= 0 || rec.Height <= 0 {
		return false
	}
	// Missing subset members map to '?' — detect only when '?' was packed in this atlas.
	// If '?' is not loaded, GetGlyphAtlasRec('?') aliases the first glyph (false positive).
	if cp != '?' && loaded != nil {
		if _, hasQ := loaded[int32('?')]; hasQ {
			qRec := rl.GetGlyphAtlasRec(font, int32('?'))
			if qRec.Width > 0 && qRec.Height > 0 &&
				rec.X == qRec.X && rec.Y == qRec.Y && rec.Width == qRec.Width && rec.Height == qRec.Height {
				return false
			}
		}
	}
	return true
}

func (r *PhosphorRegistry) drawFont(dst rl.Rectangle, name string, weight PhosphorWeight, tint rl.Color) bool {
	return remixDrawIcon(dst, name, weight, tint, 1.0)
}

func (r *PhosphorRegistry) unloadFonts() {
	r.fontMu.Lock()
	defer r.fontMu.Unlock()
	for _, f := range r.fonts {
		if f != nil && f.ready && f.font.Texture.ID != 0 {
			rl.UnloadFont(f.font)
		}
	}
	r.fonts = make(map[PhosphorWeight]*phosphorFontFace)
}
