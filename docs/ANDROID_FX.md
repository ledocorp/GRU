# FX × Android (enhancement track)

**Status:** Unblocked — M2 App Shell signed off ([ANDROID_PHASE.md](ANDROID_PHASE.md)); insets path verified ([ANDROID_INSETS.md](ANDROID_INSETS.md)).  
**Rule:** FX does **not** own drawing. GRU owns theme, layout, and GLES paint. FX adds **logic, policy, and data** behind the host.

Aligned with fxlang WRAP/GAPS: host owns the event loop; guest/caps for authority; no widget-tree port to “prove” fx.

---

## Clean look (GRU, not FX)

Ship these in M1–M2 Go work:

1. **App Shell** — `MountAppShellRoot`, bottom nav / AppBar; avoid desktop ribbon on phone demos  
2. **Touch density** — rem / `RootFontSize`, 48px hit floors, Theme v2 comfortable variants  
3. **Readable GLES text** — bitmap path for hello; improve without FX  
4. **Scroll + tap** — reliable phone input (gestures currently off globally)

---

## Where FX helps (after APK works)

| Priority | Idea | Shape |
|----------|------|--------|
| High | Guest demo logic | Caps-shaped guest computes; Go binds into widgets (`showcase_wrap` pattern) |
| High | Theme / density as data | FX emits token tables → Go `TokenSet` / variants |
| High | SAF / files via caps | JNI/Go mints path allowlist → `FsCap`/`OutCap` |
| Medium | Offline sync / transforms | Guest pipelines; UI stays Go |
| Medium | Sandboxed on-device scripts | `fxguest`-style allow/deny/fuel |
| Medium | aarch64 emit-C guest `.a` in APK | Toolchain; IR aarch64 out-of-claim |
| Avoid | FX wrapping raylib as UI | Duplicates GRU; GAPS refuses widget-tree ports |
| Avoid | WASM-in-WebView as phone UI | Not the NativeActivity path |

---

## Milestone fit

| Phase | FX? |
|-------|-----|
| M0–M1 hello APK | **No** |
| M2 grudemo paint (App Shell signed off) | **No** (optional offline-authored token JSON later) |
| M3+ packaging / SAF / IME | **Yes** — caps + guest |
| Later | Token profiles from FX; sandboxed demo guests |

---

## References

- fxlang: `docs/WRAP.md`, `docs/COMPOSITION.md`, `docs/GAPS.md`, `scaffolds/guest/`  
- GRU: `docs/widgets/theme-v2.md`, `docs/widgets/shells.md`, [ANDROID_PHASE.md](ANDROID_PHASE.md)
