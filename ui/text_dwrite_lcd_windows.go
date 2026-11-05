//go:build windows && cgo

package ui

/*
#cgo LDFLAGS: -ldwrite -lole32 -luuid
#include "dwrite_lcd.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"image"
	"sync"
	"unicode/utf16"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	dwriteLCDOnce        sync.Once
	dwriteLCDOK          bool
	dwriteLCDWarn        sync.Once
	dwriteLCDPathUTF     []uint16
	dwriteBlendGamma     float32 = 1.8
	dwriteBlendContrast  float32 = 0.5
	dwriteClearTypeLevel float32 = 0.5
)

func dwriteLCDAvailable() bool {
	dwriteLCDOnce.Do(func() {
		dwriteLCDOK = C.gru_dwrite_lcd_available() != 0
		if dwriteLCDOK {
			fmt.Println("Gru: T1.14b DirectWrite ClearType LCD scaler available")
		}
	})
	return dwriteLCDOK
}

func dwriteLCDFontPathUTF16() []uint16 {
	return dwriteLCDFontPathUTF16For("")
}

func dwriteLCDFontPathUTF16For(path string) []uint16 {
	if path == "" {
		if len(dwriteLCDPathUTF) > 0 {
			return dwriteLCDPathUTF
		}
		path = LoadedUIFontPath
		if path == "" {
			path = firstExisting(resolvedFontCandidates()...)
		}
		if path == "" {
			return nil
		}
		dwriteLCDPathUTF = utf16.Encode([]rune(path + "\x00"))
		return dwriteLCDPathUTF
	}
	return utf16.Encode([]rune(path + "\x00"))
}

// packDWriteLCDStrikeAtlas builds ClearType RGB atlas at exact size + subpixel originX.
// fontPath empty → chrome LoadedUIFontPath (Segoe).
// symmetric → CLEARTYPE_NATURAL_SYMMETRIC (aux display sizes toward browser softness).
func packDWriteLCDStrikeAtlas(pixelSize int32, originX float32, fontPath string, symmetric bool) (*image.RGBA, map[rune]ftGreyGlyph, bool) {
	if !dwriteLCDAvailable() || pixelSize < 11 {
		return nil, nil, false
	}
	pathU16 := dwriteLCDFontPathUTF16For(fontPath)
	if len(pathU16) < 2 {
		return nil, nil, false
	}
	// WS: always try 1024² first (~4MB). Auto-2048 at px≥20 was ~16MB × every
	// sticky size and pushed Studio WS from ~90MB into the 150–170MB range.
	pathLen := C.int(len(pathU16) - 1)
	cpath := (*C.uint16_t)(unsafe.Pointer(&pathU16[0]))
	mode := C.int(0)
	if symmetric {
		mode = 1
	}
	pack := func(dim int) (*image.RGBA, map[rune]ftGreyGlyph, int, bool) {
		glyphs := make(map[rune]ftGreyGlyph, 256)
		atlas := image.NewRGBA(image.Rect(0, 0, dim, dim))
		pad := 2
		x, y, rowH := pad, pad, 0
		ink := 0
		truncated := false
		for _, r := range sdfCodepoints() {
			var outRGB *C.uint8_t
			var ow, oh, bx, by C.int
			var adv C.float
			rc := C.gru_dwrite_lcd_glyph(
				cpath, pathLen, C.float(pixelSize), C.uint32_t(r), C.float(originX), mode,
				&outRGB, &ow, &oh, &bx, &by, &adv,
			)
			advPx := float32(adv)
			if rc < 0 {
				glyphs[r] = ftGreyGlyph{advance: advPx}
				continue
			}
			if rc == 0 || outRGB == nil || ow <= 0 || oh <= 0 {
				glyphs[r] = ftGreyGlyph{advance: advPx}
				if outRGB != nil {
					C.gru_dwrite_free(unsafe.Pointer(outRGB))
				}
				continue
			}
			gw, gh := int(ow), int(oh)
			if x+gw+pad > dim {
				x = pad
				y += rowH + pad
				rowH = 0
			}
			if y+gh+pad > dim {
				C.gru_dwrite_free(unsafe.Pointer(outRGB))
				truncated = true
				break
			}
			n := blitDWriteLCDRGB(atlas, x, y, gw, gh, outRGB)
			C.gru_dwrite_free(unsafe.Pointer(outRGB))
			if n == 0 {
				glyphs[r] = ftGreyGlyph{advance: advPx}
				continue
			}
			glyphs[r] = ftGreyGlyph{
				src:      rl.NewRectangle(float32(x), float32(y), float32(gw), float32(gh)),
				bearingX: float32(bx),
				bearingY: float32(by),
				advance:  advPx,
				hasInk:   true,
			}
			ink += n
			x += gw + pad
			if gh > rowH {
				rowH = gh
			}
		}
		return atlas, glyphs, ink, truncated
	}
	atlas, glyphs, ink, truncated := pack(1024)
	if truncated {
		atlas, glyphs, ink, truncated = pack(2048)
		if truncated {
			dwriteLCDWarn.Do(func() {
				fmt.Printf("Gru: DWrite LCD atlas full at size=%d — truncating\n", pixelSize)
			})
		}
	}
	if ink < 32 || len(glyphs) < 32 {
		return nil, nil, false
	}
	return atlas, glyphs, true
}

// packDWriteGreyStrikeAtlas builds greyscale (A8→RGBA) sticky strikes via
// DWRITE_RENDERING_MODE_NATURAL — preferred for Inter/Fira aux faces.
func packDWriteGreyStrikeAtlas(pixelSize int32, fontPath string) (*image.RGBA, map[rune]ftGreyGlyph, bool) {
	if !dwriteLCDAvailable() || pixelSize < 11 {
		return nil, nil, false
	}
	pathU16 := dwriteLCDFontPathUTF16For(fontPath)
	if len(pathU16) < 2 {
		return nil, nil, false
	}
	pathLen := C.int(len(pathU16) - 1)
	cpath := (*C.uint16_t)(unsafe.Pointer(&pathU16[0]))
	packGrey := func(dim int) (*image.RGBA, map[rune]ftGreyGlyph, int, bool) {
		glyphs := make(map[rune]ftGreyGlyph, 256)
		atlas := image.NewRGBA(image.Rect(0, 0, dim, dim))
		pad := 4
		x, y, rowH := pad, pad, 0
		ink := 0
		truncated := false
		for _, r := range sdfCodepoints() {
			var outA *C.uint8_t
			var ow, oh, bx, by C.int
			var adv C.float
			rc := C.gru_dwrite_grey_glyph(
				cpath, pathLen, C.float(pixelSize), C.uint32_t(r),
				&outA, &ow, &oh, &bx, &by, &adv,
			)
			advPx := float32(adv)
			if rc < 0 {
				glyphs[r] = ftGreyGlyph{advance: advPx}
				continue
			}
			if rc == 0 || outA == nil || ow <= 0 || oh <= 0 {
				glyphs[r] = ftGreyGlyph{advance: advPx}
				if outA != nil {
					C.gru_dwrite_free(unsafe.Pointer(outA))
				}
				continue
			}
			gw, gh := int(ow), int(oh)
			if x+gw+pad > dim {
				x = pad
				y += rowH + pad
				rowH = 0
			}
			if y+gh+pad > dim {
				C.gru_dwrite_free(unsafe.Pointer(outA))
				truncated = true
				break
			}
			n := blitDWriteGreyA8(atlas, x, y, gw, gh, outA)
			C.gru_dwrite_free(unsafe.Pointer(outA))
			if n == 0 {
				glyphs[r] = ftGreyGlyph{advance: advPx}
				continue
			}
			glyphs[r] = ftGreyGlyph{
				src:      rl.NewRectangle(float32(x), float32(y), float32(gw), float32(gh)),
				bearingX: float32(bx),
				bearingY: float32(by),
				advance:  advPx,
				hasInk:   true,
			}
			ink += n
			x += gw + pad
			if gh > rowH {
				rowH = gh
			}
		}
		return atlas, glyphs, ink, truncated
	}
	atlas, glyphs, ink, truncated := packGrey(1024)
	if truncated {
		atlas, glyphs, ink, truncated = packGrey(2048)
		if truncated {
			dwriteLCDWarn.Do(func() {
				fmt.Printf("Gru: DWrite grey atlas full at size=%d — truncating\n", pixelSize)
			})
		}
	}
	if truncated || ink < 32 || len(glyphs) < 32 {
		return nil, nil, false
	}
	if !auxAtlasHasASCII(glyphs) {
		return nil, nil, false
	}
	return atlas, glyphs, true
}

func blitDWriteGreyA8(atlas *image.RGBA, dstX, dstY, gw, gh int, a8 *C.uint8_t) int {
	ink := 0
	src := unsafe.Slice((*byte)(unsafe.Pointer(a8)), gw*gh)
	for py := 0; py < gh; py++ {
		for px := 0; px < gw; px++ {
			a := src[py*gw+px]
			if a == 0 {
				continue
			}
			o := atlas.PixOffset(dstX+px, dstY+py)
			atlas.Pix[o+0] = 255
			atlas.Pix[o+1] = 255
			atlas.Pix[o+2] = 255
			atlas.Pix[o+3] = a
			ink++
		}
	}
	return ink
}

func blitDWriteLCDRGB(atlas *image.RGBA, dstX, dstY, gw, gh int, rgb *C.uint8_t) int {
	ink := 0
	src := unsafe.Slice((*byte)(unsafe.Pointer(rgb)), gw*gh*3)
	for py := 0; py < gh; py++ {
		for px := 0; px < gw; px++ {
			i := (py*gw + px) * 3
			cr, cg, cb := src[i], src[i+1], src[i+2]
			o := atlas.PixOffset(dstX+px, dstY+py)
			atlas.Pix[o+0] = cr
			atlas.Pix[o+1] = cg
			atlas.Pix[o+2] = cb
			atlas.Pix[o+3] = 255
			if cr > 2 || cg > 2 || cb > 2 {
				ink++
			}
		}
	}
	return ink
}
