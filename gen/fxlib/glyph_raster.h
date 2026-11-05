#ifndef FX_GLYPH_RASTER_H
#define FX_GLYPH_RASTER_H

#include <stdint.h>

#include "fx_vec.h"

/* fx: effects: mut */
/* fx: fn clear_rgba(rgba: &mut [u8]) -> i32 */
int32_t fx_glyph_raster_clear_rgba(fx_MutSlice_u8 rgba);

/* fx: fn px_x(x: f32, scale: f32, ox: f32) -> f32 */
float fx_glyph_raster_px_x(float x, float scale, float ox);

/* fx: fn px_y(y: f32, scale: f32, oy: f32) -> f32 */
float fx_glyph_raster_px_y(float y, float scale, float oy);

/* fx: effects: mut */
/* fx: fn push_edge(edges: &mut [f32], n: i32, x0: f32, y0: f32, x1: f32, y1: f32) -> i32 */
int32_t fx_glyph_raster_push_edge(fx_MutSlice_f32 edges, int32_t n, float x0, float y0, float x1, float y1);

/* fx: fn quad_eval(p0: f32, p1: f32, p2: f32, t: f32) -> f32 */
float fx_glyph_raster_quad_eval(float p0, float p1, float p2, float t);

/* fx: fn cube_eval(p0: f32, p1: f32, p2: f32, p3: f32, t: f32) -> f32 */
float fx_glyph_raster_cube_eval(float p0, float p1, float p2, float p3, float t);

/* fx: effects: mut */
/* fx: fn flatten_outline(ops: &[i32], pts: &[f32], scale: f32, ox: f32, oy: f32, edges: &mut [f32]) -> i32 */
int32_t fx_glyph_raster_flatten_outline(fx_Slice_i32 ops, fx_Slice_f32 pts, float scale, float ox, float oy, fx_MutSlice_f32 edges);

/* fx: effects: mut */
/* fx: fn add_cover(rgba: &mut [u8], w: i32, h: i32, x: i32, y: i32, add: i32) -> i32 */
int32_t fx_glyph_raster_add_cover(fx_MutSlice_u8 rgba, int32_t w, int32_t h, int32_t x, int32_t y, int32_t add);

/* fx: effects: mut */
/* fx: fn fill_span_aa(rgba: &mut [u8], w: i32, h: i32, y: i32, x0f: f32, x1f: f32, weight: i32) -> i32 */
int32_t fx_glyph_raster_fill_span_aa(fx_MutSlice_u8 rgba, int32_t w, int32_t h, int32_t y, float x0f, float x1f, int32_t weight);

/* fx: effects: mut */
/* fx: fn fill_nonzero(edges: &mut [f32], edge_n: i32, w: i32, h: i32, xs: &mut [f32], dirs: &mut [i32], rgba: &mut [u8]) -> i32 */
int32_t fx_glyph_raster_fill_nonzero(fx_MutSlice_f32 edges, int32_t edge_n, int32_t w, int32_t h, fx_MutSlice_f32 xs, fx_MutSlice_i32 dirs, fx_MutSlice_u8 rgba);

/* fx: effects: mut */
/* fx: fn raster_outline(ops: &[i32], pts: &[f32], w: i32, h: i32, scale: f32, ox: f32, oy: f32, edges: &mut [f32], xs: &mut [f32], dirs: &mut [i32], rgba: &mut [u8]) -> i32 */
int32_t fx_glyph_raster_raster_outline(fx_Slice_i32 ops, fx_Slice_f32 pts, int32_t w, int32_t h, float scale, float ox, float oy, fx_MutSlice_f32 edges, fx_MutSlice_f32 xs, fx_MutSlice_i32 dirs, fx_MutSlice_u8 rgba);

/* fx: effects: mut */
/* fx: fn boost_rgba_alpha(rgba: &mut [u8], contrast: f32) -> i32 */
int32_t fx_glyph_raster_boost_rgba_alpha(fx_MutSlice_u8 rgba, float contrast);


#endif /* FX_GLYPH_RASTER_H */
