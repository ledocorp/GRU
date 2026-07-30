package examples

import "github.com/ledocorp/gru/ui"

// clearRootChildren removes every direct child of doc.Root (used when
// mounting a fresh page shell). Scene Build runs on a new Document, so the
// slice is usually empty.
func clearRootChildren(doc *ui.Document) {
	ids := make([]string, 0, len(doc.Root.Children()))
	for _, ch := range doc.Root.Children() {
		ids = append(ids, ch.ID())
	}
	for _, id := range ids {
		doc.Root.RemoveChild(id)
	}
}

// MountPageGrid clears the document root and adds a single full-window
// responsive page grid (see ui.NewPageGrid). Add grid children with default
// ColSpan (full row) or SetColSpan for breakpoint-aware spans.
//
// Root (absolute) → **page grid** → your Viewports / panels.
func MountPageGrid(doc *ui.Document, id string) *ui.Container {
	clearRootChildren(doc)
	grid := ui.NewPageGrid(id, float32(doc.Width), float32(doc.Height))
	doc.Root.AddChild(grid)
	return grid
}

// AppPage is the set-and-forget page chrome from MountAppPage.
type AppPage struct {
	// Frame is a full-client flex wrapper (transparent, no padding). It exists
	// only for Document.Resize — the scroll viewport is the real page surface.
	Frame  *ui.Container
	Body   *ui.Viewport
	Header *ui.Header // nil when title is empty; lives inside Body (scrolls with the page)
	// Shell aliases Frame for callers that still use page.Shell.
	Shell *ui.Container
}

// MountAppPage mounts the web-faithful scrollable page:
//
//	TitleBar (window chrome, outside doc tree)
//	Root → frame (full client, no inset) → page-scroll Viewport (1:1 with client area)
//	      └── [Header] + your grids/panels — all page content scrolls inside the viewport
//
// Pass empty title to omit the header. Add grids and panels to Body.
// Margins and the scrollbar gutter are owned by page-scroll inside the viewport.
func MountAppPage(doc *ui.Document, id, title, subtitle string) AppPage {
	clearRootChildren(doc)
	w, h := float32(doc.Width), float32(doc.Height)
	frame := ui.NewContainer(id+"-frame", 0, 0, w, h)
	frame.LayoutType = ui.LayoutFlex
	frame.FlexDirection = ui.FlexColumn
	frame.SetStyle("transparent")

	body := ui.NewViewport(id+"-vp", 0, 0, 0, 0)
	body.SetStyle("page-scroll")
	body.FlexDirection = ui.FlexColumn
	body.SetFlexGrow(1)

	var hdr *ui.Header
	if title != "" {
		hdr = ui.NewHeader(id+"-hdr", title, subtitle, 0, 0, 0, 0)
		hdr.SetStyle("header")
		body.AddChild(hdr)
	}

	frame.AddChild(body)
	doc.Root.AddChild(frame)
	doc.Root.MarkDirty()
	frame.MarkDirty()
	doc.Root.Layout()

	return AppPage{Frame: frame, Shell: frame, Body: body, Header: hdr}
}

// MountFlexPageShell mounts a scrollable page without a title header.
func MountFlexPageShell(doc *ui.Document, id string) (*ui.Container, *ui.Viewport) {
	page := MountAppPage(doc, id, "", "")
	return page.Frame, page.Body
}

// MountSceneHeader adds a page Header inside a scroll viewport (first row of the
// document). Prefer MountAppPage title/subtitle args for new scenes.
func MountSceneHeader(vp *ui.Viewport, id, title, subtitle string) *ui.Header {
	hdr := ui.NewHeader(id, title, subtitle, 0, 0, 0, 0)
	hdr.SetStyle("header")
	vp.AddChild(hdr)
	return hdr
}

// MountEdgeToEdgeRoot clears the document root and adds a full-content flex row/column
// shell (no page-shell inset). Prefer MountDesktopPageShell for rail + AppBar scenes.
func MountEdgeToEdgeRoot(doc *ui.Document, id string, row bool) *ui.Container {
	clearRootChildren(doc)
	w, h := float32(doc.Width), float32(doc.Height)
	root := ui.NewContainer(id+"-root", 0, 0, w, h)
	root.LayoutType = ui.LayoutFlex
	if row {
		root.FlexDirection = ui.FlexRow
	} else {
		root.FlexDirection = ui.FlexColumn
	}
	root.SetStyle("transparent")
	doc.Root.AddChild(root)
	doc.Root.MarkDirty()
	return root
}

// MountDesktopPageShell mounts the architecture page shell (sole direct child of Root,
// sized to doc, flex column) with a flex-row workspace for NavigationRail + content.
//
//	Root → shell (flex column) → [optional MenuBar] → workspace (flex row, flex-grow) → rail | main
//	→ [optional StatusBar]
//
// Demos insert full-width MenuBar / StatusBar as direct shell children (not beside the rail).
// Same resize contract as MountFlexPageShell (Document.fitRootChildrenToContent +
// applyShellFlexAndSyncRootLayout). Add rail and main to workspace, viewport inside main.
func MountDesktopPageShell(doc *ui.Document, id string) (*ui.Container, *ui.Container) {
	clearRootChildren(doc)
	w, h := float32(doc.Width), float32(doc.Height)
	shell := ui.NewContainer(id+"-shell", 0, 0, w, h)
	shell.LayoutType = ui.LayoutFlex
	shell.FlexDirection = ui.FlexColumn
	shell.SetStyle("transparent")

	workspace := ui.NewContainer(id+"-workspace", 0, 0, 0, 0)
	workspace.LayoutType = ui.LayoutFlex
	workspace.FlexDirection = ui.FlexRow
	workspace.SetStyle("transparent")
	workspace.SetFlexGrow(1)

	shell.AddChild(workspace)
	doc.Root.AddChild(shell)
	doc.Root.MarkDirty()
	shell.MarkDirty()
	doc.Root.Layout()
	return shell, workspace
}

// AppShellMount is a single flex-column shell filling the document content band.
// Add AppBar, page-scroll Viewport, and optional footer as direct shell children.
type AppShellMount struct {
	Shell *ui.Container
}

// MountAppShellRoot clears Root and mounts one transparent flex-column shell.
// Window chrome (title bar, launcher nav) is handled by Document — do not add
// an extra appshell-content body wrapper around the scroll viewport.
func MountAppShellRoot(doc *ui.Document, id string) AppShellMount {
	clearRootChildren(doc)
	w, h := float32(doc.Width), float32(doc.Height)
	shell := ui.NewContainer(id+"-shell", 0, 0, w, h)
	shell.LayoutType = ui.LayoutFlex
	shell.FlexDirection = ui.FlexColumn
	shell.SetStyle("transparent")
	doc.Root.AddChild(shell)
	return AppShellMount{Shell: shell}
}

// NewAppShellScrollViewport returns the page-scroll viewport for app-shell scenes.
// Add as a direct child of the shell (flex-grow 1) between pinned chrome rows.
func NewAppShellScrollViewport(id string) *ui.Viewport {
	vp := ui.NewViewport(id, 0, 0, 0, 0)
	vp.SetStyle("page-scroll")
	vp.FlexDirection = ui.FlexColumn
	vp.SetFlexGrow(1)
	return vp
}

// FinishShellMount runs layout passes after scene Build + Document.Resize.
// main.go calls this once per loadScene (scenes should not call it from Build —
// that runs before Resize and freezes wrong AutoHeight measures until a nudge).
func FinishShellMount(doc *ui.Document) {
	if doc == nil || doc.Root == nil {
		return
	}
	// Same prep as a tiny window resize: dirty the full tree, clear text measure
	// caches, then width-first layout until Panel/Card intrinsic heights stabilize.
	for pass := 0; pass < 5; pass++ {
		ui.MarkResizeLayoutDirtySubtree(doc.Root)
		ui.InvalidateAutoHeightTextMeasures(doc.Root)
		doc.Root.MarkDirty()
		doc.Root.Layout()
	}
	settlePreviewMarkdown(doc.Root)
}

// settlePreviewMarkdown reflows MarkdownView panes after shell widths are real.
func settlePreviewMarkdown(n ui.Node) {
	if n == nil {
		return
	}
	type markdownResizeNudge interface {
		SimulateResizeReflow()
	}
	if mv, ok := n.(markdownResizeNudge); ok {
		mv.SimulateResizeReflow()
	}
	for _, ch := range n.Children() {
		settlePreviewMarkdown(ch)
	}
}
