package ui

import (
	"fmt"
	"image"
	"math"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/image/font"
	xfixed "golang.org/x/image/math/fixed"
)

// T1.12 — MSDF Latin chrome (opt-in).
//
// Slice A (this file): median shader + true-SDF packed into RGB (R=G=B).
// That wires the dual-path and scales cleaner than soft raylib SDF; full
// Chlumsky edge-colored MSDF (distinct R/G/B) is the follow-on bake.
//
// Opt-in: GRU_TEXT_MSDF=1. Takes priority over ft-grey. Same Face advances as
// greyscale for RichText wrap safety. Footprint: ~32–48×DPI cells in 1024².

type msdfGlyph struct {
	src      rl.Rectangle
	bearingX float32
	bearingY float32
	advance  float32
	hasInk   bool
}

var (
	msdfReady    bool
	msdfTex      rl.Texture2D
	msdfBasePx   float32
	msdfAscentPx float32
	msdfGlyphs   map[rune]msdfGlyph
	msdfShader   rl.Shader
	msdfWarn     sync.Once
)

const (
	msdfPad     = 4
	msdfPxRange = 4.0
	// Optical match vs soft SDF / ft-grey.
	msdfOpticalScale float32 = 0.93
)

// PreferMSDFChrome is true when GRU_TEXT_MSDF=1 (opt-in T1.12).
func PreferMSDFChrome() bool {
	v := strings.ToLower(strings.TrimSpace(EnvOr("GRU_TEXT_MSDF", "GORY_TEXT_MSDF")))
	return v == "1" || v == "on" || v == "true" || v == "yes" || v == "msdf"
}

func msdfActive(mono, preview bool) bool {
	return PreferMSDFChrome() && msdfReady && !mono && !preview
}

func msdfScale(fontSize float32) float32 {
	if msdfBasePx <= 0 {
		return 1
	}
	return (fontSize / msdfBasePx) * msdfOpticalScale
}

func msdfFragSource() string {
	// Premultiply-style: RGB from tint only; coverage from median distance.
	// Outside (dist < 0.5) must be fully transparent — inverted fields paint black boxes.
	return `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
uniform sampler2D texture0;
uniform vec4 colDiffuse;
out vec4 finalColor;
float median(float r, float g, float b) {
    return max(min(r, g), min(max(r, g), b));
}
void main() {
    vec3 s = texture(texture0, fragTexCoord).rgb;
    float dist = median(s.r, s.g, s.b);
    float w = fwidth(dist) * 0.7;
    float alpha = smoothstep(0.5 - w, 0.5 + w, dist);
    vec3 rgb = fragColor.rgb * colDiffuse.rgb;
    finalColor = vec4(rgb, fragColor.a * colDiffuse.a * alpha);
}`
}

// InitMSDFFont builds an MSDF/SDF-RGB atlas at pixelSize (typically EffectiveFTGreyAtlasSize).
func InitMSDFFont(pixelSize int32) bool {
	unloadMSDFFont()
	if pixelSize < 16 {
		pixelSize = 16
	}
	path := LoadedUIFontPath
	if path == "" {
		path = firstExisting(resolvedFontCandidates()...)
	}
	if path == "" {
		return false
	}
	data, err := ReadAssetFile(path)
	if err != nil {
		return false
	}
	ttf, err := truetype.Parse(data)
	if err != nil {
		return false
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              float64(pixelSize),
		DPI:               72,
		Hinting:           font.HintingFull,
		GlyphCacheEntries: 8,
	})
	ascent := float32(face.Metrics().Ascent.Ceil())
	atlas, glyphs, ink, ok := packMSDFAtlas(face, pixelSize)
	_ = face.Close()
	if !ok {
		msdfWarn.Do(func() {
			fmt.Printf("Gru: msdf atlas failed (ink=%d) — staying on ft-grey/SDF\n", ink)
		})
		return false
	}
	shader := rl.LoadShaderFromMemory("", msdfFragSource())
	if shader.ID == 0 {
		return false
	}
	tex := GoImageToTexture(atlas)
	if tex.ID == 0 {
		rl.UnloadShader(shader)
		return false
	}
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	msdfShader = shader
	msdfGlyphs = glyphs
	msdfTex = tex
	msdfBasePx = float32(pixelSize)
	msdfAscentPx = ascent
	msdfReady = true
	return true
}

func packMSDFAtlas(face font.Face, pixelSize int32) (*image.RGBA, map[rune]msdfGlyph, int, bool) {
	glyphs := make(map[rune]msdfGlyph, 256)
	atlasW, atlasH := 1024, 1024
	if pixelSize >= 64 {
		atlasW, atlasH = 2048, 2048
	}
	atlas := image.NewRGBA(image.Rect(0, 0, atlasW, atlasH))
	// Neutral mid distance outside glyphs.
	for i := 0; i < len(atlas.Pix); i += 4 {
		atlas.Pix[i+0] = 0
		atlas.Pix[i+1] = 0
		atlas.Pix[i+2] = 0
		atlas.Pix[i+3] = 255
	}
	pad := msdfPad
	x, y, rowH := pad, pad, 0
	ink := 0
	for _, r := range sdfCodepoints() {
		dr, mask, mp, adv, ok := face.Glyph(xfixed.Point26_6{}, r)
		if !ok {
			continue
		}
		advPx := float32(adv.Round())
		gw, gh := dr.Dx(), dr.Dy()
		if mask == nil || gw <= 0 || gh <= 0 {
			glyphs[r] = msdfGlyph{advance: advPx}
			continue
		}
		cw, ch := gw+2*pad, gh+2*pad
		if x+cw+pad > atlasW {
			x = pad
			y += rowH + pad
			rowH = 0
		}
		if y+ch+pad > atlasH {
			break
		}
		field := glyphMaskToSDFField(mask, mp, gw, gh, pad, msdfPxRange)
		dst := image.Rect(x, y, x+cw, y+ch)
		n := blitMSDFField(atlas, dst, field, cw, ch)
		if n == 0 {
			glyphs[r] = msdfGlyph{advance: advPx}
			continue
		}
		glyphs[r] = msdfGlyph{
			src:      rl.NewRectangle(float32(x), float32(y), float32(cw), float32(ch)),
			bearingX: float32(dr.Min.X - pad),
			bearingY: float32(dr.Min.Y - pad),
			advance:  advPx,
			hasInk:   true,
		}
		ink += n
		x += cw + pad
		if ch > rowH {
			rowH = ch
		}
	}
	_ = pixelSize
	if ink < 32 || len(glyphs) < 32 {
		return nil, nil, ink, false
	}
	return atlas, glyphs, ink, true
}

// glyphMaskToSDFField builds a padded true-SDF (0..1, edge at 0.5) as RGB-ready bytes.
func glyphMaskToSDFField(mask image.Image, mp image.Point, gw, gh, pad int, pxRange float64) []byte {
	cw, ch := gw+2*pad, gh+2*pad
	inside := make([]bool, cw*ch)
	for py := 0; py < gh; py++ {
		for px := 0; px < gw; px++ {
			_, _, _, a := mask.At(mp.X+px, mp.Y+py).RGBA()
			inside[(py+pad)*cw+(px+pad)] = a > 0x7fff
		}
	}
	dist := edtSigned(inside, cw, ch)
	out := make([]byte, cw*ch*4)
	for i, d := range dist {
		// Median shader: smoothstep(0.5±w, dist) → opaque when dist > 0.5.
		// edtSigned is outside-positive / inside-negative — invert so inside encodes high.
		sd := 0.5 - (d/pxRange)*0.5
		if sd < 0 {
			sd = 0
		}
		if sd > 1 {
			sd = 1
		}
		v := uint8(sd*255 + 0.5)
		j := i * 4
		out[j+0] = v
		out[j+1] = v
		out[j+2] = v
		out[j+3] = 255
	}
	return out
}

func blitMSDFField(atlas *image.RGBA, dst image.Rectangle, field []byte, cw, ch int) int {
	ink := 0
	for py := 0; py < ch; py++ {
		for px := 0; px < cw; px++ {
			j := (py*cw + px) * 4
			ax := dst.Min.X + px
			ay := dst.Min.Y + py
			i := atlas.PixOffset(ax, ay)
			atlas.Pix[i+0] = field[j+0]
			atlas.Pix[i+1] = field[j+1]
			atlas.Pix[i+2] = field[j+2]
			atlas.Pix[i+3] = 255
			// Inside encodes high after invert.
			if field[j+0] > 140 {
				ink++
			}
		}
	}
	return ink
}

// edtSigned returns outside-positive / inside-negative approximate Euclidean distance (px).
func edtSigned(inside []bool, w, h int) []float64 {
	inf := float64(w+h) * float64(w+h)
	outD := make([]float64, w*h)
	inD := make([]float64, w*h)
	for i, on := range inside {
		if on {
			outD[i] = 0
			inD[i] = inf
		} else {
			outD[i] = inf
			inD[i] = 0
		}
	}
	edt2D(outD, w, h)
	edt2D(inD, w, h)
	out := make([]float64, w*h)
	for i := range out {
		o := math.Sqrt(outD[i])
		in := math.Sqrt(inD[i])
		if inside[i] {
			out[i] = -in
		} else {
			out[i] = o
		}
	}
	return out
}

// edt2D in-place squared Euclidean distance transform (Felzenszwalb & Huttenlocher).
func edt2D(d []float64, w, h int) {
	for y := 0; y < h; y++ {
		edt1D(d[y*w : y*w+w])
	}
	col := make([]float64, h)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			col[y] = d[y*w+x]
		}
		edt1D(col)
		for y := 0; y < h; y++ {
			d[y*w+x] = col[y]
		}
	}
}

func edt1D(f []float64) {
	n := len(f)
	if n == 0 {
		return
	}
	v := make([]int, n)
	z := make([]float64, n+1)
	k := 0
	v[0] = 0
	z[0] = math.Inf(-1)
	z[1] = math.Inf(1)
	for q := 1; q < n; q++ {
		s := ((f[q] + float64(q*q)) - (f[v[k]] + float64(v[k]*v[k]))) / float64(2*q-2*v[k])
		for s <= z[k] {
			k--
			s = ((f[q] + float64(q*q)) - (f[v[k]] + float64(v[k]*v[k]))) / float64(2*q-2*v[k])
		}
		k++
		v[k] = q
		z[k] = s
		z[k+1] = math.Inf(1)
	}
	k = 0
	out := make([]float64, n)
	for q := 0; q < n; q++ {
		for z[k+1] < float64(q) {
			k++
		}
		dx := float64(q - v[k])
		out[q] = dx*dx + f[v[k]]
	}
	copy(f, out)
}

func unloadMSDFFont() {
	if msdfTex.ID != 0 {
		rl.UnloadTexture(msdfTex)
		msdfTex = rl.Texture2D{}
	}
	if msdfShader.ID != 0 {
		rl.UnloadShader(msdfShader)
		msdfShader = rl.Shader{}
	}
	msdfGlyphs = nil
	msdfReady = false
	msdfBasePx = 0
	msdfAscentPx = 0
}

func measureTextMSDF(text string, fontSize float32) float32 {
	if !msdfReady || msdfBasePx <= 0 {
		return 0
	}
	scale := msdfScale(fontSize)
	var w float32
	n := 0
	for _, r := range text {
		g, ok := msdfGlyphs[r]
		if !ok {
			if sp, ok := msdfGlyphs[' ']; ok {
				w += sp.advance * scale
			} else {
				w += fontSize * 0.5 * msdfOpticalScale
			}
			continue
		}
		w += g.advance * scale
		n++
	}
	if FontSpacing != 0 && n > 1 {
		w += FontSpacing * float32(n-1)
	}
	return w
}

func drawTextMSDF(text string, x, y, fontSize float32, color rl.Color) {
	if !msdfReady || msdfBasePx <= 0 || msdfTex.ID == 0 || msdfShader.ID == 0 {
		return
	}
	scale := msdfScale(fontSize)
	ascent := msdfAscentPx * scale
	baseline := y + ascent
	penX := x
	rl.BeginShaderMode(msdfShader)
	for _, r := range text {
		g, ok := msdfGlyphs[r]
		if !ok {
			if sp, ok := msdfGlyphs[' ']; ok {
				penX += sp.advance * scale
			} else {
				penX += fontSize * 0.5 * msdfOpticalScale
			}
			continue
		}
		if g.hasInk && g.src.Width > 0 && g.src.Height > 0 {
			dst := rl.NewRectangle(
				penX+g.bearingX*scale,
				baseline+g.bearingY*scale,
				g.src.Width*scale,
				g.src.Height*scale,
			)
			rl.DrawTexturePro(msdfTex, g.src, dst, rl.NewVector2(0, 0), 0, color)
		}
		penX += g.advance * scale
		if FontSpacing != 0 {
			penX += FontSpacing
		}
	}
	rl.EndShaderMode()
}

func msdfInkHeight(fontSize float32) float32 {
	if !msdfReady || msdfBasePx <= 0 {
		return fontSize
	}
	h := msdfAscentPx * msdfScale(fontSize)
	if h <= 0 {
		return fontSize * msdfOpticalScale
	}
	return h
}
