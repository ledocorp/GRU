// Package ui (continued) — Button and IconButton.
//
// See node.go for the full package documentation.
//
// Demo scenes: **Batch 3 Live Demo** (counter +/-), **Widgets Demo**, **Form Demo**,
// **Demo Directory** nav bar, Notepad ribbon cells (via IconButton).
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// Button is an interactive widget that displays text and responds to clicks.
//
// Text is a reactive Signal so the displayed label updates automatically when
// Text.Set is called. The optional Scale field supports scale animations
// (e.g. a brief grow-shrink tween on click). Scale uses the same geometry as
// IconButton — drive both with the same tween for an identical bounce (polish §2.4).
//
// OnClick is called synchronously on the frame the mouse button is pressed.
//
// # LLM Prompt Template
//
//	btn := ui.NewButton("save", "Save", 0, 0, 0, 36)
//	btn.OnClick = func() { /* persist */ }
//	row.AddChild(btn)
//
// See examples/batch3_demo.go (Live Demo counter) and examples/widgets_demo.go.
type Button struct {
	Element
	Text    *Signal[string] // Reactive text
	OnClick func()          // Callback for click events
	// ToggleBinding, when set with style toolbar-toggle-label, tints text by on/off state.
	ToggleBinding *Signal[bool]
	hovered bool            // Internal hover state
	pressed bool            // True while mouse is held down over the button
	Scale   float32         // For animation
}

// NewButton creates a new Button with the given text.
func NewButton(id, text string, x, y, w, h float32) *Button {
	b := &Button{
		Element: NewElement(id, x, y, w, h),
		Text:    NewSignal(text),
		Scale:   1.0,
	}
	if w > 0 {
		b.PreferredWidth = w
	}
	// h=0 → intrinsic height (same contract as Label/Container), otherwise
	// flex leaves a 0-tall hit target and clicks never fire.
	if h == 0 {
		b.AutoHeight = true
	}
	b.Text.Subscribe(func() { b.MarkDirty() })
	return b
}

// Update implements Node.Update by handling mouse input.
func (b *Button) Update(dt float32) {
	if b.IsHidden() {
		return
	}

	mouse := rl.GetMousePosition()
	prevHovered := b.hovered
	prevPressed := b.pressed
	b.hovered = rl.CheckCollisionPointRec(mouse, b.Bounds())
	b.pressed = b.hovered && rl.IsMouseButtonDown(rl.MouseLeftButton)

	if b.hovered != prevHovered || b.pressed != prevPressed || b.Scale != 1.0 {
		b.MarkDrawDirty()
	}

	if b.hovered && rl.IsMouseButtonPressed(rl.MouseLeftButton) && b.OnClick != nil {
		b.OnClick()
	}
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (b *Button) ClearOverlayPointerState() {
	if !b.hovered && !b.pressed {
		return
	}
	b.hovered = false
	b.pressed = false
	b.MarkDrawDirty()
}

// buttonIntrinsicHeight is the minimum layout height for AutoHeight buttons so
// theme padding, border, and SDF text are not clipped by card/viewport body scissors.
func buttonIntrinsicHeight(style Style) float32 {
	fs := EffectiveFontSize(style)
	pad := style.Padding
	if pad <= 0 {
		pad = 10
	}
	h := fs + 2*pad + 2*style.BorderWidth + 2
	if h < 32 {
		return 32
	}
	return h
}

// toolbarBtnInnerRect is the label band inside snapped toolbar-btn chrome (border + padding).
func toolbarBtnInnerRect(snap rl.Rectangle, style Style) rl.Rectangle {
	p := toolbarStylePadding(style)
	bw := style.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	inset := p + bw
	if snap.Width <= 2*inset || snap.Height <= 2*inset {
		return snap
	}
	return rl.NewRectangle(snap.X+inset, snap.Y+inset, snap.Width-2*inset, snap.Height-2*inset)
}

// buttonNaturalWidth is the shrink-wrapped width for label + padding.
func buttonNaturalWidth(text string, style Style) float32 {
	pad := style.Padding
	if pad <= 0 {
		pad = 10
	}
	extra := float32(4)
	w := float32(measureTextS(text, style)) + 2*pad + 2*style.BorderWidth + extra
	if w < 48 {
		return 48
	}
	return w
}

// buttonToolbarNaturalWidth matches toolbarBtnInnerRect horizontal insets (no centering slack).
func buttonToolbarNaturalWidth(text string, style Style) float32 {
	p := toolbarStylePadding(style)
	bw := style.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	w := float32(measureTextS(text, style)) + 2*(p+bw)
	if w < 48 {
		return 48
	}
	return w
}

// GetPreferredWidth implements flex width hints. Buttons with w=0 in a flex
// column shrink-wrap to their label instead of stretching full panel width.
func (b *Button) GetPreferredWidth() float32 {
	if b.PreferredWidth > 0 {
		return b.PreferredWidth
	}
	return buttonNaturalWidth(b.Text.Get(), b.GetStyle())
}

// Layout implements Node.Layout.
func (b *Button) Layout() {
	if b.IsAutoHeight() {
		bounds := b.Bounds()
		want := buttonIntrinsicHeight(b.GetStyle())
		if bounds.Height < want-0.5 {
			bounds.Height = want
			b.setBoundsNoMark(bounds)
		}
	}
	b.layoutDirty = false
}

// Draw implements Node.Draw by rendering the button and text.
func (b *Button) Draw() {
	b.drawInternal() // Direct call, no caching delegation
}

// drawInternal performs the actual drawing.
func (b *Button) drawInternal() {
	if b.IsHidden() {
		return
	}

	style, stateApplied := b.ResolveStyle(buttonStyleState(b.hovered, b.pressed))

	// Apply scale (click animation)
	scaledW := b.Bounds().Width * b.Scale
	scaledH := b.Bounds().Height * b.Scale
	offsetX := (scaledW - b.Bounds().Width) / 2
	offsetY := (scaledH - b.Bounds().Height) / 2
	rect := rl.NewRectangle(b.Bounds().X-offsetX, b.Bounds().Y-offsetY, scaledW, scaledH)

	if b.styleName == "toolbar-btn" {
		snap := snapControlRect(rect)
		drawToolbarChrome(rect, b.styleName, style, b.hovered, b.pressed, false)
		inner := toolbarBtnInnerRect(snap, style)
		text := b.Text.Get()
		// Left-align in the inner band — SDF measureTextS is slightly narrower than
		// drawTextS ink, so centering looked like extra padding on the left (Sync/Clear).
		posX := int32(inner.X + 0.5)
		posY := toolbarTextPosY(inner, style)
		drawTextS(text, posX, posY, style)
		return
	}
	if b.styleName == "toolbar-toggle-label" {
		draw := style
		if b.ToggleBinding != nil && b.ToggleBinding.Get() {
			draw.TextColor = rl.NewColor(79, 70, 229, 255)
			draw.Bold = true
		} else {
			draw.TextColor = rl.NewColor(107, 114, 128, 255)
		}
		if b.hovered && !b.pressed {
			draw.TextColor = rl.NewColor(55, 65, 81, 255)
		}
		if b.pressed {
			draw.TextColor = rl.NewColor(67, 56, 202, 255)
		}
		content := toolbarContentRect(b.Bounds(), style)
		text := b.Text.Get()
		textW := measureTextS(text, draw)
		posX := int32(content.X) + (int32(content.Width)-textW)/2
		posY := toolbarTextPosY(content, draw)
		drawTextS(text, posX, posY, draw)
		return
	}

	color := style.BackgroundColor
	if !stateApplied {
		color, _ = buttonInteractionColors(style, b.hovered, b.pressed)
	}

	// Rounded fill only (ribbon-tab style — no outline stroke).
	drawRoundedControl(rect, scaledW, scaledH, style.CornerRadius, color)

	// Center text
	text := b.Text.Get()
	textW := measureTextS(text, style)
	posX := int32(b.Bounds().X) + (int32(b.Bounds().Width)-textW)/2
	posY := TextPosY(b.Bounds(), style)

	drawTextS(text, posX, posY, style)
}

// IsInteractive implements Node.IsInteractive.
func (b *Button) IsInteractive() bool { return true }

// InteractionOverlayActive implements InteractionOverlayPainter.
// Label chrome is redrawn in the SSAA cache so hover/press text stays crisp.
func (b *Button) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (b *Button) DrawInteractionOverlay() {}

func buttonStyleState(hovered, pressed bool) StyleState {
	if pressed {
		return StyleStatePressed
	}
	if hovered {
		return StyleStateHover
	}
	return StyleStateNone
}

// focusRingIndigo is the same RGB used by TextInput and SearchBar focus rings.
var focusRingIndigo = rl.NewColor(79, 70, 229, 255)

// listOptionSelectedBg is the fill for selected popup/list rows (updated by SetAppearance).
var listOptionSelectedBg = rl.NewColor(237, 239, 254, 255)

// listTileHoverOverlay is drawn over list rows on hover (updated by SetAppearance).
var listTileHoverOverlay = rl.NewColor(0, 0, 0, 10)

// buttonInteractionColors returns background and border colours for Button and
// IconButton from the current style and hover/pressed mouse state. Brightness
// deltas stay in lockstep between the two widgets (polish §2.1). When pressed,
// borders blend toward focusRingIndigo so active chrome matches focused fields,
// except for warm red/orange borders (e.g. danger) which stay brightness-only
// (polish §2.2).
func buttonInteractionColors(style Style, hovered, pressed bool) (bg rl.Color, border rl.Color) {
	bg = style.BackgroundColor
	border = style.BorderColor
	if pressed {
		bg = rl.ColorBrightness(style.BackgroundColor, -0.12)
		dim := rl.ColorBrightness(style.BorderColor, -0.1)
		if pressedBorderUsesIndigoAccent(style.BorderColor) {
			border = lerpColor(dim, focusRingIndigo, 0.38)
		} else {
			border = dim
		}
	} else if hovered {
		bg = rl.ColorBrightness(style.BackgroundColor, 0.25)
		border = rl.ColorBrightness(style.BorderColor, 0.15)
	}
	return bg, border
}

func pressedBorderUsesIndigoAccent(b rl.Color) bool {
	// Warm red / orange family: do not tint toward indigo on press.
	if b.R > 140 && int32(b.G) < int32(b.R)-15 && int32(b.B) < int32(b.R)-15 {
		return false
	}
	return true
}

const controlRoundSegments = 12

// drawRoundedControl paints a smooth filled rounded rect (no border stroke).
func drawRoundedControl(rect rl.Rectangle, w, h, cornerRadius float32, fill rl.Color) {
	roundness := buttonCornerRoundness(w, h, cornerRadius)
	if roundness <= 0 {
		rl.DrawRectangleRec(rect, fill)
		return
	}
	rl.DrawRectangleRounded(rect, roundness, controlRoundSegments, fill)
}

// buttonCornerRoundness maps style.CornerRadius and rect size to DrawRectangleRounded
// roundness (0..1), identical to the pre-refactor Button formula (polish §2.1).
func buttonCornerRoundness(width, height, cornerRadius float32) float32 {
	if cornerRadius <= 0 {
		return 0
	}
	halfMin := height / 2
	if width/2 < halfMin {
		halfMin = width / 2
	}
	if halfMin <= 0 {
		return 0
	}
	r := cornerRadius / halfMin
	if r > 1 {
		return 1
	}
	return r
}

// toolbarInteractionColors returns Bootstrap-inspired fill/border for toolbar-btn
// (btn-light) and toolbar-cell (ribbon ghost). checked is persistent toggle-on
// chrome; one-shot clicks use hovered/pressed only.
func toolbarInteractionColors(styleName string, hovered, pressed, checked bool) (bg, border rl.Color) {
	st := GetThemeStyle(styleName)
	bg = st.BackgroundColor
	border = st.BorderColor
	if bg.A == 0 {
		bg = rl.NewColor(255, 255, 255, 255)
	}
	if border.A == 0 {
		border = rl.NewColor(222, 226, 230, 255)
	}
	switch styleName {
	case "toolbar-btn", "toolbar-menu":
		if pressed {
			bg = rl.ColorBrightness(bg, -0.08)
			border = rl.ColorBrightness(border, -0.12)
		} else if hovered {
			bg = rl.ColorBrightness(bg, 0.06)
		}
	case "toolbar-cell":
		bg = rl.NewColor(0, 0, 0, 0)
		border = rl.NewColor(0, 0, 0, 0)
		if checked {
			bg = rl.NewColor(99, 102, 241, 44)
			border = rl.NewColor(99, 102, 241, 90)
		} else if pressed {
			bg = rl.NewColor(99, 102, 241, 32)
			border = rl.NewColor(99, 102, 241, 64)
		} else if hovered {
			bg = rl.NewColor(99, 102, 241, 18)
			border = rl.NewColor(99, 102, 241, 36)
		}
	}
	return bg, border
}

// toolbarStylePadding is inner inset between toolbar control chrome and icon/label.
func toolbarStylePadding(style Style) float32 {
	if style.Padding > 0 {
		return float32(style.Padding)
	}
	return 8
}

// toolbarContentRect is the area inside toolbar button chrome for icons and labels.
func toolbarContentRect(bounds rl.Rectangle, style Style) rl.Rectangle {
	p := toolbarStylePadding(style)
	b := snapControlRect(bounds)
	if b.Width <= 2*p || b.Height <= 2*p {
		return b
	}
	return snapLayoutRect(rl.NewRectangle(b.X+p, b.Y+p, b.Width-2*p, b.Height-2*p))
}

// drawToolbarChrome paints btn-light or ribbon ghost hover for toolbar controls.
// rect is the full widget slot assigned by toolbar layout (chrome fills the slot).
func drawToolbarChrome(rect rl.Rectangle, styleName string, style Style, hovered, pressed, checked bool) {
	bg, border := toolbarInteractionColors(styleName, hovered, pressed, checked)
	if bg.A == 0 && border.A == 0 {
		return
	}
	snap := snapControlRect(rect)
	if snap.Width < 1 || snap.Height < 1 {
		return
	}
	if styleName == "toolbar-cell" {
		rnCell := buttonCornerRoundness(snap.Width, snap.Height, style.CornerRadius)
		drawRoundedControl(snap, snap.Width, snap.Height, style.CornerRadius, bg)
		if border.A > 0 && (hovered || pressed || checked) {
			drawRoundedInsetBorder(snap, rnCell, 1, border, rl.NewColor(0, 0, 0, 0))
		}
		return
	}
	// toolbar-btn: flat rectangle on hover — rounded fills leave SSAA corner fringe.
	if styleName == "toolbar-btn" && (hovered || pressed) {
		rl.DrawRectangleRec(snap, bg)
		return
	}
	rn := buttonCornerRoundness(snap.Width, snap.Height, style.CornerRadius)
	drawRoundedInsetBorder(snap, rn, 1, border, bg)
}
