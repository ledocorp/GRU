// Package ui (continued)
package ui

import (
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// FileDropZone accepts OS file drops when the cursor is over the zone (Strategy 2 #27).
//
// # LLM Prompt Template
//
//	zone := ui.NewFileDropZone("import", []string{".png", ".jpg"}, func(paths []string) {
//	    importFiles(paths)
//	}, 0, 0, 320, 120)
//	panel.AddChild(zone)
//
// Demo scenes: **Batch 10** (FileDrop panel).
type FileDropZone struct {
	Element
	Extensions []string // e.g. ".png"; empty = any file
	OnFiles    func(paths []string)
	Hint       string

	hover bool
}

// NewFileDropZone creates a drop target.
func NewFileDropZone(id string, extensions []string, onFiles func(paths []string), x, y, w, h float32) *FileDropZone {
	z := &FileDropZone{
		Element:    NewElement(id, x, y, w, h),
		Extensions: extensions,
		OnFiles:    onFiles,
		Hint:       "Drop files here",
	}
	z.styleName = "filedropzone"
	return z
}

// IsInteractive implements Node.
func (z *FileDropZone) IsInteractive() bool { return true }

func (z *FileDropZone) accepts(path string) bool {
	if len(z.Extensions) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range z.Extensions {
		if strings.ToLower(e) == ext {
			return true
		}
	}
	return false
}

func (z *FileDropZone) filter(paths []string) []string {
	if len(z.Extensions) == 0 {
		return paths
	}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if z.accepts(p) {
			out = append(out, p)
		}
	}
	return out
}

func (z *FileDropZone) Update(_ float32) {
	if z.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	was := z.hover
	z.hover = rl.CheckCollisionPointRec(mouse, z.Bounds())
	if z.hover != was {
		z.MarkDrawDirty()
	}

	if rl.IsFileDropped() && z.hover {
		paths := rl.LoadDroppedFiles()
		defer rl.UnloadDroppedFiles()
		if len(paths) > 0 && z.OnFiles != nil {
			filtered := z.filter(paths)
			if len(filtered) > 0 {
				z.OnFiles(filtered)
				z.MarkDrawDirty()
			}
		}
	}
}

func (z *FileDropZone) Layout() { z.layoutDirty = false }

func (z *FileDropZone) Draw() { z.drawInternal() }

func (z *FileDropZone) drawInternal() {
	if z.IsHidden() {
		return
	}
	b := z.Bounds()
	style := z.GetStyle()
	bg := style.BackgroundColor
	if bg.A == 0 {
		bg = rl.NewColor(248, 249, 252, 255)
	}
	border := style.BorderColor
	if border.A == 0 {
		border = rl.NewColor(200, 204, 220, 255)
	}
	if z.hover {
		bg = rl.NewColor(232, 234, 255, 255)
		border = rl.NewColor(79, 70, 229, 255)
	}
	rl.DrawRectangleRounded(b, 0.06, 8, bg)
	rl.DrawRectangleRoundedLinesEx(b, 0.06, 8, 2, border)

	hint := z.Hint
	if hint == "" {
		hint = "Drop files here"
	}
	if len(z.Extensions) > 0 {
		hint += " (" + strings.Join(z.Extensions, ", ") + ")"
	}
	ts := style
	if ts.FontSize == 0 {
		ts.FontSize = 14
	}
	iw := measureTextS(hint, ts)
	drawTextS(hint, int32(b.X+(b.Width-float32(iw))/2), TextPosY(b, ts), ts)
}
