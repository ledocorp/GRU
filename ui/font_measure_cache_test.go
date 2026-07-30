package ui

import "testing"

func TestSDFMeasureCacheHit(t *testing.T) {
	if !sdfReady {
		t.Skip("SDF font not loaded")
	}
	sdfMeasureCacheClear()
	st := Style{FontSize: 16}
	w1 := measureTextS("Gallery label", st)
	w2 := measureTextS("Gallery label", st)
	if w1 != w2 || w1 <= 0 {
		t.Fatalf("widths = %d, %d", w1, w2)
	}
}

func TestMeasureTextSMatchesDrawLatinShaped(t *testing.T) {
	if !InitShapedFonts() || !sdfReady {
		t.Skip("fonts not loaded")
	}
	SetTextEngineMode(TextEngineShaped)
	st := Style{FontSize: 16}
	text := "Authoring **bold** span"
	fs := EffectiveFontSize(st)
	sdfW := measureTextSDF(text, fs, styleDrawBold(st), st.Italic, st.Mono, st.PreviewFont)
	got := measureTextS(text, st)
	SetTextEngineMode(TextEngineSDF)
	if got != int32(sdfW) {
		t.Fatalf("measureTextS = %d, SDF draw width = %.0f", got, sdfW)
	}
}
