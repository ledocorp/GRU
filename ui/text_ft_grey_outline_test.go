package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackFTGreyAtlasOutlineHasSolidR(t *testing.T) {
	candidates := []string{
		"assets/fonts/firacode/static/FiraCode-Regular.ttf",
		"assets/fonts/poppins/Poppins-Regular.ttf",
		"../assets/fonts/firacode/static/FiraCode-Regular.ttf",
		"../assets/fonts/poppins/Poppins-Regular.ttf",
	}
	var data []byte
	var path string
	for _, c := range candidates {
		b, err := os.ReadFile(c)
		if err == nil && len(b) > 0 {
			data, path = b, c
			break
		}
		if abs, err2 := filepath.Abs(c); err2 == nil {
			if b, err = os.ReadFile(abs); err == nil && len(b) > 0 {
				data, path = b, abs
				break
			}
		}
	}
	if len(data) == 0 {
		t.Skip("no font")
	}
	_ = path
	atlas, glyphs, ink, ok := packFTGreyAtlasOutline(data, 16, 4)
	if !ok {
		t.Fatalf("outline pack failed ink=%d", ink)
	}
	g, ok := glyphs['R']
	if !ok || !g.hasInk {
		t.Fatal("missing R")
	}
	// Scan R for interior near-zero bridges (opposing strong neighbors).
	holes := 0
	x0, y0 := int(g.src.X), int(g.src.Y)
	w, h := int(g.src.Width), int(g.src.Height)
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			_, _, _, a := atlas.At(x0+x, y0+y).RGBA()
			aa := int(a >> 8)
			if aa > 40 {
				continue
			}
			_, _, _, l := atlas.At(x0+x-1, y0+y).RGBA()
			_, _, _, r := atlas.At(x0+x+1, y0+y).RGBA()
			if int(l>>8) >= 120 && int(r>>8) >= 120 {
				holes++
			}
		}
	}
	if holes > 2 {
		t.Fatalf("R still has %d horizontal bridge holes after outline AA", holes)
	}
	t.Logf("R ok holes=%d ink=%d src=%.0fx%.0f path=%s", holes, ink, g.src.Width, g.src.Height, path)
}
