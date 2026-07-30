// Package ui — body width prep and clamp (docs/LAYOUT_CONTRACTS.md §6).
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// prepareBodySubtreeWidths runs width-first label fit and cap before flex layout.
func prepareBodySubtreeWidths(innerW float32, children []Node) {
	if innerW < 1 {
		return
	}
	fitSubtreeLabels(innerW, children)
	capSubtreeWidths(innerW, children)
}

// clampBodyChildren clips children to the padded body band after Layout().
func clampBodyChildren(bodyRect rl.Rectangle, padding float32, children []Node, fixedBody bool) {
	layoutBodyLabels(children, fixedBody)
	clampChildrenToBodyRect(children, bodyRect, padding)
}

// finalizeSplitPaneContainment clamps pane content after SplitView assigns pane bounds.
func finalizeSplitPaneContainment(n Node, pane rl.Rectangle) {
	switch v := n.(type) {
	case *Panel:
		style := v.GetStyle()
		titleOff := v.bodyTitleHeight()
		bodyH := pane.Height - titleOff
		if bodyH < 0 {
			bodyH = 0
		}
		bodyRect := rl.NewRectangle(pane.X, pane.Y+titleOff, pane.Width, bodyH)
		clampBodyChildren(bodyRect, style.Padding, v.Children(), true)
	case *Card:
		style := v.GetStyle()
		titleOff := v.bodyTitleHeight()
		bodyH := pane.Height - titleOff
		if bodyH < 0 {
			bodyH = 0
		}
		bodyRect := rl.NewRectangle(pane.X, pane.Y+titleOff, pane.Width, bodyH)
		clampBodyChildren(bodyRect, style.Padding, v.Children(), true)
	}
}

// capSubtreeWidths shrinks nodes wider than maxW so flex rows wrap at narrow widths.
func capSubtreeWidths(maxW float32, nodes []Node) {
	if maxW < 1 {
		return
	}
	for _, ch := range nodes {
		if ch.IsHidden() {
			continue
		}
		b := ch.Bounds()
		if b.Width > maxW {
			b.Width = maxW
			layoutSetBounds(ch, b)
			ch.MarkDirty()
			if rt, ok := ch.(*RichText); ok {
				rt.InvalidateAutoHeightMeasure()
			}
		}
		kids := ch.Children()
		if len(kids) == 0 {
			continue
		}
		innerW := maxW
		if c, ok := ch.(*Container); ok {
			pad := c.GetStyle().Padding
			innerW = maxW - 2*pad
			if innerW < 1 {
				innerW = maxW
			}
		}
		capSubtreeWidths(innerW, kids)
	}
}
