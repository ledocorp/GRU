#ifndef ZSPEC_CORE_STRING_H
#define ZSPEC_CORE_STRING_H

#include <stddef.h>

#include "zspec/core_allocator.h"
#include "zspec/core_error.h"

typedef struct core_str_view {
    const char* data;
    size_t len;
} core_str_view;

core_str_view core_str_view_from_cstr(const char* cstr);
core_str_view core_str_view_from_parts(const char* data, size_t len);

size_t core_str_view_len(core_str_view view);
int core_str_view_equals(core_str_view left, core_str_view right);
int core_str_view_equals_cstr(core_str_view left, const char* right);

core_Err core_str_dup(core_Allocator* allocator, core_str_view source, char** out);
core_Err core_str_dup_cstr(core_Allocator* allocator, const char* source, char** out);

#endif /* ZSPEC_CORE_STRING_H */
