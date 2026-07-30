// Package examples — shared Remix icon wiring for Go-first demos.
package examples

import "github.com/ledocorp/gru/ui"

// PreloadSceneIcons warms Remix icon-font glyphs used by shell and form demos.
func PreloadSceneIcons(doc *ui.Document) {
	ui.Icons.PreloadShellIcons(doc)
	for _, n := range ui.WarmGlyphNames {
		ui.Icons.EnsureLoaded(n, ui.IconRegular)
	}
	for _, n := range batch9ToolbarGlyphs {
		ui.Icons.EnsureLoaded(n, ui.IconRegular)
	}
	ui.Icons.EnsureLoaded(ui.IconPlus, ui.IconFill)
	ui.Icons.EnsureLoaded(ui.IconStar, ui.IconFill)
	ui.Icons.EnsureLoaded(ui.IconXCircle, ui.IconRegular)
	ui.Icons.EnsureLoaded(ui.IconX, ui.IconRegular)
}

// Deprecated: use PreloadSceneIcons.
func PreloadScenePhosphor(doc *ui.Document) { PreloadSceneIcons(doc) }

// batch9ToolbarGlyphs — Batch 9 toolbar/ribbon icons not in WarmGlyphNames.
var batch9ToolbarGlyphs = []string{
	ui.IconFolder,
	ui.IconFolderOpen,
	ui.IconDownload,
	ui.IconSave2,
	ui.IconFile,
	ui.IconTextBold,
	ui.IconTextItalic,
	ui.IconTextUnderline,
	ui.IconAlignLeft,
	ui.IconAlignCenter,
	ui.IconAlignRight,
	ui.IconImage,
	ui.IconLink,
}

// AppBarBackButton returns a leading AppBar control (caret-left).
func AppBarBackButton(id string) *ui.IconButton {
	btn := ui.NewIconButton(id, "", "", 0, 0, 44, 44)
	btn.SetStyle("appbar-icon")
	btn.SetIcon(ui.IconCaretLeft, ui.IconRegular)
	return btn
}

// AppBarLeadingMenu returns a leading navigation control (list / “hamburger”).
func AppBarLeadingMenu(id string) *ui.IconButton {
	btn := ui.NewIconButton(id, "☰", "", 0, 0, 44, 44)
	btn.SetStyle("appbar-icon")
	btn.SetIcon(ui.IconList, ui.IconRegular)
	btn.Symbol = "☰" // fallback if glyph not ready
	ui.Icons.EnsureLoaded(ui.IconList, ui.IconRegular)
	return btn
}

// AppBarMenuButton returns a trailing overflow control (dots-three).
func AppBarMenuButton(id string) *ui.IconButton {
	btn := ui.NewIconButton(id, "", "", 0, 0, 44, 44)
	btn.SetStyle("appbar-icon")
	btn.SetIcon(ui.IconDotsThree, ui.IconRegular)
	return btn
}
