//go:build cgo && !gru_nofx

// Package fxmeasure is the Go façade over fxlib/measure_wrap.fx emit-C (Phase 3).
//
// Source of truth: fxlib/measure_wrap.fx — regenerate with scripts/emit-fxlib.ps1.
package fxmeasure

/*
#cgo CFLAGS: -I${SRCDIR}/../../gen/fxlib
#include "measure_wrap.h"
*/
import "C"
import "unsafe"

// Enabled reports that the fx measure/wrap emit-C is linked.
func Enabled() bool { return true }

// WrapAdvances greedy-wraps advances into lines.
// canBreak[i]=1 allows a break after index i; outEnds[k] is exclusive end of line k.
// Returns line count, or negative on error.
func WrapAdvances(advances []float32, canBreak []int32, maxW float32, outEnds []int32) int32 {
	if len(canBreak) < len(advances) || len(outEnds) < 1 {
		return -1
	}
	var cadv C.fx_Slice_f32
	if len(advances) > 0 {
		cadv.data = (*C.float)(unsafe.Pointer(&advances[0]))
		cadv.len = C.size_t(len(advances))
	}
	var cbrk C.fx_Slice_i32
	if len(canBreak) > 0 {
		cbrk.data = (*C.int32_t)(unsafe.Pointer(&canBreak[0]))
		cbrk.len = C.size_t(len(canBreak))
	}
	var cout C.fx_MutSlice_i32
	cout.data = (*C.int32_t)(unsafe.Pointer(&outEnds[0]))
	cout.len = C.size_t(len(outEnds))
	return int32(C.fx_measure_wrap_wrap_advances(cadv, cbrk, C.float(maxW), cout))
}

// SumAdvances sums advances[lo:hi].
func SumAdvances(advances []float32, lo, hi int) float32 {
	var cadv C.fx_Slice_f32
	if len(advances) > 0 {
		cadv.data = (*C.float)(unsafe.Pointer(&advances[0]))
		cadv.len = C.size_t(len(advances))
	}
	return float32(C.fx_measure_wrap_sum_advances(cadv, C.int32_t(lo), C.int32_t(hi)))
}
