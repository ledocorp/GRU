package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestButtonPreferredWidthShrinkWrap(t *testing.T) {
	btn := NewButton("b", "Randomize", 0, 0, 0, 36)
	pw := btn.GetPreferredWidth()
	if pw < 48 {
		t.Fatalf("preferred width = %v, want >= 48", pw)
	}
	if pw > 400 {
		t.Fatalf("preferred width = %v, too wide for label", pw)
	}
}

func TestButtonInFlexColumnDoesNotStretchFullWidth(t *testing.T) {
	panel := NewPanel("p", "Test", 0, 0, 320, 200)
	btn := NewButton("b", "Open notifications", 0, 0, 0, 36)
	panel.AddChild(btn)
	panel.SetBounds(rl.NewRectangle(0, 0, 320, 200))
	panel.Layout()
	bw := btn.Bounds().Width
	if bw > 200 {
		t.Fatalf("button width = %v, expected shrink-wrap not full panel (~280)", bw)
	}
}

func TestButtonsInFlexRowUsePreferredWidth(t *testing.T) {
	row := NewContainer("row", 0, 0, 400, 40)
	row.LayoutType = LayoutFlex
	row.FlexDirection = FlexRow
	row.Gap = 8
	a := NewButton("a", "Fire toast", 0, 0, 0, 36)
	b := NewButton("b", "Open notifications", 0, 0, 0, 36)
	row.AddChild(a)
	row.AddChild(b)
	row.SetBounds(rl.NewRectangle(0, 0, 400, 40))
	row.Layout()
	if a.Bounds().Width < 48 || b.Bounds().Width < 48 {
		t.Fatalf("widths too small: a=%v b=%v", a.Bounds().Width, b.Bounds().Width)
	}
	if a.Bounds().Width > 220 || b.Bounds().Width > 220 {
		t.Fatalf("widths squished by equalShare: a=%v b=%v", a.Bounds().Width, b.Bounds().Width)
	}
}
