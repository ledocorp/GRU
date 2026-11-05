//go:build !cgo || gru_nofx

// Package fxbridge stub when gru_fx is not enabled (default builds).
package fxbridge

// Enabled reports that the emit-C probe is not linked.
func Enabled() bool { return false }

// Version returns the vendored pack VERSION even when emit-C is not linked.
func Version() string { return packVersion }

// Add is a pure-Go stand-in so call sites can compile without gru_fx.
func Add(a, b int32) int32 { return a + b }

// Clamp is a pure-Go stand-in.
func Clamp(v, lo, hi int32) int32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Hadd4 is a pure-Go stand-in (same math as the probe vectors).
func Hadd4(a0, a1, a2, a3, b0, b1, b2, b3 int32) int32 {
	return (a0 + b0) + (a1 + b1) + (a2 + b2) + (a3 + b3)
}

// Hadd4f is a pure-Go stand-in.
func Hadd4f(a0, a1, a2, a3, b0, b1, b2, b3 float32) float32 {
	return (a0 + b0) + (a1 + b1) + (a2 + b2) + (a3 + b3)
}

// Sum is a pure-Go stand-in.
func Sum(xs []int32) int32 {
	var n int32
	for _, v := range xs {
		n += v
	}
	return n
}
