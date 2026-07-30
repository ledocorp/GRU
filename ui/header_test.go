package ui

import (
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestHeaderLayoutClearsDirty(t *testing.T) {
	h := NewHeader("h", "Demo Directory", "subtitle", 0, 0, 400, 0)
	h.MarkDirty()
	h.SetBounds(rl.NewRectangle(0, 0, 400, 0))
	h.Layout()
	if h.IsDirty() {
		t.Fatal("Header.Layout must clear layoutDirty so SSAA cache can idle")
	}
}

func TestHeaderAutoHeightLayout(t *testing.T) {
	SetRootFontSize(20)
	h := NewHeader("h", "Demo Directory", "Subtitle line", 0, 0, 400, 0)
	if !h.IsAutoHeight() {
		t.Fatal("expected AutoHeight when h=0")
	}
	h.SetBounds(rl.NewRectangle(0, 0, 400, 0))
	h.Layout()
	got := h.Bounds().Height
	subStyle := h.headerSubtitleStyle(h.GetStyle())
	want := headerIntrinsicHeight(h.Title, h.Subtitle, h.GetStyle(), subStyle, 400, true)
	if got < want-1 || got > want+1 {
		t.Fatalf("height %.0f, want ~%.0f", got, want)
	}
}

func TestHeaderWrapSubtitleDefault(t *testing.T) {
	withSub := NewHeader("h", "Home", "hint", 0, 0, 0, 0)
	if !withSub.WrapSubtitle {
		t.Fatal("expected WrapSubtitle when subtitle is set")
	}
	titleOnly := NewHeader("h2", "Home", "", 0, 0, 0, 0)
	if titleOnly.WrapSubtitle {
		t.Fatal("expected WrapSubtitle off when subtitle empty")
	}
	if !withSub.UsesScissor() {
		t.Fatal("wrapped subtitle header should use scissor when drawing")
	}
}

func TestViewportHeaderRestacksFlexGrowPanel(t *testing.T) {
	const h = float32(640)
	vp := NewViewport("vp", 0, 0, 520, h)
	vp.SetStyle("page-scroll")
	vp.FlexDirection = FlexColumn

	hdr := NewHeader("hdr", "Markdown Preview", strings.Repeat("Responsive subtitle line. ", 80), 0, 0, 0, 0)
	panel := NewPanel("panel", "Feature Showcase", 0, 0, 0, 240)
	panel.SetFlexGrow(1)
	rt := NewRichText("preview", []TextSpan{{Text: strings.Repeat("heading ", 60), Variant: "h1"}}, 0, 0, 0, 0)
	rt.Wrap = true
	rt.AutoHeight = true
	rt.SetStyle("richtext-preview")
	panel.AddChild(rt)
	vp.AddChild(hdr)
	vp.AddChild(panel)

	layoutPage := func(w float32) {
		vp.InvalidateLayoutPassCache()
		vp.MarkDirty()
		vp.SetBounds(rl.NewRectangle(0, 0, w, h))
		vp.Layout()
	}
	layoutPage(520)

	layoutPage(240)
	if panel.Bounds().Y < hdr.Bounds().Y+hdr.Bounds().Height-0.5 {
		t.Fatalf("panel overlaps wrapped header: hdr bottom=%.1f panel top=%.1f",
			hdr.Bounds().Y+hdr.Bounds().Height, panel.Bounds().Y)
	}
	budget := vp.scrollContentWidthBudget(vp.Bounds())
	if rt.Bounds().Width > budget+4 {
		t.Fatalf("preview rich text width %.1f exceeds viewport content budget %.1f",
			rt.Bounds().Width, budget)
	}
}
