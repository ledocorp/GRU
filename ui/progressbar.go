// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ProgressBar is a non-interactive widget that displays a horizontal fill bar.
//
// Value is a Signal[float32] in [0.0, 1.0]. When Value changes, the bar
// updates reactively on the next frame. The fill color is taken from a
// "progress-fill" key in CurrentTheme (falls back to an indigo default),
// and the track color comes from the widget's own style (default "progress").
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
	// Value is the fill level in [0.0, 1.0]. Values outside this range are clamped.
	Value *Signal[float32]
}

// UsesScissor implements Node.UsesScissor.
// ProgressBar clips the fill region via BeginScissorMode.
func (pb *ProgressBar) UsesScissor() bool { return true }

// NewProgressBar creates a ProgressBar with the given initial value.
func NewProgressBar(id string, initialValue float32, x, y, w, h float32) *ProgressBar {
	if initialValue < 0 {
		initialValue = 0
	}
	if initialValue > 1 {
		initialValue = 1
	}
	pb := &ProgressBar{
		Element: NewElement(id, x, y, w, h),
		Value:   NewSignal(initialValue),
	}
	pb.styleName = "progress"
	pb.Value.Subscribe(func() { pb.MarkDirty() })
	return pb
}

// Layout implements Node.Layout (no-op for leaf widgets).
func (pb *ProgressBar) Layout() { pb.layoutDirty = false }

// Draw implements Node.Draw.
func (pb *ProgressBar) Draw() {
	pb.drawInternal()
}

// drawInternal renders the track and the filled portion.
func (pb *ProgressBar) drawInternal() {
	if pb.IsHidden() {
		return
	}

	style := pb.GetStyle()
	bounds := pb.Bounds()

	// Pre-compute roundness for pill / rounded shape
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

	// Track background
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

	// Fill colour and roundness — use "progress-fill" theme entry if present
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

	// Clamp value and compute fill width
	v := pb.Value.Get()
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}

	// Determine ancestor Viewport clip so every draw call in this widget is
	// constrained to the scroll container's visible area.
	var vpClip rl.Rectangle
	hasVP := false
	if vp := findViewport(pb); vp != nil {
		vpClip = vp.ClipBounds()
		hasVP = true
	}

	// Re-apply the Viewport scissor (or disable clipping) after a sub-scissor.
	restoreVP := func() {
		if hasVP {
			beginScissorMode(int32(vpClip.X), int32(vpClip.Y), int32(vpClip.Width), int32(vpClip.Height))
		}
	}

	fillW := bounds.Width * v
	if fillW > 0 {
		// Scissor to fill width intersected with the Viewport clip so the fill
		// never bleeds outside the scroll container on partial scroll-off.
		fillRect := rl.NewRectangle(bounds.X, bounds.Y, fillW, bounds.Height)
		if hasVP {
			fillRect = intersectRects(fillRect, vpClip)
		}
		if fillRect.Width > 0 && fillRect.Height > 0 {
			beginScissorMode(int32(fillRect.X), int32(fillRect.Y), int32(fillRect.Width), int32(fillRect.Height))
			if fillRoundness > 0 {
				rl.DrawRectangleRounded(bounds, fillRoundness, 8, fillColor)
			} else {
				rl.DrawRectangleRec(rl.NewRectangle(bounds.X, bounds.Y, fillW, bounds.Height), fillColor)
			}
			rl.EndScissorMode()
			restoreVP()
		}
	}

	// Percentage label — only when bar is tall enough to be legible (h >= 18)
	if bounds.Height >= 18 {
		const pctFontSize = int32(13)
		pct := int32(v * 100)
		label := fmt.Sprintf("%d%%", pct)
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
