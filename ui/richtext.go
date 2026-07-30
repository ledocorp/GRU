// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"
	"unicode"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TextSpan is a styled run of text inside RichText.
type TextSpan struct {
	Text    string
	Style   string
	Variant string
	Bold    bool
	Italic  bool
	Strike  bool
	Color   rl.Color
	Link      string
	LinkTitle string
}

// RichText is the first retained document text primitive. It renders multiple
// styled spans as one node, supports word wrapping, and can grow intrinsically
// when AutoHeight is enabled for parent Viewport scrolling.
//
// # LLM Prompt Template
//
//	rt := ui.NewRichText("hint", []ui.TextSpan{
//	    {Text: "Body copy with ", Variant: "muted"},
//	    {Text: "bold", Bold: true},
//	}, 0, 0, 0, 0)
//	rt.Wrap = true
//	vp.AddChild(rt)
//
// Demo scenes: **Batch 3 Live Demo**, **Document Theme Demo**, **Markdown Preview Demo**.
type RichText struct {
	Element
	Spans       []TextSpan
	Wrap        bool
	LineGap     float32
	Selectable  bool
	OnLinkClick func(link string)

	hoveredLink      string
	hoveredLinkTitle string
	selecting    bool
	selectAnchor int
	selectStart  int
	selectEnd    int
	selectPress  bool // mouse down in text; selection starts only after drag
	selectPressMouse rl.Vector2
	selectDragging   bool

	lastMeasuredW float32 // width used for last successful AutoHeight layout
}

// SetBounds marks measure stale when width changes so wrapped AutoHeight blocks
// reflow before draw (fixes preview overlap during pane/window resize).
func (rt *RichText) SetBounds(r rl.Rectangle) {
	rt.invalidateMeasureIfWidthChanged(r.Width)
	rt.Element.SetBounds(r)
}

func (rt *RichText) setBoundsNoMark(r rl.Rectangle) {
	rt.invalidateMeasureIfWidthChanged(r.Width)
	rt.Element.setBoundsNoMark(r)
}

func (rt *RichText) invalidateMeasureIfWidthChanged(newW float32) {
	if rt.bounds.Width > 0 && absF(newW-rt.bounds.Width) > 0.5 {
		rt.lastMeasuredW = 0
	}
}

// InvalidateAutoHeightMeasure forces a height remeasure on the next Layout pass.
func (rt *RichText) InvalidateAutoHeightMeasure() {
	rt.lastMeasuredW = 0
	rt.MarkDirty()
}

type richTextToken struct {
	index     int
	text      string
	style     Style
	link      string
	linkTitle string
	italic    bool
	strike    bool
	code      bool
}

// NewRichText creates a retained rich text block.
func NewRichText(id string, spans []TextSpan, x, y, w, h float32) *RichText {
	rt := &RichText{
		Element: NewElement(id, x, y, w, h),
		Spans:   spans,
		Wrap:    true,
		LineGap: 4,
	}
	rt.selectStart = -1
	rt.selectEnd = -1
	rt.selectAnchor = -1
	rt.styleName = "richtext"
	if h == 0 {
		rt.AutoHeight = true
	}
	return rt
}

// SetSpans replaces the text runs and invalidates layout/draw.
func (rt *RichText) SetSpans(spans []TextSpan) {
	rt.Spans = spans
	rt.InvalidateAutoHeightMeasure()
	rt.MarkDirty()
}

// Layout updates intrinsic height when AutoHeight is enabled. When a parent
// panel/card has already assigned a shorter height (fixed body clamp), height
// does not expand past those bounds.
func (rt *RichText) Layout() {
	defer func() { rt.layoutDirty = false }()
	if rt.IsHidden() || !rt.IsAutoHeight() {
		return
	}
	b := rt.Bounds()
	var m flexTextMeasure
	m.lastW = rt.lastMeasuredW
	res := applyFlexAutoHeightLayout(rt, &m, b, rt.measureHeight)
	rt.lastMeasuredW = m.lastW
	if !res.applied {
		return
	}
	want := res.height
	if !rt.AutoHeight && b.Height > 0 && want > b.Height+0.5 {
		res.bounds.Height = b.Height
	}
	rt.setBoundsNoMark(res.bounds)
	if res.hostDirty {
		markAutoHeightLayoutHostDirty(rt)
	}
}

// Update implements Node.Update.
func (rt *RichText) Update(_ float32) {
	if rt.IsHidden() || (!rt.hasLinks() && !rt.Selectable) {
		return
	}
	mouse := rl.GetMousePosition()
	link, title := rt.linkInfoAt(mouse)
	if link != rt.hoveredLink || title != rt.hoveredLinkTitle {
		rt.hoveredLink = link
		rt.hoveredLinkTitle = title
	}
	if link != "" {
		rl.SetMouseCursor(rl.MouseCursorPointingHand)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rt.OnLinkClick != nil {
			rt.ClearSelection()
			rt.OnLinkClick(link)
			return
		}
	}
	if rt.Selectable {
		rt.updateSelection(mouse)
	}
	if link != "" {
		// cursor already set above
	} else if rt.Selectable && rt.pointOverTextGlyph(mouse) {
		rl.SetMouseCursor(rl.MouseCursorIBeam)
	}
}

func (rt *RichText) usesFlexPlainTextPad() bool {
	return rt.IsAutoHeight() && !rt.Selectable && styleUsesFlexPlainText(rt.styleName)
}

// Draw renders the rich text block.
func (rt *RichText) Draw() {
	defer func() { rt.drawDirty = false }()
	if rt.IsHidden() {
		return
	}
	style := rt.GetStyle()
	b := rt.Bounds()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, style.BackgroundColor)
	}
	if style.BorderWidth > 0 {
		rl.DrawRectangleLinesEx(b, style.BorderWidth, style.BorderColor)
	}

	clip := effectiveLabelClip(b)
	hasViewportClip := false
	if ancestorClip, ok := ancestorClipBounds(rt); ok {
		clip = intersectRects(clip, ancestorClip)
		hasViewportClip = true
	}
	hasBodyClip := false
	if bodyClip, ok := surfaceBodyContentClip(rt); ok {
		clip = intersectRects(clip, bodyClip)
		if rt.Wrap || rt.Selectable || b.Y < bodyClip.Y-0.5 || b.Y+b.Height > bodyClip.Y+bodyClip.Height+0.5 {
			hasBodyClip = true
		}
	}
	useScissor := clip.Width > 0 && clip.Height > 0 &&
		(rt.Wrap || rt.Selectable || hasViewportClip || hasBodyClip || hasActiveDrawClip)
	if useScissor {
		beginScissorFromRect(clip)
	}
	rt.drawRuns(b)
	if useScissor {
		rl.EndScissorMode()
	}
}

// UsesScissor implements Node.UsesScissor.
func (rt *RichText) UsesScissor() bool {
	return rt.Wrap || rt.Selectable
}

// IsInteractive implements Node.IsInteractive.
func (rt *RichText) IsInteractive() bool { return rt.hasLinks() || rt.Selectable }

// contentBand is the vertical slice used for line layout inside bounds.
// Wrapped paragraph text is top-aligned under the body padding. Vertical centering
// applies only to single-line flex-row markers (!Wrap && !AutoHeight).
func (rt *RichText) contentBand(bounds rl.Rectangle) rl.Rectangle {
	style := rt.GetStyle()
	pad := style.Padding
	if rt.usesFlexPlainTextPad() {
		blockH := rt.linesBlockHeight(bounds.Width)
		bandH := bounds.Height - pad*2
		topY := bounds.Y + pad + flexTextTopPad(bandH, blockH)
		return rl.NewRectangle(bounds.X+pad, topY, bounds.Width-pad*2, blockH)
	}
	innerH := bounds.Height - pad*2
	if innerH < 1 {
		innerH = bounds.Height
	}
	contentH := rt.measureHeight(bounds.Width)
	if !rt.AutoHeight && innerH > 0 && contentH > innerH {
		contentH = innerH
	}
	if !rt.Wrap && !rt.AutoHeight && contentH > 0 && contentH < innerH {
		slack := innerH - contentH
		return rl.NewRectangle(bounds.X+pad, bounds.Y+pad+slack/2, bounds.Width-pad*2, contentH)
	}
	return rl.NewRectangle(bounds.X+pad, bounds.Y+pad, bounds.Width-pad*2, innerH)
}

func (rt *RichText) tokenLineHeight(st Style) float32 {
	return EffectiveFontSize(st) + rt.LineGap
}

func (rt *RichText) drawRuns(b rl.Rectangle) {
	style := rt.GetStyle()
	rt.forEachTokenRect(b, func(tok richTextToken, rect rl.Rectangle) bool {
		if rt.tokenSelected(tok.index) {
			rl.DrawRectangleRec(rect, rl.NewColor(79, 70, 229, 42))
		}
		fs := EffectiveFontSize(tok.style)
		var drawY float32
		if rt.styleName == "statusbar-label" {
			drawY = float32(statusBarTextPosY(rect, tok.style))
		} else {
			drawY = rect.Y + (rect.Height-fs)*0.5
		}
		if tok.code && tok.text != "" && tok.text != "\n" {
			padY := float32(2)
			padX := float32(2)
			bg := tok.style.BackgroundColor
			if bg.A == 0 {
				bg = rl.NewColor(237, 233, 254, 255)
			}
			rl.DrawRectangleRec(rl.NewRectangle(rect.X+padX, rect.Y+padY, rect.Width-padX*2, rect.Height-padY*2), bg)
		}
		if !richTextSkipDraw(tok.text) {
			drawTextS(tok.text, int32(rect.X), int32(drawY), tok.style)
		}
		if tok.strike && tok.text != "" && tok.text != "\n" {
			mid := rect.Y + rect.Height*0.55
			rl.DrawRectangle(int32(rect.X), int32(mid), int32(rect.Width), 1, tok.style.TextColor)
		}
		if tok.link != "" {
			underline := tok.style.TextColor
			if tok.link == rt.hoveredLink {
				underline = rl.ColorBrightness(underline, 0.18)
			}
			rl.DrawRectangle(int32(rect.X), int32(rect.Y+rect.Height-2), int32(rect.Width), 1, underline)
		}
		return true
	})
	_ = style
}

// MeasureContentWidth returns the pixel width of the longest line when Wrap is false.
// Used by code fences inside horizontal scroll hosts.
func (rt *RichText) MeasureContentWidth() float32 {
	style := rt.GetStyle()
	pad := style.Padding * 2
	maxLine := float32(0)
	lineW := float32(0)
	for _, tok := range rt.tokens() {
		if tok.text == "\n" {
			if lineW > maxLine {
				maxLine = lineW
			}
			lineW = 0
			continue
		}
		lineW += float32(measureTextS(tok.text, tok.style))
	}
	if lineW > maxLine {
		maxLine = lineW
	}
	if maxLine < 8 {
		maxLine = 8
	}
	return maxLine + pad
}

func (rt *RichText) linesBlockHeight(width float32) float32 {
	style := rt.GetStyle()
	pad := style.Padding
	maxW := width - pad*2
	if maxW < 8 {
		maxW = 8
	}
	baseLH := rt.tokenLineHeight(style)
	lineMax := baseLH
	linesH := float32(0)
	x := float32(0)
	for _, tok := range rt.tokens() {
		if tok.text == "\n" {
			linesH += lineMax
			lineMax = baseLH
			x = 0
			continue
		}
		w := float32(measureTextS(tok.text, tok.style))
		lh := rt.tokenLineHeight(tok.style)
		if lh > lineMax {
			lineMax = lh
		}
		if rt.Wrap && x > 0 && x+w > maxW {
			linesH += lineMax
			lineMax = lh
			x = 0
			if strings.TrimSpace(tok.text) == "" {
				continue
			}
		}
		x += w
	}
	if lineMax > 0 {
		linesH += lineMax
	}
	if linesH < baseLH {
		linesH = baseLH
	}
	return linesH
}

func (rt *RichText) measureHeight(width float32) float32 {
	style := rt.GetStyle()
	h := rt.linesBlockHeight(width) + style.Padding*2
	if rt.usesFlexPlainTextPad() {
		h += flexTextPadY
	}
	return h
}

func (rt *RichText) forEachTokenRect(b rl.Rectangle, fn func(richTextToken, rl.Rectangle) bool) {
	style := rt.GetStyle()
	band := rt.contentBand(b)
	maxW := band.Width
	if maxW < 8 {
		maxW = 8
	}
	baseLH := rt.tokenLineHeight(style)
	x := band.X
	y := band.Y
	lineStartX := x
	maxX := band.X + band.Width
	lineMax := baseLH

	flushLine := func() {
		y += lineMax
		lineMax = baseLH
	}

	for _, tok := range rt.tokens() {
		lh := rt.tokenLineHeight(tok.style)
		if lh > lineMax {
			lineMax = lh
		}
		if tok.text == "\n" {
			x = lineStartX
			flushLine()
			continue
		}
		w := float32(measureTextS(tok.text, tok.style))
		if rt.Wrap && x > lineStartX && x+w > maxX {
			x = lineStartX
			flushLine()
			if lh > lineMax {
				lineMax = lh
			}
			if strings.TrimSpace(tok.text) == "" {
				continue
			}
		}
		rect := rl.NewRectangle(x, y, w, lineMax)
		if !fn(tok, rect) {
			return
		}
		x += w
	}
}

func (rt *RichText) linkInfoAt(point rl.Vector2) (string, string) {
	if !rt.pointInTextClip(point) {
		return "", ""
	}
	var foundLink, foundTitle string
	rt.forEachTokenRect(rt.Bounds(), func(tok richTextToken, rect rl.Rectangle) bool {
		if tok.link != "" && rl.CheckCollisionPointRec(point, rect) {
			foundLink = tok.link
			foundTitle = tok.linkTitle
			return false
		}
		return true
	})
	return foundLink, foundTitle
}

func (rt *RichText) tokenAt(point rl.Vector2) int {
	if !rt.pointInTextContent(point) {
		return -1
	}
	found := -1
	rt.forEachTokenRect(rt.Bounds(), func(tok richTextToken, rect rl.Rectangle) bool {
		if tok.text == "\n" || !rl.CheckCollisionPointRec(point, rect) {
			return true
		}
		found = tok.index
		return false
	})
	return found
}

// tokenAtLineX picks the token under the cursor's X on the current line using
// layout rects (including whitespace tokens). This matches click-and-drag from
// the exact horizontal position instead of snapping to a distant word edge.
func (rt *RichText) tokenAtLineX(point rl.Vector2) int {
	if !rt.pointInTextContent(point) {
		return -1
	}
	type lineToken struct {
		index int
		rect  rl.Rectangle
	}
	var line []lineToken
	rt.forEachTokenRect(rt.Bounds(), func(tok richTextToken, rect rl.Rectangle) bool {
		if tok.text == "\n" {
			return true
		}
		if point.Y < rect.Y || point.Y > rect.Y+rect.Height {
			return true
		}
		line = append(line, lineToken{index: tok.index, rect: rect})
		return true
	})
	if len(line) == 0 {
		return -1
	}
	for i := 1; i < len(line); i++ {
		cur := line[i]
		j := i - 1
		for j >= 0 && line[j].rect.X > cur.rect.X {
			line[j+1] = line[j]
			j--
		}
		line[j+1] = cur
	}
	for _, entry := range line {
		if point.X < entry.rect.X+entry.rect.Width {
			return entry.index
		}
	}
	return line[len(line)-1].index
}

func (rt *RichText) pointInTextContent(point rl.Vector2) bool {
	b := rt.textContentBounds()
	if b.Width <= 0 || b.Height <= 0 || !rl.CheckCollisionPointRec(point, b) {
		return false
	}
	if ancestorClip, ok := ancestorClipBounds(rt); ok {
		clip := intersectRects(b, ancestorClip)
		if clip.Width <= 0 || clip.Height <= 0 || !rl.CheckCollisionPointRec(point, clip) {
			return false
		}
	}
	return true
}

// pointOverTextGlyph is true when point hits a drawn text token (not empty padding).
func (rt *RichText) pointOverTextGlyph(point rl.Vector2) bool {
	if !rt.pointInTextClip(point) {
		return false
	}
	over := false
	rt.forEachTokenRect(rt.Bounds(), func(tok richTextToken, rect rl.Rectangle) bool {
		if tok.text == "\n" {
			return true
		}
		if rl.CheckCollisionPointRec(point, rect) {
			over = true
			return false
		}
		return true
	})
	return over
}

func (rt *RichText) textContentBounds() rl.Rectangle {
	return rt.contentBand(rt.Bounds())
}

func (rt *RichText) pointInTextClip(point rl.Vector2) bool {
	clip := rt.Bounds()
	if bodyClip, ok := surfaceBodyContentClip(rt); ok {
		clip = intersectRects(clip, bodyClip)
	}
	if ancestorClip, ok := ancestorClipBounds(rt); ok {
		clip = intersectRects(clip, ancestorClip)
	}
	if clip.Width <= 0 || clip.Height <= 0 || !rl.CheckCollisionPointRec(point, clip) {
		return false
	}
	return true
}

func (rt *RichText) hasLinks() bool {
	for _, span := range rt.Spans {
		if span.Link != "" {
			return true
		}
	}
	return false
}

func (rt *RichText) tokens() []richTextToken {
	out := make([]richTextToken, 0, len(rt.Spans)*4)
	for _, span := range rt.Spans {
		st := rt.spanStyle(span)
		for _, part := range splitRichTextTokens(span.Text) {
			out = append(out, richTextToken{
				index:     len(out),
				text:      part,
				style:     st,
				link:      span.Link,
				linkTitle: span.LinkTitle,
				italic:    span.Italic,
				strike:    span.Strike,
				code:      span.Variant == "code",
			})
		}
	}
	if len(out) == 0 {
		out = append(out, richTextToken{index: 0, text: "", style: rt.GetStyle()})
	}
	return out
}

func (rt *RichText) spanStyle(span TextSpan) Style {
	st := rt.GetStyle()
	if span.Style != "" {
		if s, ok := CurrentTheme[span.Style]; ok {
			st = mergeStyle(st, s)
		}
	}
	if span.Variant != "" {
		if s, ok := CurrentTheme["richtext-"+span.Variant]; ok {
			st = mergeStyle(st, s)
		} else if s, ok := CurrentTheme[span.Variant]; ok {
			st = mergeStyle(st, s)
		}
	}
	if span.Bold {
		st.Bold = true
		if s, ok := CurrentTheme["richtext-preview-bold"]; ok && st.PreviewFont {
			st = mergeStyle(st, s)
		}
	}
	if span.Italic {
		st.Italic = true
		switch rt.StyleName() {
		case "richtext-blockquote":
			if s, ok := CurrentTheme["richtext-blockquote-italic"]; ok {
				st = mergeStyle(st, s)
			}
		default:
			if s, ok := CurrentTheme["richtext-italic"]; ok {
				st = mergeStyle(st, s)
			}
			if s, ok := CurrentTheme["richtext-preview-italic"]; ok && st.PreviewFont {
				st = mergeStyle(st, s)
			}
		}
	}
	// Preview bold/italic use dedicated Inter faces at the same nominal size as body text.
	if span.Variant == "code" || span.Variant == "code-block" || span.Variant == "math" || span.Variant == "math-display" {
		st.Mono = true
	}
	if span.Strike {
		if s, ok := CurrentTheme["richtext-strike"]; ok {
			st = mergeStyle(st, s)
		}
	}
	if span.Color.A != 0 {
		st.TextColor = span.Color
	}
	if span.Link != "" {
		if s, ok := CurrentTheme["richtext-link"]; ok {
			st = mergeStyle(st, s)
		}
	}
	if span.Variant == "footnote-back" {
		if s, ok := CurrentTheme["richtext-footnote-back"]; ok {
			st = mergeStyle(st, s)
		}
	}
	if span.Variant == "math" {
		if s, ok := CurrentTheme["richtext-math"]; ok {
			st = mergeStyle(st, s)
		}
	}
	if span.Variant == "math-display" {
		if s, ok := CurrentTheme["richtext-math-display"]; ok {
			st = mergeStyle(st, s)
		}
	}
	if base := rt.GetStyle(); base.PreviewFont {
		st.PreviewFont = true
	}
	return st
}

// richTextMaxSplitLen avoids OOM from millions of word tokens on huge code blocks.
const richTextMaxSplitLen = 8192

// richTextSkipDraw avoids drawing lone CR tokens that render as '?' in the font atlas.
func richTextSkipDraw(text string) bool {
	if text == "" || text == "\n" {
		return true
	}
	for _, r := range text {
		if r != '\r' {
			return false
		}
	}
	return true
}

func splitRichTextTokens(text string) []string {
	if len(text) > richTextMaxSplitLen {
		return []string{text}
	}
	var out []string
	var b strings.Builder
	var inSpace bool
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range text {
		if r == '\n' {
			flush()
			out = append(out, "\n")
			inSpace = false
			continue
		}
		isSpace := unicode.IsSpace(r)
		if b.Len() > 0 && isSpace != inSpace {
			flush()
		}
		inSpace = isSpace
		b.WriteRune(r)
	}
	flush()
	return out
}

// InteractionOverlayActive implements InteractionOverlayPainter (link underline).
func (rt *RichText) InteractionOverlayActive() bool {
	return !rt.IsHidden() && rt.hoveredLink != ""
}

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (rt *RichText) DrawInteractionOverlay() {
	rt.Draw()
	if rt.hoveredLink != "" && rt.hoveredLinkTitle != "" {
		var anchor rl.Rectangle
		mouse := rl.GetMousePosition()
		rt.forEachTokenRect(rt.Bounds(), func(tok richTextToken, rect rl.Rectangle) bool {
			if tok.link == rt.hoveredLink && rl.CheckCollisionPointRec(mouse, rect) {
				anchor = rect
				return false
			}
			return true
		})
		if anchor.Width > 0 {
			drawTooltipPopup(rt.hoveredLinkTitle, anchor, 1, Tooltips.winW, Tooltips.winH)
		}
	}
}
