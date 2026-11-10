#!/usr/bin/env bash
# Boot the Gru Linux emulator AVD (API 34 google_apis x86_64).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/android/env.sh"
export PATH="$ANDROID_HOME/emulator:$PATH"

AVD_NAME="${1:-gru_api34_x86_64}"

# If already running, print device and exit
if adb devices 2>/dev/null | grep -q 'emulator-'; then
  echo "Emulator already running:"
  adb devices -l
  exit 0
fi

# Cosmic/Wayland: the emulator Qt UI needs X11 and often crashes (or kills the
# session). Default is headless. For a visible mirror, set GRU_EMU_SCRCPY=1
# (uses .android/scrcpy). GRU_EMU_WINDOW=1 tries the native Qt window (fragile).
WINDOW_FLAGS=(-no-window)
if [[ "${GRU_EMU_WINDOW:-0}" == "1" ]]; then
  WINDOW_FLAGS=()
  echo "Starting AVD $AVD_NAME WITH Qt window on DISPLAY=${DISPLAY:-?}..."
else
  echo "Starting AVD $AVD_NAME headless..."
fi

# Soft GPU + no snapshots: avoids host X crashes and pending-snapshot hangs.
nohup emulator -avd "$AVD_NAME" \
  "${WINDOW_FLAGS[@]}" \
  -gpu swiftshader_indirect \
  -no-snapshot -no-snapshot-load -no-snapshot-save \
  -no-boot-anim -no-audio -no-metrics \
  >"$ROOT/.android/emulator.log" 2>&1 &
echo $! >"$ROOT/.android/emulator.pid"
echo "emulator pid $(cat "$ROOT/.android/emulator.pid") — log .android/emulator.log"

echo "Waiting for device..."
adb wait-for-device
# Wait until boot completed
for i in $(seq 1 120); do
  boot="$(adb shell getprop sys.boot_completed 2>/dev/null | tr -d '\r')"
  if [[ "$boot" == "1" ]]; then
    echo "Boot completed."
    adb devices -l
    exit 0
  fi
  sleep 2
done
echo "Timed out waiting for boot — see .android/emulator.log" >&2
exit 1
