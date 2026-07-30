// Package ui (continued) — dumb raised surface body (Phase C1.1).
//
// RaisedSurface runs flex/clamp/scissor only. Chrome and headers live on
// SurfaceShell. Panel and Card are facades over SurfaceShell.
package ui

import (
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// RaisedSurface is the dumb body box: flex layout, clamp, scissor, child draw.
type RaisedSurface struct {
	Container
}

func newRaisedSurfaceBody(id string) *RaisedSurface {
	c := NewContainer(id, 0, 0, 0, 0)
	c.SetStyle("transparent")
	return &RaisedSurface{Container: *c}
}

// GetStyle returns the parent shell style when embedded in SurfaceShell.
func (s *RaisedSurface) GetStyle() Style {
	if sh := shellForBody(s); sh != nil {
		return sh.GetStyle()
	}
	return s.Container.GetStyle()
}

func shellForBody(s *RaisedSurface) *SurfaceShell {
	for p := s.ParentNode(); p != nil; p = p.ParentNode() {
		switch v := p.(type) {
		case *SurfaceShell:
			return v
		case *Panel:
			return &v.SurfaceShell
		case *Card:
			return &v.SurfaceShell
		}
	}
	return nil
}

// AddChild sets parent and syncs preset body typography to direct children.
func (s *RaisedSurface) AddChild(child Node) {
	s.children = append(s.children, child)
	child.SetParent(s)
	s.applySurfaceBodyTypographyToChild(child)
	s.MarkDirty()
}

// Layout positions children within the body band and recurses.
func (s *RaisedSurface) Layout() {
	if s.IsHidden() {
		return
	}
	geom := !s.lastLayoutPassValid || s.lastLayoutPassW != s.bounds.Width || s.lastLayoutPassH != s.bounds.Height
	if geom {
		s.layoutDirty = true
		for _, ch := range s.children {
			if !ch.IsHidden() {
				ch.MarkDirty()
			}
		}
	}
	needsLayout := s.IsDirty() || geom
	if !needsLayout {
		for _, ch := range s.children {
			if !ch.IsHidden() && ch.IsDirty() {
				needsLayout = true
				break
			}
		}
	}
	if !needsLayout {
		return
	}
	s.layoutContent()
	for _, ch := range s.children {
		if ch.IsHidden() {
			continue
		}
		ch.Layout()
	}
	style := s.GetStyle()
	padding := style.Padding
	fixedBody := !s.AutoHeight || s.GetFlexGrow() > 0
	if s.AutoHeight && s.GetFlexGrow() == 0 {
		finalizeAutoHeightFromSubtree(
			s,
			&s.bounds,
			0,
			padding,
			s.children,
			s.GetFlexGrow(),
			s.Container.layoutFlexInRect,
			func() { layoutBodyLabels(s.children, false) },
			s.clampChildrenToBody,
		)
	}
	bodyRect := rl.NewRectangle(s.bounds.X, s.bounds.Y, s.bounds.Width, s.bounds.Height)
	clampBodyChildren(bodyRect, padding, s.children, fixedBody)
	s.layoutDirty = false
	s.lastLayoutPassW = s.bounds.Width
	s.lastLayoutPassH = s.bounds.Height
	s.lastLayoutPassValid = true
	syncLayoutExtent(s)
}

func (s *RaisedSurface) layoutContent() {
	style := s.GetStyle()
	padding := style.Padding

	if len(s.children) == 0 {
		if s.AutoHeight {
			s.bounds.Height = 2 * padding
		}
		return
	}

	origY := s.bounds.Y
	origH := s.bounds.Height
	bodyH := origH
	if bodyH < 0 {
		bodyH = 0
	}

	measureBodyH := bodyH
	intrinsic := s.AutoHeight && s.GetFlexGrow() == 0
	innerW := s.bounds.Width - 2*padding
	if innerW < 1 {
		innerW = s.bounds.Width
	}
	// Width-first: fit labels before flex when cell width is known (Phase D1).
	if intrinsic && innerW > 8 {
		prepareBodySubtreeWidths(innerW, s.children)
	}
	// Only use the tall probe band when height is still unknown (h=0). A grid or
	// flex parent that assigned a definite height must lay out at that band so
	// flex-grow SplitView / Viewport children fill vertically on resize.
	if intrinsic && origH < 1 {
		if innerW > 8 {
			measureBodyH = flexIntrinsicProbeH
		} else {
			measureBodyH = 4096
		}
	}
	bodyRect := rl.NewRectangle(s.bounds.X, origY, s.bounds.Width, measureBodyH)
	s.Container.layoutFlexInRect(bodyRect)
	fixedBody := measureBodyH < flexIntrinsicProbeH && (!s.AutoHeight || origH >= 1)
	layoutBodyLabels(s.children, fixedBody)

	if s.AutoHeight {
		contentEnd := bodyRect.Y
		if len(s.children) > 0 {
			for _, ch := range s.children {
				if ch.IsHidden() {
					continue
				}
				bottom := nodeSubtreeBottom(ch)
				if bottom > contentEnd {
					contentEnd = bottom
				}
			}
			contentEnd += padding
		} else {
			contentEnd = bodyRect.Y + padding
		}
		bodyUsed := contentEnd - bodyRect.Y
		if bodyUsed < 0 {
			bodyUsed = 0
		}
		intrinsicH := bodyUsed
		finalH := panelBodyFillHeight(s, origH, intrinsicH, s.GetFlexGrow(), s.children)
		s.bounds.Y = origY
		s.bounds.Height = finalH
		finalBodyRect := rl.NewRectangle(s.bounds.X, origY, s.bounds.Width, finalH)
		s.Container.layoutFlexInRect(finalBodyRect)
	} else {
		s.bounds.Y = origY
		s.bounds.Height = origH
	}
}

func (s *RaisedSurface) clampChildrenToBody(bodyRect rl.Rectangle, padding float32) {
	clampChildrenToBodyRect(s.children, bodyRect, padding)
}

// Draw renders body children inside the padded clip. No chrome or header.
func (s *RaisedSurface) Draw() {
	defer func() { s.drawDirty = false }()
	if s.IsHidden() {
		return
	}
	bounds := s.Bounds()
	if bounds.Width < 1 || bounds.Height < 1 {
		return
	}
	style := s.GetStyle()
	sorted := make([]Node, len(s.children))
	copy(sorted, s.children)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetZIndex() < sorted[j].GetZIndex()
	})
	pad := style.Padding
	bodyClip := panelBodyDrawClipRect(bounds, pad, sorted)
	drawChildrenInPaddedBodyClip(s, bodyClip, sorted)
}

func (s *RaisedSurface) IsInteractive() bool { return false }

func (s *RaisedSurface) applySurfaceBodyTypographyToChild(child Node) {
	sh := shellForBody(s)
	if sh == nil || !sh.surfaceUsesTintedBodyText() {
		return
	}
	st := sh.GetStyle()
	hint := bodyTypographyHintFromChrome(st)
	if !surfaceTypographyHintActive(hint) {
		return
	}
	applySurfaceBodyTypography(child, hint, chromeStyleIsDark(st))
}

// nestedInRaisedSurface is true when n sits inside another raised surface body.
func nestedInRaisedSurface(n Node) bool {
	if rs, ok := n.(*RaisedSurface); ok {
		if sh := shellForBody(rs); sh != nil {
			return nestedInRaisedSurfaceShell(sh)
		}
		return false
	}
	return nestedInRaisedSurfaceShell(n)
}

func nestedInRaisedSurfaceShell(n Node) bool {
	for p := n.ParentNode(); p != nil; p = p.ParentNode() {
		switch p.(type) {
		case *Panel, *Card, *SurfaceShell:
			return true
		}
	}
	return false
}

// layoutBodyLabels runs wrap detection and intrinsic Label.Layout for raised surface bodies.
func layoutBodyLabels(children []Node, fixedBody bool) {
	fitPanelBodyLabels(children, fixedBody)
	for _, ch := range children {
		if !ch.IsHidden() {
			ch.Layout()
		}
	}
}

// fitSubtreeLabels enables wrap/truncate on labels that exceed the given width.
func fitSubtreeLabels(width float32, nodes []Node) {
	if width < 8 {
		return
	}
	for _, ch := range nodes {
		if ch.IsHidden() {
			continue
		}
		if lbl, ok := ch.(*Label); ok {
			b := lbl.Bounds()
			if b.Width == 0 || b.Width > width {
				b.Width = width
				lbl.SetBounds(b)
			}
			lbl.ensureWrapForWidth(width)
			continue
		}
		if rt, ok := ch.(*RichText); ok {
			b := rt.Bounds()
			if b.Width == 0 || b.Width > width {
				b.Width = width
				rt.SetBounds(b)
				rt.InvalidateAutoHeightMeasure()
			}
			if !rt.Wrap && styleUsesFlexPlainText(rt.styleName) {
				rt.Wrap = true
				rt.InvalidateAutoHeightMeasure()
				rt.MarkDirty()
			}
			continue
		}
		if kids := ch.Children(); len(kids) > 0 {
			fitSubtreeLabels(width, kids)
		}
	}
}

// fitPanelBodyLabels enables wrap/truncate for labels in fixed-height surface bodies.
func fitPanelBodyLabels(children []Node, fixedBody bool) {
	for _, ch := range children {
		if ch.IsHidden() {
			continue
		}
		if lbl, ok := ch.(*Label); ok {
			b := lbl.Bounds()
			if b.Width < 8 {
				continue
			}
			if fixedBody {
				lbl.Align = LabelAlignLeft
				lbl.Wrap = true
				if !lbl.IsAutoHeight() {
					lbl.Truncate = true
				}
				lbl.MarkDirty()
				continue
			}
			style := lbl.GetStyle()
			if float32(measureTextS(lbl.Text.Get(), style)) > b.Width-4 {
				lbl.Align = LabelAlignLeft
				lbl.Wrap = true
				lbl.MarkDirty()
			}
			continue
		}
		if rt, ok := ch.(*RichText); ok {
			b := rt.Bounds()
			if b.Width < 8 {
				continue
			}
			if !rt.Wrap && styleUsesFlexPlainText(rt.styleName) {
				rt.Wrap = true
				rt.InvalidateAutoHeightMeasure()
				rt.MarkDirty()
			}
		}
	}
}

// bodyTitleHeightFor returns the header band height for a panel or card facade.
func bodyTitleHeightFor(n Node) float32 {
	switch v := n.(type) {
	case *Panel:
		return v.bodyTitleHeight()
	case *Card:
		return v.bodyTitleHeight()
	case *SurfaceShell:
		return v.bodyTitleHeight()
	case *RaisedSurface:
		return 0
	default:
		return 0
	}
}
