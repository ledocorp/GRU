//go:build android

package ui

/*
#include <stdint.h>
*/
import "C"

import rl "github.com/gen2brain/raylib-go/raylib"

//export goImeCodepoint
func goImeCodepoint(cp int32) {
	if cp > 0 {
		PushIMECodepoint(rune(cp))
	}
}

//export goImeKey
func goImeKey(key int32) {
	const (
		androidKeyDel   = 67
		androidKeyEnter = 66
	)
	switch key {
	case androidKeyDel:
		PushIMEKey(int32(rl.KeyBackspace))
	case androidKeyEnter:
		PushIMEKey(int32(rl.KeyEnter))
	default:
		PushIMEKey(key)
	}
}

//export goImeHeight
func goImeHeight(px int32) {
	prev := int32(IMEInsetHeight())
	SetIMEInsetHeight(px)
	if px != prev {
		rl.TraceLog(rl.LogWarning, "gru ime height=%d", px)
	}
}

func nodeWantsIME(n Node) bool {
	if n == nil {
		return false
	}
	switch n.(type) {
	case *TextEditor, *TextInput:
		return true
	default:
		return false
	}
}

// imeArmed is set after a user pointer gesture so Build-time SetFocus does not
// raise the keyboard (which blanked the NativeActivity GLES surface).
var imeArmed bool

// ArmIMEForUserGesture allows the next text focus to show the soft keyboard.
func ArmIMEForUserGesture() { imeArmed = true }

func syncIMEForFocus(n Node) {
	if !AndroidGLES() {
		return
	}
	if nodeWantsIME(n) {
		if imeArmed {
			PlatformShowSoftInput()
		}
		return
	}
	PlatformHideSoftInput()
}
