// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// BadgeVariant controls the color palette of a Badge.
// Each variant maps to a named theme style (e.g. BadgePrimary → "badge-primary").
type BadgeVariant int

const (
	BadgeDefault BadgeVariant = iota // Neutral grey — general-purpose tag
	BadgePrimary                     // Indigo — highlights the main category
	BadgeSuccess                     // Green — confirms a positive state
	BadgeWarning                     // Amber — cautions about a pending state
	BadgeDanger                      // Red — flags an error or destructive action
	BadgeInfo                        // Sky blue — informational, low urgency
)

// BadgeShape controls the corner style of a Badge.
type BadgeShape int

const (
	BadgeShapePill    BadgeShape = iota // Fully rounded pill — default
	BadgeShapeRounded                   // Slightly rounded corners (~20% of height)
	BadgeShapeRect                      // Sharp rectangular corners
)

// String returns a short inspector- and log-friendly name for the shape.
func (s BadgeShape) String() string {
	switch s {
	case BadgeShapePill:
		return "pill"
	case BadgeShapeRounded:
		return "rounded"
	case BadgeShapeRect:
		return "rectangle"
	default:
		return fmt.Sprintf("BadgeShape(%d)", int(s))
	}
}

// badgeStyleName maps each BadgeVariant to its theme key.
var badgeStyleName = [...]string{
	BadgeDefault: "badge",
	BadgePrimary: "badge-primary",
	BadgeSuccess: "badge-success",
	BadgeWarning: "badge-warning",
	BadgeDanger:  "badge-danger",
	BadgeInfo:    "badge-info",
}

var badgeThemeV2VariantName = [...]string{
	BadgeDefault: "default",
	BadgePrimary: "primary",
	BadgeSuccess: "success",
	BadgeWarning: "warning",
	BadgeDanger:  "danger",
	BadgeInfo:    "info",
}

// badgeCloseHit is the square hit/draw target for the dismiss control (px).
const badgeCloseHit = float32(18)

// badgeCloseTrailingPad is inset from the badge's trailing edge to the dismiss control.
const badgeCloseTrailingPad = float32(8)

// chipDefaultH is the default height for filter-chip badges (selectable badges).
const chipDefaultH = float32(32)

// badgeMinPadX is minimum horizontal inset for label text (longer chips stay readable).
const badgeMinPadX = float32(12)

// Badge is a small pill-shaped label used for tags, status chips, and counters.
// Filter “chips” are the same widget: set Selected for toggle selection and/or
// CloseButton + OnClose for dismiss — there is no separate Chip type.
//
// # Sizing
//
// Pass w=0 to let the Badge auto-size horizontally based on the text content
// and padding.  Height defaults to 24 px when h=0.
// Either dimension can be fixed by passing a positive value.
//
// # Close button
//
// Set [Badge.CloseButton] = true to show a × glyph on the right.
// Wire [Badge.OnClose] to react when the user clicks it.
// The badge does NOT hide itself on close — that is the caller's responsibility
// (call badge.Hide(), remove it from its parent, or update the backing data).
//
// # LLM Prompt Template
//
//	tag := ui.NewBadge("status", "Stable", ui.BadgeSuccess, 0, 0, 0, 24)
//	row.AddChild(tag)
//
// Demo scenes: **Batch 3b Accordion**, **Path C Demo**.
type Badge struct {
	Element

	// Text is the reactive label displayed inside the badge.
	// Changing it via Set triggers an auto-resize and redraw.
	Text *Signal[string]

	// Variant sets the color palette (Default, Primary, Success, Warning, Danger, Info).
	// Changing Variant after construction requires a SetStyle call or a full MarkDirty.
	Variant BadgeVariant

	// CloseButton, when true, renders a × glyph on the trailing edge and
	// activates the hit-test region for it.
	CloseButton bool

	// OnClose is called when the user clicks the × button.
	// It is not called when CloseButton is false.
	OnClose func()

	// Selected, when non-nil, makes the badge a toggleable filter chip.
	Selected *Signal[bool]

	// autoSize, when true, means the width was not explicitly set and will be
	// recomputed whenever the text changes.
	autoSize bool

	// Shape controls the corner style (pill / slightly rounded / sharp rectangle).
	// Rectangle uses axis-aligned primitives so fill and clipping match panels.
	Shape BadgeShape

	// hoverClose tracks whether the pointer is currently over the × region.
	hoverClose bool
	hovered    bool
}

// NewBadge creates a Badge.
//
//	id      — unique node ID
//	text    — visible label
//	variant — color palette (BadgeDefault, BadgePrimary, …)
//	x, y    — position (overridden by layout containers)
//	w       — width; 0 = auto-size from text
//	h       — height; 0 = default 24 px
func NewBadge(id, text string, variant BadgeVariant, x, y, w, h float32) *Badge {
	if h == 0 {
		h = 24
	}
	auto := w == 0

	b := &Badge{
		Element:  NewElement(id, x, y, w, h),
		Text:     NewSignal(text),
		Variant:  variant,
		autoSize: auto,
	}

	// Apply both the legacy style name and Theme v2 component/variant.
	b.styleName = badgeStyleName[variant]
	b.Element.SetStyleVariant("badge", badgeThemeV2VariantName[variant])

	// When text changes, resize if auto-sizing and trigger a layout+redraw.
	b.Text.Subscribe(func() {
		if b.autoSize {
			b.bounds.Width = b.measureWidth()
		}
		b.MarkDirty()
	})

	// Compute initial width.
	if auto {
		b.bounds.Width = b.measureWidth()
		b.PreferredWidth = b.bounds.Width
	}

	if b.Selected != nil {
		b.Selected.Subscribe(func() { b.MarkDrawDirty() })
	}
	return b
}

// ─── Public API ───────────────────────────────────────────────────────────────

// SetVariant changes the badge color palette and updates the style name.
func (b *Badge) SetVariant(v BadgeVariant) {
	if b.Variant == v {
		return
	}
	b.Variant = v
	b.styleName = badgeStyleName[v]
	b.Element.SetStyleVariant("badge", badgeThemeV2VariantName[v])
	if b.autoSize {
		b.bounds.Width = b.measureWidth()
	}
	b.MarkDrawDirty()
}

// measureWidth returns the pixel width required to fit the current text plus
// horizontal padding.  Used both at construction and whenever autoSize is true.
func (b *Badge) badgePadX(style Style) float32 {
	padX := style.Padding
	if b.autoSize && padX < badgeMinPadX {
		padX = badgeMinPadX
	}
	return padX
}

func (b *Badge) measureWidth() float32 {
	style := b.GetStyle()
	textW := float32(measureTextS(b.Text.Get(), style))
	padX := b.badgePadX(style)
	return textW + padX*2 + b.closeReserve()
}

// SetCloseButton toggles the dismiss control and refreshes auto width.
func (b *Badge) SetCloseButton(on bool) {
	if b.CloseButton == on {
		return
	}
	b.CloseButton = on
	if b.autoSize {
		b.bounds.Width = b.measureWidth()
		b.PreferredWidth = b.bounds.Width
	}
	b.MarkDrawDirty()
}

// GetPreferredWidth returns the intrinsic width for flex layout.
func (b *Badge) GetPreferredWidth() float32 {
	if b.PreferredWidth > 0 {
		return b.PreferredWidth
	}
	if b.autoSize {
		return b.measureWidth()
	}
	return b.Bounds().Width
}

// GetMinWidth pins auto-sized badges in flex and wrap rows.
func (b *Badge) GetMinWidth() float32 {
	if !b.autoSize {
		return 0
	}
	return b.measureWidth()
}

func (b *Badge) closeTextGap() float32 {
	sp := float32(measureTextS(" ", b.GetStyle()))
	if sp < 2 {
		return 3
	}
	return sp
}

func (b *Badge) closeReserve() float32 {
	if !b.CloseButton {
		return 0
	}
	return b.closeTextGap() + badgeCloseHit + badgeCloseTrailingPad
}

// ─── Node implementation ──────────────────────────────────────────────────────

// IsInteractive implements Node.
func (b *Badge) IsInteractive() bool { return b.CloseButton || b.Selected != nil }

// Update handles dismiss, filter-chip selection, and hover.
func (b *Badge) Update(_ float32) {
	if b.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	inBounds := rl.CheckCollisionPointRec(mouse, b.bounds)
	prevHover := b.hovered
	b.hovered = inBounds
	if b.CloseButton {
		b.hoverClose = rl.CheckCollisionPointRec(mouse, b.closeButtonRect())
	} else {
		b.hoverClose = false
	}
	if b.hovered != prevHover || b.hoverClose {
		b.MarkDrawDirty()
	}
	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) || !inBounds {
		return
	}
	if b.CloseButton && b.hoverClose {
		if b.OnClose != nil {
			b.OnClose()
		}
		b.Emit(EventClick, nil)
		return
	}
	if b.Selected != nil {
		b.Selected.Set(!b.Selected.Get())
	}
}

// Layout restores auto width when CloseButton is toggled after construction or
// when a flex parent assigned a band narrower than the label + dismiss control.
func (b *Badge) Layout() {
	defer func() { b.layoutDirty = false }()
	if !b.autoSize {
		return
	}
	need := b.measureWidth()
	bb := b.Bounds()
	if bb.Width+0.5 < need {
		bb.Width = need
		b.setBoundsNoMark(bb)
		b.PreferredWidth = need
	}
}

// Draw renders the badge.
func (b *Badge) Draw() {
	if b.IsHidden() {
		return
	}
	b.drawInternal()
	b.drawDirty = false
}

// drawInternal performs the actual rendering.
func (b *Badge) drawInternal() {
	bounds := b.bounds
	style := b.GetStyle()

	// ── Background fill ────────────────────────────────────────────────────────
	bg := style.BackgroundColor
	if b.Selected != nil && b.Selected.Get() {
		bg = rl.ColorBrightness(bg, 0.08)
	} else if b.hovered && b.Selected != nil {
		bg = rl.ColorBrightness(bg, 0.12)
	}
	useSharpQuad := b.Shape == BadgeShapeRect
	var roundness float32
	switch {
	case useSharpQuad:
		rl.DrawRectangleRec(bounds, bg)
	case b.Shape == BadgeShapeRounded:
		roundness = 0.2
		rl.DrawRectangleRounded(bounds, roundness, 8, bg)
	default: // BadgeShapePill
		roundness = 1.0
		rl.DrawRectangleRounded(bounds, roundness, 8, bg)
	}
	if b.Selected != nil && b.Selected.Get() {
		inset := rl.NewRectangle(bounds.X+2, bounds.Y+2, bounds.Width-4, bounds.Height-4)
		if inset.Width > 0 && inset.Height > 0 {
			rl.DrawRectangleRoundedLinesEx(inset, 1.0, 8, 2, focusRingIndigo)
		}
	}

	// ── Text ──────────────────────────────────────────────────────────────────
	text := b.Text.Get()
	textAvailW := bounds.Width - b.closeReserve()
	padX := b.badgePadX(style)
	textW := float32(measureTextS(text, style))
	textX := bounds.X + padX
	if !b.autoSize && !b.CloseButton && textW <= textAvailW {
		textX = bounds.X + (textAvailW-textW)/2
	}
	textRect := rl.NewRectangle(bounds.X, bounds.Y, textAvailW, bounds.Height)
	drawTextS(text, int32(textX), TextPosY(textRect, style), style)

	// ── Close button ──────────────────────────────────────────────────────────
	if b.CloseButton {
		cr := b.closeButtonRect()

		if b.hoverClose {
			rl.DrawCircleV(
				rl.NewVector2(cr.X+cr.Width/2, cr.Y+cr.Height/2),
				cr.Width/2-1,
				rl.NewColor(0, 0, 0, 35),
			)
		}

		iconSize := cr.Height * 0.9
		if iconSize < 18 {
			iconSize = 18
		}
		if iconSize > 20 {
			iconSize = 20
		}
		inner := rl.NewRectangle(
			cr.X+(cr.Width-iconSize)/2,
			cr.Y+(cr.Height-iconSize)/2,
			iconSize, iconSize,
		)
		Phosphor.EnsureLoaded(PhosphorXCircle, PhosphorRegular)
		if !Phosphor.Draw(inner, PhosphorXCircle, PhosphorRegular, style.TextColor) {
			Phosphor.EnsureLoaded(PhosphorX, PhosphorRegular)
			if !Phosphor.Draw(inner, PhosphorX, PhosphorRegular, style.TextColor) {
				drawTextS("x", int32(inner.X), TextPosY(cr, style), style)
			}
		}
	}

	// ── Optional border ──────────────────────────────────────────────────────────
	if style.BorderWidth > 0 {
		if useSharpQuad {
			rl.DrawRectangleLinesEx(bounds, style.BorderWidth, style.BorderColor)
		} else {
			rl.DrawRectangleRoundedLinesEx(bounds, roundness, 8, style.BorderWidth, style.BorderColor)
		}
	}
}

// InteractionOverlayActive implements InteractionOverlayPainter.
func (b *Badge) InteractionOverlayActive() bool {
	return !b.IsHidden() && b.Selected != nil && (b.hovered || b.hoverClose)
}

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (b *Badge) DrawInteractionOverlay() { b.drawInternal() }

// closeButtonRect returns the hit / draw rectangle for the × button.
func (b *Badge) closeButtonRect() rl.Rectangle {
	h := badgeCloseHit
	if b.bounds.Height < h {
		h = b.bounds.Height
	}
	return rl.NewRectangle(
		b.bounds.X+b.bounds.Width-badgeCloseTrailingPad-h,
		b.bounds.Y+(b.bounds.Height-h)/2,
		h,
		h,
	)
}
