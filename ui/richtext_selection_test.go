package ui

import (
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestRichTextTokenAtLineXUsesCursorX(t *testing.T) {
	ensureTestFonts(t)
	rt := NewRichText("rt-line-x", []TextSpan{
		{Text: "alpha beta"},
	}, 0, 0, 200, 48)
	rt.Wrap = false
	rt.SetBounds(rl.NewRectangle(0, 0, 200, 48))
	rt.Layout()
	if w := measureTextS("alpha", rt.GetStyle()); w <= 0 {
		t.Skip("text measure unavailable in unit test")
	}

	firstIdx, secondIdx := -1, -1
	rt.forEachTokenRect(rt.Bounds(), func(tok richTextToken, rect rl.Rectangle) bool {
		if strings.TrimSpace(tok.text) == "" {
			return true
		}
		if firstIdx < 0 {
			firstIdx = tok.index
			return true
		}
		secondIdx = tok.index
		return true
	})
	if firstIdx < 0 || secondIdx < 0 {
		t.Fatal("failed to find two word tokens")
	}

	var firstRect, secondRect rl.Rectangle
	rt.forEachTokenRect(rt.Bounds(), func(tok richTextToken, rect rl.Rectangle) bool {
		switch tok.index {
		case firstIdx:
			firstRect = rect
		case secondIdx:
			secondRect = rect
		}
		return true
	})

	lineY := firstRect.Y + firstRect.Height/2

	if got := rt.tokenAtLineX(rl.NewVector2(firstRect.X+firstRect.Width*0.5, lineY)); got != firstIdx {
		t.Fatalf("tokenAtLineX on first word = %d, want %d", got, firstIdx)
	}
	if got := rt.tokenAtLineX(rl.NewVector2(secondRect.X+secondRect.Width*0.5, lineY)); got != secondIdx {
		t.Fatalf("tokenAtLineX on second word = %d, want %d", got, secondIdx)
	}
}

func TestRichTextSelectionRequiresDrag(t *testing.T) {
	rt := NewRichText("rt-anchor", []TextSpan{{Text: "one two"}}, 0, 0, 200, 40)
	rt.Selectable = true
	rt.Wrap = false

	rt.selectPress = true
	rt.selectAnchor = 0
	rt.selectStart = -1
	rt.selectEnd = -1
	if rt.hasSelection() {
		t.Fatal("click without drag should not create a selection range")
	}

	rt.selecting = true
	rt.selectStart = 0
	rt.selectEnd = 1
	if !rt.hasSelection() {
		t.Fatal("drag selection should highlight a range")
	}
}
