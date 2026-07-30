// Package ui (continued) — composable surface title band (Phase C0).
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// SurfaceHeaderMode selects how the title band renders above a RaisedSurface body.
type SurfaceHeaderMode int

const (
	HeaderModeNone SurfaceHeaderMode = iota
	HeaderModeInset
	HeaderModeTitleBar
	HeaderModeGlass
)

// SurfaceHeader is the title strip above a raised surface body. C0: draw only;
// action slots arrive in C2+.
type SurfaceHeader struct {
	Title       string
	TitleHeight float32
	Mode        SurfaceHeaderMode
	// FlatBottom draws a standalone strip (rounded top, square bottom — no body below).
	FlatBottom bool
	// CollapsedOnly draws the title band as a fully rounded pill (all corners).
	CollapsedOnly bool
	// TitleInsetLeft/Right reserve space for collapse/close chrome buttons.
	TitleInsetLeft  float32
	TitleInsetRight float32
}

// Height returns the vertical space reserved above the body (0 when untitled).
func (h SurfaceHeader) Height() float32 {
	if h.Title == "" || h.Mode == HeaderModeNone {
		return 0
	}
	if h.TitleHeight <= 0 {
		return PresetSurfaceTitleHeight
	}
	return h.TitleHeight
}

// DefersUntilPostSheen reports glass-style title draw after chrome sheen.
func (h SurfaceHeader) DefersUntilPostSheen() bool {
	return h.Mode == HeaderModeGlass
}

// Draw paints the header into shell bounds. When deferGlassTitle is true the
// caller runs Draw after DrawOverFill (same rule as legacy Panel/Card glass).
func (h SurfaceHeader) Draw(shellBounds, fillBounds rl.Rectangle, bodyStyle Style, cornerRadius float32) {
	if h.Title == "" || h.Mode == HeaderModeNone {
		return
	}
	titleH := h.Height()
	switch h.Mode {
	case HeaderModeTitleBar:
		if h.CollapsedOnly {
			h.drawTitleBarCollapsedPill(shellBounds, fillBounds, cornerRadius, titleH)
		} else if h.FlatBottom {
			h.drawTitleBarFlatBottom(shellBounds, fillBounds, cornerRadius, titleH)
		} else {
			h.drawTitleBar(shellBounds, fillBounds, bodyStyle, cornerRadius, titleH)
		}
	case HeaderModeInset, HeaderModeGlass:
		h.drawInsetTitle(shellBounds, bodyStyle, titleH)
	}
}

// drawTitleBar paints the dark panel title band and clips title text before chrome buttons.
func (h SurfaceHeader) drawTitleBar(shellBounds, fillBounds rl.Rectangle, bodyStyle Style, cornerRadius float32, titleH float32) {
	titleStyle := GetThemeStyle("panel-title")
	titleRect := rl.NewRectangle(fillBounds.X, fillBounds.Y, fillBounds.Width, titleH)
	if titleRect.Height > fillBounds.Height {
		titleRect.Height = fillBounds.Height
	}
	r := cornerRadius
	if r > 0 {
		shorter := titleRect.Width
		if titleRect.Height < shorter {
			shorter = titleRect.Height
		}
		rTitle := r / (shorter / 2)
		if rTitle > 1 {
			rTitle = 1
		}
		rl.DrawRectangleRounded(titleRect, rTitle, 32, titleStyle.BackgroundColor)
		if titleRect.Height > titleH/2 {
			squareRect := rl.NewRectangle(titleRect.X, titleRect.Y+titleH/2, titleRect.Width, titleRect.Height-titleH/2)
			rl.DrawRectangleRec(squareRect, titleStyle.BackgroundColor)
		}
	} else {
		rl.DrawRectangleRec(titleRect, titleStyle.BackgroundColor)
	}

	titleDraw := titleStyle
	titlePad := titleStyle.Padding
	if titlePad <= 0 {
		titlePad = 10
	}
	textLeft := shellBounds.X + titlePad + h.TitleInsetLeft + 4
	textRight := shellBounds.X + shellBounds.Width - h.TitleInsetRight - 4
	if textRight > textLeft+8 {
		clip := rl.NewRectangle(textLeft, titleRect.Y, textRight-textLeft, titleRect.Height)
		beginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
		textY := TextPosY(titleRect, titleDraw) - 1
		drawTextS(h.Title, int32(textLeft+0.5), textY, titleDraw)
		rl.EndScissorMode()
	}

	sepY := int32(titleRect.Y + titleH)
	rl.DrawLine(int32(titleRect.X), sepY, int32(titleRect.X+titleRect.Width), sepY, rl.NewColor(0, 0, 0, 28))
}

// drawTitleBarCollapsedPill paints header-only collapsed chrome (rounded on all corners).
func (h SurfaceHeader) drawTitleBarCollapsedPill(shellBounds, fillBounds rl.Rectangle, cornerRadius float32, titleH float32) {
	titleStyle := GetThemeStyle("panel-title")
	titleRect := rl.NewRectangle(fillBounds.X, fillBounds.Y, fillBounds.Width, fillBounds.Height)
	if titleRect.Height > fillBounds.Height {
		titleRect.Height = fillBounds.Height
	}
	r := cornerRadius
	if r > 0 {
		shorter := titleRect.Width
		if titleRect.Height < shorter {
			shorter = titleRect.Height
		}
		rTitle := r / (shorter / 2)
		if rTitle > 1 {
			rTitle = 1
		}
		rl.DrawRectangleRounded(titleRect, rTitle, 32, titleStyle.BackgroundColor)
	} else {
		rl.DrawRectangleRec(titleRect, titleStyle.BackgroundColor)
	}
	titleDraw := titleStyle
	titlePad := titleStyle.Padding
	if titlePad <= 0 {
		titlePad = 10
	}
	textLeft := shellBounds.X + titlePad + h.TitleInsetLeft + 4
	textRight := shellBounds.X + shellBounds.Width - h.TitleInsetRight - 4
	if textRight > textLeft+8 {
		clip := rl.NewRectangle(textLeft, titleRect.Y, textRight-textLeft, titleRect.Height)
		beginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
		textY := TextPosY(titleRect, titleDraw) - 1
		drawTextS(h.Title, int32(textLeft+0.5), textY, titleDraw)
		rl.EndScissorMode()
	}
}

// drawTitleBarFlatBottom paints a standalone header strip: rounded top, square bottom.
func (h SurfaceHeader) drawTitleBarFlatBottom(shellBounds, fillBounds rl.Rectangle, cornerRadius float32, titleH float32) {
	titleStyle := GetThemeStyle("panel-title")
	titleRect := rl.NewRectangle(fillBounds.X, fillBounds.Y, fillBounds.Width, titleH)
	if titleRect.Height > fillBounds.Height {
		titleRect.Height = fillBounds.Height
	}
	r := cornerRadius
	if r > 0 {
		shorter := titleRect.Width
		if titleRect.Height < shorter {
			shorter = titleRect.Height
		}
		rTitle := r / (shorter / 2)
		if rTitle > 1 {
			rTitle = 1
		}
		rl.DrawRectangleRounded(titleRect, rTitle, 32, titleStyle.BackgroundColor)
		if titleRect.Height > titleH/2 {
			squareRect := rl.NewRectangle(titleRect.X, titleRect.Y+titleH/2, titleRect.Width, titleRect.Height-titleH/2)
			rl.DrawRectangleRec(squareRect, titleStyle.BackgroundColor)
		}
	} else {
		rl.DrawRectangleRec(titleRect, titleStyle.BackgroundColor)
	}
	titleDraw := titleStyle
	titlePad := titleStyle.Padding
	if titlePad <= 0 {
		titlePad = 10
	}
	textLeft := shellBounds.X + titlePad + h.TitleInsetLeft + 4
	textRight := shellBounds.X + shellBounds.Width - h.TitleInsetRight - 4
	if textRight > textLeft+8 {
		clip := rl.NewRectangle(textLeft, titleRect.Y, textRight-textLeft, titleRect.Height)
		beginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
		textY := TextPosY(titleRect, titleDraw) - 1
		drawTextS(h.Title, int32(textLeft+0.5), textY, titleDraw)
		rl.EndScissorMode()
	}
}

func (h SurfaceHeader) drawInsetTitle(shellBounds rl.Rectangle, bodyStyle Style, titleH float32) {
	titleStyle := GetThemeStyle("card-title")
	titlePad := bodyStyle.Padding
	if titlePad <= 0 {
		titlePad = PresetSurfacePadding
	}
	titleFont := titleStyle.FontSize
	if bodyStyle.FontSize > 0 {
		titleFont = bodyStyle.FontSize
	}
	textY := int32(shellBounds.Y) + (int32(titleH)-titleFont)/2
	titleDraw := titleStyle
	titleDraw.FontSize = titleFont
	if bodyStyle.TextColor.A != 0 {
		titleDraw.TextColor = bodyStyle.TextColor
	}
	drawTextS(h.Title, int32(shellBounds.X+titlePad+0.5), textY, titleDraw)

	sepY := int32(shellBounds.Y + titleH)
	rl.DrawLine(int32(shellBounds.X), sepY, int32(shellBounds.X+shellBounds.Width), sepY, rl.NewColor(0, 0, 0, 28))
}

func (h SurfaceHeader) drawGlassTitle(shellBounds rl.Rectangle, bodyStyle Style, titleH float32) {
	titleStyle := GetThemeStyle("panel-title")
	titleDraw := titleStyle
	titlePad := bodyStyle.Padding
	if titlePad <= 0 {
		titlePad = PresetSurfacePadding
	}
	if bodyStyle.TextColor.A != 0 {
		titleDraw.TextColor = bodyStyle.TextColor
	}
	titleFont := titleDraw.FontSize
	if bodyStyle.FontSize > 0 {
		titleFont = bodyStyle.FontSize
	}
	titleDraw.FontSize = titleFont
	textX := int32(shellBounds.X + titlePad + 0.5)
	textY := int32(shellBounds.Y) + (int32(titleH)-titleFont)/2
	drawTextS(h.Title, textX, textY, titleDraw)
}

// DrawGlass paints the glass preset title (no bar, no separator) after sheen.
func (h SurfaceHeader) DrawGlass(shellBounds rl.Rectangle, bodyStyle Style) {
	if h.Title == "" {
		return
	}
	h.drawGlassTitle(shellBounds, bodyStyle, h.Height())
}

// drawSurfaceChevron draws a chevron rotated by progress (0=▶ collapsed, 1=▼ expanded).
func drawSurfaceChevron(cx, cy, size, progress float32, col rl.Color) {
	angle := progress * 90
	rad := angle * (3.14159265 / 180)
	cos, sin := surfaceCossin(rad)
	halfSize := size * 0.5
	tipX := cx + cos*halfSize
	tipY := cy + sin*halfSize
	topX := cx + cos*(-halfSize) - sin*(-halfSize)
	topY := cy + sin*(-halfSize) + cos*(-halfSize)
	botX := cx + cos*(-halfSize) - sin*halfSize
	botY := cy + sin*(-halfSize) + cos*halfSize
	rl.DrawTriangle(
		rl.NewVector2(tipX, tipY),
		rl.NewVector2(botX, botY),
		rl.NewVector2(topX, topY),
		col,
	)
}

func surfaceCossin(rad float32) (c, s float32) {
	x := float64(rad)
	x2 := x * x
	x3 := x2 * x
	x4 := x2 * x2
	x5 := x4 * x
	x6 := x4 * x2
	x7 := x6 * x
	sinV := x - x3/6 + x5/120 - x7/5040
	cosV := 1 - x2/2 + x4/24 - x6/720
	return float32(cosV), float32(sinV)
}

func brightenColor(col rl.Color, delta uint8) rl.Color {
	r := int(col.R) + int(delta)
	g := int(col.G) + int(delta)
	b := int(col.B) + int(delta)
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}
	return rl.NewColor(uint8(r), uint8(g), uint8(b), col.A)
}
