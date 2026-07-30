package ui

import (
	"testing"
	"time"
)

func TestTypingGestureActiveHoldWindow(t *testing.T) {
	typingGestureUntil = time.Time{}
	if TypingGestureActive() {
		t.Fatal("expected inactive initially")
	}
	NoteTypingGesture()
	if !TypingGestureActive() {
		t.Fatal("expected active immediately after note")
	}
	typingGestureUntil = time.Now().Add(-time.Millisecond)
	if TypingGestureActive() {
		t.Fatal("expected inactive after window expires")
	}
}
