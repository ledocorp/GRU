package ui

import (
	"fmt"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Mirrors examples/responsive_demo previewChromeScroll + FlexGrow cards.
// Fixed previewBandHeight avoids nested AutoHeight collapse; FlexGrow cards need a definite host height.
func TestResponsivePreviewFlexGrowCardsVisible(t *testing.T) {
	const (
		simW              float32 = 960
		previewBandHeight float32 = 320
	)

	chrome := NewCard("prev-chrome", "", 0, 0, 0, previewBandHeight)
	chrome.AutoHeight = false
	chrome.FlexDirection = FlexColumn
	chrome.Gap = 8
	chrome.ClipChildren = true

	scroll := NewHorizontalViewport("prev-scroll", 0, 0, 0, 0)
	scroll.AutoHeight = false
	scroll.SetFlexGrow(1)
	scroll.Gap = 0
	scroll.SetStyle("list-flush")

	lane := NewContainer("prev-lane", 0, 0, simW, 0)
	lane.AutoHeight = false
	lane.SetFlexGrow(1)
	lane.FlexDirection = FlexColumn
	lane.PreferredWidth = simW
	lane.MinWidth = simW
	lane.MaxWidth = simW
	lane.ClipChildren = true
	lane.SetStyle("transparent")
	lane.SetStyleOverrides(Style{Padding: 14})

	host := NewContainer("prev-host", 0, 0, 0, 0)
	host.AutoHeight = false
	host.SetFlexGrow(1)
	host.FlexDirection = FlexRow
	host.Gap = 10
	host.ClipChildren = true
	host.SetStyle("transparent")

	for i, title := range []string{"Alpha", "Beta", "Gamma"} {
		card := NewCard(fmt.Sprintf("prev-card-%d", i), "", 0, 0, 0, 88)
		card.AutoHeight = false
		card.SetFlexGrow(1)
		card.FlexDirection = FlexColumn
		card.Gap = 6
		card.ClipChildren = true
		card.AddChild(NewPlainText(fmt.Sprintf("prev-title-%d", i), "form-value", title, 0, 0, 0, 0))
		card.AddChild(NewPlainText(fmt.Sprintf("prev-sub-%d", i), "form-label", "FlexGrow 1", 0, 0, 0, 0))
		host.AddChild(card)
	}

	bottom := NewLabel("prev-bottom", "", 0, 0, 0, 12)
	bottom.SetStyle("transparent")
	lane.AddChild(host)
	lane.AddChild(bottom)
	scroll.AddChild(lane)
	chrome.AddChild(scroll)

	// Parent assigns a definite width like a panel body; height is fixed (Batch 8 band pattern).
	chrome.SetBounds(rl.NewRectangle(20, 40, 700, previewBandHeight))
	chrome.Layout()

	if host.Bounds().Height < 80 {
		t.Fatalf("host height=%.1f want >= 80 (collapsed FlexGrow preview)", host.Bounds().Height)
	}
	if scroll.Bounds().Height < 80 {
		t.Fatalf("scroll height=%.1f want >= 80", scroll.Bounds().Height)
	}
	if chrome.Bounds().Height < 100 {
		t.Fatalf("chrome height=%.1f want >= 100", chrome.Bounds().Height)
	}
	for _, ch := range host.Children() {
		b := ch.Bounds()
		if b.Height < 80 {
			t.Fatalf("card %s height=%.1f want >= 80", ch.ID(), b.Height)
		}
		if b.Width < 50 {
			t.Fatalf("card %s width=%.1f want share of row", ch.ID(), b.Width)
		}
	}
}
