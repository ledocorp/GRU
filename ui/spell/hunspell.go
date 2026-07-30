// Package spell registers a Hunspell/gospell SpellChecker backend into package ui.
//
//	import _ "github.com/ledocorp/gru/ui/spell"
package spell

import (
	"os"
	"strings"

	"github.com/client9/gospell"
	"github.com/ledocorp/gru/ui"
)

// HunspellChecker implements ui.SpellChecker using Hunspell .aff/.dic files via gospell.
type HunspellChecker struct {
	checker *gospell.Checker
}

// NewHunspellChecker loads dictionary files and optional extra allowed tokens (app jargon).
func NewHunspellChecker(affPath, dicPath string, extraWords ...string) (*HunspellChecker, error) {
	affPath, cleanup, err := normalizeAffForGoSpell(affPath)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	gs, err := gospell.NewGoSpell(affPath, dicPath)
	if err != nil {
		return nil, err
	}
	c := gospell.NewChecker(gs)
	if len(extraWords) > 0 {
		wl := &gospell.WordList{}
		for _, w := range extraWords {
			w = normalizeSpellToken(w)
			if w != "" {
				wl.Add(w)
			}
		}
		c.AddWordList(wl)
	}
	return &HunspellChecker{checker: c}, nil
}

// Check reports whether word is spelled correctly. Empty tokens are treated as correct.
func (h *HunspellChecker) Check(word string) bool {
	if h == nil || h.checker == nil {
		return true
	}
	w := normalizeSpellToken(word)
	if w == "" || len([]rune(w)) < 2 {
		return true
	}
	if spellWordSkippable(w) {
		return true
	}
	return h.checker.Spell(w)
}

func newSpellChecker(affPath, dicPath string, extraWords ...string) (ui.SpellChecker, error) {
	return NewHunspellChecker(affPath, dicPath, extraWords...)
}

// normalizeAffForGoSpell returns a path gospell can load. Chromium/LibreOffice
// dictionaries often use "SET UTF8"; client9/gospell requires "SET UTF-8".
func normalizeAffForGoSpell(affPath string) (path string, cleanup func(), err error) {
	raw, err := os.ReadFile(affPath)
	if err != nil {
		return "", func() {}, err
	}
	fixed := strings.Replace(string(raw), "SET UTF8", "SET UTF-8", 1)
	if fixed == string(raw) {
		return affPath, func() {}, nil
	}
	f, err := os.CreateTemp("", "gru-hunspell-*.aff")
	if err != nil {
		return "", func() {}, err
	}
	name := f.Name()
	if _, err := f.WriteString(fixed); err != nil {
		f.Close()
		os.Remove(name)
		return "", func() {}, err
	}
	if err := f.Close(); err != nil {
		os.Remove(name)
		return "", func() {}, err
	}
	return name, func() { os.Remove(name) }, nil
}
