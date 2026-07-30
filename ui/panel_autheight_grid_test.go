package ui

import (
	"fmt"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestAutoHeightPanelContainsNestedLayoutGrid(t *testing.T) {
	panel := NewPanel("grid-panel", "12-Column Grid", 0, 0, 800, 0)
	panel.AutoHeight = true

	chrome := NewCard("chrome", "", 0, 0, 420, 0)
	chrome.AutoHeight = true
	chrome.PreferredWidth = 420
	chrome.MinWidth = 420
	chrome.MaxWidth = 420

	grid := NewContainer("grid", 0, 0, 0, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 12
	grid.Gap = 10

	for i := 0; i < 8; i++ {
		card := NewCard(fmt.Sprintf("card-%d", i), "", 0, 0, 0, 0)
		card.AutoHeight = true
		card.AddChild(NewLabel(fmt.Sprintf("lbl-%d", i), "Metric", 0, 0, 0, 0))
		card.SetColSpan(BreakpointXS, 12)
		card.SetColSpan(BreakpointSM, 6)
		card.SetColSpan(BreakpointMD, 4)
		card.SetColSpan(BreakpointLG, 3)
		card.SetColSpan(BreakpointXL, 3)
		grid.AddChild(card)
	}

	chrome.AddChild(grid)
	panel.AddChild(chrome)

	panel.SetBounds(rl.NewRectangle(0, 0, 800, 4096))
	panel.Layout()

	if panel.Bounds().Height >= 4096 {
		t.Fatalf("panel kept probe height: %.0f", panel.Bounds().Height)
	}

	panelBottom := panel.Bounds().Y + panel.Bounds().Height
	if bottom := nodeSubtreeBottom(chrome); bottom > panelBottom+1 {
		t.Fatalf("grid chrome bottom %.0f below panel bottom %.0f", bottom, panelBottom)
	}
}

func TestAutoHeightPanelShrinksAfterNestedGridReflow(t *testing.T) {
	panel := NewPanel("grid-panel", "12-Column Grid", 0, 0, 800, 0)
	panel.AutoHeight = true

	chrome := NewCard("chrome", "", 0, 0, 360, 0)
	chrome.AutoHeight = true

	grid := NewContainer("grid", 0, 0, 0, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 12
	grid.Gap = 10
	for i := 0; i < 8; i++ {
		card := NewCard(fmt.Sprintf("card-%d", i), "", 0, 0, 0, 0)
		card.AutoHeight = true
		card.AddChild(NewLabel(fmt.Sprintf("lbl-%d", i), "Metric", 0, 0, 0, 0))
		card.SetColSpan(BreakpointXS, 12)
		card.SetColSpan(BreakpointSM, 6)
		card.SetColSpan(BreakpointMD, 4)
		card.SetColSpan(BreakpointLG, 3)
		card.SetColSpan(BreakpointXL, 3)
		grid.AddChild(card)
	}
	chrome.AddChild(grid)
	panel.AddChild(chrome)

	setPreviewWidth := func(w float32) {
		chrome.PreferredWidth = w
		chrome.MinWidth = w
		chrome.MaxWidth = w
		chrome.InvalidateLayoutPassCache()
		chrome.MarkDirty()
		grid.InvalidateLayoutPassCache()
		grid.MarkDirty()
		panel.InvalidateLayoutPassCache()
		panel.MarkDirty()
	}

	panel.SetBounds(rl.NewRectangle(0, 0, 800, 4096))
	setPreviewWidth(360)
	panel.Layout()
	narrowH := panel.Bounds().Height

	setPreviewWidth(900)
	panel.Layout()
	wideH := panel.Bounds().Height

	setPreviewWidth(360)
	panel.Layout()
	revertH := panel.Bounds().Height

	if revertH > narrowH+1 {
		t.Fatalf("panel stuck tall after revert: narrow=%.0f wide=%.0f revert=%.0f", narrowH, wideH, revertH)
	}
}
