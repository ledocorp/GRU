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

// TypeScaleMinRoot is 1rem for phone-column through typical desktop windows.
// Stable desktop rem (see RefreshTypeScaleFromWindow): everyday resize must not
// retune chrome type. Kept near the design grid (16) so demos don’t feel oversized;
// MinFontSize floors still protect menubar/status readability.
const TypeScaleMinRoot float32 = 18

// TypeScaleLargeRoot / TypeScaleXLRoot bump 1rem only on large canvases.
// Below TypeScaleLargeWidthPx rem stays at TypeScaleMinRoot — native apps do
// not shrink body type when the window goes from 1280→1024.
const (
	TypeScaleLargeWidthPx int32   = 1920
	TypeScaleXLWidthPx    int32   = 2560
	TypeScaleLargeRoot    float32 = 20
	TypeScaleXLRoot       float32 = 22
)

// Chrome / utility typography floors (effective pixels after rem scale).
// Applied via Style.MinFontSize on menubar, statusbar-label, text-editor, etc.
// Segmented tabs share a floor so Foundry/Showcase don't sit on soft ~14–19px.
const (
	TypeScaleMinMenubarPx   int32 = 18
	TypeScaleMinStatusPx    int32 = 15
	TypeScaleMinEditorPx    int32 = 14
	TypeScaleMinSegmentedPx int32 = 15
)

// RootFontSize is 1rem in window pixels. [EffectiveFontSize] multiplies theme
// FontSize by RootFontSize/TypeScaleReference so one theme scales on large monitors
// the way web rem units do when the root font-size changes.
var RootFontSize float32 = TypeScaleReference

// SetRootFontSize sets 1rem (clamped 12–48). Call after startup or from app config.
// Desktop fluid scale stays ≤26; Android phone rem needs up to ~40 for ~16sp body.
func SetRootFontSize(px float32) {
	if px < 12 {
		px = 12
	}
	if px > 48 {
		px = 48
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

// RefreshTypeScaleFromWindow sets [RootFontSize] from client width.
//
// Contract (size stability — do not regress):
//   - Typical desktop widths (MinClientWidth … <1920) share one rem: TypeScaleMinRoot.
//   - Layout still reflows with the window; chrome type size does not.
//   - Only large canvases (≥1920 / ≥2560) bump rem — like a larger design density,
//     not a per-resize fluid clamp.
//
// Why: a prior fluid ladder (960→22, 1200→24) made initWindow 1280 seed rem=24 then
// shrink to rem=22 when the OS settled ~1024. Timing syncs cannot fix a band cliff
// between the requested window size and the real client size. Native Notepad-class
// apps keep type size across everyday resize.
//
// While [typeScaleResizeHold] is set, RootFontSize is frozen. Studio does not enable
// hold on the hot path. The API remains for tests / specialized hosts.
//
// Android: desktop rem is too small at 420dpi (1080px ≈ 411dp). Target
// ~16–18sp body (token 18 → ~42–47px) via a phone rem (see below).
func RefreshTypeScaleFromWindow(w, h int32) {
	if typeScaleResizeHold {
		return
	}
	dim := w
	if dim < 1 {
		dim = h
	}
	prev := RootFontSize
	if AndroidGLES() {
		// Phone rem on ~411dp (1080@420dpi): body ≈16–17sp (token 18 → ~45px).
		// Chrome sizes use [Dp] / [Density] — do not multiply rem into chrome
		// (docs/ANDROID_DENSITY.md).
		RootFontSize = 40
		if dim >= 1200 {
			RootFontSize = 42
		} else if dim < 720 {
			RootFontSize = 36
		}
	} else {
		switch {
		case dim >= TypeScaleXLWidthPx:
			RootFontSize = TypeScaleXLRoot
		case dim >= TypeScaleLargeWidthPx:
			RootFontSize = TypeScaleLargeRoot
		default:
			RootFontSize = TypeScaleMinRoot
		}
		if RootFontSize < TypeScaleMinRoot {
			RootFontSize = TypeScaleMinRoot
		}
	}
	if ShowTypeScaleOverlay && prev != RootFontSize {
		LogTypographyReport("resize", w, h, nil)
	}
}

var typeScaleResizeHold bool

// SetTypeScaleResizeHold freezes rem during live window resize (S1 anti-thrash).
// When released, caller should [RefreshTypeScaleFromWindow] once. Breakpoints unchanged.
func SetTypeScaleResizeHold(hold bool) {
	typeScaleResizeHold = hold
}

// TypeScaleResizeHold reports whether rem is frozen for resize.
func TypeScaleResizeHold() bool { return typeScaleResizeHold }

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
	fmt.Fprintf(os.Stderr, "\n--- GoRy typography [%s] ---\n", reason)
	fmt.Fprintf(os.Stderr, "  overlay=%v  window=%dx%d  width=%d (breakpoint)  minDim=%d\n",
		ShowTypeScaleOverlay, windowW, windowH, windowW, minDim)
	fmt.Fprintf(os.Stderr, "  1rem=%.0fpx (floor %.0f)  scale=%.3f  ref=%g  dpi=%.2f  ssaa=%.2fx\n",
		RootFontSize, TypeScaleMinRoot, typeScaleFactor(), TypeScaleReference, DisplayScale, RenderScale)
	fmt.Fprintf(os.Stderr, "  super=%v  stableDesktopRem=%v (1rem flat until width≥%d)\n",
		SuperFrameActive(), windowW < TypeScaleLargeWidthPx, TypeScaleLargeWidthPx)
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
