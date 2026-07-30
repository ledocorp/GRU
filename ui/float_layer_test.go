package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestFloatLayerViewportOverlay(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 400, 300)
	vp.SetStyle("transparent")

	tall := NewPanel("tall", "Content", 0, 0, 0, 0)
	tall.AutoHeight = true
	tall.AddChild(NewLabel("lbl", "Body", 0, 0, 0, 0))

	fl := NewFloatLayer("float")
	win := NewPanel("win", "Float", 40, 60, 200, 120)
	win.SetFloatPosition(40, 60)
	fl.AddChild(win)

	vp.AddChild(tall)
	vp.AddChild(fl)
	vp.Layout()

	flBounds := fl.Bounds()
	want := vp.floatOverlayHostRect()
	if flBounds != want {
		t.Fatalf("float layer bounds %v, want viewport host %v", flBounds, want)
	}

	before := vp.contentHeight
	vp.Layout()
	if vp.contentHeight != before {
		t.Fatalf("float layer should not affect scroll content height: got %.0f want %.0f", vp.contentHeight, before)
	}

	winBounds := win.Bounds()
	if winBounds.X != flBounds.X+40 || winBounds.Y != flBounds.Y+60 {
		t.Fatalf("float panel at %v, want (%.0f, %.0f)", winBounds, flBounds.X+40, flBounds.Y+60)
	}
}

func TestCollapseChromeCollapsed(t *testing.T) {
	p := NewPanel("p", "Title", 0, 0, 280, 0)
	p.AutoHeight = true
	p.AddChild(NewLabel("lbl", "Body", 0, 0, 0, 0))
	cb := p.EnableCollapse(true)
	cb.Expanded.Set(false)

	root := NewContainer("root", 0, 0, 320, 400)
	root.LayoutType = LayoutAbsolute
	root.AddChild(p)
	root.Layout()

	if !p.collapseChromeCollapsed() {
		t.Fatal("expected collapsed chrome flag when fully collapsed")
	}

	cb.Expanded.Set(true)
	root.Update(0.2)
	root.Layout()
	if p.collapseChromeCollapsed() {
		t.Fatal("collapsed chrome flag should be off when expanded")
	}
}

func TestPanelResizePreservesHeight(t *testing.T) {
	lane := NewFloatLayer("lane")
	lane.SetBounds(rl.NewRectangle(0, 0, 480, 360))

	p := NewPanel("p", "Resizable", 40, 40, 280, 220)
	p.SetResizable(true).SetMovable(true).SetConstrain(true)
	p.AddChild(NewLabel("lbl", "Body", 0, 0, 0, 0))
	lane.AddChild(p)

	root := NewContainer("root", 0, 0, 480, 360)
	root.LayoutType = LayoutAbsolute
	root.AddChild(lane)
	root.Layout()

	pf := p.panelFeatures
	if pf == nil {
		t.Fatal("missing panel features")
	}
	pf.resizing = true
	pf.resizeEdge = panelResizeS
	pf.resizeStartMX = 0
	pf.resizeStartMY = 0
	pf.resizeStartX = p.Bounds().X
	pf.resizeStartY = p.Bounds().Y
	pf.resizeStartW = p.Bounds().Width
	pf.resizeStartH = p.Bounds().Height
	pf.userPositioned = true
	pf.applyResize(rl.NewVector2(0, 48))
	root.Layout()

	if p.Bounds().Height < 250 {
		t.Fatalf("vertical resize should stick after layout, height=%.0f", p.Bounds().Height)
	}
}
