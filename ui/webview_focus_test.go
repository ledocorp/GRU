package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type webViewBlurSpy struct {
	webViewHostStub
	blurs int
}

func (s *webViewBlurSpy) Blur() { s.blurs++ }

func TestRouteScenePointerFocusTextEditor(t *testing.T) {
	SetScenePointerBlocked(false)
	ed := NewTextEditor("ed", "hello", 0, 0, 200, 200)
	wv := NewWebViewPanel("wv", "", 300, 0, 200, 200)
	root := NewContainer("root", 0, 0, 500, 200)
	root.LayoutType = LayoutAbsolute
	root.AddChild(ed)
	root.AddChild(wv)
	ed.SetBounds(rl.NewRectangle(0, 0, 200, 200))
	wv.SetBounds(rl.NewRectangle(300, 0, 200, 200))

	spy := &webViewBlurSpy{}
	wv.host = spy
	registerWebViewHost(spy)
	t.Cleanup(func() { unregisterWebViewHost(spy) })

	doc := &Document{Root: root, Width: 500, Height: 200}
	latchPointerClick(rl.NewVector2(50, 50))

	RouteScenePointerFocus(doc)
	if doc.Focused != ed {
		t.Fatalf("Focused=%v want TextEditor", doc.Focused)
	}
	if spy.blurs != 1 {
		t.Fatalf("web blurs=%d want 1", spy.blurs)
	}
	if PointerClickPending() {
		t.Fatal("click should be consumed")
	}
}

func TestRouteScenePointerFocusWebView(t *testing.T) {
	SetScenePointerBlocked(false)
	ed := NewTextEditor("ed", "hello", 0, 0, 200, 200)
	wv := NewWebViewPanel("wv", "", 300, 0, 200, 200)
	root := NewContainer("root", 0, 0, 500, 200)
	root.LayoutType = LayoutAbsolute
	root.AddChild(ed)
	root.AddChild(wv)
	ed.SetBounds(rl.NewRectangle(0, 0, 200, 200))
	wv.SetBounds(rl.NewRectangle(300, 0, 200, 200))

	doc := &Document{Root: root, Width: 500, Height: 200}
	doc.SetFocus(ed)
	if !ed.IsFocused() {
		t.Fatal("editor should start focused")
	}

	latchPointerClick(rl.NewVector2(350, 50))
	RouteScenePointerFocus(doc)
	if doc.Focused != nil {
		t.Fatalf("Focused=%v want nil for web", doc.Focused)
	}
	if ed.IsFocused() {
		t.Fatal("editor should blur when web clicked")
	}
}

func TestRouteScenePointerFocusOutsideBlursWeb(t *testing.T) {
	SetScenePointerBlocked(false)
	wv := NewWebViewPanel("wv", "", 100, 100, 200, 200)
	root := NewContainer("root", 0, 0, 500, 400)
	root.LayoutType = LayoutAbsolute
	root.AddChild(wv)
	wv.SetBounds(rl.NewRectangle(100, 100, 200, 200))

	spy := &webViewBlurSpy{}
	wv.host = spy
	registerWebViewHost(spy)
	t.Cleanup(func() { unregisterWebViewHost(spy) })

	doc := &Document{Root: root, Width: 500, Height: 400}
	latchPointerClick(rl.NewVector2(10, 10))
	RouteScenePointerFocus(doc)
	if spy.blurs != 1 {
		t.Fatalf("web blurs=%d want 1", spy.blurs)
	}
}

// TestRouteScenePointerFocusFlexLayout mirrors webview_focus_demo row layout.
func TestRouteScenePointerFocusFlexLayout(t *testing.T) {
	SetScenePointerBlocked(false)
	doc := &Document{Root: NewContainer("root", 0, 48, 800, 600), Width: 800, Height: 600}
	doc.SetChromeTop(48)
	root := doc.Root
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn

	row := NewContainer("row", 0, 0, 0, 0)
	row.LayoutType = LayoutFlex
	row.FlexDirection = FlexRow
	row.SetFlexGrow(1)
	root.AddChild(row)

	nativeCol := NewPanel("native", "", 0, 0, 0, 0)
	nativeCol.SetFlexGrow(1)
	ed := NewTextEditor("ed", "native", 0, 0, 0, 0)
	ed.SetFlexGrow(1)
	nativeCol.AddChild(ed)
	row.AddChild(nativeCol)

	wv := NewWebViewPanel("wv", "", 0, 0, 0, 0)
	wv.SetFlexGrow(1)
	row.AddChild(wv)

	root.MarkDirty()
	root.Layout()

	edB := ed.Bounds()
	wvB := wv.Bounds()
	if edB.Width < 1 || wvB.Width < 1 {
		t.Fatalf("layout failed: ed=%v wv=%v", edB, wvB)
	}

	latchPointerClick(rl.NewVector2(edB.X+edB.Width/2, edB.Y+edB.Height/2))
	RouteScenePointerFocus(doc)
	if doc.Focused != ed {
		t.Fatalf("editor click: Focused=%v want TextEditor", doc.Focused)
	}

	doc.SetFocus(ed)
	latchPointerClick(rl.NewVector2(wvB.X+wvB.Width/2, wvB.Y+wvB.Height/2))
	RouteScenePointerFocus(doc)
	if doc.Focused != nil {
		t.Fatalf("web click: Focused=%v want nil", doc.Focused)
	}
}
