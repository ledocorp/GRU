package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestHorizontalViewportClipBoundsMatchesContentScissor(t *testing.T) {
	vp := NewHorizontalViewport("hv", 40, 60, 320, 180)
	vp.SetBounds(rl.NewRectangle(40, 60, 320, 180))
	vp.scrollbarHeight = 10

	x, y, w, h := vp.contentScissorRect()
	want := rl.NewRectangle(x, y, w, h)
	got := vp.ClipBounds()
	if got.X != want.X || got.Y != want.Y || got.Width != want.Width || got.Height != want.Height {
		t.Fatalf("ClipBounds=%+v contentScissor=%+v", got, want)
	}
}

func TestContainerCardChildExtendsDrawClipBottom(t *testing.T) {
	host := NewContainer("host", 0, 0, 200, 100)
	host.SetStyleOverrides(Style{Padding: 8})
	card := NewCard("c", "x", 0, 0, 0, 40)
	host.AddChild(card)

	clip := containerContentDrawClip(host.Bounds(), 8, host.children)
	tight := rl.NewRectangle(8, 8, 184, 84)
	if clip.Height <= tight.Height {
		t.Fatalf("card host clip should extend past tight bottom inset: clip=%+v tight=%+v", clip, tight)
	}
	if clip.Width > tight.Width+0.5 {
		t.Fatalf("card host clip should not extend width (symmetric padding): clip=%+v tight=%+v", clip, tight)
	}
}
