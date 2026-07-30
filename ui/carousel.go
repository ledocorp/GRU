// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	carouselDefaultH    = float32(220)
	carouselDotSize     = float32(8)
	carouselDotGap      = float32(8)
	carouselDotBottom   = float32(14)
	carouselArrowW      = float32(40)
	carouselArrowMargin = float32(8)
	carouselPadTop      = float32(13) // breathing room above slide cards
	carouselPadBottom   = float32(10) // below slide when dots hidden
	carouselSlidePadX   = float32(14) // horizontal inset so slide cards do not hug arrow gutters
)

// Carousel is a horizontal slide strip with prev/next controls and dot indicators.
//
// # LLM Prompt Template
//
//	idx := ui.NewSignal(0)
//	c := ui.NewCarousel("hero", []ui.Node{slideA, slideB}, idx, 0, 0, 0, 240)
//	c.AutoPlayInterval = 4 // seconds; 0 disables
//	vp.AddChild(c)
//
// Demo scenes: **Carousel Demo**.
type Carousel struct {
	Element
	Slides           []Node
	Index            *Signal[int]
	AutoPlayInterval float32 // seconds; 0 = manual only
	ShowDots         bool
	autoplayTimer    float32
	hoverPrev        bool
	hoverNext        bool
	lastSlideVPValid bool
	lastSlideVP      rl.Rectangle
	lastSlideIdx     int
}

// NewCarousel creates a carousel. h=0 uses carouselDefaultH.
func NewCarousel(id string, slides []Node, index *Signal[int], x, y, w, h float32) *Carousel {
	if index == nil {
		index = NewSignal(0)
	}
	if len(slides) == 0 {
		slides = []Node{NewLabel(id+"-empty", "No slides", 0, 0, 0, 40)}
	}
	if index.Get() < 0 || index.Get() >= len(slides) {
		index.Set(0)
	}
	c := &Carousel{
		Element:  NewElement(id, x, y, w, h),
		Slides:   slides,
		Index:    index,
		ShowDots: true,
	}
	c.styleName = "carousel"
	if h == 0 {
		c.bounds.Height = carouselDefaultH
	}
	for _, s := range slides {
		if s != nil {
			s.SetParent(c)
		}
	}
	c.Index.Subscribe(func() {
		c.MarkDirty()
		c.MarkDrawDirty()
	})
	return c
}

// MarkDirty invalidates slide layout cache when geometry or index changes.
func (c *Carousel) MarkDirty() {
	c.Element.MarkDirty()
	c.lastSlideVPValid = false
}

// Children implements Node for inspector tree walks.
func (c *Carousel) Children() []Node { return c.Slides }

// AddChild appends a slide.
func (c *Carousel) AddChild(child Node) {
	if child == nil {
		return
	}
	c.Slides = append(c.Slides, child)
	child.SetParent(c)
	c.MarkDirty()
}

func (c *Carousel) slideCount() int { return len(c.Slides) }

func (c *Carousel) prevRect() rl.Rectangle {
	b := c.Bounds()
	return rl.NewRectangle(b.X+carouselArrowMargin, b.Y+(b.Height-carouselArrowW)/2, carouselArrowW, carouselArrowW)
}

func (c *Carousel) nextRect() rl.Rectangle {
	b := c.Bounds()
	return rl.NewRectangle(b.X+b.Width-carouselArrowMargin-carouselArrowW, b.Y+(b.Height-carouselArrowW)/2, carouselArrowW, carouselArrowW)
}

func (c *Carousel) slideViewport() rl.Rectangle {
	b := c.Bounds()
	left := b.X + carouselArrowW + carouselArrowMargin*2
	width := b.Width - 2*(carouselArrowW+carouselArrowMargin*2)
	bottom := carouselPadBottom
	if c.ShowDots {
		bottom = carouselDotBottom + carouselDotSize + 8
	}
	top := b.Y + carouselPadTop
	height := b.Height - carouselPadTop - bottom
	if height < 1 {
		height = 1
	}
	return rl.NewRectangle(left, top, width, height)
}

func carouselSlideBand(viewport rl.Rectangle) rl.Rectangle {
	w := viewport.Width - 2*carouselSlidePadX
	if w < 1 {
		w = viewport.Width
	}
	x := viewport.X + carouselSlidePadX
	if w == viewport.Width {
		x = viewport.X
	}
	return rl.NewRectangle(x, viewport.Y, w, viewport.Height)
}

func carouselShowSlide(s Node) {
	if s == nil || !s.IsHidden() {
		return
	}
	s.Show()
}

func carouselHideSlide(s Node) {
	if s == nil || s.IsHidden() {
		return
	}
	s.Hide()
}

func carouselRectsEqual(a, b rl.Rectangle) bool {
	return absF(a.X-b.X) < 0.5 && absF(a.Y-b.Y) < 0.5 &&
		absF(a.Width-b.Width) < 0.5 && absF(a.Height-b.Height) < 0.5
}

// layoutActiveSlide sizes the visible card at the slide band top — intrinsic height,
// capped when content is taller than the band (same for every carousel height).
func (c *Carousel) layoutActiveSlide(s Node, viewport rl.Rectangle) {
	if s == nil {
		return
	}
	band := carouselSlideBand(viewport)
	layoutSetBounds(s, rl.NewRectangle(band.X, band.Y, band.Width, 0))
	if el := nodeElementPtrForResize(s); el != nil {
		el.AutoHeight = true
	}
	s.Layout()
	h := s.Bounds().Height
	if h < 1 {
		h = band.Height
	}
	y := band.Y
	if h > band.Height {
		h = band.Height
	}
	if el := nodeElementPtrForResize(s); el != nil {
		el.AutoHeight = false
	}
	layoutSetBounds(s, rl.NewRectangle(band.X, y, band.Width, h))
	s.Layout()
}

func (c *Carousel) dotRect(i, count int, b rl.Rectangle) rl.Rectangle {
	totalW := float32(count)*carouselDotSize + float32(count-1)*carouselDotGap
	startX := b.X + (b.Width-totalW)/2
	x := startX + float32(i)*(carouselDotSize+carouselDotGap)
	y := b.Y + b.Height - carouselDotBottom - carouselDotSize
	return rl.NewRectangle(x, y, carouselDotSize, carouselDotSize)
}

func (c *Carousel) setIndex(i int) {
	n := c.slideCount()
	if n <= 0 {
		return
	}
	for i < 0 {
		i += n
	}
	i %= n
	if c.Index.Get() != i {
		c.Index.Set(i)
	}
	c.autoplayTimer = 0
}

// IsInteractive implements Node.
func (c *Carousel) IsInteractive() bool { return c.slideCount() > 1 }

// Layout sizes the active slide to the inner viewport.
func (c *Carousel) Layout() {
	defer func() { c.layoutDirty = false }()
	if c.IsHidden() || c.slideCount() == 0 {
		return
	}
	r := c.slideViewport()
	idx := c.Index.Get()
	if c.lastSlideVPValid && c.lastSlideIdx == idx && carouselRectsEqual(c.lastSlideVP, r) {
		return
	}
	for i, s := range c.Slides {
		if s == nil {
			continue
		}
		if i == idx {
			carouselShowSlide(s)
			c.layoutActiveSlide(s, r)
		} else {
			layoutSetBounds(s, rl.NewRectangle(r.X, r.Y, 0, 0))
			// Inactive slides must not stay layoutDirty — SubtreeNeedsRedraw only
			// skips hidden nodes, not zero-bounds cards (carousel-m2 idle guard).
			s.Layout()
			carouselHideSlide(s)
		}
	}
	c.lastSlideVP = r
	c.lastSlideIdx = idx
	c.lastSlideVPValid = true
}

// Update handles navigation, autoplay, and active slide input.
func (c *Carousel) Update(dt float32) {
	if c.IsHidden() || c.slideCount() == 0 {
		return
	}
	n := c.slideCount()
	mouse := rl.GetMousePosition()
	prev := c.prevRect()
	next := c.nextRect()
	wasPrev := c.hoverPrev
	wasNext := c.hoverNext
	c.hoverPrev = n > 1 && rl.CheckCollisionPointRec(mouse, prev)
	c.hoverNext = n > 1 && rl.CheckCollisionPointRec(mouse, next)
	if c.hoverPrev != wasPrev || c.hoverNext != wasNext {
		c.MarkDrawDirty()
	}

	if n > 1 {
		if c.hoverPrev && PointerClickConsume(prev) {
			c.setIndex(c.Index.Get() - 1)
		}
		if c.hoverNext && PointerClickConsume(next) {
			c.setIndex(c.Index.Get() + 1)
		}
		if c.ShowDots && PointerClickPending() {
			b := c.Bounds()
			for i := 0; i < n; i++ {
				dr := c.dotRect(i, n, b)
				if PointerClickConsume(dr) {
					c.setIndex(i)
					break
				}
			}
		}
	}

	if c.AutoPlayInterval > 0 && n > 1 {
		c.autoplayTimer += dt
		if c.autoplayTimer >= c.AutoPlayInterval {
			c.setIndex(c.Index.Get() + 1)
		}
	}

	idx := c.Index.Get()
	if idx >= 0 && idx < n && c.Slides[idx] != nil {
		c.Slides[idx].Update(dt)
	}
}

// Draw implements Node.Draw.
func (c *Carousel) Draw() {
	defer func() { c.drawDirty = false }()
	c.drawInternal()
}

func (c *Carousel) drawInternal() {
	if c.IsHidden() {
		return
	}
	b := c.Bounds()
	style := c.GetStyle()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRounded(b, 0.06, 8, style.BackgroundColor)
	}
	if style.BorderWidth > 0 && style.BorderColor.A > 0 {
		rl.DrawRectangleRoundedLinesEx(b, 0.06, 8, style.BorderWidth, style.BorderColor)
	}

	idx := c.Index.Get()
	if idx >= 0 && idx < len(c.Slides) && c.Slides[idx] != nil {
		c.Slides[idx].Draw()
	}

	n := c.slideCount()
	if n <= 1 {
		return
	}

	drawArrow := func(r rl.Rectangle, iconName, fallback string, hover bool) {
		bg := rl.NewColor(255, 255, 255, 200)
		if hover {
			bg = rl.NewColor(232, 234, 255, 230)
		}
		cx := r.X + r.Width/2
		cy := r.Y + r.Height/2
		rl.DrawCircleV(rl.NewVector2(cx, cy), r.Width/2, bg)
		col := style.TextColor
		if col.A == 0 {
			col = rl.NewColor(55, 48, 163, 255)
		}
		iconSize := phosphorIconDrawSize(r.Width, 0)
		dst := snapPhosphorRect(rl.NewRectangle(cx-iconSize/2, cy-iconSize/2, iconSize, iconSize))
		Phosphor.EnsureLoaded(iconName, PhosphorRegular)
		if !Phosphor.Draw(dst, iconName, PhosphorRegular, col) {
			tst := style
			tst.FontSize = 20
			tst.Bold = true
			iw := measureTextS(fallback, tst)
			drawTextS(fallback, int32(dst.X+(dst.Width-float32(iw))/2), TextPosY(dst, tst), tst)
		}
	}
	drawArrow(c.prevRect(), PhosphorCaretLeft, "<", c.hoverPrev)
	drawArrow(c.nextRect(), PhosphorCaretRight, ">", c.hoverNext)

	if c.ShowDots {
		for i := 0; i < n; i++ {
			dr := c.dotRect(i, n, b)
			col := rl.NewColor(180, 184, 200, 255)
			if i == idx {
				col = rl.NewColor(79, 70, 229, 255)
			}
			rl.DrawCircleV(rl.NewVector2(dr.X+dr.Width/2, dr.Y+dr.Height/2), dr.Width/2, col)
		}
	}
}
