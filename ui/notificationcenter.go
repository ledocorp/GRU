// Package ui (continued)
package ui

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ncWidth       = float32(360)
	ncPad         = float32(12)
	ncRowH        = float32(52)
	ncMaxHistory  = 40
	ncScrimAlpha  = float32(0.35)
	ncMargin      = float32(16)
)

// NotificationRecord is one history entry shown in the notification center.
type NotificationRecord struct {
	Message string
	Level   ToastLevel
	At      time.Time
}

var notificationHistory []NotificationRecord

func appendNotificationHistory(message string, level ToastLevel) {
	notificationHistory = append(notificationHistory, NotificationRecord{
		Message: message,
		Level:   level,
		At:      time.Now(),
	})
	if len(notificationHistory) > ncMaxHistory {
		notificationHistory = notificationHistory[len(notificationHistory)-ncMaxHistory:]
	}
}

// ClearNotificationHistory removes all stored entries.
func ClearNotificationHistory() {
	notificationHistory = nil
}

type notificationCenterManager struct {
	open, logicallyOpen, closing bool
	progress                    float32
	winW, winH                  int32
	scroll                      int
	skipFrames                  int
	CloseOnEscape               bool
	CloseOnBackdrop             bool
}

// NotificationCenterMgr is the package-level notification history panel.
var NotificationCenterMgr = &notificationCenterManager{
	CloseOnEscape:   true,
	CloseOnBackdrop: true,
}

// SetWindowSize configures overlay placement.
func (m *notificationCenterManager) SetWindowSize(w, h int32) {
	m.winW, m.winH = w, h
}

// IsNotificationCenterAnimating reports slide open/close tweens.
func (m *notificationCenterManager) IsAnimating() bool {
	if !m.open && !m.logicallyOpen {
		return false
	}
	return m.closing || m.progress < 0.999 || m.skipFrames > 0
}

// ShowNotificationCenter opens the history panel.
func ShowNotificationCenter() {
	m := NotificationCenterMgr
	m.open = true
	m.logicallyOpen = true
	m.closing = false
	m.progress = 0
	m.scroll = 0
	m.skipFrames = 2 // ignore opening click as backdrop dismiss (see ModalMgr)
	Wake(WakeOverlay, "notification-center")
}

// CloseNotificationCenter dismisses the panel.
func CloseNotificationCenter() {
	NotificationCenterMgr.closing = true
}

// IsNotificationCenterVisible reports whether the panel is open or animating.
func IsNotificationCenterVisible() bool {
	m := NotificationCenterMgr
	return m.open || m.logicallyOpen
}

func (m *notificationCenterManager) reset() {
	m.open = false
	m.logicallyOpen = false
	m.closing = false
	m.progress = 0
	m.skipFrames = 0
}

func (m *notificationCenterManager) Update(dt float32) {
	if !m.open && !m.logicallyOpen {
		return
	}
	if m.skipFrames > 0 {
		m.skipFrames--
	}
	const slide = float32(0.16)
	if m.closing {
		m.progress -= dt / slide
		if m.progress <= 0 {
			m.progress = 0
			m.open = false
			m.logicallyOpen = false
			m.closing = false
		}
		return
	}
	if m.progress < 1 {
		m.progress += dt / slide
		if m.progress > 1 {
			m.progress = 1
		}
	}
	if m.CloseOnEscape && rl.IsKeyPressed(rl.KeyEscape) {
		m.closing = true
	}
}

func (m *notificationCenterManager) panelRect() rl.Rectangle {
	sw := float32(m.winW)
	sh := float32(m.winH)
	if sw < 1 {
		sw = float32(rl.GetScreenWidth())
	}
	if sh < 1 {
		sh = float32(rl.GetScreenHeight())
	}
	band := OverlayContentBand(sw, sh)
	w := ncWidth
	if w > band.Width-ncMargin*2 {
		w = band.Width - ncMargin*2
	}
	if w < 200 {
		w = 200
	}
	x := band.X + band.Width - w - ncMargin + (1-m.progress)*w
	y := band.Y + ncMargin
	h := band.Height - 2*ncMargin
	if h < 120 {
		h = 120
	}
	return rl.NewRectangle(x, y, w, h)
}

func (m *notificationCenterManager) Draw() {
	if !m.open && m.progress <= 0 {
		return
	}
	if m.logicallyOpen && m.progress > 0 {
		sw := float32(m.winW)
		sh := float32(m.winH)
		if sw < 1 {
			sw = float32(rl.GetScreenWidth())
		}
		if sh < 1 {
			sh = float32(rl.GetScreenHeight())
		}
		band := OverlayContentBand(sw, sh)
		rl.DrawRectangleRec(band,
			rl.NewColor(0, 0, 0, uint8(ncScrimAlpha*255*m.progress)))
	}
	pr := m.panelRect()
	bg := rl.NewColor(255, 255, 255, 255)
	rl.DrawRectangleRounded(pr, 0.04, 8, bg)
	rl.DrawRectangleRoundedLinesEx(pr, 0.04, 8, 1, rl.NewColor(210, 212, 225, 255))

	title := GetThemeStyle("form-label")
	title.Bold = true
	drawTextS("Notifications", int32(pr.X+ncPad), int32(pr.Y+ncPad), title)

	rows := notificationHistory
	if len(rows) == 0 {
		hint := GetThemeStyle("form-value")
		drawTextS("No notifications yet — tap Fire toast first.", int32(pr.X+ncPad), int32(pr.Y+40), hint)
	} else {
		y := pr.Y + 36
		maxY := pr.Y + pr.Height - ncPad
		for i := len(rows) - 1 - m.scroll; i >= 0; i-- {
			if y+ncRowH > maxY {
				break
			}
			rec := rows[i]
			row := rl.NewRectangle(pr.X+ncPad, y, pr.Width-2*ncPad, ncRowH-4)
			pal := toastPalette[rec.Level]
			rl.DrawRectangleRounded(row, 0.12, 6, pal.bg)
			rl.DrawRectangleRec(rl.NewRectangle(row.X, row.Y, toastAccentW, row.Height), pal.accent)
			msgStyle := GetThemeStyle("default")
			msgStyle.FontSize = 13
			drawTextS(truncateTextS(rec.Message, row.Width-20, msgStyle), int32(row.X+10), int32(row.Y+8), msgStyle)
			lvl := GetThemeStyle("form-value")
			lvl.FontSize = 11
			drawTextS(pal.label, int32(row.X+10), int32(row.Y+26), lvl)
			y += ncRowH
		}
	}

	if m.skipFrames == 0 && m.CloseOnBackdrop && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if !rl.CheckCollisionPointRec(rl.GetMousePosition(), pr) {
			m.closing = true
		}
	}
}
