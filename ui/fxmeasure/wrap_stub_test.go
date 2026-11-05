//go:build !cgo || gru_nofx

package fxmeasure_test

import (
	"testing"

	"github.com/ledocorp/gru/ui/fxmeasure"
)

func TestStub(t *testing.T) {
	if fxmeasure.Enabled() {
		t.Fatal("Enabled")
	}
	if fxmeasure.SumAdvances([]float32{1, 2, 3}, 0, 3) != 6 {
		t.Fatal("stub sum")
	}
}
