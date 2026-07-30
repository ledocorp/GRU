// Package ui (continued)
package ui

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// VideoPlaceholder shows poster art and playback chrome without decoding video (Strategy 2 #29).
//
// # LLM Prompt Template
//
//	v := ui.NewVideoPlaceholder("clip", "assets/poster.png", 0, 0, 480, 270)
//	panel.AddChild(v)
//
// Demo scenes: **Batch 10** (media panel).
type VideoPlaceholder struct {
	Element
	PosterPath string
	playing    bool
	poster     *Image
	playBtn    *Button
}

// NewVideoPlaceholder creates a video placeholder with optional poster image path.
func NewVideoPlaceholder(id, posterPath string, x, y, w, h float32) *VideoPlaceholder {
	v := &VideoPlaceholder{
		Element:    NewElement(id, x, y, w, h),
		PosterPath: posterPath,
	}
	v.styleName = "videoplaceholder"
	if posterPath != "" {
		v.poster = NewImage(id+"-poster", posterPath, 0, 0, 0, 0)
		v.poster.SetParent(v)
	}
	v.playBtn = NewButton(id+"-play", "▶ Play (preview)", 0, 0, 140, 36)
	v.playBtn.SetStyle("primary")
	v.playBtn.SetParent(v)
	v.playBtn.OnClick = func() {
		v.playing = !v.playing
		if v.playing {
			v.playBtn.Text.Set("⏸ Preview")
			ShowToast("Video decode not implemented — placeholder only.", ToastInfo, 2*time.Second)
		} else {
			v.playBtn.Text.Set("▶ Play (preview)")
		}
		v.MarkDrawDirty()
	}
	return v
}

// Children implements Node.
func (v *VideoPlaceholder) Children() []Node {
	var ch []Node
	if v.poster != nil {
		ch = append(ch, v.poster)
	}
	if v.playBtn != nil {
		ch = append(ch, v.playBtn)
	}
	return ch
}

func (v *VideoPlaceholder) Layout() {
	if v.IsHidden() {
		return
	}
	b := v.Bounds()
	if v.poster != nil {
		layoutSetBounds(v.poster, b)
		v.poster.Layout()
	}
	if v.playBtn != nil {
		br := v.playBtn.Bounds()
		layoutSetBounds(v.playBtn, rl.NewRectangle(
			b.X+(b.Width-br.Width)/2,
			b.Y+b.Height-48,
			140, 36))
		v.playBtn.Layout()
	}
	v.layoutDirty = false
}

func (v *VideoPlaceholder) Update(dt float32) {
	if v.IsHidden() {
		return
	}
	v.Layout()
	if v.poster != nil {
		v.poster.Update(dt)
	}
	if v.playBtn != nil {
		v.playBtn.Update(dt)
	}
}

func (v *VideoPlaceholder) Draw() { v.drawInternal() }

func (v *VideoPlaceholder) drawInternal() {
	if v.IsHidden() {
		return
	}
	b := v.Bounds()
	style := v.GetStyle()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, style.BackgroundColor)
	}
	if v.poster != nil {
		v.poster.Draw()
	} else {
		rl.DrawRectangleRec(b, rl.NewColor(40, 42, 58, 255))
	}
	if v.playBtn != nil {
		v.playBtn.Draw()
	}
	badge := GetThemeStyle("form-value")
	badge.FontSize = 11
	drawTextS("Placeholder — no decoder", int32(b.X+8), int32(b.Y+6), badge)
}
