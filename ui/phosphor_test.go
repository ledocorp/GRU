package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePhosphorSelection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "selection.json")
	const sample = `{
	  "icons": [{
	    "properties": { "name": "house", "code": 58050 }
	  }, {
	    "properties": { "name": "acorn-duotone", "codes": [60314, 60315] }
	  }]
	}`
	if err := os.WriteFile(path, []byte(sample), 0o644); err != nil {
		t.Fatal(err)
	}
	glyphs, cps, err := parsePhosphorSelection(path)
	if err != nil {
		t.Fatal(err)
	}
	if glyphs["house"].primary != rune(58050) {
		t.Fatalf("house code %U", glyphs["house"].primary)
	}
	if glyphs["acorn-duotone"].secondary == 0 {
		t.Fatal("expected duotone secondary code")
	}
	if len(cps) < 3 {
		t.Fatalf("codepoints = %v", cps)
	}
}

func TestPhosphorSelectionName(t *testing.T) {
	if phosphorSelectionName("house", PhosphorRegular) != "house" {
		t.Fatal("regular name unchanged")
	}
	if phosphorSelectionName("house", PhosphorFill) != "house-fill" {
		t.Fatal("fill name suffix")
	}
	if phosphorSelectionName("house", PhosphorBold) != "house-bold" {
		t.Fatal("bold name suffix")
	}
	if phosphorSelectionName("star-fill", PhosphorFill) != "star-fill" {
		t.Fatal("fill name not doubled")
	}
}
