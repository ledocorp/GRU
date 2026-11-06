package ui

import (
	"sync"
	"sync/atomic"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Soft-IME codepoints (Android EditText proxy → JNI → here). TextEditor /
// TextInput drain these alongside rl.GetCharPressed.
var (
	imeMu    sync.Mutex
	imeRunes []rune
	imeKeys  []int32 // raylib key codes (e.g. rl.KeyBackspace)
	imeInset atomic.Int32 // keyboard overlap height in framebuffer px (adjustNothing)
)

// SetIMEInsetHeight records the soft-keyboard overlap (0 when hidden).
func SetIMEInsetHeight(px int32) {
	if px < 0 {
		px = 0
	}
	imeInset.Store(px)
}

// IMEInsetHeight returns the current soft-keyboard overlap in framebuffer pixels.
func IMEInsetHeight() float32 {
	return float32(imeInset.Load())
}

// PushIMECodepoint enqueues a Unicode scalar from the platform IME bridge.
func PushIMECodepoint(r rune) {
	if r == 0 {
		return
	}
	imeMu.Lock()
	imeRunes = append(imeRunes, r)
	imeMu.Unlock()
}

// PushIMEKey enqueues a key (e.g. rl.KeyBackspace) from the IME bridge.
func PushIMEKey(key int32) {
	if key == 0 {
		return
	}
	imeMu.Lock()
	imeKeys = append(imeKeys, key)
	imeMu.Unlock()
}

func nextIMECodepoint() rune {
	imeMu.Lock()
	defer imeMu.Unlock()
	if len(imeRunes) == 0 {
		return 0
	}
	r := imeRunes[0]
	imeRunes = imeRunes[1:]
	return r
}

func nextIMEKey() int32 {
	imeMu.Lock()
	defer imeMu.Unlock()
	if len(imeKeys) == 0 {
		return 0
	}
	k := imeKeys[0]
	imeKeys = imeKeys[1:]
	return k
}

// CharPressedOrIME returns the next character from raylib, then the IME queue.
func CharPressedOrIME() int32 {
	if c := rl.GetCharPressed(); c != 0 {
		return c
	}
	if r := nextIMECodepoint(); r != 0 {
		return int32(r)
	}
	return 0
}

// IMEKeyPressed reports whether key was delivered via the IME bridge this drain.
func IMEKeyPressed(key int32) bool {
	imeMu.Lock()
	defer imeMu.Unlock()
	for i, k := range imeKeys {
		if k == key {
			imeKeys = append(imeKeys[:i], imeKeys[i+1:]...)
			return true
		}
	}
	return false
}
