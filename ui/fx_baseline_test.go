package ui

import (
	"strings"
	"testing"
)

func TestFxPhase1BaselineReport(t *testing.T) {
	style := wrapProbeStyle()
	para := strings.Repeat("hello world wrap probe ascii only. ", 120)
	snap := CaptureFxBaseline(para, 280, style, 5)
	t.Logf("fx baseline: heap=%.1fMB sys=%.1fMB gc=%d wrap=%s match=%v pod=%.2fms classic=%.2fms glyph=%s fx_wrap=%v fx_glyph=%v",
		snap.HeapAllocMB, snap.HeapSysMB, snap.NumGC,
		snap.WrapPath, snap.WrapMatch, snap.WrapPODMs, snap.WrapClassicMs,
		snap.GlyphPath, snap.FXWrapLinked, snap.FXGlyphLinked)
	if !snap.WrapMatch {
		t.Fatalf("wrap POD vs classic mismatch path=%s", snap.WrapPath)
	}
	if snap.WrapPODMs <= 0 || snap.WrapClassicMs <= 0 {
		t.Fatalf("timing missing pod=%.2f classic=%.2f", snap.WrapPODMs, snap.WrapClassicMs)
	}
}

func TestWrapShapedClusterJoins(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	SetTextEngineMode(TextEngineShaped)
	defer SetTextEngineMode(TextEngineSDF)

	style := Style{FontSize: 16, MinFontSize: 16}
	for _, sample := range []struct {
		name string
		text string
		w    float32
	}{
		{"arabic", ComplexScriptArabicSample, 120},
		{"devanagari", ComplexScriptDevanagariSample, 140},
		{"mixed", ComplexScriptMixedSample, 200},
	} {
		t.Run(sample.name, func(t *testing.T) {
			lines := wrapParagraphByAdvances(sample.text, sample.w, style)
			if !joinWrapLines(sample.text, lines) {
				t.Fatalf("shaped wrap lost bytes: lines=%d text=%q", len(lines), sample.text)
			}
		})
	}
}

func TestFillWrapAdvancesShapedOffsets(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	SetTextEngineMode(TextEngineShaped)
	defer SetTextEngineMode(TextEngineSDF)

	style := Style{FontSize: 16, MinFontSize: 16}
	sc := &wrapScratch{}
	if !fillWrapAdvancesShaped(ComplexScriptArabicSample, style, sc) {
		t.Fatal("fillWrapAdvancesShaped failed")
	}
	if len(sc.off) < 2 || sc.off[len(sc.off)-1] != len(ComplexScriptArabicSample) {
		t.Fatalf("offsets=%v want end=%d", sc.off, len(ComplexScriptArabicSample))
	}
	if len(sc.adv) != len(sc.can) || len(sc.adv) != len(sc.off)-1 {
		t.Fatalf("adv/can/off mismatch adv=%d can=%d off=%d", len(sc.adv), len(sc.can), len(sc.off))
	}
}
