package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestNestedViewportClipIntersectsParent(t *testing.T) {
	root := NewContainer("root", 0, 0, 400, 400)
	outer := NewViewport("outer", 0, 0, 200, 120)
	outer.SetStyle("preview-scroll")
	inner := NewHorizontalViewport("inner", 0, 0, 180, 80)
	outer.AddChild(inner)
	root.AddChild(outer)
	root.Layout()

	// Place outer so its content band is mid-window, then scroll the outer so
	// the inner strip would extend above the outer clip if not intersected.
	outer.SetBounds(rl.NewRectangle(40, 80, 200, 120))
	outer.Layout()
	inner.SetBounds(rl.NewRectangle(50, 40, 180, 80)) // partly above outer

	clip := intersectRectsWithViewportAncestors(inner.ClipBounds(), inner)
	parent := outer.ClipBounds()
	if clip.Y < parent.Y-0.5 {
		t.Fatalf("nested clip Y=%.1f escapes parent Y=%.1f", clip.Y, parent.Y)
	}
	if clip.Y+clip.Height > parent.Y+parent.Height+0.5 {
		t.Fatalf("nested clip bottom escapes parent")
	}
}
