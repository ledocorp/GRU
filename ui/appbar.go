// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	appBarPadX      float32 = 12
	appBarTitleGap  float32 = 10
	appBarTrailingGap float32 = 8
	appBarDefaultH  float32 = 52
	appBarLeadingW  float32 = 40
)

// AppBar is a top application bar with a title and optional leading/trailing actions.
//
// Trailing nodes are laid out right-to-left in insertion order. Leading and trailing
// children are updated and drawn by the AppBar; add interactive widgets (IconButton,
// Button) via SetLeading / AddTrailing.
//
// # LLM Prompt Template
//
//	bar := ui.NewAppBar("top", "Settings", 0, 0, 0, 0)
//	bar.AddTrailing(ui.NewIconButton("save", "✓", "Save", 0, 0, 88, 36))
//	shell.AddChild(bar)
//
// Demo scenes: **Settings Demo**, **List Detail Demo**, Notepad chrome.
type AppBar struct {
	Element
	Title    string
	TitleSig *Signal[string]
	Leading  Node
	Trailing []Node
}

// NewAppBar creates an app bar. Pass w=0, h=0 for flex width and intrinsic height.
func NewAppBar(id, title string, x, y, w, h float32) *AppBar {
	ab := &AppBar{
		Element:  NewElement(id, x, y, w, h),
		Title:    title,
		TitleSig: NewSignal(title),
	}
	ab.styleName = "appbar"
	if h == 0 {
		ab.AutoHeight = true
	}
	ab.TitleSig.Subscribe(func() { ab.MarkDirty() })
	return ab
}

// SetLeading assigns the optional left action (menu, back, etc.).
func (ab *AppBar) SetLeading(n Node) { ab.Leading = n; ab.MarkDirty() }

// AddTrailing appends a right-side action widget.
func (ab *AppBar) AddTrailing(n Node) {
	ab.Trailing = append(ab.Trailing, n)
	ab.MarkDirty()
}

// IsInteractive implements Node — children handle input; bar itself is chrome.
func (ab *AppBar) IsInteractive() bool { return false }

// Update advances slotted children.
func (ab *AppBar) Update(dt float32) {
	if ab.IsHidden() {
		return
	}
	if ab.Leading != nil {
		ab.Leading.Update(dt)
	}
	for _, n := range ab.Trailing {
		n.Update(dt)
	}
}

// ShiftScrollSubtreeInternal moves slotted chrome when a parent Viewport scrolls.
// Leading/trailing nodes are not in Element.children, so viewport repositionOnly
// must shift them explicitly or buttons paint at stale Y.
func (ab *AppBar) ShiftScrollSubtreeInternal(dy float32) {
	if dy == 0 {
		return
	}
	shiftAppBarSlot(ab.Leading, dy)
	for _, n := range ab.Trailing {
		shiftAppBarSlot(n, dy)
	}
}

func shiftAppBarSlot(n Node, dy float32) {
	if n == nil {
		return
	}
	b := n.Bounds()
	b.Y += dy
	scrollTranslate(n, b)
	shiftSubtreeY(n.Children(), dy)
}

// Layout positions leading/trailing slots and clamps to appBarDefaultH.
func (ab *AppBar) Layout() {
	b := ab.Bounds()
	if ab.IsAutoHeight() || b.Height > appBarDefaultH+0.5 {
		b.Height = appBarDefaultH
		ab.setBoundsNoMark(b)
	}
	innerH := appBarDefaultH - 8
	innerY := b.Y + (b.Height-innerH)/2

	if ab.Leading != nil {
		lw, lh := appBarSlotSize(ab.Leading, innerH)
		lb := rl.NewRectangle(b.X+appBarPadX, innerY+(innerH-lh)/2, lw, lh)
		ab.Leading.SetBounds(lb)
		ab.Leading.Layout()
	}

	x := b.X + b.Width - appBarPadX
	for i := len(ab.Trailing) - 1; i >= 0; i-- {
		n := ab.Trailing[i]
		nw, nh := appBarSlotSize(n, innerH)
		x -= nw
		n.SetBounds(rl.NewRectangle(x, innerY+(innerH-nh)/2, nw, nh))
		n.Layout()
		x -= appBarTrailingGap
	}
	ab.layoutDirty = false
}

func appBarSlotSize(n Node, maxH float32) (w, h float32) {
	switch n.(type) {
	case *IconButton:
		h = maxH
		if h > 40 {
			h = 40
		}
		if h < 32 {
			h = 32
		}
		ib := n.(*IconButton)
		label := ib.Label.Get()
		if label == "" {
			return h, h // square icon-only control
		}
		return 88, h
	case *Button:
		h = maxH
		if h > 40 {
			h = 40
		}
		style := n.(*Button).GetStyle()
		w = float32(measureTextS(n.(*Button).Text.Get(), style)) + 28
		if w < 64 {
			w = 64
		}
		if w > 120 {
			w = 120
		}
		return w, h
	default:
		b := n.Bounds()
		if b.Width > 0 && b.Height > 0 {
			return b.Width, b.Height
		}
		return 40, maxH
	}
}

// Draw implements Node.Draw.
func (ab *AppBar) Draw() { ab.drawInternal() }

func (ab *AppBar) drawInternal() {
	if ab.IsHidden() {
		return
	}
	b := ab.Bounds()
	style := ab.GetStyle()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, style.BackgroundColor)
	}
	if style.BorderWidth > 0 && style.BorderColor.A > 0 {
		rl.DrawLineEx(
			rl.NewVector2(b.X, b.Y+b.Height),
			rl.NewVector2(b.X+b.Width, b.Y+b.Height),
			style.BorderWidth, style.BorderColor)
	}

	title := ab.Title
	if ab.TitleSig != nil {
		title = ab.TitleSig.Get()
	}
	titleStyle := style
	if titleStyle.FontSize <= 0 {
		titleStyle.FontSize = 20
	}
	titleStyle.Bold = true

	textX := b.X + appBarPadX
	if ab.Leading != nil {
		textX = ab.Leading.Bounds().X + ab.Leading.Bounds().Width + appBarTitleGap
	}
	titleY := TextPosY(b, titleStyle)
	drawTextS(title, int32(textX), titleY, titleStyle)

	if ab.Leading != nil {
		ab.Leading.Draw()
	}
	for _, n := range ab.Trailing {
		n.Draw()
	}
}
