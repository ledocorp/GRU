#!/usr/bin/env bash
# Mirror the headless AVD into a Wayland window (Cosmic-safe; Qt emu UI is fragile).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/android/env.sh"
export PATH="$ANDROID_HOME/platform-tools:$PATH"

SCRCPY="$ROOT/.android/scrcpy/scrcpy"
if [[ ! -x "$SCRCPY" ]]; then
  echo "scrcpy missing — run density/smoke flow once, or unpack scrcpy under .android/scrcpy/" >&2
  exit 1
fi

pkill -f '/.android/scrcpy/scrcpy' 2>/dev/null || true
sleep 0.5

adb wait-for-device
adb shell am start -n org.ledocorp.gru.hello/android.app.NativeActivity >/dev/null 2>&1 || true

export WAYLAND_DISPLAY="${WAYLAND_DISPLAY:-wayland-1}"
unset DISPLAY || true
export SDL_VIDEODRIVER=wayland

# Full device resolution + high bitrate (avoid --max-size; it upscales mush).
exec "$SCRCPY" \
  --window-title "Android Emulator (hello)" \
  --video-bit-rate 20M \
  --max-fps 60 \
  --no-audio \
  "$@"
