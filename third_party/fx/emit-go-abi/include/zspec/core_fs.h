#ifndef ZSPEC_CORE_FS_H
#define ZSPEC_CORE_FS_H

#include <stdint.h>

#include "zspec/core_allocator.h"
#include "zspec/core_error.h"

/* Z3 — minimal text file read/write (stdio-backed). */

core_Err core_fs_write_text(const char* path, const char* text);
core_Err core_fs_append_text(const char* path, const char* text);
core_Err core_fs_read_text(core_Allocator* allocator, const char* path, char** out);
core_Err core_fs_exists(const char* path, int32_t* out);
core_Err core_fs_remove(const char* path);
core_Err core_fs_rename(const char* old_path, const char* new_path);

/*
 * Test-harness I/O fault injection (FX-ASR-B2b / Hipp).
 * Default off. Counts fopen attempts in write/append/read/exists.
 * `n == 0` clears. Failed open returns CORE_ERR_UNSUPPORTED (exists included).
 */
void core_fs_fault_clear(void);
void core_fs_fault_fail_nth(size_t n);
size_t core_fs_fault_io_count(void);

#endif /* ZSPEC_CORE_FS_H */
