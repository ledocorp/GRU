# Rendering

Default path: **1× framebuffer** presentation with **4× MSAA**, **ClearType-style LCD text**, **high-resolution glyph raster**, and **1× shape anti-aliasing**. Main and overlay render targets are allocated eagerly; idle frames blit the last cached frame when nothing changed.

Optional environment variables can raise the framebuffer scale (1.5× or 2×) for testing or comparison — they are not required for normal use.

---

## Quality baseline

| Piece | Role |
|-------|------|
| `FlagMsaa4xHint` | Hardware MSAA on the window |
| `InitSupersampling` | Main + overlay render targets; **1× scale by default** (no env vars) |
| `InitDisplayAwareAtlases` | Icons / UI faces at display scale (not multiplied by SSAA) |
| `RenderScale` | Camera zoom + scissor coords inside the render target |
| Glyph supersampling | Text and icons raster on the CPU at higher resolution when the scene framebuffer is 1× |

Widget authors draw Raylib primitives but should use Gru text helpers (`DrawText` / measure / effective font size) — not raw Raylib text APIs — so scale stays consistent.

At 1×, quality comes from MSAA, LCD blending on opaque surfaces, coverage-aware strokes, and glyph raster — not from bilinear downscaling a 2× framebuffer.

---

## Optional framebuffer scale

Set before or during `InitSupersampling` (the helper reads env on init):

| Variable | Values | Effect |
|----------|--------|--------|
| `GRU_SSAA` | `1`, `1.5`, `2` | Explicit framebuffer multiplier |
| `GRU_RENDER_QUALITY` | `max`, `balanced`, `lean`, `minimal` | Preset tiers (`max` → 2×; `minimal` → true 1×) |

With **no variables set**, Gru keeps the **1× native default** — the recommended path for shipped apps.

---

## Present pipeline

```text
InitSupersampling(w, h)   // default 1×; main + overlay targets
…
if NeedsRedraw():
  BeginSuperFrame(clear) → Root.Draw → EndSuperFrame
else:
  cache hit — keep previous frame
BeginDrawing → BlitToScreen* → EndDrawing
```

**Eager overlay RT:** both targets allocate at init. Deferring the overlay target until first use left modals/menus soft or missing after rescale. `EnsureOverlayTarget` is a safety net, not the primary path.

On client resize: recreate textures (`ResizeWindowTextures` / rescale supersampling) — never stretch an old RT into a new client size (fat glyphs / missing columns).

The main pass paints an **opaque page fill** before scene content so LCD text and thin strokes composite correctly at 1×.

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

- Init both render targets at startup; unload on shutdown  
- Resize textures on every client size change before blit  
- Use wake reasons; prefer overlays for hover  
- Paint opaque page fill in the main pass at 1×  

**Don’t**

- Defer overlay target allocation as a footprint trick  
- Ad-hoc sleeps / per-widget FPS throttles  
- Enable global `WakeOnMouseMove` casually  
- Fix scroll cost with post-blit viewport hacks  
- Assume 2× framebuffer downscale is required for crisp text  

---

## Samples

- `cmd/hello` — `InitSupersampling` + cache loop  
- grudemo: **Batch 1 · Tooltip / TabView / Modal**, **Batch 7 · Toast / Notification**, **Counter Demo**, **Theme v2 Foundation**  

Next: [webview.md](webview.md) · [overview.md](overview.md)
