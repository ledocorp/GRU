// Package ui (continued)
package ui

// ResetTransientOverlays immediately dismisses slide/fade overlays when switching
// scenes so stale content and hit targets do not leak across demos.
func ResetTransientOverlays() {
	DrawerMgr.reset()
	BottomSheetMgr.reset()
	ContextMenuMgr.open = false
	ContextMenuMgr.hovered = -1
	if ModalMgr.host.Open {
		ModalMgr.host.Reset()
	}
	CommandPaletteMgr.reset()
	NotificationCenterMgr.reset()
	DatePickerMgr.snapClose()
	DateRangePickerMgr.snapClose()
}

// WebViewHostOccluded reports whether a blocking overlay should hide live WebView2
// HWND surfaces (modal, drawer, bottom sheet). Context menus and the title bar do
// not occlude the host — they are native raylib layers beside the panel (§5).
func WebViewHostOccluded() bool {
	return IsDrawerVisible() || IsBottomSheetVisible() || IsModalVisible()
}

// OverlayBlocksSceneInput is true while a modal overlay owns pointer input so
// clicks do not fall through to the scene underneath.
func OverlayBlocksSceneInput() bool {
	if IsDrawerVisible() || IsBottomSheetVisible() || IsModalVisible() {
		return true
	}
	if ContextMenuMgr.open {
		return true
	}
	if IsCommandPaletteVisible() {
		return true
	}
	if IsNotificationCenterVisible() {
		return true
	}
	if CalendarPopupOpen() {
		return true
	}
	return false
}

// OverlayIdleBlockers keeps ActiveFPS while transient overlays are on screen so
// fade animations and progress bars do not run at AnimationFPS/DeepIdleFPS.
func OverlayIdleBlockers() WakeReason {
	var r WakeReason
	if ActiveToastCount() > 0 {
		r |= WakeOverlay
	}
	if IsNotificationCenterVisible() {
		r |= WakeOverlay
	}
	return r
}

// OverlayAnimationWake returns wake reasons for overlay managers that are
// mid-animation (fade, slide, close). Toasts use fade-only tweens and rely on
// OverlayIdleBlockers for FPS; they are not listed here.
func OverlayAnimationWake() WakeReason {
	var r WakeReason
	if ModalMgr.IsAnimating() {
		r |= WakeAnimation
	}
	if DrawerMgr.IsAnimating() {
		r |= WakeAnimation
	}
	if BottomSheetMgr.IsAnimating() {
		r |= WakeAnimation
	}
	if IsDrawerVisible() {
		r |= WakeOverlay
	}
	if IsBottomSheetVisible() {
		r |= WakeOverlay
	}
	if ContextMenuMgr.IsAnimating() {
		r |= WakeAnimation
	}
	if CommandPaletteMgr.IsAnimating() {
		r |= WakeAnimation
	}
	if NotificationCenterMgr.IsAnimating() {
		r |= WakeAnimation
	}
	if IsCommandPaletteVisible() {
		r |= WakeOverlay
	}
	if IsNotificationCenterVisible() {
		r |= WakeOverlay
	}
	if ColorPickerMgr.IsAnimating() {
		r |= WakeAnimation
	}
	if DatePickerMgr.IsAnimating() {
		r |= WakeAnimation
	}
	if DateRangePickerMgr.IsAnimating() {
		r |= WakeAnimation
	}
	if Tooltips.IsAnimating() {
		r |= WakeAnimation
	}
	return r
}
