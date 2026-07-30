// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── Supersampling / Render-scale system ─────────────────────────────────────
//
// Gru supports optional 2× supersampled anti-aliasing (SSAA) via an
// off-screen RenderTexture2D. When active:
//
//  1. BeginSuperFrame draws the entire widget tree into a 2× texture using a
//     Camera2D zoom-2 transform, so all widget code uses normal 1× coordinates.
//  2. EndSuperFrame finalises the texture.
//  3. BlitToScreen downscales the 2× texture to the window with bilinear
//     filtering — the GPU effectively performs 4-sample SSAA for free.
//
// The scissor-coordinate wrapper (beginScissorMode, used internally by all
// widgets) multiplies coordinates by RenderScale before calling
// rl.BeginScissorMode, keeping clip regions correctly aligned in the 2× target.
//
// Usage (main.go):
//
//	ui.InitSupersampling(windowW, windowH)
//	defer ui.UnloadSupersampling()
//
//	// Inside the frame loop:
//	ui.BeginSuperFrame(clearColor)
//	doc.Root.Draw()
//	ui.EndSuperFrame()
//	rl.BeginDrawing()
//	ui.BlitToScreen(windowW, windowH)
//	rl.EndDrawing()
//
// ─── Dynamic SSAA ─────────────────────────────────────────────────────────────
//
// RescaleSupersampling(scale, w, h) recreates the off-screen target at a new
// pixel density without touching any widget code. main.go calls this when FPS
// drops below a threshold to trade visual quality for performance, and calls
// it back to 2× when headroom is restored.

// RenderScale is the current supersampling multiplier (1.0, 1.5, or 2.0).
// Widget code uses 1× logical coordinates throughout; this value is applied
// internally by BeginSuperFrame (Camera2D zoom) and beginScissorMode (scissor
// coordinate scaling). Use SetSupersamplingScale to change it at runtime.
var RenderScale float32 = 1

// superTarget is the 2× off-screen render target for cached UI content.
// overlayTarget is a separate transparent 2× target for transient overlays.
// Keeping overlays separate prevents menus/modals from being burned into the
// persistent UI cache when they close.
var superTarget rl.RenderTexture2D
var overlayTarget rl.RenderTexture2D
var superFrameActive bool
var overlayFrameActive bool

// logicalW / logicalH store the window size used at InitSupersampling so that
// SetSupersamplingScale can recreate the render target without the caller
// needing to track the window size.
var logicalW, logicalH int32

// clampDim enforces minimum dimensions for GPU render targets. Zero-sized
// textures produce GL_FRAMEBUFFER_INCOMPLETE_ATTACHMENT and can take down the GL context.
func clampDim(w, h int32) (int32, int32) {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

// InitSupersampling creates 2x render targets for the main UI and for transient
// overlays (modals, tooltips, menus). Call once after rl.InitWindow.
// Pair with defer ui.UnloadSupersampling().
//
// Both targets are allocated eagerly. Deferring the overlay RT until first use
// (EnsureOverlayTarget-only) left modals/menus soft or missing after rescale —
// do not revive that footprint shortcut for apps that paint native overlays.
func InitSupersampling(w, h int32) {
	w, h = clampDim(w, h)
	logicalW, logicalH = w, h
	superTarget = rl.LoadRenderTexture(w*2, h*2)
	overlayTarget = rl.LoadRenderTexture(w*2, h*2)
	rl.SetTextureFilter(superTarget.Texture, rl.FilterBilinear)
	rl.SetTextureFilter(overlayTarget.Texture, rl.FilterBilinear)
	RenderScale = 2
}

// EnsureOverlayTarget is a safety net if the overlay RT was cleared (e.g. mid
// rescale). Prefer InitSupersampling / RescaleSupersampling creating both RTs
// together — do not rely on lazy first-overlay allocation as the primary path.
func EnsureOverlayTarget() {
	if overlayTarget.ID != 0 || RenderScale <= 1 || logicalW < 1 || logicalH < 1 {
		return
	}
	texW, texH := clampDim(int32(float32(logicalW)*RenderScale), int32(float32(logicalH)*RenderScale))
	overlayTarget = rl.LoadRenderTexture(texW, texH)
	rl.SetTextureFilter(overlayTarget.Texture, rl.FilterBilinear)
}

// SetSupersamplingScale adjusts the render scale at runtime without requiring
// the caller to track the window size.
//
//   - 1.0 (or ≤ 1.0) — disable SSAA (render at native 1×)
//   - 1.5             — 1.5× SSAA
//   - 2.0             — 2× SSAA (default, highest quality)
//
// Changing scale recreates the GPU render target; call this only when the FPS
// monitor decides to step up or down (not every frame).
func SetSupersamplingScale(scale float32) {
	RescaleSupersampling(scale, logicalW, logicalH)
}

// UnloadSupersampling releases the GPU render target.  Call on shutdown (or via defer).
func UnloadSupersampling() {
	if superTarget.ID != 0 {
		rl.UnloadRenderTexture(superTarget)
		superTarget = rl.RenderTexture2D{}
	}
	if overlayTarget.ID != 0 {
		rl.UnloadRenderTexture(overlayTarget)
		overlayTarget = rl.RenderTexture2D{}
	}
	RenderScale = 1
}

// SupersamplingActive returns true when the scaled render target has been
// initialised and is ready for use.
func SupersamplingActive() bool { return superTarget.ID != 0 }

// RescaleSupersampling changes the supersampling multiplier at runtime,
// recreating the GPU render target if the scale or logical size changed.
//
//   - scale 1.0 (or less) → disable SSAA entirely (RenderScale = 1.0)
//   - scale 1.5           → 1.5× SSAA
//   - scale 2.0           → 2× SSAA (same as InitSupersampling)
//
// Call during the frame boundary (outside Begin/EndDrawing). Typically driven
// by the dynamic-SSAA FPS monitor in main.go or via SetSupersamplingScale.
//
// Important: when scale is unchanged but w/h differ (window resize), this still
// resizes the RT via ResizeWindowTextures. Skipping that stretches the old
// target across the new client (fat text / clipped columns).
func RescaleSupersampling(scale float32, w, h int32) {
	if scale < 1.0 {
		scale = 1.0
	}
	if scale == RenderScale {
		if w > 0 && h > 0 {
			ResizeWindowTextures(w, h)
		}
		return
	}
	if superTarget.ID != 0 {
		rl.UnloadRenderTexture(superTarget)
		superTarget = rl.RenderTexture2D{}
	}
	if overlayTarget.ID != 0 {
		rl.UnloadRenderTexture(overlayTarget)
		overlayTarget = rl.RenderTexture2D{}
	}
	// Update stored logical size so SetSupersamplingScale can call us without args.
	if w > 0 && h > 0 {
		logicalW, logicalH = w, h
	}
	if scale > 1.0 {
		texW, texH := clampDim(int32(float32(logicalW)*scale), int32(float32(logicalH)*scale))
		superTarget = rl.LoadRenderTexture(texW, texH)
		overlayTarget = rl.LoadRenderTexture(texW, texH)
		rl.SetTextureFilter(superTarget.Texture, rl.FilterBilinear)
		rl.SetTextureFilter(overlayTarget.Texture, rl.FilterBilinear)
	}
	RenderScale = scale
}

// ResizeWindowTextures re-creates the supersampled render target to match new
// window dimensions. Call this whenever rl.GetScreenWidth/Height changes.
// It is a no-op when the dimensions are unchanged or when SSAA is disabled.
func ResizeWindowTextures(w, h int32) {
	// Minimized / invalid framebuffer: keep last good logical size and GPU target
	// so we do not allocate a 0×0 or incomplete FBO (raylib logs and may abort).
	if w < 1 || h < 1 {
		return
	}
	w, h = clampDim(w, h)
	if w == logicalW && h == logicalH {
		return
	}
	logicalW, logicalH = w, h
	if superTarget.ID != 0 {
		rl.UnloadRenderTexture(superTarget)
		texW, texH := clampDim(int32(float32(w)*RenderScale), int32(float32(h)*RenderScale))
		superTarget = rl.LoadRenderTexture(texW, texH)
		rl.SetTextureFilter(superTarget.Texture, rl.FilterBilinear)
	}
	if overlayTarget.ID != 0 {
		rl.UnloadRenderTexture(overlayTarget)
		texW, texH := clampDim(int32(float32(w)*RenderScale), int32(float32(h)*RenderScale))
		overlayTarget = rl.LoadRenderTexture(texW, texH)
		rl.SetTextureFilter(overlayTarget.Texture, rl.FilterBilinear)
	}
}

// BeginSuperFrame begins the 2× supersampled draw pass.
//
// It activates the off-screen render target, clears it, and installs a
// Camera2D with zoom = RenderScale (2) so all subsequent raylib draw calls
// use normal 1× widget coordinates while rendering at 2× pixel density.
//
// Window fill must run *inside* Mode2D. Painting a logical (winW×winH) fill
// into the 2× RT *before* Mode2D only covered the top-left quarter and left a
// hard edge at mid-screen — the crosshair artifact on empty calc backgrounds.
//
// Call EndSuperFrame when the draw pass is complete.
func BeginSuperFrame(clearColor rl.Color, borderless bool, winW, winH int32) {
	superFrameActive = true
	rl.BeginTextureMode(superTarget)
	rl.ClearBackground(rl.Blank)
	cam := rl.Camera2D{}
	cam.Zoom = RenderScale
	rl.BeginMode2D(cam)
	if borderless {
		DrawBorderlessWindowFill(winW, winH, clearColor)
	} else {
		rl.DrawRectangle(0, 0, winW, winH, clearColor)
	}
}

// EndSuperFrame closes the Camera2D transform and ends the render-texture pass.
func EndSuperFrame() {
	rl.EndMode2D()
	rl.EndTextureMode()
	superFrameActive = false
}

// BeginDrawingBorderless prepares the default framebuffer before blit.
// Borderless: transparent clear + rounded fill (not ClearBackground(bgColor)).
func BeginDrawingBorderless(clearColor rl.Color, borderless bool, winW, winH int32) {
	if borderless {
		rl.ClearBackground(rl.Blank)
		DrawBorderlessWindowFill(winW, winH, clearColor)
	} else {
		rl.ClearBackground(clearColor)
	}
}

// SuperFrameActive reports whether drawing is currently happening inside the
// supersampled render target. Overlay widgets use this to choose correct scissor
// coordinates when they can be drawn either inside the SSAA pass or directly to
// the screen.
func SuperFrameActive() bool { return superFrameActive }

// BeginOverlaySuperFrame starts a transparent supersampled pass for transient
// overlays that must not become part of the cached main UI texture.
func BeginOverlaySuperFrame() {
	EnsureOverlayTarget()
	if overlayTarget.ID == 0 {
		return
	}
	overlayFrameActive = true
	superFrameActive = true
	rl.BeginTextureMode(overlayTarget)
	rl.ClearBackground(rl.Blank)
	cam := rl.Camera2D{}
	cam.Zoom = RenderScale
	rl.BeginMode2D(cam)
	rl.BeginBlendMode(rl.BlendAlpha)
}

// EndOverlaySuperFrame closes the transparent overlay pass.
func EndOverlaySuperFrame() {
	if !overlayFrameActive {
		return
	}
	rl.EndBlendMode()
	rl.EndMode2D()
	rl.EndTextureMode()
	superFrameActive = false
	overlayFrameActive = false
}

// BlitToScreen draws the supersampled texture onto the window at the given
// logical dimensions using bilinear downscaling.
// Must be called between rl.BeginDrawing / rl.EndDrawing.
//
// The RenderTexture y-axis is flipped relative to screen-space (OpenGL
// convention); negating the source height in DrawTexturePro corrects this.
func BlitToScreen(dstW, dstH int32) {
	BlitToScreenBorderless(dstW, dstH, false)
}

// BlitToScreenBorderless blits the SSAA target. Borderless uses alpha so RT
// corners show the rounded fill drawn by BeginDrawingBorderless underneath.
func BlitToScreenBorderless(dstW, dstH int32, borderless bool) {
	if !SuperTargetDrawable() || dstW < 1 || dstH < 1 {
		return
	}
	tex := superTarget.Texture
	src := rl.NewRectangle(0, 0, float32(tex.Width), -float32(tex.Height))
	dst := rl.NewRectangle(0, 0, float32(dstW), float32(dstH))
	if borderless {
		rl.BeginBlendMode(rl.BlendAlpha)
	}
	rl.DrawTexturePro(tex, src, dst, rl.NewVector2(0, 0), 0, rl.White)
	if borderless {
		rl.EndBlendMode()
	}
}

// BlitOverlayToScreen composites the transparent supersampled overlay target
// over the already-blitted main UI.
func BlitOverlayToScreen(dstW, dstH int32) {
	if !OverlayTargetDrawable() || dstW < 1 || dstH < 1 {
		return
	}
	tex := overlayTarget.Texture
	src := rl.NewRectangle(0, 0, float32(tex.Width), -float32(tex.Height))
	dst := rl.NewRectangle(0, 0, float32(dstW), float32(dstH))
	rl.BeginBlendMode(rl.BlendAlpha)
	rl.DrawTexturePro(tex, src, dst, rl.NewVector2(0, 0), 0, rl.White)
	rl.EndBlendMode()
}

// SuperTargetDrawable reports whether the SSAA render target can be bound and
// sampled safely (avoids crashes during window teardown or GL transitions).
func SuperTargetDrawable() bool {
	if !rl.IsWindowReady() || superTarget.ID == 0 {
		return false
	}
	t := superTarget.Texture
	return t.ID != 0 && t.Width > 0 && t.Height > 0
}

// OverlayTargetDrawable reports whether the transient overlay target is usable.
func OverlayTargetDrawable() bool {
	if !rl.IsWindowReady() || overlayTarget.ID == 0 {
		return false
	}
	t := overlayTarget.Texture
	return t.ID != 0 && t.Width > 0 && t.Height > 0
}

// ─── Per-frame Profiling ─────────────────────────────────────────────────────
//
// FrameStats is populated by main.go each frame and read by the Inspector and
// the optional F11 performance overlay in the nav bar.
//
// Timings are CPU-side (time.Since) and approximate the cost of each UI
// phase — useful for spotting runaway Layout calls or large draw-call counts.

// FrameStats holds per-frame profiling data.
type FrameStats struct {
	UpdateMs         float32    // time spent in doc.Root.Update (ms)
	LayoutMs         float32    // time spent in doc.Root.Layout (ms)
	DrawMs           float32    // time spent in the super-frame draw pass (ms)
	TotalMs          float32    // total frame wall time (ms) from rl.GetFrameTime
	CacheHit         bool       // true when the widget draw pass was skipped (cache valid)
	WakeReasons      WakeReason // reasons that kept/woke active mode this frame
	IdleState        string     // active, grace, soft-idle, deep-idle
	TargetFPS        int        // current render policy target FPS
	FullRedraw       bool       // true when widget tree drew into the cache/SSAA target
	QueueDrained     int        // QueueMain callbacks run at frame start
	AnimationActive  bool       // true when any widget reports time-based animation
	AnimationCount   int        // number of active animation reporters
	AnimationSources string     // compact source list for diagnostics
}

// PerfStats is the live per-frame stats written by main.go.
var PerfStats FrameStats

// ShowPerfOverlay controls the F11 nav-bar performance strip.
var ShowPerfOverlay bool

// rolling cache hit counters (written by RecordCacheHit / RecordCacheMiss)
var (
	cacheHitCount  int
	cacheMissCount int
)

// RecordCacheHit marks the current frame as a cache hit (draw pass skipped).
func RecordCacheHit() {
	cacheHitCount++
	PerfStats.CacheHit = true
}

// RecordCacheMiss marks the current frame as a cache miss (full draw ran).
func RecordCacheMiss() {
	cacheMissCount++
	PerfStats.CacheHit = false
}

// CacheHitRate returns the lifetime cache-hit rate as a value 0..1.
// Returns 0 when no frames have been recorded yet.
func CacheHitRate() float32 {
	total := cacheHitCount + cacheMissCount
	if total == 0 {
		return 0
	}
	return float32(cacheHitCount) / float32(total)
}

// ─── Runtime render capability detection ─────────────────────────────────────

// RenderCapability describes the graphics feature tier detected at startup.
type RenderCapability int

const (
	// RenderGPUFull: MSAA + 2× SSAA + SDF shader — maximum visual quality.
	RenderGPUFull RenderCapability = iota
	// RenderGPUBasic: MSAA + 2× SSAA, no SDF shader — good quality.
	RenderGPUBasic
	// RenderCPULow: neither SSAA nor SDF — software / minimal driver fallback.
	RenderCPULow
)

// CurrentRenderMode stores the tier detected during startup. Updated by
// DetectRenderCapability(); read by any code that wants to branch on quality.
var CurrentRenderMode RenderCapability

// DetectRenderCapability inspects which GPU features were successfully
// initialised and returns the appropriate tier. Call this once after
// InitSupersampling and InitSDFFont have both been attempted.
func DetectRenderCapability() RenderCapability {
	switch {
	case superTarget.ID != 0 && sdfReady:
		return RenderGPUFull
	case superTarget.ID != 0:
		return RenderGPUBasic
	default:
		return RenderCPULow
	}
}

// RenderCapabilityString returns a human-readable label for the current tier.
func RenderCapabilityString() string {
	switch CurrentRenderMode {
	case RenderGPUFull:
		return "GPU Full (MSAA + 2×SSAA + SDF)"
	case RenderGPUBasic:
		return "GPU Basic (MSAA + 2×SSAA)"
	default:
		return "CPU Low (no SSAA)"
	}
}

// beginScissorMode is the package-internal replacement for rl.BeginScissorMode.
//
// All widget draw code calls this function instead of rl.BeginScissorMode
// directly. When supersampling is active, every coordinate is multiplied by
// RenderScale (float32: 1.0, 1.5, or 2.0) before the underlying raylib call,
// keeping the scissor rect correctly aligned with the scaled render target even
// though widget code still uses 1× logical coordinates.
//
// Camera2D transforms draw-call coordinates automatically, but
// rl.BeginScissorMode always takes raw pixel coordinates — this wrapper
// bridges that gap.
//
// scissorRectI32 converts a logical clip rectangle to pixel scissor dimensions
// using Ceil on the far edges so fractional layout coords do not shave the last
// row/column (nested card borders were clipped at the bottom/right).
func scissorRectI32(r rl.Rectangle) (x, y, w, h int32) {
	x0 := int32(r.X)
	y0 := int32(r.Y)
	x1 := int32(math.Ceil(float64(r.X + r.Width)))
	y1 := int32(math.Ceil(float64(r.Y + r.Height)))
	if x1 <= x0 {
		x1 = x0 + 1
	}
	if y1 <= y0 {
		y1 = y0 + 1
	}
	return x0, y0, x1 - x0, y1 - y0
}

// beginScissorFromRect is the preferred entry for widget draw clips.
func beginScissorFromRect(r rl.Rectangle) {
	x, y, w, h := scissorRectI32(r)
	beginScissorMode(x, y, w, h)
}

func beginScissorMode(x, y, w, h int32) {
	s := RenderScale
	rl.BeginScissorMode(
		int32(float32(x)*s),
		int32(float32(y)*s),
		int32(float32(w)*s),
		int32(float32(h)*s),
	)
}
