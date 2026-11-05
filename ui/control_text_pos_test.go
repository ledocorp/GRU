package ui

import "testing"

func TestControlTextPosYMathCentersCapBox(t *testing.T) {
	// Cap equals ascent → classic mid - half.
	y := controlTextPosYMath(10, 40, 20, 20)
	if y != 20 {
		t.Fatalf("equal cap/ascent: y=%d want 20 (mid-half)", y)
	}
	// Inflated ascent must still center the cap box on mid=30.
	y = controlTextPosYMath(10, 40, 12, 24)
	// mid=30; y = 30 + 6 - 24 = 12; cap top = 12+24-12 = 24; baseline = 36; mid caps = 30
	if y != 12 {
		t.Fatalf("inflated ascent: y=%d want 12", y)
	}
	capTop := float32(y) + 24 - 12
	baseline := float32(y) + 24
	midCaps := (capTop + baseline) / 2
	if midCaps < 29.5 || midCaps > 30.5 {
		t.Fatalf("cap mid=%v want ~30", midCaps)
	}
}
