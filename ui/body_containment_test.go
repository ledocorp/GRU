package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestNarrowWidthOpenAccordionContained(t *testing.T) {
	panel := NewPanel("p", "Accordion", 0, 0, 320, 0)
	panel.AutoHeight = true
	panel.TitleHeight = 32

	acc := NewAccordion("acc", "Section", 0, 0, 0, 0)
	acc.Expanded.Set(true)
	long := "This is a long paragraph that must wrap inside a narrow panel body without spilling past the chrome border at minimum viewport width."
	acc.AddChild(NewLabel("lbl", long, 0, 0, 0, 48))
	panel.AddChild(acc)

	root := NewContainer("root", 0, 0, 320, 600)
	root.LayoutType = LayoutAbsolute
	root.AddChild(panel)
	root.Layout()

	panelBottom := panel.Bounds().Y + panel.Bounds().Height
	if bottom := nodeSubtreeBottom(acc); bottom > panelBottom+2 {
		t.Fatalf("open accordion subtree bottom %.0f below panel bottom %.0f", bottom, panelBottom)
	}
	if acc.Bounds().Width > panel.Bounds().Width+1 {
		t.Fatalf("accordion width %.0f exceeds panel %.0f", acc.Bounds().Width, panel.Bounds().Width)
	}
}

func TestNarrowWidthGridPanelContained(t *testing.T) {
	grid := NewContainer("grid", 0, 0, 320, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 1
	grid.AutoHeight = true
	grid.Gap = 12

	panel := NewPanel("p", "Grid cell", 0, 0, 0, 0)
	panel.AutoHeight = true
	panel.TitleHeight = 32
	acc := NewAccordion("acc", "Details", 0, 0, 0, 0)
	acc.Expanded.Set(true)
	acc.AddChild(NewLabel("lbl", "Wrapped content in a single-column grid at 320px minimum width.", 0, 0, 0, 40))
	panel.AddChild(acc)
	grid.AddChild(panel)

	root := NewContainer("root", 0, 0, 320, 800)
	root.LayoutType = LayoutAbsolute
	root.AddChild(grid)
	root.Layout()

	panelBottom := panel.Bounds().Y + panel.Bounds().Height
	if bottom := nodeSubtreeBottom(acc); bottom > panelBottom+2 {
		t.Fatalf("grid panel accordion bottom %.0f below panel %.0f", bottom, panelBottom)
	}
}

func TestNarrowWidthSplitViewInPanelContained(t *testing.T) {
	panel := NewPanel("p", "Split", 0, 0, 320, 0)
	panel.AutoHeight = true
	panel.TitleHeight = 32

	left := NewLabel("left", "Left pane text that wraps at narrow width.", 0, 0, 0, 24)
	right := NewLabel("right", "Right pane text that wraps at narrow width.", 0, 0, 0, 24)
	sv := NewSplitView("sv", SplitHorizontal, left, right, 0, 0, 0, 0)
	sv.SetFlexGrow(1)
	panel.AddChild(sv)

	panel.SetBounds(rl.NewRectangle(0, 0, 320, 280))
	panel.Layout()

	panelBottom := panel.Bounds().Y + panel.Bounds().Height
	if bottom := nodeSubtreeBottom(sv); bottom > panelBottom+2 {
		t.Fatalf("split subtree bottom %.0f below panel bottom %.0f", bottom, panelBottom)
	}
	if sv.Bounds().Width > panel.Bounds().Width+1 {
		t.Fatalf("split width %.0f exceeds panel %.0f", sv.Bounds().Width, panel.Bounds().Width)
	}
	if left.Bounds().Width > sv.firstRect().Width+2 {
		t.Fatalf("left pane label width %.0f exceeds pane %.0f", left.Bounds().Width, sv.firstRect().Width)
	}
}
