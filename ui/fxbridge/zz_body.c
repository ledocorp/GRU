//go:build cgo && !gru_nofx

/* Phase 0 — cgo compiles emit-C probe body from third_party pack (via -I gen). */
#include "emit_go_probe.c"
