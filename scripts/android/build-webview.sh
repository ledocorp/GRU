#!/usr/bin/env bash
# Build cmd/gruwebview as libgruwebview.so (full-window WebView host, no raylib).
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
    GOARCH=amd64
    TRIPLE=x86_64-linux-android
    ABI=x86_64
    ;;
  *)
    echo "usage: $0 [arm64|arm|x86_64]" >&2
    exit 1
    ;;
esac

OUTDIR="$ROOT/packaging/android/webview/libs/${ABI}"
mkdir -p "$OUTDIR"

echo "ANDROID_NDK_HOME=$ANDROID_NDK_HOME"
echo "Building libgruwebview.so ($ABI, API $ANDROID_API)..."

CC="${TRIPLE}${ANDROID_API}-clang" \
CGO_CFLAGS="-I${SYSROOT}/usr/include -I${SYSROOT}/usr/include/${TRIPLE} --sysroot=${SYSROOT}" \
CGO_LDFLAGS="-L${SYSROOT}/usr/lib/${TRIPLE}/${ANDROID_API} --sysroot=${SYSROOT}" \
CGO_ENABLED=1 GOOS=android GOARCH="$GOARCH" \
go build -buildmode=c-shared \
  -ldflags="-s -w -extldflags=-Wl,-soname,libgruwebview.so" \
  -o="${OUTDIR}/libgruwebview.so" \
  "$ROOT/./cmd/gruwebview"

rm -f "${OUTDIR}/libgruwebview.h"
ls -lh "${OUTDIR}/libgruwebview.so"
file "${OUTDIR}/libgruwebview.so"
echo "OK — next: ./scripts/android/assemble-webview.sh"
