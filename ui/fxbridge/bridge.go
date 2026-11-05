//go:build cgo && !gru_nofx

// Package fxbridge is the thin Go façade over fx emit-C (Phase 0 probe).
//
// Go links **emit-C only** — never fx IR. Pack: third_party/fx/emit-go-abi
// (from D:\fx dist/emit-go-abi). Rebuild pack: fx repo scripts/package-emit-go-abi.ps1
//
//	go test -tags gru_fx ./ui/fxbridge/
package fxbridge

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/fx/emit-go-abi/gen -I${SRCDIR}/../../third_party/fx/emit-go-abi/include
#cgo LDFLAGS: -L${SRCDIR}/../../third_party/fx/emit-go-abi/lib -lzspec
#include "emit_go_probe.h"
*/
import "C"
import "unsafe"

// Enabled reports that the gru_fx emit-C probe is linked.
func Enabled() bool { return true }

// Version returns the pinned fx package VERSION from the consumer pack.
func Version() string { return packVersion }

// Add returns a+b (fx_emit_go_probe_add).
func Add(a, b int32) int32 {
	return int32(C.fx_emit_go_probe_add(C.int32_t(a), C.int32_t(b)))
}

// Clamp returns v clamped to [lo, hi] (fx_emit_go_probe_clamp).
func Clamp(v, lo, hi int32) int32 {
	return int32(C.fx_emit_go_probe_clamp(C.int32_t(v), C.int32_t(lo), C.int32_t(hi)))
}

// Hadd4 is POD-only (eight int32 in, one out) — Go never sees fx_v4i32.
func Hadd4(a0, a1, a2, a3, b0, b1, b2, b3 int32) int32 {
	return int32(C.fx_emit_go_probe_hadd4(
		C.int32_t(a0), C.int32_t(a1), C.int32_t(a2), C.int32_t(a3),
		C.int32_t(b0), C.int32_t(b1), C.int32_t(b2), C.int32_t(b3),
	))
}

// Hadd4f is POD-only (eight float32 in, one out) — Go never sees fx_v4f32.
func Hadd4f(a0, a1, a2, a3, b0, b1, b2, b3 float32) float32 {
	return float32(C.fx_emit_go_probe_hadd4f(
		C.float(a0), C.float(a1), C.float(a2), C.float(a3),
		C.float(b0), C.float(b1), C.float(b2), C.float(b3),
	))
}

// Sum is buffer POD — Go builds fx_Slice_i32 { data, len } from []int32.
func Sum(xs []int32) int32 {
	var s C.fx_Slice_i32
	if len(xs) > 0 {
		s.data = (*C.int32_t)(unsafe.Pointer(&xs[0]))
		s.len = C.size_t(len(xs))
	}
	return int32(C.fx_emit_go_probe_sum(s))
}
