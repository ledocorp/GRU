package ui

// PresentHook, when set by main, redraws the current frame (cache blit + overlays).
// Used before blocking OS dialogs so context menus disappear immediately.
var PresentHook func()

// PresentScreen runs PresentHook when registered.
func PresentScreen() {
	if PresentHook != nil {
		PresentHook()
	}
}
