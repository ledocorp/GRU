//go:build cgo && !gru_nofx

package fxmeasure_test

import (
	"testing"

	"github.com/ledocorp/gru/ui/fxmeasure"
)

func TestWrapAdvances(t *testing.T) {
	if !fxmeasure.Enabled() {
		t.Fatal("Enabled")
	}
	// widths 10 each; break after spaces at 2 and 5; max 25 → "ab " / "cd " / "ef"
	adv := []float32{10, 10, 5, 10, 10, 5, 10, 10}
	brk := []int32{0, 0, 1, 0, 0, 1, 0, 0}
	out := make([]int32, 8)
	n := fxmeasure.WrapAdvances(adv, brk, 25, out)
	if n != 3 {
		t.Fatalf("lines=%d want 3 (ends=%v)", n, out[:n])
	}
	if out[0] != 3 || out[1] != 6 || out[2] != 8 {
		t.Fatalf("ends=%v", out[:n])
	}
}

func TestSumAdvances(t *testing.T) {
	adv := []float32{1, 2, 3, 4}
	if fxmeasure.SumAdvances(adv, 1, 3) != 5 {
		t.Fatalf("sum=%v", fxmeasure.SumAdvances(adv, 1, 3))
	}
}
