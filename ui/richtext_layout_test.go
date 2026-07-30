package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestRichTextInvalidatesMeasureOnSetBoundsNoMarkWidth(t *testing.T) {
	rt := NewRichText("rt", []TextSpan{{Text: "Hello"}}, 0, 0, 280, 0)
	rt.Wrap = true
	rt.AutoHeight = true
	rt.SetBounds(rl.NewRectangle(0, 0, 280, 0))
	rt.Layout()
	if rt.lastMeasuredW != 280 {
		t.Fatalf("lastMeasuredW = %v, want 280 after layout", rt.lastMeasuredW)
	}

	rt.setBoundsNoMark(rl.NewRectangle(0, 0, 140, rt.Bounds().Height))
	if rt.lastMeasuredW != 0 {
		t.Fatalf("lastMeasuredW = %v, want 0 after setBoundsNoMark width shrink", rt.lastMeasuredW)
	}
}
