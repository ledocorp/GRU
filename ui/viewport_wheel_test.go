package ui

import "testing"

func TestVerticalViewportHandlesWheelScroll(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 200, 100)
	if !vp.HandlesWheelScroll() {
		t.Fatal("vertical viewport should handle wheel scroll")
	}
	hp := NewHorizontalViewport("hp", 0, 0, 200, 100)
	if hp.HandlesWheelScroll() {
		t.Fatal("horizontal viewport should not claim vertical wheel")
	}
}

func TestNestedViewportAbsorbsParentWheelAtLimit(t *testing.T) {
	inner := NewViewport("inner-vp", 0, 0, 380, 300)
	inner.contentHeight = 900

	if !inner.AbsorbsParentWheel(-1) {
		t.Fatal("nested viewport with room to scroll down should absorb wheel")
	}
	inner.ScrollY = inner.overflowScrollY()
	if inner.AbsorbsParentWheel(-1) {
		t.Fatal("nested viewport at bottom should not absorb downward wheel")
	}
}
