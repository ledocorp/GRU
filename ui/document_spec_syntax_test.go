package ui

import "testing"

func TestDocSyntaxHighlightDisabledStripsInlineCode(t *testing.T) {
	off := false
	ctx := NewBuildContext()
	ctx.SyntaxHighlight = &off
	spans := docApplySyntaxHighlightSpans([]TextSpan{
		{Text: "Edit ", Variant: "muted"},
		{Text: "pages/gallery.gru", Variant: "code", Bold: true},
	}, ctx)
	if spans[1].Variant != "" {
		t.Fatalf("variant = %q, want plain when highlight off", spans[1].Variant)
	}
}

func TestDocCodeBlockSpansRespectsHighlightFlag(t *testing.T) {
	src := "package main\n\nfunc main() {}"
	on := true
	off := false
	ctxOn := NewBuildContext()
	ctxOn.SyntaxHighlight = &on
	ctxOff := NewBuildContext()
	ctxOff.SyntaxHighlight = &off

	colored := docCodeBlockSpans(src, "go", ctxOn)
	if len(colored) < 2 {
		t.Fatalf("colored spans = %d, want Chroma tokens", len(colored))
	}
	plain := docCodeBlockSpans(src, "go", ctxOff)
	if len(plain) != 1 || plain[0].Text != src {
		t.Fatalf("plain = %+v, want single uncolored span", plain)
	}
}

func TestDocBlockSyntaxHighlightScope(t *testing.T) {
	off := false
	spec := DocumentSpec{
		Children: []DocBlock{{
			Type: "text",
			ID:   "t",
			Spans: []TextSpan{
				{Text: "path.gru", Variant: "code"},
			},
			SyntaxHighlight: &off,
		}},
	}
	root, err := BuildDocumentSpec(spec, NewBuildContext())
	if err != nil {
		t.Fatal(err)
	}
	rt := root.Children()[0].(*RichText)
	if rt.Spans[0].Variant != "" {
		t.Fatalf("scoped variant = %q, want plain", rt.Spans[0].Variant)
	}
}
