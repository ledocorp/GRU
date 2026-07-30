// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TextInput is an editable single-line text field.
//
// # Focus
//
// TextInput processes key events only when focused. Focus is set externally by
// calling doc.SetFocus(textInput), which emits EventFocus. EventBlur is emitted
// when another widget is focused. Internally, TextInput registers On(EventFocus)
// and On(EventBlur) handlers to toggle the focused field.
//
// # Cursor & Horizontal Scroll
//
// A blinking cursor is drawn at the current character position. When typing or
// navigating causes the cursor to move outside the visible area, scrollOffset
// is adjusted so the cursor always stays in view. Text is clipped to the widget
// bounds via a scissor rect so it never overflows the box visually.
//
// # Keyboard Shortcuts
//
//	Backspace / Delete  — remove character before/after cursor
//	Left / Right arrows — move cursor one character
//	Home / End          — jump to start/end of text
//	Ctrl+A              — select all
//	Ctrl+C / Ctrl+X / Ctrl+V — copy, cut, paste
//	Shift+arrows        — extend selection
//	Right-click         — cut / copy / paste context menu
//
// Call [TextInput.IsFocused] from app code when you need to avoid overwriting
// the field from external signals while the user is typing.
//
// # LLM Prompt Template
//
//	name := ui.NewTextInput("name", "", 0, 0, 0, 40)
//	name.Placeholder = "Your name"
//	name.Text.Subscribe(func() { /* validate */ })
//	form.AddChild(name)
//
// Demo scenes: **Form Demo**, **Theme v2 Demo**, Notepad Find bar.
type TextInput struct {
	Element
	Text         *Signal[string] // Reactive text
	Placeholder  string
	OnEnter      func() // optional; Enter triggers (e.g. modal find/replace)
	Disabled     bool
	focused      bool
	hovered      bool
	lastHovered  bool
	blinkPhase   int
	cursor       int // Cursor position in text
	blinkTimer   float32
	scrollOffset int32 // Horizontal scroll offset for long text
	selAnchor    int  // Selection anchor; -1 means no selection
	dragSelect   bool // True while drag-selecting with the mouse
}

// NewTextInput creates a new TextInput with the given text.
// The widget defaults to the "input" theme key (same chrome family as Dropdown).
func NewTextInput(id, text string, x, y, w, h float32) *TextInput {
	ti := &TextInput{
		Element:   NewElement(id, x, y, w, h),
		Text:      NewSignal(text),
		cursor:    len(text),
		selAnchor: -1,
	}
	ti.styleName = "input"
	ti.Element.SetStyleVariant("input", "default")
	ti.ZIndex = 10 // Higher ZIndex for layering on top
	ti.Text.Subscribe(func() { ti.MarkDrawDirty() })
	ti.On(EventFocus, func(e Event) {
		ti.focused = !ti.Disabled
		if ti.focused {
			ti.selAnchor = -1
			ti.dragSelect = false
		}
	})
	ti.On(EventBlur, func(e Event) {
		ti.focused = false
		ti.selAnchor = -1
		ti.dragSelect = false
	})
	return ti
}

// Update implements Node.Update by handling keyboard input when focused.
func (ti *TextInput) Update(dt float32) {
	if ti.IsHidden() {
		return
	}
	ti.syncDocumentFocus()
	mouse := rl.GetMousePosition()
	ti.hovered = !ti.Disabled && rl.CheckCollisionPointRec(mouse, ti.Bounds())
	if ti.hovered != ti.lastHovered {
		ti.lastHovered = ti.hovered
		ti.MarkDrawDirty()
	}
	ti.blinkTimer += dt
	if ti.keyboardActive() {
		phase := int(ti.blinkTimer*2) % 2
		if phase != ti.blinkPhase {
			ti.blinkPhase = phase
			ti.MarkDrawDirty()
		}
	}

	if ti.keyboardActive() {
		NoteTypingGesture()
		// keyFired: initial press OR OS-driven key-repeat while held.
		keyFired := func(key int32) bool {
			return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key)
		}
		shift := textInputShiftDown()

		ti.updateSelectionShortcuts()

		// Handle text input
		if keyFired(rl.KeyBackspace) {
			if !ti.deleteSelection() && ti.cursor > 0 {
				text := ti.Text.Get()
				ti.Text.Set(text[:ti.cursor-1] + text[ti.cursor:])
				ti.cursor--
				ti.updateScroll()
				ti.MarkDrawDirty()
			}
		} else if keyFired(rl.KeyDelete) {
			if !ti.deleteSelection() && ti.cursor < len(ti.Text.Get()) {
				text := ti.Text.Get()
				ti.Text.Set(text[:ti.cursor] + text[ti.cursor+1:])
				ti.updateScroll()
				ti.MarkDrawDirty()
			}
		} else if keyFired(rl.KeyLeft) && ti.cursor > 0 {
			if shift {
				if ti.selAnchor < 0 {
					ti.selAnchor = ti.cursor
				}
			} else {
				ti.clearSelection()
			}
			ti.cursor--
			ti.updateScroll()
			ti.MarkDrawDirty()
		} else if keyFired(rl.KeyRight) && ti.cursor < len(ti.Text.Get()) {
			if shift {
				if ti.selAnchor < 0 {
					ti.selAnchor = ti.cursor
				}
			} else {
				ti.clearSelection()
			}
			ti.cursor++
			ti.updateScroll()
			ti.MarkDrawDirty()
		} else if keyFired(rl.KeyHome) {
			if shift {
				if ti.selAnchor < 0 {
					ti.selAnchor = ti.cursor
				}
			} else {
				ti.clearSelection()
			}
			ti.cursor = 0
			ti.updateScroll()
			ti.MarkDrawDirty()
		} else if keyFired(rl.KeyEnd) {
			if shift {
				if ti.selAnchor < 0 {
					ti.selAnchor = ti.cursor
				}
			} else {
				ti.clearSelection()
			}
			ti.cursor = len(ti.Text.Get())
			ti.updateScroll()
			ti.MarkDrawDirty()
		} else if keyFired(rl.KeyEnter) || keyFired(rl.KeyKpEnter) {
			if ti.OnEnter != nil {
				ti.OnEnter()
			}
		}

		// Handle character input
		char := rl.GetCharPressed()
		for char != 0 {
			if ti.hasSelection() {
				ti.deleteSelection()
			}
			text := ti.Text.Get()
			ti.Text.Set(text[:ti.cursor] + string(char) + text[ti.cursor:])
			ti.cursor++
			ti.updateScroll()
			ti.MarkDrawDirty()
			char = rl.GetCharPressed()
		}
	}

	if !ti.Disabled {
		state := StyleStateNone
		if ti.hovered {
			state |= StyleStateHover
		}
		if ti.focused {
			state |= StyleStateFocus
		}
		style, _ := ti.ResolveStyle(state)
		bounds := ti.Bounds()
		pad := int32(style.Padding + 0.5)
		if pad < 2 {
			pad = 2
		}
		ti.updateMouseSelection(rl.GetMousePosition(), style, bounds, pad)
	}
}

// updateScroll adjusts the horizontal scroll to keep the cursor visible.
func (ti *TextInput) updateScroll() {
	text := ti.Text.Get()
	style := ti.GetStyle()
	pad := int32(style.Padding + 0.5)
	if pad < 2 {
		pad = 2
	}

	cursorX := measureTextS(text[:ti.cursor], style)
	availableWidth := int32(ti.Bounds().Width) - 2*pad

	if cursorX-ti.scrollOffset > availableWidth-pad/2 {
		ti.scrollOffset = cursorX - availableWidth + pad/2
	} else if cursorX-ti.scrollOffset < pad/2 {
		ti.scrollOffset = cursorX - pad/2
		if ti.scrollOffset < 0 {
			ti.scrollOffset = 0
		}
	}
}

// Layout implements Node.Layout (no-op for leaf widgets).
func (ti *TextInput) Layout() { ti.layoutDirty = false }

// Draw implements Node.Draw by rendering the text input.
func (ti *TextInput) Draw() {
	ti.drawInternal() // Direct call, no caching delegation
}

// drawInternal performs the actual drawing.
func (ti *TextInput) drawInternal() {
	if ti.IsHidden() {
		return
	}

	state := StyleStateNone
	if ti.hovered {
		state |= StyleStateHover
	}
	if ti.Disabled {
		state |= StyleStateDisabled
	}
	style, _ := ti.ResolveStyle(state)
	chrome, _ := ti.ResolveStyle(state &^ (StyleStateFocus | StyleStateHover))
	layoutBounds := snapControlRect(ti.Bounds())
	pad := int32(style.Padding + 0.5)
	if pad < 2 {
		pad = 2
	}

	bw := chrome.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	drawTextInputChrome(layoutBounds, chrome, ti.keyboardActive())

	innerBounds := rl.NewRectangle(layoutBounds.X+bw, layoutBounds.Y+bw, layoutBounds.Width-2*bw, layoutBounds.Height-2*bw)

	// Text
	text := ti.Text.Get()
	posX := int32(layoutBounds.X) + pad - ti.scrollOffset
	posY := TextPosY(layoutBounds, style)
	showPlaceholder := textInputShowPlaceholder(text, ti.Placeholder)

	if showPlaceholder {
		phStyle := style
		phStyle.TextColor = textInputPlaceholderColor(style.TextColor)
		drawTextS(ti.Placeholder, posX, posY, phStyle)
	} else {
		clipScroll := ti.scrollOffset > 0 && innerBounds.Width > 0 && innerBounds.Height > 0
		if clipScroll {
			beginScissorMode(int32(innerBounds.X), int32(innerBounds.Y), int32(innerBounds.Width), int32(innerBounds.Height))
		}
		ti.drawSelectionHighlight(text, layoutBounds, pad, posY, style)
		if text != "" {
			drawTextS(text, posX, posY, style)
		}
		if clipScroll {
			endNestedScissor()
		}
	}

	if ti.keyboardActive() && int(ti.blinkTimer*2)%2 == 0 {
		cursorX := int32(layoutBounds.X) + pad + measureTextS(text[:ti.cursor], style) - ti.scrollOffset
		rl.DrawLine(cursorX, posY, cursorX, posY+int32(EffectiveFontSize(style)), focusRingIndigo)
	}
}

func textInputShowPlaceholder(text, placeholder string) bool {
	return placeholder != "" && strings.TrimSpace(text) == ""
}

// IsFocused reports whether this field currently has keyboard focus.
func (ti *TextInput) IsFocused() bool { return ti.keyboardActive() }

func (ti *TextInput) syncDocumentFocus() {
	if ti.Disabled {
		return
	}
	d := ActiveDocument()
	if d != nil && d.Focused == ti && !ti.focused {
		ti.focused = true
		ti.blinkTimer = 0
	}
}

func (ti *TextInput) keyboardActive() bool {
	if ti.Disabled {
		return false
	}
	if WebViewHostHoldsKeyboard() {
		return false
	}
	if ti.focused {
		return true
	}
	d := ActiveDocument()
	return d != nil && d.Focused == ti
}

// IsInteractive implements Node.IsInteractive.
func (ti *TextInput) IsInteractive() bool { return !ti.Disabled }

// UsesScissor implements Node.UsesScissor.
// TextInput clips its text content to the input bounds via BeginScissorMode.
func (ti *TextInput) UsesScissor() bool { return true }

// AnimationActive implements AnimationReporter (caret blink while focused).
func (ti *TextInput) AnimationActive() bool {
	if ti.IsHidden() || ti.Disabled || !ti.keyboardActive() {
		return false
	}
	if !rl.IsWindowFocused() {
		return false
	}
	return ti.blinkPhase == 0
}

// AnimationSource implements AnimationReporter.
func (ti *TextInput) AnimationSource() string { return ti.ID() }

// InteractionOverlayActive implements InteractionOverlayPainter.
// Hover/focus chrome is redrawn in the SSAA cache (MarkDrawDirty) — not at 1×
// after the blit, which made typed text vanish and look soft on hover.
func (ti *TextInput) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (ti *TextInput) DrawInteractionOverlay() {}

// textInputPlaceholderColor returns muted hint text for empty fields.
func textInputPlaceholderColor(_ rl.Color) rl.Color {
	muted := GetThemeStyle("modal-body").TextColor
	if muted.A == 0 {
		return rl.NewColor(136, 140, 152, 255)
	}
	return muted
}

// drawTextInputChrome paints a crisp grey inset border; focus accent is drawn underneath first.
func drawTextInputChrome(outer rl.Rectangle, chrome Style, focused bool) {
	if outer.Width < 1 || outer.Height < 1 {
		return
	}
	bg := chrome.BackgroundColor
	borderCol := chrome.BorderColor
	if borderCol.A == 0 {
		borderCol = rl.NewColor(190, 192, 200, 255)
	}
	bw := chrome.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	roundness := buttonCornerRoundness(outer.Width, outer.Height, chrome.CornerRadius)
	if focused {
		drawTextInputFocusAccent(outer, chrome.CornerRadius, focusRingIndigo)
	}
	drawRoundedInsetBorder(outer, roundness, bw, borderCol, bg)
}

// drawTextInputFocusAccent traces the bottom edge; draw before the grey border so it sits underneath.
func drawTextInputFocusAccent(outer rl.Rectangle, cornerPx float32, col rl.Color) {
	const stroke = float32(5)
	const behindPad = float32(2)
	cr := cornerPx
	if cr < 3 {
		cr = 3
	}
	if cr*2 >= outer.Width {
		cr = outer.Width * 0.18
	}
	bottom := outer.Y + outer.Height + behindPad
	y := bottom - stroke*0.5
	left := outer.X + cr
	right := outer.X + outer.Width - cr
	rl.DrawLineEx(rl.NewVector2(left, y), rl.NewVector2(right, y), stroke, col)

	innerR := cr - stroke
	if innerR < 1 {
		innerR = 1
	}
	seg := int32(8)
	blCx := outer.X + cr
	blCy := bottom - cr
	rl.DrawRing(rl.NewVector2(blCx, blCy), innerR, cr, 90, 180, seg, col)

	brCx := outer.X + outer.Width - cr
	brCy := bottom - cr
	rl.DrawRing(rl.NewVector2(brCx, brCy), innerR, cr, 0, 90, seg, col)
}
