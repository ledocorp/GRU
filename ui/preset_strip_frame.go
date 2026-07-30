// Package ui — preset comparison strip: padded flex rows + static gradient fill.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// PresetStripFrame wraps preset tiles with padding and a gradient background.
// Unlike a scroll host, it paints one fill rect from layout bounds (no clip).
type PresetStripFrame struct {
	Container
}

func newPresetStripFrame(id string, opts PresetRowOptions) *PresetStripFrame {
	c := NewContainer(id, 0, 0, 0, 0)
	c.SetStyle("transparent")
	c.Gap = opts.Gap
	c.AutoHeight = true
	pad := PresetBackdropPadding
	c.mergeStylePatch(styleJSON{Padding: &pad})
	return &PresetStripFrame{Container: *c}
}

func (p *PresetStripFrame) fillRect() rl.Rectangle {
	b := p.Bounds()
	pad := p.GetStyle().Padding
	w := b.Width - 2*pad
	h := b.Height - 2*pad
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return rl.NewRectangle(b.X+pad, b.Y+pad, w, h)
}

// Draw paints the gradient band then tiles.
func (p *PresetStripFrame) Draw() {
	if p.IsHidden() {
		return
	}
	drawDemoBackdropFill(p.fillRect())
	p.Container.drawInternal()
}
