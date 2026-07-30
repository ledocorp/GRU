package ui

import (
	"testing"
)

// Regression: scroll host must match visible body band (panel features demo).
func TestPanelVScrollViewportFitsBody(t *testing.T) {
	p := NewPanel("scroll", "Scrollable body", 0, 0, 280, 220)
	p.SetVScroll(true)
	for i := 1; i <= 8; i++ {
		lbl := NewLabel("line", "Line content", 0, 0, 0, 0)
		lbl.SetStyle("form-value")
		lbl.AutoHeight = true
		p.AddChild(lbl)
	}

	grid := NewContainer("grid", 0, 0, 900, 600)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 12
	grid.Gap = 12
	p.SetColSpan(BreakpointXS, 12)
	grid.AddChild(p)

	root := NewContainer("root", 0, 0, 900, 600)
	root.LayoutType = LayoutAbsolute
	root.AddChild(grid)
	root.Layout()

	vp := p.panelFeatures.scrollOuter
	if vp == nil {
		t.Fatal("missing scroll viewport")
	}

	titleH := p.bodyTitleHeight()
	pad := p.GetStyle().Padding
	bodyBottom := p.Bounds().Y + p.Bounds().Height
	maxVPBottom := bodyBottom - pad + 0.5
	vpBottom := vp.Bounds().Y + vp.Bounds().Height
	if vpBottom > maxVPBottom {
		t.Fatalf("scroll viewport bottom %.1f exceeds body inner bottom %.1f (panel h=%.0f title=%.0f pad=%.0f)",
			vpBottom, maxVPBottom, p.Bounds().Height, titleH, pad)
	}

	clip := vp.ClipBounds()
	if clip.Y+clip.Height > bodyBottom-pad+1 {
		t.Fatalf("viewport clip extends past body: clip bottom=%.1f body inner bottom=%.1f",
			clip.Y+clip.Height, bodyBottom-pad)
	}
}

func TestPanelVScrollLabelUsesViewportClip(t *testing.T) {
	p := NewPanel("scroll", "Scroll", 0, 0, 200, 120)
	p.SetVScroll(true)
	lbl := NewLabel("l", "Overflow line", 0, 0, 0, 0)
	lbl.AutoHeight = true
	p.AddChild(lbl)

	root := NewContainer("root", 0, 0, 200, 120)
	root.LayoutType = LayoutAbsolute
	root.AddChild(p)
	root.Layout()

	clip, ok := ancestorClipBounds(lbl)
	if !ok {
		t.Fatal("label inside scroll viewport should see ancestor viewport clip")
	}
	vpClip := p.panelFeatures.scrollOuter.ClipBounds()
	if clip.Y != vpClip.Y || clip.Height != vpClip.Height {
		t.Fatalf("label ancestor clip should match viewport content clip")
	}
}
