# Android packaging

NativeActivity APKs for Gru. Phase plan: [`docs/ANDROID_PHASE.md`](../../docs/ANDROID_PHASE.md) (**M2 App Shell signed off** 2026-11-04). Code map: [`docs/ANDROID_CODE.md`](../../docs/ANDROID_CODE.md). Insets: [`docs/ANDROID_INSETS.md`](../../docs/ANDROID_INSETS.md).

| Module | Shared lib | Application id | Milestone |
|--------|------------|----------------|-----------|
| `hello/` | `libs/*/libhello.so` | `org.ledocorp.gru.hello` | M1 done |
| `grudemo/` | `libs/*/libgrudemo.so` | `org.ledocorp.gru.demo` | **M2 done** — App Shell on AVD |
| `webview/` | `libs/*/libgruwebview.so` | `org.ledocorp.gru.webview` | Full-window WebView + Go JNI ([ANDROID_WEBVIEW.md](../../docs/ANDROID_WEBVIEW.md)) |

Evidence: `dist/android/grudemo-appshell-success.png`.

## Prerequisites

```bash
source scripts/android/env.sh   # JDK 17 + SDK + NDK (user-local paths)
```

1. **NDK r26.x** (not r27) — `ANDROID_NDK_HOME`
2. **Android SDK** + **JDK 17+** — `ANDROID_HOME` / `JAVA_HOME`
3. Go with cgo (`CGO_ENABLED=1`)

## Build

```bash
./scripts/android/build-hello.sh          # arm64-v8a libhello.so
./scripts/android/build-grudemo.sh        # arm64 / also: x86_64 for emulator
./scripts/android/build-webview.sh        # full-window WebView host (no raylib)
./scripts/android/assemble-hello.sh       # → dist/android/hello-debug.apk
./scripts/android/assemble-grudemo.sh     # → dist/android/grudemo-debug.apk
./scripts/android/assemble-webview.sh     # → dist/android/webview-debug.apk
```

APKs ship `arm64-v8a` + `x86_64` by default.

```bash
./scripts/android/run-emulator.sh          # AVD gru_api34_x86_64 (headless on Cosmic)
adb install -r dist/android/hello-debug.apk
adb shell am start -n org.ledocorp.gru.hello/android.app.NativeActivity
./scripts/android/scrcpy-hello.sh          # Wayland mirror

adb install -r dist/android/grudemo-debug.apk
adb shell am start -n org.ledocorp.gru.demo/.MainActivity
./scripts/android/scrcpy-grudemo.sh

adb install -r dist/android/webview-debug.apk
adb shell am start -n org.ledocorp.gru.webview/.MainActivity
```

`hello` / `grudemo` `MainActivity` subclasses NativeActivity (status-bar tint + `openGruUrl`). `webview` is a plain Activity + system WebView.

## Notes

- Manifest uses NativeActivity (or `MainActivity`) + `android.app.lib_name`
- Portrait locked; GLES2 required; SSAA stays 1× on Android
- Root package builds must **park** `resource.syso` (Windows PE) — scripts do this automatically
- After install, wait ~25s for Remix atlas warm before judging BottomNav icons
