package ui

import (
	"testing"
	"time"
)

func TestComboBoxFilterOptions(t *testing.T) {
	sel := NewSignal("United States")
	c := NewComboBox("cb", []string{"United States", "United Kingdom", "Canada"}, sel, 0, 0, 0, 40)
	c.filter = "united"
	got := c.filteredOptions()
	if len(got) != 2 {
		t.Fatalf("filtered len = %d, want 2", len(got))
	}
}

func TestNormalizeDateRange(t *testing.T) {
	a := time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local)
	b := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)
	s, e := normalizeDateRange(a, b)
	if s.Day() != 1 || e.Day() != 20 {
		t.Fatalf("got %v – %v", s, e)
	}
}

func TestDateRangePickerFieldLabel(t *testing.T) {
	start := NewSignal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	end := NewSignal(time.Date(2026, 1, 31, 0, 0, 0, 0, time.Local))
	drp := NewDateRangePicker("r", start, end, 0, 0, 200, 40)
	if drp.formatFieldLabel() != "2026-01-01 – 2026-01-31" {
		t.Fatalf("label = %q", drp.formatFieldLabel())
	}
}
