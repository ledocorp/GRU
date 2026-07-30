package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Mirrors examples/responsive_demo.go preview stack for layout regression tests.
func layoutPreviewMetricGrid(t *testing.T, simWidth, hostW float32) (*Card, *Card) {
	t.Helper()
	chrome := NewCard("chrome", "", 0, 0, hostW, 0)
	chrome.AutoHeight = true
	chrome.FlexDirection = FlexColumn
	chrome.Gap = 8
	chrome.ClipChildren = true
	chrome.AddChild(NewLabel("cap", "caption", 0, 0, 0, 0))

	scroll := NewHorizontalViewport("scroll", 0, 0, 0, 0)
	scroll.AutoHeight = true
	scroll.SetStyle("list-flush")

	lane := NewContainer("lane", 0, 0, simWidth, 0)
	lane.AutoHeight = true
	lane.FlexDirection = FlexColumn
	lane.PreferredWidth = simWidth
	lane.MinWidth = simWidth
	lane.MaxWidth = simWidth
	lane.ClipChildren = true
	lane.SetStyle("transparent")

	host := NewContainer("host", 0, 0, 0, 0)
	host.AutoHeight = true
	host.FlexDirection = FlexColumn
	host.ClipChildren = true
	host.SetStyle("transparent")
	lane.AddChild(host)
	lane.AddChild(NewLabel("bottom", "", 0, 0, 0, 12))

	inner := NewContainer("grid", 0, 0, 0, 0)
	inner.LayoutType = LayoutGrid
	inner.GridColumns = 12
	inner.Gap = 10
	inner.ClipChildren = true
	inner.SetStyle("transparent")

	card := NewCard("metric", "One", 0, 0, 0, 0)
	card.AutoHeight = true
	card.FlexDirection = FlexColumn
	card.Gap = 4
	card.ClipChildren = true
	card.AddChild(NewLabel("title", "One", 0, 0, 0, 0))
	card.AddChild(NewLabel("note", "xs12 sm6 md4 lg3", 0, 0, 0, 0))
	card.SetColSpan(BreakpointMD, 4)
	inner.AddChild(card)

	host.AddChild(inner)
	scroll.AddChild(lane)
	chrome.AddChild(scroll)

	chrome.SetBounds(rl.NewRectangle(0, 0, hostW, 4096))
	chrome.Layout()
	return chrome, card
}

func layoutPreviewFlexGrow(t *testing.T, simWidth, hostW float32) (*Card, []*Card) {
	t.Helper()
	chrome := NewCard("chrome", "", 0, 0, hostW, 0)
	chrome.AutoHeight = true
	chrome.FlexDirection = FlexColumn
	chrome.Gap = 8
	chrome.ClipChildren = true
	chrome.AddChild(NewLabel("cap", "caption", 0, 0, 0, 0))

	scroll := NewHorizontalViewport("scroll", 0, 0, 0, 0)
	scroll.AutoHeight = true
	scroll.SetStyle("list-flush")

	lane := NewContainer("lane", 0, 0, simWidth, 0)
	lane.AutoHeight = true
	lane.FlexDirection = FlexColumn
	lane.PreferredWidth = simWidth
	lane.MinWidth = simWidth
	lane.MaxWidth = simWidth
	lane.ClipChildren = true
	lane.SetStyle("transparent")

	host := NewContainer("host", 0, 0, 0, 0)
	host.AutoHeight = true
	host.FlexDirection = FlexRow
	host.Gap = 10
	host.ClipChildren = true
	host.SetStyle("transparent")

	var cards []*Card
	for _, title := range []string{"Alpha", "Beta", "Gamma"} {
		c := NewCard("flex-"+title, "", 0, 0, 0, 88)
		c.AutoHeight = false
		c.SetFlexGrow(1)
		c.FlexDirection = FlexColumn
		c.Gap = 6
		c.ClipChildren = true
		c.AddChild(NewLabel("t", title, 0, 0, 0, 0))
		c.AddChild(NewLabel("s", "FlexGrow 1", 0, 0, 0, 0))
		host.AddChild(c)
		cards = append(cards, c)
	}

	lane.AddChild(host)
	lane.AddChild(NewLabel("bottom", "", 0, 0, 0, 12))
	scroll.AddChild(lane)
	chrome.AddChild(scroll)

	chrome.SetBounds(rl.NewRectangle(0, 0, hostW, 4096))
	chrome.Layout()
	return chrome, cards
}

func TestPreviewMetricCardHasVisibleBody(t *testing.T) {
	_, metric := layoutPreviewMetricGrid(t, 960, 800)
	b := metric.Bounds()
	if b.Height < 24 {
		t.Fatalf("metric card height %.0f, want readable body", b.Height)
	}
	for _, ch := range metric.Children() {
		if ch.IsHidden() {
			continue
		}
		lb := ch.Bounds()
		if lb.Height < 8 {
			t.Fatalf("label height %.0f, bounds %+v", lb.Height, lb)
		}
	}
}

func TestPreviewFlexGrowInPanel(t *testing.T) {
	panel := NewPanel("panel", "Flex", 0, 0, 600, 0)
	panel.AutoHeight = true
	panel.Gap = 10
	panel.TitleHeight = 36

	chrome, cards := layoutPreviewFlexGrow(t, 960, 600)
	panel.AddChild(chrome)
	panel.SetBounds(rl.NewRectangle(0, 0, 600, 4096))
	panel.Layout()

	scroll, ok := chrome.Children()[1].(*Viewport)
	if !ok {
		t.Fatalf("expected viewport, got %T", chrome.Children()[1])
	}
	if scroll.Bounds().Height < 40 {
		t.Fatalf("viewport height %.0f in panel", scroll.Bounds().Height)
	}
	for i, c := range cards {
		b := c.Bounds()
		if b.Width < 50 || b.Height < 40 {
			t.Fatalf("card %d bounds %+v", i, b)
		}
	}
}

func TestPreviewFlexGrowCardsVisible(t *testing.T) {
	chrome, cards := layoutPreviewFlexGrow(t, 960, 800)
	scroll, ok := chrome.Children()[1].(*Viewport)
	if !ok {
		t.Fatalf("expected viewport child, got %T", chrome.Children()[1])
	}
	if scroll.Bounds().Height < 40 {
		t.Fatalf("viewport height %.0f, want room for 88px cards", scroll.Bounds().Height)
	}
	for i, c := range cards {
		b := c.Bounds()
		if b.Width < 50 {
			t.Fatalf("card %d width %.0f", i, b.Width)
		}
		if b.Height < 40 {
			t.Fatalf("card %d height %.0f", i, b.Height)
		}
	}
}
