//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &responsiveScene{} }) }

// responsiveScene teaches the layout stack:
//
//	Page shell → main Viewport → page LayoutGrid → Panels → preview frames → inner layouts
//
// Two independent width axes:
//   - Resize the window → page-grid ColSpan on each panel (breakpoint from window width).
//   - Move the preview slider → shared preview width for inner grid / flex demos only.
type responsiveScene struct {
	BaseScene
	doc           *ui.Document
	pageGrid      *ui.Container
	widthSlider   *ui.Slider
	windowBpLabel *ui.RichText
	windowBpText  *ui.Signal[string]
	simBpLabel    *ui.RichText
	simBpText     *ui.Signal[string]
	widthLabel    *ui.RichText
	widthText     *ui.Signal[string]
	simWidth      float32
	previewFrames      []*ui.Card
	previewScrollLanes []*ui.Container // fixed simWidth; scroll when window narrower
	previewScrollVps   []*ui.Viewport
	innerGrid          *ui.Container
	scenePanels        []*ui.Panel
	pendingPreviewRelayout bool
	previewRelayoutSkip    int
	lastWindowBp           ui.Breakpoint
}

func (s *responsiveScene) Title() string { return "Responsive - Breakpoints - Grid" }

func (s *responsiveScene) OnUpdate(_ *ui.Document, _ float32) {
	if !s.pendingPreviewRelayout {
		return
	}
	if s.widthSlider != nil && s.widthSlider.IsDragging() {
		s.previewRelayoutSkip++
		if s.previewRelayoutSkip%5 != 0 {
			return
		}
	}
	s.pendingPreviewRelayout = false
	s.previewRelayoutSkip = 0
	s.relayoutPreviewSubtree(s.simWidth)
}

func setSpans5ResponsivePanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func responsiveHint(id, text string) *ui.RichText {
	return FlexCopy(id, "form-value", text)
}

func (s *responsiveScene) trackPanel(p *ui.Panel) *ui.Panel {
	s.scenePanels = append(s.scenePanels, p)
	return p
}

func responsivePanel(s *responsiveScene, id, title string) *ui.Panel {
	p := ui.NewPanel(id, title, 0, 0, 0, 0)
	p.AutoHeight = true
	p.Gap = 10
	p.TitleHeight = 36
	return s.trackPanel(p)
}

func tierLabel(id, text string) (*ui.RichText, *ui.Signal[string]) {
	return FlexCopyPair(id, "form-value", text)
}

func breakpointLabel(width float32) string {
	if width < float32(ui.MinClientWidth) {
		return ui.BreakpointXS.String() + " (clamped)"
	}
	return ui.CurrentBreakpoint(width).String()
}

func (s *responsiveScene) newPreviewChromeFrame(id string, w float32, caption string) *ui.Card {
	chrome := ui.NewCard(id, "", 0, 0, 0, 0)
	chrome.AutoHeight = true
	chrome.FlexDirection = ui.FlexColumn
	chrome.Gap = 8
	chrome.ClipChildren = true

	capLbl := FlexCopy(id+"-cap", "form-label", caption)
	chrome.AddChild(capLbl)

	s.previewFrames = append(s.previewFrames, chrome)
	return chrome
}

// Preview stack: Card frame → Viewport (visible width) → Container lane (@ simWidth)
// → inner layout. Horizontal scroll appears when the window is narrower than simWidth.

const (
	previewLanePad        float32 = 14
	previewLaneBottomBand float32 = 12 // below grid / flex row so card bottom borders clear the viewport clip
)

// previewChromeScroll — lane stays simWidth px; scroll inside the frame when the window is narrower.
// Returns the content host (add grid/flex children here); the lane adds bottom inset below the host.
func (s *responsiveScene) previewChromeScroll(id string, w float32) (*ui.Card, *ui.Container) {
	chrome := s.newPreviewChromeFrame(id, w,
		fmt.Sprintf("Preview layout width: %.0f px (scroll when window is narrower)", w))

	scroll := ui.NewHorizontalViewport(id+"-scroll", 0, 0, 0, 0)
	scroll.AutoHeight = true
	scroll.Gap = 0
	scroll.SetStyle("list-flush")

	lane := ui.NewContainer(id+"-lane", 0, 0, w, 0)
	lane.AutoHeight = true
	lane.FlexDirection = ui.FlexColumn
	lane.PreferredWidth = w
	lane.MinWidth = w
	lane.MaxWidth = w
	lane.ClipChildren = true
	lane.SetStyle("transparent")
	lane.SetStyleOverrides(ui.Style{Padding: previewLanePad})

	host := ui.NewContainer(id+"-host", 0, 0, 0, 0)
	host.AutoHeight = true
	host.FlexDirection = ui.FlexColumn
	host.ClipChildren = true
	host.SetStyle("transparent")

	bottomInset := ui.NewLabel(id+"-bottom-inset", "", 0, 0, 0, previewLaneBottomBand)
	bottomInset.SetStyle("transparent")

	lane.AddChild(host)
	lane.AddChild(bottomInset)
	scroll.AddChild(lane)
	chrome.AddChild(scroll)

	s.previewScrollLanes = append(s.previewScrollLanes, lane)
	s.previewScrollVps = append(s.previewScrollVps, scroll)
	return chrome, host
}

func (s *responsiveScene) markPreviewSubtreeDirty() {
	for _, frame := range s.previewFrames {
		ui.MarkResizeLayoutDirtySubtree(frame)
		frame.InvalidateLayoutPassCache()
	}
	for _, lane := range s.previewScrollLanes {
		lane.InvalidateLayoutPassCache()
		ui.MarkResizeLayoutDirtySubtree(lane)
	}
	for _, vp := range s.previewScrollVps {
		vp.InvalidateLayoutPassCache()
		ui.MarkResizeLayoutDirtySubtree(vp)
	}
	if s.innerGrid != nil {
		s.innerGrid.InvalidateLayoutPassCache()
		ui.MarkResizeLayoutDirtySubtree(s.innerGrid)
	}
}

func (s *responsiveScene) updatePreviewLabels(w float32) {
	s.simWidth = w
	if s.widthText != nil {
		s.widthText.Set(fmt.Sprintf("%.0f px", w))
	}
	if s.simBpText != nil {
		s.simBpText.Set(breakpointLabel(w))
	}
	for _, frame := range s.previewFrames {
		if len(frame.Children()) == 0 {
			continue
		}
		capLbl, ok := frame.Children()[0].(*ui.RichText)
		if !ok {
			continue
		}
		capLbl.SetSpans([]ui.TextSpan{{
			Text: fmt.Sprintf("Preview layout width: %.0f px (scroll when window is narrower)", w),
		}})
		capLbl.InvalidateAutoHeightMeasure()
		capLbl.MarkDirty()
	}
}

func (s *responsiveScene) relayoutPreviewSubtree(w float32) {
	for _, lane := range s.previewScrollLanes {
		lane.PreferredWidth = w
		lane.MinWidth = w
		lane.MaxWidth = w
		b := lane.Bounds()
		if b.Width != w {
			b.Width = w
			lane.SetBounds(b)
		}
	}
	s.markPreviewSubtreeDirty()
	s.markScenePanelsDirty()
	if s.pageGrid != nil {
		s.pageGrid.InvalidateLayoutPassCache()
		ui.MarkResizeLayoutDirtySubtree(s.pageGrid)
	}
	if s.doc != nil && s.doc.Root != nil {
		ui.MarkResizeLayoutDirtySubtree(s.doc.Root)
	}
}

func (s *responsiveScene) applyPreviewWidth(w float32) {
	s.updatePreviewLabels(w)
	s.relayoutPreviewSubtree(w)
}

func (s *responsiveScene) onPreviewSliderChanged(w float32) {
	s.updatePreviewLabels(w)
	s.relayoutPreviewSubtree(w)
	s.pendingPreviewRelayout = false
}

func (s *responsiveScene) markScenePanelsDirty() {
	for _, p := range s.scenePanels {
		p.InvalidateLayoutPassCache()
		p.MarkDirty()
	}
}

func (s *responsiveScene) Build(doc *ui.Document) {
	const startW float32 = 960
	s.doc = doc
	s.simWidth = startW
	s.previewFrames = s.previewFrames[:0]
	s.previewScrollLanes = s.previewScrollLanes[:0]
	s.previewScrollVps = s.previewScrollVps[:0]
	s.scenePanels = s.scenePanels[:0]
	s.innerGrid = nil

	page := MountAppPage(doc, "resp",
		"Responsive Layout",
		"Resize the window to reflow panels on the page grid. Use the slider to change the width of the inner preview containers only.")
	page.Body.Gap = 12

	s.pageGrid = ui.NewContainer("resp-page-grid", 0, 0, 0, 0)
	s.pageGrid.LayoutType = ui.LayoutGrid
	s.pageGrid.GridColumns = 12
	s.pageGrid.Gap = 12
	s.pageGrid.SetStyle("page-grid")
	s.pageGrid.SetFlexGrow(1)

	s.pageGrid.AddChild(s.buildControlsPanel(doc))
	s.pageGrid.AddChild(s.buildGridPanel())
	s.pageGrid.AddChild(s.buildFlexGrowPanel())
	s.pageGrid.AddChild(s.buildResponsiveFlexPanel())
	s.pageGrid.AddChild(s.buildHorizontalScrollPanel())

	page.Body.AddChild(s.pageGrid)

	s.lastWindowBp = ui.ActiveBreakpoint.Get()
	s.applyPreviewWidth(startW)
}

func (s *responsiveScene) buildControlsPanel(doc *ui.Document) *ui.Panel {
	p := responsivePanel(s, "resp-controls-panel", "1 · Controls (connects to previews below)")
	setSpans5ResponsivePanel(p, 12, 12, 12, 12, 12)
	p.AddChild(responsiveHint("resp-controls-hint",
		"Window tier comes from the live client width (resize the window). Preview tier and all panels below use the slider width only."))

	readout := ui.NewContainer("resp-readout-row", 0, 0, 0, 0)
	readout.AutoHeight = true
	readout.FlexDirection = ui.FlexRow
	readout.Gap = 24
	readout.SetStyle("transparent")

	winCol := ui.NewContainer("resp-win-col", 0, 0, 0, 0)
	winCol.AutoHeight = true
	winCol.FlexDirection = ui.FlexColumn
	winCol.Gap = 4
	winCol.SetFlexGrow(1)
	winCol.SetStyle("transparent")
	winCol.AddChild(FlexCopy("resp-win-key", "form-label", "Window tier (page grid)"))
	s.windowBpLabel, s.windowBpText = tierLabel("resp-win-val", ui.ActiveBreakpoint.Get().String())
	winCol.AddChild(s.windowBpLabel)

	simCol := ui.NewContainer("resp-sim-col", 0, 0, 0, 0)
	simCol.AutoHeight = true
	simCol.FlexDirection = ui.FlexColumn
	simCol.Gap = 4
	simCol.SetFlexGrow(1)
	simCol.SetStyle("transparent")
	simCol.AddChild(FlexCopy("resp-sim-key", "form-label", "Preview tier (inner layouts)"))
	s.simBpLabel, s.simBpText = tierLabel("resp-sim-val", breakpointLabel(s.simWidth))
	simCol.AddChild(s.simBpLabel)

	readout.AddChild(winCol)
	readout.AddChild(simCol)
	p.AddChild(readout)

	sliderRow := ui.NewContainer("resp-slider-row", 0, 0, 0, 0)
	sliderRow.AutoHeight = true
	sliderRow.FlexDirection = ui.FlexRow
	sliderRow.Gap = 12
	sliderRow.SetStyle("transparent")
	slider := ui.NewSlider("resp-slider", 420, 1400, s.simWidth, 0, 0, 0, 32)
	slider.SetFlexGrow(1)
	slider.ShowValue = false
	s.widthSlider = slider
	s.widthLabel, s.widthText = FlexCopyPair("resp-width-label", "form-label", fmt.Sprintf("%.0f px", s.simWidth))
	sliderRow.AddChild(slider)
	sliderRow.AddChild(s.widthLabel)
	p.AddChild(sliderRow)

	ui.ActiveBreakpoint.Subscribe(func() {
		bp := ui.ActiveBreakpoint.Get()
		if s.windowBpText != nil {
			s.windowBpText.Set(bp.String())
		}
		if bp == s.lastWindowBp {
			return
		}
		s.lastWindowBp = bp
		if s.pageGrid != nil {
			s.pageGrid.InvalidateLayoutPassCache()
			s.pageGrid.MarkDirty()
		}
		s.markScenePanelsDirty()
		if s.doc != nil && s.doc.Root != nil {
			s.doc.Root.MarkDirty()
		}
	})
	slider.Value.Subscribe(func() {
		s.onPreviewSliderChanged(slider.Value.Get())
	})

	return p
}

func (s *responsiveScene) buildGridPanel() *ui.Panel {
	p := responsivePanel(s, "resp-grid-panel", "2 · 12-Column Grid (ColSpan)")
	setSpans5ResponsivePanel(p, 12, 12, 12, 12, 12)
	p.AddChild(responsiveHint("resp-grid-hint",
		"ColSpan tiers xs12/sm6/md4/lg3. Slider = layout width inside the lane; resize the window narrower to scroll inside the preview card (panel 5 shows scroll-only tiles)."))

	chrome, host := s.previewChromeScroll("resp-grid-chrome", s.simWidth)
	inner := ui.NewContainer("resp-card-grid", 0, 0, 0, 0)
	inner.LayoutType = ui.LayoutGrid
	inner.GridColumns = 12
	inner.Gap = 10
	inner.SetStyle("transparent")
	inner.SetStyleOverrides(ui.Style{Padding: 10})
	inner.ClipChildren = true
	s.innerGrid = inner

	titles := []string{"One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight"}
	for i, title := range titles {
		card := s.makeMetricCard(fmt.Sprintf("resp-grid-card-%d", i), title,
			"xs12 sm6 md4 lg3")
		card.SetColSpan(ui.BreakpointXS, 12)
		card.SetColSpan(ui.BreakpointSM, 6)
		card.SetColSpan(ui.BreakpointMD, 4)
		card.SetColSpan(ui.BreakpointLG, 3)
		card.SetColSpan(ui.BreakpointXL, 3)
		inner.AddChild(card)
	}
	host.AddChild(inner)
	p.AddChild(chrome)
	return p
}

func (s *responsiveScene) buildFlexGrowPanel() *ui.Panel {
	p := responsivePanel(s, "resp-flexgrow-panel", "3 · Flex Row + FlexGrow Cards")
	setSpans5ResponsivePanel(p, 12, 12, 12, 6, 6)
	p.AddChild(responsiveHint("resp-flexgrow-hint",
		"Explicit FlexRow (not LayoutResponsive). Each card has FlexGrow=1 and shares the slider layout width; resize the window narrower to scroll inside the preview card."))

	chrome, host := s.previewChromeScroll("resp-flexgrow-chrome", s.simWidth)
	host.FlexDirection = ui.FlexRow
	host.Gap = 10

	for i, title := range []string{"Alpha", "Beta", "Gamma"} {
		card := ui.NewCard(fmt.Sprintf("resp-flexgrow-card-%d", i), "", 0, 0, 0, 88)
		card.AutoHeight = false
		card.SetFlexGrow(1)
		card.FlexDirection = ui.FlexColumn
		card.Gap = 6
		card.ClipChildren = true
		tl := FlexCopy(fmt.Sprintf("resp-flexgrow-title-%d", i), "form-value", title)
		sub := FlexCopy(fmt.Sprintf("resp-flexgrow-sub-%d", i), "form-label", "FlexGrow 1")
		card.AddChild(tl)
		card.AddChild(sub)
		host.AddChild(card)
	}
	p.AddChild(chrome)
	return p
}

func (s *responsiveScene) buildResponsiveFlexPanel() *ui.Panel {
	p := responsivePanel(s, "resp-responsive-panel", "4 · LayoutResponsive (≥600 row)")
	setSpans5ResponsivePanel(p, 12, 12, 12, 6, 6)
	p.AddChild(responsiveHint("resp-responsive-hint",
		"LayoutResponsive switches row/column at 600 px on the slider layout width. Resize the window narrower to scroll inside the preview card."))

	chrome, host := s.previewChromeScroll("resp-responsive-chrome", s.simWidth)
	resp := ui.NewContainer("resp-responsive-inner", 0, 0, 0, 0)
	resp.AutoHeight = true
	resp.LayoutType = ui.LayoutResponsive
	resp.Gap = 10
	resp.ClipChildren = true
	resp.SetStyle("transparent")

	for i, title := range []string{"North", "South", "East"} {
		card := ui.NewCard(fmt.Sprintf("resp-responsive-card-%d", i), "", 0, 0, 0, 72)
		card.AutoHeight = false
		card.SetFlexGrow(1)
		card.FlexDirection = ui.FlexColumn
		card.Gap = 4
		card.ClipChildren = true
		card.AddChild(FlexCopy(fmt.Sprintf("resp-responsive-title-%d", i), "form-value", title))
		card.AddChild(FlexCopy(fmt.Sprintf("resp-responsive-sub-%d", i), "form-label", "LayoutResponsive"))
		resp.AddChild(card)
	}
	host.AddChild(resp)
	p.AddChild(chrome)
	return p
}

func (s *responsiveScene) buildHorizontalScrollPanel() *ui.Panel {
	p := responsivePanel(s, "resp-scroll-panel", "5 · Horizontal Viewport (panel width)")
	setSpans5ResponsivePanel(p, 12, 12, 12, 12, 12)
	p.AddChild(responsiveHint("resp-scroll-hint",
		"Uses the full panel body width (follows window/page grid). Wheel over the strip or drag the bottom scrollbar; cards keep fixed width."))

	hv := ui.NewHorizontalViewport("resp-hscroll", 0, 0, 0, 0)
	hv.AutoHeight = true
	hv.Gap = 8
	hv.SetStyle("list-flush")
	for i := 0; i < 8; i++ {
		hv.AddChild(s.makeHScrollCard(fmt.Sprintf("resp-h-card-%d", i), fmt.Sprintf("Tile %d", i+1)))
	}
	p.AddChild(hv)
	return p
}

func (s *responsiveScene) makeMetricCard(id, title, note string) *ui.Card {
	c := ui.NewCard(id, "", 0, 0, 0, 0)
	c.AutoHeight = true
	c.FlexDirection = ui.FlexColumn
	c.Gap = 4
	c.ClipChildren = true
	c.AddChild(FlexCopy(id+"-title", "form-value", title))
	c.AddChild(FlexCopy(id+"-note", "form-label", note))
	return c
}

func (s *responsiveScene) makeHScrollCard(id, label string) *ui.Card {
	c := ui.NewCard(id, "", 0, 0, 140, 96)
	c.AutoHeight = false
	c.FlexDirection = ui.FlexColumn
	c.Gap = 4
	c.ClipChildren = true
	c.AddChild(FlexCopy(id+"-label", "form-value", label))
	c.AddChild(FlexCopy(id+"-sub", "form-label", "140px wide"))
	return c
}

func (s *responsiveScene) Destroy() {}
