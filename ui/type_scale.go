// Package ui (continued) — root font size / rem-style typography scaling.
package ui

import (
	"fmt"
	"os"
)

// TypeScaleReference is the design grid [CurrentTheme] FontSize tokens are
// authored against — like html { font-size: 16px } before rem children.
// Tokens are not literal screen pixels; [EffectiveFontSize] scales them by
// RootFontSize/TypeScaleReference (see docs in RefreshTypeScaleFromWindow).
const TypeScaleReference float32 = 16

// TypeScaleMinRoot is 1rem at narrow widths (see [MinClientWidth]). Keeps body copy
// readable when the window is at minimum width: default token 18 → ~25 px eff.
const TypeScaleMinRoot float32 = 22

// Chrome / utility typography floors (effective pixels after rem scale).
// Applied via Style.MinFontSize on menubar, statusbar-label, text-editor, etc.
const (
	TypeScaleMinMenubarPx int32 = 22
	TypeScaleMinStatusPx  int32 = 19
	TypeScaleMinEditorPx  int32 = 17
)

// RootFontSize is 1rem in window pixels. [EffectiveFontSize] multiplies theme
// FontSize by RootFontSize/TypeScaleReference so one theme scales on large monitors
// the way web rem units do when the root font-size changes.
var RootFontSize float32 = TypeScaleReference

// SetRootFontSize sets 1rem (clamped 12–29). Call after startup or from app config.
func SetRootFontSize(px float32) {
	if px < 12 {
		px = 12
	}
	if px > 29 {
		px = 29
	}
	RootFontSize = px
}

// Rem returns window pixels for a rem multiple (Rem(1.25) → 20 when root is 16).
func Rem(mult float32) int32 {
	if mult < 0 {
		mult = 0
	}
	return int32(RootFontSize*mult + 0.5)
}

func typeScaleFactor() float32 {
	if RootFontSize <= 0 {
		return 1
	}
	return RootFontSize / TypeScaleReference
}

// RefreshTypeScaleFromWindow sets [RootFontSize] from client width (fluid root rem).
// Wider window → larger 1rem, like responsive html { font-size } / clamp() on the web.
// [TypeScaleMinRoot] prevents unreadable text at [MinClientWidth] (480).
func RefreshTypeScaleFromWindow(w, h int32) {
	dim := w
	if dim < 1 {
		dim = h
	}
	prev := RootFontSize
	switch {
	case dim >= 2560:
		RootFontSize = 26
	case dim >= 1920:
		RootFontSize = 25
	case dim >= 1600:
		RootFontSize = 24
	case dim >= 1400:
		RootFontSize = 23
	case dim >= 1200:
		RootFontSize = 24
	case dim >= 960:
		RootFontSize = 22
	default:
		// Narrow / phone column: always use the readability floor.
		RootFontSize = TypeScaleMinRoot
	}
	if RootFontSize < TypeScaleMinRoot {
		RootFontSize = TypeScaleMinRoot
	}
	if ShowTypeScaleOverlay && prev != RootFontSize {
		LogTypographyReport("resize", w, h, nil)
	}
}

// LogTypographyReport prints a full typography snapshot to stderr (F6 / [ ] / resize).
// Use stderr so it stays visible beside TraceLog idle lines in the terminal.
func LogTypographyReport(reason string, windowW, windowH int32, selected Node) {
	minDim := windowW
	if windowH < minDim {
		minDim = windowH
	}
	fontPath := LoadedUIFontPath
	if fontPath == "" {
		fontPath = "(raylib default)"
	}
	fmt.Fprintf(os.Stderr, "\n--- Gru typography [%s] ---\n", reason)
	fmt.Fprintf(os.Stderr, "  overlay=%v  window=%dx%d  width=%d (breakpoint)  minDim=%d\n",
		ShowTypeScaleOverlay, windowW, windowH, windowW, minDim)
	fmt.Fprintf(os.Stderr, "  1rem=%.0fpx (floor %.0f)  scale=%.3f  ref=%g  dpi=%.2f  ssaa=%.2fx\n",
		RootFontSize, TypeScaleMinRoot, typeScaleFactor(), TypeScaleReference, DisplayScale, RenderScale)
	fmt.Fprintf(os.Stderr, "  super=%v  fluidRoot=%v (1rem tracks width only, not height)\n",
		SuperFrameActive(), windowW >= 960)
	fmt.Fprintf(os.Stderr, "  backend=%s  font=%s\n", UIFontBackend(), fontPath)
	fmt.Fprintf(os.Stderr, "  icons=%s\n", Phosphor.IconFontSummary())
	for _, name := range []string{"default", "input", "button", "form-label", "form-value"} {
		st := GetThemeStyle(name)
		fmt.Fprintf(os.Stderr, "  %-12s token=%2d -> eff=%.1fpx\n", name+":", st.FontSize, EffectiveFontSize(st))
	}
	if selected != nil {
		st := GetThemeStyle(selected.StyleName())
		if styler, ok := selected.(interface{ GetStyle() Style }); ok {
			st = styler.GetStyle()
		}
		fmt.Fprintf(os.Stderr, "  selected: id=%s style=%s token=%d eff=%.1fpx\n",
			selected.ID(), selected.StyleName(), st.FontSize, EffectiveFontSize(st))
	} else {
		fmt.Fprintf(os.Stderr, "  selected: (none — F12 inspector + click)\n")
	}
	fmt.Fprintf(os.Stderr, "---\n\n")
}
