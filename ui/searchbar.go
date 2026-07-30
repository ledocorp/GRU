// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ─── SearchBar ────────────────────────────────────────────────────────────────
//
// SearchBar is an interactive single-line search field with a leading icon,
// placeholder text, a clear button, and a debounced Query output signal.
//
// # Visual Anatomy
//
// The widget is laid out as three horizontal zones inside a pill-shaped box:
//
//	┌──────────────────────────────────────────────────┐
//	│  ⊙  text content ……………………………………………………  ×  │
//	└──────────────────────────────────────────────────┘
//
//   - Left  zone (30 px) — Remix search icon (PhosphorMagnifyingGlass); hidden when ShowIcon = false.
//   - Centre zone        — editable text with cursor, horizontal scroll,
//                          placeholder hint, and scissor clipping.
//   - Right zone (28 px) — clear (×) button; visible only when text is non-empty.
//
// # Signals
//
// Two output signals are exposed:
//
//   - Text  *Signal[string] — raw text; fires on every keystroke.
//                             Use for lightweight live feedback (character counts,
//                             input echo, live filtering of tiny sets).
//
//   - Query *Signal[string] — debounced search query; fires only after the user
//                             pauses for DebounceDelay seconds (default 0.3 s).
//                             Any pending debounce is flushed immediately on blur
//                             or Escape so the consumer always has the latest value.
//                             Set DebounceDelay = 0 for immediate (Text == Query).
//                             Bind expensive operations (search, filter) here.
//
// # Focus
//
// Focus must be granted externally, exactly like TextInput:
//
//	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
//	    if rl.CheckCollisionPointRec(mouse, sb.Bounds()) {
//	        doc.SetFocus(sb)
//	    }
//	}
//
// The widget registers EventFocus / EventBlur handlers internally.
//
// # Keyboard Shortcuts
//
//	Backspace / Delete  — remove character before / after cursor
//	Left / Right arrows — move cursor one character
//	Home / End          — jump to start / end of text
//	Escape              — clear text and fire empty Query immediately
//
// # Usage
//
//	sb := ui.NewSearchBar("files-search", "Search files…", 0, 0, 320, 36)
//	sb.Query.Subscribe(func() {
//	    results = filter(allFiles, sb.Query.Get())
//	})
//	panel.AddChild(sb)
//
// See also: SetTooltip to attach hover help, doc.SetFocus for focus routing.

// sbIconW is the width of the leading search-icon zone (px).
const sbIconW = float32(30)

// sbClearW is the width of the trailing clear-button zone (px).
const sbClearW = float32(28)

// sbPadX is the horizontal text padding when ShowIcon is false.
const sbPadX = float32(10)

// sbFocusPad is layout space reserved inside bounds when focused so the 2px
// focus ring is not clipped by parent Panel body scissors.
const sbFocusPad = float32(2)

// sbClearEdgePad keeps the clear-button hover wash inside Panel scissor.
const sbClearEdgePad = float32(3)

// SearchBar is an interactive single-line search field.
//
// # LLM Prompt Template
//
//	sb := ui.NewSearchBar("search", "Search files…", 0, 0, 0, 38)
//	sb.Query.Subscribe(func() { results = filter(all, sb.Query.Get()) })
//	panel.AddChild(sb)
//
// Demo scenes: **Batch 2 SearchBar**.
type SearchBar struct {
	Element

	// Text is the raw text signal — fires on every keystroke.
	// Use this for lightweight live feedback (character count, input echo).
	Text *Signal[string]

	// Query is the debounced search query signal — fires after DebounceDelay
	// seconds of typing inactivity.  Bind expensive operations here.
	Query *Signal[string]

	// Placeholder is the grey hint text shown when the field is empty.
	Placeholder string

	// ShowIcon controls whether the magnifying-glass icon is rendered in the
	// left zone.  Default: true.
	ShowIcon bool

	// DebounceDelay is the seconds of inactivity after the last keystroke
	// before Query fires.  Set to 0 for immediate mode (Query == Text).
	// Default: 0.3 s.
	DebounceDelay float32

	// ── editing state ────────────────────────────────────────────────────────
	text         string
	focused      bool
	cursor       int
	blinkTimer   float32
	scrollOffset int32

	// ── debounce state ───────────────────────────────────────────────────────
	debounceTimer float32 // counts down from DebounceDelay after each keystroke
	lastFired     string  // last text value pushed to Query

	// ── hover state ──────────────────────────────────────────────────────────
	clearHovered bool
}

// NewSearchBar creates a SearchBar with the given placeholder, position, and
// size.  Recommended height: 34–40 px.
//
// The widget uses the "searchbar" theme style key; call SetStyle to override.
func NewSearchBar(id, placeholder string, x, y, w, h float32) *SearchBar {
	sb := &SearchBar{
		Element:       NewElement(id, x, y, w, h),
		Text:          NewSignal(""),
		Query:         NewSignal(""),
		Placeholder:   placeholder,
		ShowIcon:      true,
		DebounceDelay: 0.3,
	}
	sb.ZIndex = 10
	sb.styleName = "searchbar"
	sb.On(EventFocus, func(_ Event) {
		sb.focused = true
		sb.MarkDrawDirty()
	})
	sb.On(EventBlur, func(_ Event) {
		sb.focused = false
		sb.blinkTimer = 0
		sb.fireQuery() // flush pending debounce on blur
		sb.MarkDrawDirty()
	})
	return sb
}

// ─── Public API ───────────────────────────────────────────────────────────────

// Clear empties the search field and fires an empty Query immediately.
// Equivalent to pressing Escape while focused.
func (sb *SearchBar) Clear() {
	sb.text = ""
	sb.cursor = 0
	sb.scrollOffset = 0
	sb.debounceTimer = 0
	sb.Text.Set("")
	sb.fireQuery()
	sb.MarkDrawDirty()
}

// SetText replaces the search text programmatically, moves the cursor to the
// end, and schedules a debounced Query notification.
func (sb *SearchBar) SetText(t string) {
	sb.text = t
	sb.cursor = len(t)
	sb.updateScroll()
	sb.Text.Set(t)
	sb.scheduleDebounce()
	sb.MarkDrawDirty()
}

// GetText returns the current raw text buffer.
// This reflects every keystroke before the debounce timer fires.
func (sb *SearchBar) GetText() string { return sb.text }

// DebounceRemaining returns the seconds until the next Query notification.
// Returns 0 when no keystroke is pending.  Used by the Inspector.
func (sb *SearchBar) DebounceRemaining() float32 { return sb.debounceTimer }

// ─── Internal helpers ─────────────────────────────────────────────────────────

// fireQuery pushes the current text to Query if it has changed since the last
// notification.  Always safe to call; no-ops when text is unchanged.
func (sb *SearchBar) fireQuery() {
	if sb.text != sb.lastFired {
		sb.lastFired = sb.text
		sb.Query.Set(sb.text)
	}
}

// scheduleDebounce resets the debounce countdown.  When DebounceDelay is 0
// the Query fires immediately.
func (sb *SearchBar) scheduleDebounce() {
	if sb.DebounceDelay <= 0 {
		sb.fireQuery()
	} else {
		sb.debounceTimer = sb.DebounceDelay
	}
}

// paintBounds returns the rectangle used for painting and hit zones. Insets when
// focused (focus ring) and when non-empty (clear hover) so drawing stays inside
// Panel scissor.
func (sb *SearchBar) paintBounds() rl.Rectangle {
	b := sb.bounds
	l, t, r, bot := float32(0), float32(0), float32(0), float32(0)
	if sb.focused {
		l, t, r, bot = sbFocusPad, sbFocusPad, sbFocusPad, sbFocusPad
	}
	if sb.text != "" && r < sbClearEdgePad {
		r = sbClearEdgePad
	}
	if l == 0 && t == 0 && r == 0 && bot == 0 {
		return b
	}
	if b.Width <= l+r || b.Height <= t+bot {
		return b
	}
	return rl.NewRectangle(b.X+l, b.Y+t, b.Width-l-r, b.Height-t-bot)
}

// textLeft returns the X coordinate where the editable text zone begins.
func (sb *SearchBar) textLeft() float32 {
	b := sb.paintBounds()
	if sb.ShowIcon {
		return b.X + sbIconW
	}
	return b.X + sbPadX
}

// textRight returns the X coordinate where the editable text zone ends.
// The clear-button zone is reserved only when the field is non-empty.
func (sb *SearchBar) textRight() float32 {
	b := sb.paintBounds()
	if sb.text != "" {
		return b.X + b.Width - sbClearW
	}
	return b.X + b.Width - sbPadX
}

// clearBtnRect returns the bounding rectangle of the × clear button.
func (sb *SearchBar) clearBtnRect() rl.Rectangle {
	b := sb.paintBounds()
	return rl.NewRectangle(b.X+b.Width-sbClearW, b.Y, sbClearW, b.Height)
}

// updateScroll adjusts the horizontal scroll offset to keep the cursor visible.
func (sb *SearchBar) updateScroll() {
	style := sb.GetStyle()
	textL := sb.textLeft()
	textR := sb.textRight()
	availW := int32(textR-textL) - 4 // 2 px margin each side

	cursorX := measureTextS(sb.text[:sb.cursor], style)
	if cursorX-sb.scrollOffset > availW-4 {
		sb.scrollOffset = cursorX - availW + 4
	} else if cursorX-sb.scrollOffset < 4 {
		sb.scrollOffset = cursorX - 4
		if sb.scrollOffset < 0 {
			sb.scrollOffset = 0
		}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

// Update handles cursor blink, hover tracking, clear-button clicks, debounce
// countdown, and keyboard input (when focused).
func (sb *SearchBar) Update(dt float32) {
	if sb.IsHidden() {
		return
	}

	// ── Cursor blink ──────────────────────────────────────────────────────────
	prevPhase := int(sb.blinkTimer*2) % 2
	sb.blinkTimer += dt
	if sb.focused && int(sb.blinkTimer*2)%2 != prevPhase {
		sb.MarkDrawDirty()
	}

	// ── Debounce countdown ────────────────────────────────────────────────────
	if sb.debounceTimer > 0 {
		sb.debounceTimer -= dt
		if sb.debounceTimer <= 0 {
			sb.debounceTimer = 0
			sb.fireQuery()
		}
	}

	mouse := rl.GetMousePosition()

	// ── Clear button hover ────────────────────────────────────────────────────
	prevClearHov := sb.clearHovered
	if sb.text != "" {
		sb.clearHovered = rl.CheckCollisionPointRec(mouse, sb.clearBtnRect())
	} else {
		sb.clearHovered = false
	}
	if sb.clearHovered != prevClearHov {
		sb.MarkDrawDirty()
	}

	// ── Clear button click ────────────────────────────────────────────────────
	if sb.clearHovered && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		sb.Clear()
		return
	}

	// ── Keyboard input (focused only) ─────────────────────────────────────────
	if !sb.focused {
		return
	}
	NoteTypingGesture()

	if rl.IsKeyPressed(rl.KeyEscape) {
		sb.Clear()
		return
	}

	// keyFired: initial press OR OS-driven key-repeat while held.
	keyFired := func(key int32) bool {
		return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key)
	}

	changed := false

	if keyFired(rl.KeyBackspace) && sb.cursor > 0 {
		sb.text = sb.text[:sb.cursor-1] + sb.text[sb.cursor:]
		sb.cursor--
		sb.updateScroll()
		changed = true
	} else if keyFired(rl.KeyDelete) && sb.cursor < len(sb.text) {
		sb.text = sb.text[:sb.cursor] + sb.text[sb.cursor+1:]
		sb.updateScroll()
		changed = true
	} else if keyFired(rl.KeyLeft) && sb.cursor > 0 {
		sb.cursor--
		sb.updateScroll()
		sb.MarkDrawDirty()
	} else if keyFired(rl.KeyRight) && sb.cursor < len(sb.text) {
		sb.cursor++
		sb.updateScroll()
		sb.MarkDrawDirty()
	} else if keyFired(rl.KeyHome) {
		sb.cursor = 0
		sb.updateScroll()
		sb.MarkDrawDirty()
	} else if keyFired(rl.KeyEnd) {
		sb.cursor = len(sb.text)
		sb.updateScroll()
		sb.MarkDrawDirty()
	}

	// Character input — drain the character queue.
	char := rl.GetCharPressed()
	for char != 0 {
		sb.text = sb.text[:sb.cursor] + string(char) + sb.text[sb.cursor:]
		sb.cursor++
		sb.updateScroll()
		changed = true
		char = rl.GetCharPressed()
	}

	if changed {
		sb.Text.Set(sb.text)
		sb.scheduleDebounce()
		sb.MarkDrawDirty()
	}
}

// ─── Layout ───────────────────────────────────────────────────────────────────

// Layout is a no-op; SearchBar is a leaf widget with no children.
func (sb *SearchBar) Layout() { sb.layoutDirty = false }

// ─── Draw ─────────────────────────────────────────────────────────────────────

// Draw renders the SearchBar.
func (sb *SearchBar) Draw() {
	sb.drawInternal()
	sb.drawDirty = false
}

func (sb *SearchBar) drawInternal() {
	if sb.IsHidden() {
		return
	}

	style := sb.GetStyle()
	bounds := sb.paintBounds()

	// ── Compute roundness from CornerRadius ───────────────────────────────────
	r := style.CornerRadius
	roundness := float32(0)
	if r > 0 && bounds.Height > 0 {
		halfH := bounds.Height / 2
		roundness = r / halfH
		if roundness > 1 {
			roundness = 1
		}
	}

	bw := style.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	borderCol := style.BorderColor
	if sb.focused {
		bw = 2
		borderCol = focusRingIndigo
	}
	drawRoundedInsetBorder(bounds, roundness, bw, borderCol, style.BackgroundColor)

	// ── Search icon ───────────────────────────────────────────────────────────
	if sb.ShowIcon {
		sb.drawSearchIcon(style)
	}

	// ── Text content (scissor-clipped) ────────────────────────────────────────
	textL := sb.textLeft()
	textR := sb.textRight()
	innerBounds := rl.NewRectangle(textL, bounds.Y+1, textR-textL, bounds.Height-2)
	if vp := findViewport(sb); vp != nil {
		innerBounds = intersectRects(innerBounds, vp.ClipBounds())
		if innerBounds.Width <= 0 || innerBounds.Height <= 0 {
			return
		}
	}
	beginScissorMode(
		int32(innerBounds.X), int32(innerBounds.Y),
		int32(innerBounds.Width), int32(innerBounds.Height),
	)

	posX := int32(textL) + 2 - sb.scrollOffset
	posY := TextPosY(bounds, style)

	if sb.text == "" && sb.Placeholder != "" {
		// Muted placeholder — 38% opacity of the text colour.
		phStyle := style
		tc := style.TextColor
		phStyle.TextColor = rl.NewColor(tc.R, tc.G, tc.B, uint8(float32(tc.A)*0.38))
		drawTextS(sb.Placeholder, posX, posY, phStyle)
	} else if sb.text != "" {
		drawTextS(sb.text, posX, posY, style)
	}

	// Blinking cursor — every other half-second when focused.
	if sb.focused && int(sb.blinkTimer*2)%2 == 0 {
		cxOff := measureTextS(sb.text[:sb.cursor], style)
		cx := int32(textL) + 2 + cxOff - sb.scrollOffset
		rl.DrawLine(cx, posY, cx, posY+int32(EffectiveFontSize(style)), focusRingIndigo)
	}

	rl.EndScissorMode()

	// ── Clear button (drawn outside scissor) ─────────────────────────────────
	if sb.text != "" {
		sb.drawClearBtn(style)
	}
}

// drawSearchIcon renders the Remix search glyph in the left zone.
func (sb *SearchBar) drawSearchIcon(style Style) {
	b := sb.paintBounds()
	iconSz := b.Height * 0.46
	maxSz := sbIconW - 8
	if iconSz > maxSz {
		iconSz = maxSz
	}
	if iconSz < 12 {
		iconSz = 12
	}
	dst := rl.NewRectangle(
		b.X+(sbIconW-iconSz)/2,
		b.Y+(b.Height-iconSz)/2,
		iconSz, iconSz,
	)
	col := style.TextColor
	col.A = uint8(float32(col.A) * 0.52)
	DrawPhosphorIcon(dst, PhosphorMagnifyingGlass, PhosphorRegular, col)
}

// drawClearBtn renders the × dismiss button in the right zone.
func (sb *SearchBar) drawClearBtn(style Style) {
	cr := sb.clearBtnRect()
	cx := cr.X + cr.Width*0.50
	cy := cr.Y + cr.Height*0.50
	r := cr.Height * 0.175

	col := style.TextColor
	if sb.clearHovered {
		// Rect hover stays inside clearBtnRect (circle was clipped at panel edge).
		pad := float32(3)
		hr := rl.NewRectangle(cr.X+pad, cr.Y+pad, cr.Width-2*pad, cr.Height-2*pad)
		if hr.Width > 0 && hr.Height > 0 {
			rl.DrawRectangleRec(hr, listRowHoverBg)
		}
	} else {
		col.A = uint8(float32(col.A) * 0.50)
	}
	rl.DrawLineEx(rl.NewVector2(cx-r, cy-r), rl.NewVector2(cx+r, cy+r), 1.5, col)
	rl.DrawLineEx(rl.NewVector2(cx+r, cy-r), rl.NewVector2(cx-r, cy+r), 1.5, col)
}

// ─── Node interface ───────────────────────────────────────────────────────────

// IsInteractive returns true — SearchBar handles mouse and keyboard input.
func (sb *SearchBar) IsInteractive() bool { return true }

// IsFocused reports whether the search field owns keyboard input.
func (sb *SearchBar) IsFocused() bool { return sb.focused }

// UsesScissor returns true — SearchBar clips its text via BeginScissorMode.
func (sb *SearchBar) UsesScissor() bool { return true }
