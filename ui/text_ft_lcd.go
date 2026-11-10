package ui

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/image/font"
	xfixed "golang.org/x/image/math/fixed"
)

// T1.14 — LCD / ClearType-style Latin chrome (opt-in).
//
// RGB atlas stores per-subpixel coverage. Draw mixes against an opaque background
// (heuristic: dark text → white panel; light text → dark chrome). Transparent /
// unknown backgrounds should stay on ft-grey.
//
// Dest must match the pixels under the glyph (`PushLCDSurfaceBackdrop` for
// tiles/cards/buttons; page fill from BeginSuperFrame). Mix-in-shader + BlendAlpha
// onto the wrong fill (page grey on white tiles, or #1E1E1E on purple) reads
// as Semibold / muddy labels. Light dest: opaque mixed RGB. Dark dest: coverage
// alpha (overlay Regular). See docs/PRESENT_1X_COMPOSITE.md.
//
// Opt-in: GRU_TEXT_LCD=1. Same Face advances as ft-grey (RichText-safe).
// Footprint: same ~48×DPI / 1024² class as greyscale — do not dual-load SDF.

type ftLCDGlyph struct {
	src      rl.Rectangle
	bearingX float32
	bearingY float32
	advance  float32
	hasInk   bool
}

var (
	ftLCDReady    bool
	ftLCDTex      rl.Texture2D
	ftLCDBasePx   float32
	ftLCDAscentPx float32
	ftLCDGlyphs   map[rune]ftLCDGlyph
	ftLCDShader   rl.Shader
	ftLCDWarn     sync.Once
	ftLCDBgLoc           int32 = -1
	ftLCDGammaLoc        int32 = -1
	ftLCDContrastLoc     int32 = -1
	ftLCDClearTypeLoc    int32 = -1
	ftLCDWidgetFillLoc   int32 = -1
	ftLCDCoverLumLoc     int32 = -1
	ftLCDHostileBoostLoc int32 = -1
)

const ftLCDOpticalScale float32 = 0.93

// PreferFTLCDChrome enables ClearType-style RGB sticky strikes.
//
// Auto-on only when DirectWrite is available (Windows). FreeType "lite LCD"
// (neighbor RGB) on Linux looks soft/fringed at 1:1 — greyscale sticky is
// sharper there. Opt in anywhere with GRU_TEXT_LCD=1; opt out with =0.
func PreferFTLCDChrome() bool {
	v := strings.ToLower(strings.TrimSpace(EnvOr("GRU_TEXT_LCD", "GORY_TEXT_LCD")))
	switch v {
	case "0", "off", "false", "no", "grey", "gray":
		return false
	case "1", "on", "true", "yes", "lcd", "cleartype":
		return true
	default:
		return PreferFTGreyStrikes() && EffectiveSupersamplingScale() <= 1.01 && dwriteLCDAvailable()
	}
}

// PreferAuxFaceLCD enables DirectWrite ClearType for Fira/Inter sticky strikes.
// Default on when S1 LCD + DWrite are available (same recipe as Segoe chrome).
// Opt out to greyscale: GRU_AUX_LCD=0 (then PreferAuxFaceDWriteGrey).
func PreferAuxFaceLCD() bool {
	v := strings.ToLower(strings.TrimSpace(EnvOr("GRU_AUX_LCD", "GORY_AUX_LCD")))
	switch v {
	case "0", "off", "false", "no", "grey", "gray":
		return false
	case "1", "on", "true", "yes", "lcd", "cleartype":
		return true
	default:
		return PreferFTLCDChrome() && dwriteLCDAvailable()
	}
}

// PreferAuxFaceLCDSoft softens Inter/Fira toward browser look: NATURAL_SYMMETRIC
// bake at all sticky sizes + eased LCD blend contrast. Default on.
// Opt out: GRU_AUX_LCD_SOFT=0 (NATURAL bake like chrome).
func PreferAuxFaceLCDSoft() bool {
	if !PreferAuxFaceLCD() {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(EnvOr("GRU_AUX_LCD_SOFT", "GORY_AUX_LCD_SOFT")))
	switch v {
	case "0", "off", "false", "no", "hard", "chrome":
		return false
	default:
		return true
	}
}

// PreferAuxFaceDWriteGrey uses DirectWrite greyscale when LCD is opted out.
// Default on when DWrite is available and GRU_AUX_LCD=0. Opt out: GRU_AUX_DWRITE=0.
func PreferAuxFaceDWriteGrey() bool {
	if PreferAuxFaceLCD() {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(EnvOr("GRU_AUX_DWRITE", "GORY_AUX_DWRITE")))
	switch v {
	case "0", "off", "false", "no", "ft", "freetype":
		return false
	default:
		return dwriteLCDAvailable()
	}
}

func ftLCDActive(mono, preview bool) bool {
	if textSkipLCD > 0 {
		return false
	}
	return PreferFTLCDChrome() && ftLCDReady && !mono && !preview
}

func ftLCDScale(fontSize float32) float32 {
	if ftLCDBasePx <= 0 {
		return 1
	}
	return (fontSize / ftLCDBasePx) * ftLCDOpticalScale
}

func ftLCDFragSource() string {
	// Linear-light component-alpha ClearType (Skia/Chrome-class over opaque bg).
	// Coverage is per-channel; blend in gamma space from DWrite GetAlphaBlendParams.
	return `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
uniform sampler2D texture0;
uniform vec4 colDiffuse;
uniform vec3 uBg;
uniform float uGamma;
uniform float uContrast;
uniform float uClearType;
uniform float uWidgetFill;
// Paint-fill luminance for coverage darkBoost. When rest lifts uBg for hostile
// light ink (C22/C23), mix dest brightens but chrome stays dark — darkBoost must
// follow the painted plate or lift and boost cancel (C22 looked like a no-op).
uniform float uCoverLum;
// Midtone densify for light ink on saturated widget chrome (C24/C25).
// C25: all sat chrome uses greyscale CT (GTK4/Flutter) so Primary matches Success.
uniform float uHostileBoost;
out vec4 finalColor;

vec3 toLinear(vec3 c, float g) {
    return pow(max(c, vec3(0.0)), vec3(g));
}
vec3 toDisplay(vec3 c, float g) {
    return pow(max(c, vec3(0.0)), vec3(1.0 / g));
}

void main() {
    vec3 cover = texture(texture0, fragTexCoord).rgb;
    float ct = clamp(uClearType, 0.0, 1.0);
    float mid = cover.g;
    // Preserve RGB ClearType when ct is high; collapse fringe when soft.
    cover.r = mix(mid, cover.r, ct);
    cover.b = mix(mid, cover.b, ct);

    // Mild enhanced-contrast only — DWrite masks already carry system weight.
    // Strong ec (Chrome greyscale 1.0) on CT_3x1 eats horizontal bars.
    float ec = clamp(uContrast, 0.0, 1.0) * 0.28;
    cover = clamp(cover + ec * cover * (1.0 - cover), 0.0, 1.0);

    float g = max(uGamma, 1.0);
    vec3 text = fragColor.rgb * colDiffuse.rgb;
    vec3 bg = uBg;
    // Mid-grey chrome (captions/subtitles) reads washed on light panels — steepen
    // coverage so muted ink reaches its color sooner without baking the atlas.
    float textLum = dot(text, vec3(0.2126, 0.7152, 0.0722));
    float muteBoost = smoothstep(0.12, 0.48, textLum) * 0.38;
    // Light-on-dark (editor): slight extra cover so white/grey ink stays sharp.
    float bgLum = dot(bg, vec3(0.2126, 0.7152, 0.0722));
    float coverLum = uCoverLum >= 0.0 ? uCoverLum : bgLum;
    float darkBoost = (1.0 - smoothstep(0.12, 0.42, coverLum)) * 0.14;
    cover = clamp(cover + (muteBoost + darkBoost) * cover * (1.0 - cover), 0.0, 1.0);
    // Hostile light-on-sat: greyscale CT lacks Success/Teal fringe fill-in — densify midtones (C24).
    float hb = clamp(uHostileBoost, 0.0, 1.0);
    cover = clamp(cover + hb * 0.36 * cover * (1.0 - cover), 0.0, 1.0);
    // Per-channel component alpha in linear light.
    vec3 outL = mix(toLinear(bg, g), toLinear(text, g), cover);
    vec3 rgb = toDisplay(outL, g);
    float coverA = max(cover.r, max(cover.g, cover.b));
    if (coverA < 0.004) discard;
    // Light dest: rgb is already mixed against uBg. Output opaque so BlendAlpha
    // does not apply coverage again (grey page mix onto white tiles = fat Regular).
    // Dark dest (overlay menu/footer): keep coverage alpha — that is the thin look.
    float a = fragColor.a * colDiffuse.a;
    // Overlay menu/footer (no pushed fill): coverage alpha on dark chrome = thin Regular.
    // Widget fill was pushed (button, badge, menubar plate): rgb is already mixed against
    // that opaque fill — do not apply coverage alpha again (white-on-primary halo).
    if (bgLum <= 0.45 && uWidgetFill < 0.5) {
        a *= coverA;
    }
    finalColor = vec4(rgb, a);
}`
}

// InitFTLCDFont builds an LCD RGB atlas at pixelSize (EffectiveFTGreyAtlasSize).
func InitFTLCDFont(pixelSize int32) bool {
	unloadFTLCDFont()
	if pixelSize < 12 {
		pixelSize = 12
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
	atlas, glyphs, ink, ok := packFTLCDAtlas(face, pixelSize)
	_ = face.Close()
	if !ok {
		ftLCDWarn.Do(func() {
			fmt.Printf("Gru: ft-lcd atlas failed (ink=%d) — staying on ft-grey\n", ink)
		})
		return false
	}
	applyFTLCDGammaContrast(atlas, 1.45, 0.18)
	if !ensureFTLCDShader() {
		return false
	}
	tex := GoImageToTexture(atlas)
	if tex.ID == 0 {
		return false
	}
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	ftLCDGlyphs = glyphs
	ftLCDTex = tex
	ftLCDBasePx = float32(pixelSize)
	ftLCDAscentPx = ascent
	ftLCDReady = true
	return true
}

func packFTLCDAtlas(face font.Face, pixelSize int32) (*image.RGBA, map[rune]ftLCDGlyph, int, bool) {
	glyphs := make(map[rune]ftLCDGlyph, 256)
	atlasW, atlasH := 1024, 1024
	if pixelSize >= 64 {
		atlasW, atlasH = 2048, 2048
	}
	atlas := image.NewRGBA(image.Rect(0, 0, atlasW, atlasH))
	pad := 2
	x, y, rowH := pad, pad, 0
	ink := 0
	for _, r := range sdfCodepoints() {
		dr, mask, mp, adv, ok := face.Glyph(xfixed.Point26_6{}, r)
		if !ok {
			continue
		}
		advPx := fixed26ToPx(adv)
		gw, gh := dr.Dx(), dr.Dy()
		if mask == nil || gw <= 0 || gh <= 0 {
			glyphs[r] = ftLCDGlyph{advance: advPx}
			continue
		}
		if x+gw+pad > atlasW {
			x = pad
			y += rowH + pad
			rowH = 0
		}
		if y+gh+pad > atlasH {
			break
		}
		dst := image.Rect(x, y, x+gw, y+gh)
		n := blitFTLCDMask(atlas, dst, mask, mp)
		if n == 0 {
			glyphs[r] = ftLCDGlyph{advance: advPx}
			continue
		}
		glyphs[r] = ftLCDGlyph{
			src:      rl.NewRectangle(float32(x), float32(y), float32(gw), float32(gh)),
			bearingX: float32(dr.Min.X),
			bearingY: float32(dr.Min.Y),
			advance:  advPx,
			hasInk:   true,
		}
		ink += n
		x += gw + pad
		if gh > rowH {
			rowH = gh
		}
	}
	_ = pixelSize
	if ink < 32 || len(glyphs) < 32 {
		return nil, nil, ink, false
	}
	return atlas, glyphs, ink, true
}

// blitFTLCDMask converts a greyscale alpha mask into RGB subpixel coverage (ClearType-lite).
func blitFTLCDMask(atlas *image.RGBA, dst image.Rectangle, mask image.Image, mp image.Point) int {
	gw, gh := dst.Dx(), dst.Dy()
	at := func(px, py int) float32 {
		if py < 0 || py >= gh || px < 0 || px >= gw {
			return 0
		}
		_, _, _, a := mask.At(mp.X+px, mp.Y+py).RGBA()
		return float32(a) / 65535
	}
	ink := 0
	for py := 0; py < gh; py++ {
		for px := 0; px < gw; px++ {
			c := at(px, py)
			l := at(px-1, py)
			r := at(px+1, py)
			// R/G/B ≈ coverage at −⅓ / 0 / +⅓ px (neighbor lerp stand-in for LCD filter).
			cr := clamp01f(0.33*l + 0.67*c)
			cg := clamp01f(c)
			cb := clamp01f(0.67*c + 0.33*r)
			ax := dst.Min.X + px
			ay := dst.Min.Y + py
			i := atlas.PixOffset(ax, ay)
			atlas.Pix[i+0] = uint8(cr*255 + 0.5)
			atlas.Pix[i+1] = uint8(cg*255 + 0.5)
			atlas.Pix[i+2] = uint8(cb*255 + 0.5)
			atlas.Pix[i+3] = 255
			if cg > 0.02 || cr > 0.02 || cb > 0.02 {
				ink++
			}
		}
	}
	return ink
}

// ensureFTLCDShader loads the ClearType mix shader without building a scaled atlas.
// Used by S1 LCD strikes (exact-size RGB atlases drawn through this shader).
func ensureFTLCDShader() bool {
	if ftLCDShader.ID != 0 {
		return true
	}
	shader := rl.LoadShaderFromMemory("", ftLCDFragSource())
	if shader.ID == 0 {
		return false
	}
	ftLCDShader = shader
	ftLCDBgLoc = rl.GetShaderLocation(shader, "uBg")
	ftLCDGammaLoc = rl.GetShaderLocation(shader, "uGamma")
	ftLCDContrastLoc = rl.GetShaderLocation(shader, "uContrast")
	ftLCDClearTypeLoc = rl.GetShaderLocation(shader, "uClearType")
	ftLCDWidgetFillLoc = rl.GetShaderLocation(shader, "uWidgetFill")
	ftLCDCoverLumLoc = rl.GetShaderLocation(shader, "uCoverLum")
	ftLCDHostileBoostLoc = rl.GetShaderLocation(shader, "uHostileBoost")
	return true
}

// bindFTLCDBlendParams pushes DWrite (or default) ClearType blend knobs into the LCD shader.
func bindFTLCDBlendParams() {
	bindFTLCDBlendParamsForSize(18, 1)
}

// bindFTLCDBlendParamsForSize tunes ClearType for size/scale.
// Uses DirectWrite GetAlphaBlendParams as-is (do not force Chrome greyscale
// contrast=1.0 into LCD — that thins crossbars and looks uneven).
// softAux: Inter/Fira — ease EnhancedContrast so weight matches web Regular better.
// lightOnSat / greyOnSat: light or mid-grey ink on saturated widget fills — bump
// contrast after ClearType collapse so labels keep body weight (C18).
func bindFTLCDBlendParamsForSize(fontSize, scale float32) {
	bindFTLCDBlendParamsForSizeOpt(fontSize, scale, false, false, false, false)
}

func bindFTLCDBlendParamsForSizeOpt(fontSize, scale float32, softAux, lightOnSat, greyOnSat, hostileLightDest bool) {
	if ftLCDShader.ID == 0 {
		return
	}
	gamma := dwriteBlendGamma
	contrast := dwriteBlendContrast
	ct := dwriteClearTypeLevel
	if ct < 0.4 {
		ct = 0.4
	}
	if ct > 1 {
		ct = 1
	}
	// Optional: GRU_TEXT_LCD_CONTRAST overrides DWrite enhancedContrast only.
	if raw := strings.TrimSpace(EnvOr("GRU_TEXT_LCD_CONTRAST", "GORY_TEXT_LCD_CONTRAST")); raw != "" {
		if f, err := strconv.ParseFloat(raw, 32); err == nil {
			contrast = float32(f)
			if contrast < 0 {
				contrast = 0
			}
			if contrast > 1 {
				contrast = 1
			}
		}
	} else if softAux {
		// Browser Inter/Fira reads leaner; system EnhancedContrast fattens aux.
		// Extra ease under 20px — small NATURAL leftovers looked denser than 24+.
		if fontSize <= 20 {
			contrast *= 0.35
		} else {
			contrast *= 0.45
		}
	}
	if lightOnSat {
		// Collapsed CT on saturated fills thins white labels — restore body weight.
		if contrast < 0.62 {
			contrast = 0.62
		}
		contrast *= 1.4
		// Hostile hues (blue/red/magenta): greyscale CT + higher contrast floor so
		// rest matches hover luminance punch without washing the painted fill (C22).
		if hostileLightDest {
			if contrast < 0.92 {
				contrast = 0.92
			}
			contrast *= 1.12
		}
		if contrast > 1 {
			contrast = 1
		}
	} else if greyOnSat {
		// Mid-grey on chroma also collapses — lighter boost than pure white.
		if contrast < 0.45 {
			contrast = 0.45
		}
		contrast *= 1.2
		if contrast > 1 {
			contrast = 1
		}
	}
	// Keep ClearType fringe strong — collapsing CT made grey captions soft.
	if scale < 0.94 || scale > 1.06 {
		ct *= 0.95
	}
	if softAux {
		ct *= 0.85
	}
	if ftLCDGammaLoc >= 0 {
		rl.SetShaderValue(ftLCDShader, ftLCDGammaLoc, []float32{gamma}, rl.ShaderUniformFloat)
	}
	if ftLCDContrastLoc >= 0 {
		rl.SetShaderValue(ftLCDShader, ftLCDContrastLoc, []float32{contrast}, rl.ShaderUniformFloat)
	}
	if ftLCDClearTypeLoc >= 0 {
		rl.SetShaderValue(ftLCDShader, ftLCDClearTypeLoc, []float32{ct}, rl.ShaderUniformFloat)
	}
	_ = fontSize
}

func lcdDestSaturated(bg rl.Vector3) bool {
	mn, mx := bg.X, bg.X
	if bg.Y < mn {
		mn = bg.Y
	}
	if bg.Z < mn {
		mn = bg.Z
	}
	if bg.Y > mx {
		mx = bg.Y
	}
	if bg.Z > mx {
		mx = bg.Z
	}
	return mx-mn > 0.22
}

// lcdDestHostileForLightInk reports blue/red/magenta-dominant fills where
// ClearType RGB fringe shares channels with the backdrop, so white labels look
// thinner than on green/teal/amber (Glyph Quality Primary/Delete/Magenta vs
// Success/Warning/Teal — C19).
func lcdDestHostileForLightInk(bg rl.Vector3) bool {
	// Chromaticity (R/G/B share), not absolute levels. Hover ColorBrightness lifts
	// all channels toward white and can drop mx−mn below lcdDestSaturated, which
	// flipped Magenta/Primary/Delete into a sharper CT path on hover only (C21).
	s := bg.X + bg.Y + bg.Z
	if s < 0.08 {
		return false
	}
	nx, ny, nz := bg.X/s, bg.Y/s, bg.Z/s
	// Blue / indigo / violet family.
	if nz > 0.38 && nz-ny > 0.06 {
		return true
	}
	// Red / magenta family (magenta also high B — often caught above).
	if nx > 0.38 && nx-ny > 0.06 {
		return true
	}
	return false
}

// lcdLiftHostileMixDest matches buttonInteractionColors hover ColorBrightness(+0.25)
// for the LCD mix only. Painted chrome stays at the true fill (C22/C23).
// Already-bright dests (hover ColorBrightness) skip the lift so we do not stack.
func lcdLiftHostileMixDest(bg rl.Vector3) rl.Vector3 {
	y := 0.2126*bg.X + 0.7152*bg.Y + 0.0722*bg.Z
	// C25: skip only when hover-bright so Success (~0.50) can share the lift.
	if y >= 0.62 {
		return bg
	}
	const t = float32(0.25)
	return rl.NewVector3(bg.X+(1-bg.X)*t, bg.Y+(1-bg.Y)*t, bg.Z+(1-bg.Z)*t)
}

func lcdClearTypeForDest(ct float32, bg rl.Vector3) float32 {
	return lcdClearTypeForDestOpt(ct, bg, false)
}

// lcdClearTypeForDestOpt: chromeLight (light ink on pushed widget fill) always
// greyscale — GTK4/Flutter chrome recipe so Primary weight matches Success (C25).
// Page text keeps LCD on neutral dests; hostile page fills still collapse CT.
func lcdClearTypeForDestOpt(ct float32, bg rl.Vector3, chromeLight bool) float32 {
	if chromeLight {
		return 0
	}
	// Blue/red/magenta page fills: greyscale AA (DWRITE_PIXEL_GEOMETRY_FLAT).
	if lcdDestHostileForLightInk(bg) {
		return 0
	}
	if lcdDestSaturated(bg) {
		ct *= 0.28
		if ct < 0.12 {
			ct = 0.12
		}
	}
	return ct
}

// applyFTLCDGammaContrast tightens LCD RGB coverage (Skia-lite; dark-on-light).
func applyFTLCDGammaContrast(atlas *image.RGBA, gamma, contrast float32) {
	if atlas == nil || (gamma <= 1.001 && contrast <= 0) {
		return
	}
	invG := float32(1)
	if gamma > 1.001 {
		invG = 1 / gamma
	}
	pix := atlas.Pix
	for i := 0; i+3 < len(pix); i += 4 {
		for c := 0; c < 3; c++ {
			a := float32(pix[i+c]) / 255
			if a <= 0 || a >= 1 {
				continue
			}
			if invG != 1 {
				a = pow01(a, invG)
			}
			if contrast > 0 {
				a = a + contrast*a*(1-a)
			}
			if a > 1 {
				a = 1
			}
			pix[i+c] = uint8(a*255 + 0.5)
		}
	}
}

func pow01(a, exp float32) float32 {
	// Fast pow for 0..1 via exp(exp*ln(a)); enough for glyph masks.
	if a <= 0 {
		return 0
	}
	if a >= 1 {
		return 1
	}
	return float32(math.Pow(float64(a), float64(exp)))
}

func clamp01f(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func unloadFTLCDFont() {
	if ftLCDTex.ID != 0 {
		rl.UnloadTexture(ftLCDTex)
		ftLCDTex = rl.Texture2D{}
	}
	if ftLCDShader.ID != 0 {
		rl.UnloadShader(ftLCDShader)
		ftLCDShader = rl.Shader{}
	}
	ftLCDGlyphs = nil
	ftLCDReady = false
	ftLCDBasePx = 0
	ftLCDAscentPx = 0
	ftLCDBgLoc = -1
	ftLCDGammaLoc = -1
	ftLCDContrastLoc = -1
	ftLCDClearTypeLoc = -1
	ftLCDWidgetFillLoc = -1
	ftLCDCoverLumLoc = -1
	ftLCDHostileBoostLoc = -1
}

// lcdWidgetFillMixOutput is true when glyphs mix against a pushed opaque widget fill.
// Output stays opaque mixed RGB (no coverage-alpha blit) — see PRESENT_1X_COMPOSITE.md §3.
func lcdWidgetFillMixOutput() bool {
	return len(lcdSurfaceBackdropStack) > 0
}

func measureTextFTLCD(text string, fontSize float32) float32 {
	if !ftLCDReady || ftLCDBasePx <= 0 {
		return 0
	}
	scale := ftLCDScale(fontSize)
	var w float32
	n := 0
	for _, r := range text {
		g, ok := ftLCDGlyphs[r]
		if !ok {
			if sp, ok := ftLCDGlyphs[' ']; ok {
				w += sp.advance * scale
			} else {
				w += fontSize * 0.5 * ftLCDOpticalScale
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

// ftLCDLightBackdrop is the ClearType mix bg for dark/muted glyphs (page fill).
// Pure white was wrong for Demo Directory (#E4E5EB) and washed grey captions.
var ftLCDLightBackdrop = rl.NewVector3(228/255.0, 229/255.0, 235/255.0)

// ftLCDDarkBackdrop is the ClearType mix bg for light glyphs (editor / status).
var ftLCDDarkBackdrop = rl.NewVector3(30/255.0, 30/255.0, 30/255.0)

// lcdSurfaceBackdropStack is the opaque fill under the glyphs currently drawing
// (list-tile white, selected row, …). Mix-in-shader must match that fill, not
// the page grey from BeginSuperFrame — mismatch is the Directory-vs-overlay split.
var lcdSurfaceBackdropStack []rl.Vector3

func colorToLCDBackdrop(c rl.Color) rl.Vector3 {
	return rl.NewVector3(float32(c.R)/255, float32(c.G)/255, float32(c.B)/255)
}

// PushLCDSurfaceBackdrop sets the ClearType mix dest for nested widget fills.
// No-op (and returns false) when the fill is not opaque.
func PushLCDSurfaceBackdrop(c rl.Color) bool {
	if c.A < 250 {
		return false
	}
	lcdSurfaceBackdropStack = append(lcdSurfaceBackdropStack, colorToLCDBackdrop(c))
	return true
}

// PopLCDSurfaceBackdrop undoes a successful PushLCDSurfaceBackdrop.
func PopLCDSurfaceBackdrop(pushed bool) {
	if !pushed {
		return
	}
	n := len(lcdSurfaceBackdropStack)
	if n == 0 {
		return
	}
	lcdSurfaceBackdropStack = lcdSurfaceBackdropStack[:n-1]
}

// SetFTLCDLightBackdrop syncs ClearType light mix to the frame clear / page color.
// Call from BeginSuperFrame (and 1× cache clear) so LCD matches the real panel.
func SetFTLCDLightBackdrop(c rl.Color) {
	ftLCDLightBackdrop = rl.NewVector3(float32(c.R)/255, float32(c.G)/255, float32(c.B)/255)
}

// SetFTLCDDarkBackdrop syncs ClearType dark mix (white-on-dark editor / chrome).
func SetFTLCDDarkBackdrop(c rl.Color) {
	ftLCDDarkBackdrop = rl.NewVector3(float32(c.R)/255, float32(c.G)/255, float32(c.B)/255)
}

// lcdBgForText picks an opaque backdrop for ClearType mix.
// A pushed widget fill wins for both dark and light ink (white-on-purple must
// mix against purple, not generic #1E1E1E). Empty stack: light ink → dark chrome
// (overlay menu/footer); dark ink → page grey.
func lcdBgForText(color rl.Color) rl.Vector3 {
	if n := len(lcdSurfaceBackdropStack); n > 0 {
		return lcdSurfaceBackdropStack[n-1]
	}
	if int(color.R)+int(color.G)+int(color.B) >= 420 {
		return ftLCDDarkBackdrop
	}
	return ftLCDLightBackdrop
}

func drawTextFTLCD(text string, x, y, fontSize float32, color rl.Color) {
	if !ftLCDReady || ftLCDBasePx <= 0 || ftLCDTex.ID == 0 || ftLCDShader.ID == 0 {
		return
	}
	scale := ftLCDScale(fontSize)
	ascent := ftLCDAscentPx * scale
	baseline := y + ascent
	penX := x
	bgPaint := lcdBgForText(color)
	inkSum := int(color.R) + int(color.G) + int(color.B)
	lightInk := inkSum >= 420
	greyInk := !lightInk && inkSum >= 280 && inkSum < 420
	widgetChrome := lcdWidgetFillMixOutput()
	// C25: light labels on filled controls → greyscale AA (GTK4/Flutter chrome).
	// Same recipe for Primary/Delete/Magenta and Success/Warning/Teal.
	chromeLight := widgetChrome && lightInk
	satFill := widgetChrome && lcdDestSaturated(bgPaint)
	bgMix := bgPaint
	paintLum := 0.2126*bgPaint.X + 0.7152*bgPaint.Y + 0.0722*bgPaint.Z
	chromeBoost := float32(0)
	if chromeLight && satFill {
		bgMix = lcdLiftHostileMixDest(bgPaint)
		if paintLum < 0.62 {
			chromeBoost = 1
		}
	}
	widgetFill := float32(0)
	if widgetChrome {
		widgetFill = 1
	}
	rl.BeginShaderMode(ftLCDShader)
	if ftLCDBgLoc >= 0 {
		rl.SetShaderValue(ftLCDShader, ftLCDBgLoc, []float32{bgMix.X, bgMix.Y, bgMix.Z}, rl.ShaderUniformVec3)
	}
	if ftLCDCoverLumLoc >= 0 {
		rl.SetShaderValue(ftLCDShader, ftLCDCoverLumLoc, []float32{paintLum}, rl.ShaderUniformFloat)
	}
	if ftLCDWidgetFillLoc >= 0 {
		rl.SetShaderValue(ftLCDShader, ftLCDWidgetFillLoc, []float32{widgetFill}, rl.ShaderUniformFloat)
	}
	if ftLCDHostileBoostLoc >= 0 {
		rl.SetShaderValue(ftLCDShader, ftLCDHostileBoostLoc, []float32{chromeBoost}, rl.ShaderUniformFloat)
	}
	// Contrast floor for all light-on-sat chrome (hue-agnostic).
	bindFTLCDBlendParamsForSizeOpt(fontSize, 1, false, lightInk && satFill, greyInk && satFill, chromeLight && satFill)
	if ftLCDClearTypeLoc >= 0 {
		ct := dwriteClearTypeLevel
		if ct < 0.4 {
			ct = 0.4
		}
		if ct > 1 {
			ct = 1
		}
		ct = lcdClearTypeForDestOpt(ct, bgPaint, chromeLight)
		rl.SetShaderValue(ftLCDShader, ftLCDClearTypeLoc, []float32{ct}, rl.ShaderUniformFloat)
	}
	for _, r := range text {
		g, ok := ftLCDGlyphs[r]
		if !ok {
			if sp, ok := ftLCDGlyphs[' ']; ok {
				penX += sp.advance * scale
			} else {
				penX += fontSize * 0.5 * ftLCDOpticalScale
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
			rl.DrawTexturePro(ftLCDTex, g.src, dst, rl.NewVector2(0, 0), 0, color)
		}
		penX += g.advance * scale
		if FontSpacing != 0 {
			penX += FontSpacing
		}
	}
	rl.EndShaderMode()
}

func ftLCDInkHeight(fontSize float32) float32 {
	if !ftLCDReady || ftLCDBasePx <= 0 {
		return fontSize
	}
	h := ftLCDAscentPx * ftLCDScale(fontSize)
	if h <= 0 {
		return fontSize * ftLCDOpticalScale
	}
	return h
}
