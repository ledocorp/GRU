package ui

import "testing"

func TestSimpleSpellChecker(t *testing.T) {
	c := NewSimpleSpellChecker("xyzzy")
	if !c.Check("hello") {
		t.Fatal("expected hello in core dictionary")
	}
	if !c.Check("xyzzy") {
		t.Fatal("xyzzy should be known via Add")
	}
	if c.Check("helo") {
		t.Fatal("helo should be unknown")
	}
	if !c.Check("123") {
		t.Fatal("numeric token should skip as correct")
	}
	if !c.Check("a") {
		t.Fatal("single letter should skip as correct")
	}
}

func TestMisspelledRanges(t *testing.T) {
	c := NewSimpleSpellChecker()
	miss := MisspelledRanges("Hello teh world", c)
	if len(miss) != 1 {
		t.Fatalf("miss = %v, want [teh]", miss)
	}
	if miss[0][0] != 6 || miss[0][1] != 9 {
		t.Fatalf("range = %v", miss[0])
	}
}

func TestScanSpellWords(t *testing.T) {
	words := ScanSpellWords("don't test")
	if len(words) != 2 {
		t.Fatalf("got %d words", len(words))
	}
	if words[0].Word != "don't" {
		t.Fatalf("first = %q", words[0].Word)
	}
}
