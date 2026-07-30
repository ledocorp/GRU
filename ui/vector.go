// Package ui (continued)
// See node.go for the full package documentation.
package ui

// vector.go provides a thin integration layer between fogleman/gg (a pure-Go
// vector graphics library) and raylib textures. Use it to pre-render icons or
// any 2-D vector shapes into GPU textures that widgets can display.
//
// Typical workflow (synchronous — blocks caller until GPU upload):
//
//	gc := ui.NewIconContext(32)
//	ui.DrawIconPlus(gc)               // paint the shape
//	tex := ui.ContextToTexture(gc)    // upload to GPU (must be on main thread)
//	// … use tex in a widget …
//	ui.FreeIconTexture(tex)           // release when no longer needed
//
// Async workflow (heavy shapes — no main-thread stall):
//
//	gc := ui.NewVectorContext(128, 128)
//	ui.DrawRoundedRectWithShadow(gc, ...)
//	ui.AsyncContextToTexture(doc, gc, func(tex rl.Texture2D) {
//	    myWidget.SetVectorTex(tex)
//	    myWidget.MarkDirty()
//	})

import (
	"bytes"
	"image"
	"image/color"

	"github.com/fogleman/gg"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// NewIconContext returns a square gg drawing context pre-cleared to full
// transparency. size is both the width and height in pixels.
func NewIconContext(size int) *gg.Context {
	gc := gg.NewContext(size, size)
	gc.Clear() // sets all pixels to transparent black
	return gc
}

// GoImageToTexture uploads an RGBA image to the GPU (main thread only).
// Uses GenImageColor + ImageDrawPixel so UnloadImage never frees Go-owned memory.
func GoImageToTexture(img image.Image) rl.Texture2D {
	if img == nil {
		return rl.Texture2D{}
	}
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return rl.Texture2D{}
	}
	rlImg := rl.NewImageFromImage(img)
	if rlImg == nil || rlImg.Data == nil {
		return rl.Texture2D{}
	}
	tex := rl.LoadTextureFromImage(rlImg)
	rl.UnloadImage(rlImg)
	if tex.ID != 0 {
		rl.SetTextureFilter(tex, rl.FilterBilinear)
	}
	return tex
}

// ContextToTexture uploads a gg context to the GPU as a raylib Texture2D.
// The returned texture must be freed with FreeIconTexture when no longer needed.
func ContextToTexture(gc *gg.Context) rl.Texture2D {
	if gc == nil {
		return rl.Texture2D{}
	}
	return GoImageToTexture(gc.Image())
}

// FreeIconTexture unloads a GPU texture created by ContextToTexture.
func FreeIconTexture(t rl.Texture2D) {
	if t.ID != 0 {
		rl.UnloadTexture(t)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Pre-made icon painters
// Each function draws a centred icon on the supplied gg context using the
// context's current dimensions. The context is cleared first so callers can
// call the painter directly after NewIconContext without extra setup.
// ──────────────────────────────────────────────────────────────────────────────

// DrawIconPlus paints a + (add) icon centred in gc using white with full alpha.
func DrawIconPlus(gc *gg.Context) {
	gc.Clear()
	w, h := float64(gc.Width()), float64(gc.Height())
	pad := w * 0.25
	gc.SetRGBA(1, 1, 1, 1)
	gc.SetLineWidth(w * 0.12)
	gc.SetLineCap(gg.LineCapRound)
	// Horizontal bar
	gc.DrawLine(pad, h/2, w-pad, h/2)
	gc.Stroke()
	// Vertical bar
	gc.DrawLine(w/2, pad, w/2, h-pad)
	gc.Stroke()
}

// DrawIconCheck paints a ✓ (checkmark) icon centred in gc using white with full alpha.
func DrawIconCheck(gc *gg.Context) {
	gc.Clear()
	w, h := float64(gc.Width()), float64(gc.Height())
	gc.SetRGBA(1, 1, 1, 1)
	gc.SetLineWidth(w * 0.12)
	gc.SetLineCap(gg.LineCapRound)
	gc.SetLineJoin(gg.LineJoinRound)
	gc.MoveTo(w*0.18, h*0.52)
	gc.LineTo(w*0.40, h*0.74)
	gc.LineTo(w*0.82, h*0.26)
	gc.Stroke()
}

// DrawIconArrowRight paints a → (right arrow) icon centred in gc using white with full alpha.
func DrawIconArrowRight(gc *gg.Context) {
	gc.Clear()
	w, h := float64(gc.Width()), float64(gc.Height())
	gc.SetRGBA(1, 1, 1, 1)
	gc.SetLineWidth(w * 0.12)
	gc.SetLineCap(gg.LineCapRound)
	gc.SetLineJoin(gg.LineJoinRound)
	// Shaft
	gc.MoveTo(w*0.18, h*0.50)
	gc.LineTo(w*0.72, h*0.50)
	gc.Stroke()
	// Arrowhead
	gc.MoveTo(w*0.55, h*0.28)
	gc.LineTo(w*0.80, h*0.50)
	gc.LineTo(w*0.55, h*0.72)
	gc.Stroke()
}

// DrawIconX paints a × (close) icon centred in gc using white with full alpha.
func DrawIconX(gc *gg.Context) {
	gc.Clear()
	w, h := float64(gc.Width()), float64(gc.Height())
	pad := w * 0.22
	gc.SetRGBA(1, 1, 1, 1)
	gc.SetLineWidth(w * 0.12)
	gc.SetLineCap(gg.LineCapRound)
	gc.DrawLine(pad, pad, w-pad, h-pad)
	gc.Stroke()
	gc.DrawLine(w-pad, pad, pad, h-pad)
	gc.Stroke()
}

// ─────────────────────────────────────────────────────────────────────────────
// General-purpose context constructors
// ─────────────────────────────────────────────────────────────────────────────

// NewVectorContext returns a gg drawing context of arbitrary dimensions
// pre-cleared to full transparency. Use this instead of NewIconContext when
// you need a non-square canvas (e.g. a banner or a card background).
func NewVectorContext(w, h int) *gg.Context {
	gc := gg.NewContext(w, h)
	gc.Clear()
	return gc
}

// ─────────────────────────────────────────────────────────────────────────────
// Shape helpers
// ─────────────────────────────────────────────────────────────────────────────

// DrawRoundedRectWithShadow draws a rounded rectangle with a soft drop-shadow.
// c is the fill colour; radius is the corner radius in pixels; blur is the
// number of shadow expansion layers (4–8 is a good range).
//
// Because raylib exposes no native shadow primitive this is implemented as
// a series of progressively lighter, larger rectangles drawn before the fill —
// the classic "fake blur" used by gg-based renderers.
//
// MUST be called off the main thread or on the main thread before any draw
// pass. Never call inside BeginDrawing / EndDrawing.
func DrawRoundedRectWithShadow(gc *gg.Context, x, y, w, h, radius float64, c color.RGBA, blur int) {
	// Shadow layers: expand outward, fade from 30% → 0% alpha
	for i := blur; i >= 1; i-- {
		ex := float64(i) * 1.5
		alpha := 0.30 * float64(i) / float64(blur)
		gc.SetRGBA(0, 0, 0, alpha)
		gc.DrawRoundedRectangle(x-ex, y-ex+float64(i), w+ex*2, h+ex*2, radius+ex)
		gc.Fill()
	}
	// Fill
	gc.SetRGBA(float64(c.R)/255, float64(c.G)/255, float64(c.B)/255, float64(c.A)/255)
	gc.DrawRoundedRectangle(x, y, w, h, radius)
	gc.Fill()
}

// DrawGradientRect fills a rectangle with a vertical linear gradient from
// top colour ct to bottom colour cb.
func DrawGradientRect(gc *gg.Context, x, y, w, h float64, ct, cb color.RGBA) {
	steps := int(h)
	if steps < 1 {
		steps = 1
	}
	stepH := h / float64(steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		r := lerp01(float64(ct.R), float64(cb.R), t)
		g := lerp01(float64(ct.G), float64(cb.G), t)
		b := lerp01(float64(ct.B), float64(cb.B), t)
		a := lerp01(float64(ct.A), float64(cb.A), t)
		gc.SetRGBA(r/255, g/255, b/255, a/255)
		gc.DrawRectangle(x, y+float64(i)*stepH, w, stepH+1)
		gc.Fill()
	}
}

func lerp01(a, b, t float64) float64 { return a + (b-a)*t }

// ─────────────────────────────────────────────────────────────────────────────
// Async GPU upload
// ─────────────────────────────────────────────────────────────────────────────

// AsyncContextToTexture encodes gc as PNG on the worker pool (off main thread)
// then uploads the resulting texture on the main thread via doc.QueueMain,
// calling onDone(tex) when the texture is ready.
//
// This is the recommended path for large or complex gg renders. For small icons
// (≤ 64×64) the synchronous ContextToTexture is fast enough.
//
// onDone is called on the main goroutine — it is safe to call MarkDirty(),
// set Signal values, or directly mutate widget fields inside it.
func AsyncContextToTexture(doc *Document, gc *gg.Context, onDone func(rl.Texture2D)) {
	// Capture the pixel buffer synchronously so gc can be reused by the caller
	// after this call returns (the encode runs in a worker goroutine).
	w, h := gc.Width(), gc.Height()
	img := gc.Image() // returns *image.RGBA — safe to pass to a goroutine
	_ = w
	_ = h

	SubmitAsyncBg(func() {
		// Off-main: encode to PNG in memory (pure CPU, no OpenGL).
		var buf bytes.Buffer
		if err := gg.NewContextForImage(img).EncodePNG(&buf); err != nil {
			doc.QueueMain(func() { onDone(rl.Texture2D{}) })
			return
		}
		b := buf.Bytes()
		bLen := int32(len(b))

		// On-main: decode PNG → CPU image → GPU texture.
		doc.QueueMain(func() {
			rimg := rl.LoadImageFromMemory(".png", b, bLen)
			tex := rl.LoadTextureFromImage(rimg)
			rl.SetTextureFilter(tex, rl.FilterBilinear)
			rl.UnloadImage(rimg)
			onDone(tex)
		})
	})
}
