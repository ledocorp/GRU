package preview

import (
	"strings"

	"github.com/ledocorp/gru/ui"
)

// MathEquation renders LaTeX source to a GPU texture via go-latex (in-memory cache).
type MathEquation struct {
	*ui.Container
	img        *ui.Image
	status     *ui.RichText
	sourceLine *ui.RichText
	source     string
	display   bool
	doc       *ui.Document
	key       string
	pending   bool
	naturalW  int32
	naturalH  int32
	minFrameH float32
}

// NewMathEquation creates a math preview block (call SetDocument then SetLatex).
func NewMathEquation(id string, minFrameH float32) *MathEquation {
	root := ui.NewContainer(id, 0, 0, 0, 0)
	root.LayoutType = ui.LayoutFlex
	root.FlexDirection = ui.FlexColumn
	root.SetStyle("transparent")
	root.Gap = 6
	root.AutoHeight = true

	status := ui.NewRichText(id+"-status", []ui.TextSpan{
		{Text: "Rendering…", Variant: "muted"},
	}, 0, 0, 0, 0)
	status.AutoHeight = true

	frameH := minFrameH
	if frameH <= 0 {
		frameH = 48
	}
	img := ui.NewImage(id+"-img", "", 0, 0, 0, frameH)
	img.FitMode = ui.FitContain

	root.AddChild(status)
	root.AddChild(img)

	return &MathEquation{
		Container: root,
		img:       img,
		status:    status,
		minFrameH: frameH,
	}
}

// SetDocument enables async render/upload (required for non-blocking preview).
func (m *MathEquation) SetDocument(doc *ui.Document) {
	if m.doc == doc {
		return
	}
	m.doc = doc
	if m.source != "" {
		m.requestRender()
	}
}

// SetSourceLine optionally hides the TeX source RichText after a successful render.
func (m *MathEquation) SetSourceLine(rt *ui.RichText) {
	m.sourceLine = rt
}

// SetLatex sets the TeX source and starts render (display = larger font).
func (m *MathEquation) SetLatex(source string, display bool) {
	m.source = strings.TrimSpace(source)
	m.display = display
	m.key = mathCacheKey(m.source, m.display)
	m.pending = false
	m.requestRender()
}

func (m *MathEquation) requestRender() {
	if m.source == "" {
		m.setStatus("No math source", true)
		return
	}
	if e, ok := globalMathTextureCache.get(m.key); ok {
		m.applyTexture(e)
		m.setStatus("", false)
		return
	}
	if m.doc == nil {
		m.setStatus("", false)
		return
	}
	if m.pending {
		return
	}
	m.pending = true
	m.setStatus("Rendering…", true)
	src := m.source
	display := m.display
	key := m.key
	doc := m.doc
	ui.SubmitAsyncBg(func() {
		png, err := renderLatexPNG(src, display)
		doc.QueueMain(func() {
			if m.key != key {
				return
			}
			m.pending = false
			if err != nil {
				m.setError(err)
				return
			}
			tex, w, h := pngBytesToTexture(png)
			if tex.ID == 0 {
				m.setStatus("Render failed", true)
				return
			}
			e := mathTextureEntry{tex: tex, w: w, h: h}
			globalMathTextureCache.put(key, e)
			m.applyTexture(e)
			m.setStatus("", false)
		})
	})
}

func (m *MathEquation) applyTexture(e mathTextureEntry) {
	m.naturalW = e.w
	m.naturalH = e.h
	m.img.SetGPUTexture(e.tex)
	frameH := m.minFrameH
	if m.naturalW > 0 && m.naturalH > 0 {
		const refW float32 = 163
		scale := refW / float32(m.naturalW)
		if scale > 1 {
			scale = 1
		}
		h := float32(m.naturalH) * scale
		if h < m.minFrameH {
			h = m.minFrameH
		}
		if h > 90 {
			h = 90
		}
		frameH = h
	}
	b := m.img.Bounds()
	b.Height = frameH
	m.img.SetBounds(b)
	m.img.MarkDirty()
	if m.sourceLine != nil {
		m.sourceLine.Hide()
	}
	m.MarkDirty()
	// Texture height changes card intrinsic size — without this, math looks
	// clipped/stacked until the user nudges the window.
	ui.MarkAutoHeightHostDirty(m)
}

func (m *MathEquation) setError(err error) {
	if m.sourceLine != nil {
		m.sourceLine.Show()
	}
	msg := "TeX preview unavailable"
	if err != nil {
		e := err.Error()
		switch {
		case strings.Contains(e, "ast.Sup"), strings.Contains(e, "ast.Sub"):
			msg = "TeX preview: superscripts/subscripts not supported yet — source shown above"
		case strings.Contains(e, "not a valid unicode"), strings.Contains(e, "known TeX symbol"):
			msg = "TeX preview: unsupported symbol — source shown above"
		default:
			if len(e) > 100 {
				e = e[:97] + "..."
			}
			msg = "TeX preview: " + e
		}
	}
	m.setStatus(msg, true)
}

func (m *MathEquation) setStatus(text string, show bool) {
	if m.status == nil {
		return
	}
	if !show || text == "" {
		m.status.Hide()
		ui.MarkAutoHeightHostDirty(m)
		return
	}
	m.status.Show()
	m.status.Spans = []ui.TextSpan{{Text: text, Variant: "muted"}}
	m.status.MarkDirty()
	ui.MarkAutoHeightHostDirty(m)
}
