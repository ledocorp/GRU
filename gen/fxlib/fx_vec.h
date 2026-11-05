#ifndef FX_VEC_TYPES_H
#define FX_VEC_TYPES_H

#include <stddef.h>
#include <stdint.h>

typedef struct {
    const float* data;
    size_t len;
} fx_Slice_f32;

typedef struct {
    const int32_t* data;
    size_t len;
} fx_Slice_i32;

typedef struct {
    float* data;
    size_t len;
} fx_MutSlice_f32;

typedef struct {
    int32_t* data;
    size_t len;
} fx_MutSlice_i32;

typedef struct {
    uint8_t* data;
    size_t len;
} fx_MutSlice_u8;

#endif /* FX_VEC_TYPES_H */