//go:build android

package main

import rl "github.com/gen2brain/raylib-go/raylib"

// NativeActivity entry: raylib's android_main calls the registered callback.
// Without this, main() skips runApp on Android and the activity never starts Gru.
// See docs/ANDROID_CODE.md §2.
func init() {
	rl.SetCallbackFunc(runApp)
}
