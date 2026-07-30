package ui

import "testing"

func TestCollapseBehaviorCollapsesAutoHeightPanel(t *testing.T) {
	p := NewPanel("p", "Settings", 0, 0, 320, 0)
	p.AutoHeight = true
	lbl := NewLabel("lbl", "Body copy", 0, 0, 0, 0)
	lbl.AutoHeight = true
	p.AddChild(lbl)
	cb := p.EnableCollapse(true)
	cb.Expanded.Set(false)

	root := NewContainer("root", 0, 0, 320, 600)
	root.LayoutType = LayoutAbsolute
	root.AddChild(p)
	root.Layout()

	titleH := p.bodyTitleHeight()
	if p.Bounds().Height > titleH+4 {
		t.Fatalf("collapsed height %.0f, want ~title %.0f", p.Bounds().Height, titleH)
	}

	cb.Expanded.Set(true)
	root.Update(0.2)
	root.Layout()
	if p.Bounds().Height <= titleH+4 {
		t.Fatalf("expanded height %.0f should exceed title %.0f", p.Bounds().Height, titleH)
	}
}

func TestHeaderBandStandaloneHeight(t *testing.T) {
	hb := NewHeaderBand("hb", "Standalone strip", 0, 0, 320)
	if hb.Bounds().Height != 40 {
		t.Fatalf("height = %.0f, want 40", hb.Bounds().Height)
	}
	hb.Layout()
	if hb.header().Height() != 40 {
		t.Fatalf("header band = %.0f", hb.header().Height())
	}
}

func TestHeaderBandModeNone(t *testing.T) {
	hb := NewHeaderBand("hb", "No chrome", 0, 0, 200)
	hb.SetHeaderMode(HeaderModeNone)
	hb.Layout()
	if hb.header().Height() != 0 {
		t.Fatal("HeaderModeNone should reserve 0 header height")
	}
}

func TestCollapseExternalHeaderDisablesShellOverlay(t *testing.T) {
	p := NewPanel("p", "Linked", 0, 0, 320, 0)
	cb := p.EnableCollapse(true)
	cb.ExternalHeader = true
	// PanelFeatures draws header chrome; legacy collapse chevron overlay is skipped.
	if cb.shell != nil && cb.shell.panelFeatures == nil {
		if p.IsInteractive() {
			t.Fatal("shell should not be interactive when ExternalHeader set without PanelFeatures")
		}
	}
}
