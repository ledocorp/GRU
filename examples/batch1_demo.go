//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &batch1Scene{} }) }

// batch1Scene demonstrates all 8 Batch 1 widgets:
// Tooltip, TabView, Modal, Spinner, RadioGroup, TreeView, Form, ContextMenu.
type batch1Scene struct {
	BaseScene
	spinner   *ui.Spinner
	ctxPanel  *ui.Panel
	ctxStatus     *ui.RichText
	ctxStatusText *ui.Signal[string]
	focusable []ui.Node
}

func (s *batch1Scene) Title() string { return "Batch 1 · Tooltip / TabView / Modal" }

// OnUpdate handles right-click for context menu and left-click for text input focus.
func (s *batch1Scene) OnUpdate(d *ui.Document, _ float32) {
	if ui.OverlayBlocksSceneInput() {
		return
	}
	mouse := rl.GetMousePosition()

	// Left-click: route focus to the clicked TextInput.
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		focused := false
		for _, node := range s.focusable {
			if rl.CheckCollisionPointRec(mouse, node.Bounds()) {
				d.SetFocus(node)
				focused = true
				break
			}
		}
		if !focused {
			if hit := ui.FindInteractiveAt(d.Root, mouse); hit != nil {
				d.SetFocus(hit)
			} else {
				d.SetFocus(nil)
			}
		}
	}

	// Right-click: open context menu inside the ContextMenu panel.
	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		if s.ctxPanel != nil && rl.CheckCollisionPointRec(mouse, s.ctxPanel.Bounds()) {
			s.showCtxMenu(mouse.X, mouse.Y)
		}
	}
}

func (s *batch1Scene) showCtxMenu(x, y float32) {
	ui.ShowContextMenu([]ui.ContextMenuItem{
		{Label: "New File", Action: func() { s.ctxStatusText.Set("New File") }},
		{Label: "Open Folder…", Action: func() { s.ctxStatusText.Set("Open Folder…") }},
		{Divider: true},
		{Label: "Share", SubItems: []ui.ContextMenuItem{
			{Label: "Copy Link", Action: func() { s.ctxStatusText.Set("Copy Link") }},
			{Label: "Send via Email", Action: func() { s.ctxStatusText.Set("Send via Email") }},
			{Divider: true},
			{Label: "Export as PDF", Action: func() { s.ctxStatusText.Set("Export as PDF") }},
		}},
		{Label: "Copy Path", Action: func() { s.ctxStatusText.Set("Copy Path") }},
		{Label: "Rename…", Action: func() { s.ctxStatusText.Set("Rename…") }},
		{Divider: true},
		{Label: "Delete", Action: func() { s.ctxStatusText.Set("Delete") }},
		{Label: "Export (disabled)", Disabled: true},
	}, x, y)
}

// setSpans5 assigns 12-column spans for all breakpoint tiers. layoutGrid uses
// CurrentBreakpoint(container width) so this matches CSS container-query style
// reflow (see ui.CurrentBreakpoint / docs responsive tiers).
func setSpans5(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func b1Body(id, text string) *ui.RichText {
	return FlexCopy(id, "form-value", text)
}

func b1Status(id, text string) (*ui.RichText, *ui.Signal[string]) {
	return FlexCopyPair(id, "default", text)
}

func (s *batch1Scene) Build(doc *ui.Document) {
	PreloadScenePhosphor(doc)
	page := MountAppPage(doc, "b1",
		"Widget Batch 1 · Tooltip / TabView / Modal / …",
		"Tooltip · TabView · Modal · Spinner · RadioGroup · TreeView · Form · ContextMenu")
	page.Body.Gap = 16

	// Full-width 12-column grid: horizontal reflow only (no page-level horizontal
	// scroll). Vertical overflow scrolls in this viewport. Use HorizontalViewport
	// inside a widget when a region has a deliberate min content width.
	grid := NewBatchPageGrid("b1-grid", 16)
	// layoutGrid uses gridBreakpointForWidth: below ui.MinClientWidth all cells use
	// xs ColSpan so panels stack instead of squeezing side-by-side.

	const row1H float32 = 360
	const row2H float32 = 340
	const formRowH float32 = 400

	// ── Row 1 ─ Tooltip, TabView, Modal ───────────────────────────────────────
	pTooltip := ui.NewPanel("p-tooltip", "Tooltip", 0, 0, 0, row1H)
	setSpans5(pTooltip, 12, 12, 6, 3, 3)
	pTooltip.TitleHeight = 32
	pTooltip.Gap = 8

	pTooltip.AddChild(batchCaption("tip-cap", "Hover any button below to see a tooltip."))
	tipBtn1 := ui.NewButton("tip-btn1", "Save File", 0, 0, 0, 40)
	tipBtn1.SetStyle("default")
	tipBtn2 := ui.NewButton("tip-btn2", "Delete All", 0, 0, 0, 40)
	tipBtn2.SetStyle("default")
	tipBtn3 := ui.NewButton("tip-btn3", "More Info…", 0, 0, 0, 40)
	tipBtn3.SetStyle("default")

	ui.SetTooltip(tipBtn1, "Save the current file to disk")
	ui.SetTooltip(tipBtn2, "Permanently delete all items — cannot be undone")
	ui.SetTooltip(tipBtn3, "Open the help documentation in your browser")

	pTooltip.AddChild(tipBtn1)
	pTooltip.AddChild(tipBtn2)
	pTooltip.AddChild(tipBtn3)

	pTabView := ui.NewPanel("p-tabview", "TabView", 0, 0, 0, row1H)
	setSpans5(pTabView, 12, 12, 6, 6, 6)
	pTabView.TitleHeight = 32
	pTabView.Gap = 8

	tv := ui.NewTabView("demo-tabs", 0, 0, 0, 0)
	tv.SetFlexGrow(1)

	tab1Content := ui.NewContainer("tab1-c", 0, 0, 0, 0)
	tab1Content.FlexDirection = ui.FlexColumn
	tab1Content.Gap = 8
	tab1Content.SetFlexGrow(1)
	tab1Content.AddChild(b1Body("t1l1", "Overview tab is active."))
	tab1Content.AddChild(b1Body("t1l2", "Switch tabs using the header above."))

	tab2Content := ui.NewContainer("tab2-c", 0, 0, 0, 0)
	tab2Content.FlexDirection = ui.FlexColumn
	tab2Content.Gap = 8
	tab2Content.SetFlexGrow(1)
	tab2Content.AddChild(b1Body("t2l1", "Settings tab is active."))
	chk := ui.NewCheckbox("t2-chk", false, 0, 0, 24, 24)
	tab2Content.AddChild(chk)

	tab3Content := ui.NewContainer("tab3-c", 0, 0, 0, 0)
	tab3Content.FlexDirection = ui.FlexColumn
	tab3Content.Gap = 8
	tab3Content.SetFlexGrow(1)
	tab3Content.AddChild(b1Body("t3l1", "Analytics tab."))
	pb := ui.NewProgressBar("t3-pb", 0.65, 0, 0, 0, 24)
	pb.SetFlexGrow(1)
	tab3Content.AddChild(pb)

	tv.AddTab("Overview", tab1Content)
	tv.AddTab("Settings", tab2Content)
	tv.AddTab("Analytics", tab3Content)

	pTabView.AddChild(tv)

	pModal := ui.NewPanel("p-modal", "Modal", 0, 0, 0, row1H)
	setSpans5(pModal, 12, 12, 12, 3, 3)
	pModal.TitleHeight = 32
	pModal.Gap = 8

	modalStatus, modalStatusText := b1Status("modal-status", "No action taken yet.")

	openModalBtn := ui.NewButton("open-modal", "Open Modal", 0, 0, 0, 40)
	openModalBtn.SetStyle("default")
	openModalBtn.OnClick = func() {
		confirmContent := FlexCopy("modal-body", "modal-body",
			"Are you sure you want to proceed? This action cannot be undone.")

		ui.ShowModalSized("Confirm Action", confirmContent, []ui.ModalButton{
			{Label: "Confirm", Action: func() {
				modalStatusText.Set("Action confirmed!")
				ui.CloseModal()
			}},
			{Label: "Cancel", Action: func() {
				modalStatusText.Set("Action cancelled.")
				ui.CloseModal()
			}},
		}, 420, 280)
	}

	pModal.AddChild(batchCaption("modal-cap", "Click the button to open a blocking modal dialog."))
	pModal.AddChild(openModalBtn)
	pModal.AddChild(modalStatus)

	// ── Row 2 ─ Spinner, RadioGroup, TreeView, ContextMenu (12 cols) ───────────
	pSpinner := ui.NewPanel("p-spinner", "Spinner", 0, 0, 0, row2H)
	setSpans5(pSpinner, 12, 12, 6, 2, 2)
	pSpinner.TitleHeight = 32
	pSpinner.Gap = 8

	sp := ui.NewSpinner("demo-spinner", 0, 0, 60)
	// Paused by default so idle benchmarks cache-hit; click Resume to demo overlay animation.
	s.spinner = sp

	spinToggleBtn := ui.NewButton("spin-toggle", "Resume", 0, 0, 0, 36)
	spinToggleBtn.SetStyle("default")
	spinToggleBtn.OnClick = func() {
		active := !sp.Active.Get()
		sp.Active.Set(active)
		if active {
			spinToggleBtn.Text.Set("Pause")
		} else {
			spinToggleBtn.Text.Set("Resume")
		}
	}

	pSpinner.AddChild(sp)
	pSpinner.AddChild(spinToggleBtn)

	pRadio := ui.NewPanel("p-radio", "RadioGroup", 0, 0, 0, row2H)
	setSpans5(pRadio, 12, 12, 6, 3, 3)
	pRadio.TitleHeight = 32
	pRadio.Gap = 8

	pRadio.AddChild(batchCaption("radio-cap", "Select a theme:"))

	rg := ui.NewRadioGroup("theme-radio",
		[]string{"System Default", "Light Mode", "Dark Mode", "High Contrast"},
		0, 0, 0, 0)
	rg.SetFlexGrow(1)
	rg.Selected.Set(0)

	radioStatus, radioStatusText := b1Status("radio-status", "Selected: System Default")
	rg.Selected.Subscribe(func() {
		idx := rg.Selected.Get()
		opts := rg.Options
		if idx >= 0 && idx < len(opts) {
			radioStatusText.Set(fmt.Sprintf("Selected: %s", opts[idx]))
		}
	})

	pRadio.AddChild(rg)
	pRadio.AddChild(radioStatus)

	pTree := ui.NewPanel("p-tree", "TreeView", 0, 0, 0, row2H)
	setSpans5(pTree, 12, 12, 6, 3, 3)
	pTree.TitleHeight = 32
	pTree.Gap = 8

	rootNode := ui.NewTreeNode("Project", "Project")
	rootNode.SetExpanded(true)
	srcNode := ui.NewTreeNode("src", "src")
	srcNode.SetExpanded(true)
	srcNode.AddChild(ui.NewTreeNode("main.go", "main.go"))
	srcNode.AddChild(ui.NewTreeNode("config.go", "config.go"))
	srcNode.AddChild(ui.NewTreeNode("utils.go", "utils.go"))
	uiNode := ui.NewTreeNode("ui", "ui")
	uiNode.AddChild(ui.NewTreeNode("button.go", "button.go"))
	uiNode.AddChild(ui.NewTreeNode("modal.go", "modal.go"))
	uiNode.AddChild(ui.NewTreeNode("tooltip.go", "tooltip.go"))
	srcNode.AddChild(uiNode)
	rootNode.AddChild(srcNode)
	assetsNode := ui.NewTreeNode("assets", "assets")
	assetsNode.AddChild(ui.NewTreeNode("icon.png", "icon.png"))
	assetsNode.AddChild(ui.NewTreeNode("font.ttf", "font.ttf"))
	rootNode.AddChild(assetsNode)
	rootNode.AddChild(ui.NewTreeNode("go.mod", "go.mod"))
	rootNode.AddChild(ui.NewTreeNode("README.md", "README.md"))

	tree := ui.NewTreeView("demo-tree", rootNode, 0, 0, 0, 0)
	tree.SetFlexGrow(1)

	treeStatus, treeStatusText := b1Status("tree-status", "Click a node to select it.")
	tree.Selected.Subscribe(func() {
		n := tree.Selected.Get()
		if n != nil {
			treeStatusText.Set("Selected: " + n.Label)
		}
	})

	pTree.AddChild(tree)
	pTree.AddChild(treeStatus)

	pCtxMenu := ui.NewPanel("p-ctx", "ContextMenu", 0, 0, 0, row2H)
	setSpans5(pCtxMenu, 12, 12, 6, 4, 4)
	pCtxMenu.TitleHeight = 32
	pCtxMenu.Gap = 8

	pCtxMenu.AddChild(batchCaption("ctx-cap", "Right-click this panel or use the button below to open a context menu."))
	ctxStatus, ctxStatusText := b1Status("ctx-status", "No item selected.")
	s.ctxStatus = ctxStatus
	s.ctxStatusText = ctxStatusText
	s.ctxPanel = pCtxMenu

	openCtxBtn := ui.NewButton("open-ctx", "Open Menu", 0, 0, 0, 36)
	openCtxBtn.SetStyle("default")
	openCtxBtn.OnClick = func() {
		b := openCtxBtn.Bounds()
		s.showCtxMenu(b.X, b.Y+b.Height+4)
	}

	pCtxMenu.AddChild(openCtxBtn)
	pCtxMenu.AddChild(ctxStatus)

	// ── Row 3 ─ Form (full width) with inner vertical scroll ──────────────────
	pForm := ui.NewPanel("p-form", "Form", 0, 0, 0, formRowH)
	setSpans5(pForm, 12, 12, 12, 12, 12)
	pForm.TitleHeight = 32
	pForm.Gap = 8

	formScroll := ui.NewViewport("b1-form-scroll", 0, 0, 0, 0)
	formScroll.SetFlexGrow(1)
	formScroll.Gap = 0
	formScroll.SetStyle("transparent")

	form := ui.NewForm("demo-form", 0, 0, 0, 300)
	form.LabelW = 100
	form.RowH = 36
	form.Gap = 10

	nameInput := ui.NewTextInput("form-name", "Jane Doe", 0, 0, 0, 0)
	emailInput := ui.NewTextInput("form-email", "jane@example.com", 0, 0, 0, 0)
	volSlider := ui.NewSlider("form-vol", 0, 100, 70, 0, 0, 0, 0)
	notifToggle := ui.NewToggle("form-notif", true, 0, 0, 0, 0)

	form.AddField("Name", nameInput)
	form.AddField("Email", emailInput)
	form.AddField("Volume", volSlider)
	form.AddFieldSized("Notifications", notifToggle, 56, 28)

	s.focusable = []ui.Node{nameInput, emailInput}

	formScroll.AddChild(form)
	pForm.AddChild(formScroll)

	grid.AddChild(pTooltip)
	grid.AddChild(pTabView)
	grid.AddChild(pModal)
	grid.AddChild(pSpinner)
	grid.AddChild(pRadio)
	grid.AddChild(pTree)
	grid.AddChild(pCtxMenu)
	grid.AddChild(pForm)

	page.Body.AddChild(grid)
}
