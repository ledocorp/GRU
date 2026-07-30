// Package ui — Remix Icon font atlas (Gru default icon set).
//
// [IconRegistry.Draw] uses Remix Icon only (assets/fonts/remixicon.ttf + remixicon.css).
// Legacy Phosphor PNG/SVG assets are not used. Historical Phosphor* names remain as
// deprecated aliases in phosphor_compat.go — prefer Icon* / Icons / InitIcons / DrawIcon.
//
// # LLM Prompt Template
//
//	ui.InitIcons(128) // once after rl.InitWindow, before scenes
//	btn.SetIcon(ui.IconHouse, ui.IconRegular)
//	ui.Icons.Draw(dst, ui.IconBell, ui.IconFill, tint)
package ui

import (
	"fmt"
	"sync"
	"sync/atomic"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// IconsEagerWarm, when true before InitIcons / InitDisplayAwareAtlases, loads
// WarmGlyphNames into the startup Remix atlas (Studio / rich demos). Lean hosts
// leave this false — title-bar glyphs only; other icons load on first Draw.
var IconsEagerWarm bool

// iconWarmPending is set while a Remix atlas rebuild runs on the main thread.
var iconWarmPending atomic.Bool

// IconWarmPending reports an in-flight icon atlas rebuild (skip/placeholder draw).
func IconWarmPending() bool { return iconWarmPending.Load() }

// IconWeight selects Remix line vs fill (and legacy weight labels).
type IconWeight string

const (
	IconThin    IconWeight = "thin"
	IconLight   IconWeight = "light"
	IconRegular IconWeight = "regular"
	IconBold    IconWeight = "bold"
	IconFill    IconWeight = "fill"
	IconDuotone IconWeight = "duotone"
)

// Common icon names (kebab-case). Mapped to Remix glyphs in remix_icon.go.
const (
	IconHouse             = "house"
	IconMagnifyingGlass   = "magnifying-glass"
	IconBell              = "bell"
	IconGear              = "gear"
	IconUser              = "user"
	IconUsers             = "users"
	IconEnvelope          = "envelope"
	IconTray              = "tray"
	IconList              = "list"
	IconMarkdownLine      = "markdown-line"
	IconMarkdownFill      = "markdown-fill"
	IconCodeView          = "code-view"
	IconCodeBlock         = "code-block"
	IconHome2             = "home-2"
	IconEditBox           = "edit-box"
	IconBookRead          = "book-read"
	IconSearch            = "search"
	IconCodeBox           = "code-box"
	IconArrowDropDown     = "arrow-drop-down"
	IconArrowDropUp       = "arrow-drop-up"
	IconInfoI             = "info-i"
	IconTable             = "table"
	IconPlus              = "plus"
	IconMinus             = "minus"
	IconSquare            = "square"
	IconResize            = "resize"
	IconCopy              = "copy"
	IconX                 = "x"
	IconXCircle           = "x-circle"
	IconCheck             = "check"
	IconCheckbox          = "checkbox-fill"
	IconCheckboxBlank     = "checkbox-blank"
	IconArrowGoBack       = "arrow-go-back"
	IconArrowGoForward    = "arrow-go-forward"
	IconScissorsCut       = "scissors-cut"
	IconClipboard         = "clipboard"
	IconFindReplace       = "find-replace"
	IconZoomIn            = "zoom-in"
	IconZoomOut           = "zoom-out"
	IconFileClose         = "file-close"
	IconCaretLeft         = "caret-left"
	IconCaretRight        = "caret-right"
	IconCaretDown         = "caret-down"
	IconCaretUp           = "caret-up"
	IconCaretCircleDown   = "caret-circle-down"
	IconCaretCircleUp     = "caret-circle-up"
	IconCaretCircleLeft   = "caret-circle-left"
	IconCaretCircleRight  = "caret-circle-right"
	IconDotsThree         = "dots-three"
	IconDotsThreeVertical = "dots-three-vertical"
	IconStar              = "star"
	IconHeart             = "heart"
	IconCalendar          = "calendar"
	IconCalendarBlank     = "calendar-blank"
	IconFunnel            = "funnel"
	IconPencilSimple      = "pencil-simple"
	IconTrash             = "trash"
	IconUpload            = "upload"
	IconDownload          = "download"
	IconFolder            = "folder"
	IconFolderOpen        = "folder-open"
	IconRestart           = "restart"
	IconDatabase          = "database"
	IconSave2             = "save-2"
	IconSettings4         = "settings-4"
	IconSettings3         = "settings-3"
	IconListSettings      = "list-settings"
	IconFile              = "file"
	IconWifiHigh          = "wifi-high"
	IconMoon              = "moon"
	IconSun               = "sun"
	IconTextBold          = "bold"
	IconTextItalic        = "italic"
	IconTextUnderline     = "underline"
	IconTextWrap          = "text-wrap"
	IconAlignLeft         = "align-left"
	IconAlignCenter       = "align-center"
	IconAlignRight        = "align-right"
	IconImage             = "image"
	IconLink              = "link"
)

// Icons is the process-wide default Remix icon registry.
var Icons = NewIconRegistry("")

// SetDefaultIconRegistry replaces the process-wide icon registry used by NewIcon,
// IconButton.SetIcon, FAB, BottomNav, and Icons.Draw call sites.
// Call before InitIcons when switching to a custom pack with the same folder layout.
func SetDefaultIconRegistry(r *IconRegistry) {
	if r != nil {
		Icons = r
		Phosphor = r
	}
}

// IconRegistry resolves and draws icons via the Remix TTF atlas.
type IconRegistry struct {
	Root string

	atlasSize int32 // icon font atlas px (InitIcons)

	fontMu sync.Mutex
	fonts  map[IconWeight]*phosphorFontFace
}

// NewIconRegistry creates a registry rooted at root.
func NewIconRegistry(root string) *IconRegistry {
	return &IconRegistry{Root: root}
}

// InitIcons configures the Remix icon-font atlas. Call once after rl.InitWindow().
// Default warm set is title-bar chrome only. Set IconsEagerWarm or call
// PreloadShellIcons for the larger WarmGlyphNames set.
func InitIcons(atlasSize int32) {
	if atlasSize < 256 {
		atlasSize = 256
	}
	Icons.atlasSize = atlasSize
	Phosphor = Icons // keep deprecated alias pointing at the live registry
	initRemixIconAtlas(atlasSize)
	// Always pack title-bar controls after atlas init (maximize was flaky in lean sets).
	for _, cp := range remixTitleBarCPs {
		ensureTitleBarCodepoint(rune(cp))
	}
	if IconsEagerWarm {
		for _, name := range WarmGlyphNames {
			remixEnsureGlyph(name, IconRegular)
			remixEnsureGlyph(name, IconFill)
		}
	}
}

// Available reports whether the icon can be drawn via Remix.
func (r *IconRegistry) Available(name string, weight IconWeight) bool {
	return remixHasGlyph(name, weight)
}

// EnsureLoaded warms the Remix font subset. Safe from Draw / preload.
func (r *IconRegistry) EnsureLoaded(name string, weight IconWeight) bool {
	if !remixHasGlyph(name, weight) {
		return false
	}
	return remixEnsureGlyph(name, weight)
}

// DrawChrome renders shell chrome icons (Remix font).
func (r *IconRegistry) DrawChrome(dst rl.Rectangle, name string, weight IconWeight, tint rl.Color) bool {
	return r.Draw(dst, name, weight, tint)
}

// EnsureChromeLoaded warms Remix glyphs for chrome icons.
func (r *IconRegistry) EnsureChromeLoaded(name string, weight IconWeight) bool {
	return r.EnsureLoaded(name, weight)
}

// Draw renders the icon into dst. Returns false if missing.
func (r *IconRegistry) Draw(dst rl.Rectangle, name string, weight IconWeight, tint rl.Color) bool {
	if dst.Width <= 0 || dst.Height <= 0 {
		return false
	}
	return r.drawFont(dst, name, weight, tint)
}

// Measure returns the rendered width/height (square, fits in size).
func (r *IconRegistry) Measure(name string, weight IconWeight, size float32) (w, h float32) {
	return size, size
}

// PreloadAsync warms Remix glyphs on the main thread via doc.QueueMain.
func (r *IconRegistry) PreloadAsync(doc *Document, names []string, weight IconWeight, done func()) {
	if doc == nil || len(names) == 0 {
		if done != nil {
			done()
		}
		return
	}
	doc.QueueMain(func() {
		for _, n := range names {
			r.EnsureLoaded(n, weight)
		}
		if done != nil {
			done()
		}
	})
}

// PreloadShellIcons warms Remix glyphs used by app chrome and demos.
// Prefer calling after mount via QueueMain so startup stays lean; Studio may
// call this once after InitDisplayAwareAtlases when IconsEagerWarm was false.
func (r *IconRegistry) PreloadShellIcons(doc *Document) {
	warm := func() {
		for _, w := range []IconWeight{IconRegular, IconFill} {
			for _, n := range WarmGlyphNames {
				remixEnsureGlyph(n, w)
			}
		}
		r.EnsureLoaded(IconPlus, IconRegular)
		if doc != nil {
			doc.InvalidatePaint()
		}
	}
	if doc != nil {
		doc.QueueMain(warm)
		return
	}
	warm()
}

// PreloadShellIconsNow warms WarmGlyphNames synchronously (main thread only).
func (r *IconRegistry) PreloadShellIconsNow() {
	r.PreloadShellIcons(nil)
}

// UnloadAll frees the Remix icon font.
func (r *IconRegistry) UnloadAll() {
	unloadRemixIcons()
	r.unloadFonts()
}

// MissingHint describes how to fix a missing icon.
func (r *IconRegistry) MissingHint(name string, weight IconWeight) string {
	if weight == "" {
		weight = IconRegular
	}
	remixClass := remixClassBase(name)
	return fmt.Sprintf("icon: no Remix mapping for %q (%s) — add phosphorToRemix in ui/remix_icon.go or ri-%s in remixicon.css",
		name, weight, remixClass)
}

// snapPhosphorRect aligns icon draws to whole pixels to reduce GPU softness.
func snapPhosphorRect(r rl.Rectangle) rl.Rectangle {
	return rl.NewRectangle(
		float32(int32(r.X+0.5)),
		float32(int32(r.Y+0.5)),
		float32(int32(r.Width+0.5)),
		float32(int32(r.Height+0.5)),
	)
}

// phosphorIconSize picks a draw size for chrome icons inside a square slot.
func phosphorIconSize(slot float32, styleFontSize int32) float32 {
	size := slot * 0.62
	if size < 24 {
		size = 24
	}
	if styleFontSize > 0 {
		if fs := float32(styleFontSize); fs > size {
			size = fs
		}
	}
	if size > slot-4 {
		size = slot - 4
	}
	return size
}

// DrawIcon renders name into dst via Remix, plus a Unicode fallback for a few chrome glyphs.
func DrawIcon(dst rl.Rectangle, name string, weight IconWeight, tint rl.Color) bool {
	if dst.Width <= 0 || dst.Height <= 0 {
		return false
	}
	if weight == "" {
		weight = IconRegular
	}
	Icons.EnsureLoaded(name, weight)
	if Icons.Draw(dst, name, weight, tint) {
		return true
	}
	fallback := phosphorTextFallback(name)
	if fallback == "" {
		return false
	}
	iconSize := phosphorFontDrawSize(dst)
	fs := Style{FontSize: int32(iconSize), TextColor: tint}
	tw := measureTextS(fallback, fs)
	drawTextS(fallback, int32(dst.X+(dst.Width-float32(tw))/2), TextPosY(dst, fs), fs)
	return true
}

func phosphorTextFallback(name string) string {
	switch name {
	case IconCaretLeft, IconCaretCircleLeft:
		return "←"
	case IconCaretRight, IconCaretCircleRight:
		return "→"
	case IconCaretUp, IconCaretCircleUp:
		return "↑"
	case IconCaretDown, IconCaretCircleDown:
		return "↓"
	case IconX, IconXCircle:
		return "×"
	case IconCheck:
		return "✓"
	case IconMinus:
		return "−"
	default:
		return ""
	}
}

// phosphorIconDrawSize fits a glyph inside an explicit icon slot (e.g. [Icon] widget).
func phosphorIconDrawSize(slot float32, styleFontSize int32) float32 {
	if slot <= 0 {
		return minRenderPx
	}
	size := slot - 2
	if size < minRenderPx {
		size = minRenderPx
	}
	if styleFontSize > 0 {
		if fs := float32(styleFontSize); fs > 0 && fs < size {
			size = fs
		}
	}
	return float32(int32(size + 0.5))
}
