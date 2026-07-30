//go:build android

package tray

// Start is a no-op on Android.
func Start(c Config) {}

// Stop is a no-op on Android.
func Stop() {}
