package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestColorWellSelection(t *testing.T) {
	sw := []rl.Color{
		rl.Red,
		rl.Blue,
	}
	cw := NewColorWell("cw", rl.Red, sw, 0, 0, 0, 0)
	cw.SetBounds(rl.NewRectangle(0, 0, 80, colorWellDefaultH))
	if cw.selectedIndex() != 0 {
		t.Fatalf("selectedIndex = %d, want 0", cw.selectedIndex())
	}
	cw.Value.Set(rl.Blue)
	if cw.selectedIndex() != 1 {
		t.Fatalf("selectedIndex = %d, want 1", cw.selectedIndex())
	}
}

func TestColorWellIntrinsicWidth(t *testing.T) {
	cw := NewColorWell("cw", rl.Red, []rl.Color{rl.Red, rl.Green, rl.Blue}, 0, 0, 0, 0)
	want := colorWellCellSize*3 + colorWellCellGap*2
	if w := cw.intrinsicWidth(); w < want-0.5 || w > want+0.5 {
		t.Fatalf("intrinsicWidth = %v, want %v", w, want)
	}
}
