package ui

import (
	"testing"
)

func TestParseRemixIconCSS(t *testing.T) {
	classes, err := parseRemixIconCSS("assets/fonts/remixicon.css")
	if err != nil {
		t.Skip("remixicon.css not in workspace:", err)
	}
	if len(classes) < 1000 {
		t.Fatalf("expected large icon set, got %d classes", len(classes))
	}
	if cp, ok := classes["close-line"]; !ok || cp != 0xEB99 {
		t.Fatalf("close-line = %v ok=%v", cp, ok)
	}
}

func TestRemixCodepointForPhosphorNames(t *testing.T) {
	classes, err := parseRemixIconCSS("assets/fonts/remixicon.css")
	if err != nil {
		t.Skip(err)
	}
	cases := []struct {
		name   string
		weight PhosphorWeight
	}{
		{PhosphorHouse, PhosphorRegular},
		{PhosphorMagnifyingGlass, PhosphorRegular},
		{PhosphorBell, PhosphorRegular},
		{PhosphorStar, PhosphorFill},
		{PhosphorMinus, PhosphorRegular},
		{PhosphorSquare, PhosphorRegular},
		{PhosphorX, PhosphorRegular},
		{PhosphorCaretLeft, PhosphorRegular},
		{PhosphorDotsThreeVertical, PhosphorRegular},
		{PhosphorResize, PhosphorRegular},
	}
	for _, tc := range cases {
		if _, ok := remixCodepointFor(tc.name, tc.weight, classes); !ok {
			t.Errorf("missing remix mapping for %q weight %q", tc.name, tc.weight)
		}
	}
}
