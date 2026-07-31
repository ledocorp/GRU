// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// overlayHitRects holds screen-space rectangles for floating popups drawn above
// the scene tree (calendars, dropdown lists, combobox lists). Refreshed each
// frame via [PrepareOverlayHitTest] before doc.Root.Update.
var overlayHitRects []rl.Rectangle

// PrepareOverlayHitTest collects popup bounds so widgets underneath ignore pointer
// events while the user interacts with a floating overlay.
func PrepareOverlayHitTest(root Node) {
	overlayHitRects = overlayHitRects[:0]
	if root == nil {
		return
	}
	if DatePickerMgr.IsOpen() {
		overlayHitRects = append(overlayHitRects, DatePickerMgr.PopupBounds())
		if DatePickerMgr.target != nil {
			overlayHitRects = append(overlayHitRects, DatePickerMgr.target.Bounds())
		}
	}
	if DateRangePickerMgr.IsOpen() {
		overlayHitRects = append(overlayHitRects, DateRangePickerMgr.PopupBounds())
		if DateRangePickerMgr.target != nil {
			overlayHitRects = append(overlayHitRects, DateRangePickerMgr.target.Bounds())
		}
	}
	for _, dd := range collectOpenDropdowns(root.Children()) {
		overlayHitRects = append(overlayHitRects, dd.PopupBounds())
		overlayHitRects = append(overlayHitRects, dd.Bounds())
	}
	for _, cb := range collectOpenComboBoxes(root.Children()) {
		overlayHitRects = append(overlayHitRects, cb.PopupBounds())
		overlayHitRects = append(overlayHitRects, cb.Bounds())
	}
	if IsModalVisible() {
		overlayHitRects = append(overlayHitRects, ModalMgr.panelHitRect())
	}
	// Drawer/sheet scrims live outside the document tree — block hover/clicks on
	// scene widgets underneath for the whole content band.
	if IsDrawerVisible() {
		sw := float32(rl.GetScreenWidth())
		sh := float32(rl.GetScreenHeight())
		overlayHitRects = append(overlayHitRects, DrawerMgr.ContentBandHitRect(sw, sh))
	}
	if IsBottomSheetVisible() {
		sw := float32(rl.GetScreenWidth())
		sh := float32(rl.GetScreenHeight())
		overlayHitRects = append(overlayHitRects, BottomSheetMgr.ContentBandHitRect(sw, sh))
	}
	for _, tb := range collectOpenToolbars(root.Children()) {
		if !tb.overflowOpen {
			continue
		}
		hidden := tb.hiddenItems()
		if len(hidden) == 0 {
			continue
		}
		overlayHitRects = append(overlayHitRects, tb.overflowPopupRect(hidden))
		overlayHitRects = append(overlayHitRects, tb.Bounds())
	}
}

// WidgetBlockedByOverlay reports whether an interactive widget should ignore
// pointer input this frame because a floating overlay sits above the cursor.
// Pass only floating popup rects owned by the calling widget via exemptRects —
// do not pass the widget's base bounds when closed, or overlapping popups will
// incorrectly fall through to widgets underneath.
func WidgetBlockedByOverlay(mouse rl.Vector2, exemptRects ...rl.Rectangle) bool {
	for _, r := range overlayHitRects {
		if r.Width <= 0 || r.Height <= 0 {
			continue
		}
		if !rl.CheckCollisionPointRec(mouse, r) {
			continue
		}
		exempt := false
		for _, e := range exemptRects {
			if e.Width > 0 && e.Height > 0 && rl.CheckCollisionPointRec(mouse, e) {
				exempt = true
				break
			}
		}
		if !exempt {
			return true
		}
	}
	return false
}
