#ifndef ZSPEC_CORE_PLATFORM_H
#define ZSPEC_CORE_PLATFORM_H

#if defined(_WIN32)
#define ZSPEC_PLATFORM_WINDOWS 1
#else
#define ZSPEC_PLATFORM_WINDOWS 0
#endif

#if defined(ZSPEC_DEBUG) && ZSPEC_DEBUG
#define ZSPEC_DEBUG_BUILD 1
#else
#define ZSPEC_DEBUG_BUILD 0
#endif

#if defined(_MSC_VER)
#define ZSPEC_NORETURN __declspec(noreturn)
#else
#define ZSPEC_NORETURN __attribute__((noreturn))
#endif

ZSPEC_NORETURN void core_platform_abort(void);

#endif /* ZSPEC_CORE_PLATFORM_H */
