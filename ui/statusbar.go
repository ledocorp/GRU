// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	statusBarDefaultH   = float32(32.5) // 26px base + 25% chrome height; font size unchanged
	statusBarPadX         = float32(14)
	statusBarSectionGap = float32(16)
	statusBarDividerW     = float32(1)
	statusBarDividerGap   = float32(4)
	statusBarDividerInset = float32(6)
)

// StatusBarColumn is a weighted lane in column layout (toolbar-style proportions).
type StatusBarColumn struct {
	Weight float32    // relative width; 0 treated as 1
	Align  LabelAlign // default LabelAlignLeft
	Nodes  []Node
}

// StatusBar is a bottom desktop status strip with left, center, and right slots.
// Typical content: status text, mode indicator, cursor position.
//
// # LLM Prompt Template
//
//	sb := ui.NewStatusBar("status", 0, 0, 0, 0)
//	sb.AddLeft(ui.NewLabel("ready", "Ready", 0, 0, 0, 0))
//	sb.AddRight(ui.NewLabel("pos", "Ln 1, Col 1", 0, 0, 0, 0))
//	shell.AddChild(sb)
//
// Demo scenes: **List Pane Demo**, Notepad status bar.
type StatusBar struct {
	Element
	Left   []Node
	Center []Node
	Right  []Node
	// Columns enables weighted lanes with vertical dividers (see SetColumns).
	Columns    []StatusBarColumn
	dividerX   []float32
	columnRects []rl.Rectangle
}

// NewStatusBar creates a status bar. Pass w=0, h=0 for flex width and intrinsic height.
func NewStatusBar(id string, x, y, w, h float32) *StatusBar {
	sb := &StatusBar{
		Element: NewElement(id, x, y, w, h),
	}
	sb.styleName = "statusbar"
	if h == 0 {
		sb.AutoHeight = true
	}
	return sb
}

// AddLeft appends a widget to the leading section.
func (sb *StatusBar) AddLeft(n Node) {
	if n != nil {
		sb.Left = append(sb.Left, n)
		sb.MarkDirty()
	}
}

// AddCenter appends a widget to the centered section.
func (sb *StatusBar) AddCenter(n Node) {
	if n != nil {
		sb.Center = append(sb.Center, n)
		sb.MarkDirty()
	}
}

// AddRight appends a widget to the trailing section.
func (sb *StatusBar) AddRight(n Node) {
	if n != nil {
		sb.Right = append(sb.Right, n)
		sb.MarkDirty()
	}
}

// SetColumns replaces Left/Center/Right with weighted columns and vertical dividers.
func (sb *StatusBar) SetColumns(cols []StatusBarColumn) {
	sb.Columns = cols
	sb.Left = nil
	sb.Center = nil
	sb.Right = nil
	sb.MarkDirty()
}

// IsInteractive implements Node — slotted children may be interactive.
func (sb *StatusBar) IsInteractive() bool {
	if len(sb.Columns) > 0 {
		for _, c := range sb.Columns {
			for _, n := range c.Nodes {
				if n.IsInteractive() {
					return true
				}
			}
		}
		return false
	}
	return len(sb.Left) > 0 || len(sb.Center) > 0 || len(sb.Right) > 0
}

// Update advances slotted children.
func (sb *StatusBar) Update(dt float32) {
	if sb.IsHidden() {
		return
	}
	for _, n := range sb.Left {
		n.Update(dt)
	}
	for _, n := range sb.Center {
		n.Update(dt)
	}
	for _, n := range sb.Right {
		n.Update(dt)
	}
	if len(sb.Columns) > 0 {
		for _, col := range sb.Columns {
			for _, n := range col.Nodes {
				n.Update(dt)
			}
		}
	}
}

// Layout positions left/center/right sections on one row.
func (sb *StatusBar) Layout() {
	if sb.IsAutoHeight() {
		b := sb.Bounds()
		if b.Height < statusBarDefaultH-0.5 || b.Height > statusBarDefaultH+0.5 {
			b.Height = statusBarDefaultH
			sb.setBoundsNoMark(b)
		}
	}
	if len(sb.Columns) > 0 {
		sb.layoutColumns()
	} else {
		sb.layoutSections()
	}
	sb.layoutDirty = false
}

func (sb *StatusBar) layoutColumns() {
	b := sb.Bounds()
	innerY := b.Y
	innerH := b.Height
	if innerH < 1 {
		innerH = 1
	}
	innerX := b.X + statusBarPadX
	innerW := b.Width - 2*statusBarPadX
	if innerW < 1 {
		innerW = 1
	}

	var totalWeight float32
	for _, c := range sb.Columns {
		w := c.Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
	}
	if totalWeight <= 0 {
		return
	}

	divCount := 0
	if len(sb.Columns) > 1 {
		divCount = len(sb.Columns) - 1
	}
	dividerSpace := float32(divCount) * (statusBarDividerW + statusBarDividerGap*2)
	usable := innerW - dividerSpace
	if usable < 1 {
		usable = 1
	}

	sb.dividerX = sb.dividerX[:0]
	sb.columnRects = sb.columnRects[:0]
	x := innerX
	for i, col := range sb.Columns {
		if i > 0 {
			sb.dividerX = append(sb.dividerX, x+statusBarDividerGap)
			x += statusBarDividerGap*2 + statusBarDividerW
		}
		w := col.Weight
		if w <= 0 {
			w = 1
		}
		colW := usable * (w / totalWeight)
		sb.columnRects = append(sb.columnRects, rl.NewRectangle(x, innerY, colW, innerH))
		sb.placeColumn(col, x, innerY, colW, innerH)
		x += colW
	}
}

func (sb *StatusBar) placeColumn(col StatusBarColumn, x, y, w, h float32) {
	if len(col.Nodes) == 0 {
		return
	}
	gap := statusBarSectionGap
	var totalW float32
	sizes := make([]float32, len(col.Nodes))
	for i, n := range col.Nodes {
		nw, _ := statusBarSlotSize(n, h)
		sizes[i] = nw
		totalW += nw
		if i > 0 {
			totalW += gap
		}
	}
	startX := x
	switch col.Align {
	case LabelAlignCenter:
		startX = x + (w-totalW)/2
		if startX < x {
			startX = x
		}
	case LabelAlignRight:
		startX = x + w - totalW
		if startX < x {
			startX = x
		}
	}
	cx := startX
	for i, n := range col.Nodes {
		nw := sizes[i]
		if lbl, ok := n.(*Label); ok {
			lbl.AutoHeight = false
			if len(col.Nodes) == 1 {
				nw = w
				lbl.Truncate = true
			}
		}
		if _, ok := n.(*RichText); ok && len(col.Nodes) == 1 {
			nw = w
		}
		n.SetBounds(rl.NewRectangle(cx, y, nw, h))
		n.Layout()
		cx += nw + gap
	}
}

func (sb *StatusBar) layoutSections() {
	b := sb.Bounds()
	innerY := b.Y
	innerH := b.Height
	if innerH < 1 {
		innerH = 1
	}

	leftW := sb.measureSection(sb.Left)
	rightW := sb.measureSection(sb.Right)
	centerW := sb.measureSection(sb.Center)

	x := b.X + statusBarPadX
	sb.placeSection(sb.Left, x, innerY, innerH)
	x += leftW
	if len(sb.Left) > 0 && (len(sb.Center) > 0 || len(sb.Right) > 0) {
		x += statusBarSectionGap
	}

	centerX := b.X + (b.Width-centerW)/2
	if centerX < x {
		centerX = x
	}
	sb.placeSection(sb.Center, centerX, innerY, innerH)

	x = b.X + b.Width - statusBarPadX - rightW
	sb.placeSection(sb.Right, x, innerY, innerH)
}

func (sb *StatusBar) measureSection(nodes []Node) float32 {
	if len(nodes) == 0 {
		return 0
	}
	var w float32
	for i, n := range nodes {
		nw, nh := statusBarSlotSize(n, statusBarDefaultH-4)
		_ = nh
		w += nw
		if i > 0 {
			w += statusBarSectionGap
		}
	}
	return w
}

func (sb *StatusBar) placeSection(nodes []Node, x, y, h float32) {
	for _, n := range nodes {
		nw, _ := statusBarSlotSize(n, h)
		n.SetBounds(rl.NewRectangle(x, y, nw, h))
		n.Layout()
		x += nw + statusBarSectionGap
	}
}

func statusBarSlotSize(n Node, maxH float32) (w, h float32) {
	switch v := n.(type) {
	case *Label:
		style := v.GetStyle()
		fs := EffectiveFontSize(style)
		if fs <= 0 {
			fs = 14
		}
		textW := float32(measureTextS(v.Text.Get(), style))
		return textW + 2, h
	case *RichText:
		var textW float32
		for _, sp := range v.Spans {
			st := v.spanStyle(sp)
			textW += float32(measureTextS(sp.Text, st))
		}
		if textW < 1 {
			textW = 1
		}
		return textW + 2, maxH
	case *Toggle:
		b := v.Bounds()
		if b.Width > 0 {
			return b.Width, maxH
		}
		return 44, maxH
	case *ProgressBar:
		return 120, maxH - 4
	default:
		b := n.Bounds()
		if b.Width > 0 && b.Height > 0 {
			return b.Width, b.Height
		}
		return 48, maxH
	}
}

// Draw implements Node.Draw.
func (sb *StatusBar) Draw() {
	defer func() { sb.drawDirty = false }()
	sb.drawInternal()
}

func (sb *StatusBar) drawInternal() {
	if sb.IsHidden() {
		return
	}
	b := sb.Bounds()
	style := sb.GetStyle()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, style.BackgroundColor)
	}
	if style.BorderWidth > 0 && style.BorderColor.A > 0 {
		rl.DrawLineEx(
			rl.NewVector2(b.X, b.Y),
			rl.NewVector2(b.X+b.Width, b.Y),
			style.BorderWidth, style.BorderColor)
	}
	for _, n := range sb.Left {
		n.Draw()
	}
	for _, n := range sb.Center {
		n.Draw()
	}
	for _, n := range sb.Right {
		n.Draw()
	}
	if len(sb.Columns) > 0 {
		dividerCol := style.BorderColor
		if dividerCol.A == 0 {
			dividerCol = GetThemeStyle("contextmenu-divider").BackgroundColor
		}
		if dividerCol.A == 0 {
			dividerCol = rl.NewColor(218, 220, 232, 255)
		}
		innerH := b.Height
		y0 := b.Y + statusBarDividerInset
		y1 := b.Y + innerH - statusBarDividerInset
		for _, dx := range sb.dividerX {
			rl.DrawLineEx(
				rl.NewVector2(dx, y0),
				rl.NewVector2(dx, y1),
				statusBarDividerW, dividerCol)
		}
		for i, col := range sb.Columns {
			if i >= len(sb.columnRects) {
				continue
			}
			r := sb.columnRects[i]
			if r.Width < 1 || r.Height < 1 {
				continue
			}
			beginScissorMode(int32(r.X), int32(r.Y), int32(r.Width), int32(r.Height))
			for _, n := range col.Nodes {
				n.Draw()
			}
			rl.EndScissorMode()
		}
	}
}
