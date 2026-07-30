// Package ui (continued) — PanelFeatures behavior implementation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	panelResizeGrip  = float32(14)
	panelSideGrip    = float32(6)
	panelMinW        = float32(120)
	panelFloatMinW   = float32(260)
	panelMinH        = float32(80)
	panelAnchorInset = float32(8)
)

type panelResizeEdge int

const (
	panelResizeNone panelResizeEdge = iota
	panelResizeN
	panelResizeS
	panelResizeW
	panelResizeE
	panelResizeNW
	panelResizeNE
	panelResizeSW
	panelResizeSE
)

// PanelFeaturesBehavior implements egui-style panel chrome and interaction.
type PanelFeaturesBehavior struct {
	shell    *SurfaceShell
	config   *PanelFeatures
	collapse *CollapseBehavior
	dismiss  *DismissBehavior
	escape   *EscapeBehavior

	scrollOuter *Viewport
	scrollInner *Viewport

	hoverCollapse bool
	hoverResize   panelResizeEdge

	dragging       bool
	userPositioned bool
	dragStartMX   float32
	dragStartMY   float32
	dragStartX    float32
	dragStartY    float32

	resizing      bool
	resizeEdge    panelResizeEdge
	resizeStartMX float32
	resizeStartMY float32
	resizeStartX  float32
	resizeStartY  float32
	resizeStartW  float32
	resizeStartH  float32
}

// NewPanelFeaturesBehavior creates the default panel feature controller.
func NewPanelFeaturesBehavior() *PanelFeaturesBehavior {
	cfg := DefaultPanelFeatures()
	return &PanelFeaturesBehavior{config: &cfg}
}

// AttachShell wires the controller to a SurfaceShell.
func (pf *PanelFeaturesBehavior) AttachShell(sh *SurfaceShell) {
	pf.shell = sh
	sh.panelFeatures = pf
	pf.Apply()
}

func (pf *PanelFeaturesBehavior) Apply() {
	if pf.shell == nil || pf.config == nil {
		return
	}
	if pf.config.TitleBar {
		switch pf.shell.styleComponent {
		case "card":
			if pf.shell.headerMode == HeaderModeNone {
				pf.shell.headerMode = HeaderModeInset
			}
		default:
			if pf.shell.headerMode == HeaderModeNone {
				pf.shell.headerMode = HeaderModeTitleBar
			}
		}
	} else {
		pf.shell.headerMode = HeaderModeNone
	}
	if pf.config.Collapsible {
		if pf.collapse == nil {
			pf.collapse = NewCollapseBehavior()
			pf.collapse.ExternalHeader = false
			pf.shell.AttachBehavior(pf.collapse)
		}
	} else if pf.collapse != nil {
		pf.collapse.Expanded.Set(true)
	}
	pf.syncDismissEscape()
	pf.syncScrollHost()
	if pf.config.Movable || pf.config.Resizable {
		pf.shell.SetCachePolicy(CacheNever)
	}
	pf.shell.MarkDirty()
}

func (pf *PanelFeaturesBehavior) syncDismissEscape() {
	if pf.shell == nil || pf.config == nil {
		return
	}
	if pf.config.Closable {
		if pf.dismiss == nil {
			pf.dismiss = NewDismissBehavior()
			pf.shell.AttachBehavior(pf.dismiss)
		}
		pf.dismiss.SetActive(true)
		pf.dismiss.OnDismiss = pf.config.OnDismiss
	} else if pf.dismiss != nil {
		pf.dismiss.SetActive(false)
	}
	if pf.config.CloseOnEscape {
		if pf.dismiss == nil {
			pf.config.Closable = true
			pf.syncDismissEscape()
		}
		if pf.escape == nil && pf.dismiss != nil {
			pf.escape = NewEscapeBehavior(pf.dismiss)
			pf.shell.AttachBehavior(pf.escape)
		}
		if pf.escape != nil {
			pf.escape.Enabled = true
			pf.escape.dismiss = pf.dismiss
		}
	} else if pf.escape != nil {
		pf.escape.Enabled = false
	}
}

func (pf *PanelFeaturesBehavior) HeaderInteractive() bool { return pf.IsInteractive() }

func (pf *PanelFeaturesBehavior) IsInteractive() bool {
	if pf.config == nil {
		return false
	}
	return pf.config.Collapsible || pf.config.Closable || pf.config.Resizable ||
		(pf.config.Movable && pf.config.DragMode != PanelDragOff)
}

// ChromeHeight returns extra header band height when title bar is off but controls remain.
func (pf *PanelFeaturesBehavior) ChromeHeight() float32 {
	if pf.config == nil || pf.shell == nil {
		return 0
	}
	if pf.config.TitleBar {
		return 0
	}
	if pf.config.Collapsible || pf.config.Closable || pf.needsDragChrome() {
		return surfaceChromeRowH
	}
	return 0
}

func (pf *PanelFeaturesBehavior) needsDragChrome() bool {
	return pf.config.Movable && (pf.config.DragMode == PanelDragOnTouch || pf.config.DragMode == PanelDragTitleBar)
}

func (pf *PanelFeaturesBehavior) headerBandRect(sh *SurfaceShell) rl.Rectangle {
	return surfaceHeaderBandRect(sh)
}

func (pf *PanelFeaturesBehavior) collapseBtnRect(sh *SurfaceShell) rl.Rectangle {
	return surfaceCollapseBtnRect(sh)
}

func (pf *PanelFeaturesBehavior) closeBtnRect(sh *SurfaceShell) rl.Rectangle {
	return surfaceCloseBtnRect(sh)
}

func (pf *PanelFeaturesBehavior) userSizedShell() bool {
	return pf.userPositioned && pf.config != nil && pf.config.Resizable
}

func (pf *PanelFeaturesBehavior) minHeight() float32 {
	if pf.config != nil && pf.config.Resizable {
		return panelMinH
	}
	return 0
}

func (pf *PanelFeaturesBehavior) resizeEdgeAt(mouse rl.Vector2, b rl.Rectangle) panelResizeEdge {
	if b.Width < 1 || b.Height < 1 {
		return panelResizeNone
	}
	x, y := mouse.X, mouse.Y
	g := panelResizeGrip
	sg := panelSideGrip

	onLeft := x >= b.X && x <= b.X+sg
	onRight := x >= b.X+b.Width-sg && x <= b.X+b.Width
	onTop := y >= b.Y && y <= b.Y+sg
	onBottom := y >= b.Y+b.Height-g && y <= b.Y+b.Height

	if onBottom && x >= b.X+b.Width-g {
		return panelResizeSE
	}
	if onBottom && x <= b.X+g {
		return panelResizeSW
	}
	if onTop && x >= b.X+b.Width-g {
		return panelResizeNE
	}
	if onTop && x <= b.X+g {
		return panelResizeNW
	}
	if onLeft {
		return panelResizeW
	}
	if onRight {
		return panelResizeE
	}
	if onTop {
		return panelResizeN
	}
	if onBottom {
		return panelResizeS
	}
	return panelResizeNone
}

func (pf *PanelFeaturesBehavior) minWidth() float32 {
	if pf.config != nil && pf.config.MinWidth > 0 {
		return pf.config.MinWidth
	}
	if pf.config != nil && pf.config.Movable && pf.config.Resizable {
		return panelFloatMinW
	}
	return panelMinW
}

func (pf *PanelFeaturesBehavior) Update(dt float32) {
	if pf.shell == nil || pf.shell.IsHidden() || pf.config == nil {
		return
	}
	mouse := rl.GetMousePosition()

	prevC, prevR := pf.hoverCollapse, pf.hoverResize
	pf.hoverCollapse = false
	pf.hoverResize = panelResizeNone

	if pf.config.Collapsible {
		pf.hoverCollapse = rl.CheckCollisionPointRec(mouse, pf.collapseBtnRect(pf.shell))
	}
	if pf.config.Resizable && pf.collapseExpandedEnough() {
		pf.hoverResize = pf.resizeEdgeAt(mouse, pf.shell.Bounds())
	}

	if prevC != pf.hoverCollapse || prevR != pf.hoverResize {
		pf.shell.MarkDrawDirty()
	}

	if pf.resizing && pf.config.Resizable {
		pf.setResizeCursor(pf.resizeEdge)
		pf.applyResize(mouse)
	} else if pf.hoverResize != panelResizeNone {
		pf.setResizeCursor(pf.hoverResize)
	} else if prevR != panelResizeNone {
		// Left resize band — release so other widgets can set I-beam / default.
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}

	if !pf.resizing && pf.dragging && pf.config.Movable {
		pf.applyDrag(mouse)
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if pf.config.Collapsible && pf.hoverCollapse && pf.collapse != nil {
			pf.collapse.Toggle()
			return
		}
		if pf.config.Resizable && pf.hoverResize != panelResizeNone {
			pf.startResize(mouse, pf.hoverResize)
			return
		}
		if pf.config.Movable && pf.config.DragMode != PanelDragOff {
			zone := pf.dragZoneRect()
			if zone.Width > 0 && rl.CheckCollisionPointRec(mouse, zone) {
				pf.startDrag(mouse)
			}
		}
	}

	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		pf.dragging = false
		pf.resizing = false
		pf.resizeEdge = panelResizeNone
	}
}

func (pf *PanelFeaturesBehavior) setResizeCursor(e panelResizeEdge) {
	switch e {
	case panelResizeW, panelResizeE:
		rl.SetMouseCursor(rl.MouseCursorResizeEW)
	case panelResizeN, panelResizeS:
		rl.SetMouseCursor(rl.MouseCursorResizeNS)
	case panelResizeNW, panelResizeSE:
		rl.SetMouseCursor(rl.MouseCursorResizeNWSE)
	case panelResizeNE, panelResizeSW:
		rl.SetMouseCursor(rl.MouseCursorResizeNESW)
	default:
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}
}

func (pf *PanelFeaturesBehavior) collapseExpandedEnough() bool {
	if pf.collapse == nil {
		return true
	}
	return pf.collapse.visibleBodyH(pf.shell) > 0.5
}

func (pf *PanelFeaturesBehavior) dragZoneRect() rl.Rectangle {
	sh := pf.shell
	b := sh.Bounds()
	switch pf.config.DragMode {
	case PanelDragTitleBar:
		return pf.headerBandRect(sh)
	case PanelDragOnTouch:
		if sh.bodyTitleHeight() > 0 {
			return pf.headerBandRect(sh)
		}
		return rl.Rectangle{}
	case PanelDragAnywhere:
		return b
	default:
		return rl.Rectangle{}
	}
}

func (pf *PanelFeaturesBehavior) startDrag(mouse rl.Vector2) {
	zone := pf.dragZoneRect()
	if zone.Width <= 0 || !rl.CheckCollisionPointRec(mouse, zone) {
		return
	}
	if pf.config.Collapsible && rl.CheckCollisionPointRec(mouse, pf.collapseBtnRect(pf.shell)) {
		return
	}
	if pf.dismiss != nil && pf.dismiss.Active() && rl.CheckCollisionPointRec(mouse, surfaceCloseBtnRect(pf.shell)) {
		return
	}
	pf.dragging = true
	pf.userPositioned = true
	pf.dragStartMX = mouse.X
	pf.dragStartMY = mouse.Y
	b := pf.shell.Bounds()
	pf.dragStartX = b.X
	pf.dragStartY = b.Y
}

func (pf *PanelFeaturesBehavior) applyDrag(mouse rl.Vector2) {
	dx := mouse.X - pf.dragStartMX
	dy := mouse.Y - pf.dragStartMY
	b := pf.shell.Bounds()
	b.X = pf.dragStartX + dx
	b.Y = pf.dragStartY + dy
	if pf.config.Constrain {
		b = pf.clampPosition(b)
	}
	pf.shell.setBoundsNoMark(b)
	pf.shell.MarkDirty()
}

func (pf *PanelFeaturesBehavior) startResize(mouse rl.Vector2, edge panelResizeEdge) {
	pf.resizing = true
	pf.resizeEdge = edge
	pf.resizeStartMX = mouse.X
	pf.resizeStartMY = mouse.Y
	b := pf.shell.Bounds()
	pf.resizeStartX = b.X
	pf.resizeStartY = b.Y
	pf.resizeStartW = b.Width
	pf.resizeStartH = b.Height
	pf.shell.AutoHeight = false
	pf.userPositioned = true
}

func (pf *PanelFeaturesBehavior) applyResize(mouse rl.Vector2) {
	dx := mouse.X - pf.resizeStartMX
	dy := mouse.Y - pf.resizeStartMY
	b := pf.shell.Bounds()
	minW := pf.minWidth()
	minH := pf.minHeight()

	switch pf.resizeEdge {
	case panelResizeW:
		newW := pf.resizeStartW - dx
		if newW < minW {
			dx = pf.resizeStartW - minW
			newW = minW
		}
		b.X = pf.resizeStartX + dx
		b.Width = newW
	case panelResizeE:
		b.Width = pf.resizeStartW + dx
		if b.Width < minW {
			b.Width = minW
		}
	case panelResizeN:
		newH := pf.resizeStartH - dy
		if newH < minH {
			dy = pf.resizeStartH - minH
			newH = minH
		}
		b.Y = pf.resizeStartY + dy
		b.Height = newH
	case panelResizeS:
		b.Height = pf.resizeStartH + dy
		if b.Height < minH {
			b.Height = minH
		}
	case panelResizeNW:
		newW := pf.resizeStartW - dx
		if newW < minW {
			dx = pf.resizeStartW - minW
			newW = minW
		}
		newH := pf.resizeStartH - dy
		if newH < minH {
			dy = pf.resizeStartH - minH
			newH = minH
		}
		b.X = pf.resizeStartX + dx
		b.Y = pf.resizeStartY + dy
		b.Width = newW
		b.Height = newH
	case panelResizeNE:
		b.Width = pf.resizeStartW + dx
		if b.Width < minW {
			b.Width = minW
		}
		newH := pf.resizeStartH - dy
		if newH < minH {
			dy = pf.resizeStartH - minH
			newH = minH
		}
		b.Y = pf.resizeStartY + dy
		b.Height = newH
	case panelResizeSW:
		newW := pf.resizeStartW - dx
		if newW < minW {
			dx = pf.resizeStartW - minW
			newW = minW
		}
		b.X = pf.resizeStartX + dx
		b.Width = newW
		b.Height = pf.resizeStartH + dy
		if b.Height < minH {
			b.Height = minH
		}
	case panelResizeSE:
		b.Width = pf.resizeStartW + dx
		b.Height = pf.resizeStartH + dy
		if b.Width < minW {
			b.Width = minW
		}
		if b.Height < minH {
			b.Height = minH
		}
	default:
		return
	}

	if pf.config.Constrain {
		b = pf.clampSize(b)
	}
	pf.shell.AutoHeight = false
	pf.shell.setBoundsNoMark(b)
	pf.shell.MarkDirty()
}

func (pf *PanelFeaturesBehavior) parentBounds() rl.Rectangle {
	if pf.shell == nil {
		return rl.Rectangle{}
	}
	p := pf.shell.ParentNode()
	if p == nil {
		return rl.NewRectangle(0, 0, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))
	}
	return p.Bounds()
}

func (pf *PanelFeaturesBehavior) clampPosition(b rl.Rectangle) rl.Rectangle {
	pb := pf.parentBounds()
	if b.X < pb.X {
		b.X = pb.X
	}
	if b.Y < pb.Y {
		b.Y = pb.Y
	}
	if b.X+b.Width > pb.X+pb.Width {
		b.X = pb.X + pb.Width - b.Width
	}
	if b.Y+b.Height > pb.Y+pb.Height {
		b.Y = pb.Y + pb.Height - b.Height
	}
	return b
}

func (pf *PanelFeaturesBehavior) clampSize(b rl.Rectangle) rl.Rectangle {
	pb := pf.parentBounds()
	if b.Width > pb.Width {
		b.Width = pb.Width
	}
	if b.Height > pb.Height {
		b.Height = pb.Height
	}
	if b.X+b.Width > pb.X+pb.Width {
		b.X = pb.X + pb.Width - b.Width
	}
	if b.Y+b.Height > pb.Y+pb.Height {
		b.Y = pb.Y + pb.Height - b.Height
	}
	b = pf.clampPosition(b)
	return b
}

func (pf *PanelFeaturesBehavior) LayoutAfterBody(sh *SurfaceShell) {
	pf.syncScrollHost()
	if pf.config != nil && pf.config.FloatPosition && !pf.userPositioned {
		pf.applyFloatPosition(sh)
	} else if pf.config != nil && pf.config.Anchored && !pf.userPositioned {
		pf.applyAnchor(sh)
	}
}

func (pf *PanelFeaturesBehavior) applyFloatPosition(sh *SurfaceShell) {
	if pf.config == nil {
		return
	}
	b := sh.Bounds()
	if parent := sh.ParentNode(); parent != nil {
		pb := parent.Bounds()
		b.X = pb.X + pf.config.FloatX
		b.Y = pb.Y + pf.config.FloatY
	} else {
		b.X = pf.config.FloatX
		b.Y = pf.config.FloatY
	}
	sh.setBoundsNoMark(b)
	sh.MarkDirty()
}

// layoutScrollViewport sizes the scroll host to the body inner band after collapse settles.
func (pf *PanelFeaturesBehavior) layoutScrollViewport(sh *SurfaceShell) {
	if pf.scrollOuter == nil || sh.body == nil {
		return
	}
	b := sh.body.Bounds()
	pad := sh.GetStyle().Padding
	inner := rl.NewRectangle(b.X+pad, b.Y+pad, b.Width-2*pad, b.Height-2*pad)
	if inner.Width < 1 {
		inner.Width = 1
	}
	if inner.Height < 1 {
		inner.Height = 1
	}
	cur := pf.scrollOuter.Bounds()
	if cur.X != inner.X || cur.Y != inner.Y || cur.Width != inner.Width || cur.Height != inner.Height {
		pf.scrollOuter.setBoundsNoMark(inner)
		pf.scrollOuter.MarkDirty()
	}
	pf.scrollOuter.ClipChildren = true
	pf.scrollOuter.Layout()
}

func (pf *PanelFeaturesBehavior) applyAnchor(sh *SurfaceShell) {
	parent := sh.ParentNode()
	if parent == nil {
		return
	}
	pb := parent.Bounds()
	b := sh.Bounds()
	w, h := b.Width, b.Height

	switch pf.config.AnchorX {
	case PanelAnchorHCenter:
		b.X = pb.X + (pb.Width-w)/2
	case PanelAnchorRight:
		b.X = pb.X + pb.Width - w - panelAnchorInset
	default:
		b.X = pb.X + panelAnchorInset
	}
	switch pf.config.AnchorY {
	case PanelAnchorVCenter:
		b.Y = pb.Y + (pb.Height-h)/2
	case PanelAnchorBottom:
		b.Y = pb.Y + pb.Height - h - panelAnchorInset
	default:
		b.Y = pb.Y + panelAnchorInset
	}
	if b.X != sh.bounds.X || b.Y != sh.bounds.Y {
		sh.setBoundsNoMark(b)
		sh.MarkDirty()
	}
}

func (pf *PanelFeaturesBehavior) syncScrollHost() {
	if pf.shell == nil || pf.shell.body == nil || pf.config == nil {
		return
	}
	want := pf.config.HScroll || pf.config.VScroll
	if !want {
		if pf.scrollOuter != nil {
			pf.teardownScrollHost()
		}
		return
	}
	if pf.scrollOuter != nil {
		return
	}
	children := append([]Node{}, pf.shell.Children()...)
	for _, ch := range children {
		pf.shell.RemoveChild(ch.ID())
	}
	id := pf.shell.ID()
	if pf.config.VScroll && pf.config.HScroll {
		pf.scrollOuter = NewViewport(id+"-scroll-v", 0, 0, 0, 0)
		pf.scrollOuter.ContentClipBleed = 0
		pf.scrollInner = NewHorizontalViewport(id+"-scroll-h", 0, 0, 0, 0)
		pf.scrollInner.ContentClipBleed = 0
		pf.scrollOuter.AddChild(pf.scrollInner)
		for _, ch := range children {
			pf.scrollInner.AddChild(ch)
		}
	} else if pf.config.VScroll {
		pf.scrollOuter = NewViewport(id+"-scroll-v", 0, 0, 0, 0)
		pf.scrollOuter.ContentClipBleed = 0
		for _, ch := range children {
			pf.scrollOuter.AddChild(ch)
		}
	} else {
		pf.scrollOuter = NewHorizontalViewport(id+"-scroll-h", 0, 0, 0, 0)
		pf.scrollOuter.ContentClipBleed = 0
		for _, ch := range children {
			pf.scrollOuter.AddChild(ch)
		}
	}
	pf.scrollOuter.SetFlexGrow(1)
	pf.scrollOuter.ClipChildren = true
	pf.shell.body.ClipChildren = true
	pf.shell.body.AddChild(pf.scrollOuter)
}

func (pf *PanelFeaturesBehavior) teardownScrollHost() {
	if pf.shell == nil || pf.scrollOuter == nil {
		return
	}
	var children []Node
	if pf.scrollInner != nil {
		children = pf.scrollInner.Children()
		for _, ch := range children {
			pf.scrollInner.RemoveChild(ch.ID())
		}
	} else {
		children = pf.scrollOuter.Children()
		for _, ch := range children {
			pf.scrollOuter.RemoveChild(ch.ID())
		}
	}
	pf.shell.body.RemoveChild(pf.scrollOuter.ID())
	for _, ch := range children {
		pf.shell.body.AddChild(ch)
		ch.SetParent(pf.shell.body)
	}
	pf.scrollOuter = nil
	pf.scrollInner = nil
}

func (pf *PanelFeaturesBehavior) DrawOverlay(sh *SurfaceShell) {
	if pf.config == nil {
		return
	}
	if !pf.config.TitleBar && pf.ChromeHeight() > 0 {
		hr := pf.headerBandRect(sh)
		style := sh.GetStyle()
		rl.DrawRectangleRec(hr, style.BackgroundColor)
		sepY := int32(hr.Y + hr.Height)
		rl.DrawLine(int32(hr.X), sepY, int32(hr.X+hr.Width), sepY, rl.NewColor(0, 0, 0, 28))
	}
	style := GetThemeStyle("panel-title")
	col := style.TextColor
	if pf.config.Collapsible && pf.collapse != nil {
		btn := pf.collapseBtnRect(sh)
		rl.DrawRectangleRec(btn, surfaceTitleChromeBackground())
		iconCol := col
		if pf.hoverCollapse {
			iconCol = brightenColor(col, 30)
		}
		icon := PhosphorCaretRight
		if pf.collapse.Expanded.Get() {
			icon = PhosphorCaretDown
		}
		drawSurfacePhosphorIcon(btn, icon, iconCol)
	}
	if pf.config.Closable && pf.dismiss != nil && pf.dismiss.Active() {
		btn := pf.closeBtnRect(sh)
		rl.DrawRectangleRec(btn, surfaceTitleChromeBackground())
		iconCol := col
		if pf.dismiss.hoverClose {
			iconCol = brightenColor(col, 30)
		}
		drawSurfacePhosphorIcon(btn, PhosphorX, iconCol)
	}
	if pf.config.Resizable && pf.collapseExpandedEnough() {
		pf.drawResizeGrips(sh)
	}
}

func (pf *PanelFeaturesBehavior) drawResizeGrips(sh *SurfaceShell) {
	b := sh.Bounds()
	alpha := uint8(36)
	if pf.hoverResize != panelResizeNone {
		alpha = 72
	}
	col := rl.NewColor(60, 60, 80, alpha)
	g := panelResizeGrip
	x2 := b.X + b.Width
	y2 := b.Y + b.Height
	rl.DrawTriangle(
		rl.NewVector2(x2, y2),
		rl.NewVector2(x2-g, y2),
		rl.NewVector2(x2, y2-g),
		col,
	)
	rl.DrawTriangle(
		rl.NewVector2(b.X, y2),
		rl.NewVector2(b.X+g, y2),
		rl.NewVector2(b.X, y2-g),
		col,
	)
	midY := b.Y + b.Height*0.45
	rl.DrawLine(int32(b.X), int32(midY), int32(b.X), int32(y2-2), col)
	rl.DrawLine(int32(x2-1), int32(midY), int32(x2-1), int32(y2-2), col)
	midX := b.X + b.Width*0.5
	rl.DrawLine(int32(midX), int32(y2-1), int32(x2-2), int32(y2-1), col)
}

// panelContentTarget returns where AddChild should attach user content.
func (pf *PanelFeaturesBehavior) panelContentTarget(shell *SurfaceShell) Node {
	if pf.scrollInner != nil {
		return pf.scrollInner
	}
	if pf.scrollOuter != nil {
		return pf.scrollOuter
	}
	if shell.body != nil {
		return shell.body
	}
	return shell
}
