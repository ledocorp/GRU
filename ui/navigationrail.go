// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	navigationRailWidth    float32 = 80
	navigationRailItemH    float32 = 72
	navigationRailIconSize float32 = 24
	navigationRailIconTop  float32 = 12
	navigationRailLabelGap float32 = 4
	navigationRailActiveW  float32 = 3
)

// NavigationRail is a Material-style vertical navigation strip (icon + label per destination).
//
// Pass w=0, h=0 when used in a flex row: width is fixed at navigationRailWidth and height
// fills the cross axis from the parent. Selection is stored in Selected; optional OnChange
// runs after a tap changes the index.
//
// # LLM Prompt Template
//
//	tab := ui.NewSignal(0)
//	rail := ui.NewNavigationRail("rail", []ui.BottomNavItem{
//	    {Phosphor: ui.PhosphorHouse, Label: "Home"},
//	}, tab, 0, 0, 0, 0)
//	shell.AddChild(rail)
//
// Demo scenes: **Shell Desktop Demo**.
type NavigationRail struct {
	Element
	Items    []BottomNavItem
	Selected *Signal[int]
	OnChange func(index int)
	hoverIdx int
}

// NewNavigationRail creates a vertical navigation rail.
func NewNavigationRail(id string, items []BottomNavItem, selected *Signal[int], x, y, w, h float32) *NavigationRail {
	if selected == nil {
		selected = NewSignal(0)
	}
	if len(items) == 0 {
		items = []BottomNavItem{{Icon: "·", Label: "-"}}
	}
	if selected.Get() < 0 || selected.Get() >= len(items) {
		selected.Set(0)
	}
	nr := &NavigationRail{
		Element:  NewElement(id, x, y, w, h),
		Items:    items,
		Selected: selected,
		hoverIdx: -1,
	}
	nr.styleName = "navigationrail"
	nr.PreferredWidth = navigationRailWidth
	nr.MinWidth = navigationRailWidth
	nr.MaxWidth = navigationRailWidth
	if w == 0 {
		nr.bounds.Width = navigationRailWidth
	}
	if h == 0 {
		nr.AutoHeight = false // flex row parent stretches height to full column
	}
	nr.Selected.Subscribe(func() { nr.MarkDirty() })
	return nr
}

// IsInteractive implements Node.
func (nr *NavigationRail) IsInteractive() bool { return true }

func (nr *NavigationRail) itemBounds(i int) rl.Rectangle {
	b := nr.Bounds()
	if i < 0 || i >= len(nr.Items) {
		return b
	}
	y := b.Y + float32(i)*navigationRailItemH
	return rl.NewRectangle(b.X, y, b.Width, navigationRailItemH)
}

// Update handles item selection.
func (nr *NavigationRail) Update(_ float32) {
	if nr.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	nr.hoverIdx = -1
	for i := range nr.Items {
		if rl.CheckCollisionPointRec(mouse, nr.itemBounds(i)) {
			nr.hoverIdx = i
			break
		}
	}
	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) || nr.hoverIdx < 0 {
		return
	}
	if nr.Selected.Get() != nr.hoverIdx {
		nr.Selected.Set(nr.hoverIdx)
		if nr.OnChange != nil {
			nr.OnChange(nr.hoverIdx)
		}
	}
}

// SetBounds enforces the fixed rail width whenever flex or resize assigns bounds.
func (nr *NavigationRail) SetBounds(r rl.Rectangle) {
	if r.Width != navigationRailWidth {
		r.Width = navigationRailWidth
	}
	nr.Element.SetBounds(r)
}

// Layout keeps rail width fixed. Height comes from the flex-row parent (h=0 stretch).
func (nr *NavigationRail) Layout() {
	defer func() { nr.layoutDirty = false }()
	b := nr.Bounds()
	if b.Width != navigationRailWidth {
		b.Width = navigationRailWidth
		nr.setBoundsNoMark(b)
	}
}

// Draw implements Node.Draw.
func (nr *NavigationRail) Draw() { nr.drawInternal() }

func (nr *NavigationRail) drawInternal() {
	if nr.IsHidden() {
		return
	}
	b := nr.Bounds()
	baseStyle := nr.GetStyle()
	if baseStyle.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, baseStyle.BackgroundColor)
	}
	if baseStyle.BorderWidth > 0 && baseStyle.BorderColor.A > 0 {
		rl.DrawLineEx(
			rl.NewVector2(b.X+b.Width, b.Y),
			rl.NewVector2(b.X+b.Width, b.Y+b.Height),
			baseStyle.BorderWidth, baseStyle.BorderColor)
	}

	sel := nr.Selected.Get()
	if sel < 0 {
		sel = 0
	}
	if sel >= len(nr.Items) {
		sel = len(nr.Items) - 1
	}

	for i, item := range nr.Items {
		seg := nr.itemBounds(i)
		active := i == sel
		if active {
			accent := rl.NewColor(79, 70, 229, 255)
			rl.DrawRectangleRec(
				rl.NewRectangle(seg.X, seg.Y+8, navigationRailActiveW, seg.Height-16),
				accent)
		}

		iconColor := baseStyle.TextColor
		if active {
			iconColor = rl.NewColor(79, 70, 229, 255)
		} else if i == nr.hoverIdx {
			iconColor = rl.NewColor(55, 58, 78, 255)
		}
		iconY := seg.Y + navigationRailIconTop
		iconDst := snapPhosphorRect(rl.NewRectangle(
			seg.X+(seg.Width-navigationRailIconSize)/2,
			iconY,
			navigationRailIconSize,
			navigationRailIconSize,
		))
		if item.Phosphor != "" {
			wt := item.PhosphorWeight
			if wt == "" {
				wt = PhosphorRegular
			}
			Phosphor.EnsureLoaded(item.Phosphor, wt)
			if !Phosphor.Draw(iconDst, item.Phosphor, wt, iconColor) && item.Icon != "" {
				fallback := baseStyle
				fallback.TextColor = iconColor
				fallback.FontSize = int32(navigationRailIconSize)
				iw := measureTextS(item.Icon, fallback)
				drawTextS(item.Icon, int32(seg.X+(seg.Width-float32(iw))/2), int32(iconY), fallback)
			}
		} else if item.Icon != "" {
			fallback := baseStyle
			fallback.TextColor = iconColor
			fallback.FontSize = int32(navigationRailIconSize)
			iw := measureTextS(item.Icon, fallback)
			drawTextS(item.Icon, int32(seg.X+(seg.Width-float32(iw))/2), int32(iconY), fallback)
		}

		labelStyle := baseStyle
		if active {
			labelStyle.TextColor = rl.NewColor(79, 70, 229, 255)
			labelStyle.Bold = true
		}
		labelFS := EffectiveFontSize(labelStyle)
		labelBandY := iconY + navigationRailIconSize + navigationRailLabelGap
		labelBandH := seg.Y + seg.Height - labelBandY - 8
		if labelBandH < labelFS {
			labelBandH = labelFS
		}
		labelRect := rl.NewRectangle(seg.X, labelBandY, seg.Width, labelBandH)
		lw := measureTextS(item.Label, labelStyle)
		labelX := int32(seg.X + (seg.Width-float32(lw))/2)
		if seg.Width > 4 && labelBandH > 2 {
			beginScissorFromRect(labelRect)
			drawTextS(item.Label, labelX, TextPosY(labelRect, labelStyle), labelStyle)
			rl.EndScissorMode()
		} else {
			drawTextS(item.Label, labelX, TextPosY(labelRect, labelStyle), labelStyle)
		}
	}
}

// InteractionOverlayActive implements InteractionOverlayPainter.
func (nr *NavigationRail) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (nr *NavigationRail) DrawInteractionOverlay() {}
