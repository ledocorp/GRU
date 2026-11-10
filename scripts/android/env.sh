# Source this file to use the user-local Android toolchain (no sudo required).
#   source scripts/android/env.sh
#
# Layout (defaults):
#   JAVA_HOME      ~/.local/jdk/jdk-17.0.13+11
#   ANDROID_HOME   ~/Android/Sdk
#   ANDROID_NDK_HOME ~/Android/Sdk/ndk/26.1.10909125

_GRU_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

if [[ -f "$_GRU_ROOT/.android/java-home" ]]; then
  # file may be "JAVA_HOME=..." or a bare path — take the last non-empty line's value
  _jh="$(grep -E '(/|JAVA_HOME=)' "$_GRU_ROOT/.android/java-home" | tail -1 | sed 's/^JAVA_HOME=//')"
  [[ -n "$_jh" ]] && export JAVA_HOME="$_jh"
fi
: "${JAVA_HOME:=$HOME/.local/jdk/jdk-17.0.13+11}"

if [[ -f "$_GRU_ROOT/.android/sdk-home" ]]; then
  export ANDROID_HOME="$(cat "$_GRU_ROOT/.android/sdk-home")"
fi
: "${ANDROID_HOME:=$HOME/Android/Sdk}"
export ANDROID_SDK_ROOT="$ANDROID_HOME"

if [[ -f "$_GRU_ROOT/.android/ndk-home" ]]; then
  export ANDROID_NDK_HOME="$(cat "$_GRU_ROOT/.android/ndk-home")"
fi
: "${ANDROID_NDK_HOME:=$ANDROID_HOME/ndk/26.1.10909125}"

export PATH="$JAVA_HOME/bin:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$ANDROID_HOME/emulator:$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin:$PATH"

echo "JAVA_HOME=$JAVA_HOME"
echo "ANDROID_HOME=$ANDROID_HOME"
echo "ANDROID_NDK_HOME=$ANDROID_NDK_HOME"
java -version 2>&1 | head -1
command -v adb >/dev/null && adb version | head -1
unset _GRU_ROOT _jh
