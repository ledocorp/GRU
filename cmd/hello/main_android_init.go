//go:build android

package main

import rl "github.com/gen2brain/raylib-go/raylib"

// NativeActivity entry for the hello APK (raylib android_main → callback).
// See docs/ANDROID_CODE.md §2 and docs/ANDROID_PHASE.md M1.
func init() {
	rl.SetCallbackFunc(runApp)
}
