//go:build cgo && !gru_nofx

package ui

import (
	"image"
	"os"
	"testing"

	"github.com/fogleman/gg"
	"github.com/ledocorp/gru/ui/fxglyph"
)

func TestFxGlyphRasterMatchesGGInk(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	face := pickShapedFace(false, false, false, false)
	if face == nil {
		t.Fatal("no face")
	}
	gid, ok := face.NominalGlyph('A')
	if !ok {
		t.Fatal("no glyph")
	}
	outline, ok := shapedGlyphOutline(face, gid)
	if !ok {
		t.Fatal("no outline")
	}
	ext, ok := face.GlyphExtents(gid)
	if !ok {
		t.Fatal("no extents")
	}
	scale, w, h, ox, oy := shapedGlyphRasterLayout(outline, ext, shapedGlyphRasterPx(16), face.Upem())

	gc := gg.NewContext(w, h)
	gc.SetRGBA(0, 0, 0, 0)
	gc.Clear()
	gc.SetRGBA(1, 1, 1, 1)
	drawGlyphOutlineGG(gc, outline, scale, ox, oy)
	ggImg := gc.Image()
	if !shapedGlyphCacheHasInk(ggImg) {
		t.Fatal("gg produced no ink")
	}

	ops, pts := packGlyphOutline(outline)
	edges := make([]float32, 8192)
	xs := make([]float32, 128)
	dirs := make([]int32, 128)
	rgba := make([]byte, w*h*4)
	ink := fxglyph.RasterOutline(ops, pts, w, h, float32(scale), float32(ox), float32(oy), edges, xs, dirs, rgba)
	if ink < 1 {
		t.Fatalf("fx ink=%d", ink)
	}
	fxImg := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(fxImg.Pix, rgba)
	if !shapedGlyphCacheHasInk(fxImg) {
		t.Fatal("fx produced no ink")
	}

	ggN := countOpaque(ggImg)
	fxN := countOpaque(fxImg)
	// Even-odd vs non-zero: allow wide band; both must be in the same order of magnitude.
	if fxN < ggN/4 || fxN > ggN*4 {
		t.Fatalf("ink count mismatch gg=%d fx=%d (ratio out of band)", ggN, fxN)
	}
}

func TestFxGlyphEnvDualPath(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	if !fxGlyphBackendActive() {
		t.Fatal("fx glyph backend should be active by default in native build")
	}
	prev := os.Getenv("GRU_FX_GLYPH")
	t.Setenv("GRU_FX_GLYPH", "0")
	t.Cleanup(func() {
		if prev == "" {
			_ = os.Unsetenv("GRU_FX_GLYPH")
		} else {
			_ = os.Setenv("GRU_FX_GLYPH", prev)
		}
	})
	if fxGlyphBackendActive() {
		t.Fatal("fx glyph backend should be off when GRU_FX_GLYPH=0")
	}
	face := pickShapedFace(false, false, false, false)
	gid, ok := face.NominalGlyph('B')
	if !ok {
		t.Fatal("no glyph")
	}
	img, ok := rasterShapedGlyphImage(face, gid, 18)
	if !ok || !shapedGlyphCacheHasInk(img) {
		t.Fatal("dual-path raster failed")
	}
}

func countOpaque(img image.Image) int {
	b := img.Bounds()
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				n++
			}
		}
	}
	return n
}

func BenchmarkShapedGlyphRasterGG(b *testing.B) {
	if !InitShapedFonts() {
		b.Skip("no TTF faces")
	}
	face := pickShapedFace(false, false, false, false)
	gid, ok := face.NominalGlyph('A')
	if !ok {
		b.Fatal("no glyph")
	}
	outline, ok := shapedGlyphOutline(face, gid)
	if !ok {
		b.Fatal("no outline")
	}
	ext, _ := face.GlyphExtents(gid)
	scale, w, h, ox, oy := shapedGlyphRasterLayout(outline, ext, shapedGlyphRasterPx(16), face.Upem())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gc := gg.NewContext(w, h)
		gc.SetRGBA(0, 0, 0, 0)
		gc.Clear()
		gc.SetRGBA(1, 1, 1, 1)
		drawGlyphOutlineGG(gc, outline, scale, ox, oy)
		_ = gc.Image()
	}
}

func BenchmarkShapedGlyphRasterFX(b *testing.B) {
	if !InitShapedFonts() {
		b.Skip("no TTF faces")
	}
	if !fxglyph.Enabled() {
		b.Fatal("fxglyph not enabled")
	}
	face := pickShapedFace(false, false, false, false)
	gid, ok := face.NominalGlyph('A')
	if !ok {
		b.Fatal("no glyph")
	}
	outline, ok := shapedGlyphOutline(face, gid)
	if !ok {
		b.Fatal("no outline")
	}
	ext, _ := face.GlyphExtents(gid)
	scale, w, h, ox, oy := shapedGlyphRasterLayout(outline, ext, shapedGlyphRasterPx(16), face.Upem())
	ops, pts := packGlyphOutline(outline)
	edges := make([]float32, 8192)
	xs := make([]float32, 128)
	dirs := make([]int32, 128)
	rgba := make([]byte, w*h*4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fxglyph.RasterOutline(ops, pts, w, h, float32(scale), float32(ox), float32(oy), edges, xs, dirs, rgba)
	}
}
