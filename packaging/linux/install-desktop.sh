#!/usr/bin/env bash
# Install Cosmic/GNOME dock icon + .desktop for Gru Demo (Wayland-safe).
# Reuses packaging/icons/hicolor PNGs (same lettermark as SetWindowIcons / tray).
#
# Usage (from repo root):
#   ./packaging/linux/install-desktop.sh
#   ./packaging/linux/install-desktop.sh /absolute/path/to/grudemo
#
# WebKit demo (X11 embed, Windows-like chrome):
#   go build -tags 'grudemo,webkit' -o bin/grudemo .
#   GLFW_PLATFORM=x11 ./bin/grudemo
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ICON_SRC="$ROOT/packaging/icons/hicolor"
DESKTOP_SRC="$ROOT/packaging/linux/grudemo.desktop"
SHARE="${XDG_DATA_HOME:-$HOME/.local/share}"
ICON_DST="$SHARE/icons/hicolor"
APP_DST="$SHARE/applications"

BIN="${1:-}"
if [[ -z "$BIN" ]]; then
  if [[ -x "$ROOT/bin/grudemo" ]]; then
    BIN="$ROOT/bin/grudemo"
  elif [[ -x "$ROOT/grudemo" ]]; then
    BIN="$ROOT/grudemo"
  else
    echo "building bin/grudemo (tags=grudemo)…"
    mkdir -p "$ROOT/bin"
    (cd "$ROOT" && go build -tags grudemo -o bin/grudemo .)
    BIN="$ROOT/bin/grudemo"
  fi
fi
BIN="$(cd "$(dirname "$BIN")" && pwd)/$(basename "$BIN")"
if [[ ! -x "$BIN" ]]; then
  echo "error: binary not executable: $BIN" >&2
  exit 1
fi

mkdir -p "$ICON_DST" "$APP_DST"

# Cosmic/GTK icon lookup requires a theme index under the user hicolor tree.
if [[ ! -f "$ICON_DST/index.theme" ]]; then
  if [[ -f /usr/share/icons/hicolor/index.theme ]]; then
    cp -f /usr/share/icons/hicolor/index.theme "$ICON_DST/index.theme"
  else
    cat >"$ICON_DST/index.theme" <<'EOF'
[Icon Theme]
Name=Hicolor
Comment=Fallback icon theme
Directories=32x32/apps,48x48/apps,256x256/apps
Hidden=true
Description=Fallback
[32x32/apps]
Size=32
Context=Applications
Type=Threshold
[48x48/apps]
Size=48
Context=Applications
Type=Threshold
[256x256/apps]
Size=256
Context=Applications
Type=Threshold
EOF
  fi
fi

ICON_FILE=""
for sz in 256x256 48x48 32x32; do
  src="$ICON_SRC/$sz/apps/gru.png"
  if [[ ! -f "$src" ]]; then
    echo "error: missing $src" >&2
    exit 1
  fi
  mkdir -p "$ICON_DST/$sz/apps"
  # Install as both gru and grudemo — Cosmic often looks up Icon= by app_id.
  cp -f "$src" "$ICON_DST/$sz/apps/gru.png"
  cp -f "$src" "$ICON_DST/$sz/apps/grudemo.png"
  if [[ -f "$ICON_SRC/$sz/apps/gru-notepad.png" ]]; then
    cp -f "$ICON_SRC/$sz/apps/gru-notepad.png" "$ICON_DST/$sz/apps/gru-notepad.png"
  fi
  if [[ -z "$ICON_FILE" ]]; then
    ICON_FILE="$ICON_DST/$sz/apps/grudemo.png"
  fi
done

mkdir -p "${XDG_DATA_HOME:-$HOME/.local/share}/pixmaps"
cp -f "$ICON_FILE" "${XDG_DATA_HOME:-$HOME/.local/share}/pixmaps/grudemo.png"
cp -f "$ICON_FILE" "${XDG_DATA_HOME:-$HOME/.local/share}/pixmaps/gru.png"

# StartupWMClass + desktop id must match PrepareLinuxAppID("grudemo").
# Cosmic/Pop often fail theme lookup for user hicolor — write absolute Icon=.
tmp="$(mktemp)"
awk -v bin="$BIN" -v icon="$ICON_FILE" '
  BEGIN { FS=OFS="=" }
  /^Exec=/ { print "Exec=" bin; next }
  /^Icon=/ { print "Icon=" icon; next }
  /^StartupWMClass=/ { print "StartupWMClass=grudemo"; next }
  { print }
' "$DESKTOP_SRC" >"$tmp"
# Belt-and-suspenders: ensure Icon= line exists even if template lacked it.
if ! grep -q '^Icon=' "$tmp"; then
  printf 'Icon=%s\n' "$ICON_FILE" >>"$tmp"
fi
install -m 644 "$tmp" "$APP_DST/grudemo.desktop"
rm -f "$tmp"

if command -v gtk-update-icon-cache >/dev/null 2>&1; then
  gtk-update-icon-cache -f -t "$ICON_DST" 2>/dev/null || true
fi
if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database "$APP_DST" 2>/dev/null || true
fi
# Cosmic caches desktop/icons aggressively — poke portals if present.
if command -v dbus-send >/dev/null 2>&1; then
  dbus-send --session --dest=org.freedesktop.FileManager1 --type=method_call \
    /org/freedesktop/FileManager1 org.freedesktop.FileManager1.OpenLocations \
    array:string: string: 2>/dev/null || true
fi

echo "installed:"
echo "  $APP_DST/grudemo.desktop"
echo "  Icon=$ICON_FILE"
grep -E '^(Exec|Icon|StartupWMClass)=' "$APP_DST/grudemo.desktop" | sed 's/^/  /'
echo "  StartupWMClass=grudemo (must match PrepareLinuxAppID)"
echo "Unpin any gear dock entry, then: gtk-launch grudemo  (or pin from app library)."
