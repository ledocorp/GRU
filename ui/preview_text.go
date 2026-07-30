// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"
	"unicode/utf8"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TextViewer is a read-only multiline label: newlines are preserved, long
// lines are word-wrapped, and content scrolls when it exceeds the widget bounds
// (document-style preview). Use [TextViewer.SetViewportDocumentMode] with
// height 0 and a parent [Viewport] to let the viewport draw scrollbars.
type TextViewer struct {
	Element
	Text *Signal[string]

	scrollY float32
	scrollX float32

	// When true: no internal wheel/scroll offsets; Layout grows height to fit
	// wrapped text (AutoHeight) so a parent Viewport owns scrolling.
	viewportDoc bool
}

// PreviewText is kept as a transition alias while the document/theme work
// adopts the clearer TextViewer name.
type PreviewText = TextViewer

// NewTextViewer creates a read-only text viewer; default style is "preview-document".
func NewTextViewer(id string, x, y, w, h float32) *TextViewer {
	p := &TextViewer{
		Element: NewElement(id, x, y, w, h),
		Text:    NewSignal(""),
	}
	if h == 0 {
		p.AutoHeight = true
	}
	p.styleName = "preview-document"
	p.Text.Subscribe(func() {
		if !p.viewportDoc {
			p.scrollY, p.scrollX = 0, 0
		}
		p.MarkDirty()
	})
	return p
}

// NewPreviewText is kept for compatibility; new code should call NewTextViewer.
func NewPreviewText(id string, x, y, w, h float32) *TextViewer {
	return NewTextViewer(id, x, y, w, h)
}

// SetViewportDocumentMode configures TextViewer for embedding in a [Viewport]:
// intrinsic height follows wrapped text (use NewTextViewer with h=0), internal
// wheel scrolling is disabled, and the parent viewport shows scrollbars.
func (p *TextViewer) SetViewportDocumentMode(on bool) {
	p.viewportDoc = on
	if on && p.bounds.Height == 0 {
		p.AutoHeight = true
	}
	p.MarkDirty()
}

// ResetScroll jumps both scroll axes to the origin (e.g. when switching files).
func (p *TextViewer) ResetScroll() {
	if p.viewportDoc {
		return
	}
	p.scrollY, p.scrollX = 0, 0
	p.MarkDrawDirty()
}

// Update handles wheel scrolling when the cursor is over this widget.
func (p *TextViewer) Update(_ float32) {
	if p.viewportDoc || p.IsHidden() {
		return
	}
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	wheel := rl.GetMouseWheelMove()
	if wheel == 0 {
		return
	}
	if !rl.CheckCollisionPointRec(rl.GetMousePosition(), p.Bounds()) {
		return
	}
	style := p.GetStyle()
	fs := EffectiveFontSize(style)
	lineH := fs + 2

	maxSY := p.maxScrollY()
	maxSX := p.maxScrollX()
	if shift {
		if maxSX <= 0 {
			return
		}
		p.scrollX -= wheel * lineH * 3
		if p.scrollX < 0 {
			p.scrollX = 0
		}
		if p.scrollX > maxSX {
			p.scrollX = maxSX
		}
	} else {
		if maxSY <= 0 {
			return
		}
		p.scrollY -= wheel * lineH * 3
		if p.scrollY < 0 {
			p.scrollY = 0
		}
		if p.scrollY > maxSY {
			p.scrollY = maxSY
		}
	}
	p.MarkDrawDirty()
}

// HandlesWheelScroll implements wheelConsumer — only claim the wheel when there
// is overflow in the corresponding axis (handled by AbsorbsParentWheel logic).
func (p *TextViewer) HandlesWheelScroll() bool {
	if p.viewportDoc || p.IsHidden() {
		return false
	}
	mouse := rl.GetMousePosition()
	if !rl.CheckCollisionPointRec(mouse, p.Bounds()) {
		return false
	}
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	if shift {
		return p.maxScrollX() > 0
	}
	return p.maxScrollY() > 0
}

// AbsorbsParentWheel implements wheelScrollLimiter so an outer page scrolls
// when this preview is at the end of its range.
func (p *TextViewer) AbsorbsParentWheel(wheel float32) bool {
	if p.viewportDoc || p.IsHidden() {
		return false
	}
	mouse := rl.GetMousePosition()
	if !rl.CheckCollisionPointRec(mouse, p.Bounds()) {
		return false
	}
	const eps = float32(0.5)
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	if shift {
		maxSX := p.maxScrollX()
		if maxSX <= 0 {
			return false
		}
		if wheel < 0 && p.scrollX >= maxSX-eps {
			return false
		}
		if wheel > 0 && p.scrollX <= eps {
			return false
		}
		return true
	}
	maxSY := p.maxScrollY()
	if maxSY <= 0 {
		return false
	}
	if wheel < 0 && p.scrollY >= maxSY-eps {
		return false
	}
	if wheel > 0 && p.scrollY <= eps {
		return false
	}
	return true
}

// Layout sets intrinsic height when in viewport document mode and AutoHeight.
func (p *TextViewer) Layout() {
	defer func() { p.layoutDirty = false }()
	if !p.viewportDoc || !p.IsAutoHeight() {
		return
	}
	b := p.Bounds()
	// Until flex assigns width, use a sane wrap width so height is non-zero on first layout.
	wrapW := b.Width
	if wrapW < 32 {
		wrapW = 520
	}
	h := p.measuredContentHeight(wrapW)
	const minH = float32(160)
	if h < minH {
		h = minH
	}
	if b.Height == h {
		return
	}
	nb := b
	nb.Height = h
	p.setBoundsNoMark(nb)
}

// Draw renders a paper region and wrapped, scrollable text.
func (p *TextViewer) Draw() {
	if p.IsHidden() {
		return
	}
	style := p.GetStyle()
	b := p.Bounds()
	pad := float32(10)
	docInner := rl.NewRectangle(b.X+2, b.Y+2, b.Width-4, b.Height-4)
	paper := rl.NewColor(252, 251, 248, 255)
	border := rl.NewColor(215, 211, 199, 255)
	rl.DrawRectangleRec(docInner, paper)
	rl.DrawRectangleLinesEx(docInner, 1, border)

	maxW := b.Width - 2*pad - 4
	if maxW < 8 {
		maxW = 8
	}
	maxY := b.Y + b.Height - pad
	fs := EffectiveFontSize(style)
	lineGap := float32(2)
	lineH := fs + lineGap
	text := p.Text.Get()

	clip := intersectRects(docInner, b)
	if vp := findViewport(p); vp != nil {
		clip = intersectRects(clip, vp.ClipBounds())
	}
	if clip.Width <= 0 || clip.Height <= 0 {
		return
	}
	beginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))

	var startX, y float32
	if p.viewportDoc {
		startX = b.X + pad + 2
		y = b.Y + pad + 2
	} else {
		startX = b.X + pad + 2 - p.scrollX
		y = b.Y + pad + 2 - p.scrollY
	}

	for _, para := range strings.Split(text, "\n") {
		lines := previewWrapParagraph(para, maxW, style)
		for _, ln := range lines {
			if p.viewportDoc || (y+fs > b.Y+2 && y < maxY) {
				drawTextS(ln, int32(startX), int32(y), style)
			}
			y += lineH
		}
	}
	rl.EndScissorMode()
}

// measuredContentHeight is the total drawn height for the current Text at width w.
func (p *TextViewer) measuredContentHeight(w float32) float32 {
	style := p.GetStyle()
	pad := float32(10)
	maxWi := w - 2*pad - 4
	if maxWi < 8 {
		maxWi = 8
	}
	fs := EffectiveFontSize(style)
	lineH := fs + 2
	var n int
	for _, para := range strings.Split(p.Text.Get(), "\n") {
		n += len(previewWrapParagraph(para, maxWi, style))
	}
	if n < 1 {
		n = 1
	}
	return float32(n)*lineH + 2*pad + 4
}

// UsesScissor implements Node.
func (p *TextViewer) UsesScissor() bool { return true }

// IsInteractive implements Node.
func (p *TextViewer) IsInteractive() bool { return false }

func (p *TextViewer) maxScrollY() float32 {
	if p.viewportDoc {
		return 0
	}
	style := p.GetStyle()
	b := p.Bounds()
	pad := float32(10)
	maxW := b.Width - 2*pad - 4
	if maxW < 8 {
		maxW = 8
	}
	fs := EffectiveFontSize(style)
	lineGap := float32(2)
	lineH := fs + lineGap

	var n int
	for _, para := range strings.Split(p.Text.Get(), "\n") {
		n += len(previewWrapParagraph(para, maxW, style))
	}
	total := float32(n)*lineH + 2*pad + 4
	inner := b.Height - 4
	m := total - inner
	if m < 0 {
		return 0
	}
	return m
}

func (p *TextViewer) maxScrollX() float32 {
	if p.viewportDoc {
		return 0
	}
	style := p.GetStyle()
	b := p.Bounds()
	pad := float32(10)
	maxW := b.Width - 2*pad - 4
	if maxW < 8 {
		maxW = 8
	}
	var wMax int32
	for _, para := range strings.Split(p.Text.Get(), "\n") {
		for _, ln := range previewWrapParagraph(para, maxW, style) {
			if w := measureTextS(ln, style); w > wMax {
				wMax = w
			}
		}
	}
	inner := b.Width - 2*pad - 4
	if inner < 1 {
		inner = 1
	}
	m := float32(wMax) - inner
	if m < 0 {
		return 0
	}
	return m
}

func previewWrapParagraph(s string, maxW float32, st Style) []string {
	s = strings.TrimRight(s, "\r")
	if strings.TrimSpace(s) == "" {
		return []string{""}
	}
	maxWi := int32(maxW)
	if measureTextS(s, st) <= maxWi {
		return []string{s}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var out []string
	var line string
	flush := func() {
		if line != "" {
			out = append(out, line)
			line = ""
		}
	}
	for _, w := range words {
		trial := line
		if trial != "" {
			trial += " "
		}
		trial += w
		if measureTextS(trial, st) <= maxWi {
			line = trial
			continue
		}
		flush()
		if measureTextS(w, st) <= maxWi {
			line = w
			continue
		}
		out = append(out, previewHardBreakWord(w, maxW, st)...)
	}
	flush()
	if len(out) == 0 {
		out = append(out, "")
	}
	return out
}

func previewHardBreakWord(w string, maxW float32, st Style) []string {
	maxWi := int32(maxW)
	var lines []string
	for len(w) > 0 {
		i := len(w)
		for i > 0 {
			sub := w[:i]
			if measureTextS(sub, st) <= maxWi {
				break
			}
			_, sz := utf8.DecodeLastRuneInString(w[:i])
			if sz == 0 {
				break
			}
			i -= sz
		}
		if i == 0 {
			_, sz := utf8.DecodeRuneInString(w)
			if sz == 0 {
				break
			}
			i = sz
		}
		lines = append(lines, w[:i])
		w = w[i:]
	}
	return lines
}
