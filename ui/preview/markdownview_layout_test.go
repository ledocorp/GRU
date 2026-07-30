package preview

import (
	"testing"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestMarkdownViewFirstRealWidthRemeasures(t *testing.T) {
	mv := NewMarkdownView("mv-w", 0, 0, 0, 0)
	mv.SetMarkdown("# Hello\n\nA short paragraph that should wrap once width is known.\n")
	// Zero-width layout (pre-flex) must not lock layoutSyncedGen forever.
	mv.SetBounds(rl.NewRectangle(0, 0, 0, 400))
	mv.Layout()
	if mv.layoutSyncedGen == mv.buildGen && mv.lastLayoutW <= 1 {
		// Either not synced yet, or synced only after real width — both OK if next pass reflows.
	}
	mv.SetBounds(rl.NewRectangle(0, 0, 420, 600))
	mv.Layout()
	if mv.lane == nil || len(mv.lane.Children()) == 0 {
		t.Fatal("expected built lane")
	}
	h, ok := mv.lane.Children()[0].(*ui.RichText)
	if !ok {
		t.Fatalf("first block = %T", mv.lane.Children()[0])
	}
	if h.Bounds().Height < 8 {
		t.Fatalf("heading height %.1f after first real width — still unmeasured", h.Bounds().Height)
	}
	if mv.lastLayoutW < 400 {
		t.Fatalf("lastLayoutW=%.1f, want ~420", mv.lastLayoutW)
	}
}

func TestMarkdownViewHeadingRebuildUpdates(t *testing.T) {
	mv := NewMarkdownView("mv", 0, 0, 360, 600)
	mv.SetBounds(rl.NewRectangle(0, 0, 360, 600))
	mv.SetMarkdown("# Title\n")
	mv.Layout()

	h1, ok := mv.lane.Children()[0].(*ui.RichText)
	if !ok {
		t.Fatalf("block = %T, want heading RichText", mv.lane.Children()[0])
	}
	if h1.Spans[0].Variant != "h1" {
		t.Fatalf("variant = %q, want h1", h1.Spans[0].Variant)
	}
	h1H := h1.Bounds().Height
	if h1H < 8 {
		t.Fatalf("h1 height %.1f, want measured heading", h1H)
	}

	mv.SetMarkdown("## Title\n")
	mv.Layout()

	h2, ok := mv.lane.Children()[0].(*ui.RichText)
	if !ok {
		t.Fatalf("block = %T, want heading RichText", mv.lane.Children()[0])
	}
	if h2.Spans[0].Variant != "h2" {
		t.Fatalf("variant = %q, want h2 after rebuild", h2.Spans[0].Variant)
	}
	if h2.Bounds().Height >= h1H {
		t.Fatalf("h2 height %.1f should be less than h1 %.1f", h2.Bounds().Height, h1H)
	}
}

func TestMarkdownTableCellsWidenWithHost(t *testing.T) {
	src := "| Alpha | Beta |\n|---|---|\n| one | two |"
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d, want 1 table", len(nodes))
	}
	card, ok := nodes[0].(*ui.Card)
	if !ok {
		t.Fatalf("node = %T, want Card", nodes[0])
	}

	layout := func(w float32) float32 {
		card.InvalidateLayoutPassCache()
		card.MarkDirty()
		card.SetBounds(rl.NewRectangle(0, 0, w, 320))
		card.Layout()
		return tableFirstCellWidth(card)
	}

	wNarrow := layout(260)
	wWide := layout(460)
	if wWide <= wNarrow+4 {
		t.Fatalf("cell width narrow=%.1f wide=%.1f, want wider cells in wider host", wNarrow, wWide)
	}
}

func tableFirstCellWidth(card *ui.Card) float32 {
	scroll, ok := card.Children()[0].(*ui.Viewport)
	if !ok || len(scroll.Children()) == 0 {
		return 0
	}
	lane, ok := scroll.Children()[0].(*ui.Container)
	if !ok || len(lane.Children()) == 0 {
		return 0
	}
	row, ok := lane.Children()[0].(*ui.Container)
	if !ok || len(row.Children()) == 0 {
		return 0
	}
	for _, ch := range row.Children() {
		if rt, ok := ch.(*ui.RichText); ok {
			return rt.Bounds().Width
		}
	}
	return 0
}
