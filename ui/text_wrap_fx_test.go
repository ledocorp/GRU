package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/ledocorp/gru/ui/fxmeasure"
)

func wrapProbeStyle() Style {
	return Style{FontSize: 17, MinFontSize: 17, Mono: true}
}

func TestWrapAdvancesLines(t *testing.T) {
	adv := []float32{10, 10, 5, 10, 10, 5, 10, 10}
	brk := []int32{0, 0, 1, 0, 0, 1, 0, 0}
	out := make([]int32, 8)
	n := wrapAdvancesLines(adv, brk, 25, out)
	if n != 3 || out[0] != 3 || out[1] != 6 || out[2] != 8 {
		t.Fatalf("n=%d ends=%v", n, out)
	}
}

func TestWrapEditorLinesPODJoinsAndMatchesClassicASCII(t *testing.T) {
	style := wrapProbeStyle()
	para := strings.Repeat("hello world  wrap probe. ", 48)
	para += "trailing "
	for _, w := range []float32{72, 120, 200, 320, 800} {
		pod := wrapParagraphByAdvances(para, w, style)
		classic := wrapEditorLinesClassic(para, w, style)
		if !joinWrapLines(para, pod) {
			t.Fatalf("POD join failed width=%v lines=%d", w, len(pod))
		}
		if !joinWrapLines(para, classic) {
			t.Fatalf("classic join failed width=%v lines=%d", w, len(classic))
		}
		if ok, idx := wrapLinesEqual(pod, classic); !ok {
			var a, b string
			if idx >= 0 && idx < len(pod) {
				a = pod[idx]
			}
			if idx >= 0 && idx < len(classic) {
				b = classic[idx]
			}
			t.Fatalf("POD vs classic mismatch width=%v line=%d pod=%q classic=%q (n %d vs %d)",
				w, idx, a, b, len(pod), len(classic))
		}
	}
}

func TestWrapEditorLinesDefaultUsesPODJoin(t *testing.T) {
	style := wrapProbeStyle()
	src := "hello  world"
	lines := wrapEditorLines(src, 800, style)
	if !joinWrapLines(src, lines) {
		t.Fatalf("default wrap lost bytes: %#v", lines)
	}
}

func TestCompareEditorWrapReport(t *testing.T) {
	style := wrapProbeStyle()
	para := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20)
	rep := CompareEditorWrap(para, 220, style, 2)
	if !rep.PODJoinedOK || !rep.ClassicJoinedOK {
		t.Fatalf("join flags pod=%v classic=%v", rep.PODJoinedOK, rep.ClassicJoinedOK)
	}
	if !rep.Match {
		t.Fatalf("expected match path=%s mismatchLine=%d lines %d vs %d",
			rep.Path, rep.MismatchLine, rep.PODLines, rep.ClassicLines)
	}
	if rep.Path != "pod-go" && rep.Path != "pod-fx" {
		t.Fatalf("path=%q", rep.Path)
	}
}

func TestEditorWrapPathFX(t *testing.T) {
	if !fxmeasure.Enabled() {
		t.Skip("fx not linked")
	}
	prev := os.Getenv("GRU_FX_WRAP")
	_ = os.Unsetenv("GRU_FX_WRAP")
	t.Cleanup(func() {
		if prev == "" {
			_ = os.Unsetenv("GRU_FX_WRAP")
		} else {
			_ = os.Setenv("GRU_FX_WRAP", prev)
		}
	})
	if EditorWrapPathName() != "pod-fx" {
		t.Fatalf("path=%q want pod-fx", EditorWrapPathName())
	}
	style := wrapProbeStyle()
	para := strings.Repeat("wrap fx path probe ascii only. ", 32)
	rep := CompareEditorWrap(para, 240, style, 1)
	if rep.Path != "pod-fx" || !rep.Match {
		t.Fatalf("pod-fx mismatch path=%s match=%v line=%d", rep.Path, rep.Match, rep.MismatchLine)
	}
}

func BenchmarkWrapEditorLinesClassic(b *testing.B) {
	style := wrapProbeStyle()
	para := strings.Repeat("hello world wrap probe ascii only. ", 120)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapEditorLinesClassic(para, 280, style)
	}
}

func BenchmarkWrapEditorLinesPOD(b *testing.B) {
	style := wrapProbeStyle()
	para := strings.Repeat("hello world wrap probe ascii only. ", 120)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapParagraphByAdvances(para, 280, style)
	}
}
