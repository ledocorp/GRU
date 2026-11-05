package ui

import (
	"testing"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	xfixed "golang.org/x/image/math/fixed"
)

func TestPreferMSDFChromeOptIn(t *testing.T) {
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_MSDF", "")
		t.Setenv("GORY_TEXT_MSDF", "")
	})
	t.Setenv("GRU_TEXT_MSDF", "")
	if PreferMSDFChrome() {
		t.Fatal("default must not enable MSDF")
	}
	t.Setenv("GRU_TEXT_MSDF", "1")
	if !PreferMSDFChrome() {
		t.Fatal("env=1 enables")
	}
}

func TestPackMSDFAtlasHasField(t *testing.T) {
	path := firstExisting(fontCandidates...)
	if path == "" {
		t.Skip("no UI font")
	}
	data, err := ReadAssetFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := truetype.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              32,
		DPI:               72,
		Hinting:           font.HintingFull,
		GlyphCacheEntries: 8,
	})
	defer face.Close()
	atlas, glyphs, ink, ok := packMSDFAtlas(face, 32)
	if !ok {
		t.Fatalf("pack failed ink=%d glyphs=%d", ink, len(glyphs))
	}
	g, ok := glyphs['A']
	if !ok || !g.hasInk {
		t.Fatalf("glyph A missing: %+v", g)
	}
	foundInside := false
	for y := int(g.src.Y); y < int(g.src.Y+g.src.Height); y++ {
		for x := int(g.src.X); x < int(g.src.X+g.src.Width); x++ {
			r, _, _, _ := atlas.At(x, y).RGBA()
			if r>>8 > 160 {
				foundInside = true
				break
			}
		}
		if foundInside {
			break
		}
	}
	if !foundInside {
		t.Fatal("expected inside SDF (>~0.63) somewhere in glyph A")
	}
	// Outside padding corner should stay low (transparent in shader).
	ox, oy := int(g.src.X), int(g.src.Y)
	r0, _, _, _ := atlas.At(ox, oy).RGBA()
	if r0>>8 > 100 {
		t.Fatalf("padding corner should be outside (low), got R=%d", r0>>8)
	}
	dr, mask, _, _, ok := face.Glyph(xfixed.Point26_6{}, 'A')
	if !ok || mask == nil || dr.Empty() {
		t.Fatal("Glyph A failed")
	}
}
