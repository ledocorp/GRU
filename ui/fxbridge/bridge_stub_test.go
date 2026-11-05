//go:build !cgo || gru_nofx

package fxbridge_test

import (
	"testing"

	"github.com/ledocorp/gru/ui/fxbridge"
)

func TestStubEnabled(t *testing.T) {
	if fxbridge.Enabled() {
		t.Fatal("Enabled should be false with gru_nofx or !cgo")
	}
}

func TestStubAdd(t *testing.T) {
	if fxbridge.Add(10, 32) != 42 {
		t.Fatalf("stub Add=%d", fxbridge.Add(10, 32))
	}
}
