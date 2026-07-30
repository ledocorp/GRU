// Package ui (continued) — Remix Icon TTF backend (primary icon font for Gru).
//
// All ui.Phosphor* names resolve to glyphs in assets/fonts/remixicon.ttf via
// remixicon.css codepoints. Phosphor PNG/SVG under assets/icons/phosphor remain
// fallback only when a Remix class is missing.
//
// Migration order (May 2026):
//  1. Remix TTF primary draw path (this file) — done
//  2. Batch demo polish: SearchBar (batch2), Toolbar/Ribbon (batch9), ColorPicker split
//  3. Deprecate phosphor-fetch-fonts / Phosphor.ttf when PNG fallback unused
//
// See docs/ICONS.md and docs/WIDGET_ROADMAP.md §7 Sprint F.
package ui

import (
	"regexp"
	"sort"
	"strconv"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	remixIconCSSPath = "assets/fonts/remixicon.css"
	remixIconTTFPath = "assets/fonts/remixicon.ttf"
)

// phosphorToRemix maps legacy Phosphor kebab names to Remix CSS class bases
// (without -line/-fill). Default is identity: "{name}-line" / "{name}-fill".
var phosphorToRemix = map[string]string{
	PhosphorHouse:           "home",
	PhosphorMagnifyingGlass: "search",
	PhosphorGear:            "settings",
	PhosphorUsers:           "team",
	PhosphorEnvelope:        "mail",
	PhosphorTray:            "inbox",
	PhosphorList:            "list-unordered",
	PhosphorPlus:            "add",
	PhosphorMinus:           "subtract",
	PhosphorCopy:            "file-copy",
	PhosphorX:               "close",
	PhosphorXCircle:         "close-circle",
	PhosphorCaretLeft:       "arrow-left-s",
	PhosphorCaretRight:      "arrow-right-s",
	PhosphorCaretUp:         "arrow-up-s",
	PhosphorCaretDown:       "arrow-down-s",
	PhosphorArrowDropDown:   "arrow-drop-down",
	PhosphorArrowDropUp:     "arrow-drop-up",
	PhosphorCodeView:        "code-view",
	PhosphorCodeBlock:       "code-block",
	PhosphorInfoI:           "info-i",
	PhosphorTextWrap:        "text-wrap",
	PhosphorCaretCircleLeft:  "arrow-left-circle",
	PhosphorCaretCircleRight: "arrow-right-circle",
	PhosphorCaretCircleUp:    "arrow-up-circle",
	PhosphorCaretCircleDown:  "arrow-down-circle",
	PhosphorDotsThree:         "more",
	PhosphorDotsThreeVertical: "more-2",
	PhosphorFunnel:            "filter",
	PhosphorPencilSimple:      "pencil",
	PhosphorTrash:             "delete-bin",
	PhosphorCalendarBlank:     "calendar-2",
	PhosphorResize:            "expand-diagonal-s",
	PhosphorWifiHigh:          "wifi",
}

var remixCSSRule = regexp.MustCompile(`\.ri-([a-z0-9-]+):before\s*\{\s*content:\s*"\\([0-9a-fA-F]+)"`)

type remixIconRegistry struct {
	mu sync.Mutex

	font      rl.Font
	ready     bool
	failed    bool
	atlasSize int32

	cssClasses map[string]rune // ri class suffix → codepoint
	loadedCP   map[int32]struct{}
}

var remixIcons remixIconRegistry

func parseRemixIconCSS(path string) (map[string]rune, error) {
	data, err := ReadAssetFile(path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]rune)
	for _, m := range remixCSSRule.FindAllSubmatch(data, -1) {
		class := string(m[1])
		cp, err := strconv.ParseInt(string(m[2]), 16, 32)
		if err != nil || cp <= 0 {
			continue
		}
		out[class] = rune(cp)
	}
	return out, nil
}

func remixFilled(weight PhosphorWeight) bool {
	switch weight {
	case PhosphorFill, PhosphorDuotone, PhosphorBold:
		return true
	default:
		return false
	}
}

func remixClassBase(phosphorName string) string {
	if base, ok := phosphorToRemix[phosphorName]; ok {
		return base
	}
	return phosphorName
}

func remixLookupClass(class string, classes map[string]rune) (rune, bool) {
	if cp, ok := classes[class]; ok {
		return cp, true
	}
	return 0, false
}

func remixCodepointFor(phosphorName string, weight PhosphorWeight, classes map[string]rune) (rune, bool) {
	base := remixClassBase(phosphorName)
	if remixFilled(weight) {
		if cp, ok := remixLookupClass(base+"-fill", classes); ok {
			return cp, true
		}
	}
	if cp, ok := remixLookupClass(base+"-line", classes); ok {
		return cp, true
	}
	if cp, ok := remixLookupClass(base, classes); ok {
		return cp, true
	}
	// Direct phosphor name (e.g. "square" → square-line).
	if cp, ok := remixLookupClass(phosphorName+"-line", classes); ok {
		return cp, true
	}
	if remixFilled(weight) {
		if cp, ok := remixLookupClass(phosphorName+"-fill", classes); ok {
			return cp, true
		}
	}
	return 0, false
}

func initRemixIconAtlas(atlasSize int32) {
	remixIcons.mu.Lock()
	defer remixIcons.mu.Unlock()
	if remixIcons.ready || remixIcons.failed {
		return
	}
	if atlasSize < 128 {
		atlasSize = 128
	}
	remixIcons.atlasSize = atlasSize

	classes, err := parseRemixIconCSS(remixIconCSSPath)
	if err != nil {
		remixIcons.failed = true
		rl.TraceLog(rl.LogWarning, "Gru remixicon: failed to read %s", remixIconCSSPath)
		return
	}
	remixIcons.cssClasses = classes

	cps := remixWarmCodepoints(classes)
	if len(cps) == 0 {
		remixIcons.failed = true
		rl.TraceLog(rl.LogWarning, "Gru remixicon: no warm codepoints from %s", remixIconCSSPath)
		return
	}
	if !remixIcons.loadSubsetLocked(cps) {
		return
	}
	remixIcons.ready = true
	rl.TraceLog(rl.LogInfo, "Gru remixicon: atlas ready (%s, %d glyphs)", remixIconTTFPath, len(cps))
}

func remixWarmCodepoints(classes map[string]rune) []int32 {
	seen := make(map[int32]struct{})
	// Space + '?' stabilize raylib/stb atlas packing for sparse PUA icon subsets.
	seen[32] = struct{}{}
	seen[int32('?')] = struct{}{}
	add := func(name string, weight PhosphorWeight) {
		cp, ok := remixCodepointFor(name, weight, classes)
		if !ok {
			return
		}
		seen[int32(cp)] = struct{}{}
	}
	// Title-bar window controls (direct Remix codepoints).
	for _, cp := range remixTitleBarCPs {
		seen[cp] = struct{}{}
	}
	// Same controls via Phosphor/Remix names — keeps maximize (U+F3DC) packed when
	// direct CP rasterization is flaky in a 4-glyph atlas.
	for _, name := range []string{
		PhosphorX, PhosphorMinus, PhosphorSquare, PhosphorCopy, PhosphorResize,
	} {
		add(name, PhosphorRegular)
		add(name, PhosphorFill)
	}
	// Full warm set when requested (Studio); lean hosts still get chrome essentials above.
	if IconsEagerWarm {
		for _, name := range WarmGlyphNames {
			add(name, PhosphorRegular)
			add(name, PhosphorFill)
		}
	}
	out := make([]int32, 0, len(seen))
	for cp := range seen {
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return filterIconFontCodepoints(out)
}

func (r *remixIconRegistry) loadSubsetLocked(cps []int32) bool {
	if len(cps) == 0 {
		return false
	}
	prev := r.font
	prevReady := r.ready
	fileData, err := ReadAssetFile(remixIconTTFPath)
	if err != nil {
		r.failed = true
		rl.TraceLog(rl.LogWarning, "Gru remixicon: ReadAssetFile failed for %s", remixIconTTFPath)
		return false
	}
	font := loadFontFromMemory(fileData, r.atlasSize, cps)
	if font.Texture.ID == 0 {
		r.failed = true
		rl.TraceLog(rl.LogWarning, "Gru remixicon: LoadFontEx failed for %s", remixIconTTFPath)
		return false
	}
	// Do not mark the whole registry failed when one CP is missing — that left
	// title bars with no icons after a single bad maximize glyph (U+F3DC).
	loaded := make(map[int32]struct{}, len(cps))
	okCount := 0
	for _, cp := range cps {
		probe := make(map[int32]struct{}, len(loaded)+1)
		for k := range loaded {
			probe[k] = struct{}{}
		}
		probe[cp] = struct{}{}
		if iconFontGlyphReadyInSet(font, rune(cp), probe) {
			loaded[cp] = struct{}{}
			okCount++
			continue
		}
		rl.TraceLog(rl.LogWarning, "Gru remixicon: missing glyph U+%04X (skipped)", cp)
	}
	if okCount == 0 {
		r.failed = true
		rl.UnloadFont(font)
		rl.TraceLog(rl.LogWarning, "Gru remixicon: no glyphs packed")
		return false
	}
	if prevReady && prev.Texture.ID != 0 {
		rl.UnloadFont(prev)
	}
	rl.GenTextureMipmaps(&font.Texture)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	r.font = font
	r.loadedCP = loaded
	r.failed = false
	return true
}

// ensureTitleBarCodepoint packs a missing title-bar CP into the Remix atlas.
func ensureTitleBarCodepoint(cp rune) {
	if cp <= 0 {
		return
	}
	remixIcons.mu.Lock()
	defer remixIcons.mu.Unlock()
	if remixIcons.failed {
		return
	}
	icp := int32(cp)
	if _, has := remixIcons.loadedCP[icp]; has && remixIcons.ready {
		return
	}
	iconWarmPending.Store(true)
	defer iconWarmPending.Store(false)
	need := make([]int32, 0, len(remixIcons.loadedCP)+3)
	for k := range remixIcons.loadedCP {
		need = append(need, k)
	}
	need = append(need, icp, 32, int32('?'))
	sort.Slice(need, func(i, j int) bool { return need[i] < need[j] })
	need = filterIconFontCodepoints(need)
	if remixIcons.loadSubsetLocked(need) {
		remixIcons.ready = true
	}
}

func remixHasGlyph(phosphorName string, weight PhosphorWeight) bool {
	remixIcons.mu.Lock()
	classes := remixIcons.cssClasses
	ready := remixIcons.ready
	remixIcons.mu.Unlock()
	if len(classes) == 0 {
		return false
	}
	_, ok := remixCodepointFor(phosphorName, weight, classes)
	return ok && ready
}

func remixEnsureGlyph(phosphorName string, weight PhosphorWeight) bool {
	remixIcons.mu.Lock()
	defer remixIcons.mu.Unlock()
	if remixIcons.failed || len(remixIcons.cssClasses) == 0 {
		return false
	}
	cp, ok := remixCodepointFor(phosphorName, weight, remixIcons.cssClasses)
	if !ok {
		return false
	}
	icp := int32(cp)
	if _, has := remixIcons.loadedCP[icp]; has && remixIcons.ready {
		return true
	}
	iconWarmPending.Store(true)
	defer iconWarmPending.Store(false)
	need := make([]int32, 0, len(remixIcons.loadedCP)+1)
	for k := range remixIcons.loadedCP {
		need = append(need, k)
	}
	need = append(need, icp)
	sort.Slice(need, func(i, j int) bool { return need[i] < need[j] })
	need = filterIconFontCodepoints(need)
	if !remixIcons.loadSubsetLocked(need) {
		return false
	}
	remixIcons.ready = true
	return true
}

func remixDrawIcon(dst rl.Rectangle, phosphorName string, weight PhosphorWeight, tint rl.Color, strokeScale float32) bool {
	if !remixEnsureGlyph(phosphorName, weight) {
		return false
	}
	remixIcons.mu.Lock()
	cp, ok := remixCodepointFor(phosphorName, weight, remixIcons.cssClasses)
	loaded := remixIcons.loadedCP
	font := remixIcons.font
	ready := remixIcons.ready
	remixIcons.mu.Unlock()
	if !ok || !ready {
		return false
	}
	return remixDrawCodepoint(font, loaded, dst, cp, tint, strokeScale)
}

func remixDrawCodepoint(font rl.Font, loaded map[int32]struct{}, dst rl.Rectangle, cp rune, tint rl.Color, strokeScale float32) bool {
	if font.Texture.ID == 0 || cp <= 0 {
		return false
	}
	if !iconFontGlyphReadyInSet(font, cp, loaded) {
		return false
	}
	if strokeScale <= 0 {
		strokeScale = 1
	}
	if strokeScale > 1 {
		strokeScale = 1
	}
	dst = snapPhosphorRect(dst)
	size := phosphorFontDrawSize(dst) * strokeScale
	if size < minRenderPx {
		size = minRenderPx
	}
	text := string(cp)
	tw := rl.MeasureTextEx(font, text, size, 0).X
	if tw <= 0.5 {
		return false
	}
	posY := float32(int32(dst.Y + (dst.Height-size)/2 + 0.5))
	pos := rl.NewVector2(float32(int32(dst.X+(dst.Width-tw)/2+0.5)), posY)
	rl.DrawTextEx(font, text, pos, size, 0, tint)
	return true
}

func remixIconSummary() string {
	if remixIcons.ready {
		return "remixicon.ttf"
	}
	if remixIcons.failed {
		return "remixicon(failed)"
	}
	return "remixicon(off)"
}

func unloadRemixIcons() {
	remixIcons.mu.Lock()
	defer remixIcons.mu.Unlock()
	if remixIcons.ready && remixIcons.font.Texture.ID != 0 {
		rl.UnloadFont(remixIcons.font)
	}
	remixIcons.font = rl.Font{}
	remixIcons.loadedCP = nil
	remixIcons.cssClasses = nil
	remixIcons.ready = false
	remixIcons.failed = false
}
