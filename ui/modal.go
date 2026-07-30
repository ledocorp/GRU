// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── Modal ───────────────────────────────────────────────────────────────────
//
// Modal provides a package-level manager for blocking dialog overlays.
// Animation, scrim, backdrop, Escape, and focus tracking use OverlayHost (C5).
//
// # Basic usage
//
//	ui.ShowModal("Confirm", bodyNode, []ui.ModalButton{
//	    {Label: "OK",     Action: func() { doIt(); ui.CloseModal() }},
//	    {Label: "Cancel", Action: ui.CloseModal},
//	})
//
// # Appearance
//
// The box is centred on screen and fades in with a gentle scale-up (0.94 → 1.0
// over 150 ms). Closing is animated with a matching fade-out (120 ms).
//
// A multi-layer drop shadow, rounded corners, a title bar with an × button,
// a content area (any Node), and a right-aligned footer button row are drawn
// automatically.
//
// Button styles default to "primary" for button[0] and "button" for the rest.
// Set ModalButton.Style to a theme key to override per-button.
//
// # Configurable size
//
// Set ModalMgr.Width and ModalMgr.Height before ShowModal (or per-call via
// ShowModalSized). Zero values use the built-in defaults (500 × 340 px).
//
// # Main-loop integration
//
//	ui.ModalMgr.Update(dt)   // after doc.Root.Update
//	ui.ModalMgr.Draw()       // inside BeginDrawing, after all widget draws
//
// # LLM Prompt Template
//
//	ui.ShowModal("Confirm", bodyNode, []ui.ModalButton{
//	    {Label: "OK", Action: func() { doIt(); ui.CloseModal() }},
//	    {Label: "Cancel", Action: ui.CloseModal},
//	})
//	// main loop: ui.ModalMgr.Update(dt); ui.ModalMgr.Draw()
//
// Demo scenes: **Widgets Demo**, **Batch 1**, **Notepad** (Find/Replace/Go To).
//
// Styled via the "modal" and "modal-title" theme keys.

// ModalButton defines an action button in the modal footer.
type ModalButton struct {
	Label  string // displayed text
	Style  string // theme key ("primary", "button", …); "" = auto-assign
	Action func() // called on click
}

const (
	modalTitleBarH       = float32(48)
	modalFooterH         = float32(64)
	modalFooterBtnH      = float32(34)
	modalBtnW            = float32(110)
	modalBtnGap          = float32(10)
	modalRoundSegments   = int32(32)
)

// modalManager is the singleton implementation.
type modalManager struct {
	host *OverlayHost

	title      string
	content    Node
	buttons    []ModalButton
	hoveredBtn int  // -1 when none hovered
	xHovered   bool // × button in title bar

	// Public options — set before calling ShowModal.
	CloseOnBackdrop bool    // close when clicking outside the box (default: true)
	CloseOnEscape   bool    // close on Escape key (default: true)
	Width           float32 // box width; 0 = default (500)
	Height          float32 // box height; 0 = default (340)

}

// ModalMgr is the package-level singleton. Drive it from the main loop.
var ModalMgr = &modalManager{
	host:            DefaultOverlayHost(OverlayAnimFadeScale),
	hoveredBtn:      -1,
	CloseOnBackdrop: true,
	CloseOnEscape:   true,
}

func (m *modalManager) syncHostOptions() {
	m.host.CloseOnBackdrop = m.CloseOnBackdrop
	m.host.CloseOnEscape = m.CloseOnEscape
}

// ShowModal opens a modal dialog. If a modal is already open it is replaced.
func ShowModal(title string, content Node, buttons []ModalButton) {
	ModalMgr.title = title
	ModalMgr.content = content
	ModalMgr.buttons = buttons
	ModalMgr.hoveredBtn = -1
	ModalMgr.xHovered = false
	ModalMgr.syncHostOptions()
	ModalMgr.host.BeginOpen()
}

// ShowModalSized is like ShowModal but overrides the box dimensions for this call.
// Pass 0 to retain the ModalMgr.Width / ModalMgr.Height defaults.
func ShowModalSized(title string, content Node, buttons []ModalButton, w, h float32) {
	if w > 0 {
		ModalMgr.Width = w
	}
	if h > 0 {
		ModalMgr.Height = h
	}
	ShowModal(title, content, buttons)
}

// CloseModal begins the close animation. IsModalOpen returns false immediately.
func CloseModal() {
	ModalMgr.host.BeginClose()
}

// IsModalOpen reports whether a modal is logically open (before CloseModal is called).
// Returns false as soon as CloseModal() is called, even during the fade-out.
func IsModalOpen() bool { return ModalMgr.host.IsOpen() }

// IsModalVisible reports whether the modal is currently being rendered,
// including its close animation.
func IsModalVisible() bool { return ModalMgr.host.Open }

// IsAnimating reports fade-in, fade-out, or open settling frames.
func (m *modalManager) IsAnimating() bool { return m.host.IsAnimating() }

// DebugInfo returns a single-line status string used by the Inspector header.
func (m *modalManager) DebugInfo() string {
	if !m.host.Open {
		return "modal: closed"
	}
	state := "open"
	if m.host.Closing {
		state = "closing"
	}
	return fmt.Sprintf("modal: %s  %q  btns:%d  α:%.2f", state, m.title, len(m.buttons), m.host.Alpha)
}

// ─── Internal layout ──────────────────────────────────────────────────────────

// layoutBox returns the unscaled position and size of the modal box.
func (m *modalManager) layoutBox() rl.Rectangle {
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	boxW := m.Width
	if boxW <= 0 {
		boxW = 500
	}
	boxH := m.Height
	if boxH <= 0 {
		boxH = 340
	}
	if boxW > sw-40 {
		boxW = sw - 40
	}
	if boxH > sh-40 {
		boxH = sh - 40
	}
	boxX := (sw - boxW) / 2
	boxY := (sh - boxH) / 2
	return rl.NewRectangle(boxX, boxY, boxW, boxH)
}

// scaledBox applies the pop-in scale transform to the layout box.
func (m *modalManager) scaledBox() rl.Rectangle {
	return ScaledCenterBox(m.layoutBox(), m.host.Scale)
}

// panelHitRect is the screen-space modal shell used for overlay pointer blocking.
func (m *modalManager) panelHitRect() rl.Rectangle {
	return snapControlRect(m.scaledBox())
}

func modalFooterBtnY(boxBottom float32) float32 {
	footerTop := boxBottom - modalFooterH
	return footerTop + (modalFooterH-modalFooterBtnH)/2
}

// ─── Update ──────────────────────────────────────────────────────────────────

// Update advances animations and processes input. Call once per frame after
// doc.Root.Update. When a modal is open, mouse clicks outside the box are
// consumed so the scene behind cannot receive them.
func (m *modalManager) Update(dt float32) {
	if !m.host.Open {
		return
	}

	m.syncHostOptions()
	m.host.AdvanceAnimation(dt)
	if !m.host.Open {
		return
	}

	if !m.host.InputReady() {
		m.host.TickSkipFrames()
		if m.content != nil {
			layoutOverlaySubtree(m.content, m.contentRect())
			m.content.Update(dt)
		}
		return
	}

	if m.host.HandleEscape(CloseModal) {
		return
	}

	box := m.scaledBox()
	mouse := rl.GetMousePosition()
	inside := rl.CheckCollisionPointRec(mouse, box)

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) || rl.IsMouseButtonPressed(rl.MouseRightButton) {
		if !inside {
			PointerClickMarkUsed()
		}
	}

	if m.host.HandleBackdrop(box, box, CloseModal) {
		return
	}

	// × close button (top-right of title bar).
	xRect := rl.NewRectangle(box.X+box.Width-modalTitleBarH, box.Y, modalTitleBarH, modalTitleBarH)
	m.xHovered = rl.CheckCollisionPointRec(mouse, xRect)
	if m.xHovered && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		CloseModal()
		return
	}

	// Footer buttons.
	m.hoveredBtn = -1
	if len(m.buttons) > 0 {
		style := GetThemeStyle("modal")
		total := float32(len(m.buttons))
		startX := box.X + box.Width - style.Padding - (modalBtnW+modalBtnGap)*total + modalBtnGap
		for i, btn := range m.buttons {
			bx := startX + float32(i)*(modalBtnW+modalBtnGap)
			br := rl.NewRectangle(bx, modalFooterBtnY(box.Y+box.Height), modalBtnW, modalFooterBtnH)
			if rl.CheckCollisionPointRec(mouse, br) {
				m.hoveredBtn = i
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) && btn.Action != nil {
					btn.Action()
					return
				}
			}
		}
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && inside && m.content != nil {
		cr := m.contentRect()
		if rl.CheckCollisionPointRec(mouse, cr) {
			m.host.HandleFocusClick(m.content, cr)
		} else if !m.xHovered && m.hoveredBtn < 0 {
			if d := ActiveDocument(); d != nil {
				d.SetFocus(nil)
				m.host.FocusedWidget = nil
			}
		}
	}

	if inside && (rl.IsMouseButtonPressed(rl.MouseLeftButton) || rl.IsMouseButtonPressed(rl.MouseRightButton)) {
		PointerClickMarkUsed()
	}

	if m.content != nil {
		if d := ActiveDocument(); d != nil {
			ensureModalDocumentFocus(d)
		}
		layoutOverlaySubtree(m.content, m.contentRect())
		m.content.Update(dt)
	}
}

// EnsureModalDocumentFocus keeps document focus on the modal's default field.
// Called automatically from ModalMgr.Update; scenes may also call after ShowModal.
func EnsureModalDocumentFocus(d *Document) {
	ensureModalDocumentFocus(d)
}

func ensureModalDocumentFocus(d *Document) {
	if !IsModalOpen() || d == nil || ModalMgr.content == nil {
		return
	}
	if d.Focused != nil && subtreeContainsNode(ModalMgr.content, d.Focused) {
		return
	}
	// User clicked empty modal chrome — keep focus cleared.
	if d.Focused == nil {
		return
	}
	target := findModalDefaultFocus(ModalMgr.content)
	if target == nil {
		return
	}
	d.SetFocus(target)
	ModalMgr.host.FocusedWidget = target
}

func subtreeContainsNode(root, needle Node) bool {
	if root == nil || needle == nil {
		return false
	}
	if root == needle {
		return true
	}
	for _, ch := range root.Children() {
		if subtreeContainsNode(ch, needle) {
			return true
		}
	}
	return false
}

func findModalDefaultFocus(root Node) Node {
	var textInput Node
	var any Node
	var walk func(Node)
	walk = func(n Node) {
		if n == nil || n.IsHidden() {
			return
		}
		if n.IsInteractive() {
			if any == nil {
				any = n
			}
			if _, ok := n.(*TextInput); ok && textInput == nil {
				textInput = n
			}
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(root)
	if textInput != nil {
		return textInput
	}
	return any
}

// contentRect is the layout/draw band for modal body widgets (matches widgets_demo modal).
func (m *modalManager) contentRect() rl.Rectangle {
	box := snapControlRect(m.scaledBox())
	style := GetThemeStyle("modal")
	pad := style.Padding
	cY := box.Y + modalTitleBarH + pad
	cH := box.Height - modalTitleBarH - pad*2 - modalFooterH
	if cH < 1 {
		cH = 1
	}
	w := box.Width - pad*2
	if w < 1 {
		w = 1
	}
	return rl.NewRectangle(box.X+pad, cY, w, cH)
}

// ─── Draw ────────────────────────────────────────────────────────────────────

// Draw renders the backdrop, modal box, title bar, content, and footer buttons.
// Call inside BeginDrawing / EndDrawing, after all regular widget draws.
func (m *modalManager) Draw() {
	if !m.host.Open || m.host.Alpha <= 0 {
		return
	}

	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	m.host.DrawScrimFull(sw, sh)

	sBox := snapControlRect(m.scaledBox())
	a := m.host.Alpha
	sX, sY, sW, sH := sBox.X, sBox.Y, sBox.Width, sBox.Height

	style := GetThemeStyle("modal")
	bw := style.BorderWidth
	if bw <= 0 {
		bw = 2
	}
	roundness := buttonCornerRoundness(sW, sH, style.CornerRadius)
	if roundness < 0.08 {
		roundness = 0.08
	}
	titleStyle := GetThemeStyle("modal-title")

	// ── Drop shadow (context-menu style, uniform corners on shell) ───────────
	for _, sh := range []struct {
		off float32
		a   uint8
	}{{6, 12}, {4, 22}, {2, 34}} {
		sha := uint8(float32(sh.a) * a)
		sr := snapControlRect(rl.NewRectangle(sBox.X+sh.off, sBox.Y+sh.off, sW, sH))
		rl.DrawRectangleRounded(sr, roundness, modalRoundSegments, rl.NewColor(0, 0, 0, sha))
	}

	bg := style.BackgroundColor
	bg.A = uint8(float32(bg.A) * a)
	bc := style.BorderColor
	bc.A = uint8(float32(bc.A) * a)
	drawRoundedInsetBorder(sBox, roundness, bw, bc, bg)

	// ── Title (text + divider; shell keeps rounded top corners) ───────────────
	tbA := uint8(255 * a)
	divider := style.BorderColor
	if divider.A == 0 {
		divider = rl.NewColor(210, 213, 228, 255)
	}
	divider.A = tbA
	rl.DrawRectangle(int32(sX), int32(sY+modalTitleBarH), int32(sW), 1, divider)

	ts := titleStyle
	tc := ts.TextColor
	tc.A = tbA
	ts.TextColor = tc
	titleRect := rl.NewRectangle(sX, sY, sW, modalTitleBarH)
	drawTextS(m.title, int32(sX)+16, TextPosY(titleRect, ts), ts)

	// ── Close (Remix ri-close-line) ───────────────────────────────────────────
	const xHit = float32(modalTitleBarH)
	xRect := rl.NewRectangle(sX+sW-xHit, sY, xHit, xHit)
	xClr := rl.NewColor(120, 124, 142, tbA)
	if m.xHovered {
		xClr = rl.NewColor(45, 48, 62, tbA)
		drawTitleBarButtonHover(xRect, rl.NewColor(228, 230, 242, tbA))
	}
	drawTitleBarChromeIcon(xRect, RemixClose, xClr, 1.15, titleBarIconStrokeThin)

	// ── Content area (hidden during close animation to avoid ghost text) ──────
	if m.content != nil && !m.host.Closing {
		drawOverlaySubtreeClipped(m.content, m.contentRect(), true)
	}

	// ── Footer ────────────────────────────────────────────────────────────────
	fY := sY + sH - modalFooterH
	footerDivider := style.BorderColor
	if footerDivider.A == 0 {
		footerDivider = rl.NewColor(210, 213, 228, 255)
	}
	footerDivider.A = tbA
	rl.DrawRectangle(int32(sX), int32(fY), int32(sW), 1, footerDivider)

	if len(m.buttons) > 0 {
		total := float32(len(m.buttons))
		startX := sX + sW - style.Padding - (modalBtnW+modalBtnGap)*total + modalBtnGap
		for i, btn := range m.buttons {
			bx := startX + float32(i)*(modalBtnW+modalBtnGap)
			by := modalFooterBtnY(sY + sH)
			br := rl.NewRectangle(bx, by, modalBtnW, modalFooterBtnH)

			key := btn.Style
			if key == "" {
				if i == 0 {
					key = "primary"
				} else {
					key = "button"
				}
			}
			bs := GetThemeStyle(key)

			bgC := bs.BackgroundColor
			if m.hoveredBtn == i {
				bgC = rl.ColorBrightness(bgC, 0.15)
			}
			bgC.A = uint8(float32(bgC.A) * a)
			rl.DrawRectangleRounded(br, 0.3, 6, bgC)

			bbc := bs.BorderColor
			bbc.A = uint8(float32(bbc.A) * a)
			rl.DrawRectangleRoundedLinesEx(br, 0.3, 6, 1, bbc)

			ls := bs
			ls.FontSize = 18
			ls.FontDensity = 1.05
			ls.MinFontSize = 18
			lw := measureTextS(btn.Label, ls)
			ltc := ls.TextColor
			ltc.A = uint8(float32(ltc.A) * a)
			ls.TextColor = ltc
			drawTextS(btn.Label, int32(bx)+int32(modalBtnW/2)-lw/2, TextPosY(br, ls), ls)
		}
	}
}
