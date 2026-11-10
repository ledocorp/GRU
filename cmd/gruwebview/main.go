//go:build !android

// gruwebview — Android full-window WebView host (Wails/Tauri-style).
//
// Desktop: this binary only prints usage. Live host is the Android APK:
//
//	./scripts/android/assemble-webview.sh
//	adb install -r dist/android/webview-debug.apk
//
// See docs/ANDROID_WEBVIEW.md.
package main

import "fmt"

func main() {
	fmt.Println("gruwebview is an Android WebView host sample.")
	fmt.Println("Build: ./scripts/android/assemble-webview.sh")
	fmt.Println("Docs:  docs/ANDROID_WEBVIEW.md")
}
