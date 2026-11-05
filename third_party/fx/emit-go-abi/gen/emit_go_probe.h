#ifndef FX_EMIT_GO_PROBE_H
#define FX_EMIT_GO_PROBE_H

#include <stdint.h>

#include "fx_vec.h"

/* fx: fn add(a: i32, b: i32) -> i32 */
int32_t fx_emit_go_probe_add(int32_t a, int32_t b);

/* fx: fn clamp(v: i32, lo: i32, hi: i32) -> i32 */
int32_t fx_emit_go_probe_clamp(int32_t v, int32_t lo, int32_t hi);

/* fx: fn hadd4(a0: i32, a1: i32, a2: i32, a3: i32, b0: i32, b1: i32, b2: i32, b3: i32) -> i32 */
int32_t fx_emit_go_probe_hadd4(int32_t a0, int32_t a1, int32_t a2, int32_t a3, int32_t b0, int32_t b1, int32_t b2, int32_t b3);

/* fx: fn hadd4f(a0: f32, a1: f32, a2: f32, a3: f32, b0: f32, b1: f32, b2: f32, b3: f32) -> f32 */
float fx_emit_go_probe_hadd4f(float a0, float a1, float a2, float a3, float b0, float b1, float b2, float b3);

/* fx: fn sum(s: &[i32]) -> i32 */
int32_t fx_emit_go_probe_sum(fx_Slice_i32 s);


#endif /* FX_EMIT_GO_PROBE_H */
