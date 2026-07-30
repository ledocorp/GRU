// Package examples — docked list-pane shell (master list in SplitView).
package examples

import "github.com/ledocorp/gru/ui"

// ListPaneOptions configures a flat, borderless list sidebar.
type ListPaneOptions struct {
	ID             string
	Title          string
	TitleStyle     string // theme key; default list-pane-header-title
	PreferredWidth float32
	MinWidth       float32
	ShowCollapse   bool
	OnCollapse     func()
}

// ListPane is a flat grey sidebar: customizable header + flush scroll + row list.
type ListPane struct {
	Root            *ui.Container
	Hdr             *ui.Container
	Title           *ui.Label
	Collapse        *ui.IconButton
	Scroll          *ui.Viewport
	List            *ui.Container
	PreferredWidth  float32
	MinWidth        float32
}

// NewListPane builds a borderless list column. Add rows to List; scrollbar aligns with the split edge.
func NewListPane(doc *ui.Document, opts ListPaneOptions) *ListPane {
	if opts.ID == "" {
		opts.ID = "list-pane"
	}
	if opts.TitleStyle == "" {
		opts.TitleStyle = "list-pane-header-title"
	}
	if opts.PreferredWidth <= 0 {
		opts.PreferredWidth = 260
	}
	if opts.MinWidth <= 0 {
		opts.MinWidth = 180
	}

	lp := &ListPane{}

	root := ui.NewContainer(opts.ID, 0, 0, 0, 0)
	root.LayoutType = ui.LayoutFlex
	root.FlexDirection = ui.FlexColumn
	root.SetStyle("list-pane")
	root.Gap = 0
	root.PreferredWidth = opts.PreferredWidth
	root.MinWidth = opts.MinWidth
	root.SetFlexGrow(1)

	hdr := ui.NewContainer(opts.ID+"-hdr", 0, 0, 0, 0)
	hdr.LayoutType = ui.LayoutFlex
	hdr.FlexDirection = ui.FlexRow
	hdr.Gap = 8
	hdr.SetStyle("list-pane-header")
	hdr.AutoHeight = true

	lp.Hdr = hdr
	lp.Title = ui.NewLabel(opts.ID+"-title", opts.Title, 0, 0, 0, 0)
	lp.Title.SetStyle(opts.TitleStyle)
	lp.Title.Align = ui.LabelAlignLeft
	lp.Title.SetFlexGrow(1)
	hdr.AddChild(lp.Title)

	if opts.ShowCollapse {
		PreloadScenePhosphor(doc)
		ui.Phosphor.EnsureLoaded(ui.PhosphorCaretLeft, ui.PhosphorRegular)
		lp.Collapse = ui.NewIconButton(opts.ID+"-collapse", "", "", 0, 0, 32, 32)
		lp.Collapse.SetStyle("appbar-icon")
		lp.Collapse.SetPhosphorIcon(ui.PhosphorCaretLeft, ui.PhosphorRegular)
		if opts.OnCollapse != nil {
			lp.Collapse.OnClick = opts.OnCollapse
		}
		hdr.AddChild(lp.Collapse)
	}

	scroll := ui.NewViewport(opts.ID+"-scroll", 0, 0, 0, 0)
	scroll.SetStyle("list-pane-scroll")
	scroll.SetFlexGrow(1)

	lp.List = ui.NewContainer(opts.ID+"-rows", 0, 0, 0, 0)
	lp.List.LayoutType = ui.LayoutFlex
	lp.List.FlexDirection = ui.FlexColumn
	lp.List.Gap = 6
	lp.List.SetStyle("list-pane-list")
	lp.List.AutoHeight = true

	scroll.AddChild(lp.List)
	root.AddChild(hdr)
	root.AddChild(scroll)

	lp.Root = root
	lp.PreferredWidth = opts.PreferredWidth
	lp.MinWidth = opts.MinWidth
	lp.Scroll = scroll
	return lp
}
