# Third-party notices (maintainer checklist)

Gru depends on third-party software. **Public credits** must appear in:

1. Root [`NOTICE`](../../NOTICE)  
2. Repository [`README.md`](../../README.md) credits table (export source: `staging/PUBLIC_README.md`)

**Not legal advice.** Verify license texts before each public release.

---

## Must credit (shipped / redistributed)

| Project | Role | Action |
|---------|------|--------|
| **Raylib** (via raylib-go) | Rendering / windowing | NOTICE + README |
| **Remix Icon** | Icon atlas | NOTICE + README |
| **Bundled fonts** (Inter, Poppins, Fira Code, DM Mono) | UI / mono | NOTICE + OFL files beside fonts |
| **Microsoft Edge WebView2** | Windows web host (when used) | Document runtime terms |
| **go.mod direct deps** used by public `ui/` / demos | Various | Audit with `go-licenses`; append to NOTICE |

Legacy API names may still say `Phosphor*`; the **runtime icon set is Remix Icon**.

---

## Before public GitHub push

- [x] Credits table in README / NOTICE aligned (2026-07-29 audit; re-check on tip sync)  
- [x] `go-licenses` on public demo tags — snapshot `staging/go-licenses-grudemo.csv`  
- [ ] Tip-sync: export has no Foundry/Prism/full-Studio blank-imports (guardExport)  
- [x] Font OFL texts present for bundled faces (Inter, Poppins, Fira Code, DM Mono)  

### Re-audit command

```bash
# PowerShell
$env:GOFLAGS = "-tags=grudemo,webview2"
go-licenses csv . --ignore github.com/ledocorp/gru 2>$null | Where-Object { $_ -match '^[a-zA-Z0-9.]' } | Sort-Object -Unique | Set-Content staging/go-licenses-grudemo.csv
```

Then refresh root `NOTICE` and `staging/PUBLIC_README.md` credits.

---

## Related

- [`Licensing.md`](Licensing.md) — GPLv3 summary  
- [`COMMERCIALIZATION_DEFERRED.md`](../COMMERCIALIZATION_DEFERRED.md) — commercial pause  

---

*May 2026*
