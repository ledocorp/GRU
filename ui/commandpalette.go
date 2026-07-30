// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"
	"unicode"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── CommandPalette ──────────────────────────────────────────────────────────
//
// CommandPaletteMgr provides a VS Code-style fuzzy command picker (overlay).
//
//	ui.ShowCommandPalette([]ui.CommandPaletteItem{
//	    {Label: "Open settings", Subtitle: "Navigation", Keywords: "prefs", Action: openSettings},
//	})
//	ui.CloseCommandPalette()
//
// Drive from main.go:
//
//	ui.CommandPaletteMgr.Update(dt)
//	ui.CommandPaletteMgr.Draw() // inside drawScreenOverlays
//
// # LLM Prompt Template
//
//	ui.ShowCommandPalette([]ui.CommandPaletteItem{
//	    {Label: "Open settings", Action: openSettings},
//	})
//
// Demo scenes: **Shell Desktop Demo** (Ctrl+Shift+P).
//
// Styled via "commandpalette", "commandpalette-item", "commandpalette-selected".

const (
	cmdPaletteWidth    float32 = 520
	cmdPalettePad      float32 = 12
	cmdPaletteInputH   float32 = 40
	cmdPaletteItemH    float32 = 44
	cmdPaletteMaxItems = 8
	cmdPaletteTopInset float32 = 96 // below borderless title bar
	cmdPaletteSlideTime float32 = 0.14
	cmdPaletteScrimAlpha float32 = 0.35
)

// CommandPaletteItem is one selectable command row.
type CommandPaletteItem struct {
	Label    string
	Subtitle string
	Keywords string // extra search terms (space-separated)
	Shortcut string // display-only hint, e.g. "Ctrl+Shift+P"
	Action   func()
}

type commandPaletteManager struct {
	open, logicallyOpen, closing bool
	progress                     float32
	items                        []CommandPaletteItem
	filtered                     []int
	query                        string
	selected                     int
	scroll                       int
	skipFrames                   int
	armMouseRelease              bool
	selectionActive              bool // true after explicit Up/Down or row click
	suppressChars                int    // swallow chord key echo (Ctrl+Shift+P)
	CloseOnEscape                bool
	CloseOnBackdrop              bool
}

// CommandPaletteMgr is the package-level command palette manager.
var CommandPaletteMgr = &commandPaletteManager{
	CloseOnEscape:   true,
	CloseOnBackdrop: true,
}

// ShowCommandPalette opens the palette with items. Replaces any open palette.
func ShowCommandPalette(items []CommandPaletteItem) {
	CommandPaletteMgr.items = items
	CommandPaletteMgr.beginOpen()
}

// ShowCommandPaletteFromChord is like ShowCommandPalette but ignores the P key
// echo from Ctrl+Shift+P on the next input poll.
func ShowCommandPaletteFromChord(items []CommandPaletteItem) {
	CommandPaletteMgr.items = items
	CommandPaletteMgr.beginOpen()
	CommandPaletteMgr.suppressChars = 3
}

func (m *commandPaletteManager) beginOpen() {
	m.open = true
	m.logicallyOpen = true
	m.closing = false
	m.query = ""
	m.selected = 0
	m.scroll = 0
	m.selectionActive = false
	m.skipFrames = 2
	m.armMouseRelease = true
	if m.progress <= 0 {
		m.progress = 0
	}
	m.refilter()
}

// CloseCommandPalette dismisses the palette.
func CloseCommandPalette() {
	if !CommandPaletteMgr.open {
		return
	}
	CommandPaletteMgr.logicallyOpen = false
	CommandPaletteMgr.closing = true
}

func (m *commandPaletteManager) reset() {
	m.open = false
	m.logicallyOpen = false
	m.closing = false
	m.progress = 0
	m.skipFrames = 0
	m.armMouseRelease = false
	m.query = ""
	m.selected = 0
	m.scroll = 0
	m.suppressChars = 0
}

// IsCommandPaletteOpen reports logical open state.
func IsCommandPaletteOpen() bool { return CommandPaletteMgr.logicallyOpen }

// IsCommandPaletteVisible reports render visibility including close animation.
func IsCommandPaletteVisible() bool { return CommandPaletteMgr.open }

func (m *commandPaletteManager) IsAnimating() bool {
	if !m.open {
		return false
	}
	return m.closing || m.progress < 0.999 || m.skipFrames > 0
}

func matchCommandQuery(q, label, subtitle, keywords string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return true
	}
	hay := strings.ToLower(label + " " + subtitle + " " + keywords)
	words := strings.Fields(hay)
	for _, w := range strings.Fields(q) {
		if w == "" {
			continue
		}
		if strings.Contains(hay, w) {
			continue
		}
		found := false
		for _, word := range words {
			if strings.HasPrefix(word, w) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (m *commandPaletteManager) refilter() {
	m.filtered = m.filtered[:0]
	for i, it := range m.items {
		if matchCommandQuery(m.query, it.Label, it.Subtitle, it.Keywords) {
			m.filtered = append(m.filtered, i)
		}
	}
	if m.selected >= len(m.filtered) {
		if len(m.filtered) == 0 {
			m.selected = 0
		} else {
			m.selected = len(m.filtered) - 1
		}
	}
	m.clampScroll()
}

func (m *commandPaletteManager) clampScroll() {
	visible := cmdPaletteMaxItems
	if visible > len(m.filtered) {
		visible = len(m.filtered)
	}
	maxScroll := len(m.filtered) - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}
	if m.selected < m.scroll {
		m.scroll = m.selected
	}
	if m.selected >= m.scroll+visible {
		m.scroll = m.selected - visible + 1
	}
}

func (m *commandPaletteManager) activateSelected() {
	if m.selected < 0 || m.selected >= len(m.filtered) {
		return
	}
	idx := m.filtered[m.selected]
	if idx < 0 || idx >= len(m.items) {
		return
	}
	action := m.items[idx].Action
	m.reset()
	PointerClickMarkUsed()
	if action != nil {
		action()
	}
}

func (m *commandPaletteManager) panelRect(sw, sh float32) rl.Rectangle {
	listCount := len(m.filtered)
	if listCount > cmdPaletteMaxItems {
		listCount = cmdPaletteMaxItems
	}
	if listCount < 1 && len(m.items) > 0 {
		listCount = 1
	}
	h := cmdPalettePad*2 + cmdPaletteInputH + 8 + float32(listCount)*cmdPaletteItemH
	if h < cmdPalettePad*2+cmdPaletteInputH+8+cmdPaletteItemH {
		h = cmdPalettePad*2 + cmdPaletteInputH + 8 + cmdPaletteItemH
	}
	maxH := sh - cmdPaletteTopInset - 48
	if h > maxH {
		h = maxH
	}
	x := (sw - cmdPaletteWidth) / 2
	y := cmdPaletteTopInset + (1-m.progress)*(h+24)
	return rl.NewRectangle(x, y, cmdPaletteWidth, h)
}

func (m *commandPaletteManager) itemRect(panel rl.Rectangle, row int) rl.Rectangle {
	y := panel.Y + cmdPalettePad + cmdPaletteInputH + 8 + float32(row)*cmdPaletteItemH
	return rl.NewRectangle(panel.X+cmdPalettePad, y, panel.Width-2*cmdPalettePad, cmdPaletteItemH-2)
}

func (m *commandPaletteManager) inputRect(panel rl.Rectangle) rl.Rectangle {
	return rl.NewRectangle(
		panel.X+cmdPalettePad,
		panel.Y+cmdPalettePad,
		panel.Width-2*cmdPalettePad,
		cmdPaletteInputH,
	)
}

func (m *commandPaletteManager) handleTextInput() {
	if m.suppressChars > 0 {
		m.suppressChars--
		for {
			ch := rl.GetCharPressed()
			if ch <= 0 {
				break
			}
		}
		return
	}
	for {
		ch := rl.GetCharPressed()
		if ch <= 0 {
			break
		}
		if ch < 32 && ch != '\t' {
			continue
		}
		if !unicode.IsPrint(ch) {
			continue
		}
		m.query += string(ch)
		m.selected = 0
		m.scroll = 0
		m.selectionActive = true
		m.refilter()
	}
	if rl.IsKeyPressed(rl.KeyBackspace) && len(m.query) > 0 {
		r := []rune(m.query)
		m.query = string(r[:len(r)-1])
		m.selected = 0
		m.scroll = 0
		if len(m.query) == 0 {
			m.selectionActive = false
		}
		m.refilter()
	}
}

func (m *commandPaletteManager) handleKeys() {
	if rl.IsKeyPressed(rl.KeyDown) {
		if len(m.filtered) > 0 {
			m.selectionActive = true
			m.selected++
			if m.selected >= len(m.filtered) {
				m.selected = len(m.filtered) - 1
			}
			m.clampScroll()
		}
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		if len(m.filtered) > 0 {
			m.selectionActive = true
			m.selected--
			if m.selected < 0 {
				m.selected = 0
			}
			m.clampScroll()
		}
	}
	if rl.IsKeyPressed(rl.KeyHome) && len(m.filtered) > 0 {
		m.selectionActive = true
		m.selected = 0
		m.clampScroll()
	}
	if rl.IsKeyPressed(rl.KeyEnd) && len(m.filtered) > 0 {
		m.selectionActive = true
		m.selected = len(m.filtered) - 1
		m.clampScroll()
	}
	if (rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter)) && m.selectionActive {
		m.activateSelected()
	}
}

// Update advances animation and handles keyboard/mouse input.
func (m *commandPaletteManager) Update(dt float32) {
	if !m.open {
		return
	}
	if m.closing {
		m.progress -= dt / cmdPaletteSlideTime
		if m.progress <= 0 {
			m.progress = 0
			m.open = false
			m.closing = false
			return
		}
	} else if m.progress < 1 {
		m.progress += dt / cmdPaletteSlideTime
		if m.progress > 1 {
			m.progress = 1
		}
	}

	if m.skipFrames > 0 {
		m.skipFrames--
	}
	if m.armMouseRelease {
		if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
			m.armMouseRelease = false
		}
		return
	}

	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	panel := m.panelRect(sw, sh)

	if m.CloseOnEscape && rl.IsKeyPressed(rl.KeyEscape) {
		m.reset()
		return
	}

	m.handleTextInput()
	m.handleKeys()

	if m.CloseOnBackdrop && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if !rl.CheckCollisionPointRec(rl.GetMousePosition(), panel) {
			m.reset()
			return
		}
	}

	mouse := rl.GetMousePosition()
	visible := cmdPaletteMaxItems
	if visible > len(m.filtered) {
		visible = len(m.filtered)
	}
	for row := 0; row < visible; row++ {
		idx := row + m.scroll
		if idx >= len(m.filtered) {
			break
		}
		r := m.itemRect(panel, row)
		if rl.CheckCollisionPointRec(mouse, r) {
			m.selected = idx
		}
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, r) {
			m.selectionActive = true
			m.selected = idx
			m.activateSelected()
			return
		}
	}
}

// Draw renders scrim + palette panel. Call from drawScreenOverlays.
func (m *commandPaletteManager) Draw() {
	if !m.open || m.progress <= 0 {
		return
	}
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	a := m.progress
	scrim := uint8(cmdPaletteScrimAlpha * 255 * a)
	rl.DrawRectangle(0, 0, int32(sw), int32(sh), rl.NewColor(0, 0, 0, scrim))

	panel := m.panelRect(sw, sh)
	style := GetThemeStyle("commandpalette")
	bg := style.BackgroundColor
	bg.A = uint8(float32(bg.A) * a)
	rl.DrawRectangleRounded(panel, 0.06, 8, bg)
	bc := style.BorderColor
	bc.A = uint8(float32(bc.A) * a)
	rl.DrawRectangleLinesEx(panel, 1, bc)

	input := m.inputRect(panel)
	inBg := GetThemeStyle("commandpalette-input").BackgroundColor
	inBg.A = uint8(float32(inBg.A) * a)
	rl.DrawRectangleRounded(input, 0.12, 6, inBg)
	inStyle := GetThemeStyle("commandpalette-input")
	inStyle.TextColor.A = uint8(float32(inStyle.TextColor.A) * a)
	q := m.query
	if q == "" {
		q = "Type a command…"
		inStyle.TextColor = rl.NewColor(inStyle.TextColor.R, inStyle.TextColor.G, inStyle.TextColor.B, uint8(float32(140)*a))
	}
	drawTextS(q, int32(input.X+10), TextPosY(input, inStyle), inStyle)

	itemStyle := GetThemeStyle("commandpalette-item")
	selStyle := GetThemeStyle("commandpalette-selected")
	visible := cmdPaletteMaxItems
	if visible > len(m.filtered) {
		visible = len(m.filtered)
	}
	for row := 0; row < visible; row++ {
		idx := row + m.scroll
		if idx >= len(m.filtered) {
			break
		}
		it := m.items[m.filtered[idx]]
		r := m.itemRect(panel, row)
		if idx == m.selected {
			sb := selStyle.BackgroundColor
			sb.A = uint8(float32(sb.A) * a)
			rl.DrawRectangleRounded(r, 0.08, 6, sb)
		}
		titleStyle := itemStyle
		titleStyle.Bold = true
		titleStyle.TextColor.A = uint8(float32(titleStyle.TextColor.A) * a)
		drawTextS(it.Label, int32(r.X+8), int32(r.Y+6), titleStyle)
		if it.Subtitle != "" {
			subStyle := itemStyle
			subStyle.FontSize = int32(float32(subStyle.FontSize) * 0.9)
			if subStyle.FontSize < 12 {
				subStyle.FontSize = 12
			}
			subStyle.TextColor.A = uint8(float32(subStyle.TextColor.A) * 0.85 * a)
			drawTextS(it.Subtitle, int32(r.X+8), int32(r.Y+24), subStyle)
		}
		if it.Shortcut != "" {
			shStyle := itemStyle
			shStyle.TextColor.A = uint8(float32(shStyle.TextColor.A) * 0.7 * a)
			sw := measureTextS(it.Shortcut, shStyle)
			drawTextS(it.Shortcut, int32(r.X+r.Width-float32(sw)-8), int32(r.Y+(r.Height-float32(shStyle.FontSize))/2), shStyle)
		}
	}

	if len(m.filtered) == 0 {
		emptyStyle := itemStyle
		emptyStyle.TextColor.A = uint8(float32(emptyStyle.TextColor.A) * 0.7 * a)
		r := m.itemRect(panel, 0)
		drawTextS("No matching commands", int32(r.X+8), TextPosY(r, emptyStyle), emptyStyle)
	}
}

// KeyChordCtrlShiftP reports Ctrl+Shift+P (common command palette shortcut).
func KeyChordCtrlShiftP() bool {
	if !rl.IsKeyPressed(rl.KeyP) {
		return false
	}
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	return ctrl && shift
}
