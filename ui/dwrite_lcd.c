/* T1.14b — DirectWrite ClearType_3x1 glyph raster for Gru S1 (Windows + cgo). */
#ifdef _WIN32
#define COBJMACROS
#include <windows.h>
#include <dwrite.h>
#include <stdlib.h>
#include <string.h>

#include "dwrite_lcd.h"

/* MinGW often lacks IID_IDWriteFactory in uuid; define locally. */
static const GUID GRU_IID_IDWriteFactory =
	{0xb859ee5a, 0xd838, 0x4b5b, {0xa2, 0xe8, 0x1a, 0xdc, 0x7d, 0x93, 0xdb, 0x48}};

static IDWriteFactory *g_factory;

static int ensure_factory(void) {
	if (g_factory) {
		return 1;
	}
	HRESULT hr = DWriteCreateFactory(
		DWRITE_FACTORY_TYPE_SHARED,
		&GRU_IID_IDWriteFactory,
		(IUnknown **)&g_factory);
	return SUCCEEDED(hr) && g_factory != NULL;
}

void gru_dwrite_free(void *p) {
	free(p);
}

int gru_dwrite_lcd_available(void) {
	return ensure_factory();
}

int gru_dwrite_lcd_blend_params(float *gamma, float *enhanced_contrast, float *cleartype_level) {
	if (!gamma || !enhanced_contrast || !cleartype_level) {
		return 0;
	}
	*gamma = 1.8f;
	*enhanced_contrast = 0.5f;
	*cleartype_level = 0.5f;
	if (!ensure_factory()) {
		return 0;
	}
	IDWriteRenderingParams *params = NULL;
	HRESULT hr = IDWriteFactory_CreateRenderingParams(g_factory, &params);
	if (FAILED(hr) || !params) {
		return 0;
	}
	*gamma = IDWriteRenderingParams_GetGamma(params);
	*enhanced_contrast = IDWriteRenderingParams_GetEnhancedContrast(params);
	*cleartype_level = IDWriteRenderingParams_GetClearTypeLevel(params);
	IDWriteRenderingParams_Release(params);
	return 1;
}

int gru_dwrite_lcd_glyph(
	const uint16_t *font_path_utf16,
	int path_len,
	float em_size_px,
	uint32_t codepoint,
	float baseline_origin_x,
	int render_mode,
	uint8_t **out_rgb,
	int *out_w,
	int *out_h,
	int *bearing_x,
	int *bearing_y,
	float *advance)
{
	*out_rgb = NULL;
	*out_w = *out_h = 0;
	*bearing_x = *bearing_y = 0;
	*advance = 0;
	if (!font_path_utf16 || path_len <= 0 || em_size_px < 6.0f || !out_rgb) {
		return -1;
	}
	if (!ensure_factory()) {
		return -2;
	}

	IDWriteFontFile *font_file = NULL;
	HRESULT hr = IDWriteFactory_CreateFontFileReference(
		g_factory, (const WCHAR *)font_path_utf16, NULL, &font_file);
	if (FAILED(hr) || !font_file) {
		return -3;
	}

	IDWriteFontFace *face = NULL;
	hr = IDWriteFactory_CreateFontFace(
		g_factory, DWRITE_FONT_FACE_TYPE_TRUETYPE, 1, &font_file, 0, DWRITE_FONT_SIMULATIONS_NONE, &face);
	IDWriteFontFile_Release(font_file);
	if (FAILED(hr) || !face) {
		return -4;
	}

	UINT32 cp = codepoint;
	UINT16 glyph_index = 0;
	hr = IDWriteFontFace_GetGlyphIndices(face, &cp, 1, &glyph_index);
	if (FAILED(hr) || glyph_index == 0) {
		IDWriteFontFace_Release(face);
		return -5;
	}

	DWRITE_GLYPH_METRICS metrics;
	hr = IDWriteFontFace_GetDesignGlyphMetrics(face, &glyph_index, 1, &metrics, FALSE);
	if (FAILED(hr)) {
		IDWriteFontFace_Release(face);
		return -6;
	}
	UINT16 units_per_em = 0;
	{
		DWRITE_FONT_METRICS fm;
		IDWriteFontFace_GetMetrics(face, &fm);
		units_per_em = fm.designUnitsPerEm;
	}
	if (units_per_em == 0) {
		IDWriteFontFace_Release(face);
		return -7;
	}
	float scale = em_size_px / (float)units_per_em;
	*advance = metrics.advanceWidth * scale;

	FLOAT advances = *advance;
	DWRITE_GLYPH_OFFSET offset = {0, 0};
	DWRITE_GLYPH_RUN run;
	memset(&run, 0, sizeof(run));
	run.fontFace = face;
	run.fontEmSize = em_size_px;
	run.glyphCount = 1;
	run.glyphIndices = &glyph_index;
	run.glyphAdvances = &advances;
	run.glyphOffsets = &offset;
	run.isSideways = FALSE;
	run.bidiLevel = 0;

	/* Chrome: CLEARTYPE_NATURAL (hinted stems). Aux display: optional
	 * CLEARTYPE_NATURAL_SYMMETRIC (render_mode=1) — softer curves toward browser
	 * Inter/Fira; do not use on Segoe chrome (reads bold/fuzzy at UI sizes). */
	DWRITE_RENDERING_MODE mode = DWRITE_RENDERING_MODE_CLEARTYPE_NATURAL;
	if (render_mode == 1) {
		mode = DWRITE_RENDERING_MODE_CLEARTYPE_NATURAL_SYMMETRIC;
	}
	IDWriteGlyphRunAnalysis *analysis = NULL;
	hr = IDWriteFactory_CreateGlyphRunAnalysis(
		g_factory,
		&run,
		1.0f,
		NULL,
		mode,
		DWRITE_MEASURING_MODE_NATURAL,
		baseline_origin_x,
		0.0f,
		&analysis);
	if (FAILED(hr) || !analysis) {
		IDWriteFontFace_Release(face);
		return -8;
	}

	RECT bounds;
	hr = IDWriteGlyphRunAnalysis_GetAlphaTextureBounds(
		analysis, DWRITE_TEXTURE_CLEARTYPE_3x1, &bounds);
	if (FAILED(hr)) {
		IDWriteGlyphRunAnalysis_Release(analysis);
		IDWriteFontFace_Release(face);
		return -9;
	}
	int w = bounds.right - bounds.left;
	int h = bounds.bottom - bounds.top;
	if (w <= 0 || h <= 0) {
		IDWriteGlyphRunAnalysis_Release(analysis);
		IDWriteFontFace_Release(face);
		return 0;
	}
	UINT32 nbytes = (UINT32)(w * h * 3);
	uint8_t *buf = (uint8_t *)malloc(nbytes);
	if (!buf) {
		IDWriteGlyphRunAnalysis_Release(analysis);
		IDWriteFontFace_Release(face);
		return -10;
	}
	hr = IDWriteGlyphRunAnalysis_CreateAlphaTexture(
		analysis, DWRITE_TEXTURE_CLEARTYPE_3x1, &bounds, buf, nbytes);
	IDWriteGlyphRunAnalysis_Release(analysis);
	IDWriteFontFace_Release(face);
	if (FAILED(hr)) {
		free(buf);
		return -11;
	}
	*out_rgb = buf;
	*out_w = w;
	*out_h = h;
	*bearing_x = bounds.left;
	*bearing_y = bounds.top;
	return 1;
}

/* Shared face+metrics setup for LCD and greyscale rasters. */
static int dwrite_prepare_run(
	const uint16_t *font_path_utf16,
	int path_len,
	float em_size_px,
	uint32_t codepoint,
	IDWriteFontFace **out_face,
	UINT16 *out_glyph,
	FLOAT *out_advance,
	DWRITE_GLYPH_RUN *out_run,
	DWRITE_GLYPH_OFFSET *out_offset)
{
	(void)path_len;
	*out_face = NULL;
	*out_glyph = 0;
	*out_advance = 0;
	if (!ensure_factory()) {
		return -2;
	}
	IDWriteFontFile *font_file = NULL;
	HRESULT hr = IDWriteFactory_CreateFontFileReference(
		g_factory, (const WCHAR *)font_path_utf16, NULL, &font_file);
	if (FAILED(hr) || !font_file) {
		return -3;
	}
	IDWriteFontFace *face = NULL;
	hr = IDWriteFactory_CreateFontFace(
		g_factory, DWRITE_FONT_FACE_TYPE_TRUETYPE, 1, &font_file, 0, DWRITE_FONT_SIMULATIONS_NONE, &face);
	IDWriteFontFile_Release(font_file);
	if (FAILED(hr) || !face) {
		return -4;
	}
	UINT32 cp = codepoint;
	UINT16 glyph_index = 0;
	hr = IDWriteFontFace_GetGlyphIndices(face, &cp, 1, &glyph_index);
	if (FAILED(hr) || glyph_index == 0) {
		IDWriteFontFace_Release(face);
		return -5;
	}
	DWRITE_GLYPH_METRICS metrics;
	hr = IDWriteFontFace_GetDesignGlyphMetrics(face, &glyph_index, 1, &metrics, FALSE);
	if (FAILED(hr)) {
		IDWriteFontFace_Release(face);
		return -6;
	}
	UINT16 units_per_em = 0;
	{
		DWRITE_FONT_METRICS fm;
		IDWriteFontFace_GetMetrics(face, &fm);
		units_per_em = fm.designUnitsPerEm;
	}
	if (units_per_em == 0) {
		IDWriteFontFace_Release(face);
		return -7;
	}
	float scale = em_size_px / (float)units_per_em;
	*out_advance = metrics.advanceWidth * scale;
	*out_glyph = glyph_index;
	*out_face = face;
	out_offset->advanceOffset = 0;
	out_offset->ascenderOffset = 0;
	memset(out_run, 0, sizeof(*out_run));
	out_run->fontFace = face;
	out_run->fontEmSize = em_size_px;
	out_run->glyphCount = 1;
	out_run->glyphIndices = out_glyph;
	out_run->glyphAdvances = out_advance;
	out_run->glyphOffsets = out_offset;
	out_run->isSideways = FALSE;
	out_run->bidiLevel = 0;
	return 1;
}

int gru_dwrite_grey_glyph(
	const uint16_t *font_path_utf16,
	int path_len,
	float em_size_px,
	uint32_t codepoint,
	uint8_t **out_a,
	int *out_w,
	int *out_h,
	int *bearing_x,
	int *bearing_y,
	float *advance)
{
	*out_a = NULL;
	*out_w = *out_h = 0;
	*bearing_x = *bearing_y = 0;
	*advance = 0;
	if (!font_path_utf16 || path_len <= 0 || em_size_px < 6.0f || !out_a) {
		return -1;
	}

	IDWriteFontFace *face = NULL;
	UINT16 glyph_index = 0;
	FLOAT adv = 0;
	DWRITE_GLYPH_OFFSET offset;
	DWRITE_GLYPH_RUN run;
	int prep = dwrite_prepare_run(
		font_path_utf16, path_len, em_size_px, codepoint,
		&face, &glyph_index, &adv, &run, &offset);
	if (prep < 0) {
		return prep;
	}
	*advance = adv;

	/* Size-band: small UI → NATURAL greyscale AA (smooth curves). Larger →
	 * CLEARTYPE_NATURAL then luma fold (hinted stems without LCD draw).
	 * CT→A8 at ≤20px is why Font Faces looked stair-stepped when smaller. */
	DWRITE_RENDERING_MODE mode = DWRITE_RENDERING_MODE_CLEARTYPE_NATURAL;
	int fold_avg = 0;
	if (em_size_px <= 24.0f) {
		mode = DWRITE_RENDERING_MODE_NATURAL;
		fold_avg = 1;
	}
	IDWriteGlyphRunAnalysis *analysis = NULL;
	HRESULT hr = IDWriteFactory_CreateGlyphRunAnalysis(
		g_factory,
		&run,
		1.0f,
		NULL,
		mode,
		DWRITE_MEASURING_MODE_NATURAL,
		0.0f,
		0.0f,
		&analysis);
	if (FAILED(hr) || !analysis) {
		IDWriteFontFace_Release(face);
		return -8;
	}

	RECT bounds;
	hr = IDWriteGlyphRunAnalysis_GetAlphaTextureBounds(
		analysis, DWRITE_TEXTURE_CLEARTYPE_3x1, &bounds);
	if (FAILED(hr)) {
		IDWriteGlyphRunAnalysis_Release(analysis);
		IDWriteFontFace_Release(face);
		return -9;
	}
	int w = bounds.right - bounds.left;
	int h = bounds.bottom - bounds.top;
	if (w <= 0 || h <= 0) {
		IDWriteGlyphRunAnalysis_Release(analysis);
		IDWriteFontFace_Release(face);
		return 0;
	}
	UINT32 nbytes = (UINT32)(w * h * 3);
	uint8_t *rgb = (uint8_t *)malloc(nbytes);
	if (!rgb) {
		IDWriteGlyphRunAnalysis_Release(analysis);
		IDWriteFontFace_Release(face);
		return -10;
	}
	hr = IDWriteGlyphRunAnalysis_CreateAlphaTexture(
		analysis, DWRITE_TEXTURE_CLEARTYPE_3x1, &bounds, rgb, nbytes);
	IDWriteGlyphRunAnalysis_Release(analysis);
	IDWriteFontFace_Release(face);
	if (FAILED(hr)) {
		free(rgb);
		return -11;
	}

	uint8_t *a8 = (uint8_t *)malloc((size_t)w * (size_t)h);
	if (!a8) {
		free(rgb);
		return -10;
	}
	int i, n = w * h;
	for (i = 0; i < n; i++) {
		unsigned r = rgb[i * 3 + 0];
		unsigned g = rgb[i * 3 + 1];
		unsigned b = rgb[i * 3 + 2];
		if (fold_avg) {
			a8[i] = (uint8_t)((r + g + b) / 3);
		} else {
			a8[i] = (uint8_t)((r + 2 * g + b) / 4);
		}
	}
	free(rgb);
	*out_a = a8;
	*out_w = w;
	*out_h = h;
	*bearing_x = bounds.left;
	*bearing_y = bounds.top;
	return 1;
}

#else
#include "dwrite_lcd.h"
void gru_dwrite_free(void *p) { (void)p; }
int gru_dwrite_lcd_available(void) { return 0; }
int gru_dwrite_lcd_blend_params(float *gamma, float *enhanced_contrast, float *cleartype_level) {
	(void)gamma; (void)enhanced_contrast; (void)cleartype_level;
	return 0;
}
int gru_dwrite_lcd_glyph(
	const uint16_t *font_path_utf16, int path_len, float em_size_px, uint32_t codepoint,
	float baseline_origin_x, int render_mode,
	uint8_t **out_rgb, int *out_w, int *out_h, int *bearing_x, int *bearing_y, float *advance)
{
	(void)font_path_utf16; (void)path_len; (void)em_size_px; (void)codepoint;
	(void)baseline_origin_x; (void)render_mode;
	(void)out_rgb; (void)out_w; (void)out_h; (void)bearing_x; (void)bearing_y; (void)advance;
	return -100;
}
int gru_dwrite_grey_glyph(
	const uint16_t *font_path_utf16, int path_len, float em_size_px, uint32_t codepoint,
	uint8_t **out_a, int *out_w, int *out_h, int *bearing_x, int *bearing_y, float *advance)
{
	(void)font_path_utf16; (void)path_len; (void)em_size_px; (void)codepoint;
	(void)out_a; (void)out_w; (void)out_h; (void)bearing_x; (void)bearing_y; (void)advance;
	return -100;
}
#endif
