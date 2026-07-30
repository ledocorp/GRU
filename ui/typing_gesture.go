// Package ui (continued)
package ui

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// typingGestureWindow keeps ActiveFPS briefly after the last editor keystroke so
// deep idle (10 FPS) does not miss fast typing or key repeat between frames.
const typingGestureWindow = 900 * time.Millisecond

var typingGestureUntil time.Time

// NoteTypingGesture extends the typing hold window. Call when a focused TextEditor
// accepts keyboard input (insert, delete, navigation, paste, etc.).
func NoteTypingGesture() {
	typingGestureUntil = time.Now().Add(typingGestureWindow)
}

// TypingGestureActive reports whether editor typing recently occurred.
func TypingGestureActive() bool {
	return time.Now().Before(typingGestureUntil)
}

// SampleEditorKeyboardWake detects focused TextEditor / TextInput / SearchBar keys at the top
// of the frame without consuming GetCharPressed (those widgets drain that queue
// during Update).
func SampleEditorKeyboardWake(root Node) WakeSummary {
	if root == nil || !rl.IsWindowFocused() {
		return WakeSummary{}
	}
	if findFocusedTextEditor(root) == nil &&
		findFocusedTextInput(root) == nil &&
		findFocusedSearchBar(root) == nil {
		return WakeSummary{}
	}
	if !editorTypingKeysActive() {
		return WakeSummary{}
	}
	var out WakeSummary
	out.Add(WakeKeyboard, "text-key")
	NoteTypingGesture()
	return out
}

func findFocusedTextInput(n Node) *TextInput {
	if n == nil || n.IsHidden() {
		return nil
	}
	if ti, ok := n.(*TextInput); ok && ti.IsFocused() {
		return ti
	}
	for _, ch := range n.Children() {
		if ti := findFocusedTextInput(ch); ti != nil {
			return ti
		}
	}
	return nil
}

func findFocusedSearchBar(n Node) *SearchBar {
	if n == nil || n.IsHidden() {
		return nil
	}
	if sb, ok := n.(*SearchBar); ok && sb.IsFocused() {
		return sb
	}
	for _, ch := range n.Children() {
		if sb := findFocusedSearchBar(ch); sb != nil {
			return sb
		}
	}
	return nil
}

func editorTypingKeysActive() bool {
	if editorTypingKeyFired(rl.KeyBackspace) ||
		editorTypingKeyFired(rl.KeyDelete) ||
		editorTypingKeyFired(rl.KeyEnter) ||
		editorTypingKeyFired(rl.KeyTab) ||
		editorTypingKeyFired(rl.KeySpace) ||
		editorTypingKeyFired(rl.KeyLeft) ||
		editorTypingKeyFired(rl.KeyRight) ||
		editorTypingKeyFired(rl.KeyUp) ||
		editorTypingKeyFired(rl.KeyDown) ||
		editorTypingKeyFired(rl.KeyHome) ||
		editorTypingKeyFired(rl.KeyEnd) ||
		editorTypingKeyFired(rl.KeyPageUp) ||
		editorTypingKeyFired(rl.KeyPageDown) {
		return true
	}
	for k := int32(rl.KeyA); k <= int32(rl.KeyZ); k++ {
		if editorTypingKeyFired(k) {
			return true
		}
	}
	for k := int32(rl.KeyZero); k <= int32(rl.KeyNine); k++ {
		if editorTypingKeyFired(k) {
			return true
		}
	}
	for _, k := range []int32{
		rl.KeyApostrophe, rl.KeyComma, rl.KeyMinus, rl.KeyPeriod, rl.KeySlash,
		rl.KeySemicolon, rl.KeyEqual, rl.KeyLeftBracket, rl.KeyBackSlash,
		rl.KeyRightBracket, rl.KeyGrave,
	} {
		if editorTypingKeyFired(k) {
			return true
		}
	}
	if textInputCtrlDown() {
		for _, k := range []int32{rl.KeyZ, rl.KeyY, rl.KeyC, rl.KeyX, rl.KeyV, rl.KeyA} {
			if rl.IsKeyPressed(k) {
				return true
			}
		}
	}
	return false
}

func editorTypingKeyFired(key int32) bool {
	return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key)
}
