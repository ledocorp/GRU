// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ImageFit controls how a texture is scaled to fit the widget bounds.
type ImageFit int

const (
	// FitContain scales uniformly so the full image is visible (letter-boxed).
	FitContain ImageFit = iota
	// FitStretch stretches the texture to exactly fill the bounds.
	FitStretch
	// FitCover scales uniformly so the bounds are fully covered (cropped).
	FitCover
)

// Image is a non-interactive widget that displays a texture loaded from disk.
//
// The texture is loaded lazily on the first Draw call, or asynchronously via
// LoadAsync / LoadThumbnailAsync. If the file cannot be found or decoded a grey
// placeholder with a broken-image pattern is shown.
//
// Async loads capture an internal load generation; ResetTexture bumps it so stale
// worker callbacks cannot upload after path change or scene swap (see FilePicker.loadGen).
// Call Unload to release the GPU texture when the widget is no longer needed.
//
// # LLM Prompt Template
//
//	img := ui.NewImage("logo", "assets/logo.png", 0, 0, 320, 180)
//	img.FitMode = ui.FitContain
//	img.LoadThumbnailAsync(doc, 256) // gallery — avoids main-thread decode stall
//	panel.AddChild(img)
//
// Demo scenes: **Imaging (T2a)**, **Demo Directory** (app/tray icons), **Widget Showcase**.
type Image struct {
	Element
	FilePath string   // path to the image file (PNG, JPG, …)
	Tint     rl.Color // colour multiplied with the texture; rl.White = no tint
	FitMode  ImageFit // how to scale the texture inside Bounds

	texture  rl.Texture2D
	loaded   bool
	failed   bool
	loadGen  uint32 // bumped by ResetTexture; stale async uploads check this
}

// NewImage creates an Image widget.  The texture is loaded lazily on first Draw.
func NewImage(id, filePath string, x, y, w, h float32) *Image {
	img := &Image{
		Element:  NewElement(id, x, y, w, h),
		FilePath: filePath,
		Tint:     rl.White,
		FitMode:  FitContain,
	}
	return img
}

// Unload releases the GPU texture.  Call when the widget is no longer needed.
func (img *Image) Unload() {
	if img.loaded {
		rl.UnloadTexture(img.texture)
		img.loaded = false
	}
}

// ResetTexture clears any loaded texture so a new FilePath can be loaded.
// Bumps loadGen so in-flight LoadAsync / LoadThumbnailAsync callbacks are ignored.
func (img *Image) ResetTexture() {
	img.loadGen++
	img.Unload()
	img.failed = false
	img.MarkDirty()
}

// SetGPUTexture assigns an already-uploaded GPU texture and takes ownership.
// FilePath is cleared so lazy disk load is skipped.
func (img *Image) SetGPUTexture(tex rl.Texture2D) {
	img.Unload()
	img.texture = tex
	img.FilePath = ""
	img.loaded = tex.ID != 0
	img.failed = !img.loaded
	img.MarkDirty()
}

// LoadAsync decodes the image file on the worker pool and uploads the resulting
// texture to the GPU on the main thread via doc.QueueMain. Use this instead of
// the default lazy-load in Draw when you want to avoid a first-frame stall for
// large images.
//
// If the image is already loaded or has previously failed, LoadAsync is a
// no-op. The widget is marked dirty once the texture is ready, triggering an
// automatic redraw.
func (img *Image) LoadAsync(doc *Document) {
	if img.loaded || img.failed {
		return
	}
	path := img.FilePath
	gen := img.loadGen
	SubmitAsyncBg(func() {
		// Off-main: decode image file to CPU memory (no OpenGL calls).
		rimg := rl.LoadImage(path)
		if rimg.Width == 0 {
			doc.QueueMain(func() {
				if img.loadGen != gen {
					return
				}
				img.failed = true
				img.MarkDirty()
			})
			return
		}
		// On-main: upload CPU image to GPU texture, then free CPU copy.
		doc.QueueMain(func() {
			if img.loadGen != gen {
				rl.UnloadImage(rimg)
				return
			}
			img.texture = rl.LoadTextureFromImage(rimg)
			rl.UnloadImage(rimg)
			if img.texture.ID != 0 {
				img.loaded = true
			} else {
				img.failed = true
			}
			img.MarkDirty()
		})
	})
}

// LoadThumbnailAsync decodes path off the main thread, downsamples to maxEdge, and uploads PNG to the GPU.
// Use for gallery grids where full-resolution decode would stall the UI (T2a).
//
// Stale callbacks after ResetTexture or scene swap are ignored via loadGen.
// Very large files fall back to raylib/stb resize after the pure-Go path hits ErrImageTooLarge.
func (img *Image) LoadThumbnailAsync(doc *Document, maxEdge int) {
	if img.loaded || img.failed || img.FilePath == "" {
		return
	}
	if maxEdge < 1 {
		maxEdge = 256
	}
	path := img.FilePath
	gen := img.loadGen
	SubmitAsyncBg(func() {
		data, err := thumbnailPNGBytes(path, maxEdge)
		if err != nil {
			doc.QueueMain(func() {
				if img.loadGen != gen {
					return
				}
				img.failed = true
				img.MarkDirty()
			})
			return
		}
		doc.QueueMain(func() {
			if img.loadGen != gen {
				return
			}
			img.uploadPNG(data)
		})
	})
}

// SetPNGBytes replaces the texture with decoded PNG bytes. Call on the main thread only.
func (img *Image) SetPNGBytes(png []byte) {
	img.uploadImageBytes(".png", png)
}

// SetJPEGBytes replaces the texture with decoded JPEG bytes. Call on the main thread only.
func (img *Image) SetJPEGBytes(jpeg []byte) {
	img.uploadImageBytes(".jpg", jpeg)
}

func (img *Image) uploadPNG(png []byte) {
	img.uploadImageBytes(".png", png)
}

func (img *Image) uploadImageBytes(ext string, data []byte) {
	img.Unload()
	if len(data) == 0 {
		img.failed = true
		img.MarkDirty()
		return
	}
	rimg := rl.LoadImageFromMemory(ext, data, int32(len(data)))
	if rimg.Width == 0 {
		img.failed = true
		img.MarkDirty()
		return
	}
	img.texture = rl.LoadTextureFromImage(rimg)
	rl.UnloadImage(rimg)
	if img.texture.ID != 0 {
		img.loaded = true
		img.failed = false
	} else {
		img.failed = true
	}
	img.MarkDirty()
}

// Update is a no-op — Image is not interactive.
func (img *Image) Update(_ float32) {}

// Layout is a no-op for leaf widgets.
func (img *Image) Layout() { img.layoutDirty = false }

// Draw implements Node.Draw.
func (img *Image) Draw() { img.drawInternal() }

func (img *Image) drawInternal() {
	if img.IsHidden() {
		return
	}

	// Lazy-load texture on first frame.
	if !img.loaded && !img.failed {
		img.tryLoad()
	}

	b := img.Bounds()
	if b.Width <= 0 || b.Height <= 0 {
		return
	}
	if img.failed || !img.loaded {
		img.drawPlaceholder(b)
		return
	}

	// Clear the widget box first so FitContain letterboxing never leaves stale
	// pixels during resize.
	rl.DrawRectangleRec(b, rl.NewColor(255, 255, 255, 255))
	src, dst := img.fitRects(b, float32(img.texture.Width), float32(img.texture.Height))
	rl.DrawTexturePro(img.texture, src, dst, rl.NewVector2(0, 0), 0, img.Tint)
}

func (img *Image) tryLoad() {
	if _, err := os.Stat(img.FilePath); err != nil {
		img.failed = true
		return
	}
	img.texture = rl.LoadTexture(img.FilePath)
	if img.texture.ID != 0 {
		img.loaded = true
	} else {
		rl.UnloadTexture(img.texture)
		img.failed = true
	}
}

func (img *Image) fitRects(b rl.Rectangle, tw, th float32) (rl.Rectangle, rl.Rectangle) {
	src := rl.NewRectangle(0, 0, tw, th)
	switch img.FitMode {
	case FitStretch:
		return src, b
	case FitContain:
		scale := b.Width / tw
		if sy := b.Height / th; sy < scale {
			scale = sy
		}
		dw, dh := tw*scale, th*scale
		return src, rl.NewRectangle(b.X+(b.Width-dw)/2, b.Y+(b.Height-dh)/2, dw, dh)
	case FitCover:
		scale := b.Width / tw
		if sy := b.Height / th; sy > scale {
			scale = sy
		}
		sw := b.Width / scale
		sh := b.Height / scale
		if sw > tw {
			sw = tw
		}
		if sh > th {
			sh = th
		}
		src = rl.NewRectangle((tw-sw)/2, (th-sh)/2, sw, sh)
		return src, b
	}
	return src, b
}

func (img *Image) drawPlaceholder(b rl.Rectangle) {
	fill := rl.NewColor(218, 220, 228, 255)
	border := rl.NewColor(180, 182, 195, 255)
	iconCol := rl.NewColor(148, 150, 168, 255)

	rl.DrawRectangleRec(b, fill)
	rl.DrawRectangleLinesEx(b, 1, border)

	iconSize := b.Width * 0.22
	if iconSize > 48 {
		iconSize = 48
	}
	if iconSize < 16 {
		iconSize = 16
	}
	showMsg := b.Height > 56
	cx := b.X + b.Width/2
	cy := b.Y + b.Height/2
	if showMsg {
		cy -= iconSize * 0.35
	}
	iconRect := rl.NewRectangle(cx-iconSize/2, cy-iconSize/2, iconSize, iconSize)
	if !drawRemixIcon(iconRect, RemixImageLine, iconCol, 1) {
		Phosphor.Draw(iconRect, PhosphorImage, PhosphorRegular, iconCol)
	}

	if showMsg {
		msg := "image not found"
		if img.FilePath == "" {
			msg = "no image"
		}
		const msgSize = int32(12)
		tw := measureText(msg, msgSize)
		drawText(msg, int32(b.X)+(int32(b.Width)-tw)/2, int32(cy+iconSize*0.65), msgSize, iconCol)
		return
	}
	if iconSize < 20 {
		const fs = int32(12)
		msg := "no image"
		tw := measureText(msg, fs)
		drawText(msg, int32(b.X)+(int32(b.Width)-tw)/2, int32(b.Y+b.Height/2)-6, fs, iconCol)
	}
}

// IsInteractive implements Node.IsInteractive.
func (img *Image) IsInteractive() bool { return false }
