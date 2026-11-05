//go:build !cgo || gru_nofx

package fxglyph_test

import (
	"testing"

	"github.com/ledocorp/gru/ui/fxglyph"
)

func TestStubDisabled(t *testing.T) {
	if fxglyph.Enabled() {
		t.Fatal("Enabled")
	}
}
