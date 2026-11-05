#ifndef ZSPEC_CORE_DEBUG_H
#define ZSPEC_CORE_DEBUG_H

#include "zspec/core_platform.h"

void core_debug_init(void);
void core_debug_shutdown(void);
void core_debug_dump_leaks(void);

#if ZSPEC_DEBUG_BUILD
#define CORE_ASSERT(expr)                                                     \
    do {                                                                      \
        if (!(expr)) {                                                        \
            core_debug_assert_failed(#expr, __FILE__, __LINE__);              \
        }                                                                     \
    } while (0)
#else
#define CORE_ASSERT(expr) ((void)0)
#endif

ZSPEC_NORETURN void core_debug_assert_failed(const char* expr, const char* file, int line);

#endif /* ZSPEC_CORE_DEBUG_H */
