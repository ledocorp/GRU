package ui

import (
	"runtime"
	"strings"
	"time"

	"github.com/ledocorp/gru/ui/fxglyph"
	"github.com/ledocorp/gru/ui/fxmeasure"
)

// FxBaselineSnapshot captures Phase 1 microbench + heap stats for FX_INTEGRATION_PLAN §9.
// Full-scene RSS requires a running app; this harness covers hot subsystems in-process.
type FxBaselineSnapshot struct {
	HeapAllocMB   float64
	HeapSysMB     float64
	NumGC         uint32
	WrapPath      string
	WrapPODLines  int
	WrapClassic   int
	WrapMatch     bool
	WrapPODMs     float64
	WrapClassicMs float64
	GlyphPath     string
	FXWrapLinked  bool
	FXGlyphLinked bool
}

// CaptureFxBaseline runs editor-wrap compare + reads MemStats (call after InitSDFFont in tests).
func CaptureFxBaseline(para string, maxW float32, style Style, loops int) FxBaselineSnapshot {
	if loops < 1 {
		loops = 3
	}
	if para == "" {
		para = strings.Repeat("hello world wrap probe ascii only. ", 120)
	}
	if maxW <= 0 {
		maxW = 280
	}
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	rep := CompareEditorWrap(para, maxW, style, loops)
	snap := FxBaselineSnapshot{
		HeapAllocMB:   float64(ms.Alloc) / (1024 * 1024),
		HeapSysMB:     float64(ms.HeapSys) / (1024 * 1024),
		NumGC:         ms.NumGC,
		WrapPath:      rep.Path,
		WrapPODLines:  rep.PODLines,
		WrapClassic:   rep.ClassicLines,
		WrapMatch:     rep.Match,
		WrapPODMs:     float64(rep.PODNs) / float64(time.Millisecond),
		WrapClassicMs: float64(rep.ClassicNs) / float64(time.Millisecond),
		GlyphPath:     GlyphRasterPathName(),
		FXWrapLinked:  fxmeasure.Enabled(),
		FXGlyphLinked: fxglyph.Enabled(),
	}
	return snap
}

// GlyphRasterPathName reports the live complex-script glyph raster backend.
func GlyphRasterPathName() string {
	if fxGlyphBackendActive() {
		return "fx"
	}
	if fxOptOut("GRU_FX_GLYPH", "GORY_FX_GLYPH") {
		return "gg"
	}
	if fxglyph.Enabled() {
		return "fx"
	}
	return "gg"
}
