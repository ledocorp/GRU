package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestSplitViewRelayoutsOnBoundsChange(t *testing.T) {
	left := NewLabel("left", "Left", 0, 0, 0, 24)
	right := NewLabel("right", "Right", 0, 0, 0, 24)
	sv := NewSplitView("sv", SplitHorizontal, left, right, 0, 0, 400, 200)
	sv.Layout()

	w1 := left.Bounds().Width
	sv.SetBounds(rl.NewRectangle(0, 0, 600, 280))
	sv.Layout()
	w2 := left.Bounds().Width
	if w2 <= w1+20 {
		t.Fatalf("left pane width %.0f should grow after split resize (was %.0f)", w2, w1)
	}
}

func TestAutoHeightPanelFlexGrowSplitNotProbeHeight(t *testing.T) {
	panel := NewPanel("p", "Split", 0, 0, 600, 0)
	panel.AutoHeight = true
	panel.TitleHeight = 32

	help := NewLabel("help", "Drag the splitter.", 0, 0, 0, 18)
	panel.AddChild(help)

	left := NewLabel("left", "Files", 0, 0, 0, 24)
	right := NewLabel("right", "Editor", 0, 0, 0, 24)
	sv := NewSplitView("sv", SplitHorizontal, left, right, 0, 0, 0, 0)
	sv.SetFlexGrow(1)
	panel.AddChild(sv)

	panel.SetBounds(rl.NewRectangle(0, 0, 600, 4096))
	panel.Layout()

	if sv.Bounds().Height > 400 {
		t.Fatalf("split height %.0f should not fill 4096 probe band", sv.Bounds().Height)
	}
	if panel.Bounds().Height > 500 {
		t.Fatalf("panel height %.0f should shrink-wrap, not probe-tall", panel.Bounds().Height)
	}
}
