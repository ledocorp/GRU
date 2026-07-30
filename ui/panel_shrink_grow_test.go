package ui

import (
	"fmt"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestAutoHeightPanelGrowsWhenAccordionOpensAfterGridShrink(t *testing.T) {
	grid := NewContainer("grid", 0, 0, 320, 400)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 1
	grid.AutoHeight = true
	grid.SetFlexGrow(1)
	grid.Gap = 12

	panel := NewPanel("p", "Accordion", 0, 0, 0, 0)
	panel.AutoHeight = true
	panel.TitleHeight = 32

	acc := NewAccordion("acc", "Section", 0, 0, 0, 0)
	acc.Expanded.Set(false)
	acc.animH = 0
	acc.AddChild(NewLabel("lbl", "Hidden until expanded — panel must grow to fit this body text when opened.", 0, 0, 0, 48))
	panel.AddChild(acc)
	grid.AddChild(panel)

	root := NewContainer("root", 0, 0, 320, 400)
	root.LayoutType = LayoutAbsolute
	root.AddChild(grid)
	root.Layout()
	shrunkH := panel.Bounds().Height

	acc.Expanded.Set(true)
	acc.animH = acc.contentH
	if acc.animH <= 0 {
		acc.Layout()
		acc.animH = acc.contentH
	}
	panel.MarkDirty()
	grid.MarkDirty()
	root.Layout()

	if panel.Bounds().Height <= shrunkH+8 {
		t.Fatalf("panel height %.0f should grow after accordion open (was %.0f shrunk)", panel.Bounds().Height, shrunkH)
	}
	panelBottom := panel.Bounds().Y + panel.Bounds().Height
	if bottom := nodeSubtreeBottom(acc); bottom > panelBottom+2 {
		t.Fatalf("accordion bottom %.0f below panel %.0f after open", bottom, panelBottom)
	}
}

func TestSplitViewDoesNotInflateAutoHeightPanelFromPaneOverflow(t *testing.T) {
	panel := NewPanel("p", "Split", 0, 0, 600, 0)
	panel.AutoHeight = true
	panel.TitleHeight = 32

	lines := make([]string, 40)
	for i := range lines {
		lines[i] = "overflow line"
	}
	left := NewPanel("left", "Files", 0, 0, 0, 0)
	left.AutoHeight = false
	left.TitleHeight = 28
	for i, line := range lines {
		left.AddChild(NewLabel(fmt.Sprintf("l%d", i), line, 0, 0, 0, 18))
	}
	right := NewLabel("right", "Editor", 0, 0, 0, 24)
	sv := NewSplitView("sv", SplitHorizontal, left, right, 0, 0, 0, 0)
	sv.SetFlexGrow(1)
	panel.AddChild(sv)

	panel.SetBounds(rl.NewRectangle(0, 0, 600, 280))
	panel.Layout()

	if panel.Bounds().Height > 320 {
		t.Fatalf("panel height %.0f should not balloon from inner pane label stack", panel.Bounds().Height)
	}
	if sv.Bounds().Height > 260 {
		t.Fatalf("split height %.0f should stay within assigned panel band", sv.Bounds().Height)
	}
}

func TestDemoBackdropHeightCoversWrappedPresetRow(t *testing.T) {
	row, err := BuildPresetRow("strip", []PresetTileSpec{
		{Preset: "neo-glow-card", Text: "Tile one with enough copy to wrap."},
		{Preset: "glass-panel", Text: "Tile two with enough copy to wrap."},
		{Preset: "glass-card", Text: "Tile three with enough copy to wrap."},
	}, PresetRowOptions{MinTileWidth: 200, UseBackdrop: true})
	if err != nil {
		t.Fatal(err)
	}
	bd := row.(*PresetStripFrame)
	bd.SetBounds(rl.NewRectangle(0, 0, 440, 0))
	bd.Layout()

	bottom := bd.Bounds().Y + bd.Bounds().Height
	for _, ch := range bd.Children() {
		if sub := nodeSubtreeBottom(ch); sub > bottom+2 {
			t.Fatalf("backdrop height %.0f does not cover child bottom %.0f", bottom, sub)
		}
	}
}

func TestSplitPanelRespectsAssignedBandWhenSqueezed(t *testing.T) {
	panel := NewPanel("p", "Split", 0, 0, 600, 0)
	panel.AutoHeight = true
	panel.TitleHeight = 32
	panel.AddChild(NewLabel("help", "Drag the splitter.", 0, 0, 0, 18))

	left := NewLabel("left", "Files", 0, 0, 0, 24)
	right := NewLabel("right", "Editor", 0, 0, 0, 24)
	sv := NewSplitView("sv", SplitHorizontal, left, right, 0, 0, 0, 0)
	sv.SetFlexGrow(1)
	panel.AddChild(sv)

	panel.SetBounds(rl.NewRectangle(0, 0, 600, 200))
	panel.Layout()

	if panel.Bounds().Height > 210 {
		t.Fatalf("panel height %.0f should honor assigned 200px band", panel.Bounds().Height)
	}
	if sv.Bounds().Height > 170 {
		t.Fatalf("split height %.0f should stay inside squeezed panel body", sv.Bounds().Height)
	}
}
