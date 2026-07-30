// Package ui (continued) — truncate helpers for fixed-width text.
package ui

// truncateTextS shortens text with an ellipsis when wider than maxW window pixels.
func truncateTextS(text string, maxW float32, s Style) string {
	if text == "" || maxW <= 0 {
		return text
	}
	if float32(measureTextS(text, s)) <= maxW {
		return text
	}
	const ellipsis = "…"
	ellW := float32(measureTextS(ellipsis, s))
	if ellW >= maxW {
		return ellipsis
	}
	runes := []rune(text)
	for len(runes) > 0 {
		candidate := string(runes)
		if float32(measureTextS(candidate, s))+ellW <= maxW {
			return candidate + ellipsis
		}
		runes = runes[:len(runes)-1]
	}
	return ellipsis
}
