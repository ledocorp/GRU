// Package ui (continued) — demo backdrop for transparent preset previews.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// DemoBackdrop is a flex container that paints a colorful gradient behind its
// children so semi-transparent panels/cards read clearly in demos.
type DemoBackdrop struct {
	Container
}

// NewDemoBackdrop creates a backdrop container. Pass h=0 for intrinsic height.
func NewDemoBackdrop(id string, x, y, w, h float32) *DemoBackdrop {
	c := NewContainer(id, x, y, w, h)
	c.SetStyle("transparent")
	c.Gap = 12
	c.AutoHeight = h == 0
	return &DemoBackdrop{Container: *c}
}

// Layout delegates to the flex container.
func (d *DemoBackdrop) Layout() {
	if d.IsHidden() {
		return
	}
	d.Container.Layout()
}

// fillRect matches the gradient to laid-out children at draw time so page scroll
// cannot desync the fill band from tile positions.
func (d *DemoBackdrop) fillRect() rl.Rectangle {
	b := d.Bounds()
	pad := d.GetStyle().Padding
	host := rl.NewRectangle(b.X+pad, b.Y+pad, b.Width-2*pad, b.Height-2*pad)
	if host.Width < 1 {
		host.Width = 1
	}
	if host.Height < 1 {
		host.Height = 1
	}
	if len(d.Children()) == 0 {
		return host
	}
	minX := host.X + host.Width
	minY := host.Y + host.Height
	maxX := host.X
	maxY := host.Y
	for _, ch := range d.Children() {
		if ch.IsHidden() {
			continue
		}
		cb := ch.Bounds()
		if cb.X < minX {
			minX = cb.X
		}
		if cb.Y < minY {
			minY = cb.Y
		}
		if r := cb.X + cb.Width; r > maxX {
			maxX = r
		}
		if bottom := nodeSubtreeBottom(ch); bottom > maxY {
			maxY = bottom
		}
	}
	if minX > host.X {
		minX = host.X
	}
	if minY > host.Y {
		minY = host.Y
	}
	if maxX < host.X+host.Width {
		maxX = host.X + host.Width
	}
	maxY += pad
	h := maxY - minY
	if h < host.Height {
		h = host.Height
	}
	w := maxX - minX
	if w < host.Width {
		w = host.Width
	}
	return rl.NewRectangle(minX, minY, w, h)
}

// Draw paints the gradient fill then children.
func (d *DemoBackdrop) Draw() {
	if d.IsHidden() {
		return
	}
	drawDemoBackdropFill(d.fillRect())
	d.Container.drawInternal()
}

func drawDemoBackdropFill(bounds rl.Rectangle) {
	if bounds.Width < 1 || bounds.Height < 1 {
		return
	}
	DrawCornerGradientRect(
		bounds,
		rl.NewColor(79, 70, 229, 255),
		rl.NewColor(168, 85, 247, 255),
		rl.NewColor(99, 102, 241, 255),
		rl.NewColor(236, 72, 153, 255),
	)
	pad := float32(10)
	step := float32(28)
	for y := bounds.Y + pad; y < bounds.Y+bounds.Height-pad; y += step {
		for x := bounds.X + pad; x < bounds.X+bounds.Width-pad; x += step {
			rl.DrawCircle(int32(x), int32(y), 2, rl.NewColor(255, 255, 255, 28))
		}
	}
}
