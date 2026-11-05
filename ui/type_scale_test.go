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

// TestTypeScaleStableAcrossTypicalDesktopWidths guards the size-stability contract:
// init 1280 and settled ~1024 must share the same rem so first paint cannot
// “enlarge then shrink” when the OS client size catches up.
func TestTypeScaleStableAcrossTypicalDesktopWidths(t *testing.T) {
	prev := RootFontSize
	t.Cleanup(func() { RootFontSize = prev })

	var remAt1280 float32
	for _, w := range []int32{MinClientWidth, 800, 960, 1024, 1199, 1200, 1280, 1600, 1919} {
		RefreshTypeScaleFromWindow(w, 800)
		if RootFontSize != TypeScaleMinRoot {
			t.Fatalf("width %d: rem=%.0f, want stable desktop %.0f", w, RootFontSize, TypeScaleMinRoot)
		}
		if w == 1280 {
			remAt1280 = RootFontSize
		}
	}
	RefreshTypeScaleFromWindow(1024, 1048) // dogfood settle size from resize-fps logs
	if RootFontSize != remAt1280 {
		t.Fatalf("1024 settle rem=%.0f ≠ 1280 init rem=%.0f (band cliff regresses oversized→shrink)",
			RootFontSize, remAt1280)
	}

	RefreshTypeScaleFromWindow(TypeScaleLargeWidthPx, 1080)
	if RootFontSize != TypeScaleLargeRoot {
		t.Fatalf("1920 rem=%.0f, want %.0f", RootFontSize, TypeScaleLargeRoot)
	}
	RefreshTypeScaleFromWindow(TypeScaleXLWidthPx, 1440)
	if RootFontSize != TypeScaleXLRoot {
		t.Fatalf("2560 rem=%.0f, want %.0f", RootFontSize, TypeScaleXLRoot)
	}

	// Still monotonic overall.
	var last float32
	for _, w := range []int32{480, 1280, 1920, 2560} {
		RefreshTypeScaleFromWindow(w, 800)
		if RootFontSize < last {
			t.Fatalf("rem shrank at width %d: %.0f → %.0f", w, last, RootFontSize)
		}
		last = RootFontSize
	}
}
