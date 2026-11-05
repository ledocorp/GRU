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
	PhosphorHouse:             "home",
	PhosphorMagnifyingGlass:   "search",
	PhosphorGear:              "settings",
	PhosphorUsers:             "team",
	PhosphorEnvelope:          "mail",
	PhosphorTray:              "inbox",
	PhosphorList:              "list-unordered",
	PhosphorPlus:              "add",
	PhosphorMinus:             "subtract",
	PhosphorCopy:              "file-copy",
	PhosphorX:                 "close",
	PhosphorXCircle:           "close-circle",
	PhosphorCaretLeft:         "arrow-left-s",
	PhosphorCaretRight:        "arrow-right-s",
	PhosphorCaretUp:           "arrow-up-s",
	PhosphorCaretDown:         "arrow-down-s",
	PhosphorArrowDropDown:     "arrow-drop-down",
	PhosphorArrowDropUp:       "arrow-drop-up",
	PhosphorCodeView:          "code-view",
	PhosphorCodeBlock:         "code-block",
	PhosphorInfoI:             "info-i",
	PhosphorTextWrap:          "text-wrap",
	PhosphorCaretCircleLeft:   "arrow-left-circle",
	PhosphorCaretCircleRight:  "arrow-right-circle",
	PhosphorCaretCircleUp:     "arrow-up-circle",
	PhosphorCaretCircleDown:   "arrow-down-circle",
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

type remixIconStrike struct {
	font     rl.Font
	loadedCP map[int32]struct{}
	basePx   int32
}

type remixIconRegistry struct {
	mu sync.Mutex

	font      rl.Font // Path A / SSAA>1 single atlas (large BaseSize)
	ready     bool
	failed    bool
	atlasSize int32

	cssClasses map[string]rune // ri class suffix → codepoint
	loadedCP   map[int32]struct{}

	// G1.3 sticky strikes @ SSAA 1× — same idea as Latin FT grey strikes:
	// bake glyphs at exact client px, draw scale≈1. Greyscale only (no ClearType —
	// icon silhouettes on colored chrome must not get LCD fringe).
	strikes map[int32]*remixIconStrike
}

var remixIcons remixIconRegistry

// iconStrikeWarmSizes are baked at InitIcons under PreferIconSharp1x.
// Includes 16 for title-bar − □ ×; 20/24/28 for nav/FAB chrome.
var iconStrikeWarmSizes = []int32{16, 18, 20, 24, 28}

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
	if PreferIconSharp1x() {
		rl.TraceLog(rl.LogInfo, "Gru remixicon: sticky strikes ready (%s, %d glyphs, warm %v)", remixIconTTFPath, len(cps), iconStrikeWarmSizes)
	} else {
		rl.TraceLog(rl.LogInfo, "Gru remixicon: atlas ready (%s, %d glyphs, base %d)", remixIconTTFPath, len(cps), remixIcons.atlasSize)
	}
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
	if PreferIconSharp1x() {
		return r.loadStrikesLocked(cps)
	}
	return r.loadLegacyAtlasLocked(cps)
}

func (r *remixIconRegistry) loadLegacyAtlasLocked(cps []int32) bool {
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
		if cp >= 0x20 && cp < 0x7F {
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
	r.unloadStrikesLocked()
	rl.GenTextureMipmaps(&font.Texture)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	r.font = font
	r.loadedCP = loaded
	r.failed = false
	return true
}

func (r *remixIconRegistry) unloadStrikesLocked() {
	for _, s := range r.strikes {
		if s != nil && s.font.Texture.ID != 0 {
			rl.UnloadFont(s.font)
		}
	}
	r.strikes = nil
}

func (r *remixIconRegistry) loadStrikesLocked(cps []int32) bool {
	fileData, err := ReadAssetFile(remixIconTTFPath)
	if err != nil {
		r.failed = true
		rl.TraceLog(rl.LogWarning, "Gru remixicon: ReadAssetFile failed for %s", remixIconTTFPath)
		return false
	}
	loaded := make(map[int32]struct{}, len(cps))
	for _, cp := range cps {
		loaded[cp] = struct{}{}
	}
	// Drop legacy mega-atlas — 512→24 point downsample is what made icons jagged.
	if r.font.Texture.ID != 0 {
		rl.UnloadFont(r.font)
		r.font = rl.Font{}
	}
	if r.strikes == nil {
		r.strikes = make(map[int32]*remixIconStrike)
	}
	// Rebuild every existing strike + ensure warm sizes exist.
	sizes := make(map[int32]struct{})
	for _, sz := range iconStrikeWarmSizes {
		sizes[sz] = struct{}{}
	}
	for sz := range r.strikes {
		sizes[sz] = struct{}{}
	}
	okAny := false
	for sz := range sizes {
		if r.rebuildStrikeLocked(fileData, sz, cps) {
			okAny = true
		}
	}
	if !okAny {
		r.failed = true
		rl.TraceLog(rl.LogWarning, "Gru remixicon: no strike glyphs packed")
		return false
	}
	r.loadedCP = loaded
	r.failed = false
	return true
}

func (r *remixIconRegistry) rebuildStrikeLocked(fileData []byte, basePx int32, cps []int32) bool {
	if basePx < 14 {
		basePx = 14
	}
	font := loadFontFromMemory(fileData, basePx, cps)
	if font.Texture.ID == 0 {
		rl.TraceLog(rl.LogWarning, "Gru remixicon: strike %dpx failed", basePx)
		return false
	}
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
		}
	}
	if okCount == 0 {
		rl.UnloadFont(font)
		return false
	}
	if prev, ok := r.strikes[basePx]; ok && prev != nil && prev.font.Texture.ID != 0 {
		rl.UnloadFont(prev.font)
	}
	rl.SetTextureFilter(font.Texture, rl.FilterPoint)
	r.strikes[basePx] = &remixIconStrike{font: font, loadedCP: loaded, basePx: basePx}
	return true
}

func (r *remixIconRegistry) ensureStrikeLocked(basePx int32) *remixIconStrike {
	if basePx < 14 {
		basePx = 14
	}
	if s, ok := r.strikes[basePx]; ok && s != nil && s.font.Texture.ID != 0 {
		return s
	}
	fileData, err := ReadAssetFile(remixIconTTFPath)
	if err != nil {
		return nil
	}
	cps := make([]int32, 0, len(r.loadedCP))
	for cp := range r.loadedCP {
		cps = append(cps, cp)
	}
	sort.Slice(cps, func(i, j int) bool { return cps[i] < cps[j] })
	cps = filterIconFontCodepoints(cps)
	if !r.rebuildStrikeLocked(fileData, basePx, cps) {
		return nil
	}
	return r.strikes[basePx]
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
	return remixEnsureGlyphs([]string{phosphorName}, weight)
}

// remixEnsureGlyphs packs many Phosphor/Remix names in one atlas rebuild.
// Prefer this over N× remixEnsureGlyph — each rebuild reloads the full TTF subset.
func remixEnsureGlyphs(names []string, weight PhosphorWeight) bool {
	if len(names) == 0 {
		return true
	}
	remixIcons.mu.Lock()
	defer remixIcons.mu.Unlock()
	if remixIcons.failed || len(remixIcons.cssClasses) == 0 {
		return false
	}
	need := make([]int32, 0, len(remixIcons.loadedCP)+len(names))
	for k := range remixIcons.loadedCP {
		need = append(need, k)
	}
	added := false
	for _, name := range names {
		cp, ok := remixCodepointFor(name, weight, remixIcons.cssClasses)
		if !ok {
			continue
		}
		icp := int32(cp)
		if _, has := remixIcons.loadedCP[icp]; has && remixIcons.ready {
			continue
		}
		need = append(need, icp)
		added = true
	}
	if !added {
		return true
	}
	iconWarmPending.Store(true)
	defer iconWarmPending.Store(false)
	sort.Slice(need, func(i, j int) bool { return need[i] < need[j] })
	need = filterIconFontCodepoints(need)
	if !remixIcons.loadSubsetLocked(need) {
		return false
	}
	remixIcons.ready = true
	return true
}

func remixDrawIcon(dst rl.Rectangle, phosphorName string, weight PhosphorWeight, tint rl.Color, strokeScale float32) bool {
	if (weight == PhosphorRegular || weight == "") && PreferIconSharp1x() {
		// Line icons look anemic at true 1×; prefer fill silhouettes when available.
		weight = PhosphorFill
	}
	if !remixEnsureGlyph(phosphorName, weight) {
		if weight == PhosphorFill {
			weight = PhosphorRegular
			if !remixEnsureGlyph(phosphorName, weight) {
				return false
			}
		} else {
			return false
		}
	}
	remixIcons.mu.Lock()
	cp, ok := remixCodepointFor(phosphorName, weight, remixIcons.cssClasses)
	ready := remixIcons.ready
	var font rl.Font
	var loaded map[int32]struct{}
	drawSize := float32(0)
	if ok && ready && PreferIconSharp1x() {
		dstSnap := snapPhosphorRect(dst)
		sz := phosphorFontDrawSize(dstSnap)
		if strokeScale > 0 && strokeScale < 1 {
			sz = quantizeIconDrawSize(sz * strokeScale)
		}
		strike := remixIcons.ensureStrikeLocked(int32(sz + 0.5))
		if strike != nil {
			font = strike.font
			loaded = strike.loadedCP
			drawSize = float32(strike.basePx)
		}
	}
	if font.Texture.ID == 0 {
		font = remixIcons.font
		loaded = remixIcons.loadedCP
	}
	remixIcons.mu.Unlock()
	if !ok || !ready || font.Texture.ID == 0 {
		return false
	}
	return remixDrawCodepoint(font, loaded, dst, cp, tint, strokeScale, drawSize)
}

// remixDrawCodepoint draws one icon glyph. When exactSize>0 (sticky strike path),
// DrawTextEx uses that size so scale≈1 against the strike atlas BaseSize.
func remixDrawCodepoint(font rl.Font, loaded map[int32]struct{}, dst rl.Rectangle, cp rune, tint rl.Color, strokeScale, exactSize float32) bool {
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
	size := exactSize
	if size < minRenderPx {
		size = phosphorFontDrawSize(dst) * strokeScale
	}
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
		if PreferIconSharp1x() && len(remixIcons.strikes) > 0 {
			return "remixicon.ttf+strikes"
		}
		if PreferIconSharp1x() {
			return "remixicon.ttf+point"
		}
		return "remixicon.ttf"
	}
	if remixIcons.failed {
		return "remixicon(failed)"
	}
	return "remixicon(off)"
}

// SyncRemixIconFilter retunes Remix atlases when SSAA crosses the 1× boundary.
// Entering 1× migrates the mega-atlas into sticky strikes (text-strike analog).
func SyncRemixIconFilter() {
	remixIcons.mu.Lock()
	defer remixIcons.mu.Unlock()
	if remixIcons.failed || !remixIcons.ready {
		return
	}
	if PreferIconSharp1x() {
		if len(remixIcons.strikes) == 0 {
			cps := make([]int32, 0, len(remixIcons.loadedCP))
			for cp := range remixIcons.loadedCP {
				cps = append(cps, cp)
			}
			sort.Slice(cps, func(i, j int) bool { return cps[i] < cps[j] })
			cps = filterIconFontCodepoints(cps)
			if len(cps) > 0 {
				_ = remixIcons.loadStrikesLocked(cps)
			}
		}
		for _, s := range remixIcons.strikes {
			if s != nil && s.font.Texture.ID != 0 {
				rl.SetTextureFilter(s.font.Texture, rl.FilterPoint)
			}
		}
		return
	}
	if remixIcons.font.Texture.ID != 0 {
		rl.SetTextureFilter(remixIcons.font.Texture, rl.FilterBilinear)
	}
}

func unloadRemixIcons() {
	remixIcons.mu.Lock()
	defer remixIcons.mu.Unlock()
	if remixIcons.font.Texture.ID != 0 {
		rl.UnloadFont(remixIcons.font)
	}
	remixIcons.unloadStrikesLocked()
	remixIcons.font = rl.Font{}
	remixIcons.loadedCP = nil
	remixIcons.cssClasses = nil
	remixIcons.ready = false
	remixIcons.failed = false
}
