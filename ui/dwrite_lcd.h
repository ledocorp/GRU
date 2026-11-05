#ifndef GRU_DWRITE_LCD_H
#define GRU_DWRITE_LCD_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Returns 1 if DirectWrite factory is usable. */
int gru_dwrite_lcd_available(void);

/* Default ClearType blend knobs (gamma, enhancedContrast, clearTypeLevel). Returns 1 on success. */
int gru_dwrite_lcd_blend_params(float *gamma, float *enhanced_contrast, float *cleartype_level);

/* Raster one codepoint to ClearType_3x1 RGB coverage (3 bytes/pixel).
 * baseline_origin_x is fractional pixel phase (0 / 0.25 / 0.5 / 0.75 for rem-body).
 * render_mode: 0 = CLEARTYPE_NATURAL (chrome), 1 = CLEARTYPE_NATURAL_SYMMETRIC
 *   (softer curves — aux Inter/Fira display sizes toward browser look).
 * Returns 1 on ink, 0 on empty (advance may still be set), <0 on error.
 * Caller frees *out_rgb with gru_dwrite_free. */
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
	float *advance);

/* Greyscale coverage from CLEARTYPE_NATURAL (luma R+2G+B → A8).
 * Same hinting as Segoe chrome without LCD RGB draw — for Inter/Fira sticky. */
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
	float *advance);

void gru_dwrite_free(void *p);

#ifdef __cplusplus
}
#endif

#endif
