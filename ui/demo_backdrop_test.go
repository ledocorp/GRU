package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestDemoBackdropFillRectMatchesPaddedContent(t *testing.T) {
	strip, err := BuildPresetRow("strip", []PresetTileSpec{
		{Preset: "neo-glow-card", Text: "Tile one."},
		{Preset: "glass-panel", Text: "Tile two."},
	}, DefaultPresetRowOptions())
	if err != nil {
		t.Fatal(err)
	}
	bd := strip.(*PresetStripFrame)
	bd.SetBounds(rl.NewRectangle(0, 0, 440, 0))
	bd.Layout()

	fill := bd.fillRect()
	pad := bd.GetStyle().Padding
	b := bd.Bounds()
	if fill.X != b.X+pad || fill.Y != b.Y+pad {
		t.Fatalf("fill origin (%v) should match padded content origin", fill)
	}
	for _, ch := range bd.Children() {
		cb := ch.Bounds()
		if cb.X < fill.X-1 || cb.Y < fill.Y-1 {
			t.Fatalf("child at (%v) starts above/left of fill (%v)", cb, fill)
		}
	}
}
