#!/usr/bin/env bash
# Build libgrudemo.so + assemble debug APK (M2).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/android/env.sh"

# Default: arm64 (devices) + x86_64 (Linux host emulator).
if [[ $# -eq 0 ]]; then
  "$ROOT/scripts/android/build-grudemo.sh" arm64
  "$ROOT/scripts/android/build-grudemo.sh" x86_64
else
  "$ROOT/scripts/android/build-grudemo.sh" "$1"
fi

cd "$ROOT/packaging/android"
echo "sdk.dir=$ANDROID_HOME" > local.properties
./gradlew :grudemo:assembleDebug --no-daemon

APK="$ROOT/packaging/android/grudemo/build/outputs/apk/debug/grudemo-debug.apk"
mkdir -p "$ROOT/dist/android"
cp -a "$APK" "$ROOT/dist/android/grudemo-debug.apk"
ls -lh "$APK" "$ROOT/dist/android/grudemo-debug.apk"
unzip -l "$APK" | grep 'lib/' || true
echo "Install: adb install -r dist/android/grudemo-debug.apk"
echo "Launch:  adb shell am start -n org.ledocorp.gru.demo/android.app.NativeActivity"
