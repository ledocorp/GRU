package ui

import "unicode/utf8"

// shapedCaretStops returns byte offsets where the caret may rest (cluster ends).
// Returns nil when shaped mode is unavailable for the line.
func shapedCaretStops(line string, fontSize float32, bold, italic, mono, preview bool) []int {
	if shapedNeedsSDFFallbackMeasure(line) {
		return nil
	}
	run, ok := shapedShapeText(line, fontSize, bold, italic, mono, preview)
	if !ok {
		return nil
	}
	stops := []int{0}
	prev := 0
	for _, g := range run.out.Glyphs {
		endRune := g.TextIndex() + g.RunesCount()
		bo := byteOffsetForRuneIndex(line, endRune)
		if bo > prev {
			stops = append(stops, bo)
			prev = bo
		}
	}
	if prev != len(line) {
		stops = append(stops, len(line))
	}
	return stops
}

func byteOffsetForRuneIndex(s string, runeIdx int) int {
	if runeIdx <= 0 {
		return 0
	}
	i := 0
	ri := 0
	for i < len(s) {
		_, size := utf8.DecodeRuneInString(s[i:])
		ri++
		i += size
		if ri == runeIdx {
			return i
		}
	}
	return len(s)
}

func shapedCaretStopsFromStyle(line string, s Style) []int {
	if !shapedTextReady() {
		return nil
	}
	fs := EffectiveFontSize(s)
	return shapedCaretStops(line, fs, styleDrawBold(s), s.Italic, s.Mono, s.PreviewFont)
}

// shapedCaretOffsetAtX maps a line-local X coordinate to a byte offset using cluster stops.
func shapedCaretOffsetAtX(line string, innerX float32, s Style) (int, bool) {
	stops := shapedCaretStopsFromStyle(line, s)
	if stops == nil {
		return 0, false
	}
	if innerX <= 0 {
		return 0, true
	}
	choice := stops[0]
	for i := 1; i < len(stops); i++ {
		w := EditorMeasureWidth(line[:stops[i]], s)
		if w < innerX {
			choice = stops[i]
			continue
		}
		if w == innerX {
			return stops[i], true
		}
		wPrev := EditorMeasureWidth(line[:choice], s)
		if innerX-wPrev < w-innerX {
			return choice, true
		}
		return stops[i], true
	}
	return stops[len(stops)-1], true
}
