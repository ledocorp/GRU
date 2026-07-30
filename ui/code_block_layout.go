// Package ui (continued) — markdown / document code fence layout helpers.
package ui

// WrapCodeBlockHorizontalScroll wraps a non-wrapping code RichText in a horizontal
// viewport so long lines scroll instead of clipping at the card edge.
func WrapCodeBlockHorizontalScroll(id string, rt *RichText) Node {
	// Estimate width from rune counts — avoid MeasureContentWidth during markdown
	// build (that forces SDF font load / glyph measure and stalls the demo).
	contentW := estimateCodeRichTextWidth(rt)
	lane := NewContainer(id+"-lane", 0, 0, 0, 0)
	lane.AutoHeight = true
	lane.MinWidth = contentW
	lane.PreferredWidth = contentW
	lane.SetStyle("transparent")
	lane.AddChild(rt)

	scroll := NewHorizontalViewport(id+"-hscroll", 0, 0, 0, 0)
	scroll.AutoHeight = true
	scroll.Gap = 0
	scroll.SetStyle("list-flush")
	scroll.AddChild(lane)
	return scroll
}

func estimateCodeRichTextWidth(rt *RichText) float32 {
	if rt == nil {
		return 120
	}
	style := rt.GetStyle()
	fs := float32(style.FontSize)
	if fs < 8 {
		fs = 14
	}
	pad := float32(style.Padding) * 2
	avg := fs * 0.62 // mono advance estimate
	maxCols := 8
	cols := 0
	for _, sp := range rt.Spans {
		for _, r := range sp.Text {
			if r == '\n' {
				if cols > maxCols {
					maxCols = cols
				}
				cols = 0
				continue
			}
			cols++
		}
	}
	if cols > maxCols {
		maxCols = cols
	}
	w := float32(maxCols)*avg + pad
	if w < 80 {
		w = 80
	}
	return w
}
