package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// buildAppBarScrollFixture mirrors Path C AppBar panel: card shell with pinned
// AppBar and page-scroll body (see examples/pathc_demo.go buildAppBarPanel).
func buildAppBarScrollFixture(w, h float32) *Container {
	if w <= 0 {
		w = 360
	}
	if h <= 0 {
		h = 320
	}
	shell := NewCard("app-shell", "", 0, 0, 0, 280)
	shell.Gap = 0
	shell.SetStyle("card")

	bar := NewAppBar("app-bar", "Settings", 0, 0, 0, 0)
	bar.SetFlexGrow(0)
	bar.SetLeading(NewIconButton("menu", "≡", "", 0, 0, 36, 36))
	save := NewButton("save", "Save", 0, 0, 72, 36)
	save.SetStyle("primary")
	bar.AddTrailing(save)
	bar.AddTrailing(NewIconButton("more", "⋮", "", 0, 0, 36, 36))

	scroll := NewViewport("app-scroll", 0, 0, 0, 0)
	scroll.SetStyle("page-scroll")
	scroll.FlexDirection = FlexColumn
	scroll.Gap = 8
	scroll.SetFlexGrow(1)
	for i := 0; i < 4; i++ {
		body := NewPlainText("body-"+string(rune('0'+i)), "form-value",
			"Scroll body copy — AppBar stays pinned above this viewport.", 0, 0, 0, 0)
		scroll.AddChild(body)
	}

	shell.AddChild(bar)
	shell.AddChild(scroll)

	root := NewContainer("appbar-root", 0, 0, w, h)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	root.SetBounds(rl.NewRectangle(0, 0, w, h))
	root.AddChild(shell)
	return root
}

func TestAppBarScrollFixtureIdleAfterLayout(t *testing.T) {
	root := buildAppBarScrollFixture(360, 320)
	root.Layout()
	SimulateCacheHitFrame(root)
	AssertIdleReady(t, root, "appbar+scroll shell")
}

func TestAppBarScrollFixtureIdleAfterRelayout(t *testing.T) {
	root := buildAppBarScrollFixture(360, 320)
	root.Layout()
	clearTreeDirtyFlags(root)

	shell := root.Children()[0].(*Card)
	shell.MarkDirty()
	root.Layout()
	SimulateCacheHitFrame(root)
	AssertIdleReady(t, root, "appbar+scroll relayout")
}
