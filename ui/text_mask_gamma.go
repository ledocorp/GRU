package ui

import (
	"image"
	"math"
	"strconv"
	"strings"
)

// SkMaskGamma-lite (Raster Tragedy) for S1 glyph masks.
// Port of SkTMaskGamma_build_correcting_lut / apply_contrast — bake-time only.
// See docs/TEXT_1X_DECISION.md · https://skia.org/docs/dev/design/raster_tragedy/

const (
	// Defaults tuned toward Windows/Chrome-class dark-on-light chrome (not Linux 0.2).
	//
	// Chromium raised SK_GAMMA_CONTRAST to 1.0, but that lands on its own greyscale
	// mask + gamma LUT. Applied to these strike atlases it thickens stems on top of
	// the shader's ec/mute boosts, which is a visible weight jump on chrome copy.
	// Override with GRU_TEXT_CONTRAST=1.0 to A/B on the Glyph Quality ladder.
	maskGammaContrastDefault float32 = 0.55
	maskGammaDeviceDefault   float32 = 0 // 0 = sRGB transfer (Skia convention)
	maskGammaSrcDark         int     = 30 // near-black UI ink luminance bucket
)

// maskGammaApplyContrast is Skia's apply_contrast.
func maskGammaApplyContrast(srca, contrast float32) float32 {
	return srca + (1-srca)*contrast*srca
}

func maskGammaToLuma(gamma, c float32) float32 {
	if gamma == 0 {
		if c <= 0.04045 {
			return c / 12.92
		}
		return float32(math.Pow(float64((c+0.055)/1.055), 2.4))
	}
	if gamma == 1 {
		return c
	}
	return float32(math.Pow(float64(c), float64(gamma)))
}

func maskGammaFromLuma(gamma, luma float32) float32 {
	if gamma == 0 {
		if luma <= 0.0031308 {
			return luma * 12.92
		}
		return float32(1.055*math.Pow(float64(luma), 1/2.4) - 0.055)
	}
	if gamma == 1 {
		return luma
	}
	return float32(math.Pow(float64(luma), 1/float64(gamma)))
}

// buildMaskGammaLUT builds a 256-entry correcting table for paint luminance srcI (0–255).
func buildMaskGammaLUT(srcI int, contrast, deviceGamma float32) [256]uint8 {
	var table [256]uint8
	if srcI < 0 {
		srcI = 0
	}
	if srcI > 255 {
		srcI = 255
	}
	src := float32(srcI) / 255
	linSrc := maskGammaToLuma(deviceGamma, src)
	dst := 1 - src
	linDst := maskGammaToLuma(deviceGamma, dst)
	adjustedContrast := contrast * linDst

	if absFloat32(src-dst) < (1.0 / 256.0) {
		for i := 0; i < 256; i++ {
			raw := float32(i) / 255
			a := maskGammaApplyContrast(raw, adjustedContrast)
			if a > 1 {
				a = 1
			}
			table[i] = uint8(a*255 + 0.5)
		}
		return table
	}

	for i := 0; i < 256; i++ {
		raw := float32(i) / 255
		srca := maskGammaApplyContrast(raw, adjustedContrast)
		if srca > 1 {
			srca = 1
		}
		dsta := 1 - srca
		linOut := linSrc*srca + dsta*linDst
		out := maskGammaFromLuma(deviceGamma, linOut)
		result := (out - dst) / (src - dst)
		if result < 0 {
			result = 0
		}
		if result > 1 {
			result = 1
		}
		table[i] = uint8(result*255 + 0.5)
	}
	return table
}

func applyMaskGammaLUTAlpha(atlas *image.RGBA, table [256]uint8) {
	if atlas == nil {
		return
	}
	pix := atlas.Pix
	for i := 3; i < len(pix); i += 4 {
		pix[i] = table[pix[i]]
	}
}

func applyMaskGammaLUTRGB(atlas *image.RGBA, table [256]uint8) {
	if atlas == nil {
		return
	}
	pix := atlas.Pix
	for i := 0; i+3 < len(pix); i += 4 {
		pix[i+0] = table[pix[i+0]]
		pix[i+1] = table[pix[i+1]]
		pix[i+2] = table[pix[i+2]]
	}
}

func s1MaskGammaContrast() float32 {
	raw := strings.TrimSpace(EnvOr("GRU_TEXT_CONTRAST", "GORY_TEXT_CONTRAST"))
	if raw == "" {
		return maskGammaContrastDefault
	}
	f, err := strconv.ParseFloat(raw, 32)
	if err != nil {
		return maskGammaContrastDefault
	}
	v := float32(f)
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// enhanceS1StrikeAtlas applies SkMaskGamma-lite (+ optional fx contrast dual-path).
func enhanceS1StrikeAtlas(atlas *image.RGBA, lcd bool) {
	if atlas == nil {
		return
	}
	contrast := s1MaskGammaContrast()
	table := buildMaskGammaLUT(maskGammaSrcDark, contrast, maskGammaDeviceDefault)
	if lcd {
		applyMaskGammaLUTRGB(atlas, table)
	} else {
		applyMaskGammaLUTAlpha(atlas, table)
		// Do NOT tryFxBoostStrikeAlpha here — BoostRGBAAlpha punches white
		// speckles into greyscale sticky. Windows ClearType uses
		// enhanceS1DWriteLCDAtlas and never hits this path.
	}
}

// enhanceS1StrikeAtlasNoFx is greyscale gamma only — used for Inter/Fira aux strikes.
// fx BoostRGBAAlpha created bowl speckles / ghost stems on bundled faces.
func enhanceS1StrikeAtlasNoFx(atlas *image.RGBA) {
	if atlas == nil {
		return
	}
	contrast := s1MaskGammaContrast()
	table := buildMaskGammaLUT(maskGammaSrcDark, contrast, maskGammaDeviceDefault)
	applyMaskGammaLUTAlpha(atlas, table)
}

// enhanceS1DWriteLCDAtlas leaves DirectWrite ClearType_3x1 masks alone.
func enhanceS1DWriteLCDAtlas(atlas *image.RGBA) {
	_ = atlas
}
