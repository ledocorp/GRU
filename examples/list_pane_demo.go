//go:build !notepad

// Package examples — docked ListTile list pane in SplitView (Open Notes / master-list pattern).
// Recipe: CP-SHELL-EDGE + CP-LIST-PANE-01 (see docs/COMPOSITION_PATTERNS.md §11).
package examples

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &listPaneScene{} }) }

type listPaneDoc struct {
	ID      string
	Path    string
	Unsaved bool
	Body    string
}

func (d listPaneDoc) title() string {
	if d.Path == "" {
		return "Untitled"
	}
	return filepath.Base(d.Path)
}

func (d listPaneDoc) subtitle() string {
	if d.Unsaved {
		return "Unsaved"
	}
	if d.Path == "" {
		return ""
	}
	return truncateMiddle(filepath.Dir(d.Path), 28)
}

type listPaneScene struct {
	BaseScene
	docs       []listPaneDoc
	selectedID string
	split      *ui.SplitView
	pane       *ListPane
	detailTitle *ui.Label
	detailMeta  *ui.Label
	detailBody  *ui.Label
	statusSel   *ui.Label
}

func (s *listPaneScene) Title() string { return "List Pane (Go)" }

func (s *listPaneScene) Build(doc *ui.Document) {
	s.docs = []listPaneDoc{
		{ID: "doc-1", Path: `notes/welcome.md`, Unsaved: true,
			Body: "Welcome to the list-pane demo.\n\nSelect rows on the left — the detail card updates without rebuilding the shell."},
		{ID: "doc-2", Path: `notes/meeting-agenda.md`,
			Body: "Meeting agenda\n\n1. Status\n2. Risks\n3. Next steps"},
		{ID: "doc-3", Path: `docs/composition-patterns.md`,
			Body: "Composition patterns reference.\n\nList tiles use truncated subtitles (AP-07) so long paths do not paint past the row."},
		{ID: "doc-4", Path: "",
			Body: "Untitled scratch note — no path on disk yet."},
		{ID: "doc-5", Path: `projects/demo/list_pane.go`,
			Body: "Demo source sketch for CP-LIST-PANE-01.\n\nCard host + ListTile rows + SplitView detail."},
	}
	s.selectedID = s.docs[0].ID

	root := MountEdgeToEdgeRoot(doc, "listpane", false)
	root.Gap = 0

	s.pane = NewListPane(doc, ListPaneOptions{
		ID:             "listpane",
		Title:          "Documents",
		PreferredWidth: 260,
		MinWidth:       180,
		ShowCollapse:   true,
		OnCollapse: func() {
			ui.ShowToast("List pane collapsed (demo hook)", ui.ToastInfo, 2000)
		},
	})

	detail := ui.NewContainer("listpane-detail", 0, 0, 0, 0)
	detail.LayoutType = ui.LayoutFlex
	detail.FlexDirection = ui.FlexColumn
	detail.SetStyle("transparent")
	detail.SetFlexGrow(1)

	pad := ui.NewContainer("listpane-detail-pad", 0, 0, 0, 0)
	pad.LayoutType = ui.LayoutFlex
	pad.FlexDirection = ui.FlexColumn
	pad.SetStyle("page-scroll")
	pad.SetFlexGrow(1)
	pad.Gap = 12

	card := ui.NewCard("listpane-detail-card", "Preview", 0, 0, 0, 0)
	card.SetFlexGrow(1)
	card.Gap = 10
	card.AutoHeight = false

	s.detailTitle = ui.NewLabel("listpane-detail-title", "", 0, 0, 0, 0)
	s.detailTitle.SetStyle("form-label")
	s.detailTitle.Align = ui.LabelAlignLeft
	s.detailTitle.Wrap = true

	s.detailMeta = ui.NewLabel("listpane-detail-meta", "", 0, 0, 0, 0)
	s.detailMeta.SetStyle("form-value")
	s.detailMeta.Align = ui.LabelAlignLeft
	s.detailMeta.Wrap = true

	s.detailBody = ui.NewLabel("listpane-detail-body", "", 0, 0, 0, 0)
	s.detailBody.SetStyle("form-value")
	s.detailBody.Align = ui.LabelAlignLeft
	s.detailBody.Wrap = true
	s.detailBody.SetFlexGrow(1)

	card.AddChild(s.detailTitle)
	card.AddChild(s.detailMeta)
	card.AddChild(s.detailBody)
	pad.AddChild(card)
	detail.AddChild(pad)

	s.split = ui.NewSplitView("listpane-split", ui.SplitHorizontal, s.pane.Root, detail, 0, 0, 0, 0)
	s.split.SetFlexGrow(1)
	s.split.Ratio.Set(0.28)
	s.split.MinFirst = 180
	s.split.MinSecond = 200

	body := ui.NewContainer("listpane-body", 0, 0, 0, 0)
	body.LayoutType = ui.LayoutFlex
	body.FlexDirection = ui.FlexColumn
	body.SetStyle("transparent")
	body.SetFlexGrow(1)
	body.AddChild(s.split)

	s.statusSel = ui.NewLabel("listpane-status-sel", "", 0, 0, 0, 0)
	s.statusSel.SetStyle("statusbar-label")
	s.statusSel.Align = ui.LabelAlignLeft
	countLbl := ui.NewLabel("listpane-status-count", fmt.Sprintf("%d documents", len(s.docs)), 0, 0, 0, 0)
	countLbl.SetStyle("statusbar-label")
	countLbl.Align = ui.LabelAlignRight

	status := ui.NewStatusBar("listpane-status", 0, 0, 0, 0)
	status.SetColumns([]ui.StatusBarColumn{
		{Weight: 70, Align: ui.LabelAlignLeft, Nodes: []ui.Node{s.statusSel}},
		{Weight: 30, Align: ui.LabelAlignRight, Nodes: []ui.Node{countLbl}},
	})

	root.AddChild(body)
	root.AddChild(status)

	s.rebuildList()
	s.refreshDetail()
	FinishShellMount(doc)
}

func (s *listPaneScene) rebuildList() {
	if s.pane == nil || s.pane.List == nil {
		return
	}
	for _, ch := range append([]ui.Node(nil), s.pane.List.Children()...) {
		if strings.HasPrefix(ch.ID(), "listpane-row-") {
			s.pane.List.RemoveChild(ch.ID())
		}
	}
	for _, d := range s.docs {
		doc := d
		tile := ui.NewListTile("listpane-row-"+doc.ID, doc.title(), doc.subtitle(), 0, 0, 0, 0)
		tile.Selected = doc.ID == s.selectedID
		tile.OnClick = func() {
			s.selectedID = doc.ID
			s.rebuildList()
			s.refreshDetail()
		}
		s.pane.List.AddChild(tile)
	}
	s.pane.List.MarkDirty()
	if s.pane.Scroll != nil {
		s.pane.Scroll.MarkDirty()
	}
}

func (s *listPaneScene) refreshDetail() {
	var sel *listPaneDoc
	for i := range s.docs {
		if s.docs[i].ID == s.selectedID {
			sel = &s.docs[i]
			break
		}
	}
	if sel == nil {
		return
	}
	if s.detailTitle != nil {
		s.detailTitle.Text.Set(sel.title())
	}
	if s.detailMeta != nil {
		meta := "Path: —"
		if sel.Path != "" {
			meta = "Path: " + sel.Path
		}
		if sel.Unsaved {
			meta += " · Unsaved"
		}
		s.detailMeta.Text.Set(meta)
	}
	if s.detailBody != nil {
		body := sel.Body
		if body == "" {
			body = "No preview body for this document."
		}
		s.detailBody.Text.Set(body)
	}
	if s.statusSel != nil {
		line := sel.title()
		if sel.Unsaved {
			line += " · Unsaved"
		}
		s.statusSel.Text.Set(line)
	}
}

func (s *listPaneScene) OnUpdate(_ *ui.Document, _ float32) {}

// truncateMiddle shortens long paths for ListTile subtitles (AP-07 / CP-LIST-LISTTILE-01).
func truncateMiddle(s string, maxRunes int) string {
	if maxRunes < 8 || utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	keep := (maxRunes - 1) / 2
	return string(runes[:keep]) + "…" + string(runes[len(runes)-keep:])
}
