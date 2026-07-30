package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tdewolff/font"
)

func TestFontsConvertTTFPassthrough(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Skip(err)
	}
	src := filepath.Join(root, "assets", "fonts", "remixicon.ttf")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Skip("remixicon.ttf missing")
	}
	out, err := font.ToSFNT(data)
	if err != nil {
		t.Fatalf("ToSFNT: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("empty sfnt output")
	}
	if ext := font.Extension(out); ext != ".ttf" && ext != ".otf" {
		t.Fatalf("unexpected extension %q", ext)
	}
}
