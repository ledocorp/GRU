# v0.9.0 — S1 ship default (1× present + sticky text)

Public toolkit tip sync: GitHub `main` now matches the living Gru present/text path.

## Highlights

- **Ship default is SSAA 1×** framebuffer present (not 2× downscale). Quality comes from MSAA, sticky ClearType/LCD Latin strikes, high-res glyph raster, and coverage AA for shapes.
- Optional `GRU_SSAA=1.5|2` / `GRU_RENDER_QUALITY=…` remain for testing.
- **fx emit-C** glyph/wrap kernels ship in the public tree (`gen/fxlib`, `third_party/fx/emit-go-abi`) for default CGO builds; opt out with `-tags gru_nofx` or `GRU_FX_*=0`.
- Docs updated for the 1× present story; public dogfood host remains **`grudemo`** (full Studio / Foundry / Prism stay private).
- Export hygiene: Foundry/Prism blank-imports excluded; EqualFill grid row-slice bug fixed.

## Module

```bash
go get github.com/ledocorp/gru@v0.9.0
go run ./cmd/hello
go run -tags grudemo .
```

## Notes

- Gru Notepad showcase will pin this tag in a follow-up cut.
- Linux / Android experience branches should rebase onto this tip before resubmit.
