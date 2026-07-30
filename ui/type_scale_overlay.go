// Package ui (continued) — F6 typography debug overlay.
package ui

import (
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ShowTypeScaleOverlay toggles the typography diagnostic panel (optional; wire in main if needed).
var ShowTypeScaleOverlay bool

// Overlay chrome uses fixed readable sizes (not theme tokens).
const (
	typeOverlayW     float32 = 460
	typeOverlayPad   float32 = 14
	typeOverlayLH    float32 = 22
	typeOverlayBody  float32 = 16
	typeOverlayLabel float32 = 15
	typeOverlayTitle float32 = 20
	typeOverlayHint  float32 = 14
)

// DrawTypeScaleOverlay renders the typography debug panel. Call after the UI blit,
// inside rl.BeginDrawing (1× screen space). selected may be nil.
func DrawTypeScaleOverlay(windowW, windowH int32, selected Node) {
	if !ShowTypeScaleOverlay || windowW < 1 || windowH < 1 {
		return
	}

	panelH := float32(windowH) - 40
	if panelH < 160 {
		panelH = float32(windowH)
	}
	panel := rl.NewRectangle(10, 36, typeOverlayW, panelH)
	rl.DrawRectangleRounded(panel, 0.03, 10, rl.NewColor(12, 14, 24, 248))
	rl.DrawRectangleRoundedLinesEx(panel, 0.03, 10, 1.5, rl.NewColor(120, 140, 220, 255))

	x := panel.X + typeOverlayPad
	y := panel.Y + typeOverlayPad
	labelCol := rl.NewColor(150, 165, 210, 255)
	valueCol := rl.NewColor(235, 238, 252, 255)
	accentCol := rl.NewColor(255, 220, 120, 255)

	drawLine := func(label, value string, accent bool) {
		DrawText(label, x, y, typeOverlayLabel, labelCol)
		lw := MeasureText(label, typeOverlayLabel)
		col := valueCol
		if accent {
			col = accentCol
		}
		DrawText(value, x+lw, y, typeOverlayBody, col)
		y += typeOverlayLH
	}

	DrawText("Typography debug (F6)", x, y, typeOverlayTitle, rl.White)
	y += typeOverlayLH + 4

	drawLine("Window ", fmt.Sprintf("%d x %d", windowW, windowH), false)
	drawLine("1rem ", fmt.Sprintf("%.0f px", RootFontSize), true)
	drawLine("Scale ", fmt.Sprintf("%.3f  floor %.0f", typeScaleFactor(), TypeScaleMinRoot), false)
	drawLine("Fluid ", "1rem grows with window width", false)
	drawLine("SSAA ", fmt.Sprintf("%.2fx", RenderScale), false)
	drawLine("Backend ", UIFontBackend(), true)
	fontPath := LoadedUIFontPath
	if fontPath == "" {
		fontPath = "(default)"
	} else {
		fontPath = filepath.Base(fontPath)
	}
	drawLine("Font ", fontPath, false)
	drawLine("Icons ", Phosphor.IconFontSummary(), true)

	y += 6
	DrawText("Theme token -> screen px", x, y, typeOverlayLabel, labelCol)
	y += typeOverlayLH

	for _, name := range []string{
		"default", "input", "button", "form-label", "form-value",
		"settings-row-label", "bottomnav", "tooltip",
	} {
		st := GetThemeStyle(name)
		drawLine(name+" ", fmt.Sprintf("%d -> %.1f", st.FontSize, EffectiveFontSize(st)), false)
	}

	y += 4
	if selected != nil {
		st := GetThemeStyle(selected.StyleName())
		if styler, ok := selected.(interface{ GetStyle() Style }); ok {
			st = styler.GetStyle()
		}
		drawLine("Selected ", selected.ID(), true)
		drawLine("  style ", fmt.Sprintf("%s  %d -> %.1f px", selected.StyleName(), st.FontSize, EffectiveFontSize(st)), false)
	} else {
		drawLine("Selected ", "F12 inspector + click widget", false)
	}

	y += 8
	DrawText("Live samples", x, y, typeOverlayLabel, labelCol)
	y += typeOverlayLH
	for _, pair := range []struct {
		style string
		text  string
	}{
		{"default", "Body (default) — The quick brown fox"},
		{"form-value", "Caption (form-value) — status line"},
		{"form-label", "Label (form-label) — Country"},
	} {
		st := GetThemeStyle(pair.style)
		fs := EffectiveFontSize(st)
		DrawText(pair.text, x, y, fs, rl.NewColor(240, 242, 255, 255))
		y += fs + 8
	}

	y += 4
	DrawText("[ ] = 1rem +/- 1   full log -> stderr", x, y, typeOverlayHint, rl.NewColor(180, 190, 220, 255))
}

// TypeScaleOverlayAdjustRoot changes RootFontSize when the overlay is open ([ decrement, ] increment).
func TypeScaleOverlayAdjustRoot(delta float32, windowW, windowH int32, selected Node) {
	if !ShowTypeScaleOverlay || delta == 0 {
		return
	}
	SetRootFontSize(RootFontSize + delta)
	LogTypographyReport(fmt.Sprintf("root%+.0f", delta), windowW, windowH, selected)
}
