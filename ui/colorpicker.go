// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ── Popup geometry constants ───────────────────────────────────────────────────

const (
	cpPopW = float32(300) // popup width
	cpPad  = float32(14)  // popup inner padding
	cpSvH  = float32(150) // saturation/value box height
	cpBarH = float32(20)  // hue / alpha bar height
	cpGap  = float32(10)  // vertical gap between sections
	cpPrvH = float32(36)  // preview row height (old + new swatches + hex)

	cpSwatchW = float32(52) // Bootstrap-ish input height row (fixed swatch width)
	cpSwatchH = float32(36)

	// Computed total popup heights (top-pad + sections + gaps + bottom-pad).
	cpPopHNoAlpha = cpPad + cpSvH + cpGap + cpBarH + cpGap + cpPrvH + cpPad                  // ≈ 254
	cpPopHAlpha   = cpPad + cpSvH + cpGap + cpBarH + cpGap + cpBarH + cpGap + cpPrvH + cpPad // ≈ 284
)

// ── ColorPicker widget (swatch in the widget tree) ────────────────────────────

// ColorPicker renders a rounded color swatch that opens an HSV popup picker
// when clicked.  The selected color is exposed via the reactive [Value] signal.
//
// Layout: embed in any FlexColumn or FlexRow container.  Pass w=0 for the default
// compact swatch (52×36 px); set explicit w/h for custom sizes.
//
// Popup: click the swatch to open a 300×254 (or 300×284 with alpha) overlay
// showing an SV box, hue bar, optional alpha slider, and a preview row with
// old/new swatches and a hex string.  Close with Escape or click outside.
//
// # LLM Prompt Template
//
//	cp := ui.NewColorPicker("accent", rl.Blue, 0, 0, 52, 36)
//	cp.Value.Subscribe(func() { applyAccent(cp.Value.Get()) })
//	form.AddChild(cp)
//
// Demo scenes: **Batch 26 ColorPicker**.
type ColorPicker struct {
	Element

	// Value is the reactive currently-selected color.
	// Read, subscribe to, or set it programmatically at any time.
	Value *Signal[rl.Color]

	// ShowAlpha, when true, adds an alpha slider to the popup.  Default: false.
	ShowAlpha bool

	hovered bool
}

// NewColorPicker creates a ColorPicker swatch widget.
//
//	id      – unique node ID
//	initial – starting color
//	x, y    – position (ignored when managed by a layout container)
//	w, h    – size; pass 0 for default compact swatch (52×36)
func NewColorPicker(id string, initial rl.Color, x, y, w, h float32) *ColorPicker {
	if h <= 0 {
		h = cpSwatchH
	}
	if w <= 0 {
		w = cpSwatchW
	}
	cp := &ColorPicker{
		Element: NewElement(id, x, y, w, h),
		Value:   NewSignal(initial),
	}
	cp.PreferredWidth = w
	cp.styleName = "colorpicker"
	return cp
}

// GetPreferredWidth keeps swatches compact in flex layouts (Bootstrap input-group sizing).
func (cp *ColorPicker) GetPreferredWidth() float32 {
	if cp.PreferredWidth > 0 {
		return cp.PreferredWidth
	}
	if b := cp.Bounds(); b.Width > 0 {
		return b.Width
	}
	return cpSwatchW
}

// IsInteractive satisfies Node so the inspector's FindInteractiveAt resolves
// this widget on click.
func (cp *ColorPicker) IsInteractive() bool { return true }

// Update detects hover/click and delegates to [ColorPickerMgr] on press.
func (cp *ColorPicker) Update(_ float32) {
	if cp.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	prev := cp.hovered
	cp.hovered = rl.CheckCollisionPointRec(mouse, cp.bounds)
	if cp.hovered != prev {
		cp.MarkDrawDirty()
	}
	if cp.hovered && PointerClickConsume(cp.bounds) {
		// If another picker's popup is open and active, swallow the click rather
		// than immediately hijacking the manager.  The click-outside detection in
		// ColorPickerMgr.Update will close the other popup; the user can then
		// click again to open this one.  Without this guard, clicks inside an
		// open popup that spatially overlaps this swatch steal the manager state
		// mid-frame (widget Update runs before ColorPickerMgr.Update).
		if ColorPickerMgr.open && !ColorPickerMgr.closing && ColorPickerMgr.target != cp {
			return
		}
		b := cp.bounds
		ColorPickerMgr.Open(cp, b.X, b.Y+b.Height+4)
	}
}

// Layout is a no-op for leaf widgets; bounds are set by the layout engine.
func (cp *ColorPicker) Layout() { cp.layoutDirty = false }

// Draw renders the swatch.  The popup overlay is drawn by ColorPickerMgr.Draw.
func (cp *ColorPicker) Draw() {
	if cp.IsHidden() {
		return
	}
	cp.drawInternal()
	cp.drawDirty = false
}

func (cp *ColorPicker) drawInternal() {
	b := snapControlRect(cp.bounds)
	c := cp.Value.Get()
	style := cp.GetStyle()

	// Checkerboard background for semi-transparent colors.
	if c.A < 255 {
		ck := int32(8)
		bx0, by0 := int32(b.X), int32(b.Y)
		bx1, by1 := int32(b.X+b.Width), int32(b.Y+b.Height)
		for cx := bx0; cx < bx1; cx += ck {
			for cy := by0; cy < by1; cy += ck {
				light := ((cx-bx0)/ck+(cy-by0)/ck)%2 == 0
				col := rl.NewColor(180, 180, 180, 255)
				if light {
					col = rl.NewColor(240, 240, 240, 255)
				}
				w := ck
				if bx1-cx < ck {
					w = bx1 - cx
				}
				h := ck
				if by1-cy < ck {
					h = by1 - cy
				}
				rl.DrawRectangle(cx, cy, w, h, col)
			}
		}
	}

	bw := style.BorderWidth
	if bw < 1 {
		bw = 1
	}
	shorter := b.Width
	if b.Height < shorter {
		shorter = b.Height
	}
	rn := float32(0)
	if style.CornerRadius > 0 && shorter > 0 {
		rn = style.CornerRadius / (shorter / 2)
		if rn > 1 {
			rn = 1
		}
	}

	borderCol := style.BorderColor
	if borderCol.A == 0 {
		borderCol = rl.NewColor(222, 226, 230, 255)
	}
	isOpen := ColorPickerMgr.open && ColorPickerMgr.target == cp
	if cp.hovered || isOpen {
		borderCol = rl.NewColor(79, 70, 229, 255)
	}

	drawRoundedInsetBorder(b, rn, bw, borderCol, c)

	// Chevron hint (bottom-right).
	ax := b.X + b.Width - 10
	ay := b.Y + b.Height - 7
	rl.DrawTriangle(
		rl.NewVector2(ax-4, ay-2),
		rl.NewVector2(ax+4, ay-2),
		rl.NewVector2(ax, ay+2),
		rl.NewColor(255, 255, 255, 200),
	)
}

// ── HSV ↔ RGB helpers (package-level; used by ColorPicker and Inspector) ─────

// colorToHSVA decomposes c into H [0,360), S [0,1], V [0,1], A [0,1].
func colorToHSVA(c rl.Color) (h, s, v, a float32) {
	r := float32(c.R) / 255
	g := float32(c.G) / 255
	b := float32(c.B) / 255
	a = float32(c.A) / 255

	mx := r
	if g > mx {
		mx = g
	}
	if b > mx {
		mx = b
	}
	mn := r
	if g < mn {
		mn = g
	}
	if b < mn {
		mn = b
	}
	v = mx
	if mx == 0 {
		return 0, 0, 0, a
	}
	s = (mx - mn) / mx
	if mx == mn {
		return 0, s, v, a
	}
	delta := mx - mn
	switch mx {
	case r:
		h = (g - b) / delta
		if h < 0 {
			h += 6
		}
	case g:
		h = (b-r)/delta + 2
	default: // b
		h = (r-g)/delta + 4
	}
	h *= 60
	return
}

// hsvToColor converts H [0,360), S [0,1], V [0,1], A [0,1] to rl.Color.
func hsvToColor(h, s, v, a float32) rl.Color {
	if s == 0 {
		c := uint8(v * 255)
		return rl.NewColor(c, c, c, uint8(a*255))
	}
	h = float32(math.Mod(float64(h), 360))
	if h < 0 {
		h += 360
	}
	sector := h / 60
	i := int(sector)
	f := sector - float32(i)

	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	var r, g, b float32
	switch i {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	default:
		r, g, b = v, p, q
	}
	return rl.NewColor(uint8(r*255), uint8(g*255), uint8(b*255), uint8(a*255))
}

// cpClamp01 clamps f to [0, 1].
func cpClamp01(f float32) float32 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// ── colorPickerManager — singleton overlay ─────────────────────────────────────

// colorPickerManager manages the single popup overlay shared by all ColorPicker
// swatches.  Only one popup is open at a time; opening a new one closes any
// previous one.
//
// Wire into the main loop (immediately after doc.Root.Update):
//
//	ui.ColorPickerMgr.Update(dt)
//
// And in the draw section, after all other overlays:
//
//	ui.ColorPickerMgr.Draw()
type colorPickerManager struct {
	open   bool
	target *ColorPicker

	// Editing state in HSV+A.
	h, s, v  float32  // hue [0,360), saturation [0,1], value [0,1]
	aVal     float32  // alpha [0,1]
	original rl.Color // color at popup-open time (shown in "old" swatch)

	// Popup position and size.
	popX, popY float32
	popW, popH float32

	// Drag tracking.
	draggingSV    bool
	draggingHue   bool
	draggingAlpha bool

	// Fade animation [0,1].
	fade    float32
	closing bool // true while fading out

	// skipFrames prevents the same mouse-press that opened the popup from
	// immediately closing it via the click-outside check.
	skipFrames int
}

// ColorPickerMgr is the package-level singleton. Wire it in main.go.
var ColorPickerMgr = &colorPickerManager{}

// IsOpen reports whether any picker popup is currently visible.
func (m *colorPickerManager) IsOpen() bool { return m.open }

// IsAnimating reports fade-in, fade-out, or open settling frames.
func (m *colorPickerManager) IsAnimating() bool {
	if !m.open {
		return false
	}
	return m.closing || m.fade < 1 || m.skipFrames > 0
}

// OpenColorPickerAt opens the HSV popup for cp anchored at screen (x, y).
func OpenColorPickerAt(cp *ColorPicker, anchorX, anchorY float32) {
	if cp == nil {
		return
	}
	ColorPickerMgr.Open(cp, anchorX, anchorY)
}

// Open initialises and shows the popup anchored below (or above) the swatch.
func (m *colorPickerManager) Open(cp *ColorPicker, anchorX, anchorY float32) {
	m.target = cp
	m.h, m.s, m.v, m.aVal = colorToHSVA(cp.Value.Get())
	m.original = cp.Value.Get()

	m.popW = cpPopW
	if cp.ShowAlpha {
		m.popH = cpPopHAlpha
	} else {
		m.popH = cpPopHNoAlpha
	}

	// Clamp to screen; flip above swatch if not enough room below.
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	px := anchorX
	py := anchorY
	if px+m.popW > sw-8 {
		px = sw - m.popW - 8
	}
	if px < 8 {
		px = 8
	}
	if py+m.popH > sh-40 { // -40 reserve for nav bar
		py = anchorY - cp.bounds.Height - 4 - m.popH
	}
	if py < 8 {
		py = 8
	}
	m.popX = px
	m.popY = py

	m.open = true
	m.closing = false
	m.fade = 0
	m.skipFrames = 2
	m.draggingSV = false
	m.draggingHue = false
	m.draggingAlpha = false
}

func (m *colorPickerManager) close() {
	m.closing = true
}

func (m *colorPickerManager) commitColor() {
	if m.target == nil {
		return
	}
	c := hsvToColor(m.h, m.s, m.v, m.aVal)
	m.target.Value.Set(c)
	m.target.MarkDrawDirty()
}

// ── Geometry helpers ──────────────────────────────────────────────────────────

func (m *colorPickerManager) svBoxRect() rl.Rectangle {
	return rl.NewRectangle(m.popX+cpPad, m.popY+cpPad, m.popW-2*cpPad, cpSvH)
}

func (m *colorPickerManager) hueBarRect() rl.Rectangle {
	return rl.NewRectangle(m.popX+cpPad, m.popY+cpPad+cpSvH+cpGap, m.popW-2*cpPad, cpBarH)
}

func (m *colorPickerManager) alphaBarRect() rl.Rectangle {
	base := m.popY + cpPad + cpSvH + cpGap + cpBarH + cpGap
	return rl.NewRectangle(m.popX+cpPad, base, m.popW-2*cpPad, cpBarH)
}

func (m *colorPickerManager) previewRowRect() rl.Rectangle {
	base := m.popY + cpPad + cpSvH + cpGap + cpBarH + cpGap
	if m.target != nil && m.target.ShowAlpha {
		base += cpBarH + cpGap
	}
	return rl.NewRectangle(m.popX+cpPad, base, m.popW-2*cpPad, cpPrvH)
}

// ── Update ────────────────────────────────────────────────────────────────────

// Update handles input and drives the fade animation.
// Call once per frame, after doc.Root.Update, before doc.Root.Layout.
func (m *colorPickerManager) Update(dt float32) {
	if !m.open {
		return
	}

	// Animate fade.
	const fadeIn = float32(0.12)
	const fadeOut = float32(0.10)
	if m.closing {
		m.fade -= dt / fadeOut
		if m.fade <= 0 {
			m.fade = 0
			m.open = false
			m.target = nil
		}
		return // no input while closing
	}
	if m.fade < 1 {
		m.fade += dt / fadeIn
		if m.fade > 1 {
			m.fade = 1
		}
	}

	// Skip first frames to absorb the opening click.
	if m.skipFrames > 0 {
		m.skipFrames--
		return
	}

	mouse := rl.GetMousePosition()
	svRect := m.svBoxRect()
	hueRect := m.hueBarRect()
	popRect := rl.NewRectangle(m.popX, m.popY, m.popW, m.popH)
	showAlpha := m.target != nil && m.target.ShowAlpha
	var alphaRect rl.Rectangle
	if showAlpha {
		alphaRect = m.alphaBarRect()
	}

	// Escape closes.
	if rl.IsKeyPressed(rl.KeyEscape) {
		m.close()
		return
	}

	// On new click: start drag or close if outside popup.
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		switch {
		case rl.CheckCollisionPointRec(mouse, svRect):
			m.draggingSV = true
		case rl.CheckCollisionPointRec(mouse, hueRect):
			m.draggingHue = true
		case showAlpha && rl.CheckCollisionPointRec(mouse, alphaRect):
			m.draggingAlpha = true
		case !rl.CheckCollisionPointRec(mouse, popRect):
			m.close()
			return
		}
	}

	// Release drags on mouse-up.
	if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
		m.draggingSV = false
		m.draggingHue = false
		m.draggingAlpha = false
	}

	// Apply drag updates.
	dirty := false
	if m.draggingSV {
		ns := cpClamp01((mouse.X - svRect.X) / svRect.Width)
		nv := cpClamp01(1 - (mouse.Y-svRect.Y)/svRect.Height)
		if ns != m.s || nv != m.v {
			m.s = ns
			m.v = nv
			dirty = true
		}
	}
	if m.draggingHue {
		nh := cpClamp01((mouse.X-hueRect.X)/hueRect.Width) * 360
		if nh != m.h {
			m.h = nh
			dirty = true
		}
	}
	if m.draggingAlpha {
		na := cpClamp01((mouse.X - alphaRect.X) / alphaRect.Width)
		if na != m.aVal {
			m.aVal = na
			dirty = true
		}
	}
	if dirty {
		m.commitColor()
	}
}

// ── Draw ──────────────────────────────────────────────────────────────────────

// Draw renders the popup overlay.
// Call once per frame, at the very end of the draw pass (after all other
// overlays) so it appears on top.
func (m *colorPickerManager) Draw() {
	if !m.open || m.fade <= 0 {
		return
	}

	// mul scales c.A by the current fade factor.
	mul := func(c rl.Color) rl.Color {
		c.A = uint8(float32(c.A) * m.fade)
		return c
	}
	a := uint8(m.fade * 255)

	// ── Drop shadow ───────────────────────────────────────────────────────────
	shadowRec := rl.NewRectangle(m.popX+3, m.popY+5, m.popW, m.popH)
	rl.DrawRectangleRounded(shadowRec, 0.06, 8, mul(rl.NewColor(0, 0, 20, 55)))

	// ── Popup background + border ─────────────────────────────────────────────
	popRec := rl.NewRectangle(m.popX, m.popY, m.popW, m.popH)
	rl.DrawRectangleRounded(popRec, 0.06, 8, mul(rl.NewColor(250, 250, 253, 255)))
	rl.DrawRectangleRoundedLinesEx(popRec, 0.06, 8, 1, mul(rl.NewColor(210, 212, 228, 255)))

	// ── SV box ────────────────────────────────────────────────────────────────
	m.drawSVBox(a, mul)

	// ── Hue bar ───────────────────────────────────────────────────────────────
	m.drawHueBar(a, mul)

	// ── Alpha bar (optional) ──────────────────────────────────────────────────
	if m.target != nil && m.target.ShowAlpha {
		m.drawAlphaBar(a, mul)
	}

	// ── Preview row ───────────────────────────────────────────────────────────
	m.drawPreviewRow(a, mul)
}

// drawSVBox draws the saturation/value picker row-by-row (opaque, standard HSV read).
func (m *colorPickerManager) drawSVBox(_ uint8, mul func(rl.Color) rl.Color) {
	r := m.svBoxRect()
	rows := int(r.Height)
	if rows < 1 {
		rows = 1
	}
	for yi := 0; yi < rows; yi++ {
		v := float32(1)
		if rows > 1 {
			v = 1 - float32(yi)/float32(rows-1)
		}
		left := mul(hsvToColor(m.h, 0, v, 1))
		right := mul(hsvToColor(m.h, 1, v, 1))
		row := rl.NewRectangle(r.X, r.Y+float32(yi), r.Width, 1)
		DrawHorizontalGradientRect(row, left, right)
	}
	rl.DrawRectangleLinesEx(r, 1, mul(rl.NewColor(200, 202, 218, 255)))

	// Thumb circle at (s, 1-v) within the box.
	tx := r.X + m.s*r.Width
	ty := r.Y + (1-m.v)*r.Height
	rl.DrawCircleV(rl.NewVector2(tx, ty), 6, mul(rl.NewColor(255, 255, 255, 220)))
	rl.DrawCircleLines(int32(tx), int32(ty), 6, mul(rl.NewColor(0, 0, 0, 150)))
}

// drawHueBar draws the rainbow hue strip and its vertical thumb marker.
func (m *colorPickerManager) drawHueBar(a uint8, mul func(rl.Color) rl.Color) {
	r := m.hueBarRect()

	// Six hue stops covering the full 360° wheel.
	stops := [7]rl.Color{
		mul(rl.NewColor(255, 0, 0, 255)),   // 0°   red
		mul(rl.NewColor(255, 255, 0, 255)), // 60°  yellow
		mul(rl.NewColor(0, 255, 0, 255)),   // 120° green
		mul(rl.NewColor(0, 255, 255, 255)), // 180° cyan
		mul(rl.NewColor(0, 0, 255, 255)),   // 240° blue
		mul(rl.NewColor(255, 0, 255, 255)), // 300° magenta
		mul(rl.NewColor(255, 0, 0, 255)),   // 360° red
	}
	for i := 0; i < 6; i++ {
		startX := r.X + float32(i)*r.Width/6
		endX := r.X + float32(i+1)*r.Width/6
		seg := rl.NewRectangle(startX, r.Y, endX-startX+0.5, r.Height)
		DrawCornerGradientRect(seg, stops[i], stops[i], stops[i+1], stops[i+1])
	}
	rl.DrawRectangleLinesEx(r, 1, mul(rl.NewColor(200, 202, 218, 255)))

	// Vertical thumb line.
	tx := r.X + (m.h/360)*r.Width
	rl.DrawRectangle(int32(tx)-1, int32(r.Y)-2, 3, int32(r.Height)+4, rl.NewColor(255, 255, 255, a))
	rl.DrawRectangleLinesEx(rl.NewRectangle(tx-2, r.Y-3, 5, r.Height+6), 1,
		rl.NewColor(0, 0, 0, uint8(float32(a)*0.45)))
}

// drawAlphaBar draws the checkerboard + gradient alpha slider and its thumb.
func (m *colorPickerManager) drawAlphaBar(a uint8, mul func(rl.Color) rl.Color) {
	r := m.alphaBarRect()

	// Checkerboard background.
	ck := int32(8)
	bx0, by0 := int32(r.X), int32(r.Y)
	bx1, by1 := int32(r.X+r.Width), int32(r.Y+r.Height)
	for cx := bx0; cx < bx1; cx += ck {
		for cy := by0; cy < by1; cy += ck {
			light := ((cx-bx0)/ck+(cy-by0)/ck)%2 == 0
			col := rl.NewColor(180, 180, 180, 255)
			if light {
				col = rl.NewColor(240, 240, 240, 255)
			}
			w := ck
			if bx1-cx < ck {
				w = bx1 - cx
			}
			h := ck
			if by1-cy < ck {
				h = by1 - cy
			}
			rl.DrawRectangle(cx, cy, w, h, col)
		}
	}

	// Color gradient: transparent (left) → opaque current color (right).
	cc := hsvToColor(m.h, m.s, m.v, 1)
	transparent := rl.NewColor(cc.R, cc.G, cc.B, 0)
	opaque := rl.NewColor(cc.R, cc.G, cc.B, a)
	DrawHorizontalGradientRect(r, transparent, opaque)
	rl.DrawRectangleLinesEx(r, 1, mul(rl.NewColor(200, 202, 218, 255)))

	// Vertical thumb.
	tx := r.X + m.aVal*r.Width
	rl.DrawRectangle(int32(tx)-1, int32(r.Y)-2, 3, int32(r.Height)+4, rl.NewColor(255, 255, 255, a))
	rl.DrawRectangleLinesEx(rl.NewRectangle(tx-2, r.Y-3, 5, r.Height+6), 1,
		rl.NewColor(0, 0, 0, uint8(float32(a)*0.45)))
}

// drawPreviewRow draws old/new color swatches and the hex string.
func (m *colorPickerManager) drawPreviewRow(_ uint8, mul func(rl.Color) rl.Color) {
	r := m.previewRowRect()
	const sw = float32(44)

	newColor := hsvToColor(m.h, m.s, m.v, m.aVal)

	// Old color swatch.
	oldRec := rl.NewRectangle(r.X, r.Y+(r.Height-sw)*0.5, sw, sw)
	rl.DrawRectangleRounded(oldRec, 0.2, 6, mul(m.original))
	rl.DrawRectangleRoundedLinesEx(oldRec, 0.2, 6, 1, mul(rl.NewColor(0, 0, 0, 50)))

	// ">" arrow.
	arrowX := r.X + sw + 8
	arrowY := r.Y + r.Height/2 - 8
	DrawText(">", arrowX, arrowY, 14, mul(rl.NewColor(140, 144, 168, 255)))

	// New color swatch.
	newRec := rl.NewRectangle(r.X+sw+22, r.Y+(r.Height-sw)*0.5, sw, sw)
	rl.DrawRectangleRounded(newRec, 0.2, 6, mul(newColor))
	rl.DrawRectangleRoundedLinesEx(newRec, 0.2, 6, 1, mul(rl.NewColor(0, 0, 0, 50)))

	// Hex string to the right.
	var hexStr string
	if m.target != nil && m.target.ShowAlpha {
		hexStr = fmt.Sprintf("#%02X%02X%02X%02X",
			newColor.R, newColor.G, newColor.B, uint8(m.aVal*255))
	} else {
		hexStr = fmt.Sprintf("#%02X%02X%02X", newColor.R, newColor.G, newColor.B)
	}
	textX := r.X + sw*2 + 28
	textY := r.Y + r.Height/2 - 8
	DrawText(hexStr, textX, textY, 13, mul(rl.NewColor(40, 44, 70, 255)))

	// Small "old" / "new" labels below each swatch.
	labelY := r.Y + r.Height - 12
	labelCol := mul(rl.NewColor(140, 144, 168, 255))
	DrawText("old", r.X+4, labelY, 10, labelCol)
	DrawText("new", r.X+sw+22+4, labelY, 10, labelCol)
}
