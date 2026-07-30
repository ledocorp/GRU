// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─────────────────────────────────────────────────────────────────────────────
// Layout constants
// ─────────────────────────────────────────────────────────────────────────────

const (
	kbActionFontSize int32 = 18 // action label font size
	kbTokenFontSize  int32 = 16 // font size inside each key pill
	kbIconFontSize   int32 = 20 // leading symbol glyph font size

	kbTokenPadX float32 = 7  // horizontal padding inside each key pill
	kbTokenPadY float32 = 4  // vertical padding inside each key pill
	kbTokenGap  float32 = 4  // gap between adjacent key pills
	kbTokenSepW float32 = 14 // total width reserved for the "+" separator between pills
	kbActionGap float32 = 14 // gap from the action text to the first key pill
	kbIconGap   float32 = 8  // gap from the leading symbol to the action text

	kbDefaultH float32 = 36 // default widget height when h = 0
)

// ─────────────────────────────────────────────────────────────────────────────
// Palette (package-level, unexported)
// ─────────────────────────────────────────────────────────────────────────────

var (
	kbTokenBg     = rl.NewColor(243, 244, 246, 255) // key pill fill   #F3F4F6
	kbTokenBorder = rl.NewColor(209, 213, 219, 255) // key pill border #D1D5DB
	kbTokenText   = rl.NewColor(31, 41, 55, 255)    // key pill text   #1F2937
	kbSepColor    = rl.NewColor(156, 163, 175, 255) // "+" separator   #9CA3AF
	kbHoverTint   = rl.NewColor(79, 70, 229, 12)    // clickable row hover tint
	kbShadowColor = rl.NewColor(180, 185, 200, 80)  // key cap bottom shadow
)

// ─────────────────────────────────────────────────────────────────────────────
// KeyboardShortcut
// ─────────────────────────────────────────────────────────────────────────────

// KeyboardShortcut renders an action name alongside one or more key-cap pills
// that represent its keyboard shortcut. It is designed for inline use in
// toolbars, menus, command palettes, and settings panels.
//
// # Visual layout
//
//	[Symbol]  Action name      [Ctrl] + [Shift] + [S]
//
// Each modifier and key is drawn as an individual pill-shaped box with a
// subtle bottom shadow to evoke a physical key cap. Common modifier names are
// automatically replaced with canonical display labels:
//
//	"Ctrl"  -> "Ctrl"   "Shift"  -> "Shift"   "Alt"     -> "Alt"
//	"Cmd"   -> "Cmd"    "Win"    -> "Win"     "Tab"     -> "Tab"
//	"Enter" -> "Enter"  "Backspace" -> "Back"  "Escape" -> "Esc"
//	Arrow keys -> "Up" "Down" "Left" "Right"    F-keys -> "F1"..."F12"
//
// # Auto-sizing
//
// Passing w=0 causes the widget to compute its own width from the action label
// and key tokens. The width is recomputed whenever Action or Keys changes.
// Pass a fixed width to constrain the widget (required for Spread mode).
//
// # Spread mode
//
// When [KeyboardShortcut.Spread] is true the action text is placed at the left
// edge and the key tokens are pushed to the right edge of the bounding box.
// This reproduces the classic menu-row layout:
//
//	Open File…              [Ctrl] [O]
//
// Spread requires a fixed width; with w=0 the result is identical to
// the default inline layout.
//
// # Example — toolbar shortcut label (non-interactive):
//
//	ks := ui.NewKeyboardShortcut("ks-save", "Save", "Ctrl+S", 0, 0, 0, 28)
//	toolbar.AddChild(ks)
//
// # Example — menu row (Spread, fixed width):
//
//	ks := ui.NewKeyboardShortcut("ks-open", "Open File…", "Ctrl+O", 0, 0, 300, 28)
//	ks.Spread = true
//	menu.AddChild(ks)
//
// # Example — clickable shortcut (fires OnClick when pressed):
//
//	ks := ui.NewKeyboardShortcut("ks-undo", "Undo", "Ctrl+Z", 0, 0, 0, 32)
//	ks.OnClick = func() { doc.Undo() }
//	parent.AddChild(ks)
//
// # LLM Prompt Template
//
//	ks := ui.NewKeyboardShortcut("ks-save", "Save", "Ctrl+S", 0, 0, 300, 28)
//	ks.Spread = true
//	menu.AddChild(ks)
//
// Demo scenes: **Batch 5 KeyboardShortcut**.
type KeyboardShortcut struct {
	Element

	// Action is the human-readable action name shown to the left of the keys.
	// Reactive — call Action.Set to update at runtime.
	// Set to "" to display key tokens only.
	Action *Signal[string]

	// Keys is the raw key combination string, e.g. "Ctrl+S" or "Ctrl+Shift+Z".
	// Individual keys are separated by "+". Each token is automatically
	// canonicalised (see [canonicalKeyToken]).
	// Reactive — call Keys.Set to update at runtime.
	// Set to "" to display the action name only.
	Keys *Signal[string]

	// Symbol is an optional UTF-8 glyph drawn to the left of Action.
	// Accepts any printable Unicode character: "⌘", "🔍", "✎", etc.
	Symbol string

	// SymbolColor overrides the theme TextColor for the Symbol glyph only.
	// Leave at its zero value to use the theme colour.
	SymbolColor rl.Color

	// Spread, when true, places the action name flush-left and the key tokens
	// flush-right within the bounding box. Requires a fixed (non-zero) width.
	Spread bool

	// OnClick, when non-nil, enables row interactivity: the widget reports
	// IsInteractive() = true, a hover highlight is drawn, and OnClick fires on
	// every left-mouse-button press over the widget.
	OnClick func()

	// autoSize is true when w=0 was passed to NewKeyboardShortcut.
	autoSize bool

	// keyTokens holds the canonical key strings parsed from Keys.Get().
	keyTokens []string

	hovered bool
}

// NewKeyboardShortcut creates a KeyboardShortcut widget.
//
//	id     — unique node ID
//	action — human-readable action name (e.g. "Save"); may be ""
//	keys   — key combo string (e.g. "Ctrl+Shift+S"); may be ""
//	x, y   — position (overridden by parent layout)
//	w      — width; 0 = auto-size from content
//	h      — height; 0 = default 28 px
func NewKeyboardShortcut(id, action, keys string, x, y, w, h float32) *KeyboardShortcut {
	if h == 0 {
		h = kbDefaultH
	}
	auto := w == 0

	ks := &KeyboardShortcut{
		Element:   NewElement(id, x, y, w, h),
		Action:    NewSignal(action),
		Keys:      NewSignal(keys),
		autoSize:  auto,
		keyTokens: parseKeyTokens(keys),
	}
	ks.styleName = "keyboard-shortcut"

	ks.Action.Subscribe(func() {
		if ks.autoSize {
			ks.bounds.Width = ks.measureTotalWidth()
		}
		ks.MarkDrawDirty()
	})

	ks.Keys.Subscribe(func() {
		ks.keyTokens = parseKeyTokens(ks.Keys.Get())
		if ks.autoSize {
			ks.bounds.Width = ks.measureTotalWidth()
		}
		ks.MarkDrawDirty()
	})

	if auto {
		ks.bounds.Width = ks.measureTotalWidth()
	}

	return ks
}

// ─────────────────────────────────────────────────────────────────────────────
// Public API
// ─────────────────────────────────────────────────────────────────────────────

// SetAction is a convenience wrapper for Action.Set(v).
func (ks *KeyboardShortcut) SetAction(v string) { ks.Action.Set(v) }

// SetKeys is a convenience wrapper for Keys.Set(v).
func (ks *KeyboardShortcut) SetKeys(v string) { ks.Keys.Set(v) }

// ─────────────────────────────────────────────────────────────────────────────
// Node interface
// ─────────────────────────────────────────────────────────────────────────────

// IsInteractive returns true when OnClick is set.
func (ks *KeyboardShortcut) IsInteractive() bool { return ks.OnClick != nil }

// UsesScissor returns false — KeyboardShortcut never modifies the scissor rect.
func (ks *KeyboardShortcut) UsesScissor() bool { return false }

// Layout is a no-op for this leaf widget.
func (ks *KeyboardShortcut) Layout() { ks.layoutDirty = false }

// Update handles hover detection and optional click.
func (ks *KeyboardShortcut) Update(_ float32) {
	if ks.IsHidden() || ks.OnClick == nil {
		if ks.hovered {
			ks.hovered = false
			ks.MarkDrawDirty()
		}
		return
	}
	mouse := rl.GetMousePosition()
	prev := ks.hovered
	ks.hovered = rl.CheckCollisionPointRec(mouse, ks.bounds)
	if ks.hovered != prev {
		ks.MarkDrawDirty()
	}
	if ks.hovered && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		ks.OnClick()
	}
}

// Draw renders the widget.
func (ks *KeyboardShortcut) Draw() {
	if ks.IsHidden() {
		return
	}
	ks.drawInternal()
	ks.drawDirty = false
}

// ─────────────────────────────────────────────────────────────────────────────
// Drawing
// ─────────────────────────────────────────────────────────────────────────────

func (ks *KeyboardShortcut) drawInternal() {
	b := ks.bounds
	style := ks.GetStyle()

	// Optional hover highlight (only when clickable).
	if ks.hovered {
		rl.DrawRectangleRounded(b, 0.4, 6, kbHoverTint)
	}

	midY := b.Y + b.Height/2

	// ── Leading symbol glyph ─────────────────────────────────────────────────
	curX := b.X
	if ks.Symbol != "" {
		symColor := style.TextColor
		if ks.SymbolColor.A != 0 {
			symColor = ks.SymbolColor
		}
		symW := float32(measureText(ks.Symbol, kbIconFontSize))
		iconFS := float32(kbIconFontSize) * GlobalFontScale
		drawTextF(ks.Symbol, curX, midY-iconFS/2, iconFS, symColor, false, false, false, false)
		curX += symW + kbIconGap
	}

	// ── Action name ──────────────────────────────────────────────────────────
	actionText := ks.Action.Get()
	hasTokens := len(ks.keyTokens) > 0

	if actionText != "" {
		if ks.Spread && hasTokens {
			// In spread mode just draw the action at the current left edge;
			// tokens will be positioned from the right edge below.
			actionFS := float32(kbActionFontSize) * GlobalFontScale
			drawTextF(actionText, curX, midY-actionFS/2, actionFS, style.TextColor, false, false, false, false)
			// curX is not advanced; token positioning overrides it below.
		} else {
			actionFS := float32(kbActionFontSize) * GlobalFontScale
			drawTextF(actionText, curX, midY-actionFS/2, actionFS, style.TextColor, false, false, false, false)
			curX += float32(measureText(actionText, kbActionFontSize))
			if hasTokens {
				curX += kbActionGap
			}
		}
	}

	if !hasTokens {
		return
	}

	// ── Token start X (spread vs inline) ─────────────────────────────────────
	if ks.Spread {
		tokW := ks.measureTokensWidth()
		curX = b.X + b.Width - tokW
	}

	// ── Key token pills ───────────────────────────────────────────────────────
	tokenFS := float32(kbTokenFontSize) * GlobalFontScale
	tokenH := tokenFS + kbTokenPadY*2
	tokenY := midY - tokenH/2

	for i, tok := range ks.keyTokens {
		// "+" separator between pills.
		if i > 0 {
			sepTextW := float32(measureText("+", kbTokenFontSize))
			sepX := curX + (kbTokenSepW-sepTextW)/2
			drawTextF("+", sepX, midY-tokenFS/2, tokenFS, kbSepColor, false, false, false, false)
			curX += kbTokenSepW
		}

		tokTextW := float32(measureText(tok, kbTokenFontSize))
		pillW := tokTextW + kbTokenPadX*2
		pillRect := rl.NewRectangle(curX, tokenY, pillW, tokenH)

		// Background fill.
		rl.DrawRectangleRounded(pillRect, 0.28, 4, kbTokenBg)
		// Border.
		rl.DrawRectangleRoundedLinesEx(pillRect, 0.28, 4, 1.0, kbTokenBorder)
		// Bottom shadow — gives a subtle key-cap depth effect.
		shadowRect := rl.NewRectangle(curX+1, tokenY+tokenH, pillW-2, 2)
		rl.DrawRectangleRounded(shadowRect, 0.5, 4, kbShadowColor)
		// Key text.
		drawTextF(tok, curX+kbTokenPadX, tokenY+kbTokenPadY, tokenFS, kbTokenText, false, false, false, false)

		curX += pillW + kbTokenGap
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Width measurement
// ─────────────────────────────────────────────────────────────────────────────

// measureTotalWidth returns the minimum pixel width for the current content
// in the default inline layout (Spread is not considered here).
func (ks *KeyboardShortcut) measureTotalWidth() float32 {
	w := float32(0)

	if ks.Symbol != "" {
		w += float32(measureText(ks.Symbol, kbIconFontSize)) + kbIconGap
	}

	action := ks.Action.Get()
	if action != "" {
		w += float32(measureText(action, kbActionFontSize))
		if len(ks.keyTokens) > 0 {
			w += kbActionGap
		}
	}

	w += ks.measureTokensWidth()

	if w < 10 {
		w = 10
	}
	return w
}

// measureTokensWidth returns the pixel width consumed by all key token pills
// including inter-token separators, but excluding the trailing gap after the
// last pill.
func (ks *KeyboardShortcut) measureTokensWidth() float32 {
	if len(ks.keyTokens) == 0 {
		return 0
	}
	w := float32(0)
	for i, tok := range ks.keyTokens {
		if i > 0 {
			w += kbTokenSepW
		}
		w += float32(measureText(tok, kbTokenFontSize)) + kbTokenPadX*2
	}
	// kbTokenGap is added between pills during drawing but the last pill has no
	// trailing gap, so do not add it to the measurement.
	return w
}

// ─────────────────────────────────────────────────────────────────────────────
// Key parsing
// ─────────────────────────────────────────────────────────────────────────────

// parseKeyTokens splits a "+" delimited string into canonical key tokens.
// For example "ctrl+shift+s" -> ["Ctrl", "Shift", "S"].
// An empty string returns nil.
func parseKeyTokens(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, "+")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, canonicalKeyToken(p))
		}
	}
	return out
}

// canonicalKeyToken maps a raw key name (case-insensitive) to its display
// form. Well-known names use ASCII-safe labels so demos do not render missing
// glyph placeholders on fonts without symbol coverage. Single letters are
// upper-cased; everything else is title-cased.
func canonicalKeyToken(s string) string {
	switch strings.ToLower(s) {
	case "ctrl", "control":
		return "Ctrl"
	case "shift":
		return "Shift"
	case "alt", "option", "opt":
		return "Alt"
	case "cmd", "command", "meta":
		return "Cmd"
	case "win", "windows", "super":
		return "Win"
	case "tab":
		return "Tab"
	case "enter", "return":
		return "Enter"
	case "backspace":
		return "Back"
	case "delete", "del":
		return "Del"
	case "escape", "esc":
		return "Esc"
	case "space":
		return "Space"
	case "up", "arrowup":
		return "Up"
	case "down", "arrowdown":
		return "Down"
	case "left", "arrowleft":
		return "Left"
	case "right", "arrowright":
		return "Right"
	case "pageup", "pgup":
		return "PgUp"
	case "pagedown", "pgdn", "pgdown":
		return "PgDn"
	case "home":
		return "Home"
	case "end":
		return "End"
	case "insert", "ins":
		return "Ins"
	case "printscreen", "prtscn", "prtsc":
		return "PrtSc"
	case "f1":
		return "F1"
	case "f2":
		return "F2"
	case "f3":
		return "F3"
	case "f4":
		return "F4"
	case "f5":
		return "F5"
	case "f6":
		return "F6"
	case "f7":
		return "F7"
	case "f8":
		return "F8"
	case "f9":
		return "F9"
	case "f10":
		return "F10"
	case "f11":
		return "F11"
	case "f12":
		return "F12"
	default:
		// Single character: upper-case.
		if len(s) == 1 {
			return strings.ToUpper(s)
		}
		// Multi-char: title-case the first ASCII byte; rest lower.
		lower := strings.ToLower(s)
		return strings.ToUpper(lower[:1]) + lower[1:]
	}
}
