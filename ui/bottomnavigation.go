// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	bottomNavDefaultH    = float32(64)
	bottomNavCompactH    = float32(60) // compact strip (~25% taller than 48px)
	bottomNavIconSize    = float32(28)
	bottomNavCompactIcon = float32(26)
	bottomNavIconTop     = float32(8)
	bottomNavLabelGap    = float32(4)
	bottomNavBadgeH      = float32(16)
)

// BottomNavCompactHeight is the intrinsic bar height when Compact is true.
func BottomNavCompactHeight() float32 { return bottomNavCompactH }

// BottomNavItem is one destination in a BottomNavigationBar.
type BottomNavItem struct {
	Icon           string         // UTF-8 fallback when Phosphor is empty
	Phosphor       string         // Phosphor icon name (e.g. ui.PhosphorHouse)
	PhosphorWeight PhosphorWeight // default PhosphorRegular
	Label          string
	Badge          string // optional count or dot text; empty hides the badge
}

// BottomNavigationBar is a Material-style bottom bar with icon + label items.
//
// Pass w=0, h=0 for flex width and intrinsic height. Selection is stored in
// Selected; optional OnChange runs after a tap changes the index.
//
// Example:
//
//	tab := ui.NewSignal(0)
//	nav := ui.NewBottomNavigationBar("nav", []ui.BottomNavItem{
//	    {Icon: "⌂", Label: "Home"},
//	    {Icon: "🔍", Label: "Search"},
//	}, tab, 0, 0, 0, 0)
type BottomNavigationBar struct {
	Element
	Items    []BottomNavItem
	Selected *Signal[int]
	OnChange func(index int)
	// Compact uses a shorter bar (~25%) and omits labels (icon-only strip).
	Compact bool
	// MenuMode fires OnChange on every tap, even when re-selecting the same item.
	MenuMode bool
	hoverIdx int
}

// NewBottomNavigationBar creates a bottom navigation bar.
func NewBottomNavigationBar(id string, items []BottomNavItem, selected *Signal[int], x, y, w, h float32) *BottomNavigationBar {
	if selected == nil {
		selected = NewSignal(0)
	}
	if len(items) == 0 {
		items = []BottomNavItem{{Icon: "·", Label: "-"}}
	}
	if selected.Get() < 0 || selected.Get() >= len(items) {
		selected.Set(0)
	}
	bn := &BottomNavigationBar{
		Element:  NewElement(id, x, y, w, h),
		Items:    items,
		Selected: selected,
		hoverIdx: -1,
	}
	bn.styleName = "bottomnav"
	if h == 0 {
		bn.AutoHeight = true
	}
	bn.Selected.Subscribe(func() { bn.MarkDirty() })
	return bn
}

// IsInteractive implements Node.
func (bn *BottomNavigationBar) IsInteractive() bool { return true }

func (bn *BottomNavigationBar) itemBounds(i int) rl.Rectangle {
	b := bn.Bounds()
	n := len(bn.Items)
	if n <= 0 {
		return b
	}
	w := b.Width / float32(n)
	return rl.NewRectangle(b.X+float32(i)*w, b.Y, w, b.Height)
}

// ItemMenuAnchor returns a screen point above item i suitable for ShowContextMenu.
func (bn *BottomNavigationBar) ItemMenuAnchor(i int) (x, y float32) {
	seg := bn.itemBounds(i)
	return seg.X + seg.Width/2, seg.Y
}

// Update handles item selection.
func (bn *BottomNavigationBar) Update(_ float32) {
	if bn.IsHidden() {
		return
	}
	if ScenePointerBlocked() {
		bn.hoverIdx = -1
		return
	}
	mouse := rl.GetMousePosition()
	bn.hoverIdx = -1
	for i := range bn.Items {
		if rl.CheckCollisionPointRec(mouse, bn.itemBounds(i)) {
			bn.hoverIdx = i
			break
		}
	}
	if bn.hoverIdx < 0 || !PointerClickConsume(bn.itemBounds(bn.hoverIdx)) {
		return
	}
	prev := bn.Selected.Get()
	if prev != bn.hoverIdx {
		bn.Selected.Set(bn.hoverIdx)
	}
	if bn.OnChange != nil && (bn.MenuMode || prev != bn.hoverIdx) {
		bn.OnChange(bn.hoverIdx)
	}
}

func (bn *BottomNavigationBar) barHeight() float32 {
	if bn.Compact {
		return bottomNavCompactH
	}
	return bottomNavDefaultH
}

func (bn *BottomNavigationBar) iconSize() float32 {
	if bn.Compact {
		return bottomNavCompactIcon
	}
	return bottomNavIconSize
}

func (bn *BottomNavigationBar) iconsOnly() bool {
	if bn.Compact {
		return true
	}
	for _, item := range bn.Items {
		if strings.TrimSpace(item.Label) != "" {
			return false
		}
	}
	return true
}

// Layout clamps to intrinsic height when AutoHeight is set.
func (bn *BottomNavigationBar) Layout() {
	if !bn.IsAutoHeight() {
		bn.layoutDirty = false
		return
	}
	want := bn.barHeight()
	b := bn.Bounds()
	if b.Height < want-0.5 || b.Height > want+0.5 {
		b.Height = want
		bn.setBoundsNoMark(b)
	}
	bn.layoutDirty = false
}

// Draw implements Node.Draw.
func (bn *BottomNavigationBar) Draw() {
	defer func() { bn.drawDirty = false }()
	bn.drawInternal()
}

func (bn *BottomNavigationBar) drawInternal() {
	if bn.IsHidden() {
		return
	}
	b := bn.Bounds()
	baseStyle := bn.GetStyle()
	if baseStyle.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, baseStyle.BackgroundColor)
	}
	if baseStyle.BorderWidth > 0 && baseStyle.BorderColor.A > 0 {
		rl.DrawLineEx(
			rl.NewVector2(b.X, b.Y),
			rl.NewVector2(b.X+b.Width, b.Y),
			baseStyle.BorderWidth, baseStyle.BorderColor)
	}

	sel := bn.Selected.Get()
	if sel < 0 {
		sel = 0
	}
	if sel >= len(bn.Items) {
		sel = len(bn.Items) - 1
	}

	for i, item := range bn.Items {
		seg := bn.itemBounds(i)
		active := i == sel
		if active {
			accent := rl.NewColor(79, 70, 229, 255)
			rl.DrawLineEx(
				rl.NewVector2(seg.X+seg.Width*0.2, seg.Y+2),
				rl.NewVector2(seg.X+seg.Width*0.8, seg.Y+2),
				2, accent)
		}

		iconColor := baseStyle.TextColor
		if active {
			iconColor = rl.NewColor(79, 70, 229, 255)
		} else if i == bn.hoverIdx {
			iconColor = rl.NewColor(55, 58, 78, 255)
		}
		iconSz := bn.iconSize()
		var iconY float32
		if bn.iconsOnly() {
			iconY = seg.Y + (seg.Height-iconSz)/2
		} else {
			iconY = seg.Y + bottomNavIconTop
		}
		iconDst := snapPhosphorRect(rl.NewRectangle(
			seg.X+(seg.Width-iconSz)/2,
			iconY,
			iconSz,
			iconSz,
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
				fallback.FontSize = int32(bottomNavIconSize)
				iw := measureTextS(item.Icon, fallback)
				drawTextS(item.Icon, int32(seg.X+(seg.Width-float32(iw))/2), int32(iconY), fallback)
			}
		} else if item.Icon != "" {
			fallback := baseStyle
			fallback.TextColor = iconColor
			fallback.FontSize = int32(iconSz)
			iw := measureTextS(item.Icon, fallback)
			drawTextS(item.Icon, int32(seg.X+(seg.Width-float32(iw))/2), int32(iconY), fallback)
		}

		if !bn.iconsOnly() && strings.TrimSpace(item.Label) != "" {
			labelStyle := baseStyle
			if active {
				labelStyle.TextColor = rl.NewColor(79, 70, 229, 255)
				labelStyle.Bold = true
			}
			labelFS := EffectiveFontSize(labelStyle)
			labelBandY := iconY + iconSz + bottomNavLabelGap
			labelBandH := seg.Y + seg.Height - labelBandY - 6
			if labelBandH < labelFS {
				labelBandH = labelFS
			}
			labelRect := rl.NewRectangle(seg.X, labelBandY, seg.Width, labelBandH)
			lw := measureTextS(item.Label, labelStyle)
			labelX := int32(seg.X + (seg.Width-float32(lw))/2)
			drawTextS(item.Label, labelX, TextPosY(labelRect, labelStyle), labelStyle)
		}

		if item.Badge != "" {
			badgeStyle := Style{TextColor: rl.White, FontSize: 10, Bold: true}
			bw := measureTextS(item.Badge, badgeStyle) + 10
			if bw < 18 {
				bw = 18
			}
			bx := seg.X + seg.Width/2 + 8
			by := iconY - 2
			rl.DrawRectangleRounded(
				rl.NewRectangle(bx, by, float32(bw), bottomNavBadgeH),
				0.5, 4, rl.NewColor(239, 68, 68, 255))
			drawTextS(item.Badge, int32(bx+5), int32(by+2), badgeStyle)
		}
	}
}

// InteractionOverlayActive implements InteractionOverlayPainter.
// Hover is handled via MarkDrawDirty so icons stay in the SSAA cache (sharp).
func (bn *BottomNavigationBar) InteractionOverlayActive() bool {
	return false
}

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (bn *BottomNavigationBar) DrawInteractionOverlay() {}
