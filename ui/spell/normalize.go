package spell

import (
	"strings"
	"unicode"
)

func normalizeSpellToken(word string) string {
	return strings.ToLower(strings.Trim(word, "'\""))
}

func spellWordSkippable(w string) bool {
	hasLetter := false
	for _, r := range w {
		if unicode.IsLetter(r) {
			hasLetter = true
		} else if unicode.IsDigit(r) {
			return true
		} else if r != '\'' && r != '-' {
			return true
		}
	}
	return !hasLetter
}
