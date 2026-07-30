// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ─── Tooltip ─────────────────────────────────────────────────────────────────
//
// Tooltip is a decorator widget that wraps any Node and shows a hover popup.
//
// # Usage
//
//	btn := ui.NewButton("save", "Save", 0, 0, 140, 40)
//	tip := ui.NewTooltip("save-tip", btn, "Save the current file to disk")
//	panel.AddChild(tip) // add tip in place of btn
//
// Tooltip forwards all Update, Layout, and Draw calls to its Target so the
// underlying widget behaves exactly as if it were added directly to the tree.
// SetStyle on the Tooltip changes the popup's look (not the Target's style).
//
// # Positioning
//
// The popup is placed on whichever of the four sides of the Target has the
// most vertical/horizontal space (bottom preferred, then top, right, left).
// It is always clamped to the screen bounds. Call Tooltips.SetWindowSize(w, h)
// once after rl.InitWindow so the manager knows the screen dimensions.
//
// # Fade-in
//
// The popup fades in over ~150 ms after the hover delay expires. It disappears
// instantly when the mouse leaves the Target or a mouse button is clicked
// outside the Target.
//
// # SetTooltip convenience API
//
// When restructuring the widget tree is not convenient, use the package-level
// SetTooltip to attach a popup to any existing Node without wrapping it:
//
//	ui.SetTooltip(existingButton, "tooltip text")
//
// # Main-loop integration
//
//	// After doc.Root.Update:
//	ui.Tooltips.Update(dt)
//	// Inside BeginDrawing, after all widget draws:
//	ui.Tooltips.Draw()
//
// Styled via the "tooltip" theme key.

// ─── Internal helpers ─────────────────────────────────────────────────────────

// tooltipChooseSide returns (x, y) for a popup of size (popW × popH) near
// anchor. Preference order: bottom, top, right, left. The result is clamped
// to a 4-pixel screen margin.
func tooltipChooseSide(anchor rl.Rectangle, popW, popH float32, winW, winH int32) (x, y float32) {
	const gap = float32(8)
	sw := float32(winW)
	sh := float32(winH)

	// Centre the popup on the perpendicular axis.
	midX := anchor.X + anchor.Width/2 - popW/2
	midY := anchor.Y + anchor.Height/2 - popH/2

	spaceBottom := sh - (anchor.Y + anchor.Height)
	spaceTop := anchor.Y
	spaceRight := sw - (anchor.X + anchor.Width)

	switch {
	case spaceBottom >= popH+gap:
		x, y = midX, anchor.Y+anchor.Height+gap
	case spaceTop >= popH+gap:
		x, y = midX, anchor.Y-popH-gap
	case spaceRight >= popW+gap:
		x, y = anchor.X+anchor.Width+gap, midY
	default: // left side
		x, y = anchor.X-popW-gap, midY
	}

	// Clamp to screen.
	if x < 4 {
		x = 4
	}
	if x+popW > sw-4 {
		x = sw - popW - 4
	}
	if y < 4 {
		y = 4
	}
	if y+popH > sh-4 {
		y = sh - popH - 4
	}
	return
}

// drawTooltipPopup renders a styled popup box near anchor.
// alpha is multiplied into every colour channel to produce the fade effect.
func drawTooltipPopup(text string, anchor rl.Rectangle, alpha float32, winW, winH int32) {
	style := GetThemeStyle("tooltip")
	fs := EffectiveFontSize(style)
	pad := style.Padding
	tw := float32(measureTextS(text, style))
	popW := tw + pad*2
	popH := fs + pad*2

	x, y := tooltipChooseSide(anchor, popW, popH, winW, winH)

	// Roundness: follows the same CornerRadius / half-min-side formula used by Button.
	roundness := float32(0)
	if style.CornerRadius > 0 {
		halfMin := popH / 2
		if popW/2 < halfMin {
			halfMin = popW / 2
		}
		if halfMin > 0 {
			roundness = style.CornerRadius / halfMin
			if roundness > 1 {
				roundness = 1
			}
		}
	}

	// Drop shadow — 2 px right, 3 px down, dark translucent.
	shadowA := uint8(52 * alpha)
	rl.DrawRectangleRounded(
		rl.NewRectangle(x+2, y+3, popW, popH), roundness, 6,
		rl.NewColor(0, 0, 0, shadowA))

	// Background.
	bg := style.BackgroundColor
	bg.A = uint8(float32(bg.A) * alpha)
	r := rl.NewRectangle(x, y, popW, popH)
	rl.DrawRectangleRounded(r, roundness, 6, bg)

	// Border.
	if style.BorderWidth > 0 {
		bc := style.BorderColor
		bc.A = uint8(float32(bc.A) * alpha)
		rl.DrawRectangleRoundedLinesEx(r, roundness, 6, style.BorderWidth, bc)
	}

	// Text — use drawTextS so sizing matches measureTextS (DrawText would apply
	// chromeFontSize again and spill past the popup at larger type scales).
	textStyle := style
	tc := style.TextColor
	tc.A = uint8(float32(tc.A) * alpha)
	textStyle.TextColor = tc
	posY := int32(y) + int32((popH-fs)/2)
	drawTextS(text, int32(x+pad), posY, textStyle)
}

// ─── Tooltip Widget ───────────────────────────────────────────────────────────

// Tooltip is a decorator widget that wraps any Node and shows a hover popup.
//
// Tooltip embeds Element and satisfies the full Node interface. All tree
// operations — bounds, dirty propagation, layout, draw — are forwarded to the
// wrapped Target so the parent Container sees the Tooltip exactly as if the
// Target were added directly to the tree.
//
// The popup itself is always rendered at screen level by TooltipManager.Draw()
// so it is never scissor-clipped by a parent Viewport.
//
// # LLM Prompt Template
//
//	btn := ui.NewButton("save", "Save", 0, 0, 140, 40)
//	tip := ui.NewTooltip("save-tip", btn, "Save the current file")
//	panel.AddChild(tip) // add tip, not btn
//
// Demo scenes: **Batch 1**, **Batch 5**, **Widgets Demo**.
type Tooltip struct {
	Element
	Target Node            // the wrapped/decorated widget
	Text   *Signal[string] // tooltip text; reactive — changing it takes effect immediately
	Delay  float32         // seconds of hover before popup appears (default 0.4)
	// internal render state
	hovered bool
	hoverT  float32
	visible bool
	alpha   float32 // fade-in progress [0 .. 1]
}

// NewTooltip wraps target inside a Tooltip with the given help text.
//
// Pass the Tooltip to the parent container instead of the Target directly.
// The Tooltip sizes itself to match the Target on every Layout pass.
//
//	tip := ui.NewTooltip("id", someWidget, "Help text")
//	container.AddChild(tip)   // ← tip, not someWidget
func NewTooltip(id string, target Node, text string) *Tooltip {
	// Start with zero bounds — Layout will copy the Target's rendered size.
	var w, h float32
	if target != nil {
		b := target.Bounds()
		w, h = b.Width, b.Height
	}
	t := &Tooltip{
		Element: NewElement(id, 0, 0, w, h),
		Target:  target,
		Text:    NewSignal(text),
		Delay:   0.4,
	}
	t.styleName = "tooltip" // popup style key; SetStyle("x") overrides the popup look
	t.Text.Subscribe(func() { t.MarkDrawDirty() })
	// Wire the target's dirty propagation through this Tooltip so MarkDirty on
	// the target bubbles: target → Tooltip.Element → Tooltip's parent.
	if target != nil {
		target.SetParent(t)
	}
	return t
}

// Children overrides Element.Children so the Inspector tree shows the Target
// as a child of the Tooltip node.
func (t *Tooltip) Children() []Node {
	if t.Target != nil {
		return []Node{t.Target}
	}
	return nil
}

// Update handles hover tracking, fade-in animation, and input forwarding.
func (t *Tooltip) Update(dt float32) {
	if t.IsHidden() {
		return
	}
	// Always forward input to the wrapped widget first.
	if t.Target != nil {
		t.Target.Update(dt)
	}

	mouse := rl.GetMousePosition()
	prevHovered := t.hovered
	t.hovered = rl.CheckCollisionPointRec(mouse, t.Bounds())

	if t.hovered {
		t.hoverT += dt
		if !t.visible && t.hoverT >= t.Delay {
			// Delay expired — begin fade-in.
			t.visible = true
			t.alpha = 0
		}
		// Advance fade-in (~150 ms to full opacity).
		if t.visible && t.alpha < 1 {
			t.alpha += dt / 0.15
			if t.alpha > 1 {
				t.alpha = 1
			}
			t.MarkDrawDirty()
		}
	} else {
		// Mouse left — reset timer and hide instantly.
		t.hoverT = 0
		if t.visible {
			t.visible = false
			t.alpha = 0
			t.MarkDrawDirty()
		}
	}

	// Dismiss on click anywhere outside the Target.
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !t.hovered && t.visible {
		t.visible = false
		t.alpha = 0
		t.MarkDrawDirty()
	}

	if t.hovered != prevHovered {
		t.MarkDrawDirty()
	}

	// Register with the singleton so Draw() will render our popup this frame.
	if t.visible {
		Tooltips.setActiveWidget(t)
	}
}

// Layout sizes the Target to match the Tooltip's bounds and delegates.
func (t *Tooltip) Layout() {
	defer func() { t.layoutDirty = false }()
	if t.Target != nil {
		t.Target.SetBounds(t.Bounds())
		t.Target.Layout()
	}
}

// Draw renders the Target widget. The popup overlay is drawn later by
// TooltipManager.Draw() at screen level so it is never clipped.
func (t *Tooltip) Draw() {
	if t.IsHidden() {
		return
	}
	if t.Target != nil {
		t.Target.Draw()
	}
}

// drawInternal renders the popup overlay for this Tooltip.
// Called exclusively by TooltipManager.Draw(); not part of the normal Draw pass.
func (t *Tooltip) drawInternal(winW, winH int32) {
	if !t.visible || t.alpha <= 0 || t.IsHidden() {
		return
	}
	text := t.Text.Get()
	if text == "" {
		return
	}
	drawTooltipPopup(text, t.Bounds(), t.alpha, winW, winH)
}

// IsInteractive delegates to the Target so the Inspector correctly reports
// input-handling capability.
func (t *Tooltip) IsInteractive() bool {
	if t.Target != nil {
		return t.Target.IsInteractive()
	}
	return false
}

// UsesScissor delegates to the Target.
func (t *Tooltip) UsesScissor() bool {
	if t.Target != nil {
		return t.Target.UsesScissor()
	}
	return false
}

// ─── TooltipManager ───────────────────────────────────────────────────────────

// TooltipManager is the singleton rendering backend for tooltip popups.
// It supports two APIs that can coexist:
//
//  1. Tooltip widget: create NewTooltip(id, target, text) and add it to the
//     widget tree. The widget registers itself via setActiveWidget each frame
//     it is visible.
//
//  2. SetTooltip convenience: attach a popup to any existing Node without
//     restructuring the widget tree. The manager polls hover state in Update.
//
// Drive the manager from the main loop:
//
//	ui.Tooltips.Update(dt)  // after doc.Root.Update
//	ui.Tooltips.Draw()      // inside BeginDrawing, after all widget draws

// tooltipEntry pairs a Node with its text for the SetTooltip convenience API.
type tooltipEntry struct {
	target Node
	text   string
}

// TooltipManager tracks hover state and renders tooltip popups at screen level.
type TooltipManager struct {
	HoverDelay float32 // seconds before popup appears (SetTooltip API default 0.4)
	winW, winH int32   // screen dimensions for positioning; set via SetWindowSize

	// SetTooltip() API state.
	entries   []tooltipEntry
	hoverID   string
	hoverTime float32
	mousePos  rl.Vector2

	// hoverArmed suppresses popups until the pointer moves after scene load or
	// ClearTooltipEntries — avoids phantom hovers when the cursor rests on a
	// control before the first layout pass completes.
	hoverArmed   bool
	armMouse     rl.Vector2
	armMouseSeen bool

	// Tooltip widget API — set by Tooltip.Update() each frame it is visible.
	// The first caller wins (topmost widget under the cursor).
	activeWidget *Tooltip
}

// Tooltips is the package-level singleton. Drive it from the main loop.
var Tooltips = &TooltipManager{HoverDelay: 0.4, winW: 1280, winH: 720}

// SetWindowSize configures the screen dimensions used for smart positioning.
// Call once after rl.InitWindow with the same width/height.
func (m *TooltipManager) SetWindowSize(w, h int32) { m.winW = w; m.winH = h }

// IsActive reports whether a tooltip is visible or a hover target is waiting
// for its popup delay.
func (m *TooltipManager) IsActive() bool {
	return m.activeWidget != nil || m.hoverID != ""
}

// IsAnimating reports tooltip fade-in progress (not the hover delay wait).
func (m *TooltipManager) IsAnimating() bool {
	if m.activeWidget == nil {
		return false
	}
	return m.activeWidget.visible && m.activeWidget.alpha < 1
}

// setActiveWidget is called by Tooltip.Update() to register the currently
// visible Tooltip widget. The first registration per frame wins (topmost under
// cursor). This method is unexported; application code uses NewTooltip instead.
func (m *TooltipManager) setActiveWidget(t *Tooltip) {
	if m.activeWidget == nil {
		m.activeWidget = t
	}
}

// ClearTooltipEntries removes all SetTooltip registrations. Call when tearing
// down a scene so stale node pointers cannot show tooltips on unrelated widgets.
func ClearTooltipEntries() {
	Tooltips.entries = nil
	Tooltips.hoverID = ""
	Tooltips.hoverTime = 0
	Tooltips.activeWidget = nil
	Tooltips.resetHoverArm()
}

func (m *TooltipManager) resetHoverArm() {
	m.hoverArmed = false
	m.armMouse = rl.Vector2{}
	m.armMouseSeen = false
}

func (m *TooltipManager) updateHoverArm() {
	if m.hoverArmed {
		return
	}
	if !m.armMouseSeen {
		m.armMouse = m.mousePos
		m.armMouseSeen = true
		return
	}
	dx := m.mousePos.X - m.armMouse.X
	dy := m.mousePos.Y - m.armMouse.Y
	if dx*dx+dy*dy >= 4 {
		m.hoverArmed = true
	}
}

// SetTooltip registers (or replaces) a tooltip text for an existing Node.
// Use this when you do not want to restructure the widget tree with NewTooltip.
// The target reference is always updated so that scene rebuilds (after Tab
// switching) point to the fresh node rather than a stale one.
func SetTooltip(target Node, text string) {
	for i, e := range Tooltips.entries {
		if e.target.ID() == target.ID() {
			Tooltips.entries[i].text = text
			Tooltips.entries[i].target = target // refresh stale reference
			return
		}
	}
	Tooltips.entries = append(Tooltips.entries, tooltipEntry{target: target, text: text})
}

// Update advances hover timers for the SetTooltip convenience API.
// Call once per frame; order relative to doc.Root.Update does not matter.
func (m *TooltipManager) Update(dt float32) {
	// SetTooltip() API: poll hover state for all registered entries.
	// activeWidget is reset at the END of Draw() so Tooltip widget
	// registrations (which happen during doc.Root.Update) are always visible.
	if len(m.entries) == 0 {
		return
	}
	m.mousePos = rl.GetMousePosition()
	m.updateHoverArm()
	if !m.hoverArmed {
		m.hoverID = ""
		m.hoverTime = 0
		return
	}
	currentID := ""
	for _, e := range m.entries {
		if !e.target.IsHidden() && rl.CheckCollisionPointRec(m.mousePos, e.target.Bounds()) {
			currentID = e.target.ID()
			break
		}
	}
	if currentID == m.hoverID {
		if currentID != "" {
			m.hoverTime += dt
		}
	} else {
		m.hoverID = currentID
		m.hoverTime = 0
	}
}

// Draw renders the active tooltip popup at screen level.
// Call inside BeginDrawing / EndDrawing, after all widget draws.
func (m *TooltipManager) Draw() {
	// Tooltip widget API takes priority over the SetTooltip() legacy path.
	if m.activeWidget != nil {
		m.activeWidget.drawInternal(m.winW, m.winH)
		// Reset now so the slot is clean for the next frame's registrations.
		m.activeWidget = nil
		return
	}

	// SetTooltip() API fallback.
	if m.hoverID == "" || m.hoverTime < m.HoverDelay {
		return
	}
	var text string
	var anchor rl.Rectangle
	for _, e := range m.entries {
		if e.target.ID() == m.hoverID {
			text = e.text
			anchor = e.target.Bounds()
			break
		}
	}
	if text == "" {
		return
	}
	// Legacy path renders at full opacity (no fade-in by design).
	drawTooltipPopup(text, anchor, 1.0, m.winW, m.winH)
}
