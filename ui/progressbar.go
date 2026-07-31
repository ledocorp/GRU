// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ProgressBar is a non-interactive widget that displays a horizontal fill bar.
//
// Value is a Signal[float32] in [0.0, 1.0]. Continuous demos should call
// [ProgressBar.SetLiveAnimation](true) while Value advances: the SSAA cache
// keeps a static track (like Spinner), and fill + percent repaint together in
// [DrawAnimationOverlay] at 1× so AnimationFPS (36) does not drop the fill.
//
// # LLM Prompt Template
//
//	pct := ui.NewSignal(float32(0.65))
//	bar := ui.NewProgressBar("upload", pct.Get(), 0, 0, 400, 24)
//	pct.Subscribe(func() { bar.Value.Set(pct.Get()) })
//	panel.AddChild(bar)
//
// Demo scenes: **Batch 25 ProgressBar**.
type ProgressBar struct {
	Element
	Value *Signal[float32]
	live  bool
}

// UsesScissor — false, matching Spinner. Nested BeginScissorMode inside
// DrawAnimationOverlay clears the overlay clip and the indigo fill vanishes
// when idle drops to AnimationFPS (cache-hit frames).
func (pb *ProgressBar) UsesScissor() bool { return false }

// NewProgressBar creates a ProgressBar with the given initial value.
func NewProgressBar(id string, initialValue float32, x, y, w, h float32) *ProgressBar {
	initialValue = progressClamp01(initialValue)
	pb := &ProgressBar{
		Element: NewElement(id, x, y, w, h),
		Value:   NewSignal(initialValue),
	}
	pb.styleName = "progress"
	pb.Value.Subscribe(func() {
		if pb.live {
			return
		}
		pb.MarkDrawDirty()
	})
	return pb
}

// SetLiveAnimation keeps fill updates off the document dirty path (AnimationFPS).
// Call false when the run ends to bake the final fill into the SSAA cache.
func (pb *ProgressBar) SetLiveAnimation(on bool) {
	if pb.live == on {
		return
	}
	pb.live = on
	pb.MarkDrawDirty()
}

// AnimationActive implements AnimationReporter.
func (pb *ProgressBar) AnimationActive() bool {
	return pb.live && !pb.IsHidden()
}

// AnimationSource implements AnimationReporter.
func (pb *ProgressBar) AnimationSource() string { return pb.ID() }

// Layout implements Node.Layout.
func (pb *ProgressBar) Layout() { pb.layoutDirty = false }

// Draw implements Node.Draw — Spinner-style: track only while live.
func (pb *ProgressBar) Draw() {
	if pb.AnimationActive() {
		pb.paint(0, false)
		return
	}
	pb.paint(progressClamp01(pb.Value.Get()), true)
}

// DrawAnimationOverlay redraws track + fill + label together at 1× (Spinner).
func (pb *ProgressBar) DrawAnimationOverlay() {
	if !pb.AnimationActive() {
		return
	}
	pb.paint(progressClamp01(pb.Value.Get()), true)
}

func progressClamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (pb *ProgressBar) paint(v float32, showFill bool) {
	if pb.IsHidden() {
		return
	}
	style := pb.GetStyle()
	bounds := pb.Bounds()
	if bounds.Width < 1 || bounds.Height < 1 {
		return
	}

	shorter := bounds.Width
	if bounds.Height < shorter {
		shorter = bounds.Height
	}
	roundness := float32(0)
	if style.CornerRadius > 0 && shorter > 0 {
		roundness = style.CornerRadius / (shorter / 2)
		if roundness > 1 {
			roundness = 1
		}
	}

	if roundness > 0 {
		rl.DrawRectangleRounded(bounds, roundness, 8, style.BackgroundColor)
	} else {
		rl.DrawRectangleRec(bounds, style.BackgroundColor)
	}
	if style.BorderWidth > 0 {
		if roundness > 0 {
			rl.DrawRectangleRoundedLinesEx(bounds, roundness, 8, style.BorderWidth, style.BorderColor)
		} else {
			rl.DrawRectangleLinesEx(bounds, style.BorderWidth, style.BorderColor)
		}
	}
	if !showFill {
		return
	}

	fillColor := rl.NewColor(79, 70, 229, 255)
	fillRoundness := roundness
	if fs, ok := CurrentTheme["progress-fill"]; ok {
		fillColor = fs.BackgroundColor
		if fs.CornerRadius > 0 && shorter > 0 {
			fr := fs.CornerRadius / (shorter / 2)
			if fr > 1 {
				fr = 1
			}
			fillRoundness = fr
		}
	}

	// No nested scissor — overlay path already clips via DrawAnimationOverlays.
	fillW := bounds.Width * v
	if fillW >= 0.5 {
		fill := rl.NewRectangle(bounds.X, bounds.Y, fillW, bounds.Height)
		if fillRoundness > 0 {
			rl.DrawRectangleRounded(fill, fillRoundness, 8, fillColor)
		} else {
			rl.DrawRectangleRec(fill, fillColor)
		}
	}

	if bounds.Height >= 18 {
		const pctFontSize = int32(13)
		label := fmt.Sprintf("%d%%", int32(v*100+0.5))
		textW := measureText(label, pctFontSize)
		textX := int32(bounds.X) + (int32(bounds.Width)-textW)/2
		textY := int32(bounds.Y) + (int32(bounds.Height)-pctFontSize)/2
		textColor := style.TextColor
		if v > 0.45 {
			textColor = rl.NewColor(255, 255, 255, 210)
		}
		drawText(label, textX, textY, pctFontSize, textColor)
	}
}

// IsInteractive implements Node.IsInteractive (false — display only).
func (pb *ProgressBar) IsInteractive() bool { return false }
