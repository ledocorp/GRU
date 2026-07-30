package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestBuildPresetRowSingleWrapRow(t *testing.T) {
	strip, err := BuildPresetRow("strip", []PresetTileSpec{
		{Preset: "neo-glow-card", Text: "Neo glow body copy."},
		{Preset: "glass-panel", Text: "Glass panel body."},
	}, DefaultPresetRowOptions())
	if err != nil {
		t.Fatal(err)
	}
	bd, ok := strip.(*PresetStripFrame)
	if !ok {
		t.Fatalf("type = %T, want PresetStripFrame", strip)
	}
	if bd.ClipChildren {
		t.Fatal("preset strip should not create a nested scroll clip")
	}
	row := findNodeByID(strip, "strip-row")
	if row == nil {
		t.Fatal("strip-row not found")
	}
	c := row.(*Container)
	if !c.FlexWrap || c.FlexDirection != FlexRow {
		t.Fatal("expected flex row wrap")
	}
}

func TestBuildPresetRowColumnsChunksRows(t *testing.T) {
	opts := DefaultPresetRowOptions()
	opts.Columns = 3
	strip, err := BuildPresetRow("strip", []PresetTileSpec{
		{Preset: "neo-glow-card", Text: "a"},
		{Preset: "glass-panel", Text: "b"},
		{Preset: "glass-panel-dark", Text: "c"},
		{Preset: "glass-card", Text: "d"},
		{Preset: "callout-tip", Text: "e"},
	}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if findNodeByID(strip, "strip-row-0") == nil || findNodeByID(strip, "strip-row-1") == nil {
		t.Fatal("expected two explicit row chunks for columns=3")
	}
}

func TestBuildPresetRowLayoutHorizontal(t *testing.T) {
	strip, err := BuildPresetRow("strip", []PresetTileSpec{
		{Preset: "neo-glow-card", Text: "Neo glow."},
		{Preset: "glass-panel", Text: "Glass."},
	}, DefaultPresetRowOptions())
	if err != nil {
		t.Fatal(err)
	}
	strip.SetBounds(rl.NewRectangle(0, 0, 640, 4096))
	strip.Layout()
	row := findNodeByID(strip, "strip-row").(*Container)
	children := row.Children()
	if len(children) < 2 {
		t.Fatal("need two tiles")
	}
	a, b := children[0].Bounds(), children[1].Bounds()
	if b.X <= a.X+a.Width*0.25 {
		t.Fatalf("tiles not horizontal: %+v %+v", a, b)
	}
}

func TestBuildPresetTileCardRejectsUnknownPreset(t *testing.T) {
	_, err := BuildPresetTileCard("x", PresetTileSpec{Preset: "not-a-preset", Text: "n"}, 200)
	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
}
