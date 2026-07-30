// Package appicon applies OS-facing GRU icons using raylib (SetWindowIcons) and embedded PNGs/ICOs.
//
// Regenerate all assets: go run ./cmd/gru icons regen  (or go generate .)
// See docs/APP_ICONS.md for the full workflow.
package appicon

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed data/window-16.png
var window16PNG []byte

//go:embed data/window-32.png
var window32PNG []byte

//go:embed data/window-48.png
var window48PNG []byte

//go:embed data/window-dark-16.png
var windowDark16PNG []byte

//go:embed data/window-dark-32.png
var windowDark32PNG []byte

//go:embed data/window-dark-48.png
var windowDark48PNG []byte

//go:embed data/notify-16.png
var notify16PNG []byte

//go:embed data/notify-32.png
var notify32PNG []byte

//go:embed data/notify-dark-16.png
var notifyDark16PNG []byte

//go:embed data/notify-dark-32.png
var notifyDark32PNG []byte

//go:embed data/app.ico
var appICO []byte

//go:embed data/notify.ico
var notifyICO []byte

//go:embed data/notify-dark.ico
var notifyDarkICO []byte

const pngRel = "packaging/icons/hicolor/256x256/apps/gru-notepad.png"
const notifyPNGRel = "packaging/icons/hicolor/32x32/apps/gru-notify.png"

var (
	applyOnce      sync.Once
	preferDarkIcon bool
	preferDarkMu   sync.RWMutex
)

// SetPreferDarkIcon selects the light-chrome or dark-chrome PNG set for runtime icons
// (title bar + SetWindowIcons). The embedded .exe icon stays on the light variant.
func SetPreferDarkIcon(dark bool) {
	preferDarkMu.Lock()
	preferDarkIcon = dark
	preferDarkMu.Unlock()
	InvalidateTitleBarIcon()
}

// PreferDarkIcon reports whether the dark-chrome icon set is active.
func PreferDarkIcon() bool {
	preferDarkMu.RLock()
	defer preferDarkMu.RUnlock()
	return preferDarkIcon
}

// PNGPath returns the baked lettermark PNG on disk, or "" if missing.
func PNGPath() string { return resolveRel(pngRel) }

// NotifyPNGPath returns the notify tile PNG on disk, or "" if missing.
func NotifyPNGPath() string { return resolveRel(notifyPNGRel) }

func resolveRel(rel string) string {
	if _, err := os.Stat(rel); err == nil {
		return rel
	}
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, filepath.FromSlash(rel))
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// ApplyWindowIcon sets the taskbar/window icon via raylib SetWindowIcons (GLFW WM_SETICON).
// Safe to call multiple times; icons are applied once. Call after InitWindow when IsWindowReady.
//
// This is the supported raylib path for borderless GLFW apps on Windows — do not use LoadImage .ico.
func ApplyWindowIcon() {
	if !rl.IsWindowReady() {
		return
	}
	applyOnce.Do(applyWindowIcons)
}

// ReapplyWindowIcon forces a fresh SetWindowIcons (e.g. after first frame / window show).
func ReapplyWindowIcon() {
	applyOnce = sync.Once{}
	applyWindowIcons()
}

func applyWindowIcons() {
	icons := loadIconImages()
	if len(icons) == 0 {
		if iconDebug() {
			fmt.Fprintln(os.Stderr, "appicon: no window icon PNGs loaded (run: go run ./cmd/gru icons regen)")
		}
		return
	}
	if iconDebug() {
		fmt.Fprintf(os.Stderr, "appicon: SetWindowIcons %d sizes\n", len(icons))
	}
	rl.SetWindowIcons(icons, int32(len(icons)))
	for i := range icons {
		rl.UnloadImage(&icons[i])
	}
}

func iconDebug() bool {
	if os.Getenv("GRU_ICON_DEBUG") != "" {
		return true
	}
	return os.Getenv("GORY_ICON_DEBUG") != ""
}

func loadIconImages() []rl.Image {
	if PreferDarkIcon() {
		return loadEmbeddedWindowIcons(windowDark16PNG, windowDark32PNG, windowDark48PNG)
	}
	// Prefer on-disk master PNG so taskbar matches Demo Directory preview after gen.
	if path := PNGPath(); path != "" {
		if icons := loadScaledIcons(path); len(icons) > 0 {
			return icons
		}
	}
	return loadEmbeddedWindowIcons(window16PNG, window32PNG, window48PNG)
}

func loadEmbeddedWindowIcons(s16, s32, s48 []byte) []rl.Image {
	var out []rl.Image
	for _, png := range [][]byte{s16, s32, s48} {
		if img := loadPNG(png); img != nil {
			out = append(out, *img)
		}
	}
	return out
}

func loadScaledIcons(path string) []rl.Image {
	img := rl.LoadImage(path)
	if img == nil || img.Width == 0 {
		return nil
	}
	defer rl.UnloadImage(img)
	var out []rl.Image
	for _, sz := range []int32{16, 32, 48} {
		dup := rl.ImageCopy(img)
		if dup == nil {
			continue
		}
		rl.ImageResize(dup, sz, sz)
		if dup.Format != rl.UncompressedR8g8b8a8 {
			rl.ImageFormat(dup, rl.UncompressedR8g8b8a8)
		}
		out = append(out, *dup)
	}
	return out
}

func loadPNG(b []byte) *rl.Image {
	if len(b) == 0 {
		return nil
	}
	img := rl.LoadImageFromMemory(".png", b, int32(len(b)))
	if img == nil || img.Width == 0 {
		return nil
	}
	// GLFW SetWindowIcons requires R8G8B8A8; stb/PNG often loads as 24-bit RGB.
	if img.Format != rl.UncompressedR8g8b8a8 {
		rl.ImageFormat(img, rl.UncompressedR8g8b8a8)
	}
	return img
}

// TrayIconBytes returns bytes for the system tray icon.
// Windows LoadImage requires classic .ico; Linux/macOS accept PNG.
func TrayIconBytes() []byte {
	if PreferDarkIcon() {
		if runtime.GOOS == "windows" {
			return append([]byte(nil), notifyDarkICO...)
		}
		return append([]byte(nil), notifyDark32PNG...)
	}
	if runtime.GOOS == "windows" {
		return append([]byte(nil), notifyICO...)
	}
	return append([]byte(nil), notify32PNG...)
}

// TrayPNGBytes returns embedded 16×16 notify tile PNG for the active chrome variant.
func TrayPNGBytes() []byte {
	if PreferDarkIcon() {
		return append([]byte(nil), notifyDark16PNG...)
	}
	return append([]byte(nil), notify16PNG...)
}

// NotifyPNGBytes returns embedded notify tile PNG bytes (32×32).
func NotifyPNGBytes() []byte { return append([]byte(nil), notify32PNG...) }

// WindowPNGBytes returns embedded 32×32 window icon PNG bytes for the active variant.
func WindowPNGBytes() []byte {
	if PreferDarkIcon() {
		return append([]byte(nil), windowDark32PNG...)
	}
	return append([]byte(nil), window32PNG...)
}

func activeWindow32PNG() []byte {
	if PreferDarkIcon() {
		return windowDark32PNG
	}
	return window32PNG
}

// AppICORaw returns embedded lettermark .ico (for goversioninfo / packaging only).
func AppICORaw() []byte { return appICO }

// NotifyICORaw returns embedded notify .ico.
func NotifyICORaw() []byte { return notifyICO }
