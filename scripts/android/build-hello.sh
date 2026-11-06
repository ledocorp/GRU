#!/usr/bin/env bash
# Build cmd/hello as a NativeActivity shared library for Android.
# Requires NDK r26.x (not r27). See docs/ANDROID_PHASE.md / ANDROID_CODE.md.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ARCH="${1:-arm64}" # arm64 | arm | x86_64

if [[ -z "${ANDROID_NDK_HOME:-}" ]]; then
  if [[ -f "$ROOT/.android/ndk-home" ]]; then
    ANDROID_NDK_HOME="$(cat "$ROOT/.android/ndk-home")"
  elif [[ -d "$HOME/Android/Sdk/ndk/26.1.10909125" ]]; then
    ANDROID_NDK_HOME="$HOME/Android/Sdk/ndk/26.1.10909125"
  else
    echo "Set ANDROID_NDK_HOME to NDK r26.x" >&2
    exit 1
  fi
fi
export ANDROID_NDK_HOME
ANDROID_API="${ANDROID_API:-21}"

PREBUILT="${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64"
SYSROOT="${PREBUILT}/sysroot"
export PATH="${PREBUILT}/bin:${PATH}"

case "$ARCH" in
  arm64|aarch64)
    GOARCH=arm64
    TRIPLE=aarch64-linux-android
    ABI=arm64-v8a
    ;;
  arm|armeabi-v7a)
    GOARCH=arm
    TRIPLE=armv7a-linux-androideabi
    ABI=armeabi-v7a
    ANDROID_API="${ANDROID_API_ARM:-21}"
    ;;
  x86_64|x64|amd64)
    # Linux host emulator (google_apis;x86_64)
    GOARCH=amd64
    TRIPLE=x86_64-linux-android
    ABI=x86_64
    ;;
  *)
    echo "usage: $0 [arm64|arm|x86_64]" >&2
    exit 1
    ;;
esac

OUTDIR="$ROOT/packaging/android/hello/libs/${ABI}"
mkdir -p "$OUTDIR"

# Stage fonts/icons into APK assets (best-effort).
ASSETS="$ROOT/packaging/android/hello/assets"
mkdir -p "$ASSETS"
for d in fonts icons; do
  if [[ -d "$ROOT/assets/$d" ]]; then
    mkdir -p "$ASSETS/$d"
    cp -a "$ROOT/assets/$d/." "$ASSETS/$d/"
  fi
done

echo "ANDROID_NDK_HOME=$ANDROID_NDK_HOME"
echo "Building libhello.so ($ABI, API $ANDROID_API)..."

# native_app_glue.h — ui/android_insets_android.go (contentRect / window flags).
GLUE_INC="${ANDROID_NDK_HOME}/sources/android/native_app_glue"

# NDK r26+ defines __ANDROID_API__ from the clang target — do not -D redefine it.
CC="${TRIPLE}${ANDROID_API}-clang" \
CGO_CFLAGS="-I${SYSROOT}/usr/include -I${SYSROOT}/usr/include/${TRIPLE} -I${GLUE_INC} --sysroot=${SYSROOT}" \
CGO_LDFLAGS="-L${SYSROOT}/usr/lib/${TRIPLE}/${ANDROID_API} --sysroot=${SYSROOT}" \
CGO_ENABLED=1 GOOS=android GOARCH="$GOARCH" \
go build -buildmode=c-shared \
  -ldflags="-s -w -extldflags=-Wl,-soname,libhello.so" \
  -o="${OUTDIR}/libhello.so" \
  "$ROOT/./cmd/hello"

# Drop the companion .h unless needed for JNI
rm -f "${OUTDIR}/libhello.h"

ls -lh "${OUTDIR}/libhello.so"
file "${OUTDIR}/libhello.so"
echo "OK — next: packaging/android + gradlew :hello:assembleDebug (needs ANDROID_HOME + JDK)"
