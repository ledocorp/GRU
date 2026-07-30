package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestAccordionPanelShrinksWhenCollapsed(t *testing.T) {
	panel := NewPanel("p", "Accordion", 0, 0, 320, 0)
	panel.AutoHeight = true
	panel.TitleHeight = 32

	acc := NewAccordion("acc", "Section", 0, 0, 0, 0)
	acc.AddChild(NewLabel("lbl", "Body text that should not spill past the panel when collapsed.", 0, 0, 0, 48))
	panel.AddChild(acc)

	root := NewContainer("root", 0, 0, 320, 600)
	root.LayoutType = LayoutAbsolute
	root.AddChild(panel)
	root.Layout()
	collapsedH := panel.Bounds().Height

	acc.Expanded.Set(false)
	acc.animH = 0
	acc.bounds.Height = acc.TitleHeight
	panel.MarkDirty()
	root.Layout()
	if panel.Bounds().Height > collapsedH+8 {
		t.Fatalf("collapsed panel %.0f should stay near %.0f", panel.Bounds().Height, collapsedH)
	}
	panelBottom := panel.Bounds().Y + panel.Bounds().Height
	if bottom := nodeSubtreeBottom(acc); bottom > panelBottom+2 {
		t.Fatalf("accordion subtree bottom %.0f below panel bottom %.0f", bottom, panelBottom)
	}
}

func TestAccordionPreExpandedBeforeWidth(t *testing.T) {
	panel := NewPanel("p", "Stack", 0, 0, 320, 0)
	panel.AutoHeight = true
	panel.TitleHeight = 32
	panel.Gap = 8

	acc1 := NewAccordion("a1", "Open", 0, 0, 0, 0)
	acc1.AddChild(NewLabel("b1", "First body line", 0, 0, 0, 24))
	acc1.Expanded.Set(true)

	acc2 := NewAccordion("a2", "Below", 0, 0, 0, 0)
	acc2.AddChild(NewLabel("b2", "Second body line", 0, 0, 0, 24))

	panel.AddChild(acc1)
	panel.AddChild(acc2)
	panel.SetBounds(rl.NewRectangle(0, 0, 320, 0))
	panel.Layout()

	acc1Bottom := acc1.Bounds().Y + acc1.Bounds().Height
	if acc2.Bounds().Y < acc1Bottom-1 {
		t.Fatalf("acc2 y=%.0f overlaps acc1 bottom=%.0f (acc1 h=%.0f animH=%.0f)",
			acc2.Bounds().Y, acc1Bottom, acc1.Bounds().Height, acc1.animH)
	}
}

func TestNestedViewportDoesNotInflatePanel(t *testing.T) {
	panel := NewPanel("p", "Scroll", 0, 0, 400, 0)
	panel.AutoHeight = true

	vp := NewViewport("vp", 0, 0, 0, 200)
	vp.AddChild(NewLabel("lbl", "Line", 0, 0, 0, 24))
	panel.AddChild(vp)

	panel.SetBounds(rl.NewRectangle(0, 0, 400, 4096))
	panel.Layout()

	if panel.Bounds().Height >= 4096 {
		t.Fatalf("panel kept probe height %.0f", panel.Bounds().Height)
	}
	if panel.Bounds().Height > 280 {
		t.Fatalf("panel height %.0f, want ~viewport band not scroll content", panel.Bounds().Height)
	}
}
