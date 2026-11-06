# Linux present — X11 + Wayland (general)

Not Cosmic-specific. Same contract on GNOME, KDE, Sway, Cosmic, and X11.

## Policy

**Window size = client size. No DPI expand / Camera densify / FB-as-layout.**

| Piece | Linux (X11 + Wayland) | Windows |
|-------|----------------------|---------|
| Client size | `GetScreenWidth/Height` only | same (already = FB) |
| `SCALE_FRAMEBUFFER` | forced **off** after raylib reset | N/A |
| `DisplayScale` | **1** (ignore compositor DPI for rem) | follows monitor DPI |
| SSAA / Camera | quality tier **1×** only | 1× lean |
| MSAA | **off** on S1 hosts | on |
| Text | sticky greyscale via **outline→fx AA** (gg fallback); no gamma / no `BoostRGBAAlpha` | DWrite ClearType |

Full text journey (scale → holes → outline AA → soft-bold → kern → snap fights → baseline lock), failed experiments, code map, and **Windows bold verification note for the linux-experience PR**: **[docs/architecture/linux-s1-text.md](docs/architecture/linux-s1-text.md)**.

## Why ignore compositor scale

Cursor/Brave *honor* fractional scale with a denser buffer (same window size, sharp).  
Gru’s Linux S1 choice is simpler **1:1**: do not zoom or enlarge for DPI. If a compositor still upscales a 1× buffer, that is outside the app — we do not keep adding Cosmic-only compensation layers.

## Opt into X11 explicitly

```bash
GLFW_PLATFORM=x11 GRU_DEBUG_CLIENT=1 /tmp/grunotepad-s1
```

`PrepareLinuxWindowHints` applies the GLFW platform hint when that env is set.
