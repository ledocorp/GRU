// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	fabMiniDiameter   = float32(56)
	fabMiniIconSize   = float32(24)
	fabExtendedHeight = float32(48)
	fabMargin         = float32(16)
	fabBottomLift     = float32(8) // extra gap above parent bottom edge
	fabShadowOffsetY  = float32(2)
	fabShadowExpand   = float32(1)
	fabShadowAlpha    = uint8(28)
)

// FAB is a floating action button anchored to the bottom-right of Anchor or its parent.
//
// Leave Label empty for a circular mini FAB; set Label for an extended pill.
// Set Anchor to the content area (e.g. the shell body container) when the FAB is
// a root-level sibling so it is not stretched by flex layout.
//
// # LLM Prompt Template
//
//	fab := ui.NewFAB("compose", "+", "", nil, 0, 0, 0, 0)
//	fab.OnClick = func() { openCompose() }
//	body.AddChild(fab)
//
// Demo scenes: **AppShell Demo**.
type FAB struct {
	Element
	Icon           string
	phosphorName   string
	phosphorWeight IconWeight
	Label          string
	OnClick        func()
	Anchor         Node    // optional layout reference; defaults to parent
	Margin         float32 // inset from anchor bottom-right; 0 uses fabMargin
	hovered        bool
	pressed        bool
	Scale          float32
}

// SetIcon draws the FAB glyph from the shared Remix icon registry.
func (f *FAB) SetIcon(name string, weight IconWeight) {
	f.phosphorName = name
	f.phosphorWeight = weight
	if f.Icon == "" && name == IconPlus {
		f.Icon = "+"
	}
	f.MarkDrawDirty()
}

// Deprecated: use SetIcon.
func (f *FAB) SetPhosphorIcon(name string, weight IconWeight) {
	f.SetIcon(name, weight)
}

// NewFAB creates a FAB. Pass w=0, h=0; size is computed in Layout.
func NewFAB(id, icon, label string, onClick func(), x, y, w, h float32) *FAB {
	dw, dh := fabMiniDiameter, fabMiniDiameter
	if label != "" {
		style := Style{
			TextColor: rl.White,
			FontSize:  14,
			Bold:      true,
			Padding:   16,
		}
		labelW := float32(measureTextS(label, style))
		iconW := float32(measureTextS(icon, style))
		pad := float32(20)
		if style.Padding > 0 {
			pad = style.Padding
		}
		dw = pad*2 + iconW + labelW + 12
		if dw < fabMiniDiameter {
			dw = fabMiniDiameter
		}
		dh = fabExtendedHeight
	}
	f := &FAB{
		Element: NewElement(id, x, y, dw, dh),
		Icon:    icon,
		Label:   label,
		OnClick: onClick,
		Scale:   1.0,
	}
	f.styleName = "fab"
	return f
}

// IsInteractive implements Node.
func (f *FAB) IsInteractive() bool { return true }

func (f *FAB) margin() float32 {
	if f.Margin > 0 {
		return f.Margin
	}
	return fabMargin
}

func (f *FAB) desiredSize() (w, h float32) {
	if f.Label == "" {
		return fabMiniDiameter, fabMiniDiameter
	}
	style := f.GetStyle()
	labelW := float32(measureTextS(f.Label, style))
	iconW := float32(measureTextS(f.Icon, style))
	pad := float32(20)
	if style.Padding > 0 {
		pad = style.Padding
	}
	w = pad*2 + iconW + labelW + 12
	if w < fabMiniDiameter {
		w = fabMiniDiameter
	}
	return w, fabExtendedHeight
}

// Layout pins the FAB to the anchor's bottom-right corner.
func (f *FAB) Layout() {
	defer func() { f.layoutDirty = false }()
	ref := f.Anchor
	if ref == nil {
		ref = f.ParentNode()
	}
	if ref == nil {
		return
	}
	pb := ref.Bounds()
	w, h := f.desiredSize()
	m := f.margin()
	x := pb.X + pb.Width - m - w
	y := pb.Y + pb.Height - m - fabBottomLift - h
	f.setBoundsNoMark(rl.NewRectangle(x, y, w, h))
}

// Update handles click and hover.
func (f *FAB) Update(_ float32) {
	if f.IsHidden() {
		return
	}
	// Anchor position must be current before hit-testing (Root is LayoutAbsolute).
	f.Layout()
	prevHovered := f.hovered
	prevPressed := f.pressed
	mouse := rl.GetMousePosition()
	b := f.Bounds()
	f.hovered = rl.CheckCollisionPointRec(mouse, b)
	f.pressed = f.hovered && rl.IsMouseButtonDown(rl.MouseLeftButton)
	if f.hovered != prevHovered || f.pressed != prevPressed {
		f.MarkDrawDirty()
	}
	if f.hovered && PointerClickConsume(b) {
		if f.OnClick != nil {
			f.OnClick()
		}
	}
}

// Draw implements Node.Draw.
func (f *FAB) Draw() { f.drawInternal() }

func (f *FAB) drawInternal() {
	if f.IsHidden() {
		return
	}
	b := f.Bounds()
	style := f.GetStyle()
	bg, _ := buttonInteractionColors(style, f.hovered, f.pressed)

	scaledW := b.Width * f.Scale
	scaledH := b.Height * f.Scale
	offsetX := (scaledW - b.Width) / 2
	offsetY := (scaledH - b.Height) / 2
	rect := rl.NewRectangle(b.X-offsetX, b.Y-offsetY, scaledW, scaledH)

	if f.Label == "" {
		cx := rect.X + rect.Width/2
		cy := rect.Y + rect.Height/2
		r := rect.Width / 2
		if rect.Height < rect.Width {
			r = rect.Height / 2
		}
		drawFABShadow(cx, cy, r)
		rl.DrawCircleV(rl.NewVector2(cx, cy), r, bg)
		iconStyle := style
		iconSize := fabMiniIconSize
		if iconSize > rect.Width-12 {
			iconSize = rect.Width - 12
		}
		iconDst := snapPhosphorRect(rl.NewRectangle(cx-iconSize/2, cy-iconSize/2, iconSize, iconSize))
		drawn := false
		if f.phosphorName != "" {
			wt := f.phosphorWeight
			if wt == "" {
				wt = PhosphorRegular
			}
			drawn = Phosphor.Draw(iconDst, f.phosphorName, wt, iconStyle.TextColor)
		}
		if !drawn && f.Icon != "" {
			iconStyle.FontSize = int32(iconSize * 0.85)
			if iconStyle.FontSize < 14 {
				iconStyle.FontSize = 14
			}
			iw := measureTextS(f.Icon, iconStyle)
			drawTextS(f.Icon, int32(cx-float32(iw)/2), int32(cy-float32(iconStyle.FontSize)/2), iconStyle)
		}
		return
	}

	radius := style.CornerRadius
	if radius <= 0 {
		radius = fabExtendedHeight / 2
	}
	drawFABShadow(rect.X+rect.Width/2, rect.Y+rect.Height/2, radius)
	drawRoundedControl(rect, scaledW, scaledH, radius, bg)
	iconStyle := style
	iconW := measureTextS(f.Icon, iconStyle)
	tx := int32(rect.X + 16)
	ty := TextPosY(rect, iconStyle)
	drawTextS(f.Icon, tx, ty, iconStyle)
	labelX := int32(rect.X + 16 + float32(iconW) + 8)
	drawTextS(f.Label, labelX, ty, iconStyle)
}

func drawFABShadow(cx, cy, r float32) {
	shadow := rl.NewColor(0, 0, 0, fabShadowAlpha)
	rl.DrawCircleV(rl.NewVector2(cx, cy+fabShadowOffsetY), r+fabShadowExpand, shadow)
}
