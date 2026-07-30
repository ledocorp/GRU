// Package ui — spell-check helpers for TextEditor and other text surfaces.
//
// v1 uses [SimpleSpellChecker] (compact in-memory list) to prove the TextEditor hook,
// squiggles, and auto-correct. Production checking uses Hunspell via
// [TryHunspellChecker] after blank-importing ui/spell (en_US .aff/.dic);
// scenes implement [SpellChecker] only.
package ui

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SpellChecker reports whether a single word token is spelled correctly.
// Implementations should treat an empty return from Check as correct (skip).
type SpellChecker interface {
	Check(word string) bool
}

// SpellWord is a word token in source text (byte offsets into UTF-8 string).
type SpellWord struct {
	Start int
	End   int
	Word  string
}

// SimpleSpellChecker is a case-insensitive in-memory dictionary (English stub v1).
type SimpleSpellChecker struct {
	words map[string]struct{}
}

// NewSimpleSpellChecker builds a checker seeded with common English words plus extras.
func NewSimpleSpellChecker(extra ...string) *SimpleSpellChecker {
	c := &SimpleSpellChecker{words: make(map[string]struct{}, len(spellEnglishCore)+len(extra))}
	for _, w := range strings.Fields(spellEnglishCore) {
		c.words[normalizeSpellToken(w)] = struct{}{}
	}
	for _, w := range extra {
		w = normalizeSpellToken(w)
		if w != "" {
			c.words[w] = struct{}{}
		}
	}
	return c
}

// Add registers additional correctly-spelled tokens.
func (c *SimpleSpellChecker) Add(words ...string) {
	if c == nil {
		return
	}
	for _, w := range words {
		w = normalizeSpellToken(w)
		if w != "" {
			c.words[w] = struct{}{}
		}
	}
}

// Check reports whether word is in the dictionary. Unknown tokens with digits
// or no letters are treated as correct (skipped).
func (c *SimpleSpellChecker) Check(word string) bool {
	if c == nil {
		return true
	}
	w := normalizeSpellToken(word)
	if w == "" || len([]rune(w)) < 2 {
		return true
	}
	if spellWordSkippable(w) {
		return true
	}
	_, ok := c.words[w]
	return ok
}

// ScanSpellWords returns word tokens in text (letters with optional internal apostrophe).
func ScanSpellWords(text string) []SpellWord {
	var out []SpellWord
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		if isSpellWordRune(r) {
			start := i
			i += size
			for i < len(text) {
				r2, sz := utf8.DecodeRuneInString(text[i:])
				if !isSpellInnerRune(r2) {
					break
				}
				i += sz
			}
			word := text[start:i]
			out = append(out, SpellWord{Start: start, End: i, Word: word})
			continue
		}
		i += size
	}
	return out
}

// MisspelledRanges returns byte ranges [start,end) of words failing checker.
func MisspelledRanges(text string, checker SpellChecker) [][2]int {
	if checker == nil {
		return nil
	}
	var out [][2]int
	for _, tok := range ScanSpellWords(text) {
		if !checker.Check(tok.Word) {
			out = append(out, [2]int{tok.Start, tok.End})
		}
	}
	return out
}

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

func isSpellWordRune(r rune) bool {
	return unicode.IsLetter(r)
}

func isSpellInnerRune(r rune) bool {
	return unicode.IsLetter(r) || r == '\''
}

// spellEnglishCore is a compact v1 English stub (expand via SimpleSpellChecker.Add).
const spellEnglishCore = `
a about above after again all also am an and any are as at be because been before
being below between both but by can could did do does doing done down during each
either else enough even ever every for from further get got had has have having
he her here hers herself him his how however i if in into is it its itself just
know let like made make many may me might more most much must my myself no nor
not now of off on once only or other our ours ourselves out over own same see
she should so some such than that the their theirs them themselves then there
these they this those through to too under until up us use very was we were what
when where which while who whom why will with would you your yours yourself
able across add almost already although always another around away back become
been being best better big both bring came come could day did different does
done early end even every few find first found give go good great group hand
high however important include including keep kind large last late left life
little long look made make man may mean men might much must name need never new
next night number often old once only open own part people place point public
put read right same say see seem several show side since small so some something
still such take tell than that their them then there these they thing think
those though three through time to together too took turn two under until use
used using want way well went were what when where which while who why will
with within without work world would year years yet young
the and for are but not you all can had her was one our out day get has him his
how its may new now old see way who boy did has let put say she too use
hello world text editor notepad markdown preview spell check correct word
`
