#ifndef FX_VEC_TYPES_H
#define FX_VEC_TYPES_H

#include <stddef.h>
#include <stdint.h>

typedef struct {
    const int32_t* data;
    size_t len;
} fx_Slice_i32;

typedef struct {
    int32_t v[4];
} fx_v4i32;

typedef struct {
    float v[4];
} fx_v4f32;

#endif /* FX_VEC_TYPES_H */
