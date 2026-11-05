# emit-go-abi consumer pack (FX-EMIT-GO-ABI-2â€¦5)

Staged for external Go/cgo (Gru `ui/fxbridge`). **Go links emit-C only — never fx IR.**

## Layout

- `gen/` â€” `emit_go_probe.h` / `.c` / `.fxmap` (from `fx emit-c --debug-source`)
- `include/zspec/` â€” zspec headers
- `lib/libzspec.a` â€” link when the library allocates (probe is POD-only but pack is complete)
- `VERSION` â€” fx package version used to generate

## Go sketch

```go
package fxprobe

/*
#cgo CFLAGS: -I${SRCDIR}/gen -I${SRCDIR}/include
#cgo LDFLAGS: -L${SRCDIR}/lib -lzspec
#include "emit_go_probe.h"
*/
import "C"
import "unsafe"

func Add(a, b int32) int32 {
    return int32(C.fx_emit_go_probe_add(C.int32_t(a), C.int32_t(b)))
}

// Hadd4 is POD-only (eight int32 in, one out) â€” Go never sees fx_v4i32.
func Hadd4(a0, a1, a2, a3, b0, b1, b2, b3 int32) int32 {
    return int32(C.fx_emit_go_probe_hadd4(
        C.int32_t(a0), C.int32_t(a1), C.int32_t(a2), C.int32_t(a3),
        C.int32_t(b0), C.int32_t(b1), C.int32_t(b2), C.int32_t(b3),
    ))
}

// Hadd4f is POD-only (eight float32 in, one out) â€” Go never sees fx_v4f32.
func Hadd4f(a0, a1, a2, a3, b0, b1, b2, b3 float32) float32 {
    return float32(C.fx_emit_go_probe_hadd4f(
        C.float(a0), C.float(a1), C.float(a2), C.float(a3),
        C.float(b0), C.float(b1), C.float(b2), C.float(b3),
    ))
}

// Sum is buffer POD â€” Go builds fx_Slice_i32 { data, len } from []int32.
func Sum(xs []int32) int32 {
    var s C.fx_Slice_i32
    if len(xs) > 0 {
        s.data = (*C.int32_t)(unsafe.Pointer(&xs[0]))
        s.len = C.size_t(len(xs))
    }
    return int32(C.fx_emit_go_probe_sum(s))
}
```

Compile the generated `.c` into the Go package (cgo picks up sibling `.c` files, or `#include` the body as in fx `bindings/go/emitlib`).

## Debug

Open `gen/emit_go_probe.fxmap` or rebuild with `--debug-source` to map C lines back to fx.

## Regenerate

From fx repo root:

```powershell
.\scripts\package-emit-go-abi.ps1
.\scripts\test-emit-go-abi.ps1
```

Authority: `docs/specs/EMIT_C_GO_ABI_LITE.md`.
