package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestModalOverlayLayoutLabel(t *testing.T) {
	lbl := NewLabel("modal-test", "Discard unsaved changes?", 0, 0, 0, 0)
	lbl.SetStyle("modal-body")
	lbl.Align = LabelAlignLeft
	lbl.Wrap = true
	rect := rl.NewRectangle(120, 180, 360, 96)
	layoutOverlaySubtree(lbl, rect)
	b := lbl.Bounds()
	if b.Width < 200 || b.Height < 8 {
		t.Fatalf("label bounds too small: %+v", b)
	}
}

func TestModalOverlayLayoutRenameForm(t *testing.T) {
	cap := NewLabel("cap", "Name in Open Notes:", 0, 0, 0, 0)
	cap.SetStyle("modal-body")
	in := NewTextInput("in", "", 0, 0, 0, 44)
	in.SetStyle("modal-input")
	body := NewContainer("body", 0, 0, 0, 0)
	body.LayoutType = LayoutFlex
	body.FlexDirection = FlexColumn
	body.Gap = 8
	body.AutoHeight = true
	body.SetStyle("transparent")
	body.AddChild(cap)
	body.AddChild(in)
	pad := NewContainer("pad", 0, 0, 0, 8)
	pad.SetStyle("transparent")
	body.AddChild(pad)

	// Content band for 440×280 single-field modal (modal.go contentRect math).
	rect := rl.NewRectangle(100, 160, 400, 128)
	layoutOverlaySubtree(body, rect)
	inB := in.Bounds()
	if inB.Width < 300 {
		t.Fatalf("input too narrow: %+v", inB)
	}
	if inB.Height < 40 {
		t.Fatalf("input height %+v", inB)
	}
	bottom := inB.Y + inB.Height
	if bottom > rect.Y+rect.Height {
		t.Fatalf("input extends past content band: bottom=%.1f max=%.1f", bottom, rect.Y+rect.Height)
	}
}

func TestModalOverlayLayoutFormBody(t *testing.T) {
	cap := NewLabel("cap", "Find what:", 0, 0, 0, 0)
	cap.SetStyle("form-label")
	in := NewTextInput("in", "", 0, 0, 0, 40)
	body := NewContainer("body", 0, 0, 0, 0)
	body.LayoutType = LayoutFlex
	body.FlexDirection = FlexColumn
	body.Gap = 8
	body.AutoHeight = true
	body.SetStyle("transparent")
	body.AddChild(cap)
	body.AddChild(in)

	rect := rl.NewRectangle(100, 160, 380, 140)
	layoutOverlaySubtree(body, rect)
	if in.Bounds().Height < 30 {
		t.Fatalf("input height %+v", in.Bounds())
	}
	if cap.Bounds().Height < 8 {
		t.Fatalf("caption height %+v", cap.Bounds())
	}
}
