# Android code map

How Gru targets Android (raylib **NativeActivity**). Phase status: [ANDROID_PHASE.md](ANDROID_PHASE.md). FX track: [ANDROID_FX.md](ANDROID_FX.md).

---

## 1. Mental model

```
NativeActivity (AndroidManifest)
        │
        ▼
raylib android_main (C)  ──calls──►  android_run (cgo export)
        │                                    │
        │                                    ▼
        │                         rl.SetCallbackFunc(runApp)   // init in *_android_init.go
        │                                    │
        ▼                                    ▼
   EGL / GLES2                          Go runApp() loop
   AAssetManager                        ui.Document + widgets
```

- **GOOS=android** also matches the `linux` build tag — desktop X11/gsettings files **must** use `linux && !android`.
- App code uses `runtime.GOOS == "android"` (helpers: `isAndroidApp()` in root, `isAndroid()` in `cmd/hello`).

---

## 2. Entry points

| Package | File | Role |
|---------|------|------|
| root (`grudemo`) | `main_android_init.go` | `init()` → `rl.SetCallbackFunc(runApp)` |
| root | `main.go` | `main()` skips `runApp` on Android; `runApp` owns the loop |
| `cmd/hello` | `cmd/hello/main_android_init.go` | Same pattern for hello APK |
| `cmd/hello` | `cmd/hello/main.go` | No chrome / no MSAA / no SSAA on Android; back key exits |

raylib-go **v0.55.1** uses `SetCallbackFunc` (newer upstream examples may show `SetMain` — same idea).

---

## 3. Platform stubs (Android)

| Concern | File | Behavior |
|---------|------|----------|
| Tray | `internal/tray/tray_android.go` | No-op |
| OpenBrowser | `ui/openurl_android.go` | ACTION_VIEW via `gruOpenBrowser` → Activity `openGruUrl` |
| Full-window WebView | `cmd/gruwebview` + `packaging/android/webview` | System WebView Activity + JNI (not raylib); [ANDROID_WEBVIEW.md](ANDROID_WEBVIEW.md) |
| File dialogs | `examples/filedialog_android.go` | SAF via `StartTextFileOpen` / `StartTextFileSave` (async; MainActivity intents) |
| Window chrome | `platform_other.go` (`android` tag) | No X11 raise / borderless DWM |
| Appearance | `ui/appearance_system.go` | No gsettings; light default |
| Single-instance | `examples/appinstance/instance_other.go` | No flock / X11 raise |

---

## 4. Desktop-only (excluded on Android)

Retagged to `//go:build linux && !android`:

- `platform_linux.go`
- `ui/appearance_system_linux.go`
- `examples/appinstance/instance_linux.go`
- `x11util/focus_linux.go` (CGO `-lX11`)

Stub side: `(!windows && !linux) || android` on `platform_other.go`, `ui/appearance_system.go`, `examples/appinstance/instance_other.go`.

---

## 5. Assets & GLES / S1 text (v0.9+)

| API | File | Notes |
|-----|------|-------|
| `rl.OpenAsset` | `ui/assetfs.go` | APK `assets/` tree |
| `AndroidGLES()` | `ui/assetfs.go` | True on Android |
| SDF fonts | `ui/font.go` | Skipped on GLES (GLSL 330) |
| Sticky text | `ui/text_ft_grey*.go` | **Same S1 path as Linux**: outline→fx AA greyscale strikes (not desktop-only). Bitmap/SDF are fallback only. |
| Soft-bold | `assets/fonts/poppins/` | Must include **Medium** + **SemiBold** (copied by `scripts/android/build-*.sh` from repo `assets/`). Without them chrome Bold falls back to 700. |

`scripts/android/build-grudemo.sh` (and hello) sync `assets/fonts|icons|web` into `packaging/android/<app>/assets/` at build time — keep Medium/SemiBold in the **repo** `assets/fonts/poppins/` tree (packaging copies are gitignored).

Present: SSAA **1×**, MSAA **off** (EGL). Density / `Dp` for chrome: [ANDROID_DENSITY.md](ANDROID_DENSITY.md). Do **not** copy Linux `DisplayScale=1` onto phones — Android keeps width/dpi buckets for T1.9 atlases while rem/chrome use `Density`.

---

## 6. Host Activity contract (JNI)

Go calls into the running Activity by method name (any package). Optional methods:

| Method | Role |
|--------|------|
| `showGruIme` / `hideGruIme` | Soft keyboard via EditText proxy |
| `hapticTick` | Short confirmation vibration |
| `openGruUrl(String)` | `ACTION_VIEW` browser handoff |
| `gruPickOpenDocument` / `gruPickSaveDocument(...)` | SAF open/save |

Java→Go natives (`nativeImeCodepoint`, `nativeImeKey`, `nativeImeHeight`, `nativeFileOpenResult`, `nativeFileSaveResult`) are `RegisterNatives`'d on the **live** Activity class the first time Go attaches (product-agnostic — no hardcoded package).

In-tree samples: `hello` / `grudemo` implement tint + `openGruUrl`. Full IME/SAF host example lives out of tree (e.g. gru-notepad).

## 7. Packaging layout

```
packaging/android/
  build.gradle          # AGP classpath
  settings.gradle       # :hello + :grudemo + :webview
  hello/                # NativeActivity sample
  grudemo/              # M2 App Shell sample
  webview/              # Full-window system WebView + Go JNI (no raylib)
```

Build `.so`:

```bash
./scripts/android/build-hello.sh           # arm64-v8a libhello.so
./scripts/android/build-hello.sh arm       # optional armeabi-v7a
```

APK (SDK + JDK required):

```bash
cd packaging/android && ./gradlew :hello:assembleDebug
```

---

## 8. Window / input policy (Android)

From `platform_app.go` + hello adaptations:

- No MSAA (`FlagMsaa4xHint` breaks many EGL configs)
- No custom title bar / borderless chrome
- Portrait normalize when width > height
- Hardware **Back** closes the app (v1)
- SSAA off (1× GLES draw)
- **System-bar insets** — after `InitWindow`, clear raylib `AWINDOW_FLAG_FULLSCREEN` and apply NDK layout+inset-decor flags; map `contentRect` → `Document` chrome (`ui.EnsureAndroidInsetFlags` / `ApplyAndroidSafeInsets`). **Verified** on App Shell AVD 2026-11-04. Details: [ANDROID_INSETS.md](ANDROID_INSETS.md) · sign-off: [ANDROID_PHASE.md](ANDROID_PHASE.md#m2-success--app-shell-go-on-avd).

---

## 9. Windows `resource.syso`

The root package may contain `resource.syso` (PE versioninfo for Windows). The Android (and non-Windows) linker rejects it (`unknown file type`). `scripts/android/build-grudemo.sh` parks/restores it around the link. Do not commit a permanently deleted syso without coordinating Windows packaging.

## 10. What not to do

- Do not import `x11util` from Android-tagged code
- Do not leave NDK `CC`/`CGO_*` in the environment when building desktop
- Do not use NDK **r27** with current raylib-go (see raylib#4166)
- Do not treat FX as the paint layer — see [ANDROID_FX.md](ANDROID_FX.md)
