// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── BottomSheet ─────────────────────────────────────────────────────────────
//
// BottomSheetMgr provides a slide-up panel from the bottom of the screen.
// Animation, scrim, backdrop, and Escape use OverlayHost (C5).
//
//	ui.ShowBottomSheet(content, 320)   // fixed height in px
//	ui.ShowBottomSheet(content, 0.45)  // fraction of window height when 0 < h <= 1
//	ui.CloseBottomSheet()
//
// Drive from main.go:
//
//	ui.BottomSheetMgr.Update(dt)
//	ui.BottomSheetMgr.Draw() // inside drawScreenOverlays
//
// Styled via "bottomsheet" theme key.

const (
	bottomSheetHandleW   float32 = 40
	bottomSheetHandleH   float32 = 4
	bottomSheetHandleTop float32 = 10
	bottomSheetChromeH   float32 = 28 // handle + top padding before content
)

type bottomSheetManager struct {
	host *OverlayHost

	content Node
	height  float32 // px, or (0,1] as fraction of screen
	CloseOnBackdrop bool
	CloseOnEscape   bool
	dragging         bool
	dragStartY       float32
	dragBaseProgress float32
}

// BottomSheetMgr is the package-level bottom sheet manager.
var BottomSheetMgr = &bottomSheetManager{
	host:            DefaultOverlayHost(OverlayAnimSlideBottom),
	CloseOnBackdrop: true,
	CloseOnEscape:   true,
}

func (b *bottomSheetManager) syncHostOptions() {
	b.host.CloseOnBackdrop = b.CloseOnBackdrop
	b.host.CloseOnEscape = b.CloseOnEscape
}

// ShowBottomSheet opens a bottom sheet. height: pixels, or fraction in (0,1], or 0 for 40% of screen.
func ShowBottomSheet(content Node, height float32) {
	if content == nil {
		return
	}
	BottomSheetMgr.content = content
	BottomSheetMgr.height = height
	BottomSheetMgr.beginOpen()
}

func (b *bottomSheetManager) beginOpen() {
	b.syncHostOptions()
	b.host.BeginOpen()
	b.dragging = false
}

// CloseBottomSheet begins the slide-down animation.
func CloseBottomSheet() {
	BottomSheetMgr.beginClose()
}

func (b *bottomSheetManager) beginClose() {
	if !b.host.Open {
		return
	}
	b.host.BeginClose()
	b.dragging = false
}

func (b *bottomSheetManager) reset() {
	b.host.Reset()
	b.dragging = false
}

// IsBottomSheetOpen reports logical open state.
func IsBottomSheetOpen() bool { return BottomSheetMgr.host.IsOpen() }

// IsBottomSheetVisible reports render visibility including close animation.
func IsBottomSheetVisible() bool { return BottomSheetMgr.host.Open }

func (b *bottomSheetManager) IsAnimating() bool {
	if !b.host.Open {
		return false
	}
	return b.host.IsAnimating() || b.dragging
}

func (b *bottomSheetManager) sheetHeight(sh float32) float32 {
	h := b.height
	if h <= 0 {
		return sh * 0.4
	}
	if h > 0 && h <= 1 {
		return sh * h
	}
	if h > sh-48 {
		h = sh - 48
	}
	if h < 120 {
		h = 120
	}
	return h
}

func (b *bottomSheetManager) sheetRect(sw, sh float32) rl.Rectangle {
	sheetH := b.sheetHeight(sh)
	return b.host.SlideSheetRect(sw, sh, sheetH)
}

// ContentBandHitRect is the area that should block scene pointer hit-tests.
func (b *bottomSheetManager) ContentBandHitRect(sw, sh float32) rl.Rectangle {
	return b.host.ContentBand(sw, sh)
}

func (b *bottomSheetManager) handleRect(sheet rl.Rectangle) rl.Rectangle {
	return rl.NewRectangle(
		sheet.X+(sheet.Width-bottomSheetHandleW)/2,
		sheet.Y+bottomSheetHandleTop,
		bottomSheetHandleW,
		bottomSheetHandleH,
	)
}

// Update advances animation and handles input (backdrop, escape, drag-to-dismiss on handle).
func (b *bottomSheetManager) Update(dt float32) {
	if !b.host.Open {
		return
	}
	b.syncHostOptions()

	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	sheet := b.sheetRect(sw, sh)
	sheetH := b.sheetHeight(sh)

	if b.dragging {
		mouse := rl.GetMousePosition()
		dy := mouse.Y - b.dragStartY
		b.host.Progress = b.dragBaseProgress - dy/sheetH
		if b.host.Progress < 0 {
			b.host.Progress = 0
		}
		if b.host.Progress > 1 {
			b.host.Progress = 1
		}
		if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
			b.dragging = false
			if b.host.Progress < 0.35 {
				b.beginClose()
				return
			}
			b.host.Closing = false
			b.host.LogicallyOpen = true
		}
	} else {
		b.host.AdvanceAnimation(dt)
		if !b.host.Open {
			return
		}
	}

	if !b.host.InputReady() {
		b.host.TickSkipFrames()
		return
	}
	if b.host.HandleEscape(b.beginClose) {
		return
	}

	mouse := rl.GetMousePosition()
	inside := rl.CheckCollisionPointRec(mouse, sheet)
	handle := b.handleRect(sheet)

	if !b.dragging && rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, handle) {
		b.dragging = true
		b.dragStartY = mouse.Y
		b.dragBaseProgress = b.host.Progress
	}

	if b.host.HandleBackdrop(sheet, sheet, b.beginClose) {
		return
	}

	if inside && b.content != nil {
		inner := b.contentInnerRect(sheet)
		layoutOverlaySubtree(b.content, inner)
		b.content.Update(dt)
	}
}

func (b *bottomSheetManager) contentInnerRect(sheet rl.Rectangle) rl.Rectangle {
	pad := GetThemeStyle("bottomsheet").Padding
	contentY := sheet.Y + bottomSheetChromeH
	inner := rl.NewRectangle(sheet.X+pad, contentY, sheet.Width-2*pad, sheet.Height-bottomSheetChromeH-pad)
	if inner.Height < 1 {
		inner.Height = 1
	}
	return inner
}

// Draw renders scrim + sheet. Call from drawScreenOverlays.
func (b *bottomSheetManager) Draw() {
	if !b.host.Open || b.host.Progress <= 0 {
		return
	}
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	a := b.host.Opacity()
	b.host.DrawScrimFull(sw, sh)

	sheet := b.sheetRect(sw, sh)
	style := GetThemeStyle("bottomsheet")
	bg := style.BackgroundColor
	bg.A = uint8(float32(bg.A) * a)
	const roundness = float32(0.08)
	const segments = int32(8)
	rl.DrawRectangleRounded(sheet, roundness, segments, bg)
	bc := style.BorderColor
	bc.A = uint8(float32(bc.A) * a)
	rl.DrawRectangle(int32(sheet.X), int32(sheet.Y), int32(sheet.Width), 1, bc)

	handle := b.handleRect(sheet)
	hc := rl.NewColor(180, 184, 200, uint8(float32(220)*a))
	rl.DrawRectangleRounded(handle, 1, 6, hc)

	if b.content != nil && !b.host.Closing {
		drawOverlaySubtree(b.content, b.contentInnerRect(sheet))
	}
}

// DebugInfo returns inspector status text.
func (b *bottomSheetManager) DebugInfo() string {
	if !b.host.Open {
		return "bottomsheet: closed"
	}
	state := "open"
	if b.host.Closing {
		state = "closing"
	}
	return fmt.Sprintf("bottomsheet: %s  progress:%.2f", state, b.host.Progress)
}
