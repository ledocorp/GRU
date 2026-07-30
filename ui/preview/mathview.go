package preview

import (
	"github.com/ledocorp/gru/ui"
)

// MathView renders LaTeX to GPU textures via go-latex (in-memory cache, no disk files).
type MathView struct {
	*ui.Container
	source  string
	display bool
	doc     *ui.Document
}

// NewMathView creates a math preview host.
func NewMathView(id string, x, y, w, h float32) *MathView {
	root := ui.NewContainer(id, x, y, w, h)
	root.LayoutType = ui.LayoutFlex
	root.FlexDirection = ui.FlexColumn
	root.SetStyle("transparent")
	root.Gap = 12
	root.AutoHeight = true
	return &MathView{Container: root}
}

// SetDocument enables async PNG render + texture upload on the main thread.
func (m *MathView) SetDocument(doc *ui.Document) {
	m.doc = doc
	if m.source != "" {
		m.rebuild()
	}
}

// SetLatex sets the TeX source and display mode (block vs inline).
func (m *MathView) SetLatex(source string, display bool) {
	m.source = source
	m.display = display
	m.rebuild()
}

func (m *MathView) rebuild() {
	for _, ch := range m.Children() {
		m.RemoveChild(ch.ID())
	}
	label := "Inline math"
	if m.display {
		label = "Display math"
	}
	hdr := ui.NewRichText(m.ID()+"-label", []ui.TextSpan{
		{Text: label, Bold: true, Variant: "muted"},
	}, 0, 0, 0, 0)
	hdr.AutoHeight = true
	m.AddChild(hdr)
	m.AddChild(BuildMathPreviewBlockWithDoc(m.ID()+"-block", m.source, m.display, m.doc))
	m.MarkDirty()
}
