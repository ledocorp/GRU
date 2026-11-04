//go:build !android

package ui

// EnsureAndroidInsetFlags is a no-op off Android.
func EnsureAndroidInsetFlags() {}

// AndroidContentInsets returns zeros off Android.
func AndroidContentInsets(screenW, screenH int32) (top, bottom, left, right float32) {
	return 0, 0, 0, 0
}

// ApplyAndroidSafeInsets is a no-op off Android.
func ApplyAndroidSafeInsets(doc *Document, extraBottom float32, screenW, screenH int32) bool {
	return false
}
