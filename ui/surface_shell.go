// Package ui (continued) — chrome + header shell around a dumb RaisedSurface body.
//
// SurfaceShell orchestrates draw order and body bounds; RaisedSurface alone runs
// flex/clamp/scissor (see docs/SURFACE_COMPOSITION_PLAN.md §8.13).
package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// SurfaceShell owns outer bounds, chrome, and the title band. Body layout lives
// on the inner RaisedSurface child only.
type SurfaceShell struct {
	Container
	Title       string
	TitleHeight float32
	headerMode  SurfaceHeaderMode
	body        *RaisedSurface
	behaviors   []SurfaceBehavior
	collapse      *CollapseBehavior
	panelFeatures *PanelFeaturesBehavior
	dismiss       *DismissBehavior
	escape        *EscapeBehavior
}

func (sh *SurfaceShell) attachBody(id string) {
	if sh.body != nil {
		return
	}
	sh.body = newRaisedSurfaceBody(id + "-body")
	sh.Container.children = append(sh.Container.children, sh.body)
	sh.body.SetParent(sh)
}

func (sh *SurfaceShell) surfaceHeader() SurfaceHeader {
	mode := sh.headerMode
	if sh.ChromeKind() == ChromeGlass {
		mode = HeaderModeGlass
	}
	return SurfaceHeader{
		Title:          sh.Title,
		TitleHeight:    sh.TitleHeight,
		Mode:           mode,
		TitleInsetLeft: sh.headerTitleInsetLeft(),
		TitleInsetRight: sh.headerTitleInsetRight(),
	}
}

func (sh *SurfaceShell) headerTitleInsetLeft() float32 {
	return surfaceHeaderInsetLeft(sh)
}

func (sh *SurfaceShell) headerTitleInsetRight() float32 {
	return surfaceHeaderInsetRight(sh)
}

func (sh *SurfaceShell) bodyTitleHeight() float32 {
	h := sh.surfaceHeader().Height()
	if pf := sh.panelFeatures; pf != nil {
		if ch := pf.ChromeHeight(); ch > h {
			return ch
		}
	}
	return h
}

// AttachBehavior adds a SurfaceBehavior plugin (C2+).
func (sh *SurfaceShell) AttachBehavior(b SurfaceBehavior) {
	if b == nil {
		return
	}
	if cb, ok := b.(*CollapseBehavior); ok && sh.collapse == nil {
		sh.collapse = cb
	}
	if pf, ok := b.(*PanelFeaturesBehavior); ok {
		sh.panelFeatures = pf
	}
	if db, ok := b.(*DismissBehavior); ok {
		sh.dismiss = db
	}
	if eb, ok := b.(*EscapeBehavior); ok {
		sh.escape = eb
	}
	b.AttachShell(sh)
	sh.behaviors = append(sh.behaviors, b)
	sh.MarkDirty()
}

func (sh *SurfaceShell) collapseBehavior() *CollapseBehavior {
	return sh.collapse
}

// Children returns user content in the body, not the internal body node.
func (sh *SurfaceShell) Children() []Node {
	if sh.body == nil {
		return nil
	}
	return sh.body.Children()
}

// RemoveChild removes a user content node from the body.
func (sh *SurfaceShell) RemoveChild(id string) {
	if sh.body == nil {
		return
	}
	sh.body.RemoveChild(id)
	sh.MarkDirty()
}

// ShiftScrollSubtreeInternal moves the internal RaisedSurface body and its
// descendants by dy without moving the shell bounds again. Viewport
// repositionOnly already updated the shell; Card/Panel body bands must follow
// or preview table headers paint at stale Y and bleed over chrome above.
func (sh *SurfaceShell) ShiftScrollSubtreeInternal(dy float32) {
	if dy == 0 || sh.body == nil {
		return
	}
	bb := sh.body.Bounds()
	bb.Y += dy
	scrollTranslate(sh.body, bb)
	shiftSubtreeY(sh.body.Children(), dy)
}

// InvalidateLayoutPassCache clears shell and body layout snapshots.
func (sh *SurfaceShell) InvalidateLayoutPassCache() {
	sh.Container.InvalidateLayoutPassCache()
	if sh.body != nil {
		sh.body.InvalidateLayoutPassCache()
	}
}

// layoutWalkChildren returns nodes for resize/dirty walks: internal body plus user content.
func (sh *SurfaceShell) layoutWalkChildren() []Node {
	if sh.body == nil {
		return nil
	}
	out := make([]Node, 0, 1+len(sh.body.children))
	out = append(out, sh.body)
	out = append(out, sh.body.children...)
	return out
}

// Update runs behavior plugins then body descendants.
func (sh *SurfaceShell) Update(dt float32) {
	if sh.IsHidden() {
		return
	}
	for _, b := range sh.behaviors {
		b.Update(dt)
	}
	UpdateChildrenOverlayAware(sh.Children(), dt)
}

func (sh *SurfaceShell) syncBodyFromShell() {
	if sh.body == nil {
		return
	}
	sh.syncStyleToBody()
	titleOff := sh.bodyTitleHeight()
	b := sh.bounds
	bodyH := b.Height - titleOff
	if cb := sh.collapse; cb != nil && cb.contentH > 0.5 {
		if sh.panelUserSized() && cb.visibleBodyH(sh) > 0.5 {
			bodyH = b.Height - titleOff
		} else {
			bodyH = cb.contentH
		}
	}
	if bodyH < 0 {
		bodyH = 0
	}
	newRect := rl.NewRectangle(b.X, b.Y+titleOff, b.Width, bodyH)
	if sh.body.bounds != newRect {
		sh.body.bounds = newRect
		sh.body.MarkDirty()
	}
	sh.body.Gap = sh.Gap
	sh.body.AutoHeight = sh.AutoHeight
	sh.body.MinWidth = sh.MinWidth
	sh.body.MaxWidth = sh.MaxWidth
	sh.body.PreferredWidth = sh.PreferredWidth
	sh.body.ClipChildren = sh.ClipChildren
}

func (sh *SurfaceShell) syncStyleToBody() {
	if sh.body == nil {
		return
	}
	b := sh.body
	b.styleName = sh.styleName
	b.styleComponent = sh.styleComponent
	b.styleVariant = sh.styleVariant
	b.styleOverrides = sh.styleOverrides
	b.stylePatch = sh.stylePatch
	b.presetName = sh.presetName
	b.presetGlowIntensity = sh.presetGlowIntensity
	b.presetGlowSet = sh.presetGlowSet
	b.presetHoverLift = sh.presetHoverLift
	b.styleDirty = true
	b.resolvedStyleValid = false
}

// Layout offsets the body band and delegates flex/clamp to RaisedSurface.
func (sh *SurfaceShell) Layout() {
	if sh.IsHidden() {
		return
	}
	geom := !sh.lastLayoutPassValid || sh.lastLayoutPassW != sh.bounds.Width || sh.lastLayoutPassH != sh.bounds.Height
	if geom {
		sh.layoutDirty = true
		for _, ch := range sh.Children() {
			if !ch.IsHidden() {
				ch.MarkDirty()
			}
		}
	}
	needsLayout := sh.IsDirty() || geom
	if !needsLayout {
		for _, ch := range sh.Children() {
			if !ch.IsHidden() && ch.IsDirty() {
				needsLayout = true
				break
			}
		}
	}
	if !needsLayout {
		return
	}

	origH := sh.bounds.Height
	sh.syncBodyFromShell()
	sh.body.Layout()

	for _, b := range sh.behaviors {
		b.LayoutAfterBody(sh)
	}

	if sh.AutoHeight && sh.GetFlexGrow() == 0 && !sh.panelUserSized() {
		titleOff := sh.bodyTitleHeight()
		bodyH := sh.body.Bounds().Height
		if cb := sh.collapse; cb != nil {
			bodyH = cb.visibleBodyH(sh)
		}
		intrinsic := titleOff + bodyH
		finalH := panelBodyFillHeight(sh, origH, intrinsic, sh.GetFlexGrow(), sh.Children())
		if sh.bounds.Height != finalH {
			sh.bounds.Height = finalH
			sh.syncBodyFromShell()
			sh.body.Layout()
		}
	}

	if pf := sh.panelFeatures; pf != nil {
		pf.layoutScrollViewport(sh)
	}

	sh.layoutDirty = false
	sh.lastLayoutPassW = sh.bounds.Width
	sh.lastLayoutPassH = sh.bounds.Height
	sh.lastLayoutPassValid = true
	syncLayoutExtent(sh)
}

// collapseChromeCollapsed is true when the body band is fully hidden.
func (sh *SurfaceShell) collapseChromeCollapsed() bool {
	cb := sh.collapse
	if cb == nil {
		return false
	}
	return cb.visibleBodyH(sh) < 0.5
}

func (sh *SurfaceShell) panelUserSized() bool {
	if pf := sh.panelFeatures; pf != nil {
		return pf.userSizedShell()
	}
	return false
}

// Draw paints chrome, header, then body children.
func (sh *SurfaceShell) Draw() {
	defer func() { sh.drawDirty = false }()
	if sh.IsHidden() {
		return
	}
	bounds := snapControlRect(sh.Bounds())
	if bounds.Width < 1 || bounds.Height < 1 {
		return
	}
	style := sh.GetStyle()
	header := sh.surfaceHeader()
	collapsedOnly := sh.collapseChromeCollapsed()
	header.CollapsedOnly = collapsedOnly
	deferGlassTitle := header.DefersUntilPostSheen()
	profile := ResolveChromeProfile(&sh.Element)
	chromeCtx := SurfaceChromeCtxFor(&sh.Element, bounds, style, nestedInRaisedSurface(sh))
	if collapsedOnly {
		chromeCtx.SkipOuterShadow = true
	}

	profile.DrawShadow(chromeCtx)

	fillBounds := bounds
	bw := style.BorderWidth
	// Fractional borders (e.g. panel 1.5) fringe under SSAA; snap to whole pixels.
	if bw > 0 {
		bw = float32(int32(bw + 0.5))
		if bw < 1 {
			bw = 1
		}
		fillBounds = chromeFillBounds(bounds, bw)
	}
	fillRoundness := chromeRoundness(fillBounds, style.CornerRadius)

	// Prefer seam-safe inset fill over DrawRectangleRounded — triangulation
	// leaves a vertical mid-seam under 2× SSAA on wide empty Card/Panel bodies
	// (calc Desk card). Compact controls still use drawRoundedInsetBorder.
	if bw > 0 && !collapsedOnly {
		drawSeamSafeInsetBorder(bounds, style.CornerRadius, bw, style.BorderColor, style.BackgroundColor)
	} else if style.CornerRadius > 0.5 {
		drawSeamSafeRoundedFill(fillBounds, style.CornerRadius, style.BackgroundColor)
	} else if fillRoundness > 0 {
		rl.DrawRectangleRounded(fillBounds, fillRoundness, 32, style.BackgroundColor)
	} else {
		rl.DrawRectangleRec(fillBounds, style.BackgroundColor)
	}

	if !deferGlassTitle {
		header.Draw(bounds, fillBounds, style, style.CornerRadius)
	}

	profile.DrawOverFill(chromeCtx)

	// Border already painted via inset path when BorderWidth > 0.
	if bw > 0 && collapsedOnly {
		borderBounds := chromeBorderBounds(bounds, bw)
		bodyR := chromeRoundness(borderBounds, style.CornerRadius)
		if bodyR > 0 {
			rl.DrawRectangleRoundedLinesEx(borderBounds, bodyR, 32, bw, style.BorderColor)
		} else {
			rl.DrawRectangleLinesEx(borderBounds, bw, style.BorderColor)
		}
	}

	profile.DrawPostBorder(chromeCtx)

	if deferGlassTitle {
		header.DrawGlass(bounds, style)
	}

	for _, b := range sh.behaviors {
		b.DrawOverlay(sh)
	}

	if sh.body != nil {
		if cb := sh.collapse; cb != nil {
			titleOff := sh.bodyTitleHeight()
			visible := cb.visibleBodyH(sh)
			if visible > 0 && visible < cb.contentH-0.5 {
				clip := rl.NewRectangle(bounds.X, bounds.Y+titleOff, bounds.Width, visible)
				rl.BeginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
				sh.body.Draw()
				rl.EndScissorMode()
			} else if visible > 0 {
				sh.body.Draw()
			}
		} else {
			sh.body.Draw()
		}
	}
}

func (sh *SurfaceShell) IsInteractive() bool {
	if pf := sh.panelFeatures; pf != nil && pf.IsInteractive() {
		return true
	}
	if d := sh.dismiss; d != nil && d.HeaderInteractive() {
		return true
	}
	return sh.collapse != nil && !sh.collapse.ExternalHeader
}

func (sh *SurfaceShell) applySurfaceBodyTypographyToChild(child Node) {
	if !sh.surfaceUsesTintedBodyText() {
		return
	}
	st := sh.GetStyle()
	hint := bodyTypographyHintFromChrome(st)
	if !surfaceTypographyHintActive(hint) {
		return
	}
	applySurfaceBodyTypography(child, hint, chromeStyleIsDark(st))
}

func (sh *SurfaceShell) applySurfaceBodyTypographyToChildren() {
	if !sh.surfaceUsesTintedBodyText() {
		return
	}
	st := sh.GetStyle()
	hint := bodyTypographyHintFromChrome(st)
	if !surfaceTypographyHintActive(hint) {
		return
	}
	dark := chromeStyleIsDark(st)
	for _, ch := range sh.Children() {
		applySurfaceBodyTypography(ch, hint, dark)
	}
}

func (sh *SurfaceShell) surfaceUsesTintedBodyText() bool {
	return sh.Element.surfaceUsesTintedBodyText()
}
