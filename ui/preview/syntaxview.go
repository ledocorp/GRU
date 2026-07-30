package preview

import (
	"github.com/ledocorp/gru/ui"
)

// SyntaxView shows editable source with a Chroma-highlighted mirror pane below.
//
// Example:
//
//	sv := preview.NewSyntaxView("code", 0, 0, 600, 400)
//	sv.SetLanguage("go")
//	sv.SetText("package main\n")
type SyntaxView struct {
	*ui.Container
	Editor    *ui.TextEditor
	Highlight *ui.RichText
	language  string
	dirty     bool
}

// NewSyntaxView creates an editor + highlighted preview stack.
func NewSyntaxView(id string, x, y, w, h float32) *SyntaxView {
	root := ui.NewContainer(id, x, y, w, h)
	root.LayoutType = ui.LayoutFlex
	root.FlexDirection = ui.FlexColumn
	root.SetStyle("transparent")
	root.Gap = 8
	root.SetFlexGrow(1)

	ed := ui.NewTextEditor(id+"-editor", "", 0, 0, 0, 0)
	edStyle := ui.GetThemeStyle("text-editor")
	edStyle.Mono = true
	edStyle.FontSize = 15
	ed.SetStyleOverrides(edStyle)
	ed.SetFlexGrow(1)
	ed.WordWrap = false
	ed.AutoHeight = false

	hlVP := ui.NewViewport(id+"-hl-vp", 0, 0, 0, 0)
	hlVP.SetStyle("panel")
	hlVP.SetFlexGrow(1)
	hlVP.Gap = 0

	hl := highlightedCodeRichText(id+"-hl", "go", "")
	hlVP.AddChild(hl)

	root.AddChild(ed)
	root.AddChild(hlVP)

	sv := &SyntaxView{
		Container: root,
		Editor:    ed,
		Highlight: hl,
		language:  "go",
		dirty:     true,
	}
	ed.OnChange = func() { sv.dirty = true }
	return sv
}

// SetLanguage sets the Chroma lexer name (e.g. "go", "python").
func (s *SyntaxView) SetLanguage(lang string) {
	s.language = lang
	s.dirty = true
}

// SetText replaces editor buffer and refreshes highlight.
func (s *SyntaxView) SetText(text string) {
	s.Editor.SetTextContent(text)
	s.dirty = true
}

func (s *SyntaxView) refreshHighlight() {
	s.Highlight.Spans = highlightSpans(s.Editor.Text.Get(), s.language)
	s.Highlight.MarkDirty()
}

// Update refreshes the highlight mirror when the editor changed.
func (s *SyntaxView) Update(dt float32) {
	s.Container.Update(dt)
	if !s.dirty {
		return
	}
	s.dirty = false
	s.refreshHighlight()
}

// Layout runs layout then ensures highlight viewport height.
func (s *SyntaxView) Layout() {
	s.Container.Layout()
}

// Draw draws the syntax view subtree.
func (s *SyntaxView) Draw() {
	if s.dirty {
		s.refreshHighlight()
		s.dirty = false
	}
	s.Container.Draw()
}
