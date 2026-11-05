//go:build cgo && !gru_nofx

package ui

import (
	"image"

	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/ledocorp/gru/ui/fxglyph"
)

func fxGlyphRasterEnvEnabled() bool {
	if fxOptOut("GRU_FX_GLYPH", "GORY_FX_GLYPH") {
		return false
	}
	return fxglyph.Enabled()
}

func fxGlyphBackendActive() bool {
	return fxGlyphRasterEnvEnabled() && fxglyph.Enabled()
}

// tryRasterShapedGlyphFX dual-paths outline→RGBA via fx emit-C when GRU_FX_GLYPH=1.
func tryRasterShapedGlyphFX(outline font.GlyphOutline, scale float64, w, h int, ox, oy float64) (image.Image, bool) {
	if !fxGlyphRasterEnvEnabled() || !fxglyph.Enabled() || w <= 0 || h <= 0 {
		return nil, false
	}
	ops, pts := packGlyphOutline(outline)
	if len(ops) == 0 {
		return nil, false
	}
	edgeCap := len(ops) * 128
	if edgeCap < 8192 {
		edgeCap = 8192
	}
	edges := make([]float32, edgeCap)
	xs := make([]float32, 256)
	dirs := make([]int32, 256)
	rgba := make([]byte, w*h*4)
	ink := fxglyph.RasterOutline(
		ops, pts, w, h,
		float32(scale), float32(ox), float32(oy),
		edges, xs, dirs, rgba,
	)
	if ink < 1 {
		return nil, false
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, rgba)
	return img, true
}

func packGlyphOutline(outline font.GlyphOutline) (ops []int32, pts []float32) {
	for _, seg := range outline.Segments {
		args := seg.ArgsSlice()
		switch seg.Op {
		case ot.SegmentOpMoveTo:
			if len(args) < 1 {
				continue
			}
			ops = append(ops, 0)
			pts = append(pts, args[0].X, args[0].Y)
		case ot.SegmentOpLineTo:
			if len(args) < 1 {
				continue
			}
			ops = append(ops, 1)
			pts = append(pts, args[0].X, args[0].Y)
		case ot.SegmentOpQuadTo:
			if len(args) < 2 {
				continue
			}
			ops = append(ops, 2)
			pts = append(pts, args[0].X, args[0].Y, args[1].X, args[1].Y)
		case ot.SegmentOpCubeTo:
			if len(args) < 3 {
				continue
			}
			ops = append(ops, 3)
			pts = append(pts, args[0].X, args[0].Y, args[1].X, args[1].Y, args[2].X, args[2].Y)
		}
	}
	return ops, pts
}
