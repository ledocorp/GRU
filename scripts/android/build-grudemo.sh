#!/usr/bin/env bash
# Build root package (-tags grudemo) as NativeActivity libgrudemo.so.
# Parks Windows resource.syso during the link (PE objects break android ld).
# See docs/ANDROID_PHASE.md M2 · docs/ANDROID_CODE.md §6.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ARCH="${1:-arm64}"

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

OUTDIR="$ROOT/packaging/android/grudemo/libs/${ABI}"
mkdir -p "$OUTDIR"

ASSETS="$ROOT/packaging/android/grudemo/assets"
mkdir -p "$ASSETS"
for d in fonts icons web; do
  if [[ -d "$ROOT/assets/$d" ]]; then
    mkdir -p "$ASSETS/$d"
    cp -a "$ROOT/assets/$d/." "$ASSETS/$d/"
  fi
done

SYS_O="$ROOT/resource.syso"
SYS_BAK="$ROOT/resource.syso.android-park"
parked=0
if [[ -f "$SYS_O" ]]; then
  mv "$SYS_O" "$SYS_BAK"
  parked=1
  echo "parked resource.syso for Android link"
fi
restore_syso() {
  if [[ "$parked" -eq 1 && -f "$SYS_BAK" ]]; then
    mv "$SYS_BAK" "$SYS_O"
    echo "restored resource.syso"
  fi
}
trap restore_syso EXIT

# native_app_glue.h — needed by ui/android_insets_android.go (contentRect / window flags).
GLUE_INC="${ANDROID_NDK_HOME}/sources/android/native_app_glue"

echo "Building libgrudemo.so ($ABI, API $ANDROID_API)..."
CC="${TRIPLE}${ANDROID_API}-clang" \
CGO_CFLAGS="-I${SYSROOT}/usr/include -I${SYSROOT}/usr/include/${TRIPLE} -I${GLUE_INC} --sysroot=${SYSROOT}" \
CGO_LDFLAGS="-L${SYSROOT}/usr/lib/${TRIPLE}/${ANDROID_API} --sysroot=${SYSROOT}" \
CGO_ENABLED=1 GOOS=android GOARCH="$GOARCH" \
go build -tags grudemo -buildmode=c-shared \
  -ldflags="-s -w -extldflags=-Wl,-soname,libgrudemo.so" \
  -o="${OUTDIR}/libgrudemo.so" \
  "$ROOT"

rm -f "${OUTDIR}/libgrudemo.h"
ls -lh "${OUTDIR}/libgrudemo.so"
file "${OUTDIR}/libgrudemo.so"
echo "OK — next: ./scripts/android/assemble-grudemo.sh (needs ANDROID_HOME + JDK)"
