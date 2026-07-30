package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestSurfaceShellBodyIsDumbBox(t *testing.T) {
	p := NewPanel("shell", "Title", 0, 0, 200, 120)
	if p.body == nil {
		t.Fatal("missing body")
	}
	if p.bodyTitleHeight() == 0 {
		t.Fatal("titled shell should reserve header band")
	}
	if len(p.body.Children()) != len(p.Children()) {
		t.Fatal("Children() should expose body content only")
	}
}

func TestSurfaceShellUntitledBodyFillsShell(t *testing.T) {
	c := NewCard("untitled", "", 0, 0, 200, 100)
	c.Layout()
	if c.bodyTitleHeight() != 0 {
		t.Fatalf("title band = %v, want 0", c.bodyTitleHeight())
	}
	body := c.body.Bounds()
	if body.Y != c.Bounds().Y {
		t.Fatalf("body Y=%v shell Y=%v, want same when untitled", body.Y, c.Bounds().Y)
	}
	if body.Height != c.Bounds().Height {
		t.Fatalf("body H=%v shell H=%v", body.Height, c.Bounds().Height)
	}
}

func TestSurfaceShellLayoutBoundary(t *testing.T) {
	p := NewPanel("p", "Panel", 0, 0, 240, 160)
	rt := NewRichText("copy", []TextSpan{{Text: "Body copy wraps inside the dumb box."}}, 0, 0, 0, 0)
	p.AddChild(rt)
	p.SetBounds(rl.NewRectangle(0, 0, 240, 160))
	p.Layout()
	if rt.Bounds().Y <= p.Bounds().Y {
		t.Fatalf("rich text should sit below header: text Y=%v shell Y=%v", rt.Bounds().Y, p.Bounds().Y)
	}
}

func TestNestedSurfaceShadowSkipThroughShell(t *testing.T) {
	outer := NewPanel("outer", "Outer", 0, 0, 400, 200)
	inner := NewCard("inner", "Inner", 0, 0, 200, 80)
	outer.AddChild(inner)
	if !nestedInRaisedSurface(inner) {
		t.Fatal("nested card should report nestedInRaisedSurface")
	}
	if nestedInRaisedSurface(outer.body) {
		t.Fatal("outer body should not count as nested")
	}
}

func TestPanelCardFacadeSameLayoutHeaderless(t *testing.T) {
	const w, h = float32(320), float32(120)
	panel := NewPanel("p", "", 0, 0, w, h)
	card := NewCard("c", "", 0, 0, w, h)
	lblP := NewLabel("lp", "Same copy", 0, 0, 0, 24)
	lblC := NewLabel("lc", "Same copy", 0, 0, 0, 24)
	panel.AddChild(lblP)
	card.AddChild(lblC)
	panel.SetBounds(rl.NewRectangle(0, 0, w, h))
	card.SetBounds(rl.NewRectangle(0, 0, w, h))
	panel.Layout()
	card.Layout()
	if lblP.Bounds() != lblC.Bounds() {
		t.Fatalf("label bounds differ: panel %+v card %+v", lblP.Bounds(), lblC.Bounds())
	}
}

func TestGlassHeaderModeFromChromeKind(t *testing.T) {
	p := NewPanel("g", "Glass", 0, 0, 200, 100)
	if err := p.SetStylePreset("glass-panel", PresetProps{}); err != nil {
		t.Fatal(err)
	}
	h := p.surfaceHeader()
	if h.Mode != HeaderModeGlass {
		t.Fatalf("mode = %v, want HeaderModeGlass", h.Mode)
	}
	if !h.DefersUntilPostSheen() {
		t.Fatal("glass preset should defer title")
	}
}
