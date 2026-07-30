// Package ui (continued)
package ui

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Pointer input latch — raylib's IsMouseButtonPressed is only true for one poll.
// At DeepIdleFPS (10) taps can fall between frames; we keep a pending click from
// press until a widget consumes it or the gesture goes stale. Input/click wakes
// bump target FPS to ActiveFPS before the next interaction.
var (
	ptrLeftWasDown      bool
	ptrClickPending     bool
	ptrClickPos         rl.Vector2
	ptrClickUsed        bool
	ptrClickPendingSince time.Time
	scenePointerBlocked bool

	// Focus handoff click — reserved at press time, consumed only by RouteScenePointerFocus
	// so other widgets cannot clear the latch before post-layout hit testing (§5).
	focusHandoffPending bool
	focusHandoffPos     rl.Vector2
)

const pointerClickStaleDuration = 750 * time.Millisecond

func latchPointerClick(pos rl.Vector2) {
	ptrClickPending = true
	ptrClickPos = pos
	ptrClickUsed = false
	ptrClickPendingSince = time.Now()
	focusHandoffPending = true
	focusHandoffPos = pos
}

func clearPointerClickLatch() {
	ptrClickPending = false
	ptrClickUsed = false
	ptrClickPendingSince = time.Time{}
}

func expireStalePointerClick(now time.Time) {
	if !ptrClickPending || ptrClickUsed || ptrLeftWasDown {
		return
	}
	if ptrClickPendingSince.IsZero() {
		return
	}
	if now.Sub(ptrClickPendingSince) > pointerClickStaleDuration {
		clearPointerClickLatch()
		clearFocusHandoffClick()
	}
}

// PeekFocusHandoffClick returns the reserved press position for web/native focus routing.
func PeekFocusHandoffClick() (rl.Vector2, bool) {
	if !focusHandoffPending {
		return rl.Vector2{}, false
	}
	return focusHandoffPos, true
}

func clearFocusHandoffClick() {
	focusHandoffPending = false
}

// PreparePointerInput must run once per frame at the top of the main loop,
// before layout/Update, so widgets can use PointerClickConsume.
func PreparePointerInput() {
	now := time.Now()
	mouse := rl.GetMousePosition()
	down := rl.IsMouseButtonDown(rl.MouseLeftButton)
	pressed := rl.IsMouseButtonPressed(rl.MouseLeftButton)

	if pressed || (!ptrLeftWasDown && down) {
		latchPointerClick(mouse)
	}
	// Do not clear pending on release — widgets consume on the next frame at idle FPS.
	ptrLeftWasDown = down
	expireStalePointerClick(now)
}

// PointerClickPending reports an unhandled primary click gesture in progress.
func PointerClickPending() bool { return ptrClickPending && !ptrClickUsed }

// PointerClickPosition is the latched press position for the current gesture.
func PointerClickPosition() rl.Vector2 { return ptrClickPos }

// PointerInputActive is true while the primary button is down or a click is pending.
func PointerInputActive() bool {
	return ptrClickPending || rl.IsMouseButtonDown(rl.MouseLeftButton)
}

// PointerClickMarkUsed marks the latched gesture handled (e.g. after a toggle down-edge).
func PointerClickMarkUsed() { clearPointerClickLatch() }

// PointerClickHandled reports whether an overlay or menu consumed the current click.
func PointerClickHandled() bool { return ptrClickUsed }

// SetScenePointerBlocked suppresses PointerClickConsume in the document tree while
// transient overlays handle input (drawer, bottom sheet, modal, etc.).
func SetScenePointerBlocked(block bool) { scenePointerBlocked = block }

// ScenePointerBlocked reports whether the document tree should ignore latched clicks.
func ScenePointerBlocked() bool { return scenePointerBlocked }

// PointerClickConsume reports whether this frame delivered a latched click inside hit.
func PointerClickConsume(hit rl.Rectangle) bool {
	if scenePointerBlocked {
		return false
	}
	if !ptrClickPending || ptrClickUsed || hit.Width < 1 || hit.Height < 1 {
		return false
	}
	if !rl.CheckCollisionPointRec(ptrClickPos, hit) {
		return false
	}
	clearPointerClickLatch()
	return true
}

// PointerClickSuppressTileBody marks the click handled when the cursor is on the
// locked row body but not on switchHit. Uses the latched press position when
// pending so deep-idle FPS can still suppress stray body hits.
func PointerClickSuppressTileBody(tile, switchHit rl.Rectangle) {
	if PointerClickHandled() {
		return
	}
	var pt rl.Vector2
	if PointerClickPending() {
		pt = ptrClickPos
	} else if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
		return
	} else {
		pt = rl.GetMousePosition()
	}
	if tile.Width < 1 || tile.Height < 1 {
		return
	}
	if !rl.CheckCollisionPointRec(pt, tile) {
		return
	}
	if switchHit.Width >= 1 && switchHit.Height >= 1 &&
		rl.CheckCollisionPointRec(pt, switchHit) {
		return
	}
	clearPointerClickLatch()
}

// PointerInside is true when the cursor is over hit (for hover chrome).
func PointerInside(hit rl.Rectangle) bool {
	if hit.Width < 1 || hit.Height < 1 {
		return false
	}
	return rl.CheckCollisionPointRec(rl.GetMousePosition(), hit)
}
