#ifndef ZSPEC_CORE_ERROR_H
#define ZSPEC_CORE_ERROR_H

typedef enum core_Err {
    CORE_OK = 0,
    CORE_ERR_OUT_OF_MEMORY,
    CORE_ERR_INVALID_ARG,
    CORE_ERR_OVERFLOW,
    CORE_ERR_UNSUPPORTED,
} core_Err;

static inline int core_err_is_ok(core_Err err)
{
    return err == CORE_OK;
}

static inline int core_err_is_err(core_Err err)
{
    return err != CORE_OK;
}

#define CORE_IGNORE_ERR(expr) \
    do {                      \
        (void)(expr);         \
    } while (0)

#endif /* ZSPEC_CORE_ERROR_H */
