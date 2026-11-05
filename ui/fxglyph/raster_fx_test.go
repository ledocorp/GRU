//go:build cgo && !gru_nofx

package fxglyph_test

import (
	"testing"

	"github.com/ledocorp/gru/ui/fxglyph"
)

func TestRasterOutlineSquare(t *testing.T) {
	if !fxglyph.Enabled() {
		t.Fatal("Enabled")
	}
	ops := []int32{0, 1, 1, 1, 1}
	pts := []float32{
		0, 0,
		100, 0,
		100, 100,
		0, 100,
		0, 0,
	}
	w, h := 4, 4
	scale := float32(3) / 100
	ox := float32(0.5)
	oy := float32(3.5)
	edges := make([]float32, 256)
	xs := make([]float32, 64)
	dirs := make([]int32, 64)
	rgba := make([]byte, w*h*4)
	ink := fxglyph.RasterOutline(ops, pts, w, h, scale, ox, oy, edges, xs, dirs, rgba)
	if ink < 1 {
		t.Fatalf("ink=%d", ink)
	}
	var alpha int
	for i := 3; i < len(rgba); i += 4 {
		if rgba[i] != 0 {
			alpha++
		}
	}
	if alpha < 1 {
		t.Fatal("no alpha pixels")
	}
}

func TestRasterOutlineCoverageAA(t *testing.T) {
	if !fxglyph.Enabled() {
		t.Fatal("Enabled")
	}
	// Axis-aligned unit square in font units → soft edge alphas (not binary 0/255 only).
	ops := []int32{0, 1, 1, 1, 1}
	pts := []float32{
		0, 0,
		100, 0,
		100, 100,
		0, 100,
		0, 0,
	}
	w, h := 8, 8
	scale := float32(6) / 100
	// Non-integer origin so edges land between pixels → greyscale coverage.
	ox := float32(1.3)
	oy := float32(7.3)
	edges := make([]float32, 256)
	xs := make([]float32, 64)
	dirs := make([]int32, 64)
	rgba := make([]byte, w*h*4)
	ink := fxglyph.RasterOutline(ops, pts, w, h, scale, ox, oy, edges, xs, dirs, rgba)
	if ink < 1 {
		t.Fatalf("ink=%d", ink)
	}
	var soft, solid int
	for i := 3; i < len(rgba); i += 4 {
		a := rgba[i]
		if a == 0 {
			continue
		}
		if a < 255 {
			soft++
		} else {
			solid++
		}
	}
	if soft < 1 {
		t.Fatalf("expected coverage AA edge pixels (soft=%d solid=%d)", soft, solid)
	}
	if solid < 1 {
		t.Fatalf("expected interior solid pixels (soft=%d solid=%d)", soft, solid)
	}
}
