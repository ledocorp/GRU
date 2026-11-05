// Package ui (continued) — soft polylines at SSAA 1× (charts).
package ui

import "math"

import rl "github.com/gen2brain/raylib-go/raylib"

// drawCoveragePolyline strokes a polyline with coverage-AA capsules.
// Targeted geometry AA (same family as gauges/rings) — not full-frame MSAA.
// MSAA helps triangle edges but is costly, inconsistent across GPUs, and does
// not replace ClearType for text; prefer coverage stamps for charts/chrome.
func drawCoveragePolyline(pts []rl.Vector2, thickness float32, col rl.Color) {
	if len(pts) < 2 || thickness < 0.5 || col.A == 0 {
		return
	}
	r := thickness * 0.5
	if r < 0.55 {
		r = 0.55
	}
	if r > 1.1 {
		r = 1.1
	}
	for i := 1; i < len(pts); i++ {
		drawCoverageLineSeg(pts[i-1].X, pts[i-1].Y, pts[i].X, pts[i].Y, r, col)
	}
}

func drawCoverageLineSeg(x0, y0, x1, y1, radius float32, col rl.Color) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	length := math.Hypot(dx, dy)
	if length < 0.01 {
		DrawCoverageCircle(x0, y0, radius, col)
		return
	}
	// Denser stamps = smoother diagonals without a thicker look.
	steps := int(length/0.4 + 0.5)
	if steps < 1 {
		steps = 1
	}
	inv := 1.0 / float64(steps)
	for i := 0; i <= steps; i++ {
		t := float64(i) * inv
		DrawCoverageCircle(x0+float32(dx*t), y0+float32(dy*t), radius, col)
	}
}

// insetPolylineToPlot keeps stroke capsules inside the plot frame (end caps
// used to straddle the grid border).
func insetPolylineToPlot(pts []rl.Vector2, plot rl.Rectangle, pad float32) {
	if pad < 0.5 || len(pts) == 0 {
		return
	}
	minX := plot.X + pad
	maxX := plot.X + plot.Width - pad
	minY := plot.Y + pad
	maxY := plot.Y + plot.Height - pad
	if maxX <= minX || maxY <= minY {
		return
	}
	for i := range pts {
		if pts[i].X < minX {
			pts[i].X = minX
		}
		if pts[i].X > maxX {
			pts[i].X = maxX
		}
		if pts[i].Y < minY {
			pts[i].Y = minY
		}
		if pts[i].Y > maxY {
			pts[i].Y = maxY
		}
	}
}
