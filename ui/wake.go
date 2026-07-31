// Package ui (continued)
package ui

import (
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// WakeReason explains why the renderer should stay responsive this frame.
// Reasons are bitflags so a frame can report multiple causes.
type WakeReason uint32

const (
	WakeInput WakeReason = 1 << iota
	WakeScroll
	WakeKeyboard
	WakeAnimation
	WakeOverlay
	WakeDataUpdate
	WakeResize
	WakeScene
	// WakeWebView keeps the frame budget high enough for WebView2 COM while a host is alive.
	WakeWebView
)

// WakeSignal is a low-cost cross-goroutine signal. Raylib input is still polled
// on the main goroutine; this channel is mainly for async data work.
type WakeSignal struct {
	Reason WakeReason
	Source string
}

// WakeSummary is the per-frame aggregate drained at the top of the main loop.
type WakeSummary struct {
	Reasons WakeReason
	Sources []string
}

// AnimationReporter is implemented by widgets that can keep frames active due
// to time-based visuals. It is intentionally separate from Node so normal
// widgets do not pay API surface or bookkeeping cost.
type AnimationReporter interface {
	AnimationActive() bool
	AnimationSource() string
}

const (
	ActiveFPS    = 60
	AnimationFPS = 36
	// ScrollFPS is the presentation rate while wheel gestures are active and the
	// SSAA cache is still valid (blit-only frames). Main bumps here immediately
	// on wheel so deep idle (10 FPS) does not miss slow ticks.
	ScrollFPS = 30
	// WebViewIdleFPS is the idle tier while WebView2 hosts are alive (focused or not).
	// Empirically ~12 is the lowest safe tier with compensating COM pump; 10 correlated with exits.
	WebViewIdleFPS = 12
	// WebViewMinSafeFPS documents the approximate floor — do not idle below this with hosts alive.
	WebViewMinSafeFPS = 12
	// SoftIdleFPS is unused: after grace we jump straight to DeepIdleFPS (no soft-idle ramp).
	SoftIdleFPS = 10
	DeepIdleFPS = 10
)

const (
	activityGraceDuration = 800 * time.Millisecond
	// sceneLoadGraceDuration keeps ActiveFPS after loadScene / first present so
	// deep idle does not engage during atlas settle and the first paint.
	sceneLoadGraceDuration = 1500 * time.Millisecond
	deepIdleDelay         = 6 * time.Second
	// ResizeHoldDuration keeps ActiveFPS after each resize dimension change or
	// borderless grip drag. Time-based so low actual FPS during heavy relayout
	// does not expire the hold early (unlike a frame counter).
	ResizeHoldDuration = 1200 * time.Millisecond
	// ResizeBurstHoldDuration extends the tail when multiple dimension changes land
	// in a short window (continuous edge drag on heavy scenes).
	ResizeBurstHoldDuration = 2800 * time.Millisecond
	// ResizeBurstWindow is the rolling window for counting burst resize steps.
	ResizeBurstWindow = 1800 * time.Millisecond
	// ResizeSettleDuration extends the hold while layout is still dirty shortly
	// after the last resize step (flex/grid reflow finishing).
	ResizeSettleDuration = 1200 * time.Millisecond
	// ResizeIdleCooldown prevents RenderIdlePolicy from stepping down immediately
	// after a resize burst ends — avoids the hard 60→10 FPS cliff between gestures.
	ResizeIdleCooldown = 1800 * time.Millisecond
)

// ResizeHoldTracker keeps target FPS at ActiveFPS through resize bursts and a
// short cooldown tail. Call NoteDimensionChange / NoteGesture from the main
// loop; Active / KeepActiveFPS drive idle policy and SetTargetFPS.
type ResizeHoldTracker struct {
	activeUntil   time.Time
	lastActivity  time.Time
	lastDimChange time.Time
	burstUntil    time.Time
	burstSteps    int
}

// NoteDimensionChange records an OS client-area size change.
func (t *ResizeHoldTracker) NoteDimensionChange(now time.Time) {
	if t == nil {
		return
	}
	t.lastActivity = now
	t.lastDimChange = now
	if now.After(t.burstUntil) {
		t.burstSteps = 0
	}
	t.burstSteps++
	t.burstUntil = now.Add(ResizeBurstWindow)
	hold := ResizeHoldDuration
	if t.burstSteps >= 2 {
		hold = ResizeBurstHoldDuration
	}
	t.extend(now, hold)
}

// NoteGesture records an in-flight borderless edge/corner drag (no dim change yet).
func (t *ResizeHoldTracker) NoteGesture(now time.Time) {
	if t == nil {
		return
	}
	t.lastActivity = now
	t.extend(now, ResizeHoldDuration)
}

// NoteHeavyFrame extends the hold when a full redraw still follows a recent resize
// step (layout/draw still catching up at low actual FPS).
func (t *ResizeHoldTracker) NoteHeavyFrame(now time.Time) {
	if t == nil || t.lastDimChange.IsZero() {
		return
	}
	if now.Sub(t.lastDimChange) > ResizeSettleDuration {
		return
	}
	t.lastActivity = now
	t.extend(now, ResizeHoldDuration)
}

// NoteLayoutSettling extends the hold while the tree is still layout-dirty after resize.
func (t *ResizeHoldTracker) NoteLayoutSettling(now time.Time) {
	if t == nil || t.lastDimChange.IsZero() {
		return
	}
	if now.Sub(t.lastDimChange) > ResizeSettleDuration {
		return
	}
	t.lastActivity = now
	t.extend(now, ResizeHoldDuration)
}

func (t *ResizeHoldTracker) extend(now time.Time, d time.Duration) {
	until := now.Add(d)
	if until.After(t.activeUntil) {
		t.activeUntil = until
	}
}

// Active reports whether the resize burst hold is still in effect.
func (t *ResizeHoldTracker) Active(now time.Time) bool {
	if t == nil {
		return false
	}
	return now.Before(t.activeUntil)
}

// KeepActiveFPS reports whether target FPS must stay at ActiveFPS — includes the
// post-burst cooldown so policy does not rush to DeepIdleFPS between drag steps.
func (t *ResizeHoldTracker) KeepActiveFPS(now time.Time) bool {
	if t == nil {
		return false
	}
	if t.Active(now) {
		return true
	}
	if t.lastActivity.IsZero() {
		return false
	}
	return now.Sub(t.lastActivity) < ResizeIdleCooldown
}

// RecentActivity reports whether a resize happened recently (for heavy-frame extension).
func (t *ResizeHoldTracker) RecentActivity(now time.Time) bool {
	if t == nil || t.lastDimChange.IsZero() {
		return false
	}
	return now.Sub(t.lastDimChange) < ResizeSettleDuration
}

// ResizeHoldSnapshot is a point-in-time view of resize hold state for debug logging.
type ResizeHoldSnapshot struct {
	HoldActive          bool
	KeepActive          bool
	InPostCooldown      bool
	MsUntilHoldEnd      int
	MsSinceLastActivity int
	MsSinceLastDim      int
	BurstSteps          int
}

// Snapshot returns hold tracker fields at now (for F7 resize FPS debug).
func (t *ResizeHoldTracker) Snapshot(now time.Time) ResizeHoldSnapshot {
	if t == nil {
		return ResizeHoldSnapshot{}
	}
	var snap ResizeHoldSnapshot
	snap.HoldActive = t.Active(now)
	snap.KeepActive = t.KeepActiveFPS(now)
	snap.InPostCooldown = !snap.HoldActive && snap.KeepActive
	if !t.activeUntil.IsZero() && now.Before(t.activeUntil) {
		snap.MsUntilHoldEnd = int(t.activeUntil.Sub(now).Milliseconds())
	}
	if !t.lastActivity.IsZero() {
		snap.MsSinceLastActivity = int(now.Sub(t.lastActivity).Milliseconds())
	}
	if !t.lastDimChange.IsZero() {
		snap.MsSinceLastDim = int(now.Sub(t.lastDimChange).Milliseconds())
	}
	snap.BurstSteps = t.burstSteps
	return snap
}

var wakeChan = make(chan WakeSignal, 64)

// WakeOnMouseMove controls whether cursor motion alone bumps idle FPS to ActiveFPS.
// Default false: only scroll, clicks, keys, and other wake reasons restore 60 FPS so
// hovering at 12 FPS idle does not prevent the low-power path. Set true if hover-driven
// UI (tooltips, hover chrome) must stay at full presentation rate without a click.
//
// Exception: SampleChromeHoverWake — title-bar / WebView client hover without enabling
// global mouse-move wake.
var WakeOnMouseMove = false

// SampleChromeHoverWake keeps ActiveFPS while the cursor rests over the borderless
// title band or a live WebView client area (position still tracks under HWND).
// Does not enable WakeOnMouseMove for the rest of the UI.
func SampleChromeHoverWake(windowFocused bool, windowW, windowH int32) WakeSummary {
	var out WakeSummary
	if !windowFocused || windowW < 1 || windowH < 1 || !rl.IsWindowReady() {
		return out
	}
	m := rl.GetMousePosition()
	if m.X < 0 || m.Y < 0 || m.X >= float32(windowW) || m.Y >= float32(windowH) {
		return out
	}
	top := overlayChromeTop
	if top <= 0 {
		if doc := ActiveDocument(); doc != nil && doc.ChromeTop() > 0 {
			top = doc.ChromeTop()
		} else {
			top = TitleBarHeight
		}
	}
	if m.Y < top {
		out.Add(WakeInput, "title-hover")
		return out
	}
	if WebViewHostsActive() {
		out.Add(WakeInput, "webview-hover")
	}
	return out
}

// Wake posts a non-blocking wake signal. It is safe from background goroutines.
func Wake(reason WakeReason, source string) {
	if reason == 0 {
		return
	}
	select {
	case wakeChan <- WakeSignal{Reason: reason, Source: source}:
	default:
		// Dropping a wake is acceptable: the next dirty/input/queue check will
		// reassert activity. Never block a worker on UI wake bookkeeping.
	}
}

// DrainWakeSignals drains all pending wake signals without blocking.
func DrainWakeSignals() WakeSummary {
	var out WakeSummary
	for {
		select {
		case sig := <-wakeChan:
			out.Add(sig.Reason, sig.Source)
		default:
			return out
		}
	}
}

// Add merges a reason/source into this summary.
func (s *WakeSummary) Add(reason WakeReason, source string) {
	if reason == 0 {
		return
	}
	s.Reasons |= reason
	if source != "" {
		s.Sources = append(s.Sources, source)
	}
}

// Merge combines two summaries.
func (s WakeSummary) Merge(other WakeSummary) WakeSummary {
	s.Reasons |= other.Reasons
	s.Sources = append(s.Sources, other.Sources...)
	return s
}

// Any reports whether the summary contains at least one reason.
func (s WakeSummary) Any() bool { return s.Reasons != 0 }

// CollectAnimationWake walks a widget tree and reports active time-based
// visuals. Call from the main loop after Update so widget state is current.
func CollectAnimationWake(root Node) WakeSummary {
	var out WakeSummary
	collectAnimationWake(root, &out)
	return out
}

func collectAnimationWake(n Node, out *WakeSummary) {
	if n == nil || n.IsHidden() {
		return
	}
	if anim, ok := n.(AnimationReporter); ok && anim.AnimationActive() {
		out.Add(WakeAnimation, anim.AnimationSource())
	}
	for _, child := range n.Children() {
		collectAnimationWake(child, out)
	}
}

// AnyDropdownOpen reports whether a dropdown or combobox list is expanded
// anywhere under root. The main loop uses this to stay at ActiveFPS while a
// menu is open.
func AnyDropdownOpen(root Node) bool {
	if root == nil {
		return false
	}
	if len(collectOpenDropdowns([]Node{root})) > 0 {
		return true
	}
	return len(collectOpenComboBoxes([]Node{root})) > 0
}

// String returns a compact, stable reason list for logs and overlays.
func (r WakeReason) String() string {
	if r == 0 {
		return "none"
	}
	parts := make([]string, 0, 8)
	if r&WakeInput != 0 {
		parts = append(parts, "input")
	}
	if r&WakeScroll != 0 {
		parts = append(parts, "scroll")
	}
	if r&WakeKeyboard != 0 {
		parts = append(parts, "keyboard")
	}
	if r&WakeAnimation != 0 {
		parts = append(parts, "animation")
	}
	if r&WakeOverlay != 0 {
		parts = append(parts, "overlay")
	}
	if r&WakeDataUpdate != 0 {
		parts = append(parts, "data")
	}
	if r&WakeResize != 0 {
		parts = append(parts, "resize")
	}
	if r&WakeScene != 0 {
		parts = append(parts, "scene")
	}
	if r&WakeWebView != 0 {
		parts = append(parts, "webview")
	}
	return strings.Join(parts, "|")
}

// RenderIdlePolicy tracks active/soft-idle/deep-idle FPS selection.
type RenderIdlePolicy struct {
	targetFPS           int
	state               string
	lastWake            time.Time
	lastInteractiveWake time.Time
	lastReason          WakeReason
	sceneGraceUntil     time.Time
}

// NewRenderIdlePolicy starts in active mode.
func NewRenderIdlePolicy(now time.Time) *RenderIdlePolicy {
	return &RenderIdlePolicy{
		targetFPS:           ActiveFPS,
		state:               "active",
		lastWake:            now,
		lastInteractiveWake: now,
		lastReason:          WakeScene,
	}
}

// Update returns the target FPS for the current frame window.
// When resizeHold is true the policy stays at ActiveFPS regardless of clean cache
// hits — used during resize bursts and the post-burst cooldown tail.
func (p *RenderIdlePolicy) Update(now time.Time, wake WakeSummary, clean bool, blockers WakeReason, resizeHold bool) int {
	if p == nil {
		return ActiveFPS
	}
	if resizeHold {
		p.lastWake = now
		p.lastInteractiveWake = now
		p.lastReason = WakeResize
		p.targetFPS = ActiveFPS
		p.state = "resize-hold"
		return ActiveFPS
	}
	if !p.sceneGraceUntil.IsZero() && now.Before(p.sceneGraceUntil) {
		p.targetFPS = ActiveFPS
		p.state = "scene-grace"
		return ActiveFPS
	}
	reasons := wake.Reasons | blockers
	if reasons != 0 || !clean {
		nonAnimationReasons := reasons &^ WakeAnimation
		p.lastWake = now
		if reasons != 0 {
			p.lastReason = reasons
		} else {
			p.lastReason = WakeDataUpdate
		}
		if nonAnimationReasons != 0 || !clean {
			p.lastInteractiveWake = now
		}
		if clean && reasons != 0 && reasons&^WakeAnimation == 0 {
			if now.Sub(p.lastInteractiveWake) < activityGraceDuration {
				p.targetFPS = ActiveFPS
				p.state = "grace"
				return p.targetFPS
			}
			// Animation pixels draw in DrawAnimationOverlays after cache blit.
			p.targetFPS = AnimationFPS
			p.state = "animation"
			return p.targetFPS
		}
		if clean && reasons == WakeScroll {
			p.targetFPS = ScrollFPS
			p.state = "scroll"
			return p.targetFPS
		}
		p.targetFPS = ActiveFPS
		p.state = "active"
		return p.targetFPS
	}

	idleFor := now.Sub(p.lastInteractiveWake)
	if idleFor >= activityGraceDuration {
		// One idle tier: no soft-idle ramp. Input/scroll still wake to ActiveFPS
		// at the top of the frame via sampleInputWake before this runs.
		p.targetFPS = DeepIdleFPS
		if idleFor >= deepIdleDelay {
			p.state = "deep-idle"
		} else {
			p.state = "idle"
		}
	} else {
		p.targetFPS = ActiveFPS
		p.state = "grace"
	}
	return p.targetFPS
}

func (p *RenderIdlePolicy) TargetFPS() int {
	if p == nil || p.targetFPS == 0 {
		return ActiveFPS
	}
	return p.targetFPS
}

func (p *RenderIdlePolicy) State() string {
	if p == nil || p.state == "" {
		return "active"
	}
	return p.state
}

func (p *RenderIdlePolicy) LastReason() WakeReason {
	if p == nil {
		return 0
	}
	return p.lastReason
}

// NoteInteractiveWake extends the active/grace window after mid-frame input (clicks,
// scene navigation) so the next clean frame does not drop to DeepIdleFPS immediately.
func (p *RenderIdlePolicy) NoteInteractiveWake(now time.Time) {
	if p == nil {
		return
	}
	p.lastWake = now
	p.lastInteractiveWake = now
}

// NoteSceneLoad holds ActiveFPS for sceneLoadGraceDuration after a scene Build /
// first present so startup and Tab switches do not drop to DeepIdleFPS mid-settle.
func (p *RenderIdlePolicy) NoteSceneLoad(now time.Time) {
	if p == nil {
		return
	}
	p.lastWake = now
	p.lastInteractiveWake = now
	p.lastReason = WakeScene
	p.sceneGraceUntil = now.Add(sceneLoadGraceDuration)
	p.targetFPS = ActiveFPS
	p.state = "scene-grace"
}
