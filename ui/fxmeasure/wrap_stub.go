//go:build !cgo || gru_nofx

package fxmeasure

// Enabled reports that fx measure/wrap is not linked.
func Enabled() bool { return false }

// WrapAdvances stub — returns -1.
func WrapAdvances(advances []float32, canBreak []int32, maxW float32, outEnds []int32) int32 {
	return -1
}

// SumAdvances stub sums in pure Go so call sites can compile without gru_fx.
func SumAdvances(advances []float32, lo, hi int) float32 {
	if lo < 0 {
		lo = 0
	}
	if hi > len(advances) {
		hi = len(advances)
	}
	if lo > hi {
		return 0
	}
	var s float32
	for i := lo; i < hi; i++ {
		s += advances[i]
	}
	return s
}
