//go:build cgo && !gru_nofx

package fxbridge_test

import (
	"testing"

	"github.com/ledocorp/gru/ui/fxbridge"
)

func TestEnabled(t *testing.T) {
	if !fxbridge.Enabled() {
		t.Fatal("Enabled")
	}
	if fxbridge.Version() == "" {
		t.Fatal("Version empty")
	}
}

func TestAdd(t *testing.T) {
	if fxbridge.Add(10, 32) != 42 {
		t.Fatalf("Add(10,32)=%d", fxbridge.Add(10, 32))
	}
}

func TestClamp(t *testing.T) {
	if fxbridge.Clamp(100, 0, 50) != 50 {
		t.Fatal("clamp high")
	}
	if fxbridge.Clamp(-5, 0, 50) != 0 {
		t.Fatal("clamp low")
	}
	if fxbridge.Clamp(25, 0, 50) != 25 {
		t.Fatal("clamp mid")
	}
}

func TestHadd4(t *testing.T) {
	if fxbridge.Hadd4(1, 2, 3, 4, 10, 20, 30, 40) != 110 {
		t.Fatalf("Hadd4=%d want 110", fxbridge.Hadd4(1, 2, 3, 4, 10, 20, 30, 40))
	}
}

func TestHadd4f(t *testing.T) {
	got := fxbridge.Hadd4f(1, 2, 3, 4, 0.5, 0.5, 0.5, 0.5)
	if got != 12 {
		t.Fatalf("Hadd4f=%v want 12", got)
	}
}

func TestSum(t *testing.T) {
	if fxbridge.Sum([]int32{10, 20, 12}) != 42 {
		t.Fatalf("Sum=%d want 42", fxbridge.Sum([]int32{10, 20, 12}))
	}
	if fxbridge.Sum(nil) != 0 {
		t.Fatal("Sum(nil)")
	}
}
