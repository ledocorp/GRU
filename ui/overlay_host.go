// Package ui (continued) — shared overlay host for Modal, Drawer, BottomSheet (Phase C5).
//
// OverlayHost unifies open/close state, scrim drawing, fade/slide animation,
// backdrop dismiss, Escape, skip-frame input gating, and focus tracking.
//
// # LLM Prompt Template
//
//	host := ui.DefaultOverlayHost(ui.OverlayAnimSlideLeft)
//	host.BeginOpen()
//	host.AdvanceAnimation(dt)
//	host.DrawScrimRect(host.ContentBand(sw, sh))
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const overlaySkipFramesOpen = 2

// OverlayAnimKind selects the overlay entrance animation.
type OverlayAnimKind int

const (
	// OverlayAnimFadeScale — centered modal fade + scale-up (0.94 → 1).
	OverlayAnimFadeScale OverlayAnimKind = iota
	// OverlayAnimSlideLeft — drawer slides in from the left edge.
	OverlayAnimSlideLeft
	// OverlayAnimSlideBottom — bottom sheet slides up from the bottom edge.
	OverlayAnimSlideBottom
)

// OverlayHost owns shared overlay lifecycle, animation, scrim, and input gating.
type OverlayHost struct {
	Open, LogicallyOpen, Closing bool

	Kind OverlayAnimKind

	// FadeScale fields (modal).
	Alpha float32
	Scale float32

	// Slide fields (drawer, bottom sheet).
	Progress float32

	SkipFrames      int
	ArmMouseRelease bool

	ContentTop    float32
	ContentBottom float32

	CloseOnBackdrop bool
	CloseOnEscape   bool

	FadeInTime    float32
	FadeOutTime   float32
	ScaleFrom     float32
	SlideTime     float32
	ScrimStrength float32 // 0–255 scrim alpha at full opacity

	FocusedWidget Node
}

// DefaultOverlayHost returns a host preconfigured for the given animation kind.
func DefaultOverlayHost(kind OverlayAnimKind) *OverlayHost {
	h := &OverlayHost{
		Kind:            kind,
		CloseOnBackdrop: true,
		CloseOnEscape:   true,
	}
	switch kind {
	case OverlayAnimFadeScale:
		h.FadeInTime = 0.15
		h.FadeOutTime = 0.12
		h.ScaleFrom = 0.94
		h.ScrimStrength = 140
	case OverlayAnimSlideLeft:
		h.SlideTime = 0.22
		h.ScrimStrength = 0.45 * 255
	case OverlayAnimSlideBottom:
		h.SlideTime = 0.24
		h.ScrimStrength = 0.4 * 255
	}
	return h
}

// BeginOpen starts the open animation and input settling frames.
func (h *OverlayHost) BeginOpen() {
	h.Open = true
	h.LogicallyOpen = true
	h.Closing = false
	h.SkipFrames = overlaySkipFramesOpen
	switch h.Kind {
	case OverlayAnimFadeScale:
		h.Alpha = 0
		h.Scale = h.ScaleFrom
		h.ArmMouseRelease = false
	default:
		if h.Progress < 0 {
			h.Progress = 0
		}
		h.ArmMouseRelease = true
	}
}

// BeginClose starts the close animation. IsOpen() becomes false immediately.
func (h *OverlayHost) BeginClose() {
	if !h.Open {
		return
	}
	h.LogicallyOpen = false
	h.Closing = true
	if h.FocusedWidget != nil {
		h.FocusedWidget.Emit(EventBlur, nil)
		h.FocusedWidget = nil
	}
}

// Reset immediately dismisses the overlay with no animation.
func (h *OverlayHost) Reset() {
	h.Open = false
	h.LogicallyOpen = false
	h.Closing = false
	h.Alpha = 0
	h.Scale = 0
	h.Progress = 0
	h.SkipFrames = 0
	h.ArmMouseRelease = false
	h.FocusedWidget = nil
}

// IsOpen reports logical open state (false once BeginClose is called).
func (h *OverlayHost) IsOpen() bool { return h.LogicallyOpen }

// IsVisible reports whether the overlay is rendered, including close animation.
func (h *OverlayHost) IsVisible() bool {
	if !h.Open {
		return false
	}
	return h.Opacity() > 0
}

// Opacity returns the current scrim/content opacity multiplier [0..1].
func (h *OverlayHost) Opacity() float32 {
	switch h.Kind {
	case OverlayAnimFadeScale:
		return h.Alpha
	default:
		return h.Progress
	}
}

// IsAnimating reports fade/slide in progress or open settling frames.
func (h *OverlayHost) IsAnimating() bool {
	if !h.Open {
		return false
	}
	switch h.Kind {
	case OverlayAnimFadeScale:
		return h.Closing || h.Alpha < 1 || h.Scale < 0.999 || h.SkipFrames > 0
	default:
		return h.Closing || h.Progress < 0.999 || h.SkipFrames > 0
	}
}

// AdvanceAnimation steps fade or slide tweens. When fully closed, Open becomes false.
func (h *OverlayHost) AdvanceAnimation(dt float32) {
	if !h.Open {
		return
	}
	switch h.Kind {
	case OverlayAnimFadeScale:
		h.advanceFadeScale(dt)
	default:
		h.advanceSlide(dt)
	}
}

func (h *OverlayHost) advanceFadeScale(dt float32) {
	if h.Closing {
		h.Alpha -= dt / h.FadeOutTime
		h.Scale = h.ScaleFrom + h.Alpha*(1-h.ScaleFrom)
		if h.Alpha <= 0 {
			h.Reset()
		}
		return
	}
	if h.Alpha < 1 {
		h.Alpha += dt / h.FadeInTime
		if h.Alpha > 1 {
			h.Alpha = 1
		}
	}
	if h.Scale < 1 {
		h.Scale += dt / h.FadeInTime * (1 - h.ScaleFrom)
		if h.Scale > 1 {
			h.Scale = 1
		}
	}
}

func (h *OverlayHost) advanceSlide(dt float32) {
	if h.Closing {
		h.Progress -= dt / h.SlideTime
		if h.Progress <= 0 {
			h.Reset()
		}
		return
	}
	if h.Progress < 1 {
		h.Progress += dt / h.SlideTime
		if h.Progress > 1 {
			h.Progress = 1
		}
	}
}

// TickSkipFrames decrements the post-open input skip counter.
func (h *OverlayHost) TickSkipFrames() {
	if h.SkipFrames > 0 {
		h.SkipFrames--
	}
}

// InputReady is false during skip frames or while the opening click is still held.
func (h *OverlayHost) InputReady() bool {
	if h.SkipFrames > 0 {
		return false
	}
	if h.ArmMouseRelease {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			return false
		}
		h.ArmMouseRelease = false
	}
	return true
}

// SetContentInsets limits overlay panels and scrim to the band below top chrome
// and above bottom chrome (e.g. title bar + launcher nav).
func (h *OverlayHost) SetContentInsets(top, bottom float32) {
	if top < 0 {
		top = 0
	}
	if bottom < 0 {
		bottom = 0
	}
	h.ContentTop = top
	h.ContentBottom = bottom
}

// ContentBand returns the rectangle overlays may occupy between chrome insets.
func (h *OverlayHost) ContentBand(sw, sh float32) rl.Rectangle {
	top := h.ContentTop
	bottom := h.ContentBottom
	bandH := sh - top - bottom
	if bandH < 1 {
		bandH = 1
	}
	return rl.NewRectangle(0, top, sw, bandH)
}

// DrawScrimRect fills rect with the host scrim color at current opacity.
func (h *OverlayHost) DrawScrimRect(r rl.Rectangle) {
	a := h.Opacity()
	if a <= 0 {
		return
	}
	scrim := uint8(h.ScrimStrength * a)
	if scrim > 255 {
		scrim = 255
	}
	rl.DrawRectangleRec(r, rl.NewColor(0, 0, 0, scrim))
}

// DrawScrimFull draws the scrim over the full screen.
func (h *OverlayHost) DrawScrimFull(sw, sh float32) {
	h.DrawScrimRect(rl.NewRectangle(0, 0, sw, sh))
}

// HandleEscape calls onDismiss when Escape is pressed and CloseOnEscape is set.
func (h *OverlayHost) HandleEscape(onDismiss func()) bool {
	if h.CloseOnEscape && rl.IsKeyPressed(rl.KeyEscape) {
		onDismiss()
		return true
	}
	return false
}

// HandleBackdrop calls onDismiss when the user clicks outside panelRect.
// For slide-left overlays, clicks must land inside bandRect but outside panelRect.
func (h *OverlayHost) HandleBackdrop(panelRect, bandRect rl.Rectangle, onDismiss func()) bool {
	if !h.CloseOnBackdrop || !rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		return false
	}
	mouse := rl.GetMousePosition()
	inPanel := rl.CheckCollisionPointRec(mouse, panelRect)
	if h.Kind == OverlayAnimSlideLeft {
		inBand := rl.CheckCollisionPointRec(mouse, bandRect)
		if inBand && !inPanel {
			onDismiss()
			return true
		}
		return false
	}
	if !inPanel {
		onDismiss()
		return true
	}
	return false
}

// HandleFocusClick tracks keyboard focus for interactive content inside panelRect.
func (h *OverlayHost) HandleFocusClick(content Node, panelRect rl.Rectangle) {
	if content == nil || !rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		return
	}
	mouse := rl.GetMousePosition()
	if !rl.CheckCollisionPointRec(mouse, panelRect) {
		return
	}
	hit := FindInteractiveAt(content, mouse)
	if hit == h.FocusedWidget {
		return
	}
	if d := ActiveDocument(); d != nil {
		d.SetFocus(hit)
		h.FocusedWidget = hit
		return
	}
	if h.FocusedWidget != nil {
		h.FocusedWidget.Emit(EventBlur, nil)
	}
	h.FocusedWidget = hit
	if hit != nil {
		hit.Emit(EventFocus, nil)
	}
}

// ScaledCenterBox applies a pop-in scale transform around the box centre.
func ScaledCenterBox(box rl.Rectangle, scale float32) rl.Rectangle {
	sW := box.Width * scale
	sH := box.Height * scale
	sX := box.X + (box.Width-sW)/2
	sY := box.Y + (box.Height-sH)/2
	return rl.NewRectangle(sX, sY, sW, sH)
}

// SlidePanelRect returns the drawer panel rectangle for the current slide progress.
func (h *OverlayHost) SlidePanelRect(band rl.Rectangle, panelW float32) rl.Rectangle {
	x := band.X - panelW + h.Progress*panelW
	return rl.NewRectangle(x, band.Y, panelW, band.Height)
}

// SlideSheetRect returns the bottom sheet rectangle for the current slide progress.
func (h *OverlayHost) SlideSheetRect(sw, sh, sheetH float32) rl.Rectangle {
	y := sh - h.Progress*sheetH
	return rl.NewRectangle(0, y, sw, sheetH)
}

// layoutOverlaySubtree positions and lays out overlay content outside the document tree.
func layoutOverlaySubtree(content Node, rect rl.Rectangle) {
	if content == nil {
		return
	}
	content.SetBounds(rect)
	content.Layout()
}

// drawOverlaySubtree lays out and draws overlay content at 1× screen pixels when needed.
func drawOverlaySubtree(content Node, rect rl.Rectangle) {
	drawOverlaySubtreeClipped(content, rect, false)
}

// drawOverlaySubtreeClipped lays out overlay content; when clip is true, content is
// scissored to rect so modal bodies cannot paint into the footer band.
func drawOverlaySubtreeClipped(content Node, rect rl.Rectangle, clip bool) {
	if content == nil {
		return
	}
	layoutOverlaySubtree(content, rect)
	prevScale := RenderScale
	if !SuperFrameActive() {
		RenderScale = 1.0
	}
	defer func() {
		if !SuperFrameActive() {
			RenderScale = prevScale
		}
	}()

	if !clip || rect.Width < 1 || rect.Height < 1 {
		content.Draw()
		return
	}

	// Clip leaf nodes directly; for flex bodies draw children with per-widget
	// scissor restore (TextInput/Label call EndScissorMode).
	if c, ok := content.(*Container); ok {
		drawChildrenInRectClip(c, rect, c.sortedChildren())
		return
	}
	withClipRestore(content, rect, func() {
		content.Draw()
	})
}
