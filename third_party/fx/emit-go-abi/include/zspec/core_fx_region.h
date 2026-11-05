#ifndef ZSPEC_CORE_FX_REGION_H
#define ZSPEC_CORE_FX_REGION_H

#include "zspec/core_allocator.h"

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct core_Region core_Region;

typedef struct core_FxRegionDesc {
    const char* label;
    size_t initial_size;
} core_FxRegionDesc;

core_Region* core_fx_region_new(
    core_Allocator* backing,
    core_Region* parent,
    const core_FxRegionDesc* desc
);
void core_fx_region_reset(core_Region* region);
void core_fx_region_destroy(core_Region* region);

core_Region* core_fx_region_parent(const core_Region* region);
core_Allocator* core_fx_region_allocator(core_Region* region);
const char* core_fx_region_label(const core_Region* region);

#ifdef __cplusplus
}
#endif

#endif /* ZSPEC_CORE_FX_REGION_H */
