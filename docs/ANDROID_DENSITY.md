# Android density, chrome, and IME (Gru sync)

**Status:** Landed — density + IME chars + async SAF + JSON persistence  
**Last updated:** 2026-11-06 (v0.9 / Linux S1 cross-check)  
**Related:** [ANDROID_PHASE.md](ANDROID_PHASE.md) · [ANDROID_INSETS.md](ANDROID_INSETS.md) · [architecture/linux-s1-text.md](architecture/linux-s1-text.md)

### vs Linux S1 present policy

Linux desktop forces **`DisplayScale = 1`** (window size = client; no compositor densify).  
Android **keeps** width/dpi buckets for `DisplayScale` (T1.9 atlases) and **`Density` / `Dp`** for chrome. Copying Linux unity onto phones would undersize touch targets. Shared with Linux: SSAA 1×, MSAA off, sticky **outline→fx** greyscale text + Poppins Medium/SemiBold soft-bold.

---

## Problem

On ~420dpi AVDs (1080×2400), Notepad chrome felt oversized and typing failed:

- Layout is in **framebuffer pixels**, while Android UIs author in **dp**.
- Phone bumps **stacked**: `RootFontSize≈40` × `PhoneChromeScale` 1.5 × ribbon cell ×1.9 → billboard tool strip (~200px).
- Soft keyboard (**IME**) was never wired (M4) — focus alone cannot type on a phone.
- BottomNav hit Y lagged draw Y (`chromeTop` flex band) until a footer hit-band hack.

---

## Synthesis (pros kept / cons rejected)

| Keep (sync) | Reject |
|-------------|--------|
| Gru retained-mode tree + `Document` chrome as inset bus | Second layout engine / Compose port |
| **One** `Density` → `Dp(v)` for chrome & touch | Stacked `PhoneChromeScale` × ribbon multipliers |
| Material-sized **tokens** (tool strip ~52dp, bottom ~48dp, touch ≥48dp) | Scaling the desktop Office ribbon to “fill” the phone |
| Fyne-thin JNI/NDK for IME show/hide + metrics | Early GameActivity migration for density math |
| Flutter-style: keyboard height adjusts chrome/content | Haptics as a substitute for wrong hit targets |
| iOS: sparse haptics later, separate from layout | Per-widget `if Android` size forks in demos |

**Steady model**

```text
DisplayMetrics / AConfiguration densityDpi
        │
        ▼
  ui.Density (160dpi → 1.0)
        │
        ├─ ui.Dp(v)     → chrome, touch, BottomNav, tool strip
        ├─ rem / type   → typography only (no extra chrome multiply)
        └─ DisplayScale → glyph atlases only (SSAA stays 1× on GLES)

WindowInsets / content rect / IME height
        │
        ▼
  Document.ChromeTop / ChromeBottom  → content band (draw == hit)
```

---

## API (`ui`)

| Symbol | Role |
|--------|------|
| `RefreshDensity()` | Read device density (Android) or set 1 (desktop) |
| `Density()` | Scale where **160dpi = 1.0** (Android `densityDpi/160`) |
| `Dp(v)` | `v * Density()` — density-independent → framebuffer px |
| `PhoneChromeScale()` | **Bridge only** — `min(Density(), 2)` on Android; prefer `Dp` at new call sites |

Phone chrome tokens (Notepad / BottomNav):

| Token | dp | ~px @420dpi |
|-------|-----|-------------|
| Tool strip height | 48 | ~126 |
| Ribbon cell | 48×48 (fills strip) | ~126×126 |
| Ribbon caption | 10 | ~26 |
| BottomNav bar | 56 | ~147 |
| BottomNav icon | 24 | ~63 |
| BottomNav label | 11 | ~29 |
| Min touch | 48 | ~126 |

IME: `MainActivity` hosts a 1×1 `EditText` proxy; Go arms IME on user tap then `showGruIme` / `hideGruIme` (`adjustNothing` keeps GLES size stable); codepoints via JNI → `PushIMECodepoint` → `CharPressedOrIME`; IME height → `IMEInsetHeight` → `ChromeBottom`. Haptics: `hapticTick` on BottomNav change (`VIBRATE` permission).

Typography keeps the phone rem floor (~40) so body text stays ~16sp; **do not** multiply rem into chrome again.

---

## IME foothold

1. On `Document.SetFocus` to a text field → `ANativeActivity_showSoftInput`.
2. On blur / non-text focus → `hideSoftInput`.
3. IME height (when available) adds into `ChromeBottom` so the editor is not covered.
4. Full `InputConnection` / GameActivity remains M4+ if raylib key delivery is incomplete.

---

## Product chrome (hosts)

Hosts that force phone BottomNav on Android (e.g. out-of-tree Notepad) should keep:

- **BottomNav** for primary modes
- **Top strip** = tool buttons only (`HideRibbonTabs`); height from `Dp(PhoneToolStripHDp)`, not a scaled desktop ribbon
- Menubar out of the Android tree

---

## Verification (toolkit)

```bash
source scripts/android/env.sh
./scripts/android/assemble-grudemo.sh x86_64
adb install -r dist/android/grudemo-debug.apk
adb shell am start -n org.ledocorp.gru.demo/.MainActivity
```

Product hosts verify density/IME in their own packaging against this branch.
