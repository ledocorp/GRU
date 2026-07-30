//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"strings"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &formScene{} }) }

// formScene demonstrates TextInput, Slider, Checkbox, Toggle, Dropdown, and Form.
// Recipe: CP-SHELL-PAGE + NewBatchPageGrid (LAYOUT_CONTRACTS §8).
type formScene struct {
	BaseScene
	focusable []ui.Node
}

func (s *formScene) Title() string { return "Form Demo" }

func setSpans5FormPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func formHint(id, text string) *ui.RichText {
	return batchCaption(id, text)
}

func formPanel(id, title string) *ui.Panel {
	p := ui.NewPanel(id, title, 0, 0, 0, 0)
	p.AutoHeight = true
	p.Gap = 10
	p.TitleHeight = 32
	return p
}

func buttonRow(id string) *ui.Container {
	row := ui.NewContainer(id, 0, 0, 0, 40)
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.Gap = 8
	row.AutoHeight = false
	row.SetStyle("transparent")
	return row
}

func (s *formScene) Build(doc *ui.Document) {
	s.focusable = s.focusable[:0]

	page := MountAppPage(doc, "form",
		"Form Demo",
		"Profile, preferences, and account — TextInput, Slider, Checkbox, Toggle, Dropdown, Form.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("form-grid", 12)

	nameInput := ui.NewTextInput("form-name", "", 0, 0, 0, 0)
	emailInput := ui.NewTextInput("form-email", "", 0, 0, 0, 0)
	volumeSlider := ui.NewSlider("form-volume", 0, 100, 50, 0, 0, 0, 0)
	notifyCb := ui.NewCheckbox("form-notify", true, 0, 0, ui.CheckboxDefaultSize, ui.CheckboxDefaultSize)
	digestToggle := ui.NewToggle("form-digest", false, 0, 0, 52, 28)
	darkToggle := ui.NewToggle("form-dark", false, 0, 0, 52, 28)
	betaToggle := ui.NewToggle("form-beta", false, 0, 0, 52, 28)
	regionOpts := []string{"Americas", "Europe", "Asia Pacific", "Africa"}
	regionDD := ui.NewDropdown("form-region", regionOpts, 0, 0, 0, 0, 36)

	status := ui.NewSignal("Ready")
	statusLbl := FlexCopyMirror("form-status", "form-label", status, "Status: ")

	profilePanel, profileFields := s.buildProfilePanel(nameInput, emailInput)
	prefsPanel := s.buildPreferencesPanel(volumeSlider, notifyCb, digestToggle)
	accountPanel := s.buildAccountPanel(regionDD, darkToggle, betaToggle)
	reviewPanel := s.buildReviewPanel(nameInput, emailInput, volumeSlider, regionDD, regionOpts, profileFields, status, statusLbl)

	grid.AddChild(profilePanel)
	grid.AddChild(prefsPanel)
	grid.AddChild(accountPanel)
	grid.AddChild(reviewPanel)
	page.Body.AddChild(grid)

	s.focusable = []ui.Node{nameInput, emailInput, volumeSlider, notifyCb, digestToggle, regionDD, darkToggle, betaToggle}
}

func (s *formScene) buildProfilePanel(nameInput, emailInput *ui.TextInput) (*ui.Panel, *ui.Form) {
	p := formPanel("form-profile-panel", "Profile")
	setSpans5FormPanel(p, 12, 12, 12, 6, 6)
	p.AddChild(formHint("form-profile-hint", "Two-column form layout with full-width text inputs."))

	f := ui.NewForm("form-profile-fields", 0, 0, 0, 0)
	f.AutoHeight = true
	f.LabelW = 72
	f.RowH = 36
	f.Gap = 8
	f.FieldGap = 10
	f.AddField("Name", nameInput)
	f.AddField("Email", emailInput)
	p.AddChild(f)
	return p, f
}

func (s *formScene) buildPreferencesPanel(volumeSlider *ui.Slider, notifyCb *ui.Checkbox, digestToggle *ui.Toggle) *ui.Panel {
	p := formPanel("form-preferences-panel", "Preferences")
	setSpans5FormPanel(p, 12, 12, 12, 6, 6)
	p.AddChild(formHint("form-preferences-hint", "Checkbox and toggle sizing, with a live volume readout."))

	volumeLbl, volumeDisplay := FlexCopyPair("form-volume-label", "form-label", "Volume: 50")
	volumeBar := ui.NewProgressBar("form-volume-bar", 0.5, 0, 0, 0, 12)
	ui.NewEffect(func() {
		v := volumeSlider.Value.Get()
		volumeDisplay.Set(fmt.Sprintf("Volume: %.0f", v))
		volumeBar.Value.Set(v / 100)
	})

	f := ui.NewForm("form-preferences-fields", 0, 0, 0, 0)
	f.AutoHeight = true
	f.LabelW = 108
	f.RowH = 36
	f.Gap = 8
	f.FieldGap = 10
	f.AddField("Volume", volumeSlider)
	f.AddFieldSized("Notify", notifyCb, ui.CheckboxDefaultSize, ui.CheckboxDefaultSize)
	f.AddFieldSized("Daily digest", digestToggle, 52, 28)

	// Size samples for visual QA (compact → default). No 16px — too small to read.
	sizeRow := ui.NewContainer("form-cb-sizes", 0, 0, 0, 0)
	sizeRow.LayoutType = ui.LayoutFlex
	sizeRow.FlexDirection = ui.FlexRow
	sizeRow.Gap = 12
	sizeRow.AutoHeight = true
	sizeRow.SetStyle("transparent")
	for _, sz := range []float32{18, 20, 24} {
		cb := ui.NewCheckbox(fmt.Sprintf("form-cb-sz-%.0f", sz), true, 0, 0, sz, sz)
		sizeRow.AddChild(cb)
		sizeRow.AddChild(ui.NewPlainText(fmt.Sprintf("form-cb-sz-lbl-%.0f", sz), "form-label",
			fmt.Sprintf("%.0fpx", sz), 0, 0, 0, 0))
	}

	p.AddChild(volumeLbl)
	p.AddChild(volumeBar)
	p.AddChild(f)
	p.AddChild(batchCaption("form-cb-size-hint", "Checkbox sizes: 18 / 20 / 24 (Notify uses 24)."))
	p.AddChild(sizeRow)
	return p
}

func (s *formScene) buildAccountPanel(regionDD *ui.Dropdown, darkToggle, betaToggle *ui.Toggle) *ui.Panel {
	p := formPanel("form-account-panel", "Account")
	setSpans5FormPanel(p, 12, 12, 12, 6, 6)
	p.AddChild(formHint("form-account-hint", "Dropdown plus compact account toggles."))

	f := ui.NewForm("form-account-fields", 0, 0, 0, 0)
	f.AutoHeight = true
	f.LabelW = 88
	f.RowH = 36
	f.Gap = 8
	f.FieldGap = 10
	f.AddField("Region", regionDD)
	f.AddFieldSized("Dark mode", darkToggle, 52, 28)
	f.AddFieldSized("Beta UI", betaToggle, 52, 28)
	p.AddChild(f)
	return p
}

func (s *formScene) buildReviewPanel(
	nameInput, emailInput *ui.TextInput,
	volumeSlider *ui.Slider,
	regionDD *ui.Dropdown,
	regionOpts []string,
	profileForm *ui.Form,
	status *ui.Signal[string],
	statusLbl *ui.RichText,
) *ui.Panel {
	p := formPanel("form-review-panel", "Submit And Validation")
	setSpans5FormPanel(p, 12, 12, 12, 6, 6)
	p.AddChild(formHint("form-review-hint", "Submit validates required fields and reports current state."))

	actionRow := buttonRow("form-action-row")
	submitBtn := ui.NewButton("form-submit", "Submit", 0, 0, 112, 40)
	submitBtn.SetStyle("primary")
	clearBtn := ui.NewButton("form-clear", "Clear", 0, 0, 96, 40)
	clearBtn.SetStyle("button")

	submitBtn.OnClick = func() {
		profileForm.ClearAllErrors()
		name := strings.TrimSpace(nameInput.Text.Get())
		email := strings.TrimSpace(emailInput.Text.Get())
		if name == "" {
			profileForm.SetError("Name", "Name is required")
			status.Set("Name is required")
			return
		}
		if email == "" || !strings.Contains(email, "@") {
			profileForm.SetError("Email", "Enter a valid email address")
			status.Set("Valid email is required")
			return
		}
		region := regionOpts[regionDD.SelectedIndex.Get()]
		status.Set(fmt.Sprintf("Submitted %s / %s / %s / volume %.0f", name, email, region, volumeSlider.Value.Get()))
	}
	clearBtn.OnClick = func() {
		profileForm.ClearAllErrors()
		nameInput.Text.Set("")
		emailInput.Text.Set("")
		status.Set("Ready")
	}

	actionRow.AddChild(submitBtn)
	actionRow.AddChild(clearBtn)
	p.AddChild(actionRow)
	p.AddChild(statusLbl)

	summary, summaryDisplay := FlexCopyPair("form-summary", "form-value", "Summary: waiting for input")
	ui.NewEffect(func() {
		name := strings.TrimSpace(nameInput.Text.Get())
		email := strings.TrimSpace(emailInput.Text.Get())
		if name == "" {
			name = "(no name)"
		}
		if email == "" {
			email = "(no email)"
		}
		summaryDisplay.Set(fmt.Sprintf("Summary: %s, %s", name, email))
	})
	p.AddChild(summary)
	return p
}

func (s *formScene) OnUpdate(d *ui.Document, _ float32) {
	if ui.OverlayBlocksSceneInput() {
		return
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		mouse := rl.GetMousePosition()
		matched := false
		for _, node := range s.focusable {
			if rl.CheckCollisionPointRec(mouse, node.Bounds()) {
				d.SetFocus(node)
				matched = true
				break
			}
		}
		if !matched {
			d.SetFocus(nil)
		}
	}
}
