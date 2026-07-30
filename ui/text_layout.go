// Package ui — shared flex-column text layout for Label and RichText.
//
// Both widgets must share width-first measure, wrap, height remeasure on resize,
// and viewport restack when dimensions change. Use NewPlainText for new
// form/caption/hint copy instead of Label when text wraps in flex columns.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const flexTextMinMeasureW = float32(32)

// flexTextPadY is total vertical inset for PlainText / flex-column copy (matches labelPadY).
// Keep modest — stacked selection/status lines (Batch 6/7) read as oversized gaps when this is large.
const flexTextPadY = float32(10)

// flexTextTopPad vertically positions a text block inside flexTextPadY (matches Label).
func flexTextTopPad(boundsH, blockH float32) float32 {
	top := (boundsH - blockH) * 0.5
	const minTop float32 = 4
	if top < minTop {
		top = minTop
	}
	return top
}

// flexTextMeasure tracks the last width used for intrinsic height measurement.
type flexTextMeasure struct {
	lastW float32
}

func (m *flexTextMeasure) invalidateIfWidthChanged(oldW, newW float32) {
	if oldW > 0 && absF(newW-oldW) > 0.5 {
		m.lastW = 0
	}
}

func (m *flexTextMeasure) reset() { m.lastW = 0 }

// resolveFlexTextWidth picks the width for measure/layout (parent width when cell is still 0).
func resolveFlexTextWidth(boundsW float32, parent Node) float32 {
	w := boundsW
	if w < flexTextMinMeasureW && parent != nil {
		if pw := parent.Bounds().Width; pw >= flexTextMinMeasureW {
			w = pw
		}
	}
	return w
}

type flexAutoHeightResult struct {
	bounds    rl.Rectangle
	measuredW float32
	height    float32
	hostDirty bool
	applied   bool
}

// applyFlexAutoHeightLayout is the single contract for AutoHeight text in flex columns.
func applyFlexAutoHeightLayout(
	node Node,
	measure *flexTextMeasure,
	bounds rl.Rectangle,
	measureHeight func(width float32) float32,
) flexAutoHeightResult {
	out := flexAutoHeightResult{bounds: bounds}
	w := resolveFlexTextWidth(bounds.Width, node.ParentNode())
	if w < flexTextMinMeasureW {
		return out
	}
	if bounds.Width > w+0.5 {
		bounds.Width = w
	}
	want := measureHeight(w)
	if want < 1 {
		want = 1
	}
	widthChanged := measure.lastW > 0 && absF(w-measure.lastW) > 0.5
	if !widthChanged && measure.lastW == 0 && bounds.Width >= flexTextMinMeasureW && absF(w-bounds.Width) > 0.5 {
		widthChanged = true
	}
	heightChanged := bounds.Height < want-0.5 || bounds.Height > want+0.5
	widthAssigned := bounds.Width < flexTextMinMeasureW && w >= flexTextMinMeasureW
	if !widthChanged && !heightChanged && !widthAssigned && measure.lastW == w {
		return out
	}
	if bounds.Width < flexTextMinMeasureW || widthChanged {
		bounds.Width = w
	}
	bounds.Height = want
	measure.lastW = w
	out.bounds = bounds
	out.measuredW = w
	out.height = want
	out.hostDirty = heightChanged || widthChanged || widthAssigned
	out.applied = true
	return out
}

// styleUsesFlexPlainText is true for theme keys that should always wrap left in flex columns.
func styleUsesFlexPlainText(styleName string) bool {
	switch styleName {
	case "form-label", "form-field-caption", "form-value", "label", "richtext", "":
		return true
	default:
		return false
	}
}

// plainTextOverflowsWidth reports whether a single line exceeds the cell.
func plainTextOverflowsWidth(text string, width float32, style Style) bool {
	if text == "" || width < 8 {
		return false
	}
	return float32(measureTextS(text, style)) > width-4
}
