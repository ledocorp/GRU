package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestAccordionBodyLabelsReflowOnWidthChange(t *testing.T) {
	acc := NewAccordion("acc", "Section", 0, 0, 320, 0)
	col := NewContainer("col", 0, 0, 0, 0)
	col.FlexDirection = FlexColumn
	col.SetStyle("transparent")
	lbl := NewLabel("lbl", "Gru is a retained-mode UI library for raylib-go with widgets, signals, themes, panels, accordions, and responsive layout reflow on every resize.", 0, 0, 0, 18)
	lbl.SetStyle("form-value")
	col.AddChild(lbl)
	acc.AddChild(col)
	acc.Expanded.Set(true)

	root := NewContainer("root", 0, 0, 520, 400)
	root.AddChild(acc)
	root.Layout()

	acc.SetBounds(rl.NewRectangle(0, 0, 280, acc.Bounds().Height))
	acc.MarkDirty()
	root.Layout()
	if !lbl.Wrap || !lbl.AutoHeight {
		t.Fatal("accordion body label should use wrap + AutoHeight")
	}
	if lbl.Bounds().Width > 290 {
		t.Fatalf("label width %.0f should cap to accordion body", lbl.Bounds().Width)
	}
}
