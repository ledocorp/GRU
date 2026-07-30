package ui

import (
	"os"
	"path/filepath"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func loadGalleryNode(t *testing.T) Node {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(wd, "..", "pages", "gallery.gru"))
	if err != nil {
		t.Fatal(err)
	}
	node, err := LoadDocumentSpec(data, NewBuildContext())
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func TestPresetTileCardIntrinsicHeightAfterFlexMeasure(t *testing.T) {
	card, err := BuildPresetTileCard("tile", PresetTileSpec{
		Preset: "neo-glow-card",
		Text:   "Dark indigo card with outer halo and inner glow rings.",
	}, 200)
	if err != nil {
		t.Fatal(err)
	}
	card.SetBounds(rl.NewRectangle(0, 0, 200, 0))
	card.Layout()
	if card.Bounds().Height < 80 {
		t.Fatalf("card height = %.0f, want intrinsic shrink-wrap >= 80", card.Bounds().Height)
	}
}

func TestFlexRowWrapUsesPostLayoutAutoHeight(t *testing.T) {
	row := NewContainer("row", 0, 0, 640, 0)
	row.FlexDirection = FlexRow
	row.SetFlexWrap(true)
	row.AutoHeight = true
	row.Gap = 12
	card, err := BuildPresetTileCard("tile", PresetTileSpec{
		Preset: "glass-panel",
		Text:   "Light frosted panel with native top gradient sheen.",
	}, 200)
	if err != nil {
		t.Fatal(err)
	}
	row.AddChild(card)
	row.SetBounds(rl.NewRectangle(0, 0, 640, 4096))
	row.Layout()
	if card.Bounds().Height < 80 {
		t.Fatalf("wrapped card height = %.0f, want intrinsic >= 80", card.Bounds().Height)
	}
}

func TestGalleryPresetTileNotTinyAfterLayout(t *testing.T) {
	node := loadGalleryNode(t)
	node.SetBounds(rl.NewRectangle(0, 0, 920, 4096))
	node.Layout()

	card := findNodeByID(node, "gallery-presets-strip-neo-glow-card-0")
	if card == nil {
		card = findNodeByID(node, "gallery-presets-strip-neo-glow-card")
	}
	if card == nil {
		t.Fatal("neo-glow preset tile missing")
	}
	b := card.Bounds()
	if b.Width < 180 {
		t.Fatalf("tile width %.0f too narrow", b.Width)
	}
	if b.Height < 80 {
		t.Fatalf("tile height %.0f too short (tiny rectangle bug)", b.Height)
	}
	c, ok := card.(*Card)
	if !ok {
		t.Fatalf("type = %T", card)
	}
	if c.body == nil || c.body.Bounds().Height < 40 {
		t.Fatalf("body height = %.0f", c.body.Bounds().Height)
	}
	rt, ok := c.Children()[0].(*RichText)
	if !ok {
		t.Fatalf("child type = %T", c.Children()[0])
	}
	if rt.Bounds().Width < 120 {
		t.Fatalf("richtext width %.0f", rt.Bounds().Width)
	}
}
