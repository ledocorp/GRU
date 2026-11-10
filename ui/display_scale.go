// Package ui (continued) — monitor DPI for supersampling quality.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// DisplayScale is the monitor DPI factor (1.0 at 100%, ~1.25 at 125%, ~1.5 at 150%).
// Refreshed at startup and on window resize via [RefreshDisplayScale].
// Drives T1.9 source atlases — not SSAA (S1 stays 1× present).
var DisplayScale float32 = 1

// PhoneChromeScale lifts fixed chrome sizes (AppBar height, FAB, shared pads)
// on Android. Prefer [Dp] for new chrome. This bridge caps [Density] so older
// call sites (×1.5-era) do not become billboards again on 420dpi devices.
// Typography stays on [RootFontSize] — do not multiply rem into chrome.
// See docs/ANDROID_DENSITY.md.
func PhoneChromeScale() float32 {
	if !AndroidGLES() {
		return 1
	}
	d := Density()
	if d < 1 {
		return 1
	}
	if d > 2 {
		return 2
	}
	return d
}

// BaseSupersamplingScale is the user-facing SSAA tier (native ship default 1×).
// Android GLES ignores it — [EffectiveSupersamplingScale] forces 1.
var BaseSupersamplingScale float32 = 1

// RefreshDisplayScale updates [DisplayScale].
//   - Android: approximate from framebuffer width (GetWindowScaleDPI is
//     unimplemented there) and refresh [Density].
//   - Linux (X11+Wayland): keep 1 — window size is the client size; no
//     compositor DPI inflate.
//   - Windows: follow monitor DPI for T1.9 atlases.
func RefreshDisplayScale() {
	if AndroidGLES() {
		DisplayScale = androidDisplayScaleFromWidth(rl.GetScreenWidth())
		RefreshDensity()
		return
	}
	if platformKeepDisplayScaleUnity() {
		DisplayScale = 1
		return
	}
	dpi := rl.GetWindowScaleDPI()
	s := (dpi.X + dpi.Y) * 0.5
	if s < 1.01 {
		if ms, ok := platformMonitorContentScale(); ok && ms > 1.01 {
			s = ms
		}
	}
	if s < 1 {
		s = 1
	}
	if s > 2.5 {
		s = 2.5
	}
	DisplayScale = s
}

// androidDisplayScaleFromWidth maps NativeActivity pixel width to a DPR bucket.
// 1080-wide @ ~420dpi ≈ 2.6×; we clamp to 2.5 (atlas cap) for xxhdpi phones.
func androidDisplayScaleFromWidth(w int) float32 {
	switch {
	case w >= 1400:
		return 2.5
	case w >= 1000:
		return 2.5 // 1080p phones (Pixel / AVD 420dpi)
	case w >= 720:
		return 2.0
	case w >= 480:
		return 1.5
	default:
		return 1.25
	}
}

// EffectiveSupersamplingScale is the SSAA multiplier applied to the UI framebuffer.
// DPI is tracked separately ([DisplayScale]) for T1.9 source-atlas scaling only;
// multiplying SSAA × DPI over-downscales text and makes it look soft/pixelated.
// Android GLES: always 1× — 2× of 1080×2400 (2160×4800) exceeds common emulator
// max_texture_size (4096) and leaves a black surface.
func EffectiveSupersamplingScale() float32 {
	if AndroidGLES() {
		return 1
	}
	s := BaseSupersamplingScale
	if s < 1 {
		return 1
	}
	if s > 2.5 {
		return 2.5
	}
	return s
}

// ApplyDisplayAwareSupersampling recreates SSAA targets using [EffectiveSupersamplingScale].
// When monitor DPI changes, SDF and Remix source atlases are rebuilt (T1.9).
// Returns true when atlases were reloaded (caller may flush widget GPU caches).
func ApplyDisplayAwareSupersampling(w, h int32) bool {
	if w < 1 || h < 1 {
		return false
	}
	reinit := ReinitDisplayAwareAtlasesIfNeeded()
	if AndroidGLES() {
		RenderScale = 1
		return reinit
	}
	RescaleSupersampling(EffectiveSupersamplingScale(), w, h)
	SyncRemixIconFilter()
	return reinit
}
