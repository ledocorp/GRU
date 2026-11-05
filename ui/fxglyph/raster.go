//go:build cgo && !gru_nofx

// Package fxglyph is the Go façade over fxlib/glyph_raster.fx emit-C (Phase 2).
//
// Source of truth: fxlib/glyph_raster.fx — regenerate with scripts/emit-fxlib.ps1.
// Go links emit-C only; never fx IR.
package fxglyph

/*
#cgo CFLAGS: -I${SRCDIR}/../../gen/fxlib
#include "glyph_raster.h"
*/
import "C"
import "unsafe"

// Enabled reports that the fx glyph raster emit-C is linked.
func Enabled() bool { return true }

// RasterOutline flattens ops/pts and non-zero-fills into preallocated RGBA (w*h*4).
// ops: 0=Move 1=Line 2=Quad 3=Cube; pts: flat x,y font units.
// edges/xs/dirs are scratch; lengths should be generous (edges >= 8192, xs/dirs >= 128).
// Returns ink pixel count, or negative error from fx_glyph_raster_raster_outline.
func RasterOutline(ops []int32, pts []float32, w, h int, scale, ox, oy float32, edges, xs []float32, dirs []int32, rgba []byte) int32 {
	if w <= 0 || h <= 0 || len(rgba) < w*h*4 {
		return -1
	}
	if len(dirs) < len(xs) {
		return -1
	}
	var cops C.fx_Slice_i32
	if len(ops) > 0 {
		cops.data = (*C.int32_t)(unsafe.Pointer(&ops[0]))
		cops.len = C.size_t(len(ops))
	}
	var cpts C.fx_Slice_f32
	if len(pts) > 0 {
		cpts.data = (*C.float)(unsafe.Pointer(&pts[0]))
		cpts.len = C.size_t(len(pts))
	}
	var cedges C.fx_MutSlice_f32
	if len(edges) > 0 {
		cedges.data = (*C.float)(unsafe.Pointer(&edges[0]))
		cedges.len = C.size_t(len(edges))
	}
	var cxs C.fx_MutSlice_f32
	if len(xs) > 0 {
		cxs.data = (*C.float)(unsafe.Pointer(&xs[0]))
		cxs.len = C.size_t(len(xs))
	}
	var cdirs C.fx_MutSlice_i32
	if len(dirs) > 0 {
		cdirs.data = (*C.int32_t)(unsafe.Pointer(&dirs[0]))
		cdirs.len = C.size_t(len(dirs))
	}
	var crgba C.fx_MutSlice_u8
	crgba.data = (*C.uint8_t)(unsafe.Pointer(&rgba[0]))
	crgba.len = C.size_t(len(rgba))
	return int32(C.fx_glyph_raster_raster_outline(
		cops, cpts,
		C.int32_t(w), C.int32_t(h),
		C.float(scale), C.float(ox), C.float(oy),
		cedges, cxs, cdirs, crgba,
	))
}

// BoostRGBAAlpha applies Skia-lite contrast to alpha channels in RGBA8 (stride 4).
// Used by S1 greyscale strike enhance when GRU_FX_GLYPH=1 (not DWrite LCD).
func BoostRGBAAlpha(rgba []byte, contrast float32) int32 {
	if contrast <= 0 || len(rgba) < 4 {
		return 0
	}
	var c C.fx_MutSlice_u8
	c.data = (*C.uint8_t)(unsafe.Pointer(&rgba[0]))
	c.len = C.size_t(len(rgba))
	return int32(C.fx_glyph_raster_boost_rgba_alpha(c, C.float(contrast)))
}
