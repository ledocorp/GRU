// Package ui (continued) — preset surface body typography sync.
package ui

func chromeStyleIsDark(st Style) bool {
	bg := st.BackgroundColor
	if bg.A < 32 {
		return false
	}
	lum := 0.299*float64(bg.R) + 0.587*float64(bg.G) + 0.114*float64(bg.B)
	return lum < 96
}

func bodyTypographyHintFromChrome(st Style) Style {
	hint := GetThemeStyle("surface-body")
	if st.TextColor.A != 0 {
		hint.TextColor = st.TextColor
	}
	if st.FontSize > 0 {
		hint.FontSize = st.FontSize
	}
	if st.FontDensity > 0 {
		hint.FontDensity = st.FontDensity
	}
	return hint
}

func surfaceTypographyHintActive(hint Style) bool {
	return hint.TextColor.A != 0 || hint.FontSize > 0
}

// surfaceUsesTintedBodyText reports whether body labels should inherit chrome
// typography (presets and non-default Theme v2 variants only).
func (e *Element) surfaceUsesTintedBodyText() bool {
	// Blockquote cards set body typography via richtext-blockquote (16px Inter).
	if e.styleVariant == "blockquote" || e.styleVariant == "blockquote-nested" {
		return false
	}
	if e.presetName != "" {
		return true
	}
	return e.styleVariant != "" && e.styleVariant != "default"
}

func applySurfaceBodyTypography(n Node, hint Style, darkChrome bool) {
	if !surfaceTypographyHintActive(hint) {
		return
	}
	switch w := n.(type) {
	case *Label:
		base := Style{}
		if w.styleOverrides != nil {
			base = *w.styleOverrides
		}
		w.SetStyleOverrides(mergeStyle(base, hint))
	case *RichText:
		// Markdown code fences keep richtext-code-block sizing (mono ~14px); do not
		// merge surface-body (18px) from tinted card chrome onto syntax blocks.
		if w.StyleName() == "richtext-code-block" {
			return
		}
		if darkChrome && w.StyleName() != "richtext-on-dark" {
			w.SetStyle("richtext-on-dark")
		}
		base := Style{}
		if w.styleOverrides != nil {
			base = *w.styleOverrides
		}
		merged := mergeStyle(base, hint)
		merged.Padding = 0
		w.SetStyleOverrides(merged)
		w.LineGap = 4
		w.MarkDirty()
	}
}
