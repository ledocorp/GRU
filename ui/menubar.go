// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	menuBarDefaultH    float32 = 35 // 28px base + 25% chrome height; font size unchanged
	menuBarPadX        float32 = 8
	menuBarItemPadX    float32 = 10
	menuBarItemGap     float32 = 2
	menuBarItemMinW    float32 = 36
)

// MenuBarMenu is one top-level menu (File, Edit, View) with dropdown items.
// Items reuse ContextMenuItem so MenuBar composes the existing ContextMenu manager.
type MenuBarMenu struct {
	Label string
	Items []ContextMenuItem
	// ItemsFunc, when set, supplies items at open time (e.g. View menu checkmarks).
	ItemsFunc func() []ContextMenuItem
}

// MenuBar is a horizontal desktop menu strip. Clicking a label opens a dropdown
// via ShowContextMenu anchored below that label. While a dropdown is open, hovering
// another top-level label switches to that menu (Windows-style menu bar).
//
// # LLM Prompt Template
//
//	mb := ui.NewMenuBar("menu", []ui.MenuBarMenu{{
//	    Label: "File",
//	    Items: []ui.ContextMenuItem{{Label: "New", Action: func() { newDoc() }}},
//	}}, 0, 0, 0, 0)
//	shell.AddChild(mb)
//
// Demo scenes: **Shell Desktop Demo**, **Settings Desktop**, Notepad menubar.
type MenuBar struct {
	Element
	Menus             []MenuBarMenu
	hoverIdx                 int
	openMenuIdx              int  // top-level menu whose dropdown is active; -1 when none
	menuSessionActive        bool // true after a click opened a dropdown this session
	suppressHoverUntilLeave  bool // after dismiss, ignore hover until cursor leaves the bar
	menuRects                []rl.Rectangle
}

// NewMenuBar creates a menu bar. Pass w=0, h=0 for flex width and intrinsic height.
func NewMenuBar(id string, menus []MenuBarMenu, x, y, w, h float32) *MenuBar {
	if len(menus) == 0 {
		menus = []MenuBarMenu{{Label: "Menu", Items: nil}}
	}
	mb := &MenuBar{
		Element:     NewElement(id, x, y, w, h),
		Menus:       menus,
		hoverIdx:    -1,
		openMenuIdx: -1,
	}
	mb.styleName = "menubar"
	mb.ZIndex = 100
	if h == 0 {
		mb.AutoHeight = true
	}
	return mb
}

// IsInteractive implements Node.
func (mb *MenuBar) IsInteractive() bool { return true }

// Update handles hover and opens dropdown menus on click.
func (mb *MenuBar) Update(_ float32) {
	if mb.IsHidden() {
		return
	}
	mb.layoutMenuRects()

	if !IsContextMenuOpen() {
		if mb.openMenuIdx >= 0 || mb.menuSessionActive {
			mouse := rl.GetMousePosition()
			onBar := rl.CheckCollisionPointRec(mouse, mb.Bounds())
			// Only suppress hover after an outside dismiss — not when switching
			// top-level labels on the bar (File → Edit) or clicking to reopen.
			if !onBar {
				mb.suppressHoverUntilLeave = true
			}
			mb.MarkDrawDirty()
		}
		mb.menuSessionActive = false
		mb.openMenuIdx = -1
	}

	mouse := rl.GetMousePosition()
	onBar := rl.CheckCollisionPointRec(mouse, mb.Bounds())
	if mb.suppressHoverUntilLeave && !onBar {
		mb.suppressHoverUntilLeave = false
	}

	prev := mb.hoverIdx
	mb.hoverIdx = -1
	if !mb.suppressHoverUntilLeave {
		mb.hoverIdx = mb.menuIndexAt(mouse)
	}
	if mb.hoverIdx != prev {
		mb.MarkDrawDirty()
	}

	// After a click opened a menu, hovering another label switches dropdowns.
	if mb.menuSessionActive && IsContextMenuOpen() && mb.hoverIdx >= 0 &&
		mb.hoverIdx != mb.openMenuIdx {
		mb.openDropdown(mb.hoverIdx)
		return
	}

	clickIdx := mb.menuClickIndex(mouse)
	if clickIdx < 0 {
		return
	}
	r := mb.menuRects[clickIdx]
	if !mb.menuClickOn(r) {
		return
	}
	mb.openDropdown(clickIdx)
	mb.menuSessionActive = true
	mb.suppressHoverUntilLeave = false
}

func (mb *MenuBar) menuIndexAt(pos rl.Vector2) int {
	for i, r := range mb.menuRects {
		if rl.CheckCollisionPointRec(pos, r) {
			return i
		}
	}
	return -1
}

// menuClickIndex resolves the label under a click even when hover is suppressed
// after an outside dismiss (user can still click File/Edit without leaving the bar).
func (mb *MenuBar) menuClickIndex(mouse rl.Vector2) int {
	if idx := mb.menuIndexAt(mouse); idx >= 0 {
		return idx
	}
	if PointerClickPending() {
		return mb.menuIndexAt(PointerClickPosition())
	}
	return -1
}

func (mb *MenuBar) menuClickOn(r rl.Rectangle) bool {
	if PointerClickConsume(r) {
		return true
	}
	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		return false
	}
	mouse := rl.GetMousePosition()
	return rl.CheckCollisionPointRec(mouse, r) ||
		(PointerClickPending() && rl.CheckCollisionPointRec(PointerClickPosition(), r))
}

func (mb *MenuBar) openDropdown(idx int) {
	if idx < 0 || idx >= len(mb.Menus) {
		return
	}
	menu := mb.Menus[idx]
	items := menu.Items
	if menu.ItemsFunc != nil {
		items = menu.ItemsFunc()
	}
	if len(items) == 0 {
		return
	}
	r := mb.menuRects[idx]
	mb.openMenuIdx = idx
	ShowContextMenu(items, r.X, r.Y+r.Height)
}

// Layout clamps to menuBarDefaultH when AutoHeight is set.
func (mb *MenuBar) Layout() {
	if !mb.IsAutoHeight() {
		mb.layoutMenuRects()
		mb.layoutDirty = false
		return
	}
	b := mb.Bounds()
	if b.Height < menuBarDefaultH-0.5 || b.Height > menuBarDefaultH+0.5 {
		b.Height = menuBarDefaultH
		mb.setBoundsNoMark(b)
	}
	mb.layoutMenuRects()
	mb.layoutDirty = false
}

func (mb *MenuBar) layoutMenuRects() {
	b := mb.Bounds()
	style := mb.GetStyle()
	fs := EffectiveFontSize(style)
	if fs <= 0 {
		fs = 14
	}
	mb.menuRects = mb.menuRects[:0]
	x := b.X + menuBarPadX
	for _, menu := range mb.Menus {
		textW := float32(measureTextS(menu.Label, style))
		w := textW + menuBarItemPadX*2
		if w < menuBarItemMinW {
			w = menuBarItemMinW
		}
		mb.menuRects = append(mb.menuRects, rl.NewRectangle(x, b.Y, w, b.Height))
		x += w + menuBarItemGap
	}
}

// Draw implements Node.Draw.
func (mb *MenuBar) Draw() {
	defer func() { mb.drawDirty = false }()
	mb.drawInternal()
}

func (mb *MenuBar) highlightMenuIdx() int {
	if mb.suppressHoverUntilLeave {
		return -1
	}
	return mb.hoverIdx
}

func (mb *MenuBar) drawInternal() {
	if mb.IsHidden() {
		return
	}
	if len(mb.menuRects) != len(mb.Menus) {
		mb.layoutMenuRects()
	}
	b := mb.Bounds()
	style := mb.GetStyle()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, style.BackgroundColor)
	}
	if style.BorderWidth > 0 && style.BorderColor.A > 0 {
		rl.DrawLineEx(
			rl.NewVector2(b.X, b.Y+b.Height),
			rl.NewVector2(b.X+b.Width, b.Y+b.Height),
			style.BorderWidth, style.BorderColor)
	}
	highlight := mb.highlightMenuIdx()
	hoverStyle := GetThemeStyle("menubar-hover")
	for i, menu := range mb.Menus {
		r := mb.menuRects[i]
		if i == highlight && hoverStyle.BackgroundColor.A > 0 {
			rl.DrawRectangleRec(r, hoverStyle.BackgroundColor)
		}
		textStyle := style
		if i == highlight && hoverStyle.TextColor.A > 0 {
			textStyle.TextColor = hoverStyle.TextColor
		}
		tw := measureTextS(menu.Label, textStyle)
		tx := int32(r.X + (r.Width-float32(tw))/2)
		ty := TextPosY(r, textStyle)
		drawTextS(menu.Label, tx, ty, textStyle)
	}
}
