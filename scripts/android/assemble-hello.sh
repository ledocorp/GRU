#!/usr/bin/env bash
# Build libhello.so + assemble debug APK.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/android/env.sh"

# Default: arm64 (devices) + x86_64 (Linux host emulator). Pass a single arch to override.
if [[ $# -eq 0 ]]; then
  "$ROOT/scripts/android/build-hello.sh" arm64
  "$ROOT/scripts/android/build-hello.sh" x86_64
else
  "$ROOT/scripts/android/build-hello.sh" "$1"
fi

cd "$ROOT/packaging/android"
echo "sdk.dir=$ANDROID_HOME" > local.properties
./gradlew :hello:assembleDebug --no-daemon

APK="$ROOT/packaging/android/hello/build/outputs/apk/debug/hello-debug.apk"
mkdir -p "$ROOT/dist/android"
cp -a "$APK" "$ROOT/dist/android/hello-debug.apk"
ls -lh "$APK" "$ROOT/dist/android/hello-debug.apk"
unzip -l "$APK" | grep 'lib/' || true
echo "Install: adb install -r dist/android/hello-debug.apk"
echo "Emulator: ./scripts/android/run-emulator.sh"
