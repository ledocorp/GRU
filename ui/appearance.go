package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	AppearanceLight  = "light"
	AppearanceDark   = "dark"
	AppearanceSystem = "system"
)

var (
	lightTheme Theme
	darkTheme  Theme

	currentAppearanceMode = AppearanceSystem
)

func init() {
	lightTheme = cloneTheme(CurrentTheme)
	darkTheme = cloneTheme(CurrentTheme)
	applyDarkAppearancePatches(darkTheme)
}

func cloneTheme(src Theme) Theme {
	out := make(Theme, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func applyDarkAppearancePatches(t Theme) {
	patch := func(name string, bg, fg rl.Color) {
		s, ok := t[name]
		if !ok {
			return
		}
		if bg.A > 0 {
			s.BackgroundColor = bg
		}
		if fg.A > 0 {
			s.TextColor = fg
		}
		t[name] = s
	}
	patchBorder := func(name string, border rl.Color) {
		s, ok := t[name]
		if !ok {
			return
		}
		s.BorderColor = border
		t[name] = s
	}
	patchTextOnly := func(name string, fg rl.Color) {
		s, ok := t[name]
		if !ok {
			return
		}
		s.TextColor = fg
		t[name] = s
	}

	surface := rl.NewColor(22, 24, 32, 255)
	surfaceRaised := rl.NewColor(32, 34, 44, 255)
	surfaceMuted := rl.NewColor(28, 30, 40, 255)
	surfaceHover := rl.NewColor(48, 50, 64, 255)
	textPrimary := rl.NewColor(232, 234, 244, 255)
	textMuted := rl.NewColor(160, 164, 184, 255)
	textSubtle := rl.NewColor(120, 126, 148, 255)
	border := rl.NewColor(58, 62, 78, 255)
	borderSoft := rl.NewColor(48, 52, 68, 255)
	accentTab := rl.NewColor(155, 163, 228, 255)       // active ribbon tab — softer than indigo-400
	accentDropdown := rl.NewColor(202, 190, 255, 255) // selected dropdown option text
	accent := rl.NewColor(129, 140, 248, 255)
	transparent := rl.NewColor(0, 0, 0, 0)

	// Core surfaces
	for _, name := range []string{
		"default", "text-editor", "editor-scroll", "input", "dropdown", "combobox",
		"settings-body", "settings-flat-band", "settings-scroll", "surface-muted",
		"list-pane", "list-pane-scroll", "list-pane-list", "list-pane-header",
		"panel", "card", "spinbox", "searchbar",
	} {
		bg := surface
		switch name {
		case "surface-muted", "list-pane-scroll", "settings-scroll", "list-pane", "list-pane-list":
			bg = surfaceMuted
		case "panel", "card", "settings-flat-band":
			bg = surfaceRaised
		}
		patch(name, bg, textPrimary)
	}

	// Chrome strips
	patch("menubar", surfaceRaised, textPrimary)
	patch("menubar-hover", surfaceHover, textPrimary)
	patch("statusbar", surfaceRaised, textMuted)
	patchBorder("statusbar", borderSoft)
	patch("statusbar-label", transparent, textMuted)
	patch("statusbar-pipe", transparent, rl.NewColor(160, 164, 180, 255))

	// Ribbon / toolbar
	patch("toolbar-ribbon", surfaceRaised, textPrimary)
	patch("toolbar-ribbon-tab", transparent, textSubtle)
	patch("toolbar-ribbon-tab-active", transparent, accentTab)
	patchTextOnly("dropdown-selected", accentDropdown)
	patch("toolbar-cell", transparent, textPrimary)
	patch("toolbar-btn", surfaceRaised, textPrimary)
	patch("toolbar-menu", surfaceRaised, textPrimary)
	patch("toolbar-toggle-label", transparent, textMuted)
	patch("bottomnav", surfaceRaised, textMuted)
	patchBorder("bottomnav", borderSoft)
	patchTextOnly("appbar-icon", textPrimary)

	// List tiles (open notes, settings rows)
	for _, name := range []string{"list-tile", "list-tile-pane", "list-tile-bordered"} {
		patch(name, surfaceRaised, textPrimary)
		patchBorder(name, borderSoft)
	}
	patchTextOnly("list-tile-subtitle", textMuted)
	patch("list-pane-header-title", transparent, textPrimary)

	// Settings typography
	patch("form-value", transparent, textMuted)
	patch("form-label", transparent, textMuted)
	patch("settings-page-title", transparent, textPrimary)
	patch("settings-section", transparent, textPrimary)

	for _, name := range []string{
		"menubar", "statusbar", "panel", "card", "list-pane", "toolbar-ribbon",
		"toolbar-btn", "toolbar-menu", "combobox", "dropdown",
	} {
		patchBorder(name, border)
	}

	// Context menus and popup rows
	patch("contextmenu", surfaceRaised, textPrimary)
	patchBorder("contextmenu", border)
	patch("contextmenu-item", transparent, textPrimary)
	patch("contextmenu-hover", surfaceHover, textPrimary)
	patchTextOnly("contextmenu-shortcut", textMuted)
	patch("contextmenu-divider", borderSoft, transparent)

	// Markdown preview lane
	previewText := textPrimary
	previewMuted := textMuted
	previewCodeBg := rl.NewColor(38, 40, 52, 255)
	previewCodeFg := rl.NewColor(196, 181, 253, 255)
	previewBlockBg := rl.NewColor(36, 38, 50, 255)
	previewSurface := surfaceRaised

	for _, name := range []string{
		"richtext-preview", "richtext-blockquote", "richtext-callout",
		"richtext-footnote", "preview-image-caption",
	} {
		patchTextOnly(name, previewText)
	}
	patchTextOnly("richtext-preview-bold", previewText)
	patchTextOnly("richtext-preview-italic", previewMuted)
	patchTextOnly("richtext-strike", textSubtle)
	patchTextOnly("richtext-list-marker", accent)
	patch("richtext-code", previewCodeBg, previewCodeFg)
	patch("richtext-math", previewCodeBg, previewCodeFg)
	patchTextOnly("richtext-math-inline", previewCodeFg)
	patchTextOnly("richtext-math-display", previewCodeFg)
	patch("richtext-code-block", previewBlockBg, previewText)
	patch("preview-blockquote", previewBlockBg, previewText)
	patch("preview-footnote-shell", previewBlockBg, previewText)
	patch("preview-image-wrap", previewSurface, previewText)
	patchBorder("preview-image-wrap", borderSoft)
	patchBorder("preview-image-frame", borderSoft)
	patchTextOnly("richtext-footnote-ref", accent)
	patchTextOnly("richtext-footnote-back", accent)
	patchTextOnly("preview-footnote-back-icon", accent)
	patchTextOnly("preview-task-icon", previewText)

	// Title bar + splitters + markdown tables
	patch("titlebar", surfaceRaised, textPrimary)
	patchBorder("titlebar", borderSoft)
	patch("titlebar-hover", rl.NewColor(255, 255, 255, 18), transparent)
	patchBorder("splitview-splitter", borderSoft)
	patch("table-header-row", surfaceHover, textPrimary)
	patch("table-body-row", surfaceRaised, textPrimary)
	patchBorder("table-body-row", borderSoft)

	// Modals and search fields
	patch("modal", surfaceRaised, textPrimary)
	patchBorder("modal", border)
	patchTextOnly("modal-title", textMuted)
	patchTextOnly("modal-body", textMuted)
	patch("modal-input", surfaceRaised, textPrimary)
	patchBorder("modal-input", borderSoft)
	patch("input", surfaceRaised, textPrimary)
	patchBorder("input", borderSoft)
	patch("colorwell", transparent, textPrimary)
	patch("colorpicker", surfaceRaised, textPrimary)
	patchBorder("colorpicker", borderSoft)

	previewCard := rl.NewColor(36, 38, 50, 255)
	patch("card-code", previewCard, textPrimary)
	patchBorder("card-code", rl.NewColor(64, 68, 84, 255))
	patch("card-blockquote", previewCard, textPrimary)
	patchBorder("card-blockquote", rl.NewColor(64, 68, 84, 255))
	patchTextOnly("richtext-blockquote-italic", textPrimary)
	patch("list-flush", previewCard, textPrimary)
}

// AppearanceMode returns the normalized mode: light, dark, or system.
func AppearanceMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case AppearanceDark:
		return AppearanceDark
	case AppearanceSystem:
		return AppearanceSystem
	default:
		return AppearanceLight
	}
}

// EffectiveAppearance resolves system to light or dark for the current platform.
func EffectiveAppearance(mode string) string {
	mode = AppearanceMode(mode)
	if mode != AppearanceSystem {
		return mode
	}
	if systemPrefersDarkAppearance() {
		return AppearanceDark
	}
	return AppearanceLight
}

// SetAppearance applies light, dark, or system (resolved) to CurrentTheme.
// Call Document.UnloadCache after switching so SSAA picks up new colors.
func SetAppearance(mode string) {
	currentAppearanceMode = AppearanceMode(mode)
	src := lightTheme
	dark := EffectiveAppearance(mode) == AppearanceDark
	if dark {
		src = darkTheme
	}
	for k, v := range src {
		CurrentTheme[k] = v
	}
	if dark {
		SetViewportScrollbarColors(
			rl.NewColor(38, 40, 50, 255),
			rl.NewColor(96, 104, 122, 255),
		)
		setListPopupColors(
			rl.NewColor(48, 50, 64, 255),
			rl.NewColor(55, 58, 82, 255),
			rl.NewColor(255, 255, 255, 12),
		)
		setToggleTrackColors(
			rl.NewColor(88, 92, 104, 255),
			rl.NewColor(79, 70, 229, 255),
		)
		setEditorAccentColors(
			rl.NewColor(255, 254, 250, 255),
			rl.NewColor(255, 82, 82, 255),
		)
	} else {
		SetViewportScrollbarColors(viewportScrollTrackLight, viewportScrollThumbLight)
		setListPopupColors(
			rl.NewColor(237, 239, 254, 255),
			rl.NewColor(237, 239, 254, 255),
			rl.NewColor(0, 0, 0, 10),
		)
		setToggleTrackColors(
			rl.NewColor(210, 213, 222, 255),
			rl.NewColor(79, 70, 229, 255),
		)
		setEditorAccentColors(
			rl.NewColor(40, 42, 54, 255),
			rl.NewColor(239, 48, 48, 255),
		)
	}
	SyncThemeV2MarkdownCards(dark)
	SyncThemeV2FormControls()
	bumpThemeRevision()
}

// ThemeIsDark reports whether the active appearance resolves to dark chrome.
func ThemeIsDark() bool {
	return EffectiveAppearance(currentAppearanceMode) == AppearanceDark
}

func setListPopupColors(hover, selected, tileHover rl.Color) {
	listRowHoverBg = hover
	listOptionSelectedBg = selected
	listTileHoverOverlay = tileHover
}

func setToggleTrackColors(off, on rl.Color) {
	toggleTrackOff = off
	toggleTrackOn = on
}

func setEditorAccentColors(caret, spell rl.Color) {
	textEditorCaretColor = caret
	spellUnderlineColor = spell
}

func bumpThemeRevision() {
	themeRevisionV2++
	if currentThemeV2 != nil {
		currentThemeV2.Revision = themeRevisionV2
	}
}
