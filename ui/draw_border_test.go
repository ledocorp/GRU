package ui

import "testing"

func TestSnapChromeBorderWidthSSAA1KeepsHairline(t *testing.T) {
	old := RenderScale
	RenderScale = 1
	t.Cleanup(func() { RenderScale = old })

	if got := snapChromeBorderWidth(1.5); got != 1 {
		t.Fatalf("1.5 at SSAA 1× = %v, want 1", got)
	}
	if got := snapChromeBorderWidth(1); got != 1 {
		t.Fatalf("1 at SSAA 1× = %v, want 1", got)
	}
	if got := snapChromeBorderWidth(2); got != 2 {
		t.Fatalf("2 at SSAA 1× = %v, want 2", got)
	}
}

func TestSnapChromeBorderWidthSSAA2RoundsNearest(t *testing.T) {
	old := RenderScale
	RenderScale = 2
	t.Cleanup(func() { RenderScale = old })

	if got := snapChromeBorderWidth(1.5); got != 2 {
		t.Fatalf("1.5 at SSAA 2× = %v, want 2", got)
	}
}
