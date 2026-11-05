#ifndef ZSPEC_CORE_ALLOCATOR_H
#define ZSPEC_CORE_ALLOCATOR_H

#include <stddef.h>

#include "zspec/core_error.h"

typedef struct core_Allocator core_Allocator;

core_Allocator* core_allocator_new(void);
void core_allocator_destroy(core_Allocator* allocator);
core_Allocator* core_default_allocator(void);
core_Allocator* core_arena_new(core_Allocator* parent, size_t initial_size);

void* core_mem_alloc(core_Allocator* allocator, size_t size);
void* core_mem_calloc(core_Allocator* allocator, size_t count, size_t size);
void* core_mem_realloc(core_Allocator* allocator, void* pointer, size_t new_size);
void  core_mem_free(core_Allocator* allocator, void* pointer);
void  core_mem_reset(core_Allocator* allocator);

size_t core_mem_stats_live_bytes(const core_Allocator* allocator);

/*
 * Test-harness fault injection (FX-ASR-B2 / Hipp N-th alloc).
 * Default off. Not required for normal product paths.
 * `core_mem_fault_fail_nth(n)` — the n-th `core_mem_alloc` call returns NULL (1-based).
 * `n == 0` clears. Affects calloc (via alloc); does not rewrite production call sites.
 */
void core_mem_fault_clear(void);
void core_mem_fault_fail_nth(size_t n);
size_t core_mem_fault_alloc_count(void);

#endif /* ZSPEC_CORE_ALLOCATOR_H */
