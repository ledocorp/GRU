package ui

// prepareAccordionBodyContent reflows accordion body copy at the definitive
// content width. Labels in column/stack layouts get wrap + AutoHeight so resize
// reflows; flex-row rows (badges, form rows) keep intrinsic widths.
func prepareAccordionBodyContent(contentW float32, nodes []Node) {
	if contentW < 1 {
		return
	}
	fitSubtreeLabels(contentW, nodes)
	capSubtreeWidths(contentW, nodes)
	for _, ch := range nodes {
		if !ch.IsHidden() {
			prepareAccordionBodyNode(ch, contentW, false)
		}
	}
}

func prepareAccordionBodyNode(ch Node, contentW float32, inFlexRow bool) {
	if ch.IsHidden() {
		return
	}
	if row, ok := ch.(*Container); ok {
		rowW := contentW
		pad := row.GetStyle().Padding
		rowW = contentW - 2*pad
		if rowW < 1 {
			rowW = contentW
		}
		b := row.Bounds()
		if b.Width == 0 || b.Width > contentW {
			b.Width = contentW
			layoutSetBounds(row, b)
			row.MarkDirty()
		}
		childInRow := row.FlexDirection == FlexRow
		for _, kid := range row.Children() {
			prepareAccordionBodyNode(kid, rowW, inFlexRow || childInRow)
		}
		return
	}
	if lbl, ok := ch.(*Label); ok {
		if inFlexRow {
			return
		}
		b := lbl.Bounds()
		b.Width = contentW
		lbl.Align = LabelAlignLeft
		lbl.Wrap = true
		lbl.Truncate = false
		lbl.AutoHeight = true
		layoutSetBounds(lbl, b)
		lbl.MarkDirty()
		return
	}
	b := ch.Bounds()
	if b.Width == 0 || b.Width > contentW {
		b.Width = contentW
		layoutSetBounds(ch, b)
		ch.MarkDirty()
	}
	for _, kid := range ch.Children() {
		prepareAccordionBodyNode(kid, contentW, inFlexRow)
	}
}
