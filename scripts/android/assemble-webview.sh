#!/usr/bin/env bash
# Build libgruwebview.so + assemble debug APK (full-window WebView sample).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/android/env.sh"

if [[ $# -eq 0 ]]; then
  "$ROOT/scripts/android/build-webview.sh" arm64
  "$ROOT/scripts/android/build-webview.sh" x86_64
else
  "$ROOT/scripts/android/build-webview.sh" "$1"
fi

cd "$ROOT/packaging/android"
echo "sdk.dir=$ANDROID_HOME" > local.properties
./gradlew :webview:assembleDebug --no-daemon

APK="$ROOT/packaging/android/webview/build/outputs/apk/debug/webview-debug.apk"
mkdir -p "$ROOT/dist/android"
cp -a "$APK" "$ROOT/dist/android/webview-debug.apk"
ls -lh "$APK" "$ROOT/dist/android/webview-debug.apk"
unzip -l "$APK" | grep 'lib/' || true
echo "Install: adb install -r dist/android/webview-debug.apk"
echo "Start:   adb shell am start -n org.ledocorp.gru.webview/.MainActivity"
