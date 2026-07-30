// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	breadcrumbSep      = " / "
	breadcrumbDefaultH = float32(36)
	breadcrumbPadX     = float32(12)
	breadcrumbPadY     = float32(8)
)

// Breadcrumbs shows a clickable path (e.g. Home / Settings / Wi-Fi).
//
// The last segment is the current page (not clickable). Earlier segments call
// OnClick with their index when clicked.
//
// # LLM Prompt Template
//
//	bc := ui.NewBreadcrumbs("crumbs", []string{"Inbox", "Message 3"}, 0, 0, 0, 0)
//	bc.OnClick = func(i int) { popTo(i) }
//	toolbar.AddChild(bc)
//
// Demo scenes: **Batch 17 Breadcrumbs**, **List Detail Demo**.
type Breadcrumbs struct {
	Element
	Items    []string
	OnClick  func(index int)
	hoverIdx int
}

// NewBreadcrumbs creates a breadcrumb trail. Pass w=0 for intrinsic width.
func NewBreadcrumbs(id string, items []string, x, y, w, h float32) *Breadcrumbs {
	if len(items) == 0 {
		items = []string{"Home"}
	}
	bc := &Breadcrumbs{
		Element:  NewElement(id, x, y, w, h),
		Items:    items,
		hoverIdx: -1,
	}
	bc.styleName = "breadcrumbs"
	if h == 0 {
		bc.AutoHeight = true
	}
	if w == 0 {
		b := bc.Bounds()
		b.Width = bc.measureWidth()
		bc.setBoundsNoMark(b)
	}
	return bc
}

// SetItems replaces the path and refreshes intrinsic width when w was auto-sized.
func (bc *Breadcrumbs) SetItems(items []string) {
	if len(items) == 0 {
		items = []string{"Home"}
	}
	bc.Items = items
	b := bc.Bounds()
	if b.Width == 0 || bc.IsAutoHeight() {
		b.Width = bc.measureWidth()
		bc.setBoundsNoMark(b)
	}
	bc.MarkDirty()
}

func (bc *Breadcrumbs) breadcrumbPad() (padX, padY float32) {
	padX, padY = breadcrumbPadX, breadcrumbPadY
	style := bc.GetStyle()
	if style.Padding > 0 {
		padX = style.Padding
		padY = style.Padding * 0.65
		if padY < breadcrumbPadY {
			padY = breadcrumbPadY
		}
	}
	return padX, padY
}

func (bc *Breadcrumbs) measureWidth() float32 {
	style := bc.GetStyle()
	padX, _ := bc.breadcrumbPad()
	sepW := breadcrumbTextWidth(breadcrumbSep, style)
	var total float32
	for i, item := range bc.Items {
		if i > 0 {
			total += sepW
		}
		total += breadcrumbTextWidth(item, style)
	}
	return total + padX*2
}

// breadcrumbTextWidth measures label width; uses a rune estimate when fonts are not loaded (headless tests).
func breadcrumbTextWidth(text string, style Style) float32 {
	w := float32(measureTextS(text, style))
	if w > 0 {
		return w
	}
	fs := EffectiveFontSize(style)
	if fs <= 0 {
		fs = 14
	}
	return float32(len(text)) * fs * 0.52
}

func (bc *Breadcrumbs) segmentAt(x float32) int {
	style := bc.GetStyle()
	padX, _ := bc.breadcrumbPad()
	sepW := float32(measureTextS(breadcrumbSep, style))
	cursor := bc.Bounds().X + padX
	last := len(bc.Items) - 1
	for i, item := range bc.Items {
		if i > 0 {
			cursor += sepW
		}
		tw := float32(measureTextS(item, style))
		if x >= cursor && x < cursor+tw {
			if i < last {
				return i
			}
			return -1
		}
		cursor += tw
	}
	return -1
}

// IsInteractive implements Node.
func (bc *Breadcrumbs) IsInteractive() bool { return bc.OnClick != nil && len(bc.Items) > 1 }

// Update handles segment clicks and hover.
func (bc *Breadcrumbs) Update(_ float32) {
	if bc.IsHidden() || bc.OnClick == nil || len(bc.Items) < 2 {
		return
	}
	mouse := rl.GetMousePosition()
	prevHover := bc.hoverIdx
	bc.hoverIdx = -1
	if rl.CheckCollisionPointRec(mouse, bc.Bounds()) {
		bc.hoverIdx = bc.segmentAt(mouse.X)
	}
	if bc.hoverIdx != prevHover {
		bc.MarkDrawDirty()
	}
	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) || bc.hoverIdx < 0 {
		return
	}
	bc.OnClick(bc.hoverIdx)
}

// Layout clamps height when AutoHeight.
func (bc *Breadcrumbs) Layout() {
	defer func() { bc.layoutDirty = false }()
	if !bc.IsAutoHeight() {
		return
	}
	b := bc.Bounds()
	wantH := breadcrumbDefaultH
	if style := bc.GetStyle(); style.FontSize > 0 {
		fs := EffectiveFontSize(style)
		padX, padY := bc.breadcrumbPad()
		_ = padX
		if h := fs + 2*padY; h > wantH {
			wantH = h
		}
	}
	if b.Height < wantH-0.5 || b.Height > wantH+0.5 {
		b.Height = wantH
		bc.setBoundsNoMark(b)
	}
}

// Draw implements Node.Draw.
func (bc *Breadcrumbs) Draw() { bc.drawInternal() }

func (bc *Breadcrumbs) drawInternal() {
	if bc.IsHidden() {
		return
	}
	b := bc.Bounds()
	style := bc.GetStyle()
	if style.FontSize <= 0 {
		style.FontSize = 14
	}
	sepStyle := style
	sepStyle.TextColor = rl.NewColor(160, 165, 185, 255)

	padX, _ := bc.breadcrumbPad()
	x := b.X + padX
	y := TextPosY(b, style)
	last := len(bc.Items) - 1
	for i, item := range bc.Items {
		if i > 0 {
			drawTextS(breadcrumbSep, int32(x), y, sepStyle)
			x += float32(measureTextS(breadcrumbSep, sepStyle))
		}
		labelStyle := style
		if i == last {
			labelStyle.Bold = true
			labelStyle.TextColor = rl.NewColor(30, 32, 50, 255)
		} else {
			labelStyle.TextColor = rl.NewColor(79, 70, 229, 255)
			if i == bc.hoverIdx {
				labelStyle.TextColor = rl.NewColor(99, 90, 249, 255)
			}
		}
		drawTextS(item, int32(x), y, labelStyle)
		x += float32(measureTextS(item, labelStyle))
	}
}

// InteractionOverlayActive implements InteractionOverlayPainter.
func (bc *Breadcrumbs) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (bc *Breadcrumbs) DrawInteractionOverlay() {}
