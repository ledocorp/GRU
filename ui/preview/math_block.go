package preview

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/ledocorp/gru/ui"
)

// BuildMathPreviewBlock renders LaTeX to a texture inside a card (source shown as label).
func BuildMathPreviewBlock(id, source string, display bool) ui.Node {
	return BuildMathPreviewBlockWithDoc(id, source, display, nil)
}

// BuildMathPreviewBlockWithDoc renders LaTeX asynchronously when doc is non-nil.
func BuildMathPreviewBlockWithDoc(id, source string, display bool, doc *ui.Document) ui.Node {
	title := "Math"
	if display {
		title = "LaTeX"
	}
	card := ui.NewCard(id, title, 0, 0, 0, 0)
	card.TitleHeight = 28
	card.AutoHeight = true
	card.Gap = 8
	card.SetStyleVariant("card", "code")
	if st, ok := ui.MergedComponentVariantStyle("card", "code"); ok {
		st.Padding = 12
		if display {
			st.BackgroundColor = rl.NewColor(248, 247, 255, 255)
			st.BorderColor = rl.NewColor(221, 214, 254, 255)
		} else {
			st.BackgroundColor = rl.NewColor(252, 251, 255, 255)
			st.BorderColor = rl.NewColor(228, 224, 248, 255)
		}
		card.SetStyleOverrides(st)
	}

	src := strings.TrimSpace(source)
	if src == "" {
		src = "(empty)"
	}

	srcLine := ui.NewRichText(id+"-src", []ui.TextSpan{
		{Text: src, Variant: "math"},
	}, 0, 0, 0, 0)
	srcLine.SetStyle("richtext-math-inline")
	srcLine.AutoHeight = true

	frameH := float32(30)
	if display {
		frameH = 40
	}
	eq := NewMathEquation(id+"-eq", frameH)
	if doc != nil {
		eq.SetDocument(doc)
	}
	eq.SetSourceLine(srcLine)
	eq.SetLatex(src, display)

	card.AddChild(srcLine)
	card.AddChild(eq)
	return card
}

func (ctx *mdBuildCtx) buildMathCard(id, source string, display bool) ui.Node {
	var doc *ui.Document
	if ctx != nil {
		doc = ctx.doc
	}
	return BuildMathPreviewBlockWithDoc(id, source, display, doc)
}
