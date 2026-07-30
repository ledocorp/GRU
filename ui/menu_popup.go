// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const menuPopupMaxHeight = float32(280)

type menuPopupPlacement struct {
	openAbove bool
	listCap   float32
}

// menuPopupHostRect is the padded client of the nearest vertical Viewport, or
// the screen content band below window chrome when no viewport encloses anchor.
func menuPopupHostRect(anchor Node) rl.Rectangle {
	if anchor != nil {
		for p := anchor.ParentNode(); p != nil; p = p.ParentNode() {
			if vp, ok := p.(*Viewport); ok && !vp.IsHidden() && vp.Orientation != ScrollHorizontal {
				return vp.viewportPaddedClientRect()
			}
		}
	}
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	return OverlayContentBand(sw, sh)
}

// computeMenuPopupPlacement picks mobile-style direction: open downward when the
// list fits below the face; otherwise flip upward (same as native pickers).
func computeMenuPopupPlacement(anchor Node, face rl.Rectangle, leadingChrome, listContentH, optionH float32) menuPopupPlacement {
	host := menuPopupHostRect(anchor)
	spaceBelow := host.Y + host.Height - (face.Y + face.Height)
	spaceAbove := face.Y - host.Y

	minDown := leadingChrome + optionH
	openAbove := false
	if spaceBelow < minDown && spaceAbove > spaceBelow {
		openAbove = true
	}

	space := spaceBelow
	if openAbove {
		space = spaceAbove
	}
	listCap := space - leadingChrome
	if listCap < optionH {
		listCap = optionH
	}
	if listCap > menuPopupMaxHeight {
		listCap = menuPopupMaxHeight
	}
	if listCap > listContentH {
		listCap = listContentH
	}
	return menuPopupPlacement{openAbove: openAbove, listCap: listCap}
}

// menuPopupChromeColors returns panel colors for dropdown/combobox popups from
// the widget style, falling back to the contextmenu theme when unset.
func menuPopupChromeColors(style Style) (panelBg, panelBorder, divider rl.Color) {
	panelBg = style.BackgroundColor
	panelBorder = style.BorderColor
	if panelBg.A == 0 {
		if cm, ok := CurrentTheme["contextmenu"]; ok {
			panelBg = cm.BackgroundColor
		}
		if panelBg.A == 0 {
			panelBg = rl.NewColor(255, 255, 255, 255)
		}
	}
	if panelBorder.A == 0 {
		if cm, ok := CurrentTheme["contextmenu"]; ok {
			panelBorder = cm.BorderColor
		}
		if panelBorder.A == 0 {
			panelBorder = rl.NewColor(200, 200, 212, 255)
		}
	}
	divider = panelBorder
	if hover, ok := CurrentTheme["contextmenu-divider"]; ok && hover.BackgroundColor.A > 0 {
		divider = hover.BackgroundColor
	}
	return panelBg, panelBorder, divider
}

func menuPopupFilterColors(style Style) (idle, focused rl.Color) {
	idle = rl.ColorBrightness(style.BackgroundColor, 0.04)
	focused = rl.ColorBrightness(style.BackgroundColor, -0.06)
	if idle.A == 0 {
		idle = rl.NewColor(248, 249, 252, 255)
		focused = rl.NewColor(237, 239, 254, 255)
	}
	return idle, focused
}

func drawMenuPopupInHost(anchor Node, draw func()) {
	host := menuPopupHostRect(anchor)
	beginScissorFromRect(host)
	draw()
	rl.EndScissorMode()
}

// scrollAncestorViewportsToRevealRect scrolls each vertical ancestor Viewport so
// rect (screen space) fits inside the padded client area. Repositions children
// immediately so popup hit targets stay aligned on the opening frame.
func scrollAncestorViewportsToRevealRect(anchor Node, rect rl.Rectangle) {
	if rect.Width <= 0 || rect.Height <= 0 {
		return
	}
	for p := anchor.ParentNode(); p != nil; p = p.ParentNode() {
		vp, ok := p.(*Viewport)
		if !ok || vp.IsHidden() || vp.Orientation == ScrollHorizontal {
			continue
		}
		client := vp.viewportPaddedClientRect()
		top := client.Y
		bottom := client.Y + client.Height

		if overflow := rect.Y + rect.Height - bottom; overflow > 0 {
			vp.ScrollY += overflow
			if max := vp.overflowScrollY(); vp.ScrollY > max {
				vp.ScrollY = max
			}
			vp.scrollDirty = true
			vp.MarkDirty()
			vp.repositionOnly()
			rect.Y -= overflow
		}
		if underflow := top - rect.Y; underflow > 0 {
			vp.ScrollY -= underflow
			if vp.ScrollY < 0 {
				vp.ScrollY = 0
			}
			vp.scrollDirty = true
			vp.MarkDirty()
			vp.repositionOnly()
			rect.Y += underflow
		}
	}
}

func ensureMenuPopupVisible(anchor Node, popupBounds func() rl.Rectangle) {
	for p := anchor.ParentNode(); p != nil; p = p.ParentNode() {
		if vp, ok := p.(*Viewport); ok && !vp.IsHidden() && vp.Orientation != ScrollHorizontal {
			vp.MarkDirty()
		}
	}
	for pass := 0; pass < 2; pass++ {
		scrollAncestorViewportsToRevealRect(anchor, popupBounds())
	}
}

func clampMenuPopupScroll(scrollY, contentH, viewH float32) float32 {
	max := contentH - viewH
	if max < 0 {
		max = 0
	}
	if scrollY < 0 {
		return 0
	}
	if scrollY > max {
		return max
	}
	return scrollY
}

// menuPopupListWidth returns popup width from face width and longest option label.
func menuPopupListWidth(anchor Node, faceW float32, options []string, style Style) float32 {
	if faceW <= 0 {
		faceW = 120
	}
	pad := style.Padding*2 + 28
	maxW := faceW
	for _, opt := range options {
		w := float32(measureTextS(opt, style)) + pad
		if w > maxW {
			maxW = w
		}
	}
	host := menuPopupHostRect(anchor)
	if host.Width > 0 {
		cap := host.Width * 0.92
		if maxW > cap {
			maxW = cap
		}
	}
	return maxW
}

// clampMenuPopupRect keeps a popup list inside the anchor viewport host.
func clampMenuPopupRect(anchor Node, popup rl.Rectangle) rl.Rectangle {
	host := menuPopupHostRect(anchor)
	if host.Width <= 0 || host.Height <= 0 {
		return popup
	}
	if popup.Width > host.Width {
		popup.Width = host.Width
	}
	if popup.Height > host.Height {
		popup.Height = host.Height
	}
	if popup.X+popup.Width > host.X+host.Width {
		popup.X = host.X + host.Width - popup.Width
	}
	if popup.X < host.X {
		popup.X = host.X
	}
	if popup.Y+popup.Height > host.Y+host.Height {
		popup.Y = host.Y + host.Height - popup.Height
	}
	if popup.Y < host.Y {
		popup.Y = host.Y
	}
	return popup
}
