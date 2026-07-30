package ui

import (
	"fmt"
	"testing"
)

func TestPageViewportFillsShellNotContent(t *testing.T) {
	const w, h = float32(800), float32(600)
	shell := NewContainer("shell", 0, 0, w, h)
	shell.LayoutType = LayoutFlex
	shell.FlexDirection = FlexColumn
	shell.SetStyle("page-shell")

	vp := NewViewport("vp", 0, 0, 0, 0)
	vp.SetFlexGrow(1)

	hdr := NewHeader("hdr", "Demo", "Subtitle", 0, 0, 0, 64)
	grid := NewContainer("grid", 0, 0, 0, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 12
	grid.Gap = 12
	grid.SetFlexGrow(1)
	for i := 0; i < 6; i++ {
		p := NewPanel(fmt.Sprintf("p%d", i), fmt.Sprintf("Panel %d", i), 0, 0, 0, 180)
		p.SetColSpan(BreakpointXS, 12)
		grid.AddChild(p)
	}
	vp.AddChild(hdr)
	vp.AddChild(grid)
	shell.AddChild(vp)

	root := NewContainer("root", 0, 0, w, h)
	root.LayoutType = LayoutAbsolute
	root.AddChild(shell)
	root.MarkDirty()
	root.Layout()

	padding := shell.GetStyle().Padding
	wantInnerH := h - 2*padding
	if vp.Bounds().Height > wantInnerH+4 {
		t.Fatalf("viewport height %.0f exceeds shell inner %.0f (elongated)", vp.Bounds().Height, wantInnerH)
	}
	if vp.Bounds().Height < wantInnerH-40 {
		t.Fatalf("viewport height %.0f too short vs shell inner %.0f", vp.Bounds().Height, wantInnerH)
	}
	if vp.overflowScrollY() < 200 {
		t.Fatalf("overflowScrollY %.0f, want scrollable tall grid content", vp.overflowScrollY())
	}
}

func TestShellMainViewportBoundsFlushRight(t *testing.T) {
	const w, h = float32(800), float32(600)
	shell := NewContainer("shell", 0, 0, w, h)
	shell.SetStyle("page-shell")
	inner := ShellMainViewportBounds(shell)
	pad := shell.GetStyle().Padding
	if inner.X != pad {
		t.Fatalf("inner.X = %.0f, want left pad %.0f", inner.X, pad)
	}
	if inner.Y != pad {
		t.Fatalf("inner.Y = %.0f, want top pad %.0f", inner.Y, pad)
	}
	wantW := w - pad
	if inner.Width != wantW {
		t.Fatalf("inner.Width = %.0f, want flush-right %.0f", inner.Width, wantW)
	}
	if inner.Height != h-2*pad {
		t.Fatalf("inner.Height = %.0f, want %.0f", inner.Height, h-2*pad)
	}
}

func TestPageShellSyncsScrollViewportFlushRight(t *testing.T) {
	const w, h = float32(800), float32(600)
	shell := NewContainer("shell", 0, 0, w, h)
	shell.LayoutType = LayoutFlex
	shell.FlexDirection = FlexColumn
	shell.SetStyle("page-shell")
	vp := NewViewport("vp", 0, 0, 0, 0)
	vp.SetStyle("page-scroll")
	vp.SetFlexGrow(1)
	shell.AddChild(vp)
	root := NewContainer("root", 0, 0, w, h)
	root.LayoutType = LayoutAbsolute
	root.AddChild(shell)
	root.MarkDirty()
	root.Layout()
	pad := shell.GetStyle().Padding
	if vp.Bounds().Width != w-pad {
		t.Fatalf("vp width %.0f, want flush-right %.0f", vp.Bounds().Width, w-pad)
	}
	if vp.Bounds().X != pad {
		t.Fatalf("vp x %.0f, want %.0f", vp.Bounds().X, pad)
	}
}

func TestPageScrollContentRespectsPadding(t *testing.T) {
	const w, h = float32(800), float32(600)
	frame := NewContainer("frame", 0, 0, w, h)
	frame.LayoutType = LayoutFlex
	frame.FlexDirection = FlexColumn
	frame.SetStyle("transparent")
	vp := NewViewport("vp", 0, 0, 0, 0)
	vp.SetStyle("page-scroll")
	vp.SetFlexGrow(1)
	hdr := NewHeader("hdr", "Title", "Subtitle", 0, 0, 0, 64)
	vp.AddChild(hdr)
	frame.AddChild(vp)
	root := NewContainer("root", 0, 0, w, h)
	root.LayoutType = LayoutAbsolute
	root.AddChild(frame)
	root.MarkDirty()
	root.Layout()

	padL, _, _, _ := vp.scrollContentPadding()
	wantW := vp.scrollContentWidthBudget(vp.Bounds())
	b := hdr.Bounds()
	if b.X != vp.Bounds().X+padL {
		t.Fatalf("header X %.0f, want content left %.0f", b.X, vp.Bounds().X+padL)
	}
	if b.Width > wantW+0.5 {
		t.Fatalf("header width %.0f exceeds content budget %.0f", b.Width, wantW)
	}
	if vp.ContentClipBleed != 0 {
		t.Fatalf("page-scroll bleed = %v, want 0", vp.ContentClipBleed)
	}
}

func TestPageViewportFillsClientFrame(t *testing.T) {
	const w, h = float32(800), float32(600)
	frame := NewContainer("frame", 0, 0, w, h)
	frame.LayoutType = LayoutFlex
	frame.FlexDirection = FlexColumn
	frame.SetStyle("transparent")
	vp := NewViewport("vp", 0, 0, 0, 0)
	vp.SetStyle("page-scroll")
	vp.SetFlexGrow(1)
	frame.AddChild(vp)
	root := NewContainer("root", 0, 0, w, h)
	root.LayoutType = LayoutAbsolute
	root.AddChild(frame)
	root.MarkDirty()
	root.Layout()

	if vp.Bounds().X != frame.Bounds().X || vp.Bounds().Y != frame.Bounds().Y {
		t.Fatalf("viewport origin should match frame")
	}
	if vp.Bounds().Width != frame.Bounds().Width || vp.Bounds().Height != frame.Bounds().Height {
		t.Fatalf("viewport size (%.0f×%.0f) should match frame (%.0f×%.0f)",
			vp.Bounds().Width, vp.Bounds().Height, frame.Bounds().Width, frame.Bounds().Height)
	}
}

func TestPageScrollSymmetricHorizontalMargins(t *testing.T) {
	const w, h = float32(800), float32(600)
	vp := NewViewport("vp", 0, 0, w, h)
	vp.SetStyle("page-scroll")
	client := vp.viewportPaddedClientRect()
	if client.X != pageScrollPadH || client.Width != w-2*pageScrollPadH {
		t.Fatalf("client band x=%.0f w=%.0f, want x=%.0f w=%.0f", client.X, client.Width, pageScrollPadH, w-2*pageScrollPadH)
	}
	budget := vp.scrollContentWidthBudget(vp.Bounds())
	if budget != client.Width {
		t.Fatalf("content budget %.0f != client width %.0f", budget, client.Width)
	}
	_, trackY, _, trackH := vp.pageScrollVertTrackRect(vp.Bounds())
	if trackY != pageScrollPadV+pageScrollBarVertInset {
		t.Fatalf("track Y %.0f, want pad+inset %.0f", trackY, pageScrollPadV+pageScrollBarVertInset)
	}
	wantTrackH := h - 2*pageScrollPadV - 2*pageScrollBarVertInset
	if trackH != wantTrackH {
		t.Fatalf("track H %.0f, want %.0f", trackH, wantTrackH)
	}
}

func TestSettingsScrollFlushVerticalPadding(t *testing.T) {
	const w, h = float32(800), float32(500)
	vp := NewViewport("vp", 0, 40, w, h)
	vp.SetStyle("settings-scroll")
	padL, padT, padR, padB := vp.scrollContentPadding()
	if padT != 0 || padB != 0 {
		t.Fatalf("vertical padding = top %.0f bottom %.0f, want 0", padT, padB)
	}
	if padL != settingsScrollPadH || padR != settingsScrollPadH {
		t.Fatalf("horizontal padding = %.0f / %.0f, want %.0f", padL, padR, settingsScrollPadH)
	}
	client := vp.viewportPaddedClientRect()
	if client.Y != vp.Bounds().Y || client.Height != h {
		t.Fatalf("client Y/H = %.0f/%.0f, want flush %.0f/%.0f", client.Y, client.Height, vp.Bounds().Y, h)
	}
}

func TestViewportSetFlexGrowClearsAutoHeight(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 0, 0)
	if !vp.AutoHeight {
		t.Fatal("h=0 viewport should start AutoHeight")
	}
	vp.SetFlexGrow(1)
	if vp.AutoHeight {
		t.Fatal("flex-grow viewport should not stay AutoHeight")
	}
}
