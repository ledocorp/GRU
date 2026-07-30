package ui

import (
	"os"
	"path/filepath"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestGalleryPresetRowLayoutHorizontal(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "..", "pages", "gallery.gru"))
	if err != nil {
		t.Fatal(err)
	}
	node, err := LoadDocumentSpec(data, NewBuildContext())
	if err != nil {
		t.Fatal(err)
	}
	row := findNodeByID(node, "gallery-presets-strip-row-0")
	if row == nil {
		t.Fatal("gallery-presets-strip-row-0 not found")
	}
	c, ok := row.(*Container)
	if !ok {
		t.Fatalf("row type = %T, want *Container", row)
	}
	if c.FlexDirection != FlexRow {
		t.Fatalf("row direction = %v, want FlexRow", c.FlexDirection)
	}
	if !c.FlexWrap {
		t.Fatal("preset row should use flex wrap")
	}
	if len(c.Children()) < 2 {
		t.Fatalf("row children = %d, want >= 2 preset tiles", len(c.Children()))
	}

	node.SetBounds(rl.NewRectangle(0, 0, 920, 4096))
	node.Layout()

	var tiles []Node
	for _, ch := range c.Children() {
		if ch.IsHidden() {
			continue
		}
		tiles = append(tiles, ch)
	}
	if len(tiles) < 2 {
		t.Fatal("need at least two visible preset tiles")
	}
	a, b := tiles[0].Bounds(), tiles[1].Bounds()
	if a.Width < 120 {
		t.Fatalf("first tile width %.0f too narrow", a.Width)
	}
	if b.X <= a.X+a.Width*0.25 {
		t.Fatalf("second tile should be to the right of first: x0=%.0f w0=%.0f x1=%.0f", a.X, a.Width, b.X)
	}
	backdrop := findNodeByID(node, "gallery-presets-strip-strip")
	if backdrop == nil {
		backdrop = findNodeByID(node, "gallery-presets-strip-backdrop")
	}
	if backdrop == nil {
		t.Fatal("preset strip frame not found")
	}
	if _, ok := backdrop.(*PresetStripFrame); !ok {
		if db, ok := backdrop.(*DemoBackdrop); ok && db.ClipChildren {
			t.Fatal("legacy preset backdrop should not clip")
		}
	}
}

func TestGalleryPresetRowWrapsWhenNarrow(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "..", "pages", "gallery.gru"))
	if err != nil {
		t.Fatal(err)
	}
	node, err := LoadDocumentSpec(data, NewBuildContext())
	if err != nil {
		t.Fatal(err)
	}
	row := findNodeByID(node, "gallery-presets-strip-row-0")
	if row == nil {
		t.Fatal("gallery-presets-strip-row-0 not found")
	}
	node.SetBounds(rl.NewRectangle(0, 0, 460, 4096))
	node.Layout()

	c := row.(*Container)
	var tiles []Node
	for _, ch := range c.Children() {
		if !ch.IsHidden() {
			tiles = append(tiles, ch)
		}
	}
	if len(tiles) < 2 {
		t.Fatalf("tiles = %d, want >= 2", len(tiles))
	}
	rows := map[int32]int{}
	for _, tile := range tiles {
		rows[int32(tile.Bounds().Y+0.5)]++
	}
	if len(rows) < 2 {
		t.Fatalf("expected wrapped rows at width 460, got %d row(s)", len(rows))
	}
}

func TestGalleryPresetTileWidthsInPresetsCard(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "..", "pages", "gallery.gru"))
	if err != nil {
		t.Fatal(err)
	}
	node, err := LoadDocumentSpec(data, NewBuildContext())
	if err != nil {
		t.Fatal(err)
	}
	presets := findNodeByID(node, "gallery-presets-card")
	if presets == nil {
		t.Fatal("presets card missing")
	}
	presets.SetBounds(rl.NewRectangle(0, 0, 800, 4096))
	presets.Layout()

	row := findNodeByID(node, "gallery-presets-strip-row-0").(*Container)
	for _, ch := range row.Children() {
		b := ch.Bounds()
		if b.Width > 350 {
			t.Fatalf("tile %s width %.0f looks like full-row stack", ch.ID(), b.Width)
		}
	}
}

func TestDocumentSpecPresetRow(t *testing.T) {
	node, err := LoadDocumentSpec([]byte(`{
		"id": "root",
		"children": [{
			"type": "presetRow",
			"id": "demo-strip",
			"columns": 2,
			"items": [
				{ "preset": "neo-glow-card", "text": "Neo" },
				{ "preset": "glass-panel", "text": "Glass" },
				{ "preset": "glass-card", "text": "Card" }
			]
		}]
	}`), NewBuildContext())
	if err != nil {
		t.Fatal(err)
	}
	if findNodeByID(node, "demo-strip-row-0") == nil {
		t.Fatal("missing first row chunk")
	}
	if findNodeByID(node, "demo-strip-row-1") == nil {
		t.Fatal("missing second row chunk")
	}
}
