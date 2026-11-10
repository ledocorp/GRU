# Android phase plan

**Branch:** `android-experience` (from `main`; does not touch `linux-experience`)  
**Status:** **M2 App Shell signed off** · M4 partial (IME / SAF / OpenBrowser) · full-window WebView sample  
**Last updated:** 2026-11-04  
**Out of tree:** product hosts (e.g. [gru-notepad](https://github.com/shawnlondono/gru-notepad)) consume this branch; they are not part of this PR.

Desktop Windows/Linux paths stay behind `linux && !android` / `windows` tags. Android work is additive.

---

## M2 success — App Shell (Go) on AVD

**Signed off 2026-11-04** on AVD `gru_api34_x86_64` via Cosmic-safe headless emulator + scrcpy.

Evidence: `dist/android/grudemo-appshell-success.png` (also `grudemo-appshell-icons-1.png`, `grudemo-appshell-insets-1.png`).

| Check | Result |
|-------|--------|
| NativeActivity + raylib GLES | Paints App Shell full-bleed, ~60 fps |
| Status / gesture insets | Gray Gru clear under status (not black); BottomNav above gesture pill |
| AppBar | Hamburger + overflow glyphs visible |
| BottomNav | Home / Search / Profile Remix icons + labels; Search badge; active underline |
| FAB | Purple `+` above BottomNav, not clipped |
| Shell chrome | Drawer / overflow / BottomSheet wiring intact (demo copy) |
| Density | rem≈40; `PhoneChromeScale` 1.5 (AppBar/FAB); BottomNav icons 56 / bar 100 on Android (`grudemo-appshell-scale-2.png`) |

**What landed for this win**

- NDK content-rect insets — clear raylib `FULLSCREEN`, `LAYOUT_IN_SCREEN|LAYOUT_INSET_DECOR` → `SetChromeTop/Bottom` ([ANDROID_INSETS.md](ANDROID_INSETS.md))
- Thin Java tint (`MainActivity` + `GruDemoTheme`) for status-bar color only
- Remix atlas warm (320) + `DrawIcon` / phosphor text fallbacks for lean-atlas CPs
- Android: no BottomNav auto-hide; `demoNavBottom=0`; SSAA stays 1×
- Page shell padding 20; phone rem floor in `RefreshTypeScaleFromWindow`

**Out of M2 / next product**

- Optional rem / icon-slot bump if phone type still reads small
- Product hosts (Notepad, etc.) live in their own repos; verify against this branch via `replace`
- M3 script polish · optional Custom Tabs / sqlite / touch-scroll

---

## Milestones

| ID | Goal | Success criteria | Status |
|----|------|------------------|--------|
| **M0** | Toolchain + compile | NDK r26; `GOOS=android` c-shared links for hello (and grudemo); no X11/gsettings on Android | **Done** — arm64 `libhello.so` + `libgrudemo.so`; retags + init |
| **M1** | Hello APK | NativeActivity APK installs; paints hello scene; tap button; back exits | **Smoke + tap OK** — density rem/DPR pass (full ANativeWindow, rem≈40, DisplayScale≈2.5) |
| **M2** | grudemo APK | `-tags grudemo` APK; App Shell boots; insets + icons; readable GLES | **Done** — App Shell signed off on AVD (`grudemo-appshell-success.png`) |
| **M3** | Repeatable scripts | `scripts/android/build-hello.sh` (+ grudemo); assets staging; `gru` CLI later | **Partial** — both build scripts landed |
| **M4** | Platform parity | SAF file pickers, `OpenBrowser` intent, IME | **Partial** — IME + SAF + OpenBrowser landed; polish remaining |
| **FX+** | FX enhancements | Guest logic, caps for SAF, theme/density token tables | **Unblocked** — see [ANDROID_FX.md](ANDROID_FX.md) |

---

## Non-goals (this branch)

- Linux WebKit / Cosmic chrome (stays on `linux-experience`)
- FX as a UI renderer (GRU owns paint — [ANDROID_FX.md](ANDROID_FX.md))
- NDK r27 (unsupported by raylib — use **r26.x**)

---

## Toolchain (machine notes)

User-local install (no sudo). Source once per shell:

```bash
source scripts/android/env.sh
```

| Piece | Location / note |
|-------|-----------------|
| JDK 17 | `~/.local/jdk/jdk-17.0.13+11` (Temurin) |
| NDK | `ANDROID_NDK_HOME` → **r26b** `~/Android/Sdk/ndk/26.1.10909125` |
| SDK | `ANDROID_HOME` → `~/Android/Sdk` (platform-tools, platforms;android-34, build-tools;34.0.0) |
| API | Prefer **21+** for arm64 (`aarch64-linux-android21-clang`) |
| Go | Module Go version from `go.mod`; `CGO_ENABLED=1` |

Hints also live under `.android/{java,sdk,ndk}-home` (gitignored).  
Do **not** leave `CC` / `CGO_*` Android vars set when building desktop — scrub PATH of NDK clang first.

---

## Build cheat sheet

```bash
source scripts/android/env.sh

# .so only (arm64 device · x86_64 Linux emulator)
./scripts/android/build-hello.sh arm64
./scripts/android/build-hello.sh x86_64
./scripts/android/build-grudemo.sh

# Hello APK end-to-end (both ABIs) → dist/android/hello-debug.apk
./scripts/android/assemble-hello.sh

# Emulator AVD (API 34 google_apis x86_64)
./scripts/android/run-emulator.sh
adb install -r dist/android/hello-debug.apk
adb shell am start -n org.ledocorp.gru.hello/android.app.NativeActivity
```

Details: [ANDROID_CODE.md](ANDROID_CODE.md) · packaging: [`packaging/android/`](../packaging/android/README.md) · public build index: [BUILD.md](BUILD.md).

---

## Verification log

| Date | Check | Result |
|------|-------|--------|
| 2026-11-04 | `linux && !android` retags | Done |
| 2026-11-04 | `main_android_init.go` + `cmd/hello` Android path | Done |
| 2026-11-04 | `libhello.so` arm64 c-shared | **OK** (~7.8 MB stripped) |
| 2026-11-04 | `libgrudemo.so` arm64 (park `resource.syso`) | **OK** (~11 MB stripped) |
| 2026-11-04 | Desktop `go build ./cmd/hello` | OK |
| 2026-11-04 | Desktop `go build -tags grudemo .` | OK (after scrubbing NDK from env) |
| 2026-11-04 | JDK 17 + Android SDK (user-local) | **OK** |
| 2026-11-04 | `assembleDebug` hello APK | **OK** — dual ABI `dist/android/hello-debug.apk` (~7.4 MB) |
| 2026-11-04 | Emulator + system image | **OK** — AVD `gru_api34_x86_64` (API 34 google_apis x86_64, KVM) |
| 2026-11-04 | Emulator smoke | **OK** — install + launch; EGL ~60 fps; “Getting started” card visible (`dist/android/hello-emulator-3.png`) |
| 2026-11-04 | Hello title / Click me / density | **OK** — tap works; density rem/DPR + full-window size (was 1920 letterbox on 2400) |
| 2026-11-04 | grudemo APK assemble + smoke | **OK** — `dist/android/grudemo-debug.apk`; Counter Demo on AVD (`grudemo-emulator-2.png`) |
| 2026-11-04 | App Shell phone polish | **Partial** — BottomNav/FAB/rem OK; black status band diagnosed as raylib FULLSCREEN vs WM insets |
| 2026-11-04 | NDK content-rect insets | **OK** — top band is Gru clear `(228,229,235)` not black (`grudemo-appshell-insets-1.png`); [ANDROID_INSETS.md](ANDROID_INSETS.md) |
| 2026-11-04 | BottomNav icons | **OK** — house/search/profile + AppBar menu glyphs (`grudemo-appshell-icons-1.png`) |
| 2026-11-04 | **M2 App Shell sign-off** | **OK** — full shell (AppBar + list + Wi-Fi toggle + FAB + BottomNav) on AVD; user: “looks fantastic”; rem still slightly small (`grudemo-appshell-success.png`) |
| 2026-11-04 | Density bump | **OK** — rem 34→40; BottomNav icons 28→56 / bar 100; AppBar/FAB ×1.5 (`grudemo-appshell-scale-2.png`) |
| 2026-11-04 | Notepad Android smoke | **OK** — `gru-notepad` APK paints on AVD (`dist/android/notepad-android-3.png`); sqlite deferred (seccomp); mobile bottom nav forced |
| 2026-11-04 | Density sync design | **OK** — [ANDROID_DENSITY.md](ANDROID_DENSITY.md): one `Density`/`Dp`, retire stacked chrome multipliers; IME show/hide on focus |
| 2026-11-04 | Density+IME foothold ship | **OK** — tool strip/BottomNav via `Dp`; insets every frame; Search tap OK (`notepad-density-2.png`); IME show on text focus (char delivery TBD on AVD) |
| 2026-11-04 | IME defer + adjustNothing | **OK** — arm IME on tap (no startup blank); GLES stays painted with keyboard (`notepad-ime-3/4.png`) |
| 2026-11-04 | Chrome pad + IME chars | **OK** — strip fills cell band; `JNI_OnLoad` RegisterNatives; typed “Hello” on AVD (`notepad-pad-6.png`); Search nav + haptic (`notepad-pad-8.png`) |
| 2026-11-04 | IME inset + SAF + JSON store | **OK** — JSON `store.json` welcome restore; async SAF Open (no ANR); IME height poll + ChromeBottom max(); Save As SAF wired |
| 2026-11-04 | OpenBrowser + WebView host | **OK** — `OpenBrowser` → ACTION_VIEW (`openGruUrl`); full-window WebView sample `org.ledocorp.gru.webview` ([ANDROID_WEBVIEW.md](ANDROID_WEBVIEW.md)) |
| 2026-11-04 | Notepad preview → Chrome | **OK** — tap GFM link in MarkdownView opens Chrome `example.com` (`notepad-link-16.png`) |
| Next | Optional polish | Viewport touch-drag scroll; phone sheets when width≥LG; Custom Tabs; arm64 sqlite experiment; richer WebView bridge |
