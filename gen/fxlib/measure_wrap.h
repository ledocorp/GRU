#ifndef FX_MEASURE_WRAP_H
#define FX_MEASURE_WRAP_H

#include <stdint.h>

#include "fx_vec.h"

/* fx: effects: mut */
/* fx: fn wrap_advances(advances: &[f32], can_break: &[i32], max_w: f32, out_ends: &mut [i32]) -> i32 */
int32_t fx_measure_wrap_wrap_advances(fx_Slice_f32 advances, fx_Slice_i32 can_break, float max_w, fx_MutSlice_i32 out_ends);

/* fx: fn sum_advances(advances: &[f32], lo: i32, hi: i32) -> f32 */
float fx_measure_wrap_sum_advances(fx_Slice_f32 advances, int32_t lo, int32_t hi);


#endif /* FX_MEASURE_WRAP_H */
