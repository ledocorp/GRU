# Android system-bar insets (status / nav / cutout)

**Status:** **Verified** — M2 App Shell sign-off 2026-11-04 (`dist/android/grudemo-appshell-success.png`)  
**Last updated:** 2026-11-04

How Gru clears the black status band and gesture-bar clip on NativeActivity, and why. Path below shipped and matches the signed-off AVD capture (gray status band, BottomNav above the gesture pill, FAB unclipped).

---

## Symptom

On Pixel-class AVDs (`mAppBounds` top ≈ 128px @ 420dpi):

- Black slab above AppBar / hello content (OS status bar over a surface that does not own that band)
- BottomNav / FAB clipped by the gesture inset
- Java `setStatusBarColor` alone lost to **raylib** `AWINDOW_FLAG_FULLSCREEN` (set in `rcore_android.c` at `InitPlatform`)

---

## Peer survey (what others do)

| System | Method | Java required? |
|--------|--------|----------------|
| **Fyne** | `GoNativeActivity` reads `WindowInsets` → JNI → Go `size.Event` inset fields → pad layout | **Yes** (thin Activity) |
| **Ebiten** | Safe area still open (`Monitor.SafeArea`) | Not solved |
| **GameActivity (AGDK)** | `GameActivity_getWindowInsets` in C; Google’s NativeActivity successor | Activity class + native API |
| **NDK NativeActivity** | `AWINDOW_FLAG_LAYOUT_IN_SCREEN \| LAYOUT_INSET_DECOR` + `onContentRectChanged` → `ARect` | **No** for the rect |
| **raylib** | Forces `AWINDOW_FLAG_FULLSCREEN` (game-first; ignore chrome) | No inset story |
| **fxlang** | Does **not** own GLES/insets — host (GRU) does ([ANDROID_FX.md](ANDROID_FX.md)) | N/A |

**Did Gru need Java for paint?** No — hello/grudemo already paint on stock NativeActivity.  
**For polished insets/colors?** Fyne-style Java is the proven UI-toolkit path; NDK content-rect is enough for *padding* without Java. Status-bar *tint* can stay a thin `MainActivity` helper.

---

## Chosen path (M2)

Stay on raylib NativeActivity. After `rl.InitWindow`:

1. **Clear** `AWINDOW_FLAG_FULLSCREEN`
2. **Set** `FORCE_NOT_FULLSCREEN | LAYOUT_IN_SCREEN | LAYOUT_INSET_DECOR`
3. Install `onContentRectChanged` → cache `ARect`
4. Map rect → `Document.SetChromeTop` / `SetChromeBottom` (gesture pad under BottomNav)

Code: `ui/android_insets_android.go` · stubs: `ui/android_insets_stub.go`  
Call sites: `main.go` (grudemo), `cmd/hello/main.go`  
Build: `-I$ANDROID_NDK_HOME/sources/android/native_app_glue` in `scripts/android/build-*.sh`

Fallback when the WM has not published a rect yet: ≈48dp top / ≈24dp bottom @ 420dpi (126 / 63 px).

Optional Java (`MainActivity` + `GruDemoTheme`): light `statusBarColor` matching Gru clear `(228,229,235)` so icons sit on gray, not black — does not replace the chrome inset math.

---

## Explicitly rejected (for now)

| Approach | Why |
|----------|-----|
| `IMMERSIVE_STICKY` | Blanked GLES behind OS “Viewing full screen” tip |
| `Theme.*.Light` on NativeActivity | Black screencap / failed composite on API 34 AVD |
| `SetChromeTop(128)` *without* clearing raylib fullscreen | Double-padded (surface already below status) and shoved BottomNav off-screen |
| GameActivity migration | Best long-term insets/IME, but leaves raylib’s NativeActivity glue — later |

---

## Later options

1. **Fyne-style JNI insets** if content-rect proves flaky across OEMs  
2. **GameActivity** when IME / SAF need Jetpack (M4+)  
3. **FX** stays caps/logic only — no compositor ([ANDROID_FX.md](ANDROID_FX.md))

---

## Verify

```bash
source scripts/android/env.sh
./scripts/android/assemble-grudemo.sh x86_64
adb install -r dist/android/grudemo-debug.apk
adb shell am start -n org.ledocorp.gru.demo/.MainActivity
# Wait ~25s for App Shell atlas warm; screencap should show gray (not black) top band
# and full BottomNav labels above the gesture pill.
```

Pass criteria (met 2026-11-04): AppBar + body + FAB + BottomNav icons/labels; no black status slab; no gesture clip. Phase log: [ANDROID_PHASE.md](ANDROID_PHASE.md#m2-success--app-shell-go-on-avd).
