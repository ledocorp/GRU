//go:build !windows

package ui

func hostWindowTransparencyHint() string {
	return "Whole-window opacity may work via Raylib SetWindowOpacity on this OS; " +
		"Win32 layered / FLAG_WINDOW_TRANSPARENT experiments are Windows-focused."
}

// EnableHostLayeredStyle is a no-op off Windows.
func EnableHostLayeredStyle() bool { return false }

// ApplyHostLayeredAlpha is a no-op off Windows.
func ApplyHostLayeredAlpha(alpha uint8) { _ = alpha }

// HostWindowExStyleReport is a stub off Windows.
func HostWindowExStyleReport() string { return "layered style: n/a (non-Windows)" }

// ClearHostColorKey is a no-op off Windows.
func ClearHostColorKey() {}

// EnableHostAcrylicBlur is a no-op off Windows.
func EnableHostAcrylicBlur(tintABGR uint32) bool { _ = tintABGR; return false }

// EnableHostBlurBehind is a no-op off Windows.
func EnableHostBlurBehind(tintABGR uint32) bool { _ = tintABGR; return false }

// DisableHostBlurBehind is a no-op off Windows.
func DisableHostBlurBehind() {}
