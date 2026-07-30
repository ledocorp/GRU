# Rendering

Quality path: **MSAA + 2× SSAA**, SDF fonts/atlases, eager main **and** overlay render targets, then a wake/idle FPS policy. Footprint trimming must not defer overlay SSAA or invent per-widget FPS sleeps.

---

## Quality baseline

| Piece | Role |
|-------|------|
| `FlagMsaa4xHint` | Hardware MSAA |
| `InitSupersampling` at **2.0×** | SSAA render targets |
| `InitDisplayAwareAtlases` | Icons / UI faces at display scale |
| `RenderScale` | Camera zoom + scissor coords in the superframe |

Widget authors draw Raylib primitives but should use Gru text helpers (`DrawText` / measure / effective font size) — not raw Raylib text APIs — so scale stays consistent.

There is no public “1× performance mode”; it was removed after quality regressions.

---

## SSAA pipeline

```text
InitSupersampling(w, h)   // main + overlay targets at 2×, FilterBilinear
…
if NeedsRedraw():
  BeginSuperFrame(clear) → Root.Draw → EndSuperFrame
else:
  cache hit — keep previous supersample
BeginDrawing → BlitToScreen* → EndDrawing
```

**Eager overlay RT:** both targets allocate at init. Deferring the overlay target until first use left modals/menus soft or missing after rescale. `EnsureOverlayTarget` is a safety net, not the primary path.

On client resize: recreate textures (`ResizeWindowTextures` / rescale supersampling) — never stretch an old RT into a new client size (fat glyphs / missing columns).

`ApplyUIOptimizations` disables depth test / backface culling / gestures as appropriate; audio is not initialized by default.

---

## Cache vs redraw

- Clean visible tree → blit-only (`NeedsRedraw` false); deep idle ~**10 FPS**
- Transient hover: prefer **interaction overlay** painters after blit — do **not** `MarkDrawDirty` for pure hover chrome
- Animation-only overlays can run at `AnimationFPS` (**36**) when the main cache is clean
- Still mark draw dirty for layout, value changes, selection, and open menus that bake into the cache

---

## Wake / idle policy

Wake reasons (bitflags) include input, scroll, keyboard, animation, overlay, data, resize, scene, WebView.

| Constant | Typical FPS | Role |
|----------|-------------|------|
| ActiveFPS | 60 | Interactive |
| AnimationFPS | 36 | Animation overlays only |
| ScrollFPS | 30 | Wheel + clean cache |
| WebViewIdleFPS | 12 | Keep WebView hosts alive |
| DeepIdleFPS | 10 | Deep idle |

Policy highlights:

- Drain wake signals before Update; let the idle policy own `SetTargetFPS`
- Global `WakeOnMouseMove` defaults **off**
- Exception: **`SampleChromeHoverWake`** — title band or live WebView client → ActiveFPS without enabling global mouse wake
- Resize hold tracker: stay Active through resize burst + cooldown (avoid 60→10 cliffs)
- Scene-load grace (~1.5s) and activity grace keep startup/snappy feel

---

## Overlay managers

Drawn after the main UI blit: modals, drawers, bottom sheets, context menus, tooltips, toasts, etc.

- Call `SetOverlayChromeInsets` so overlays respect the title band
- Modal / drawer / bottom sheet → `WebViewHostOccluded` hides the WebView HWND (not passthrough)

See [../widgets/overlays.md](../widgets/overlays.md).

---

## Do / Don’t

**Do**

- Init both SSAA targets at startup; unload on shutdown  
- Resize textures on every client size change before blit  
- Use wake reasons; prefer overlays for hover  

**Don’t**

- Defer overlay SSAA as a lean footprint trick  
- Ad-hoc sleeps / per-widget FPS throttles  
- Enable global `WakeOnMouseMove` casually  
- Fix scroll cost with post-blit viewport hacks  

---

## Samples

- `cmd/hello` — InitSupersampling + cache loop  
- grudemo: **Batch 1 · Tooltip / TabView / Modal**, **Batch 7 · Toast / Notification**, **Counter Demo**, **Theme v2 Foundation**  

Next: [webview.md](webview.md) · [overview.md](overview.md)
