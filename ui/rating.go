// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	ratingDefaultStars = 5
	ratingStarGap      = float32(4)
	ratingStarSize     = float32(24)
	ratingStarMinSize  = float32(16)
	ratingStarFilled   = "★"
	ratingStarEmpty    = "☆"
)

// Rating is a star-based score input (whole stars).
//
// Stars use Phosphor Regular (outline) and Fill (solid) via the icon font
// (U+E46A). Unicode ★/☆ are drawn only when the font glyph is unavailable.
//
// # LLM Prompt Template
//
//	score := ui.NewSignal(float32(3))
//	r := ui.NewRating("stars", score, 5, 0, 0, 0, 0)
//	form.AddChild(r)
//
// Demo scenes: **Batch 12 Rating**, **Settings Demo**.
type Rating struct {
	Element
	MaxStars  int
	Value     *Signal[float32]
	hoverStar int
}

// NewRating creates a rating control. maxStars <= 0 defaults to 5.
// Pass w=0 for intrinsic width (stars * size + gaps).
func NewRating(id string, value *Signal[float32], maxStars int, x, y, w, h float32) *Rating {
	if value == nil {
		value = NewSignal(float32(0))
	}
	if maxStars <= 0 {
		maxStars = ratingDefaultStars
	}
	if h == 0 {
		h = ratingStarSize
	}
	autoW := w == 0
	if autoW {
		n := maxStars
		if n <= 0 {
			n = ratingDefaultStars
		}
		w = float32(n)*ratingStarSize + float32(n-1)*ratingStarGap
	}
	r := &Rating{
		Element:   NewElement(id, x, y, w, h),
		MaxStars:  maxStars,
		Value:     value,
		hoverStar: -1,
	}
	r.styleName = "rating"
	r.Value.Subscribe(func() { r.MarkDrawDirty() })
	if !autoW && r.PreferredWidth <= 0 {
		r.PreferredWidth = w
	}
	return r
}

func (r *Rating) intrinsicWidth() float32 {
	n := r.MaxStars
	if n <= 0 {
		n = ratingDefaultStars
	}
	return float32(n)*ratingStarSize + float32(n-1)*ratingStarGap
}

// starLayout returns the star size and gap for the current bounds, scaling down
// when the flex parent assigned a band narrower than the default strip.
func (r *Rating) starLayout() (starSize, gap float32) {
	starSize = ratingStarSize
	gap = ratingStarGap
	need := r.intrinsicWidth()
	b := r.Bounds()
	if b.Width > 0 && b.Width < need-0.5 {
		scale := b.Width / need
		starSize *= scale
		gap *= scale
		if starSize < ratingStarMinSize {
			starSize = ratingStarMinSize
			gap = ratingStarGap * (ratingStarMinSize / ratingStarSize)
		}
	}
	return starSize, gap
}

func (r *Rating) stripWidth(starSize, gap float32) float32 {
	n := r.MaxStars
	if n <= 0 {
		n = ratingDefaultStars
	}
	return float32(n)*starSize + float32(n-1)*gap
}

// GetPreferredWidth implements flex width hints so a rating row is never
// compressed narrower than its star strip (which would clip stars 2–5).
func (r *Rating) GetPreferredWidth() float32 {
	if r.PreferredWidth > 0 {
		return r.PreferredWidth
	}
	return r.intrinsicWidth()
}

// GetMinWidth allows flex parents to assign a band narrower than the default strip.
func (r *Rating) GetMinWidth() float32 { return 0 }

// IsInteractive implements Node.
func (r *Rating) IsInteractive() bool { return true }

func (r *Rating) starBounds(i int) rl.Rectangle {
	b := r.Bounds()
	starSize, gap := r.starLayout()
	stripW := r.stripWidth(starSize, gap)
	x0 := b.X
	if b.Width > stripW {
		x0 += (b.Width - stripW) / 2
	}
	x := x0 + float32(i)*(starSize+gap)
	return rl.NewRectangle(x, b.Y+(b.Height-starSize)/2, starSize, starSize)
}

func (r *Rating) drawStar(seg rl.Rectangle, filled bool, col rl.Color) {
	weight := PhosphorRegular
	if filled {
		weight = PhosphorFill
	}
	Phosphor.EnsureLoaded(PhosphorStar, PhosphorRegular)
	Phosphor.EnsureLoaded(PhosphorStar, PhosphorFill)
	if Phosphor.Draw(seg, PhosphorStar, weight, col) {
		return
	}
	glyph := ratingStarEmpty
	if filled {
		glyph = ratingStarFilled
	}
	fs := seg.Height * 0.82
	if fs < minRenderPx {
		fs = minRenderPx
	}
	tw := measureTextF(glyph, fs, false, false, false, false)
	x := seg.X + (seg.Width-tw)/2
	y := seg.Y + (seg.Height-fs)/2
	drawTextF(glyph, x, y, fs, col, false, false, false, false)
}

// Update handles star selection and hover preview.
func (r *Rating) Update(_ float32) {
	if r.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	prevHover := r.hoverStar
	r.hoverStar = -1
	for i := 0; i < r.MaxStars; i++ {
		if rl.CheckCollisionPointRec(mouse, r.starBounds(i)) {
			r.hoverStar = i
			break
		}
	}
	if r.hoverStar != prevHover {
		r.MarkDrawDirty()
	}
	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) || r.hoverStar < 0 {
		return
	}
	r.Value.Set(float32(r.hoverStar + 1))
}

// Layout ensures a minimum row height; star width scales down inside narrow bands.
func (r *Rating) Layout() {
	defer func() { r.layoutDirty = false }()
	b := r.Bounds()
	changed := false
	if b.Height < ratingStarSize-0.5 {
		b.Height = ratingStarSize
		changed = true
	}
	if changed {
		r.setBoundsNoMark(b)
	}
}

// Draw implements Node.Draw.
func (r *Rating) Draw() { r.drawInternal() }

func (r *Rating) drawInternal() {
	if r.IsHidden() {
		return
	}
	val := int(r.Value.Get() + 0.5)
	if val < 0 {
		val = 0
	}
	if val > r.MaxStars {
		val = r.MaxStars
	}
	filled := rl.NewColor(250, 204, 21, 255)
	empty := rl.NewColor(180, 188, 200, 255)
	hover := rl.NewColor(253, 224, 71, 255)
	style := r.GetStyle()
	if style.TextColor.A > 0 {
		filled = style.TextColor
	}
	for i := 0; i < r.MaxStars; i++ {
		seg := r.starBounds(i)
		switch {
		case i < val:
			r.drawStar(seg, true, filled)
		case r.hoverStar >= 0 && i <= r.hoverStar:
			r.drawStar(seg, true, hover)
		default:
			r.drawStar(seg, false, empty)
		}
	}
}

// InteractionOverlayActive implements InteractionOverlayPainter.
func (r *Rating) InteractionOverlayActive() bool {
	return !r.IsHidden() && r.hoverStar >= 0
}

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (r *Rating) DrawInteractionOverlay() { r.drawInternal() }
