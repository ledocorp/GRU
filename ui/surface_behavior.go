// Package ui — surface behavior plugins (Phase C2–C3).
//
// Behaviors attach to SurfaceShell (Panel/Card) or coordinate with a standalone
// HeaderBand. Headers may have features (collapse chevron), none, or sit alone
// without a panel body — see HeaderBand and docs/SURFACE_COMPOSITION_PLAN.md §11.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// SurfaceBehavior is a composable plugin for SurfaceShell (collapse, dismiss, float…).
type SurfaceBehavior interface {
	AttachShell(sh *SurfaceShell)
	Update(dt float32)
	LayoutAfterBody(sh *SurfaceShell)
	DrawOverlay(sh *SurfaceShell)
	HeaderInteractive() bool
}

// CollapseBehavior tweens shell body visibility (Accordion-style height band).
type CollapseBehavior struct {
	shell        *SurfaceShell
	Expanded     *Signal[bool]
	AnimDuration float32
	animH        float32
	contentH     float32
	tween        *Tween
	hoverHeader  bool
	subscribed   bool
	// ExternalHeader when true: HeaderBand (or other external row) owns header clicks.
	ExternalHeader bool
}

// NewCollapseBehavior creates a collapse plugin. Attach with SurfaceShell.AttachBehavior
// or Panel.EnableCollapse.
func NewCollapseBehavior() *CollapseBehavior {
	return &CollapseBehavior{
		Expanded:     NewSignal(true),
		AnimDuration: 0.11,
	}
}

// AttachShell wires the behavior to a shell.
func (c *CollapseBehavior) AttachShell(sh *SurfaceShell) {
	c.shell = sh
	if c.subscribed {
		return
	}
	c.subscribed = true
	c.Expanded.Subscribe(func() {
		if c.shell == nil {
			return
		}
		c.shell.MarkDirty()
		if c.Expanded.Get() {
			c.startExpand()
		} else {
			c.startCollapse()
		}
	})
}

func (c *CollapseBehavior) Update(dt float32) {
	if c.shell == nil || c.shell.IsHidden() {
		return
	}
	if c.tween != nil && c.tween.IsActive {
		c.tween.Update(dt)
	} else if c.tween != nil && !c.tween.IsActive {
		c.tween = nil
	}
	if c.ExternalHeader || !c.HeaderInteractive() {
		return
	}
	if c.shell != nil && c.shell.panelFeatures != nil {
		return
	}
	mouse := rl.GetMousePosition()
	headerRect := c.headerRect()
	prev := c.hoverHeader
	c.hoverHeader = rl.CheckCollisionPointRec(mouse, headerRect)
	if c.hoverHeader != prev {
		c.shell.MarkDrawDirty()
	}
	if c.hoverHeader && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		c.Toggle()
	}
}

func (c *CollapseBehavior) LayoutAfterBody(sh *SurfaceShell) {
	c.measureNaturalBody(sh)
	if c.tween != nil {
		c.applyVisibleBody(sh)
		return
	}
	if c.Expanded.Get() {
		c.animH = c.contentH
		c.applyVisibleBody(sh)
		return
	}
	if c.animH != 0 {
		c.animH = 0
		c.applyVisibleBody(sh)
	}
}

func (c *CollapseBehavior) DrawOverlay(sh *SurfaceShell) {
	if c.ExternalHeader || sh == nil || sh.Title == "" {
		return
	}
	if sh.panelFeatures != nil {
		return
	}
	progress := c.chevronProgress()
	if progress < 0 {
		return
	}
	hr := c.headerRect()
	style := GetThemeStyle("panel-title")
	col := style.TextColor
	if c.hoverHeader {
		col = brightenColor(col, 20)
	}
	drawSurfaceChevron(hr.X+hr.Width-28, hr.Y+hr.Height/2, 10, progress, col)
}

func (c *CollapseBehavior) HeaderInteractive() bool { return true }

// Toggle flips expanded state.
func (c *CollapseBehavior) Toggle() {
	c.Expanded.Set(!c.Expanded.Get())
}

func (c *CollapseBehavior) headerRect() rl.Rectangle {
	if c.shell == nil {
		return rl.Rectangle{}
	}
	b := c.shell.Bounds()
	titleH := c.shell.bodyTitleHeight()
	if titleH <= 0 {
		titleH = c.shell.TitleHeight
	}
	return rl.NewRectangle(b.X, b.Y, b.Width, titleH)
}

func (c *CollapseBehavior) chevronProgress() float32 {
	if c.contentH <= 0 {
		if c.Expanded.Get() {
			return 1
		}
		return 0
	}
	return c.animH / c.contentH
}

func (c *CollapseBehavior) visibleBodyH(sh *SurfaceShell) float32 {
	if c.tween != nil {
		return c.animH
	}
	if c.Expanded.Get() {
		return c.contentH
	}
	return 0
}

func (c *CollapseBehavior) measureNaturalBody(sh *SurfaceShell) {
	if sh == nil || sh.body == nil {
		return
	}
	titleOff := sh.bodyTitleHeight()
	innerW := sh.bounds.Width - 2*sh.GetStyle().Padding
	if innerW < 1 {
		innerW = sh.bounds.Width
	}
	if innerW < 1 {
		return
	}
	pad := sh.body.GetStyle().Padding
	bodyY := sh.bounds.Y + titleOff
	contentEnd := bodyY + pad
	gap := sh.body.Gap
	for i, ch := range sh.Children() {
		if ch.IsHidden() {
			continue
		}
		if i > 0 && gap > 0 {
			contentEnd += gap
		}
		mb := ch.Bounds()
		mb.Width = innerW
		mb.Height = 0
		mb.X = sh.bounds.X + sh.GetStyle().Padding
		mb.Y = contentEnd
		layoutSetBounds(ch, mb)
		ch.MarkDirty()
		ch.Layout()
		if sub := nodeSubtreeBottom(ch); sub > contentEnd {
			contentEnd = sub
		}
	}
	contentEnd += pad
	c.contentH = contentEnd - bodyY
	if c.contentH < 0 {
		c.contentH = 0
	}
}

func (c *CollapseBehavior) applyVisibleBody(sh *SurfaceShell) {
	if sh == nil || sh.body == nil {
		return
	}
	titleOff := sh.bodyTitleHeight()
	visible := c.visibleBodyH(sh)
	if visible < 0 {
		visible = 0
	}
	if visible > c.contentH && c.contentH > 0 {
		visible = c.contentH
	}
	userSized := false
	if pf := sh.panelFeatures; pf != nil {
		userSized = pf.userSizedShell()
	}
	layoutH := c.contentH
	if userSized && visible > 0.5 {
		layoutH = sh.bounds.Height - titleOff
		if layoutH < 1 {
			layoutH = visible
		}
	}
	if layoutH < 1 {
		layoutH = visible
	}
	bodyRect := rl.NewRectangle(sh.bounds.X, sh.bounds.Y+titleOff, sh.bounds.Width, layoutH)
	layoutSetBounds(sh.body, bodyRect)
	sh.body.ClipChildren = visible < layoutH-0.5
	sh.body.MarkDirty()
	sh.body.Layout()

	want := titleOff + visible
	if userSized && visible > 0.5 {
		if sh.bounds.Height < want {
			sh.bounds.Height = want
		}
	} else if sh.bounds.Height != want {
		sh.bounds.Height = want
	}
	if sh.AutoHeight && sh.GetFlexGrow() == 0 && !userSized {
		sh.bounds.Height = want
	}
	sh.MarkDirty()
	syncLayoutExtent(sh)
}

func (c *CollapseBehavior) startExpand() {
	if c.shell == nil {
		return
	}
	c.measureNaturalBody(c.shell)
	start := c.animH
	end := c.contentH
	if end <= 0 {
		c.animH = 0
		c.applyVisibleBody(c.shell)
		return
	}
	if start >= end {
		c.animH = end
		c.applyVisibleBody(c.shell)
		c.tween = nil
		return
	}
	dur := c.AnimDuration * (1 - start/end)
	if dur < 0.04 {
		dur = 0.04
	}
	sh := c.shell
	c.tween = NewTween(start, end, dur, EaseInOutQuad,
		func(v float32) {
			c.animH = v
			c.applyVisibleBody(sh)
		},
		func() {
			c.animH = c.contentH
			c.applyVisibleBody(sh)
			c.tween = nil
		},
	)
}

func (c *CollapseBehavior) startCollapse() {
	if c.shell == nil {
		return
	}
	start := c.animH
	if start <= 0 {
		c.animH = 0
		c.applyVisibleBody(c.shell)
		c.tween = nil
		return
	}
	if c.contentH <= 0 {
		c.animH = 0
		c.applyVisibleBody(c.shell)
		return
	}
	dur := c.AnimDuration * (start / c.contentH)
	if dur < 0.04 {
		dur = 0.04
	}
	sh := c.shell
	c.tween = NewTween(start, 0, dur, EaseInOutQuad,
		func(v float32) {
			c.animH = v
			c.applyVisibleBody(sh)
		},
		func() {
			c.animH = 0
			c.applyVisibleBody(sh)
			c.tween = nil
		},
	)
}
