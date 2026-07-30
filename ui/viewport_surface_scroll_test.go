package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Preview markdown tables live in Cards; viewport repositionOnly must shift the
// internal RaisedSurface body or header rows paint at stale Y (bleed over chrome).
func TestViewportRepositionOnlyShiftsCardBody(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 400, 300)
	vp.SetStyle("preview-scroll")
	vp.ContentClipBleed = 0

	lane := NewContainer("lane", 0, 0, 0, 0)
	lane.LayoutType = LayoutFlex
	lane.FlexDirection = FlexColumn
	lane.AutoHeight = true
	lane.SetStyle("transparent")

	card := NewCard("table-card", "", 0, 0, 0, 0)
	card.AutoHeight = true
	card.Title = ""
	card.TitleHeight = 0
	card.Gap = 0
	card.SetStyleVariant("card", "table")
	row := NewContainer("hdr", 0, 0, 0, 32)
	row.SetStyle("table-header-row")
	card.AddChild(row)
	lane.AddChild(card)
	vp.AddChild(lane)

	vp.SetBounds(vp.Bounds())
	vp.Layout()

	bodyBefore := card.body.Bounds()

	vp.ScrollY = 80
	vp.scrollDirty = true
	vp.MarkDirty()
	vp.Layout()

	bodyAfter := card.body.Bounds()
	wantY := bodyBefore.Y - 80
	if absF(bodyAfter.Y-wantY) > 0.5 {
		t.Fatalf("card body Y after scroll = %.1f, want %.1f (before=%.1f)", bodyAfter.Y, wantY, bodyBefore.Y)
	}
}

func TestViewportFloatOverlayClipUsesContentScissor(t *testing.T) {
	vp := NewViewport("vp", 40, 60, 320, 200)
	vp.SetStyle("preview-scroll")
	vp.SetBounds(vp.Bounds())
	vp.Layout()

	clip := vp.floatOverlayClipRect()
	innerX, innerY, innerW, innerH := vp.contentScissorRect()
	if clip.X != innerX || clip.Y != innerY || clip.Width != innerW || clip.Height != innerH {
		t.Fatalf("float overlay clip %v != content scissor %v",
			clip, rl.NewRectangle(innerX, innerY, innerW, innerH))
	}
}
