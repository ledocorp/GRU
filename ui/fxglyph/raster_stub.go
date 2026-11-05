//go:build !cgo || gru_nofx

// Package fxglyph stub when CGO is off or gru_nofx is set (testing fallback).
package fxglyph

// Enabled reports that the fx glyph raster is not linked.
func Enabled() bool { return false }

// RasterOutline is a stub that always fails so callers fall back to gg.
func RasterOutline(ops []int32, pts []float32, w, h int, scale, ox, oy float32, edges, xs []float32, dirs []int32, rgba []byte) int32 {
	return -1
}

// BoostRGBAAlpha stub — callers use Go contrast path.
func BoostRGBAAlpha(rgba []byte, contrast float32) int32 {
	_ = rgba
	_ = contrast
	return 0
}
