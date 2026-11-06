#!/usr/bin/env bash
# Mirror headless AVD showing grudemo (Cosmic-safe).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/android/env.sh"
export PATH="$ANDROID_HOME/platform-tools:$PATH"

SCRCPY="$ROOT/.android/scrcpy/scrcpy"
if [[ ! -x "$SCRCPY" ]]; then
  echo "scrcpy missing under .android/scrcpy/" >&2
  exit 1
fi

pkill -f '/.android/scrcpy/scrcpy' 2>/dev/null || true
sleep 0.5

adb wait-for-device
adb shell am start -n org.ledocorp.gru.demo/android.app.NativeActivity >/dev/null 2>&1 || true

export WAYLAND_DISPLAY="${WAYLAND_DISPLAY:-wayland-1}"
unset DISPLAY || true
export SDL_VIDEODRIVER=wayland

exec "$SCRCPY" \
  --window-title "Android Emulator (grudemo)" \
  --video-bit-rate 20M \
  --max-fps 60 \
  --no-audio \
  "$@"
