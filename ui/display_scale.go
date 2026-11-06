// Package ui (continued) — monitor DPI for supersampling quality.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// DisplayScale is the monitor DPI factor (1.0 at 100%, ~1.25 at 125%, ~1.5 at 150%).
// Refreshed at startup and on window resize via [RefreshDisplayScale].
// Drives T1.9 source atlases — not SSAA (S1 stays 1× present).
var DisplayScale float32 = 1

// BaseSupersamplingScale is the user-facing SSAA tier (native ship default 1×).
var BaseSupersamplingScale float32 = 1

// RefreshDisplayScale updates [DisplayScale].
// Linux (X11+Wayland): keep 1 — window size is the client size; no compositor
// DPI inflate. Windows: follow monitor DPI for T1.9 atlases.
func RefreshDisplayScale() {
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

// EffectiveSupersamplingScale is the SSAA multiplier applied to the UI framebuffer.
// DPI is tracked separately ([DisplayScale]) for T1.9 source-atlas scaling only;
// multiplying SSAA × DPI over-downscales text and makes it look soft/pixelated.
func EffectiveSupersamplingScale() float32 {
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
	RescaleSupersampling(EffectiveSupersamplingScale(), w, h)
	SyncRemixIconFilter()
	return reinit
}
