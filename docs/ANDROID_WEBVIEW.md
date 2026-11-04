# Android full-window WebView

Wails/Tauri-style option: a **system `WebView` fills the Activity**; Go runs in `libgruwebview.so` and talks over JNI. This is **not** a raylib/GLES embed and is separate from desktop `WebViewPanel` / WebView2.

## When to use

| Path | Use |
|------|-----|
| NativeActivity + raylib (`hello`, `grudemo`, product hosts) | Retained-mode Gru UI |
| `OpenBrowser` / ACTION_VIEW | Hand off to Chrome / Custom Tabs |
| **This sample** (`org.ledocorp.gru.webview`) | App whose primary surface is HTML + a Go bridge |

## Layout

```
Activity (MainActivity)
   └── WebView (match_parent)
         ├── assets/www/index.html
         └── JavascriptInterface "gru"  ←→  JNI  ←→  Go (c-shared)
```

| Piece | Path |
|-------|------|
| Go entry | `cmd/gruwebview/` (`main_android.go` registers natives) |
| Activity | `packaging/android/webview/java/.../MainActivity.java` |
| Sample page | `packaging/android/webview/assets/www/index.html` |
| Build | `scripts/android/build-webview.sh` / `assemble-webview.sh` |

## Bridge contract (v0)

- **JS → Go:** `gru.postMessage(string)` → `nativeOnJsMessage`
- **Go → JS:** Activity `evaluateJs(script)` (called from native after echo)
- Sample: Ping button round-trips through Go and appends `go→js: …` in the page log

## Build & run

```bash
source scripts/android/env.sh
./scripts/android/assemble-webview.sh          # arm64 + x86_64
adb install -r dist/android/webview-debug.apk
adb shell am start -n org.ledocorp.gru.webview/.MainActivity
adb logcat -s gruwebview
```

## Related

- External browser from NativeActivity apps: `ui.OpenBrowser` → `openGruUrl` (ACTION_VIEW)
- Desktop FillClient WebView2: [WEBVIEW.md](WEBVIEW.md)
- Phase log: [ANDROID_PHASE.md](ANDROID_PHASE.md)
