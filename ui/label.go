// Package ui provides a reactive label widget.
// See node.go for the full package documentation.
package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// LabelAlign controls horizontal placement of text inside Label bounds.
type LabelAlign int

const (
	// LabelAlignCenter (default) — each line is centred when Wrap is false.
	LabelAlignCenter LabelAlign = iota
	// LabelAlignLeft — text starts near the left edge.
	LabelAlignLeft
	// LabelAlignRight — text ends near the right edge.
	LabelAlignRight
)

// Label is a non-interactive text display for fixed chrome (status bar, centered
// toolbar captions, single-line metrics). For flex-column copy that wraps on
// resize (form captions, hints, body text), use NewPlainText instead — it is
// a RichText node with the document flex/reflow contract.
//
// Pass h=0 for AutoHeight intrinsic sizing with Wrap enabled by default.
//
// # LLM Prompt Template
//
//	lbl := ui.NewLabel("status", "Ready", 0, 0, 0, 0)
//	sb.AddLeft(lbl)
//
// Prefer NewPlainText for form copy in flex columns. Demo: **StatusBar**, toolbar captions.
type Label struct {
	Element
	Text     *Signal[string]
	Align    LabelAlign
	Truncate bool // single-line scissor clip
	Wrap     bool // multi-line word wrap to bounds.Width
	measure  flexTextMeasure
}

// NewLabel creates a new Label. h=0 enables AutoHeight (intrinsic sizing).
func NewLabel(id, text string, x, y, w, h float32) *Label {
	l := &Label{
		Element: NewElement(id, x, y, w, h),
		Text:    NewSignal(text),
	}
	if h == 0 {
		l.AutoHeight = true
		l.Wrap = true
		l.Align = LabelAlignLeft
	}
	l.SetStyle("label")
	l.Text.Subscribe(func() { l.MarkDirty() })
	return l
}

// Update implements Node.Update (no-op for label).
func (l *Label) Update(dt float32) {}

// SetBounds marks wrap measure stale when width changes (same contract as RichText).
func (l *Label) SetBounds(r rl.Rectangle) {
	l.measure.invalidateIfWidthChanged(l.bounds.Width, r.Width)
	l.Element.SetBounds(r)
}

func (l *Label) setBoundsNoMark(r rl.Rectangle) {
	l.measure.invalidateIfWidthChanged(l.bounds.Width, r.Width)
	l.Element.setBoundsNoMark(r)
}

func (l *Label) ensureWrapForWidth(w float32) {
	if w < 8 {
		return
	}
	if l.Wrap {
		return
	}
	style := l.GetStyle()
	text := l.Text.Get()
	if plainTextOverflowsWidth(text, w, style) || styleUsesFlexPlainText(l.styleName) {
		l.Wrap = true
		l.Align = LabelAlignLeft
	}
}

// Layout sets intrinsic height from one or more wrapped lines.
func (l *Label) Layout() {
	defer func() { l.layoutDirty = false }()
	if l.IsHidden() {
		return
	}
	if l.styleName == "form-label" || l.styleName == "form-field-caption" || l.styleName == "form-value" {
		l.Align = LabelAlignLeft
	}
	l.ensureWrapForWidth(l.Bounds().Width)

	if !l.IsAutoHeight() {
		return
	}

	res := applyFlexAutoHeightLayout(l, &l.measure, l.Bounds(), func(w float32) float32 {
		return l.intrinsicHeight(w)
	})
	if !res.applied {
		return
	}
	l.setBoundsNoMark(res.bounds)
	if res.hostDirty {
		markAutoHeightLayoutHostDirty(l)
	}
}

func (l *Label) intrinsicHeight(width float32) float32 {
	style := l.GetStyle()
	fs := EffectiveFontSize(style)
	if !l.Wrap || width < 24 {
		return EffectiveFontSize(style) + labelPadY
	}
	lines := wrapLabelLines(l.Text.Get(), width, style)
	if len(lines) == 0 {
		return fs + labelPadY
	}
	const lineGap = labelWrappedLineGap
	return float32(len(lines))*fs + float32(len(lines)-1)*float32(lineGap) + labelPadY
}

func wrapLabelLines(text string, maxW float32, style Style) []string {
	if maxW < 8 {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	var cur string
	for _, w := range words {
		trial := w
		if cur != "" {
			trial = cur + " " + w
		}
		if float32(measureTextS(trial, style)) <= maxW-4 {
			cur = trial
			continue
		}
		if cur != "" {
			lines = append(lines, cur)
			cur = w
			continue
		}
		lines = append(lines, w)
		cur = ""
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func labelTopPad(boundsH float32, blockH int32) int32 {
	top := int32((boundsH - float32(blockH)) * 0.5)
	const minTop int32 = 6
	if top < minTop {
		top = minTop
	}
	return top
}

func (l *Label) Draw() {
	defer func() { l.drawDirty = false }()
	l.drawInternal()
}

// labelPadY is total vertical inset (top + bottom) for label bounds.
const labelPadY = flexTextPadY

const labelPadX = int32(2)

func labelHorizPad(style Style) int32 {
	if style.Padding > 0 {
		return int32(style.Padding)
	}
	return labelPadX
}

// labelWrappedLineGap is line spacing for multi-line labels.
const labelWrappedLineGap = int32(4)

func (l *Label) drawInternal() {
	if l.IsHidden() {
		return
	}

	style := l.GetStyle()
	b := l.Bounds()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, style.BackgroundColor)
	}

	text := l.Text.Get()
	fs := EffectiveFontSize(style)
	const lineGap = labelWrappedLineGap

	drawLine := func(line string, y int32) {
		textW := measureTextS(line, style)
		padX := labelHorizPad(style)
		var posX int32
		switch l.Align {
		case LabelAlignLeft:
			posX = int32(b.X) + padX
		case LabelAlignRight:
			posX = int32(b.X+b.Width) - textW - padX
		default:
			posX = int32(b.X) + (int32(b.Width)-textW)/2
		}
		drawTextS(line, posX, y, style)
	}

	clip := effectiveLabelClip(b)
	hasViewportClip := false
	if ancestorClip, ok := ancestorClipBounds(l); ok {
		clip = intersectRects(clip, ancestorClip)
		hasViewportClip = true
	}
	hasBodyClip := false
	if bodyClip, ok := surfaceBodyContentClip(l); ok {
		clip = intersectRects(clip, bodyClip)
		if l.Truncate || l.Wrap || b.Y < bodyClip.Y-0.5 || b.Y+b.Height > bodyClip.Y+bodyClip.Height+0.5 {
			hasBodyClip = true
		}
	}
	useScissor := clip.Width > 0 && clip.Height > 0 && (l.Truncate || l.Wrap || hasViewportClip || hasBodyClip || hasActiveDrawClip)
	if !useScissor && !l.Wrap && b.Width >= 8 && float32(measureTextS(text, style)) > b.Width-4 {
		useScissor = true
	}
	if useScissor {
		beginScissorFromRect(clip)
	}

	if l.Wrap && b.Width >= 24 {
		lines := wrapLabelLines(text, b.Width, style)
		blockH := int32(len(lines))*int32(fs) + int32(len(lines)-1)*lineGap
		startY := int32(b.Y) + labelTopPad(b.Height, blockH)
		for i, line := range lines {
			drawLine(line, startY+int32(i)*(int32(fs)+lineGap))
		}
	} else if !l.Wrap && float32(measureTextS(text, style)) > b.Width-4 && b.Width >= 8 {
		// Single-line overflow: clip to cell instead of painting over siblings.
		drawLine(text, TextPosY(b, style))
	} else if l.styleName == "statusbar-label" {
		drawLine(text, statusBarTextPosY(b, style))
	} else if l.styleName == "toolbar-caption" {
		drawLine(text, toolbarTextPosY(b, style))
	} else {
		drawLine(text, TextPosY(b, style))
	}

	if useScissor {
		rl.EndScissorMode()
	}
}

// UsesScissor implements Node.UsesScissor.
func (l *Label) UsesScissor() bool {
	if l.Truncate || l.Wrap {
		return true
	}
	if hasActiveDrawClip {
		return true
	}
	if _, ok := surfaceBodyContentClip(l); ok {
		return true
	}
	_, ok := ancestorClipBounds(l)
	return ok
}

// IsInteractive implements Node.IsInteractive.
func (l *Label) IsInteractive() bool { return false }
