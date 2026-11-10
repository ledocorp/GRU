# Linux packaging helpers

## Dock / app-grid icon (Wayland)

GLFW cannot set the window icon on Wayland. Cosmic matches **Wayland `app_id`** to `.desktop` `StartupWMClass`. The demo calls `PrepareLinuxAppID("grudemo")` before `InitWindow`.

```bash
go build -tags grudemo -o bin/grudemo .
./packaging/linux/install-desktop.sh "$PWD/bin/grudemo"
# App library → “Gru Demo” → pin (4-tile lettermark). A generic gear means app_id mismatch.
```

Files:

| Path | Role |
|------|------|
| `packaging/icons/hicolor/*/apps/gru.png` | Icon theme tiles (32 / 48 / 256) |
| `packaging/icons/hicolor/*/apps/gru-notepad.png` | Same tiles for Notepad packaging name |
| `packaging/linux/grudemo.desktop` | Template (`StartupWMClass=grudemo`) |
| `packaging/linux/install-desktop.sh` | Copies into `~/.local/share` |

## WebView on Linux

Child-of-GLFW host (mirrors Windows WebView2). Requires X11 GLFW:

```bash
sudo apt install libwebkit2gtk-4.1-dev
GLFW_PLATFORM=x11 go run -tags webkit ./cmd/webviewhello
GLFW_PLATFORM=x11 go run -tags 'grudemo,webkit' .
```

See [docs/architecture/webview.md](../../docs/architecture/webview.md).
