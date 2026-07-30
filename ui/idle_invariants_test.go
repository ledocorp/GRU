package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// buildNotepadIdleFixture returns a tree shaped like Notepad desktop:
// mainSplit(notes | editorArea) and editorSplit(editorVP | previewVP).
func buildNotepadIdleFixture() (*Container, *SplitView, *SplitView) {
	notesList := NewContainer("notes-list", 0, 0, 0, 0)
	notesList.LayoutType = LayoutFlex
	notesList.FlexDirection = FlexColumn
	notesList.AutoHeight = true
	notesList.AddChild(NewLabel("note-row", "Note one", 0, 0, 0, 0))

	notesScroll := NewViewport("notes-scroll", 0, 0, 0, 0)
	notesScroll.AddChild(notesList)

	notesRoot := NewContainer("notes-root", 0, 0, 0, 0)
	notesRoot.LayoutType = LayoutFlex
	notesRoot.FlexDirection = FlexColumn
	notesRoot.SetFlexGrow(1)
	notesRoot.AddChild(notesScroll)

	editorVP := NewViewport("editor-vp", 0, 0, 0, 0)
	editorVP.AddChild(NewLabel("editor", "hello world", 0, 0, 0, 0))

	previewVP := NewViewport("preview-vp", 0, 0, 0, 0)
	previewLane := NewContainer("preview-lane", 0, 0, 0, 0)
	previewLane.AutoHeight = true
	previewLane.AddChild(NewLabel("preview-md", "Preview", 0, 0, 0, 0))
	previewVP.AddChild(previewLane)

	editorSplit := NewSplitView("editor-split", SplitHorizontal, editorVP, previewVP, 0, 0, 0, 0)
	editorSplit.Ratio.Set(0.52)
	editorSplit.SetFlexGrow(1)

	editorArea := NewContainer("editor-area", 0, 0, 0, 0)
	editorArea.LayoutType = LayoutFlex
	editorArea.FlexDirection = FlexColumn
	editorArea.SetFlexGrow(1)
	editorArea.AddChild(editorSplit)

	mainSplit := NewSplitView("main-split", SplitHorizontal, notesRoot, editorArea, 0, 0, 0, 0)
	mainSplit.Ratio.Set(0.24)
	mainSplit.SetFlexGrow(1)

	root := NewContainer("notepad-root", 0, 0, 800, 600)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	root.SetBounds(rl.NewRectangle(0, 0, 800, 600))
	root.AddChild(mainSplit)

	return root, mainSplit, editorSplit
}

func clearTreeDirtyFlags(root Node) {
	if root == nil {
		return
	}
	if c, ok := root.(*Container); ok {
		c.layoutDirty = false
		c.drawDirty = false
	}
	if sv, ok := root.(*SplitView); ok {
		sv.layoutDirty = false
		sv.drawDirty = false
	}
	if vp, ok := root.(*Viewport); ok {
		vp.layoutDirty = false
		vp.drawDirty = false
	}
	if l, ok := root.(*Label); ok {
		l.layoutDirty = false
		l.drawDirty = false
	}
	for _, ch := range root.Children() {
		clearTreeDirtyFlags(ch)
	}
}

func TestDemoPageIdleReadyAfterLayout(t *testing.T) {
	vp := NewViewport("dir-scroll", 0, 0, 520, 640)
	vp.SetStyle("page-scroll")
	vp.FlexDirection = FlexColumn

	body := NewContainer("dir-body", 0, 0, 0, 0)
	body.LayoutType = LayoutFlex
	body.FlexDirection = FlexColumn
	body.AutoHeight = true
	body.Gap = 10

	hdr := NewHeader("dir-hdr", "Demo Directory", "Open any registered scene.", 0, 0, 0, 0)
	body.AddChild(hdr)
	for i := 0; i < 3; i++ {
		tile := NewListTile("dir-row", "Demo scene", "subtitle", 0, 0, 0, 0)
		tile.OnClick = func() {}
		body.AddChild(tile)
	}
	vp.AddChild(body)

	root := NewContainer("dir-root", 0, 0, 520, 640)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	root.SetBounds(rl.NewRectangle(0, 0, 520, 640))
	root.AddChild(vp)

	root.Layout()
	SimulateCacheHitFrame(root)
	AssertIdleReady(t, root, "demo directory page")
}

func TestViewportScheduleLayoutPassClearsOnLayout(t *testing.T) {
	vp := NewViewport("dir-vp", 0, 0, 520, 640)
	vp.SetStyle("page-scroll")
	body := NewContainer("body", 0, 0, 0, 0)
	body.LayoutType = LayoutFlex
	body.AutoHeight = true
	body.AddChild(NewHeader("hdr", "Title", "Sub", 0, 0, 0, 0))
	vp.AddChild(body)
	vp.SetBounds(rl.NewRectangle(0, 0, 520, 640))
	vp.Layout()
	clearTreeDirtyFlags(vp)
	scheduleLayoutPass(vp)
	if !vp.IsDirty() {
		t.Fatal("want layout dirty after scheduleLayoutPass")
	}
	vp.Layout()
	if vp.IsDirty() {
		t.Fatalf("viewport still layout-dirty after Layout: %s", NotIdleReason(vp))
	}
}

func TestViewportOrphanLayoutDirtyClearedByParent(t *testing.T) {
	frame := NewContainer("dir-frame", 0, 0, 520, 640)
	frame.LayoutType = LayoutFlex
	frame.FlexDirection = FlexColumn
	frame.SetStyle("transparent")

	vp := NewViewport("dir-vp", 0, 0, 520, 640)
	vp.SetStyle("page-scroll")
	vp.SetFlexGrow(1)
	body := NewContainer("dir-body", 0, 0, 0, 0)
	body.LayoutType = LayoutFlex
	body.FlexDirection = FlexColumn
	body.AutoHeight = true
	body.Gap = 10
	body.AddChild(NewHeader("dir-hdr", "Demo Directory", "Open scenes.", 0, 0, 0, 0))
	vp.AddChild(body)
	frame.AddChild(vp)

	root := NewContainer("root", 0, 0, 520, 640)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	root.SetBounds(rl.NewRectangle(0, 0, 520, 640))
	root.AddChild(frame)

	root.Layout()
	clearTreeDirtyFlags(root)

	scheduleLayoutPass(vp)
	if root.IsDirty() {
		t.Fatal("scheduleLayoutPass must not bubble to root")
	}
	if !vp.IsDirty() {
		t.Fatal("viewport should be layout-dirty")
	}
	if !SubtreeNeedsRedraw(root) {
		t.Fatal("orphan viewport dirty should block idle")
	}

	root.Layout()
	if vp.IsDirty() {
		t.Fatalf("orphan viewport still layout-dirty after root Layout: %s", NotIdleReason(root))
	}
	SimulateCacheHitFrame(root)
	AssertIdleReady(t, root, "after orphan viewport layout")
}

func TestDirectoryPageIdleAfterUpdateLoop(t *testing.T) {
	vp := NewViewport("dir-vp", 0, 0, 664, 949)
	vp.SetStyle("page-scroll")
	vp.FlexDirection = FlexColumn
	vp.SetFlexGrow(1)

	body := NewContainer("dir-body", 0, 0, 0, 0)
	body.LayoutType = LayoutFlex
	body.FlexDirection = FlexColumn
	body.AutoHeight = true
	body.Gap = 10
	body.AddChild(NewHeader("dir-hdr", "Demo Directory", "Open any registered scene.", 0, 0, 0, 0))
	for i := 0; i < 8; i++ {
		tile := NewListTile("dir-row", "Demo scene", "subtitle", 0, 0, 0, 0)
		tile.OnClick = func() {}
		body.AddChild(tile)
	}
	vp.AddChild(body)

	frame := NewContainer("dir-frame", 0, 0, 664, 949)
	frame.LayoutType = LayoutFlex
	frame.FlexDirection = FlexColumn
	frame.SetStyle("transparent")
	frame.AddChild(vp)

	root := NewContainer("root", 0, 0, 664, 949)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	root.SetBounds(rl.NewRectangle(0, 0, 664, 949))
	root.AddChild(frame)

	root.Layout()
	SimulateCacheHitFrame(root)

	const dt float32 = 1.0 / 60.0
	for i := 0; i < 120; i++ {
		if root.IsDirty() || SubtreeLayoutDirty(root) {
			root.Layout()
		}
		root.Update(dt)
		if root.IsDirty() || SubtreeLayoutDirty(root) {
			root.Layout()
		}
	}
	AssertIdleReady(t, root, "directory page after 120 update frames")
}

func TestNotepadLikeFixtureIdleAfterMainSplitRelayout(t *testing.T) {
	root, mainSplit, _ := buildNotepadIdleFixture()
	root.Layout()
	clearTreeDirtyFlags(root)

	mainSplit.MarkDirty()
	root.Layout()
	SimulateCacheHitFrame(root)
	AssertIdleReady(t, root, "mainSplit relayout")
}

func TestNotepadLikeFixtureIdleAfterNestedSplitRelayout(t *testing.T) {
	root, mainSplit, editorSplit := buildNotepadIdleFixture()
	root.Layout()
	clearTreeDirtyFlags(root)

	mainSplit.MarkDirty()
	editorSplit.MarkDirty()
	root.Layout()
	SimulateCacheHitFrame(root)
	AssertIdleReady(t, root, "nested split relayout")
}

func TestNotepadLikeFixtureIdleAfterCacheHitSimulation(t *testing.T) {
	root, _, _ := buildNotepadIdleFixture()
	root.Layout()
	// Residual draw-only dirty from chrome is OK if cleared after blit.
	SimulateCacheHitFrame(root)
	if SubtreeNeedsRedraw(root) {
		// Layout dirty must be cleared by Layout; draw-only should be gone.
		reports := CollectDirtyReports(root, 8)
		for _, r := range reports {
			if r.LayoutDirty {
				t.Fatalf("layout still dirty after settle: %+v", reports)
			}
		}
	}
}

func TestLayoutChildAfterSetBoundsDoesNotDirtyRoot(t *testing.T) {
	left := NewLabel("left", "L", 0, 0, 0, 20)
	right := NewLabel("right", "R", 0, 0, 0, 20)
	sv := NewSplitView("sv", SplitHorizontal, left, right, 0, 0, 400, 200)
	root := NewContainer("root", 0, 0, 400, 200)
	root.AddChild(sv)
	root.Layout()
	root.layoutDirty = false

	layoutChildAfterSetBounds(left, left.Bounds())
	if root.IsDirty() {
		t.Fatal("layoutChildAfterSetBounds must not bubble MarkDirty to root")
	}
}

func TestListTileLayoutClearsDirty(t *testing.T) {
	closeBtn := NewIconButton("close", "x", "", 0, 0, 24, 24)
	tile := NewListTile("sess-row", "Document one", "", 0, 0, 280, 36)
	tile.SetTrailing(closeBtn)
	tile.MarkDirty()
	tile.Layout()
	if tile.IsDirty() {
		t.Fatalf("ListTile still dirty after Layout: layout=%v draw=%v", tile.layoutDirty, tile.drawDirty)
	}
	if closeBtn.layoutDirty {
		t.Fatal("trailing IconButton layoutDirty not cleared after ListTile.Layout")
	}
}

func TestSeparatorLayoutClearsDirty(t *testing.T) {
	sep := NewSeparator("md-rule", "", 0, 0, 400, 1)
	sep.MarkDirty()
	sep.Layout()
	if sep.IsDirty() {
		t.Fatalf("Separator still dirty after Layout: layout=%v draw=%v", sep.layoutDirty, sep.drawDirty)
	}
}

func TestLabelAutoHeightLayoutClearsDirty(t *testing.T) {
	lbl := NewLabel("notes-title", "Open notes", 0, 0, 0, 0)
	lbl.MarkDirty()
	lbl.Layout()
	if lbl.layoutDirty {
		t.Fatal("Label layoutDirty not cleared after Layout")
	}
}

func TestOpenNotesLikeFixtureIdleAfterSessionListRelayout(t *testing.T) {
	root := NewContainer("open-notes", 0, 0, 320, 400)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	root.SetBounds(rl.NewRectangle(0, 0, 320, 400))

	title := NewLabel("open-notes-title", "Open notes", 0, 0, 0, 0)
	title.MarkDirty()
	root.AddChild(title)

	list := NewContainer("sess-list", 0, 0, 0, 0)
	list.LayoutType = LayoutFlex
	list.FlexDirection = FlexColumn
	list.AutoHeight = true
	for i := 0; i < 3; i++ {
		btn := NewIconButton("menu", "⋮", "", 0, 0, 36, 36)
		row := NewListTile("sess-doc", "Doc", "", 0, 0, 0, 36)
		row.SetTrailing(btn)
		row.MarkDirty()
		list.AddChild(row)
	}
	root.AddChild(list)

	root.Layout()
	SimulateCacheHitFrame(root)
	AssertIdleReady(t, root, "session list relayout")
}

func TestPreviewCardRichTextIdleAfterDrawClears(t *testing.T) {
	lane := NewContainer("preview-lane", 0, 0, 320, 400)
	lane.LayoutType = LayoutFlex
	lane.FlexDirection = FlexColumn
	lane.AutoHeight = true
	lane.SetBounds(rl.NewRectangle(0, 0, 320, 400))

	for i := 0; i < 3; i++ {
		card := NewCard("md-b"+string(rune('0'+i)), "Block", 0, 0, 0, 0)
		rt := NewRichText("md-b"+string(rune('0'+i))+"-code", []TextSpan{{Text: "fmt.Println()"}}, 0, 0, 0, 0)
		card.AddChild(rt)
		card.MarkDrawDirty()
		rt.MarkDrawDirty()
		lane.AddChild(card)
	}

	lane.Layout()
	// Simulate post-Draw dirty clear for preview chrome (Draw needs raylib).
	clearDrawDirtySubtree(lane)
	AssertIdleReady(t, lane, "preview card+richtext after draw")
}

func TestEditorChainIdleAfterPostRedrawClear(t *testing.T) {
	ed := NewTextEditor("notepad-editor", "hello", 0, 0, 0, 0)
	ed.AutoHeight = true
	vp := NewViewport("notepad-editor-vp", 0, 0, 320, 200)
	vp.AddChild(ed)
	split := NewSplitView("notepad-split", SplitHorizontal, vp, nil, 0, 0, 320, 200)
	area := NewContainer("notepad-editor-area", 0, 0, 320, 200)
	area.AddChild(split)
	area.SetBounds(rl.NewRectangle(0, 0, 320, 200))
	area.Layout()

	ed.MarkDrawDirty()
	SimulateCacheHitFrame(area)
	AssertIdleReady(t, area, "editor chain post-redraw clear")
}
