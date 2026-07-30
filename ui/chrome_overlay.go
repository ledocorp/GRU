// Chrome overlay drawing — Gru Studio footer, side panel, and F12 inspector.
//
// Text and vector chrome render through the supersampled overlay pass when SSAA
// is active so labels stay as sharp as in-tree widgets (1× screen draws look
// soft or drop missing glyphs).
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ChromeMinFontPx is the smallest effective size for launcher chrome (footer,
// studio panel, F12 inspector). Matches the library-wide 14px readability floor.
const ChromeMinFontPx int32 = 14

// DrawLauncherChromeOverlays renders devtools chrome (studio panel, inspector)
// through the overlay SSAA path when available.
func DrawLauncherChromeOverlays(winW, winH int32, draw func()) {
	if draw == nil {
		return
	}
	if SupersamplingActive() && OverlayTargetDrawable() {
		BeginOverlaySuperFrame()
		draw()
		EndOverlaySuperFrame()
		BlitOverlayToScreen(winW, winH)
		return
	}
	draw()
}

// ChromeTitleStyle is semibold launcher chrome (panel title, inspector header).
func ChromeTitleStyle() Style {
	s := GetThemeStyle("form-value")
	s.Bold = true
	s.MinFontSize = ChromeMinFontPx
	s.TextColor = rl.NewColor(235, 237, 245, 255)
	return s
}

// ChromeBodyStyle is default launcher chrome body text.
func ChromeBodyStyle() Style {
	s := GetThemeStyle("form-label")
	s.MinFontSize = ChromeMinFontPx
	s.TextColor = rl.NewColor(220, 222, 235, 255)
	return s
}

// ChromeFooterHintStyle is the scene index / title in the slim footer strip.
func ChromeFooterHintStyle() Style {
	s := GetThemeStyle("form-label")
	s.MinFontSize = ChromeMinFontPx
	s.NoCaptionBold = true
	s.TextColor = rl.NewColor(200, 202, 215, 255)
	return s
}

// ChromeFooterButtonStyle is a compact label for footer chrome (Menu, Directory).
func ChromeFooterButtonStyle() Style {
	s := GetThemeStyle("form-label")
	s.Bold = true
	s.MinFontSize = ChromeMinFontPx
	s.TextColor = rl.NewColor(235, 237, 245, 255)
	return s
}

// ChromeMutedStyle is secondary launcher chrome text.
func ChromeMutedStyle() Style {
	s := GetThemeStyle("form-label")
	s.MinFontSize = ChromeMinFontPx
	s.TextColor = rl.NewColor(150, 155, 178, 255)
	return s
}

// ChromeDimStyle is tertiary stats / hints on inspector chrome.
func ChromeDimStyle() Style {
	s := GetThemeStyle("form-label")
	s.MinFontSize = ChromeMinFontPx
	s.TextColor = rl.NewColor(130, 135, 160, 255)
	return s
}

// ChromeInspectorTreeStyle is widget-tree rows in the F12 inspector.
func ChromeInspectorTreeStyle() Style {
	s := GetThemeStyle("form-value")
	s.MinFontSize = ChromeMinFontPx
	return s
}

// DrawChromeText draws a single line using the same path as widgets (drawTextS).
func DrawChromeText(text string, x, y float32, s Style) {
	if text == "" {
		return
	}
	fs := EffectiveFontSize(s)
	drawTextF(text, x, y, fs, s.TextColor, styleDrawBold(s), s.Italic, s.Mono, s.PreviewFont)
}

// MeasureChromeText returns pixel width for DrawChromeText.
func MeasureChromeText(text string, s Style) float32 {
	return measureTextDrawAligned(text, s)
}

// ChromeTextCenterY returns Y to vertically center text in a rectangle using
// measured glyph metrics (same as toolbar labels).
func ChromeTextCenterY(bounds rl.Rectangle, s Style) float32 {
	return float32(toolbarTextPosY(bounds, s))
}
