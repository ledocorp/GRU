// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"time"

	"github.com/fogleman/gg"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─────────────────────────────────────────────────────────────────────────────
// ToastLevel
// ─────────────────────────────────────────────────────────────────────────────

// ToastLevel controls the colour scheme and icon of a toast notification.
type ToastLevel int

const (
	// ToastInfo uses a blue palette — neutral information or tips.
	ToastInfo ToastLevel = iota
	// ToastSuccess uses a green palette — a completed operation.
	ToastSuccess
	// ToastWarning uses an amber palette — caution or partial success.
	ToastWarning
	// ToastError uses a red palette — a failure or critical alert.
	ToastError
)

// ─────────────────────────────────────────────────────────────────────────────
// Toast entry
// ─────────────────────────────────────────────────────────────────────────────

// toast is a single active notification entry owned by ToastManager.
//
// Fields are intentionally unexported — callers interact only through
// ShowToast / ShowToastClickable / DismissAll.
type toast struct {
	message   string
	level     ToastLevel
	duration  float32 // total visible seconds; 0 = sticky
	remaining float32 // seconds until auto-dismiss

	// Animation state.
	alpha float32 // 0..1 opacity applied to all drawn colours

	// Callbacks and tweens.
	onClick      func() // optional whole-card click (ShowToastClickable)
	actionLabel  string // optional trailing action (e.g. "Undo")
	onAction     func()
	slideTween *Tween
	fadeTween  *Tween
	dismissing bool
}

// ─────────────────────────────────────────────────────────────────────────────
// Layout constants
// ─────────────────────────────────────────────────────────────────────────────

const (
	toastW         float32 = 340 // toast card width
	toastMinH      float32 = 60  // minimum card height (toastCardHeight may exceed)
	toastPadX      float32 = 14  // horizontal inner padding (left of content)
	toastPadY      float32 = 11  // vertical inner padding (top of content)
	toastIconSz    float32 = 22  // rendered icon size in pixels
	toastAccentW   float32 = 4   // coloured left-accent bar width
	toastGap       float32 = 8   // vertical gap between stacked toasts
	toastMarginR   float32 = 20  // right margin from window right edge
	toastMarginB   float32 = 54  // bottom margin (clears the nav bar)
	toastRadius    float32 = 8   // card corner radius
	toastProgressH float32 = 3   // progress bar height
	toastMaxCount  int     = 5   // maximum simultaneous toasts

	toastFadeInDur  float32 = 0.15 // entry fade seconds
	toastFadeOutDur float32 = 0.12 // exit fade seconds
)

// iconCtxSz is the gg context size used to pre-render icons. Chosen at 40 for
// 2× oversampling when icons are rendered at toastIconSz = 20–22 px.
const iconCtxSz = 40

// ─────────────────────────────────────────────────────────────────────────────
// Level palette
// ─────────────────────────────────────────────────────────────────────────────

type toastPaletteEntry struct {
	bg, border, accent rl.Color
	label              string
	progressColor      rl.Color
}

// toastPalette is indexed by ToastLevel.
var toastPalette = [4]toastPaletteEntry{
	// ToastInfo — blue
	{
		bg:            rl.NewColor(239, 246, 255, 255),
		border:        rl.NewColor(186, 212, 253, 255),
		accent:        rl.NewColor(59, 130, 246, 255),
		label:         "Info",
		progressColor: rl.NewColor(59, 130, 246, 200),
	},
	// ToastSuccess — green
	{
		bg:            rl.NewColor(240, 253, 244, 255),
		border:        rl.NewColor(134, 239, 172, 255),
		accent:        rl.NewColor(34, 197, 94, 255),
		label:         "Success",
		progressColor: rl.NewColor(34, 197, 94, 200),
	},
	// ToastWarning — amber
	{
		bg:            rl.NewColor(255, 251, 235, 255),
		border:        rl.NewColor(252, 211, 77, 255),
		accent:        rl.NewColor(245, 158, 11, 255),
		label:         "Warning",
		progressColor: rl.NewColor(245, 158, 11, 200),
	},
	// ToastError — red
	{
		bg:            rl.NewColor(254, 242, 242, 255),
		border:        rl.NewColor(252, 165, 165, 255),
		accent:        rl.NewColor(239, 68, 68, 255),
		label:         "Error",
		progressColor: rl.NewColor(239, 68, 68, 200),
	},
}

// ─────────────────────────────────────────────────────────────────────────────
// ToastManager
// ─────────────────────────────────────────────────────────────────────────────

// ToastManager manages a live stack of toast notifications and renders them
// as an always-on-top screen overlay.
//
// # Positioning
//
// Toasts are anchored to the bottom-right corner of the window. The newest
// toast appears first (lowest), and each subsequent toast is stacked above.
// At most toastMaxCount (5) toasts are visible simultaneously; excess toasts
// displace the oldest one.
//
// # Animation
//
// Entry: fade in over 150 ms (EaseOutQuad). Exit: fade out over 120 ms when
// auto-dismissed or clicked. Fade-only keeps motion smooth when frame time spikes;
// visible toasts also hold ActiveFPS via OverlayIdleBlockers.
//
// # Progress bar
//
// Timed toasts show a thin progress bar at the bottom of the card that
// depletes linearly from full to empty as the timer counts down.
//
// # Sticky toasts
//
// Pass 0 as duration to ShowToast to create a toast that never
// auto-dismisses. The user must click it to remove it.
//
// # Main-loop integration
//
//	// After doc.Root.Update(dt):
//	ui.Toasts.Update(dt)
//
//	// Inside rl.BeginDrawing() / rl.EndDrawing(), after other overlay draws:
//	ui.Toasts.Draw()
//
//	// On shutdown:
//	defer ui.Toasts.Unload()
//
// # Example
//
//	ui.ShowToast("File saved successfully.", ui.ToastSuccess, 3*time.Second)
//	ui.ShowToastClickable("Click to open settings.", ui.ToastInfo, 0, func() {
//	    doc.SetScene(settingsScene)
//	})
type ToastManager struct {
	toasts     []*toast
	winW, winH int32

	// Pre-rendered gg vector icons, uploaded on first Draw call.
	icons       [4]rl.Texture2D
	iconsLoaded bool
}

// Toasts is the package-level singleton. Wire it into main.go's Update and
// Draw calls exactly as Tooltips, ModalMgr, etc.
var Toasts = &ToastManager{winW: 1280, winH: 720}

// SetWindowSize configures the screen dimensions used for overlay positioning.
// Call once after rl.InitWindow with the same width/height as the window.
func (m *ToastManager) SetWindowSize(w, h int32) { m.winW = w; m.winH = h }

// ─────────────────────────────────────────────────────────────────────────────
// Public API
// ─────────────────────────────────────────────────────────────────────────────

// ShowToast queues a new toast notification.
//
// # LLM Prompt Template
//
//	ui.ShowToast("Saved", ui.ToastSuccess, 2*time.Second)
//	// main loop: ui.Toasts.Update(dt); ui.Toasts.Draw()
//
// Demo scenes: **AppShell Demo**, **Batch 1**, Notepad Find/Replace.
//
//	message  — the displayed text (kept to one line; long messages are clipped).
//	level    — one of ToastInfo / ToastSuccess / ToastWarning / ToastError.
//	duration — how long the toast stays visible (use 0 for a sticky toast).
func ShowToast(message string, level ToastLevel, duration time.Duration) {
	ShowToastClickable(message, level, duration, nil)
}

// ShowToastWithAction shows a toast with a trailing action label (Material-style).
// Clicking the action runs onAction and dismisses the toast. The message area
// dismisses on click only when onClick is set via ShowToastClickable.
func ShowToastWithAction(message string, level ToastLevel, duration time.Duration, actionLabel string, onAction func()) {
	if actionLabel == "" || onAction == nil {
		ShowToast(message, level, duration)
		return
	}
	m := Toasts
	Wake(WakeOverlay, "toast")
	if len(m.toasts) >= toastMaxCount {
		for i, t := range m.toasts {
			if !t.dismissing {
				m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
				break
			}
		}
	}
	secs := float32(duration.Seconds())
	t := &toast{
		message:     message,
		level:       level,
		duration:    secs,
		remaining:   secs,
		alpha:       0,
		actionLabel: actionLabel,
		onAction:    onAction,
	}
	startToastFadeIn(t)
	m.toasts = append(m.toasts, t)
	appendNotificationHistory(message, level)
}

// ShowToastClickable is like ShowToast but fires onClick when the user clicks
// the card. The toast is always dismissed after the click regardless of whether
// onClick is nil.
func ShowToastClickable(message string, level ToastLevel, duration time.Duration, onClick func()) {
	m := Toasts
	Wake(WakeOverlay, "toast")
	// Evict the oldest non-dismissing entry when the cap is reached.
	if len(m.toasts) >= toastMaxCount {
		for i, t := range m.toasts {
			if !t.dismissing {
				m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
				break
			}
		}
	}

	secs := float32(duration.Seconds())
	t := &toast{
		message:   message,
		level:     level,
		duration:  secs,
		remaining: secs,
		alpha:     0,
		onClick:   onClick,
	}
	startToastFadeIn(t)
	m.toasts = append(m.toasts, t)
	appendNotificationHistory(message, level)
}

// DismissAll triggers the exit animation on every visible toast.
func DismissAll() {
	for _, t := range Toasts.toasts {
		if !t.dismissing {
			startDismiss(t)
		}
	}
}

// ActiveToastCount returns the number of toasts currently in the queue
// (including those still in their exit animation). Safe to call at any time.
func ActiveToastCount() int { return len(Toasts.toasts) }

// IsAnimating reports active fade tweens on any toast.
func (m *ToastManager) IsAnimating() bool {
	for _, t := range m.toasts {
		if t.slideTween != nil && t.slideTween.IsActive {
			return true
		}
		if t.fadeTween != nil && t.fadeTween.IsActive {
			return true
		}
	}
	return false
}

// ─────────────────────────────────────────────────────────────────────────────
// Update
// ─────────────────────────────────────────────────────────────────────────────

// Update advances timers and animations, and handles click-to-dismiss.
// Call once per frame, after doc.Root.Update(dt).
func (m *ToastManager) Update(dt float32) {
	// Phase 1 — advance all tweens and timers.
	for _, t := range m.toasts {
		advanceToastTween(t.slideTween, dt)
		advanceToastTween(t.fadeTween, dt)
		// Auto-dismiss countdown (only while alive and timed).
		if !t.dismissing && t.duration > 0 {
			t.remaining -= dt
			if t.remaining <= 0 {
				t.remaining = 0
				startDismiss(t)
			}
		}
	}

	// Phase 2 — click hit-test (must run BEFORE the cleanup pass so that
	// indices into m.toasts still match toastRect(i) correctly).
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		mouse := rl.GetMousePosition()
		n := len(m.toasts)
		for i := 0; i < n; i++ {
			t := m.toasts[i]
			if t.dismissing || t.alpha < 0.01 {
				continue
			}
			rect := m.toastRect(i)
			if !rl.CheckCollisionPointRec(mouse, rect) {
				continue
			}
			if ar := m.toastActionRect(i, t, rect); ar.Width > 0 && rl.CheckCollisionPointRec(mouse, ar) {
				if t.onAction != nil {
					t.onAction()
				}
				startDismiss(t)
				break
			}
			if t.onClick != nil {
				t.onClick()
			}
			startDismiss(t)
			break
		}
	}

	// Phase 3 — remove toasts whose exit animation has finished.
	live := m.toasts[:0]
	for _, t := range m.toasts {
		if !t.dismissing {
			live = append(live, t)
			continue
		}
		// Keep while exit tween is running or while still visible.
		if (t.fadeTween != nil && t.fadeTween.IsActive) || t.alpha > 0.005 {
			live = append(live, t)
		}
	}
	m.toasts = live
}

// ─────────────────────────────────────────────────────────────────────────────
// Draw
// ─────────────────────────────────────────────────────────────────────────────

// Draw renders all active toasts to the screen. Call inside rl.BeginDrawing()
// after all other overlays (or order them by desired z-layering preference).
// Icons are uploaded to the GPU on the first call; subsequent calls are free.
func (m *ToastManager) Draw() {
	if len(m.toasts) == 0 {
		return
	}
	m.ensureIcons()
	for i, t := range m.toasts {
		if t.alpha < 0.005 {
			continue
		}
		m.drawOne(i, t)
	}
}

// Unload releases the GPU icon textures. Call once on application shutdown
// (e.g. via defer).
func (m *ToastManager) Unload() {
	for i := range m.icons {
		FreeIconTexture(m.icons[i])
		m.icons[i] = rl.Texture2D{}
	}
	m.iconsLoaded = false
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────────────────────────────────────

// toastRect returns the screen rectangle for the toast at m.toasts[idx].
// Newest toast (last index) is anchored at the bottom-right; older toasts
// stack upward.
// toastActionRect is the hit target for the optional trailing action button.
func (m *ToastManager) toastActionRect(idx int, t *toast, rect rl.Rectangle) rl.Rectangle {
	if t.actionLabel == "" || t.onAction == nil {
		return rl.Rectangle{}
	}
	style := GetThemeStyle("toast-action")
	w := float32(measureTextS(t.actionLabel, style)) + 20
	if w < 56 {
		w = 56
	}
	if w > rect.Width*0.45 {
		w = rect.Width * 0.45
	}
	return rl.NewRectangle(rect.X+rect.Width-w-8, rect.Y+8, w, rect.Height-16)
}

// toastCardHeight derives card height from effective label/body sizes + padding.
func toastCardHeight() float32 {
	labelFS := EffectiveFontSize(GetThemeStyle("toast-label"))
	bodyFS := EffectiveFontSize(GetThemeStyle("toast-body"))
	h := toastPadY + labelFS + 4 + bodyFS + toastPadY + toastProgressH + 4
	if h < toastMinH {
		h = toastMinH
	}
	return h
}

func (m *ToastManager) toastRect(idx int) rl.Rectangle {
	n := len(m.toasts)
	// stackPos 0 = bottom (newest), increasing = higher up the screen.
	stackPos := float32(n - 1 - idx)
	cardH := toastCardHeight()
	w := float32(m.winW)
	h := float32(m.winH)
	if w < toastW+16 {
		w = toastW + 16
	}
	if h < cardH+toastMarginB+16 {
		h = cardH + toastMarginB + 16
	}
	baseX := w - toastW - toastMarginR
	if baseX < 12 {
		baseX = 12
	}
	baseY := h - toastMarginB - cardH
	y := baseY - stackPos*(cardH+toastGap)
	if y < 8 {
		y = 8
	}
	return rl.NewRectangle(baseX, y, toastW, cardH)
}

func startToastFadeIn(t *toast) {
	t.slideTween = NewTween(0, 1, toastFadeInDur, EaseOutQuad,
		func(v float32) { t.alpha = v },
		func() { t.alpha = 1 })
}

// advanceToastTween steps a tween; on a large frame hitch, snap to the end so
// a single slow frame does not drag the fade across many visible steps.
func advanceToastTween(tw *Tween, dt float32) {
	if tw == nil || !tw.IsActive {
		return
	}
	if dt > tw.Duration*0.5 {
		tw.Update(tw.Duration)
		return
	}
	tw.Update(dt)
}

// startDismiss begins the fade-out exit animation on t.
func startDismiss(t *toast) {
	if t.dismissing {
		return
	}
	t.dismissing = true
	if t.slideTween != nil {
		t.slideTween.IsActive = false
	}
	startAlpha := t.alpha
	t.fadeTween = NewTween(0, 1, toastFadeOutDur, LinearEasing,
		func(v float32) { t.alpha = startAlpha * (1 - v) },
		func() { t.alpha = 0 })
}

// drawOne renders a single toast card.
func (m *ToastManager) drawOne(idx int, t *toast) {
	rect := m.toastRect(idx)
	pal := toastPalette[t.level]
	a := t.alpha

	applyA := func(c rl.Color) rl.Color {
		c.A = uint8(float32(c.A) * a)
		return c
	}

	cardH := rect.Height
	roundness := toastRadius / (cardH / 2)
	if roundness > 1 {
		roundness = 1
	}
	fading := t.alpha < 0.995 ||
		(t.slideTween != nil && t.slideTween.IsActive) ||
		(t.fadeTween != nil && t.fadeTween.IsActive)
	segs := int32(8)
	if fading {
		segs = 4
	}

	if !fading {
		rl.DrawRectangleRounded(
			rl.NewRectangle(rect.X+1, rect.Y+4, rect.Width+1, rect.Height),
			roundness, segs, applyA(rl.NewColor(0, 0, 0, 28)))
	}
	rl.DrawRectangleRounded(
		rl.NewRectangle(rect.X+2, rect.Y+2, rect.Width, rect.Height),
		roundness, segs, applyA(rl.NewColor(0, 0, 0, 18)))

	// ── Card background ───────────────────────────────────────────────────────
	rl.DrawRectangleRounded(rect, roundness, segs, applyA(pal.bg))
	if !fading {
		rl.DrawRectangleRoundedLinesEx(rect, roundness, segs, 1.5, applyA(pal.border))
	}

	// ── Left accent bar ───────────────────────────────────────────────────────
	accentRec := rl.NewRectangle(rect.X, rect.Y, toastAccentW, cardH)
	accentR := toastAccentW / (cardH / 2)
	if accentR > 1 {
		accentR = 1
	}
	rl.DrawRectangleRounded(accentRec, accentR, segs, applyA(pal.accent))

	// ── Icon ─────────────────────────────────────────────────────────────────
	iconTex := m.icons[t.level]
	iconX := rect.X + toastAccentW + toastPadX
	iconY := rect.Y + (cardH-toastIconSz)/2
	if iconTex.ID != 0 {
		scale := toastIconSz / float32(iconCtxSz)
		rl.DrawTextureEx(iconTex, rl.NewVector2(iconX, iconY), 0, scale, applyA(rl.White))
	}

	// ── Text ─────────────────────────────────────────────────────────────────
	textX := iconX + toastIconSz + 8
	labelY := rect.Y + toastPadY

	labelStyle := GetThemeStyle("toast-label")
	labelStyle.TextColor = applyA(pal.accent)
	labelFS := EffectiveFontSize(labelStyle)
	drawTextS(pal.label, int32(textX), int32(labelY), labelStyle)

	msgStyle := GetThemeStyle("toast-body")
	msgStyle.TextColor = applyA(msgStyle.TextColor)
	msgY := labelY + labelFS + 4
	msgMaxW := rect.Width - (textX - rect.X) - toastPadX
	if t.actionLabel != "" {
		ar := m.toastActionRect(idx, t, rect)
		msgMaxW = ar.X - textX - 4
	}
	if msgMaxW < 40 {
		msgMaxW = 40
	}
	drawTextS(truncateTextS(t.message, msgMaxW, msgStyle), int32(textX), int32(msgY), msgStyle)

	// ── Action button ─────────────────────────────────────────────────────────
	if t.actionLabel != "" && t.onAction != nil {
		ar := m.toastActionRect(idx, t, rect)
		actionStyle := GetThemeStyle("toast-action")
		actionStyle.TextColor = applyA(actionStyle.TextColor)
		aw := float32(measureTextS(t.actionLabel, actionStyle))
		ax := int32(ar.X + (ar.Width-aw)/2)
		ay := TextPosY(ar, actionStyle)
		mouse := rl.GetMousePosition()
		if rl.CheckCollisionPointRec(mouse, ar) {
			rl.DrawRectangleRounded(ar, 0.25, 6, applyA(rl.NewColor(0, 0, 0, 14)))
		}
		drawTextS(t.actionLabel, ax, ay, actionStyle)
	}

	// ── Progress bar (timed toasts only) ─────────────────────────────────────
	if t.duration > 0 && !t.dismissing {
		barY := rect.Y + cardH - toastProgressH - 2
		barX := rect.X + toastRadius
		barW := rect.Width - 2*toastRadius
		// Track.
		rl.DrawRectangleRounded(
			rl.NewRectangle(barX, barY, barW, toastProgressH),
			1.0, 4, applyA(rl.NewColor(0, 0, 0, 18)))
		// Fill — depletes left-to-right as time counts down.
		ratio := t.remaining / t.duration
		if ratio < 0 {
			ratio = 0
		}
		fillW := barW * ratio
		if fillW > 1 {
			rl.DrawRectangleRounded(
				rl.NewRectangle(barX, barY, fillW, toastProgressH),
				1.0, 4, applyA(pal.progressColor))
		}
	}

	// ── Hover highlight for clickable toasts ─────────────────────────────────
	if t.onClick != nil {
		mouse := rl.GetMousePosition()
		if rl.CheckCollisionPointRec(mouse, rect) {
			rl.DrawRectangleRounded(rect, roundness, segs, rl.NewColor(0, 0, 0, 12))
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Icon pre-rendering
// ─────────────────────────────────────────────────────────────────────────────

// ensureIcons uploads the gg-rendered icon textures on the first Draw call.
// Using a lazy approach ensures the OpenGL context is always ready.
func (m *ToastManager) ensureIcons() {
	if m.iconsLoaded {
		return
	}
	m.icons[ToastInfo] = renderInfoIcon()
	m.icons[ToastSuccess] = renderSuccessIcon()
	m.icons[ToastWarning] = renderWarningIcon()
	m.icons[ToastError] = renderErrorIcon()
	m.iconsLoaded = true
}

// renderInfoIcon draws a circle with a white 'i' symbol.
func renderInfoIcon() rl.Texture2D {
	sz := iconCtxSz
	gc := NewIconContext(sz)
	s := float64(sz)
	// Filled blue circle.
	gc.SetRGBA(0.231, 0.510, 0.965, 1.0)
	gc.DrawCircle(s/2, s/2, s*0.44)
	gc.Fill()
	// White dot (top).
	gc.SetRGBA(1, 1, 1, 1)
	gc.DrawCircle(s/2, s*0.32, s*0.07)
	gc.Fill()
	// White vertical bar (body of 'i').
	gc.SetLineWidth(s * 0.13)
	gc.SetLineCap(gg.LineCapRound)
	gc.DrawLine(s/2, s*0.46, s/2, s*0.72)
	gc.Stroke()
	return ContextToTexture(gc)
}

// renderSuccessIcon draws a circle with a white checkmark.
func renderSuccessIcon() rl.Texture2D {
	sz := iconCtxSz
	gc := NewIconContext(sz)
	s := float64(sz)
	// Green circle.
	gc.SetRGBA(0.133, 0.773, 0.369, 1.0)
	gc.DrawCircle(s/2, s/2, s*0.44)
	gc.Fill()
	// White checkmark.
	gc.SetRGBA(1, 1, 1, 1)
	gc.SetLineWidth(s * 0.13)
	gc.SetLineCap(gg.LineCapRound)
	gc.SetLineJoin(gg.LineJoinRound)
	gc.MoveTo(s*0.26, s*0.52)
	gc.LineTo(s*0.44, s*0.70)
	gc.LineTo(s*0.74, s*0.32)
	gc.Stroke()
	return ContextToTexture(gc)
}

// renderWarningIcon draws a triangle with a white '!' inside.
func renderWarningIcon() rl.Texture2D {
	sz := iconCtxSz
	gc := NewIconContext(sz)
	s := float64(sz)
	// Amber triangle.
	gc.SetRGBA(0.961, 0.620, 0.043, 1.0)
	gc.MoveTo(s*0.50, s*0.07)
	gc.LineTo(s*0.95, s*0.88)
	gc.LineTo(s*0.05, s*0.88)
	gc.ClosePath()
	gc.Fill()
	// White exclamation mark body.
	gc.SetRGBA(1, 1, 1, 1)
	gc.SetLineWidth(s * 0.13)
	gc.SetLineCap(gg.LineCapRound)
	gc.DrawLine(s/2, s*0.35, s/2, s*0.62)
	gc.Stroke()
	// White exclamation dot.
	gc.DrawCircle(s/2, s*0.75, s*0.07)
	gc.Fill()
	return ContextToTexture(gc)
}

// renderErrorIcon draws a circle with a white '×' symbol.
func renderErrorIcon() rl.Texture2D {
	sz := iconCtxSz
	gc := NewIconContext(sz)
	s := float64(sz)
	// Red circle.
	gc.SetRGBA(0.937, 0.267, 0.267, 1.0)
	gc.DrawCircle(s/2, s/2, s*0.44)
	gc.Fill()
	// White ×.
	gc.SetRGBA(1, 1, 1, 1)
	gc.SetLineWidth(s * 0.13)
	gc.SetLineCap(gg.LineCapRound)
	pad := s * 0.28
	gc.DrawLine(pad, pad, s-pad, s-pad)
	gc.Stroke()
	gc.DrawLine(s-pad, pad, pad, s-pad)
	gc.Stroke()
	return ContextToTexture(gc)
}
