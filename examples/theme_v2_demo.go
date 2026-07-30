//go:build !notepad

package examples

import (
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &themeV2Scene{} }) }

type themeV2Scene struct {
	BaseScene
}

func (s *themeV2Scene) Title() string { return "Theme v2 Foundation" }

func (s *themeV2Scene) OnUpdate(d *ui.Document, _ float32) {
	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		return
	}
	focusClickedTextInput(d)
}

func (s *themeV2Scene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "theme-v2",
		"Theme v2 Foundation",
		"Component + variant style resolution layered over the existing theme.")
	page.Body.Gap = 12

	grid := ui.NewContainer("theme-v2-grid", 0, 0, 0, 0)
	grid.LayoutType = ui.LayoutGrid
	grid.GridColumns = 12
	grid.Gap = 12
	grid.SetStyle("page-grid")
	grid.SetFlexGrow(1)

	grid.AddChild(themeV2ButtonPanel())
	grid.AddChild(themeV2BadgePanel())
	grid.AddChild(themeV2ControlsPanel())
	grid.AddChild(themeV2NotesPanel())

	page.Body.AddChild(grid)
	FinishShellMount(doc)
}

func themeV2Panel(id, title string) *ui.Panel {
	// AutoHeight — fixed 180–250 panels clipped button/badge/input labels.
	p := ui.NewPanel(id, title, 0, 0, 0, 0)
	p.AutoHeight = true
	p.Gap = 10
	p.TitleHeight = 32
	p.SetColSpan(ui.BreakpointXS, 12)
	p.SetColSpan(ui.BreakpointSM, 12)
	p.SetColSpan(ui.BreakpointMD, 6)
	p.SetColSpan(ui.BreakpointLG, 6)
	p.SetColSpan(ui.BreakpointXL, 6)
	return p
}

func themeV2Hint(id, text string) *ui.RichText {
	return ui.NewPlainText(id, "form-label", text, 0, 0, 0, 0)
}

func themeV2ButtonPanel() *ui.Panel {
	p := themeV2Panel("theme-v2-buttons", "Button Variants")
	p.AddChild(themeV2Hint("theme-v2-buttons-hint",
		"These buttons use SetStyleVariant(component, variant)."))

	row1 := themeV2Row("theme-v2-button-row-1")
	row1.AddChild(themeV2Button("theme-v2-default", "Default", "button", "default"))
	row1.AddChild(themeV2Button("theme-v2-primary", "Primary", "button", "primary"))

	row2 := themeV2Row("theme-v2-button-row-2")
	row2.AddChild(themeV2Button("theme-v2-danger", "Danger", "button", "danger"))
	row2.AddChild(themeV2Button("theme-v2-legacy", "Legacy SetStyle", "", ""))

	p.AddChild(row1)
	p.AddChild(row2)
	return p
}

func themeV2Row(id string) *ui.Container {
	row := ui.NewContainer(id, 0, 0, 0, 0)
	row.FlexDirection = ui.FlexRow
	row.Gap = 10
	row.AutoHeight = true
	row.SetStyle("transparent")
	return row
}

func themeV2Button(id, label, component, variant string) *ui.Button {
	btn := ui.NewButton(id, label, 0, 0, 0, 40)
	btn.SetFlexGrow(1)
	if component != "" {
		btn.SetStyleVariant(component, variant)
	} else {
		btn.SetStyle("primary")
	}
	return btn
}

func themeV2BadgePanel() *ui.Panel {
	p := themeV2Panel("theme-v2-badges", "Badge Variants")
	p.AddChild(themeV2Hint("theme-v2-badges-hint",
		"Badges resolve through the badge component variants."))

	row := themeV2Row("theme-v2-badge-row")
	row.AddChild(ui.NewBadge("theme-v2-badge-default", "Default", ui.BadgeDefault, 0, 0, 0, 24))
	row.AddChild(ui.NewBadge("theme-v2-badge-primary", "Primary", ui.BadgePrimary, 0, 0, 0, 24))
	row.AddChild(ui.NewBadge("theme-v2-badge-success", "Success", ui.BadgeSuccess, 0, 0, 0, 24))
	row.AddChild(ui.NewBadge("theme-v2-badge-danger", "Danger", ui.BadgeDanger, 0, 0, 0, 24))
	p.AddChild(row)
	return p
}

func themeV2ControlsPanel() *ui.Panel {
	p := themeV2Panel("theme-v2-controls", "Input + Dropdown")
	p.AddChild(themeV2Hint("theme-v2-controls-hint",
		"TextInput and Dropdown opt into Theme v2 defaults and state styles."))

	input := ui.NewTextInput("theme-v2-input", "Theme v2 input", 0, 0, 0, 0)
	p.AddChild(input)

	dropdown := ui.NewDropdown("theme-v2-dropdown", []string{"Default", "Primary", "Danger"}, 0, 0, 0, 0, 40)
	p.AddChild(dropdown)

	disabled := ui.NewDropdown("theme-v2-dropdown-disabled", []string{"Disabled dropdown"}, 0, 0, 0, 0, 40)
	disabled.Disabled = true
	p.AddChild(disabled)
	return p
}

func themeV2NotesPanel() *ui.Panel {
	p := themeV2Panel("theme-v2-notes", "Notes")
	p.AddChild(themeV2Hint("theme-v2-notes-1",
		"New JSON can use component + variant + overrides."))
	p.AddChild(themeV2Hint("theme-v2-notes-2",
		"Button/IconButton, Badge, Panel/Card, TextInput, and Dropdown use resolver-backed variants."))
	return p
}
