package ui

import "testing"

func TestStatusBarPipeCapsAtMaxFontSize(t *testing.T) {
	RefreshTypeScaleFromWindow(1920, 1080)
	pipe := GetThemeStyle("statusbar-pipe")
	got := EffectiveFontSize(pipe)
	if got != 16 {
		t.Fatalf("statusbar-pipe eff=%.1f, want literal 16px cap", got)
	}
	rt := NewRichText("rt", nil, 0, 0, 200, 20)
	rt.SetStyle("statusbar-label")
	st := rt.spanStyle(TextSpan{Text: " | ", Style: "statusbar-pipe"})
	got = EffectiveFontSize(st)
	if got != 16 {
		t.Fatalf("merged pipe span eff=%.1f, want 16 (not label floor %d)", got, TypeScaleMinStatusPx)
	}
	label := EffectiveFontSize(GetThemeStyle("statusbar-label"))
	if got >= label {
		t.Fatalf("pipe eff=%.1f must be smaller than label eff=%.1f", got, label)
	}
}

func TestTypeScaleAtMinClientWidth(t *testing.T) {
	RefreshTypeScaleFromWindow(MinClientWidth, 800)
	if RootFontSize != TypeScaleMinRoot {
		t.Fatalf("1rem = %.0f, want floor %.0f at %dpx width", RootFontSize, TypeScaleMinRoot, MinClientWidth)
	}
	def := GetThemeStyle("default")
	got := EffectiveFontSize(def)
	want := float32(def.FontSize) * TypeScaleMinRoot / TypeScaleReference
	if got < want-0.5 || got > want+0.5 {
		t.Fatalf("default eff=%.1f want≈%.1f (token %d)", got, want, def.FontSize)
	}
}
