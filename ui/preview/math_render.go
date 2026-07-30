package preview

import (
	"bytes"
	"fmt"
	"strings"

	"codeberg.org/go-latex/latex/drawtex/drawimg"
	"codeberg.org/go-latex/latex/mtex"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	mathRenderDPI       = 96 // sharper glyphs; displayed at ~125% of prior compact size
	mathInlineFontSize  = 12
	mathDisplayFontSize = 17
)

// renderLatexPNG renders a math expression to PNG bytes (safe on worker goroutines).
func renderLatexPNG(source string, display bool) (png []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("latex render: %v", r)
		}
	}()
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, fmt.Errorf("empty math source")
	}
	expr := wrapMathExpression(source)
	size := float64(mathInlineFontSize)
	if display {
		size = float64(mathDisplayFontSize)
	}
	var buf bytes.Buffer
	dst := drawimg.NewRenderer(&buf)
	if err := mtex.Render(dst, expr, size, mathRenderDPI, nil); err != nil {
		return nil, err
	}
	if buf.Len() == 0 {
		return nil, fmt.Errorf("empty PNG output")
	}
	return buf.Bytes(), nil
}

func wrapMathExpression(source string) string {
	source = strings.TrimSpace(source)
	if strings.HasPrefix(source, "$") && strings.HasSuffix(source, "$") {
		return source
	}
	return "$" + source + "$"
}

func pngBytesToTexture(png []byte) (rl.Texture2D, int32, int32) {
	if len(png) == 0 {
		return rl.Texture2D{}, 0, 0
	}
	img := rl.LoadImageFromMemory(".png", png, int32(len(png)))
	if img.Width <= 0 || img.Height <= 0 {
		return rl.Texture2D{}, 0, 0
	}
	tex := rl.LoadTextureFromImage(img)
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	w, h := img.Width, img.Height
	rl.UnloadImage(img)
	return tex, w, h
}
