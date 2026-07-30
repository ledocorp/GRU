// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ─── TabView ─────────────────────────────────────────────────────────────────
//
// TabView is a container that displays one tab at a time. Tabs are added via
// AddTab; the active tab is controlled by the reactive Active signal.
//
//	tv := ui.NewTabView("tabs", 10, 10, 600, 400)
//	tv.AddTab("Overview", overviewNode)
//	tv.AddTabWithIcon("Settings", "⚙", settingsNode)
//	tv.Active.Set(0)
//
// # Layouts
//
// Horizontal (default): tab bar is 34 px tall across the top; content fills
// the remaining area below.
//
// Vertical: set Vertical = true before adding to the tree. The tab bar is
// 160 px wide on the left; content fills the remaining area to the right.
//
// # Icons
//
// AddTabWithIcon prepends a single UTF-8 glyph to the tab label:
//
//	tv.AddTabWithIcon("Help", "?", helpNode)
//
// # Disabled tabs
//
// Disabled tabs render with muted colours and do not respond to clicks:
//
//	tv.AddTabWithOptions("Draft", "", draftNode, true) // disabled=true
//	tv.DisableTab(2, true)
//
// # Signals
//
// Active is a *Signal[int]; subscribe to be notified on tab switches:
//
//	tv.Active.Subscribe(func() { fmt.Println("tab", tv.Active.Get()) })
//
// # Style keys used
//
//   - "tab"           — inactive header
//   - "tab-active"    — active header (indigo bottom/right accent)
//   - "tab-hover"     — hovered-inactive header
//   - "tab-disabled"  — disabled header

// tabBarH is the height (px) of the horizontal tab bar.
// Also referenced by ui/json.go (jsonTabView builds content rect from it).
const tabBarH = float32(34)

// tabBarW is the width (px) of the vertical tab bar.
const tabBarW = float32(160)

// tabEntry stores a single tab's metadata and content node.
type tabEntry struct {
	title    string
	icon     string // optional leading glyph (e.g. "★", "⚙")
	content  Node
	disabled bool
}

// TabView is a switchable tab container.
//
// Embed in your widget tree and populate with AddTab / AddTabWithIcon.
//
// # LLM Prompt Template
//
//	tv := ui.NewTabView("tabs", 0, 0, 0, 400)
//	tv.AddTab("Overview", overviewNode)
//	tv.AddTab("Settings", settingsNode)
//	panel.AddChild(tv)
//
// Demo scenes: **Batch 1**, **Widgets Demo**.
type TabView struct {
	Element
	// Active is the 0-based index of the currently visible tab.
	// Reactive: external code can read/write it freely.
	Active *Signal[int]
	// Vertical switches to a left-side tab bar layout.
	// Set before adding the TabView to the widget tree.
	Vertical bool
	tabs     []tabEntry
	hovered  int // index of hovered tab header, -1 when none
}

// NewTabView creates an empty TabView at the given position and size.
func NewTabView(id string, x, y, w, h float32) *TabView {
	tv := &TabView{
		Element: NewElement(id, x, y, w, h),
		Active:  NewSignal(0),
		hovered: -1,
	}
	tv.Active.Subscribe(func() { tv.MarkDirty() })
	return tv
}

// AddTab appends a plain text tab. content's bounds are managed by Layout.
func (tv *TabView) AddTab(title string, content Node) *TabView {
	return tv.AddTabWithOptions(title, "", content, false)
}

// AddTabWithIcon appends a tab with a leading icon glyph before the title.
// icon should be a single UTF-8 character such as "★", "⚙", or "✓".
func (tv *TabView) AddTabWithIcon(title, icon string, content Node) *TabView {
	return tv.AddTabWithOptions(title, icon, content, false)
}

// AddTabWithOptions is the full tab constructor.
// Pass disabled=true to make the tab non-interactive and visually muted.
func (tv *TabView) AddTabWithOptions(title, icon string, content Node, disabled bool) *TabView {
	if content != nil {
		content.SetParent(tv)
	}
	tv.tabs = append(tv.tabs, tabEntry{
		title:    title,
		icon:     icon,
		content:  content,
		disabled: disabled,
	})
	tv.MarkDirty()
	return tv
}

// DisableTab enables or disables the tab at index.
func (tv *TabView) DisableTab(index int, disabled bool) {
	if index >= 0 && index < len(tv.tabs) {
		tv.tabs[index].disabled = disabled
		tv.MarkDrawDirty()
	}
}

// TabCount returns the number of registered tabs.
func (tv *TabView) TabCount() int { return len(tv.tabs) }

// Children returns all tab content nodes so the Inspector tree can walk them.
func (tv *TabView) Children() []Node {
	out := make([]Node, 0, len(tv.tabs))
	for _, t := range tv.tabs {
		if t.content != nil {
			out = append(out, t.content)
		}
	}
	return out
}

// IsInteractive returns true because tab headers are clickable.
func (tv *TabView) IsInteractive() bool { return true }

// ─── Update ──────────────────────────────────────────────────────────────────

// Update handles hover/click on tab headers and delegates to active content.
func (tv *TabView) Update(dt float32) {
	if tv.IsHidden() {
		return
	}
	b := tv.Bounds()
	mouse := rl.GetMousePosition()
	prevHovered := tv.hovered
	tv.hovered = -1

	if tv.Vertical {
		tv.updateVertical(b, mouse)
	} else {
		tv.updateHorizontal(b, mouse)
	}

	if rl.CheckCollisionPointRec(mouse, b) {
		tv.handleKeyboard()
	}
	if tv.hovered != prevHovered {
		tv.MarkDrawDirty()
	}

	// Delegate Update to active content only.
	idx := tv.Active.Get()
	if idx >= 0 && idx < len(tv.tabs) && tv.tabs[idx].content != nil {
		tv.tabs[idx].content.Update(dt)
	}
}

func (tv *TabView) updateHorizontal(b rl.Rectangle, mouse rl.Vector2) {
	if len(tv.tabs) == 0 {
		return
	}
	tabW := b.Width / float32(len(tv.tabs))
	for i := range tv.tabs {
		r := rl.NewRectangle(b.X+float32(i)*tabW, b.Y, tabW, tabBarH)
		if rl.CheckCollisionPointRec(mouse, r) {
			tv.hovered = i
			if !tv.tabs[i].disabled && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				tv.Active.Set(i)
			}
			break
		}
	}
}

// handleKeyboard switches tabs via arrow keys and Ctrl+Tab when the pointer
// is over this TabView (Path B keyboard polish).
func (tv *TabView) handleKeyboard() {
	if len(tv.tabs) == 0 {
		return
	}
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	if ctrl && rl.IsKeyPressed(rl.KeyTab) {
		tv.stepActiveTab(1)
		return
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		tv.stepActiveTab(1)
		return
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		tv.stepActiveTab(-1)
		return
	}
	if rl.IsKeyPressed(rl.KeyHome) {
		tv.setActiveTab(0)
		return
	}
	if rl.IsKeyPressed(rl.KeyEnd) {
		tv.setActiveTab(len(tv.tabs) - 1)
	}
}

func (tv *TabView) setActiveTab(i int) {
	if i < 0 || i >= len(tv.tabs) || tv.tabs[i].disabled {
		return
	}
	tv.Active.Set(i)
}

func (tv *TabView) stepActiveTab(dir int) {
	n := len(tv.tabs)
	if n == 0 {
		return
	}
	cur := tv.Active.Get()
	for k := 0; k < n; k++ {
		cur += dir
		if cur < 0 {
			cur = n - 1
		}
		if cur >= n {
			cur = 0
		}
		if !tv.tabs[cur].disabled {
			tv.Active.Set(cur)
			return
		}
	}
}

func (tv *TabView) updateVertical(b rl.Rectangle, mouse rl.Vector2) {
	if len(tv.tabs) == 0 {
		return
	}
	tabH := b.Height / float32(len(tv.tabs))
	for i := range tv.tabs {
		r := rl.NewRectangle(b.X, b.Y+float32(i)*tabH, tabBarW, tabH)
		if rl.CheckCollisionPointRec(mouse, r) {
			tv.hovered = i
			if !tv.tabs[i].disabled && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				tv.Active.Set(i)
			}
			break
		}
	}
}

// ─── Layout ──────────────────────────────────────────────────────────────────

// Layout positions all content nodes in the area below/right of the tab bar.
// All tabs receive the same content bounds so switching is instant.
func (tv *TabView) Layout() {
	defer func() { tv.layoutDirty = false }()
	b := tv.Bounds()
	var contentRect rl.Rectangle
	if tv.Vertical {
		contentRect = rl.NewRectangle(b.X+tabBarW, b.Y, b.Width-tabBarW, b.Height)
	} else {
		contentRect = rl.NewRectangle(b.X, b.Y+tabBarH, b.Width, b.Height-tabBarH)
	}
	for _, t := range tv.tabs {
		if t.content != nil {
			t.content.SetBounds(contentRect)
			t.content.Layout()
		}
	}
}

// ─── Draw ────────────────────────────────────────────────────────────────────

// Draw renders the tab bar headers and the active tab's content node.
func (tv *TabView) Draw() {
	if tv.IsHidden() || len(tv.tabs) == 0 {
		return
	}
	if tv.Vertical {
		tv.drawVertical(true)
	} else {
		tv.drawHorizontal(true)
	}
}

// InteractionOverlayActive implements InteractionOverlayPainter.
// Tab header labels redraw in the SSAA cache so hover text stays crisp.
func (tv *TabView) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (tv *TabView) DrawInteractionOverlay() {}

// tabLabel returns "icon  title" or just "title" when the tab has no icon.
func (tv *TabView) tabLabel(t tabEntry) string {
	if t.icon == "" {
		return t.title
	}
	return t.icon + "  " + t.title
}

// tabStyle selects the appropriate style for a tab header.
func (tv *TabView) tabStyle(i, activeIdx int) Style {
	t := tv.tabs[i]
	switch {
	case t.disabled:
		return GetThemeStyle("tab-disabled")
	case i == activeIdx:
		return GetThemeStyle("tab-active")
	case tv.hovered == i:
		return GetThemeStyle("tab-hover")
	default:
		return GetThemeStyle("tab")
	}
}

func (tv *TabView) drawHorizontal(drawContent bool) {
	b := tv.Bounds()
	activeIdx := tv.Active.Get()
	tabW := b.Width / float32(len(tv.tabs))

	sTab := GetThemeStyle("tab")

	// Full-width tab bar background strip.
	rl.DrawRectangle(int32(b.X), int32(b.Y), int32(b.Width), int32(tabBarH),
		rl.NewColor(232, 234, 243, 255))

	for i, t := range tv.tabs {
		x := b.X + float32(i)*tabW
		tabRect := rl.NewRectangle(x, b.Y, tabW, tabBarH)
		isActive := i == activeIdx
		s := tv.tabStyle(i, activeIdx)

		rl.DrawRectangleRec(tabRect, s.BackgroundColor)

		// Right-edge divider between tabs (skip between active and neighbour).
		if i < len(tv.tabs)-1 && !isActive && tv.hovered != i+1 {
			rl.DrawRectangle(int32(x+tabW-1), int32(b.Y+4), 1,
				int32(tabBarH-8), sTab.BorderColor)
		}

		// Label centred in the tab header.
		label := tv.tabLabel(t)
		tw := measureTextS(label, s)
		tx := int32(x) + (int32(tabW)-tw)/2
		ty := int32(b.Y) + (int32(tabBarH)-s.FontSize)/2
		drawTextS(label, tx, ty, s)
	}

	// Bottom rule between tab bar and content; active accent sits flush on the seam.
	ruleY := int32(b.Y + tabBarH - 1)
	rl.DrawRectangle(int32(b.X), ruleY, int32(b.Width), 1, sTab.BorderColor)
	if activeIdx >= 0 && activeIdx < len(tv.tabs) {
		x := b.X + float32(activeIdx)*tabW
		s := tv.tabStyle(activeIdx, activeIdx)
		// Cover the gray rule under the active tab, then full-width accent on the seam.
		rl.DrawRectangle(int32(x), ruleY, int32(tabW), 1, s.BackgroundColor)
		rl.DrawRectangle(int32(x), ruleY-1, int32(tabW), 2, s.BorderColor)
	}

	if drawContent && activeIdx >= 0 && activeIdx < len(tv.tabs) && tv.tabs[activeIdx].content != nil {
		tv.tabs[activeIdx].content.Draw()
	}
}

func (tv *TabView) drawVertical(drawContent bool) {
	b := tv.Bounds()
	activeIdx := tv.Active.Get()
	tabH := b.Height / float32(len(tv.tabs))

	sTab := GetThemeStyle("tab")

	// Tab bar background strip.
	rl.DrawRectangle(int32(b.X), int32(b.Y), int32(tabBarW), int32(b.Height),
		rl.NewColor(232, 234, 243, 255))

	// Separator line right of tab bar.
	rl.DrawRectangle(int32(b.X+tabBarW-1), int32(b.Y), 1, int32(b.Height), sTab.BorderColor)

	for i, t := range tv.tabs {
		y := b.Y + float32(i)*tabH
		tabRect := rl.NewRectangle(b.X, y, tabBarW, tabH)
		isActive := i == activeIdx
		s := tv.tabStyle(i, activeIdx)

		rl.DrawRectangleRec(tabRect, s.BackgroundColor)

		// Horizontal divider between tabs (skip last).
		if i < len(tv.tabs)-1 {
			rl.DrawRectangle(int32(b.X+4), int32(y+tabH-1),
				int32(tabBarW-8), 1, sTab.BorderColor)
		}

		// Active tab — 2 px indigo right accent.
		if isActive {
			rl.DrawRectangle(int32(b.X+tabBarW-2), int32(y+2),
				2, int32(tabH-4), s.BorderColor)
		}

		// Label left-aligned with padding.
		label := tv.tabLabel(t)
		pad := int32(s.Padding)
		ty := int32(y) + (int32(tabH)-s.FontSize)/2
		drawTextS(label, int32(b.X)+pad, ty, s)
	}

	if drawContent && activeIdx >= 0 && activeIdx < len(tv.tabs) && tv.tabs[activeIdx].content != nil {
		tv.tabs[activeIdx].content.Draw()
	}
}
