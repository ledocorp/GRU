package ui

import (
	"image"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// T1.8 — dynamic shelf-packed glyph atlas (Skia strike-cache pattern).

const (
	shapedAtlasDim  = 2048
	shapedAtlasCell = 2 // gutter between packed glyphs
)

type shapedGlyphAtlas struct {
	tex              rl.Texture2D
	pixels           []color.RGBA
	cursorX, cursorY int
	shelfH           int
	ready            bool
}

var shapedAtlas shapedGlyphAtlas

func initShapedGlyphAtlas() {
	shapedAtlas = shapedGlyphAtlas{
		pixels: make([]color.RGBA, shapedAtlasDim*shapedAtlasDim),
	}
}

func unloadShapedGlyphAtlas() {
	if shapedAtlas.tex.ID != 0 {
		rl.UnloadTexture(shapedAtlas.tex)
	}
	shapedAtlas = shapedGlyphAtlas{
		pixels: make([]color.RGBA, shapedAtlasDim*shapedAtlasDim),
	}
}

func shapedAtlasEnsureTexture() bool {
	if shapedAtlas.ready && shapedAtlas.tex.ID != 0 {
		return true
	}
	img := rl.GenImageColor(shapedAtlasDim, shapedAtlasDim, rl.NewColor(0, 0, 0, 0))
	if img == nil || img.Data == nil {
		return false
	}
	tex := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	if tex.ID == 0 {
		return false
	}
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	shapedAtlas.tex = tex
	shapedAtlas.ready = true
	return true
}

func shapedAtlasAlloc(w, h int) (image.Rectangle, bool) {
	if w <= 0 || h <= 0 || w > shapedAtlasDim || h > shapedAtlasDim {
		return image.Rectangle{}, false
	}
	pad := shapedAtlasCell
	if shapedAtlas.cursorX+pad+w+pad > shapedAtlasDim {
		shapedAtlas.cursorX = 0
		shapedAtlas.cursorY += shapedAtlas.shelfH + pad
		shapedAtlas.shelfH = 0
	}
	if shapedAtlas.cursorY+pad+h+pad > shapedAtlasDim {
		return image.Rectangle{}, false
	}
	x0 := shapedAtlas.cursorX + pad
	y0 := shapedAtlas.cursorY + pad
	rect := image.Rect(x0, y0, x0+w, y0+h)
	shapedAtlas.cursorX = x0 + w + pad
	if h > shapedAtlas.shelfH {
		shapedAtlas.shelfH = h
	}
	return rect, true
}

func shapedAtlasUpload(img image.Image, slot image.Rectangle) bool {
	if !shapedAtlasEnsureTexture() {
		return false
	}
	b := img.Bounds()
	if b.Dx() != slot.Dx() || b.Dy() != slot.Dy() {
		return false
	}
	region := make([]color.RGBA, slot.Dx()*slot.Dy())
	i := 0
	for y := slot.Min.Y; y < slot.Max.Y; y++ {
		for x := slot.Min.X; x < slot.Max.X; x++ {
			srcX := b.Min.X + (x - slot.Min.X)
			srcY := b.Min.Y + (y - slot.Min.Y)
			c := color.RGBAModel.Convert(img.At(srcX, srcY)).(color.RGBA)
			region[i] = c
			shapedAtlas.pixels[y*shapedAtlasDim+x] = c
			i++
		}
	}
	rec := rl.NewRectangle(float32(slot.Min.X), float32(slot.Min.Y), float32(slot.Dx()), float32(slot.Dy()))
	rl.UpdateTextureRec(shapedAtlas.tex, rec, region)
	return true
}

func shapedAtlasPackGlyph(img image.Image) (rl.Texture2D, rl.Rectangle, bool) {
	if img == nil {
		return rl.Texture2D{}, rl.Rectangle{}, false
	}
	b := img.Bounds()
	slot, ok := shapedAtlasAlloc(b.Dx(), b.Dy())
	if !ok {
		return rl.Texture2D{}, rl.Rectangle{}, false
	}
	if !shapedAtlasUpload(img, slot) {
		return rl.Texture2D{}, rl.Rectangle{}, false
	}
	src := rl.NewRectangle(float32(slot.Min.X), float32(slot.Min.Y), float32(slot.Dx()), float32(slot.Dy()))
	return shapedAtlas.tex, src, true
}
