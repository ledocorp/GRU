// Package ui (continued) — shared surface header chrome geometry (Phase C3).
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	surfaceHeaderBtn   = float32(28)
	surfaceHeaderIcon  = float32(22)
	surfaceHeaderGap   = float32(8)
	surfaceChromeRowH  = float32(32)
)

// surfaceHeaderBandRect is the title/chrome row above the body.
func surfaceHeaderBandRect(sh *SurfaceShell) rl.Rectangle {
	if sh == nil {
		return rl.Rectangle{}
	}
	b := sh.Bounds()
	th := sh.bodyTitleHeight()
	return rl.NewRectangle(b.X, b.Y, b.Width, th)
}

func surfaceCollapseBtnRect(sh *SurfaceShell) rl.Rectangle {
	hr := surfaceHeaderBandRect(sh)
	return rl.NewRectangle(hr.X+surfaceHeaderGap, hr.Y+(hr.Height-surfaceHeaderBtn)/2, surfaceHeaderBtn, surfaceHeaderBtn)
}

func surfaceCloseBtnRect(sh *SurfaceShell) rl.Rectangle {
	hr := surfaceHeaderBandRect(sh)
	return rl.NewRectangle(hr.X+hr.Width-surfaceHeaderBtn-surfaceHeaderGap, hr.Y+(hr.Height-surfaceHeaderBtn)/2, surfaceHeaderBtn, surfaceHeaderBtn)
}

func surfaceHeaderInsetRight(sh *SurfaceShell) float32 {
	if sh == nil {
		return 0
	}
	if sh.dismiss != nil && sh.dismiss.Active() {
		return surfaceHeaderBtn + surfaceHeaderGap
	}
	if pf := sh.panelFeatures; pf != nil && pf.config != nil && pf.config.Closable {
		return surfaceHeaderBtn + surfaceHeaderGap
	}
	return 0
}

func surfaceHeaderInsetLeft(sh *SurfaceShell) float32 {
	if sh == nil {
		return 0
	}
	if pf := sh.panelFeatures; pf != nil && pf.config != nil && pf.config.Collapsible {
		return surfaceHeaderBtn + surfaceHeaderGap
	}
	return 0
}

func drawSurfacePhosphorIcon(dst rl.Rectangle, name string, tint rl.Color) {
	size := surfaceHeaderIcon
	inner := rl.NewRectangle(
		dst.X+(dst.Width-size)/2,
		dst.Y+(dst.Height-size)/2,
		size, size,
	)
	Phosphor.EnsureLoaded(name, PhosphorRegular)
	if !Phosphor.Draw(inner, name, PhosphorRegular, tint) {
		Phosphor.Draw(inner, name, PhosphorFill, tint)
	}
}

func surfaceTitleChromeBackground() rl.Color {
	return GetThemeStyle("panel-title").BackgroundColor
}
