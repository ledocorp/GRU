// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const headerTextInset = float32(10)

// headerTitleSubtitleGap is vertical space between title and subtitle lines.
const headerTitleSubtitleGap = float32(8)

// headerWrappedLineGap is line spacing for wrapped subtitle text.
const headerWrappedLineGap = float32(4)

// Header is a non-interactive widget that displays a large title and an
// optional subtitle beneath it.  It is designed to introduce a screen or
// major section and provides visual hierarchy above panels and forms.
//
// Subtitles wrap to the header width (like Label with Wrap). Pass h=0 for
// AutoHeight; Layout uses bounds.Width after the flex parent assigns width.
//
// Styles used:
//
//	"header"          — controls BackgroundColor, TextColor, FontSize (title)
//	"header-subtitle" — controls TextColor and FontSize for the subtitle line
//
// # LLM Prompt Template
//
//	hdr := ui.NewHeader("page-hdr", "Dashboard", "Overview of today's metrics", 0, 0, 0, 0)
//	vp.AddChild(hdr)
//
// Demo scenes: **List Detail Demo**, demo page shells (`page_shell.go`).
type Header struct {
	Element
	Title    string // Primary large text
	Subtitle string // Secondary smaller text; empty = title only
	// WrapSubtitle enables word wrap on the subtitle (default true when subtitle set).
	WrapSubtitle bool
	AccentBar    bool // Draw a 4 px indigo left bar (good for panel section headers)
}

// NewHeader creates a Header widget. Pass an empty subtitle for a title-only
// header. Bounds default to 0,0,0,0 — set them explicitly or let a flex
// parent size the widget.
func NewHeader(id, title, subtitle string, x, y, w, h float32) *Header {
	hdr := &Header{
		Element:      NewElement(id, x, y, w, h),
		Title:        title,
		Subtitle:     subtitle,
		WrapSubtitle: subtitle != "",
		AccentBar:    false,
	}
	hdr.styleName = "header"
	if h == 0 {
		hdr.AutoHeight = true
	}
	return hdr
}

func (h *Header) headerSubtitleStyle(titleStyle Style) Style {
	subStyle, ok := CurrentTheme["header-subtitle"]
	if !ok {
		subStyle = titleStyle
		subStyle.FontSize = titleStyle.FontSize - 6
		if subStyle.FontSize < 12 {
			subStyle.FontSize = 12
		}
		subStyle.TextColor = rl.NewColor(
			titleStyle.TextColor.R,
			titleStyle.TextColor.G,
			titleStyle.TextColor.B,
			180,
		)
	}
	return subStyle
}

func (h *Header) textWidth(bounds rl.Rectangle, pad float32) float32 {
	w := bounds.Width - 2*pad - headerTextInset
	if h.AccentBar {
		w -= 8
	}
	if w < 24 {
		w = 24
	}
	return w
}

func headerTextBlockHeight(text string, style Style, maxW float32, wrap bool) float32 {
	fs := EffectiveFontSize(style)
	if text == "" {
		return 0
	}
	if !wrap || maxW < 24 {
		return fs
	}
	lines := wrapLabelLines(text, maxW, style)
	if len(lines) == 0 {
		return fs
	}
	const lineGap = headerWrappedLineGap
	return float32(len(lines))*fs + float32(len(lines)-1)*lineGap
}

// headerIntrinsicHeight returns the layout height for title + optional subtitle at width.
func headerIntrinsicHeight(title, subtitle string, titleStyle, subStyle Style, width float32, wrapSub bool) float32 {
	pad := titleStyle.Padding
	textW := width - 2*pad - headerTextInset
	if textW < 24 {
		textW = 24
	}
	titleH := headerTextBlockHeight(title, titleStyle, textW, false)
	inner := titleH
	if subtitle != "" {
		inner = titleH + headerTitleSubtitleGap + headerTextBlockHeight(subtitle, subStyle, textW, wrapSub)
	}
	return inner + 2*pad
}

// Update is a no-op — Header is not interactive.
func (h *Header) Update(_ float32) {}

// Layout sets intrinsic height when AutoHeight or when bounds are shorter than content.
func (h *Header) Layout() {
	style := h.GetStyle()
	subStyle := h.headerSubtitleStyle(style)
	textW := h.textWidth(h.Bounds(), style.Padding)
	want := headerIntrinsicHeight(h.Title, h.Subtitle, style, subStyle, textW+2*style.Padding+headerTextInset, h.WrapSubtitle)
	b := h.Bounds()
	if h.IsAutoHeight() || b.Height < want-0.5 {
		if b.Height < want-0.5 || b.Height > want+0.5 {
			b.Height = want
			h.setBoundsNoMark(b)
			// Viewport/panel stacks position siblings from measured heights; bubble up
			// so the next layout pass does not overlap content under a wrapped subtitle.
			if p := h.ParentNode(); p != nil {
				p.MarkDirty()
			}
		}
	}
	h.layoutDirty = false
}

// Draw implements Node.Draw.
func (h *Header) Draw() { h.drawInternal() }

func (h *Header) drawHeaderLines(text string, style Style, x int32, y int32, maxW float32, wrap bool) int32 {
	fs := int32(EffectiveFontSize(style))
	const lineGap = int32(headerWrappedLineGap)
	if !wrap || maxW < 24 {
		drawTextS(text, x, y, style)
		return y + fs
	}
	lines := wrapLabelLines(text, maxW, style)
	for i, line := range lines {
		drawTextS(line, x, y, style)
		y += fs
		if i < len(lines)-1 {
			y += lineGap
		}
	}
	return y
}

func (h *Header) drawInternal() {
	if h.IsHidden() {
		return
	}

	style := h.GetStyle()
	b := h.Bounds()
	pad := style.Padding

	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, style.BackgroundColor)
	}

	textX := int32(b.X) + int32(headerTextInset)
	if h.AccentBar {
		accentColor := rl.NewColor(79, 70, 229, 255)
		rl.DrawRectangle(int32(b.X), int32(b.Y+4), 4, int32(b.Height-8), accentColor)
		textX = int32(b.X) + 18
	}

	maxW := h.textWidth(b, pad)
	subStyle := h.headerSubtitleStyle(style)
	titleFS := EffectiveFontSize(style)

	var blockH int32
	if h.Subtitle == "" {
		blockH = int32(titleFS)
	} else {
		subH := int32(headerTextBlockHeight(h.Subtitle, subStyle, maxW, h.WrapSubtitle))
		blockH = int32(titleFS) + int32(headerTitleSubtitleGap) + subH
	}
	startY := int32(b.Y) + (int32(b.Height)-blockH)/2

	clip := b
	if h.WrapSubtitle && h.Subtitle != "" {
		if ancestorClip, ok := ancestorClipBounds(h); ok {
			clip = intersectRects(clip, ancestorClip)
		}
	}
	useScissor := h.WrapSubtitle && h.Subtitle != "" && clip.Width > 0 && clip.Height > 0
	if useScissor {
		beginScissorFromRect(clip)
	}

	y := startY
	y = h.drawHeaderLines(h.Title, style, textX, y, maxW, false)
	if h.Subtitle != "" {
		y += int32(headerTitleSubtitleGap)
		h.drawHeaderLines(h.Subtitle, subStyle, textX, y, maxW, h.WrapSubtitle)
	}

	if useScissor {
		rl.EndScissorMode()
	}

	rl.DrawLine(
		int32(b.X), int32(b.Y+b.Height-1),
		int32(b.X+b.Width), int32(b.Y+b.Height-1),
		rl.NewColor(0, 0, 0, 18),
	)
}

// UsesScissor implements Node.UsesScissor.
func (h *Header) UsesScissor() bool {
	return h.WrapSubtitle && h.Subtitle != ""
}

// IsInteractive implements Node.IsInteractive.
func (h *Header) IsInteractive() bool { return false }
