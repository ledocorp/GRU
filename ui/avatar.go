// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const avatarDefaultD = float32(40)

// Avatar shows a circular profile image, or initials when ImagePath is empty or fails to load.
//
// Pass explicit width/height for a fixed diameter, or w=0, h=0 for avatarDefaultD.
// Set ShowStatus and StatusOnline for a small presence dot at the bottom-right.
//
// # LLM Prompt Template
//
//	av := ui.NewAvatar("user", "", "SL", 0, 0, 0, 0)
//	av.ShowStatus = true
//	av.StatusOnline = true
//	profileRow.AddChild(av)
//
// Demo scenes: **Settings Demo**.
type Avatar struct {
	Element
	ImagePath    string
	Initials     string
	ShowStatus   bool
	StatusOnline bool
	texture      rl.Texture2D
	loaded       bool
	failed       bool
}

// NewAvatar creates an avatar widget.
func NewAvatar(id, imagePath, initials string, x, y, w, h float32) *Avatar {
	a := &Avatar{
		Element:  NewElement(id, x, y, w, h),
		ImagePath: imagePath,
		Initials:  strings.ToUpper(strings.TrimSpace(initials)),
	}
	a.styleName = "avatar"
	if w == 0 && h == 0 {
		a.SetBounds(rl.NewRectangle(x, y, avatarDefaultD, avatarDefaultD))
	}
	return a
}

// Unload releases a loaded GPU texture.
func (a *Avatar) Unload() {
	if a.loaded {
		rl.UnloadTexture(a.texture)
		a.loaded = false
	}
}

func (a *Avatar) diameter() float32 {
	b := a.Bounds()
	if b.Width > 0 && b.Height > 0 {
		if b.Width < b.Height {
			return b.Width
		}
		return b.Height
	}
	return avatarDefaultD
}

// Layout keeps width and height equal (circle).
func (a *Avatar) Layout() {
	defer func() { a.layoutDirty = false }()
	b := a.Bounds()
	d := a.diameter()
	if b.Width != d || b.Height != d {
		b.Width = d
		b.Height = d
		a.setBoundsNoMark(b)
	}
}

// Draw implements Node.Draw.
func (a *Avatar) Draw() {
	if a.IsHidden() {
		return
	}
	b := a.Bounds()
	d := a.diameter()
	cx := b.X + d/2
	cy := b.Y + d/2
	r := d / 2
	style := a.GetStyle()

	if a.ImagePath != "" {
		a.ensureTexture()
	}
	if a.loaded && a.texture.ID != 0 {
		src := rl.NewRectangle(0, 0, float32(a.texture.Width), float32(a.texture.Height))
		dst := rl.NewRectangle(b.X, b.Y, d, d)
		rl.DrawTexturePro(a.texture, src, dst, rl.Vector2Zero(), 0, rl.White)
	} else {
		fill := style.BackgroundColor
		if fill.A == 0 {
			fill = rl.NewColor(199, 210, 254, 255)
		}
		rl.DrawCircleV(rl.NewVector2(cx, cy), r, fill)
		text := a.Initials
		if text == "" {
			text = "?"
		}
		if len(text) > 2 {
			text = text[:2]
		}
		labelStyle := style
		if labelStyle.FontSize <= 0 {
			labelStyle.FontSize = int32(d * 0.38)
		}
		tw := measureTextS(text, labelStyle)
		drawTextS(text, int32(cx-float32(tw)/2), int32(cy-float32(labelStyle.FontSize)/2), labelStyle)
	}

	if a.ShowStatus {
		dotR := float32(5)
		if d < 32 {
			dotR = 4
		}
		dotX := b.X + d - dotR*1.2
		dotY := b.Y + d - dotR*1.2
		ring := rl.NewColor(255, 255, 255, 255)
		fill := rl.NewColor(156, 163, 175, 255)
		if a.StatusOnline {
			fill = rl.NewColor(34, 197, 94, 255)
		}
		rl.DrawCircleV(rl.NewVector2(dotX, dotY), dotR+1.5, ring)
		rl.DrawCircleV(rl.NewVector2(dotX, dotY), dotR, fill)
	}
}

func (a *Avatar) ensureTexture() {
	if a.loaded || a.failed || a.ImagePath == "" {
		return
	}
	img := rl.LoadImage(a.ImagePath)
	if img.Width == 0 {
		a.failed = true
		return
	}
	a.texture = rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	a.loaded = a.texture.ID != 0
	if !a.loaded {
		a.failed = true
	}
}
