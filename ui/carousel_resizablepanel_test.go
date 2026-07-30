package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestNormalizeRatios(t *testing.T) {
	got := normalizeRatios([]float32{1, 2, 1}, 3)
	want := []float32{0.25, 0.5, 0.25}
	for i := range want {
		if got[i] < want[i]-0.001 || got[i] > want[i]+0.001 {
			t.Fatalf("ratio[%d] = %v, want %v", i, got[i], want[i])
		}
	}
	equal := normalizeRatios(nil, 4)
	for _, r := range equal {
		if r < 0.24 || r > 0.26 {
			t.Fatalf("expected equal split, got %v", equal)
		}
	}
}

func TestResizablePanelPaneLayout(t *testing.T) {
	p0 := NewLabel("p0", "A", 0, 0, 0, 20)
	p1 := NewLabel("p1", "B", 0, 0, 0, 20)
	p2 := NewLabel("p2", "C", 0, 0, 0, 20)
	rp := NewResizablePanel("rp", SplitHorizontal, []Node{p0, p1, p2},
		[]float32{0.25, 0.5, 0.25}, 0, 0, 400, 200)
	rp.SetBounds(rl.NewRectangle(0, 0, 400, 200))
	rp.Layout()

	var totalW float32
	for _, p := range []Node{p0, p1, p2} {
		totalW += p.Bounds().Width
	}
	totalW += float32(rp.splitterCount()) * rp.SplitterW
	if totalW < 399 || totalW > 401 {
		t.Fatalf("pane widths + splitters = %v, want ~400", totalW)
	}
	if p1.Bounds().Width <= p0.Bounds().Width {
		t.Fatalf("center pane should be widest: %v vs %v", p1.Bounds().Width, p0.Bounds().Width)
	}
	right := p2.Bounds().X + p2.Bounds().Width
	if right < 399 || right > 401 {
		t.Fatalf("right pane should flush host edge, got right=%.0f", right)
	}
}

func TestResizablePanelNarrowWidthDoesNotJamRightPane(t *testing.T) {
	tree := NewPanel("tree", "Tree", 0, 0, 0, 0)
	editor := NewPanel("editor", "Editor", 0, 0, 0, 0)
	props := NewPanel("props", "Props", 0, 0, 0, 0)
	rp := NewResizablePanel("rp", SplitHorizontal, []Node{tree, editor, props},
		[]float32{0.22, 0.56, 0.22}, 0, 0, 360, 200)
	rp.MinSizes = []float32{100, 120, 100}
	rp.SetBounds(rl.NewRectangle(0, 0, 360, 200))
	rp.Layout()

	if props.Bounds().Width < 40 {
		t.Fatalf("props pane too narrow: %.0f", props.Bounds().Width)
	}
	right := props.Bounds().X + props.Bounds().Width
	if right < 359 || right > 361 {
		t.Fatalf("props pane right edge %.0f, want ~360", right)
	}
}

func TestResizablePanelLayoutNotRootDirty(t *testing.T) {
	root := NewContainer("root", 0, 0, 800, 600)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	p0 := NewLabel("p0", "A", 0, 0, 0, 20)
	p1 := NewLabel("p1", "B", 0, 0, 0, 20)
	p2 := NewLabel("p2", "C", 0, 0, 0, 20)
	rp := NewResizablePanel("rp", SplitHorizontal, []Node{p0, p1, p2},
		[]float32{0.25, 0.5, 0.25}, 0, 0, 800, 400)
	root.AddChild(rp)
	root.SetBounds(rl.NewRectangle(0, 0, 800, 600))
	root.Layout()
	root.layoutDirty = false
	rp.Layout()
	if root.IsDirty() {
		t.Fatal("layoutSetBounds should not re-dirty root on unchanged split layout")
	}
}

func TestCarouselLayoutActiveSlide(t *testing.T) {
	idx := NewSignal(1)
	s0 := NewLabel("s0", "A", 0, 0, 0, 20)
	s1 := NewLabel("s1", "B", 0, 0, 0, 20)
	c := NewCarousel("c", []Node{s0, s1}, idx, 0, 0, 300, 200)
	c.SetBounds(rl.NewRectangle(0, 0, 300, 200))
	c.Layout()
	if s1.Bounds().Width <= 0 {
		t.Fatal("active slide should have width")
	}
	if s0.Bounds().Width != 0 {
		t.Fatalf("inactive slide should be zero width, got %v", s0.Bounds().Width)
	}
}

func TestCarouselLayoutTallSlideUsesFullViewport(t *testing.T) {
	slide := NewCard("slide", "Manual only", 0, 0, 0, 0)
	slide.TitleHeight = 36
	slide.Gap = 14
	slide.AddChild(NewPlainText("body", "form-value",
		"AutoPlayInterval = 0 keeps slides until you click. "+
			"Extra lines force the card taller than a short carousel viewport.", 0, 0, 0, 0))
	c := NewCarousel("c", []Node{slide}, NewSignal(0), 0, 0, 200, 120)
	c.SetBounds(rl.NewRectangle(0, 0, 200, 120))
	c.Layout()
	vp := c.slideViewport()
	if slide.Bounds().Height < vp.Height-1 {
		t.Fatalf("slide height %.0f, want full viewport %.0f", slide.Bounds().Height, vp.Height)
	}
	if slide.Bounds().Y > vp.Y+0.5 {
		t.Fatalf("tall slide Y=%.0f, want top-aligned at %.0f", slide.Bounds().Y, vp.Y)
	}
}

func TestCarouselSameHeightSameSlideLayout(t *testing.T) {
	const h = float32(180)
	s1 := carouselTestSlide("carousel-s1", "Welcome", "Build desktop UIs in Go.")
	m1 := carouselTestSlide("carousel-m1", "Manual only", "AutoPlayInterval = 0 keeps slides until you click.")
	hero := NewCarousel("hero", []Node{s1}, NewSignal(0), 0, 0, 720, h)
	manual := NewCarousel("manual", []Node{m1}, NewSignal(0), 0, 0, 720, h)
	hero.SetBounds(rl.NewRectangle(0, 0, 720, h))
	manual.SetBounds(rl.NewRectangle(0, 0, 720, h))
	hero.Layout()
	manual.Layout()
	if s1.Bounds() != m1.Bounds() {
		t.Fatalf("hero vs manual slide bounds differ: hero %+v manual %+v", s1.Bounds(), m1.Bounds())
	}
}

func TestCarouselLayoutMediumSlideKeepsIntrinsicHeight(t *testing.T) {
	slide := NewCard("slide", "Compose real screens", 0, 0, 0, 0)
	slide.TitleHeight = 36
	slide.Gap = 20
	slide.AddChild(NewPlainText("body", "form-value",
		"Demos prove widgets work together — flex, scroll, signals, and overlays.", 0, 0, 0, 0))
	c := NewCarousel("c", []Node{slide}, NewSignal(0), 0, 0, 300, 240)
	c.SetBounds(rl.NewRectangle(0, 0, 300, 240))
	c.Layout()
	band := carouselSlideBand(c.slideViewport())
	intrinsic := slide.Bounds().Height
	if intrinsic >= band.Height-1 {
		t.Fatalf("medium slide stretched to band %.0f", band.Height)
	}
	if intrinsic < 40 {
		t.Fatalf("medium slide too short: %.0f", intrinsic)
	}
}

func carouselDemoSlides() []Node {
	return []Node{
		carouselTestSlide("carousel-s1", "Welcome to Gru", "Build desktop and mobile UIs in Go."),
		carouselTestSlide("carousel-s2", "Compose real screens", "Demos prove widgets work together."),
		carouselTestSlide("carousel-s3", "Ship faster", "Theme tokens and a single package ui."),
	}
}

func carouselTestSlide(id, title, body string) *Card {
	c := NewCard(id, title, 0, 0, 0, 0)
	c.TitleHeight = 36
	c.Gap = 20
	c.AddChild(NewPlainText(id+"-body", "form-value", body, 0, 0, 0, 0))
	return c
}

func TestCarouselIdleReadyAfterLayout(t *testing.T) {
	idx := NewSignal(0)
	hero := NewCarousel("carousel-hero", carouselDemoSlides(), idx, 0, 0, 0, 180)
	hero.SetFlexGrow(0)

	manualIdx := NewSignal(0)
	manual := NewCarousel("carousel-manual", []Node{
		carouselTestSlide("carousel-m1", "Manual only", "AutoPlayInterval = 0 keeps slides until you click."),
		carouselTestSlide("carousel-m2", "Dots optional", "ShowDots stays on by default for wayfinding."),
	}, manualIdx, 0, 0, 0, 180) // same band height as hero

	vp := NewViewport("page-scroll", 0, 0, 520, 640)
	vp.SetStyle("page-scroll")
	vp.FlexDirection = FlexColumn
	vp.AddChild(hero)
	vp.AddChild(manual)

	root := NewContainer("root", 0, 0, 520, 640)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	root.SetBounds(rl.NewRectangle(0, 0, 520, 640))
	root.AddChild(vp)

	root.Layout()
	SimulateCacheHitFrame(root)
	AssertIdleReady(t, root, "carousel page")

	m2 := manual.Slides[1]
	if !m2.IsHidden() {
		t.Fatal("inactive carousel-m2 should be hidden so it cannot pin idle FPS")
	}
	if SubtreeNeedsRedraw(m2) {
		t.Fatalf("hidden inactive slide still needs redraw: %s", NotIdleReason(m2))
	}
}

func TestCarouselIndexWrap(t *testing.T) {
	idx := NewSignal(0)
	c := NewCarousel("c", []Node{
		NewLabel("a", "A", 0, 0, 0, 20),
		NewLabel("b", "B", 0, 0, 0, 20),
	}, idx, 0, 0, 200, 120)
	c.setIndex(-1)
	if idx.Get() != 1 {
		t.Fatalf("index = %d, want 1", idx.Get())
	}
	c.setIndex(2)
	if idx.Get() != 0 {
		t.Fatalf("index = %d, want 0", idx.Get())
	}
}
