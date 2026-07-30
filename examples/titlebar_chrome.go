package examples

import (
	"github.com/ledocorp/gru/ui"
)

// TitleBarChrome optionally overrides custom title bar content.
// OS window title (taskbar tooltip) may still be set separately via rl.SetWindowTitle.
type TitleBarChrome interface {
	TitleBarShowTitle() bool
	TitleBarShowAppIcon() bool
	TitleBarCenterText() string
}

// ConfigureTitleBar applies scene chrome preferences to the borderless title bar.
func ConfigureTitleBar(tb *ui.TitleBar, scene Scene) {
	if tb == nil || scene == nil {
		return
	}
	tb.ShowTitle = true
	tb.ShowAppIcon = false
	tb.Title = scene.Title()
	if c, ok := scene.(TitleBarChrome); ok {
		tb.ShowTitle = c.TitleBarShowTitle()
		tb.ShowAppIcon = c.TitleBarShowAppIcon()
		tb.Title = c.TitleBarCenterText()
	}
}
