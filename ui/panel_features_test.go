package ui

import "testing"

func TestPanelFeaturesDefaults(t *testing.T) {
	p := NewPanel("p", "Title", 0, 0, 320, 200)
	if p.panelFeatures == nil {
		t.Fatal("expected default PanelFeaturesBehavior")
	}
	f := p.Features()
	if !f.TitleBar {
		t.Fatal("default TitleBar should be true")
	}
	if f.Collapsible || f.Closable || f.Movable || f.Resizable {
		t.Fatal("optional features should default off")
	}
}

func TestPanelSetCollapsible(t *testing.T) {
	p := NewPanel("p", "Settings", 0, 0, 320, 0)
	p.AutoHeight = true
	p.SetCollapsible(true)
	lbl := NewLabel("lbl", "Body", 0, 0, 0, 0)
	lbl.AutoHeight = true
	p.AddChild(lbl)

	cb := p.CollapseBehavior()
	if cb == nil {
		t.Fatal("expected collapse behavior")
	}
	cb.Expanded.Set(false)

	root := NewContainer("root", 0, 0, 320, 600)
	root.LayoutType = LayoutAbsolute
	root.AddChild(p)
	root.Layout()

	titleH := p.bodyTitleHeight()
	if p.Bounds().Height > titleH+4 {
		t.Fatalf("collapsed height %.0f, want ~title %.0f", p.Bounds().Height, titleH)
	}
}

func TestPanelVScrollHost(t *testing.T) {
	p := NewPanel("p", "Scroll", 0, 0, 200, 120)
	p.SetVScroll(true)
	p.AddChild(NewLabel("l", "Line", 0, 0, 0, 0))
	if p.panelFeatures.scrollOuter == nil {
		t.Fatal("expected scroll viewport after SetVScroll")
	}
	if len(p.body.Children()) != 1 {
		t.Fatalf("body should contain scroll host, got %d children", len(p.body.Children()))
	}
}

func TestPanelFloatPositionRelativeToParent(t *testing.T) {
	lane := NewContainer("lane", 40, 80, 400, 320)
	lane.LayoutType = LayoutAbsolute

	p := NewPanel("float", "Float", 0, 0, 200, 120)
	p.SetFloatPosition(16, 24)
	lane.AddChild(p)

	root := NewContainer("root", 0, 0, 800, 600)
	root.LayoutType = LayoutAbsolute
	root.AddChild(lane)
	root.Layout()

	if p.Bounds().X != 56 || p.Bounds().Y != 104 {
		t.Fatalf("float bounds = (%v,%v), want (56,104) parent-relative", p.Bounds().X, p.Bounds().Y)
	}
}

func TestPanelScrollHostNoContentBleed(t *testing.T) {
	p := NewPanel("p", "Scroll", 0, 0, 200, 120)
	p.SetVScroll(true)
	if p.panelFeatures.scrollOuter == nil {
		t.Fatal("missing scroll host")
	}
	if p.panelFeatures.scrollOuter.ContentClipBleed != 0 {
		t.Fatalf("panel scroll bleed = %v, want 0", p.panelFeatures.scrollOuter.ContentClipBleed)
	}
}

func TestPanelEnableCollapseCompat(t *testing.T) {
	p := NewPanel("p", "Linked", 0, 0, 320, 0)
	cb := p.EnableCollapse(true)
	if cb == nil {
		t.Fatal("EnableCollapse should return behavior")
	}
	if !p.Features().Collapsible {
		t.Fatal("EnableCollapse should enable Collapsible feature")
	}
}
