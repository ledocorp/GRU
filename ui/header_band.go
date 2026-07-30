// Package ui (continued) — standalone title strip (Phase C2).
//
// HeaderBand is an independent box: title chrome only, no RaisedSurface body.
// Use it alone in a grid cell, above a sibling panel, or with BindCollapse to
// drive a collapsible Panel's Expanded signal from a separate header row.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const headerBandRibbonH = float32(36)

// HeaderBand is a standalone surface title strip (not a Panel body host).
type HeaderBand struct {
	Element
	Title       string
	TitleHeight float32
	Mode        SurfaceHeaderMode
	ShowChevron bool
	collapse    *CollapseBehavior
	hoverHeader bool
	ribbon      *Container
}

// NewHeaderBand creates a standalone title strip at the given width.
func NewHeaderBand(id, title string, x, y, w float32) *HeaderBand {
	titleH := float32(40)
	hb := &HeaderBand{
		Element:     NewElement(id, x, y, w, titleH),
		Title:       title,
		TitleHeight: titleH,
		Mode:        HeaderModeTitleBar,
	}
	hb.styleName = "panel"
	hb.Element.SetStyleVariant("panel", "default")
	return hb
}

// SetHeaderMode changes inset/title-bar/glass/none presentation.
func (hb *HeaderBand) SetHeaderMode(mode SurfaceHeaderMode) {
	hb.Mode = mode
	hb.MarkDrawDirty()
}

// Ribbon returns an optional toolbar row below the title (ribbon-style controls).
func (hb *HeaderBand) Ribbon() *Container {
	if hb.ribbon == nil {
		hb.ribbon = NewContainer(hb.ID()+"-ribbon", 0, 0, 0, headerBandRibbonH)
		hb.ribbon.FlexDirection = FlexRow
		hb.ribbon.Gap = 8
		hb.ribbon.SetStyle("transparent")
		hb.MarkDirty()
	}
	return hb.ribbon
}

// BindCollapse links this band's chevron and clicks to a panel's CollapseBehavior.
func (hb *HeaderBand) BindCollapse(panel *Panel) *CollapseBehavior {
	if panel == nil {
		return nil
	}
	cb := panel.collapseBehavior()
	if cb == nil {
		cb = panel.EnableCollapse(true)
	}
	hb.collapse = cb
	hb.ShowChevron = true
	cb.ExternalHeader = true
	return cb
}

// BindCollapseBehavior shares an existing collapse plugin.
func (hb *HeaderBand) BindCollapseBehavior(cb *CollapseBehavior) {
	hb.collapse = cb
	hb.ShowChevron = cb != nil
	if cb != nil {
		cb.ExternalHeader = true
	}
}

func (hb *HeaderBand) header() SurfaceHeader {
	return SurfaceHeader{
		Title:       hb.Title,
		TitleHeight: hb.TitleHeight,
		Mode:        hb.Mode,
		FlatBottom:  true,
	}
}

func (hb *HeaderBand) totalHeight() float32 {
	h := hb.header().Height()
	if h <= 0 {
		h = hb.TitleHeight
	}
	if hb.ribbon != nil && len(hb.ribbon.Children()) > 0 {
		h += headerBandRibbonH
	}
	return h
}

func (hb *HeaderBand) IsInteractive() bool {
	return hb.collapse != nil || hb.ribbon != nil
}

func (hb *HeaderBand) Update(dt float32) {
	if hb.IsHidden() {
		return
	}
	if hb.ribbon != nil {
		for _, ch := range hb.ribbon.Children() {
			if !ch.IsHidden() {
				ch.Update(dt)
			}
		}
	}
	if hb.collapse == nil {
		return
	}
	mouse := rl.GetMousePosition()
	hr := rl.NewRectangle(hb.bounds.X, hb.bounds.Y, hb.bounds.Width, hb.header().Height())
	prev := hb.hoverHeader
	hb.hoverHeader = rl.CheckCollisionPointRec(mouse, hr)
	if hb.hoverHeader != prev {
		hb.MarkDrawDirty()
	}
	if hb.hoverHeader && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		hb.collapse.Toggle()
	}
}

func (hb *HeaderBand) Layout() {
	if hb.IsHidden() {
		return
	}
	totalH := hb.totalHeight()
	if hb.bounds.Height != totalH {
		hb.bounds.Height = totalH
	}
	if hb.ribbon != nil && len(hb.ribbon.Children()) > 0 {
		titleH := hb.header().Height()
		pad := hb.GetStyle().Padding
		if pad <= 0 {
			pad = 12
		}
		innerW := hb.bounds.Width - 2*pad
		if innerW < 1 {
			innerW = hb.bounds.Width
		}
		rb := rl.NewRectangle(hb.bounds.X+pad, hb.bounds.Y+titleH, innerW, headerBandRibbonH)
		hb.ribbon.SetBounds(rb)
		hb.ribbon.Layout()
	}
	hb.layoutDirty = false
	syncLayoutExtent(hb)
}

func (hb *HeaderBand) Draw() {
	if hb.IsHidden() {
		return
	}
	bounds := hb.Bounds()
	if bounds.Width < 1 || bounds.Height < 1 {
		return
	}
	style := hb.GetStyle()
	header := hb.header()
	titleH := header.Height()
	fillBounds := bounds
	if style.BorderWidth > 0 {
		fillBounds = chromeFillBounds(bounds, style.BorderWidth)
	}
	chromeCtx := SurfaceChromeCtxFor(&hb.Element, bounds, style, true)
	fillRoundness := chromeRoundness(fillBounds, style.CornerRadius)

	drawFlatBottomChromeFill(fillBounds, style.CornerRadius, style.BackgroundColor)
	header.Draw(bounds, fillBounds, style, style.CornerRadius)

	if hb.ribbon != nil && len(hb.ribbon.Children()) > 0 {
		ribbonRect := rl.NewRectangle(bounds.X, bounds.Y+titleH, bounds.Width, headerBandRibbonH)
		ribbonFill := ribbonRect
		if style.BorderWidth > 0 {
			ribbonFill = chromeFillBounds(ribbonRect, style.BorderWidth)
		}
		rl.DrawRectangleRec(ribbonFill, style.BackgroundColor)
		sepY := int32(ribbonRect.Y)
		rl.DrawLine(int32(bounds.X), sepY, int32(bounds.X+bounds.Width), sepY, rl.NewColor(0, 0, 0, 28))
		hb.ribbon.Draw()
	}

	if style.BorderWidth > 0 {
		drawFlatBottomChromeBorder(bounds, style.BorderWidth, style.CornerRadius, style.BorderColor)
		if fillRoundness > 0 {
			profile := ResolveChromeProfile(&hb.Element)
			postCtx := chromeCtx
			postCtx.SkipOuterShadow = true
			profile.DrawPostBorder(postCtx)
		}
	}

	if hb.ShowChevron && hb.collapse != nil {
		progress := hb.collapse.chevronProgress()
		if progress >= 0 {
			col := GetThemeStyle("panel-title").TextColor
			if hb.hoverHeader {
				col = brightenColor(col, 20)
			}
			drawSurfaceChevron(bounds.X+bounds.Width-28, bounds.Y+titleH/2, 10, progress, col)
		}
	}
}

// Children implements ribbon child walk for hit testing.
func (hb *HeaderBand) Children() []Node {
	if hb.ribbon == nil {
		return nil
	}
	return hb.ribbon.Children()
}
