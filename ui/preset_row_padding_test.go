package ui

import (
	"testing"
)

func TestPresetRowBackdropHasPadding(t *testing.T) {
	row, err := BuildPresetRow("strip", []PresetTileSpec{
		{Preset: "neo-glow-card", Text: "Tile"},
	}, DefaultPresetRowOptions())
	if err != nil {
		t.Fatal(err)
	}
	bd := row.(*PresetStripFrame)
	if got := bd.GetStyle().Padding; got != PresetBackdropPadding {
		t.Fatalf("backdrop padding = %.0f, want %.0f", got, PresetBackdropPadding)
	}
}

func TestFitSubtreeLabelsWrapsWithoutTruncate(t *testing.T) {
	lbl := NewLabel("lbl", "Any accordion body copy.", 0, 0, 320, 48)
	fitSubtreeLabels(200, []Node{lbl})
	if lbl.Bounds().Width > 200.5 {
		t.Fatalf("label width %.0f should cap to 200", lbl.Bounds().Width)
	}
	if lbl.Truncate {
		t.Fatal("fitSubtreeLabels must not force truncate; wrapping should grow height instead")
	}
}
