// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── ContextMenu ─────────────────────────────────────────────────────────────
//
// ContextMenu provides a package-level manager for transient right-click popup
// menus. Show the menu with ShowContextMenu and it closes automatically on
// Escape, an outside click, or when the user activates an item.
//
// # Usage
//
//	items := []ui.ContextMenuItem{
//	    {Label: "Cut",   Action: func() { doCut() }},
//	    {Label: "Copy",  Action: func() { doCopy() }},
//	    {Divider: true},
//	    {Label: "More", SubItems: []ui.ContextMenuItem{
//	        {Label: "Option A", Action: func() { doA() }},
//	        {Label: "Option B", Action: func() { doB() }},
//	    }},
//	    {Label: "Paste", Disabled: true},
//	}
//	// On right-click (in scene OnUpdate):
//	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
//	    mp := rl.GetMousePosition()
//	    ui.ShowContextMenu(items, mp.X, mp.Y)
//	}
//
// Drive the manager every frame from main.go:
//
//	ui.ContextMenuMgr.Update(dt)  // after doc.Root.Update so MenuBar can switch menus first
//	ui.ContextMenuMgr.Draw()      // inside BeginDrawing, after all widget draws
//
// # Sub-menus
//
// Set SubItems on any item to nest one additional level. Hovering the parent
// item opens the sub-menu to the right (auto-flips left near the screen edge).
// Clicking a sub-menu item activates it and closes both menus.
//
// # Fade animation
//
// ShowContextMenu resets a Tween that drives alpha 0→255 over 100 ms
// (EaseOutQuad). The entire menu fades in as a unit.
//
// # Styles
//
// Styled via "contextmenu", "contextmenu-item", "contextmenu-hover", and
// "contextmenu-divider" theme keys.
//
// # LLM Prompt Template
//
//	items := []ui.ContextMenuItem{
//	    {Label: "Copy", Action: func() { doCopy() }},
//	    {Label: "Paste", Disabled: true},
//	}
//	ui.ShowContextMenu(items, mp.X, mp.Y)
//	// main loop: ui.ContextMenuMgr.Update(dt); ui.ContextMenuMgr.Draw()
//
// Demo scenes: **Batch 1**, **AppShell Demo**, Notepad tab context menu.

const (
	contextMenuCheckCol      = float32(20)
	contextMenuShortcutGap   = float32(24)
	contextMenuMinWidth      = float32(220)
	contextMenuRoundSegments = int32(32)
)

// ContextMenuItem describes one row in the popup menu.
type ContextMenuItem struct {
	Label    string            // display text (ignored when Divider is true)
	Shortcut string            // optional hint shown trailing, e.g. "Ctrl+S"
	Action   func()            // callback on activation; nil = no-op
	Disabled bool              // greyed-out and non-interactive
	Divider  bool              // renders a thin separator line (Label is ignored)
	Checked  bool              // shows a checkmark when true (toggle / option state)
	SubItems []ContextMenuItem // non-empty → hovering opens a nested sub-menu
}

// contextMenuManager is the singleton implementation backing ContextMenuMgr.
type contextMenuManager struct {
	open  bool
	items []ContextMenuItem
	x, y  float32

	hovered int // index of hovered main-menu item (-1 = none)

	// per-item geometry constants
	itemH float32
	divH  float32
	menuW float32
	menuH float32

	// fade-in animation: drives alpha 0→255 on ShowContextMenu
	fadeAlpha float32
	fadeTween *Tween

	// justOpened is set true by ShowContextMenu and cleared at the start of the
	// next Update call. It prevents the "outside click" auto-close from firing
	// on the same frame that a button's OnClick opens the menu.
	justOpened bool

	// sub-menu state (one level deep)
	subOpen    bool    // sub-menu is currently visible
	subIndex   int     // index in m.items whose SubItems are shown (-1 = none)
	subX, subY float32 // top-left of sub-menu box
	subMenuH   float32 // cached height of sub-menu box
	subHovered int     // hovered index inside sub-menu (-1 = none)
}

// ContextMenuMgr is the singleton context-menu manager.
// Drive it every frame from your main loop.
var ContextMenuMgr = &contextMenuManager{
	hovered:    -1,
	subIndex:   -1,
	subHovered: -1,
	itemH:      36,
	divH:       10,
	menuW:      208,
	fadeAlpha:  255,
}

// ShowContextMenu opens the popup at (x, y) with the given items and starts
// the fade-in animation. Replaces any previously open menu.
func ShowContextMenu(items []ContextMenuItem, x, y float32) {
	// Native popup menus take pointer/keyboard focus away from embedded web hosts.
	ReleaseWebKeyboardFocus()
	m := ContextMenuMgr
	m.items = items
	m.x = x
	m.y = y
	m.menuW = contextMenuMeasureWidth(items)
	m.hovered = -1
	m.subOpen = false
	m.subIndex = -1
	m.subHovered = -1
	m.open = true
	m.justOpened = true
	m.fadeAlpha = 0
	m.recalcGeometry()

	// 100 ms fade-in (EaseOutQuad).
	m.fadeTween = NewTween(0, 255, 0.10, EaseOutQuad,
		func(v float32) { m.fadeAlpha = v },
		nil,
	)
}

// CloseContextMenu closes the menu immediately and presents one frame so the menu
// vanishes before blocking actions (native file dialogs, etc.).
func CloseContextMenu() {
	ContextMenuMgr.open = false
	PresentScreen()
}

// IsContextMenuOpen reports whether the popup is currently visible.
func IsContextMenuOpen() bool { return ContextMenuMgr.open }

// IsAnimating reports the menu fade-in tween.
func (m *contextMenuManager) IsAnimating() bool {
	return m.open && m.fadeTween != nil && m.fadeTween.IsActive
}

// DebugInfo returns a one-line summary used by the Inspector.
func (m *contextMenuManager) DebugInfo() string {
	if !m.open {
		return "closed"
	}
	sub := "no sub"
	if m.subOpen {
		sub = fmt.Sprintf("sub[%d] open", m.subIndex)
	}
	return fmt.Sprintf("open  items:%d  alpha:%.0f  %s",
		len(m.items), m.fadeAlpha, sub)
}

// ─── geometry ─────────────────────────────────────────────────────────────────

// recalcGeometry recomputes menuH and clamps x/y so the menu stays on screen.
func (m *contextMenuManager) recalcGeometry() {
	style := GetThemeStyle("contextmenu")
	h := style.Padding * 2
	for _, item := range m.items {
		if item.Divider {
			h += m.divH
		} else {
			h += m.itemH
		}
	}
	m.menuH = h

	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	if m.x+m.menuW > sw-4 {
		m.x -= m.menuW
		if m.x < 4 {
			m.x = 4
		}
	}
	if m.y+m.menuH > sh-4 {
		m.y = sh - m.menuH - 4
		if m.y < 4 {
			m.y = 4
		}
	}
}

func contextMenuMeasureWidth(items []ContextMenuItem) float32 {
	styleItem := GetThemeStyle("contextmenu-item")
	shortcutStyle := GetThemeStyle("contextmenu-shortcut")
	maxW := contextMenuMinWidth
	var measure func([]ContextMenuItem)
	measure = func(list []ContextMenuItem) {
		for _, item := range list {
			if item.Divider {
				continue
			}
			w := float32(measureTextS(item.Label, styleItem)) + styleItem.Padding*2
			if item.Shortcut != "" {
				w += contextMenuShortcutGap + float32(measureTextS(item.Shortcut, shortcutStyle))
			}
			if item.Checked {
				w += contextMenuCheckCol + 8
			}
			if len(item.SubItems) > 0 {
				w += 16
			}
			if w > maxW {
				maxW = w
			}
			if len(item.SubItems) > 0 {
				measure(item.SubItems)
			}
		}
	}
	measure(items)
	return maxW
}

// calcSubHeight returns the pixel height for the sub-menu at items[subIndex].
func (m *contextMenuManager) calcSubHeight(subIndex int) float32 {
	if subIndex < 0 || subIndex >= len(m.items) {
		return 0
	}
	sub := m.items[subIndex].SubItems
	style := GetThemeStyle("contextmenu")
	h := style.Padding * 2
	for _, item := range sub {
		if item.Divider {
			h += m.divH
		} else {
			h += m.itemH
		}
	}
	return h
}

// ─── Update ───────────────────────────────────────────────────────────────────

// Update advances the fade animation, tracks hover, handles item activation,
// and auto-closes on Escape or an outside click.
// Call once per frame before Draw, passing the frame delta-time.
func (m *contextMenuManager) Update(dt float32) {
	if !m.open {
		return
	}

	if !rl.IsWindowFocused() {
		m.open = false
		m.subOpen = false
		return
	}

	// If ShowContextMenu was called this frame (e.g. from a button's OnClick),
	// skip the outside-click check so the menu doesn't immediately self-close.
	if m.justOpened {
		m.justOpened = false
		return
	}

	// Advance fade-in tween.
	if m.fadeTween != nil && m.fadeTween.IsActive {
		m.fadeTween.Update(dt)
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		m.open = false
		return
	}

	mouse := rl.GetMousePosition()
	menuRect := rl.NewRectangle(m.x, m.y, m.menuW, m.menuH)
	inMain := rl.CheckCollisionPointRec(mouse, menuRect)

	// Build sub-menu rect (recompute height in case sub changed).
	var subRect rl.Rectangle
	inSub := false
	if m.subOpen && m.subIndex >= 0 {
		m.subMenuH = m.calcSubHeight(m.subIndex)
		subRect = rl.NewRectangle(m.subX, m.subY, m.menuW, m.subMenuH)
		inSub = rl.CheckCollisionPointRec(mouse, subRect)
	}

	// Close on any click outside both menus.
	leftClick := rl.IsMouseButtonPressed(rl.MouseLeftButton)
	rightClick := rl.IsMouseButtonPressed(rl.MouseRightButton)
	if (leftClick || rightClick) && !inMain && !inSub {
		m.open = false
		ReleaseWebKeyboardFocus()
		return
	}

	style := GetThemeStyle("contextmenu")

	// ── Main menu ─────────────────────────────────────────────────────────
	if inMain {
		m.hovered = -1
		rowY := m.y + style.Padding
		for i, item := range m.items {
			if item.Divider {
				rowY += m.divH
				continue
			}
			rr := rl.NewRectangle(m.x+style.Padding, rowY, m.menuW-style.Padding*2, m.itemH)
			if rl.CheckCollisionPointRec(mouse, rr) && !item.Disabled {
				m.hovered = i
				if len(item.SubItems) > 0 {
					// Open / refresh sub-menu anchored to this row.
					m.openSubMenu(i, rowY)
				} else {
					// Hovering a leaf item closes any open sub-menu.
					m.subOpen = false
					m.subIndex = -1
					if leftClick {
						m.open = false
						PointerClickMarkUsed()
						if item.Action != nil {
							item.Action()
						}
						return
					}
				}
			}
			rowY += m.itemH
		}
	}

	// ── Sub-menu ──────────────────────────────────────────────────────────
	if m.subOpen && inSub && m.subIndex >= 0 {
		sub := m.items[m.subIndex].SubItems
		m.subHovered = -1
		rowY := m.subY + style.Padding
		for i, item := range sub {
			if item.Divider {
				rowY += m.divH
				continue
			}
			rr := rl.NewRectangle(m.subX+style.Padding, rowY, m.menuW-style.Padding*2, m.itemH)
			if rl.CheckCollisionPointRec(mouse, rr) && !item.Disabled {
				m.subHovered = i
				if leftClick {
					m.open = false
					PointerClickMarkUsed()
					if item.Action != nil {
						item.Action()
					}
					return
				}
			}
			rowY += m.itemH
		}
	}
}

// openSubMenu positions and shows the sub-menu for items[mainIndex].
// rowY is the screen-space top of the parent item row.
func (m *contextMenuManager) openSubMenu(mainIndex int, rowY float32) {
	if m.subOpen && m.subIndex == mainIndex {
		return // already showing this sub-menu
	}
	m.subIndex = mainIndex
	m.subOpen = true
	m.subHovered = -1
	m.subMenuH = m.calcSubHeight(mainIndex)

	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())

	sx := m.x + m.menuW
	if sx+m.menuW > sw-4 {
		sx = m.x - m.menuW
		if sx < 4 {
			sx = 4
		}
	}
	sy := rowY
	if sy+m.subMenuH > sh-4 {
		sy = sh - m.subMenuH - 4
		if sy < 4 {
			sy = 4
		}
	}
	m.subX = sx
	m.subY = sy
}

// ─── Draw ─────────────────────────────────────────────────────────────────────

// Draw renders the context menu and any open sub-menu on top of all other
// content. Call inside BeginDrawing / EndDrawing, after all widget draws.
func (m *contextMenuManager) Draw() {
	if !m.open {
		return
	}

	alpha := uint8(m.fadeAlpha)
	style := GetThemeStyle("contextmenu")
	styleItem := GetThemeStyle("contextmenu-item")
	styleHover := GetThemeStyle("contextmenu-hover")
	styleDivider := GetThemeStyle("contextmenu-divider")

	m.drawMenuBox(m.x, m.y, m.menuH, m.items, m.hovered, alpha,
		style, styleItem, styleHover, styleDivider)

	if m.subOpen && m.subIndex >= 0 && m.subIndex < len(m.items) {
		sub := m.items[m.subIndex].SubItems
		m.drawMenuBox(m.subX, m.subY, m.subMenuH, sub, m.subHovered, alpha,
			style, styleItem, styleHover, styleDivider)
	}
}

// drawMenuBox renders one popup box (main or sub-menu) at (x, y).
func (m *contextMenuManager) drawMenuBox(
	x, y, boxH float32,
	items []ContextMenuItem,
	hovered int,
	alpha uint8,
	style, styleItem, styleHover, styleDivider Style,
) {
	menuRect := snapControlRect(rl.NewRectangle(x, y, m.menuW, boxH))
	alphaf := float32(alpha) / 255.0
	round := buttonCornerRoundness(menuRect.Width, menuRect.Height, style.CornerRadius)
	if round < 0.08 {
		round = 0.08
	}
	seg := contextMenuRoundSegments

	// Multi-layer drop shadow.
	shadowDefs := [3]struct {
		off float32
		a   uint8
	}{{6, 12}, {4, 22}, {2, 34}}
	for _, s := range shadowDefs {
		a := uint8(float32(s.a) * alphaf)
		sr := snapControlRect(rl.NewRectangle(menuRect.X+s.off, menuRect.Y+s.off, menuRect.Width, menuRect.Height))
		rl.DrawRectangleRounded(sr, round, seg, rl.NewColor(0, 0, 0, a))
	}

	bg := style.BackgroundColor
	bg.A = alpha
	borderCol := style.BorderColor
	borderCol.A = uint8(float32(borderCol.A) * alphaf)
	drawRoundedInsetBorder(menuRect, round, style.BorderWidth, borderCol, bg)

	shortcutStyle := GetThemeStyle("contextmenu-shortcut")
	shortcutStyle.TextColor.A = uint8(float32(shortcutStyle.TextColor.A) * alphaf)
	rowRound := float32(0.2)
	if m.itemH > 0 {
		rowRound = 6 / m.itemH
		if rowRound > 0.35 {
			rowRound = 0.35
		}
	}

	rowY := y + style.Padding
	for i, item := range items {
		if item.Divider {
			divCol := styleDivider.BackgroundColor
			divCol.A = uint8(float32(divCol.A) * alphaf)
			rl.DrawRectangle(
				int32(x+style.Padding), int32(rowY+m.divH/2-1),
				int32(m.menuW-style.Padding*2), 1,
				divCol,
			)
			rowY += m.divH
			continue
		}

		rr := rl.NewRectangle(x+style.Padding, rowY, m.menuW-style.Padding*2, m.itemH)

		// Hover highlight.
		if hovered == i {
			hoverBg := styleHover.BackgroundColor
			hoverBg.A = uint8(float32(hoverBg.A) * alphaf)
			rl.DrawRectangleRounded(rr, rowRound, contextMenuRoundSegments, hoverBg)
		}

		// Label text.
		textColor := styleItem.TextColor
		if item.Disabled {
			textColor = rl.NewColor(160, 162, 180, 180)
		}
		textColor.A = uint8(float32(textColor.A) * alphaf)
		itemStyle := styleItem
		itemStyle.TextColor = textColor
		itemStyle.MinFontSize = 15
		tx := int32(rr.X) + int32(styleItem.Padding)
		drawTextS(item.Label, tx, TextPosY(rr, itemStyle), itemStyle)

		trail := rr.X + rr.Width - styleItem.Padding
		if len(item.SubItems) > 0 {
			trail -= 14
		}
		if item.Checked {
			iconSize := float32(16)
			checkX := trail - iconSize - 4
			checkY := rr.Y + (m.itemH-iconSize)/2
			dst := rl.NewRectangle(checkX, checkY, iconSize, iconSize)
			DrawPhosphorIcon(dst, PhosphorCheck, PhosphorBold, textColor)
			trail = checkX - 8
		}
		if item.Shortcut != "" {
			sw := float32(measureTextS(item.Shortcut, shortcutStyle))
			drawTextS(item.Shortcut, int32(trail-sw), TextPosY(rr, shortcutStyle), shortcutStyle)
		}

		// Right-pointing arrow for sub-menu items.
		if len(item.SubItems) > 0 {
			arrowMY := rowY + m.itemH/2
			arrowR := x + m.menuW - style.Padding - 2
			rl.DrawTriangle(
				rl.NewVector2(arrowR-7, arrowMY-4),
				rl.NewVector2(arrowR-7, arrowMY+4),
				rl.NewVector2(arrowR, arrowMY),
				textColor,
			)
		}

		rowY += m.itemH
	}
}
