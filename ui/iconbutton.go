// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"os"

	"github.com/fogleman/gg"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// IconButton is a compact button that draws a small icon to the left of its
// text label.  The icon source is chosen in priority order:
//  1. VectorPainter — a gg painter func that is rendered off-main the first
//     time the widget is drawn, producing a crisp vector texture.
//  2. IconPath — a PNG/JPG file loaded lazily on first draw.
//  3. Symbol — a UTF-8 glyph drawn as text (lowest cost, zero setup).
//
// When none of the three is set the widget behaves exactly like a regular
// Button.
//
// Scale applies the same center-pivot resize as Button; use the same tween
// curve on both widgets for a matched click bounce (polish §2.4).
//
// Styling is identical to Button — set the style name to "primary", "danger",
// "button", or the new "icon-button" default via SetStyle.
//
// Usage:
//
//	add := ui.NewIconButton("add-btn", "+", "Add Item", 0, 0, 160, 44)
//	add.SetStyle("primary")
//	add.OnClick = func() { ... }
//
//	// With a gg vector painter:
//	btn := ui.NewIconButton("vec-btn", "", "Label", 0, 0, 160, 44)
//	btn.VectorPainter = ui.DrawIconPlus   // any func(*gg.Context)
//	btn.LoadVectorIconAsync(doc, 32)      // render off-main, 32×32 px
//
// # LLM Prompt Template
//
//	btn := ui.NewIconButton("add", "+", "Add", 0, 0, 120, 36)
//	btn.SetStyle("primary")
//	btn.OnClick = func() { /* action */ }
//	toolbar.AddChild(btn)
//
// Demo scenes: **Widgets Demo**, Notepad ribbon, **Settings** back button.
type IconButton struct {
	Element
	Label          *Signal[string]   // Reactive button text
	Symbol         string            // Single-glyph fallback drawn as the icon (UTF-8 OK)
	IconPath       string            // Optional PNG/JPG path; preferred over Symbol when loaded
	VectorPainter  func(*gg.Context) // gg painter; takes precedence over IconPath and Symbol
	OnClick        func()            // Called on click
	// Checked, when set, mirrors persistent toggle state (ribbon pane toggles).
	// One-shot commands leave Checked nil so click chrome does not latch purple.
	Checked *Signal[bool]
	// ToggleCheckedOnClick flips Checked on click when true. Ribbon toggles bind
	// Checked to app state and set this false so OnClick owns the flip.
	ToggleCheckedOnClick bool
	// SuppressCheckedChrome skips forced hover/press chrome when Checked is true
	// (for icon-only toggles where the glyph change is enough feedback).
	SuppressCheckedChrome bool
	hovered        bool
	pressed        bool
	Scale          float32 // For click-bounce animation (same as Button)
	tex            rl.Texture2D
	texOK          bool
	texFail        bool
	usePhosphor    bool
	phosphorName   string
	phosphorWeight PhosphorWeight
	Stacked        bool         // icon above label (ribbon cells)
	vecTex         rl.Texture2D // GPU texture from VectorPainter
	vecTexOK       bool         // true once vecTex has been uploaded
	vecPending     bool         // true while async render is in flight
}

// NewIconButton creates an IconButton with the given symbol and label.
// The widget defaults to the "icon-button" style; call SetStyle to override.
func NewIconButton(id, symbol, label string, x, y, w, h float32) *IconButton {
	ib := &IconButton{
		Element: NewElement(id, x, y, w, h),
		Label:   NewSignal(label),
		Symbol:  symbol,
		Scale:   1.0,
	}
	if w > 0 {
		ib.PreferredWidth = w
	}
	ib.styleName = "icon-button"
	ib.Label.Subscribe(func() { ib.MarkDirty() })
	return ib
}

// GetPreferredWidth shrink-wraps icon + label when width is not explicit.
func (ib *IconButton) GetPreferredWidth() float32 {
	if ib.PreferredWidth > 0 {
		return ib.PreferredWidth
	}
	style := ib.GetStyle()
	pad := style.Padding
	if pad <= 0 {
		pad = 10
	}
	iconW := float32(24)
	if ib.Symbol != "" {
		iconW = float32(measureTextS(ib.Symbol, style)) + 6
	}
	w := iconW + float32(measureTextS(ib.Label.Get(), style)) + 2*pad + 2*style.BorderWidth + 8
	if w < 48 {
		return 48
	}
	return w
}

// Unload releases the GPU texture if one was loaded.
func (ib *IconButton) Unload() {
	if ib.texOK {
		rl.UnloadTexture(ib.tex)
		ib.texOK = false
	}
	if ib.vecTexOK {
		rl.UnloadTexture(ib.vecTex)
		ib.vecTexOK = false
	}
}

// LoadVectorIconAsync renders VectorPainter into a size×size GPU texture using
// the worker pool, then uploads it on the main thread. It is a no-op if
// VectorPainter is nil, the render is already in flight, or the texture is
// already ready. Call this once after Build — the first Draw will then use
// the vector texture instead of the symbol fallback.
func (ib *IconButton) LoadVectorIconAsync(doc *Document, size int) {
	if ib.VectorPainter == nil || ib.vecPending || ib.vecTexOK {
		return
	}
	ib.vecPending = true
	gc := NewIconContext(size)
	ib.VectorPainter(gc)
	AsyncContextToTexture(doc, gc, func(tex rl.Texture2D) {
		ib.vecTex = tex
		ib.vecTexOK = tex.ID != 0
		ib.vecPending = false
		ib.MarkDrawDirty()
	})
}

func (ib *IconButton) tryLoadIcon() {
	if _, err := os.Stat(ib.IconPath); err != nil {
		ib.texFail = true
		return
	}
	ib.tex = rl.LoadTexture(ib.IconPath)
	if ib.tex.ID != 0 {
		ib.texOK = true
	} else {
		rl.UnloadTexture(ib.tex)
		ib.texFail = true
	}
}

// stackedRibbonCell is true for DevExpress-style ribbon icon+label cells.
func (ib *IconButton) stackedRibbonCell() bool {
	return ib.Stacked && ib.styleName == "toolbar-cell"
}

// Update implements Node.Update by handling mouse input.
func (ib *IconButton) Update(_ float32) {
	if ib.IsHidden() {
		return
	}
	if ScenePointerBlocked() {
		if ib.hovered || ib.pressed {
			ib.hovered = false
			ib.pressed = false
			if !ib.ghostChrome() {
				ib.MarkDrawDirty()
			}
		}
		return
	}
	mouse := rl.GetMousePosition()
	prevHovered := ib.hovered
	prevPressed := ib.pressed
	ib.hovered = rl.CheckCollisionPointRec(mouse, ib.Bounds())
	ib.pressed = ib.hovered && rl.IsMouseButtonDown(rl.MouseLeftButton)
	if ib.hovered != prevHovered || ib.pressed != prevPressed {
		if !ib.ghostChrome() {
			ib.MarkDrawDirty()
		}
	}
	if ib.Scale != 1.0 && (ib.hovered != prevHovered || ib.pressed != prevPressed) {
		ib.MarkDrawDirty()
	}
	// PointerClickConsume only — raw IsMouseButtonPressed bypassed drawer/modal blocks.
	if ib.hovered && PointerClickConsume(ib.Bounds()) {
		if ib.OnClick != nil {
			ib.OnClick()
		}
		if ib.Checked != nil && ib.ToggleCheckedOnClick {
			ib.Checked.Set(!ib.Checked.Get())
			ib.MarkDrawDirty()
		}
	}
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (ib *IconButton) ClearOverlayPointerState() {
	if !ib.hovered && !ib.pressed {
		return
	}
	ib.hovered = false
	ib.pressed = false
	ib.MarkDrawDirty()
}

// Layout implements Node.Layout (leaf — bounds come from parent Toolbar/Container).
func (ib *IconButton) Layout() {
	ib.layoutDirty = false
}

// InteractionOverlayActive implements InteractionOverlayPainter.
func (ib *IconButton) InteractionOverlayActive() bool {
	checked := ib.Checked != nil && ib.Checked.Get()
	if ib.stackedRibbonCell() {
		return false
	}
	return !ib.IsHidden() && (ib.hovered || ib.pressed || ib.Scale != 1.0 || checked)
}

// CollectRibbonIconWake returns WakeInput when the cursor is over a stacked
// ribbon cell. Merged before Update so deep idle ramps to ActiveFPS on hover —
// the latched click then lands on the first frame at full rate.
func CollectRibbonIconWake(root Node) WakeSummary {
	var out WakeSummary
	collectRibbonIconWake(root, &out)
	return out
}

func collectRibbonIconWake(n Node, out *WakeSummary) {
	if n == nil || n.IsHidden() {
		return
	}
	if ib, ok := n.(*IconButton); ok && ib.stackedRibbonCell() {
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), ib.Bounds()) {
			out.Add(WakeInput, ib.ID())
		}
	}
	for _, ch := range n.Children() {
		collectRibbonIconWake(ch, out)
	}
}

// ghostChrome is true for transparent icon buttons (appbar-icon). Toolbar buttons use
// solid chrome and must redraw icon + background together on hover (overlay is opaque).
func (ib *IconButton) ghostChrome() bool {
	if ib.styleName == "toolbar-cell" || ib.styleName == "toolbar-btn" {
		return false
	}
	if ib.styleName == "appbar-icon" {
		return true
	}
	return ib.GetStyle().BackgroundColor.A == 0
}

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (ib *IconButton) DrawInteractionOverlay() {
	ib.drawChrome(ib.hovered, ib.pressed)
	if !ib.ghostChrome() {
		ib.drawContent()
	}
}

// Draw implements Node.Draw.
func (ib *IconButton) Draw() {
	defer func() { ib.drawDirty = false }()
	if ib.IsHidden() {
		return
	}
	if ib.ghostChrome() {
		if ib.hovered || ib.pressed {
			ib.drawChrome(ib.hovered, ib.pressed)
		}
	} else {
		ib.drawChrome(ib.hovered, ib.pressed)
	}
	ib.drawContent()
}

func (ib *IconButton) scaledRect() rl.Rectangle {
	scaledW := ib.Bounds().Width * ib.Scale
	scaledH := ib.Bounds().Height * ib.Scale
	ox := (scaledW - ib.Bounds().Width) / 2
	oy := (scaledH - ib.Bounds().Height) / 2
	return rl.NewRectangle(ib.Bounds().X-ox, ib.Bounds().Y-oy, scaledW, scaledH)
}

func (ib *IconButton) drawChrome(hovered, pressed bool) {
	if ib.IsHidden() {
		return
	}
	checked := ib.Checked != nil && ib.Checked.Get() && !ib.SuppressCheckedChrome
	style, stateApplied := ib.ResolveStyle(buttonStyleState(hovered, pressed))
	rect := ib.scaledRect()

	if ib.styleName == "toolbar-btn" || ib.styleName == "toolbar-cell" {
		drawToolbarChrome(rect, ib.styleName, style, hovered, pressed, checked)
		return
	}

	bgColor := style.BackgroundColor
	if !stateApplied {
		bgColor, _ = buttonInteractionColors(style, hovered, pressed)
	}
	if style.BackgroundColor.A > 0 || style.BorderWidth > 0 {
		drawRoundedControl(rect, rect.Width, rect.Height, style.CornerRadius, bgColor)
	} else if hovered || pressed {
		hoverBg := rl.NewColor(0, 0, 0, 18)
		if pressed {
			hoverBg = rl.NewColor(0, 0, 0, 32)
		}
		if ib.styleName == "appbar-icon" && ThemeIsDark() {
			hoverBg = rl.NewColor(255, 255, 255, 18)
			if pressed {
				hoverBg = rl.NewColor(255, 255, 255, 32)
			}
		}
		r := rect.Width / 2
		if rect.Height < rect.Width {
			r = rect.Height / 2
		}
		cx := rect.X + rect.Width/2
		cy := rect.Y + rect.Height / 2
		if ib.styleName == "appbar-icon" {
			inset := float32(3)
			if r > inset {
				r -= inset
			}
			cy -= 1
		}
		rl.DrawCircleV(rl.NewVector2(cx, cy), r, hoverBg)
	}
}

func (ib *IconButton) drawContent() {
	if ib.IsHidden() {
		return
	}

	// Lazy-load icon texture on first draw
	if ib.IconPath != "" && !ib.texOK && !ib.texFail {
		ib.tryLoadIcon()
	}

	style, _ := ib.ResolveStyle(StyleStateNone)
	rect := ib.scaledRect()

	// ── Content: icon + label centred as a group ──────────────────────────────
	// Priority: vector → phosphor cache → file texture → symbol text
	hasVecIcon := ib.vecTexOK
	hasPhosphor := !hasVecIcon && ib.usePhosphor
	hasTexIcon := !hasVecIcon && !hasPhosphor && ib.texOK
	hasSymIcon := !hasVecIcon && !hasPhosphor && !hasTexIcon && ib.Symbol != ""

	label := ib.Label.Get()
	iconOnly := label == "" && (hasPhosphor || hasVecIcon || hasTexIcon || hasSymIcon)

	if ib.Stacked && label != "" && (hasPhosphor || hasVecIcon || hasTexIcon || hasSymIcon) {
		ib.drawStackedRibbonCell(style, rect, label, hasPhosphor, hasVecIcon, hasTexIcon, hasSymIcon)
		return
	}

	content := rect
	if ib.styleName == "toolbar-btn" {
		content = toolbarContentRect(rect, style)
	}

	textW := measureTextS(label, style)
	iconSize := float32(style.FontSize)
	if iconOnly {
		iconSize = phosphorIconSize(content.Height, style.FontSize)
	}
	const iconGap = int32(8)

	contentW := textW
	if hasVecIcon || hasPhosphor || hasTexIcon || hasSymIcon {
		contentW += int32(iconSize) + iconGap
	}

	startX := int32(content.X) + (int32(content.Width)-contentW)/2
	textY := TextPosY(content, style)

	if hasVecIcon {
		src := rl.NewRectangle(0, 0, float32(ib.vecTex.Width), float32(ib.vecTex.Height))
		dst := rl.NewRectangle(float32(startX), float32(textY), iconSize, iconSize)
		rl.DrawTexturePro(ib.vecTex, src, dst, rl.NewVector2(0, 0), 0, style.TextColor)
		startX += int32(iconSize) + iconGap
	} else if hasPhosphor {
		dst := snapPhosphorRect(rl.NewRectangle(
			content.X+(content.Width-iconSize)/2,
			content.Y+(content.Height-iconSize)/2,
			iconSize, iconSize,
		))
		if label != "" {
			dst = snapPhosphorRect(rl.NewRectangle(float32(startX), float32(textY), iconSize, iconSize))
		}
		if phosphorIconReady(ib.phosphorName, ib.phosphorWeight) {
			Phosphor.Draw(dst, ib.phosphorName, ib.phosphorWeight, style.TextColor)
		}
		if label != "" {
			startX += int32(iconSize) + iconGap
		}
	} else if hasTexIcon {
		// Draw texture tinted to match the button text colour (e.g. white-on-indigo)
		src := rl.NewRectangle(0, 0, float32(ib.tex.Width), float32(ib.tex.Height))
		dst := rl.NewRectangle(float32(startX), float32(textY), float32(iconSize), float32(iconSize))
		rl.DrawTexturePro(ib.tex, src, dst, rl.NewVector2(0, 0), 0, style.TextColor)
		startX += int32(iconSize) + iconGap
	} else if hasSymIcon {
		// Symbol text: centre within the icon column
		symW := measureTextS(ib.Symbol, style)
		symX := startX + (int32(iconSize)-symW)/2
		drawTextS(ib.Symbol, symX, textY, style)
		startX += int32(iconSize) + iconGap
	}

	drawTextS(label, startX, textY, style)
}

// drawStackedRibbonCell lays out icon + caption inside a ribbon cell with even inner padding.
func (ib *IconButton) drawStackedRibbonCell(style Style, cell rl.Rectangle, label string, hasPhosphor, hasVecIcon, hasTexIcon, hasSymIcon bool) {
	const innerPad = float32(7)
	capStyle := style
	capStyle.FontSize = tbRibbonCaptionFS
	labelH := EffectiveFontSize(capStyle)
	textTop := cell.Y + cell.Height - innerPad - labelH
	iconAreaH := textTop - cell.Y - innerPad
	if iconAreaH < 10 {
		iconAreaH = 10
	}
	iconSize := phosphorIconSize(iconAreaH, 22)
	if iconSize > cell.Width-2*innerPad {
		iconSize = cell.Width - 2*innerPad
	}
	if iconSize > iconAreaH {
		iconSize = iconAreaH
	}
	iconY := cell.Y + innerPad + (iconAreaH-iconSize)/2
	iconX := cell.X + (cell.Width-iconSize)/2
	if hasPhosphor && phosphorIconReady(ib.phosphorName, ib.phosphorWeight) {
		dst := snapPhosphorRect(rl.NewRectangle(iconX, iconY, iconSize, iconSize))
		Phosphor.Draw(dst, ib.phosphorName, ib.phosphorWeight, style.TextColor)
	} else if hasVecIcon {
		src := rl.NewRectangle(0, 0, float32(ib.vecTex.Width), float32(ib.vecTex.Height))
		dst := rl.NewRectangle(iconX, iconY, iconSize, iconSize)
		rl.DrawTexturePro(ib.vecTex, src, dst, rl.NewVector2(0, 0), 0, style.TextColor)
	} else if hasTexIcon {
		src := rl.NewRectangle(0, 0, float32(ib.tex.Width), float32(ib.tex.Height))
		dst := rl.NewRectangle(iconX, iconY, iconSize, iconSize)
		rl.DrawTexturePro(ib.tex, src, dst, rl.NewVector2(0, 0), 0, style.TextColor)
	} else if hasSymIcon {
		symW := measureTextS(ib.Symbol, style)
		drawTextS(ib.Symbol, int32(iconX+(iconSize-float32(symW))/2), int32(iconY+(iconSize-labelH)/2), style)
	}
	textW := measureTextS(label, capStyle)
	maxLabelW := cell.Width - 2*innerPad
	if maxLabelW > 0 && float32(textW) > maxLabelW {
		label = truncateTextS(label, maxLabelW, capStyle)
		textW = measureTextS(label, capStyle)
	}
	tx := int32(cell.X + (cell.Width-float32(textW))/2)
	ty := int32(textTop)
	drawTextS(label, tx, ty, capStyle)
}

// IsInteractive implements Node.IsInteractive.
func (ib *IconButton) IsInteractive() bool { return true }
