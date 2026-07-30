// Package ui — shared flex-column text layout tests.
package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestApplyFlexAutoHeightLayoutWidthChange(t *testing.T) {
	l := NewLabel("t", "layout contract probe", 0, 0, 320, 0)
	l.Wrap = true
	var m flexTextMeasure

	// Stub heights — avoids depending on raylib font metrics in unit tests.
	heightAt := func(w float32) float32 {
		if w >= 200 {
			return 36
		}
		return 96
	}

	res := applyFlexAutoHeightLayout(l, &m, rl.NewRectangle(0, 0, 320, 0), heightAt)
	if !res.applied || res.height != 36 {
		t.Fatalf("wide layout = %+v, want height 36", res)
	}
	l.setBoundsNoMark(res.bounds)

	m.lastW = 0
	res = applyFlexAutoHeightLayout(l, &m, rl.NewRectangle(0, 0, 72, 36), heightAt)
	if !res.applied {
		t.Fatal("narrow layout not applied")
	}
	if res.height != 96 {
		t.Fatalf("narrow height = %.0f, want 96", res.height)
	}
	if !res.hostDirty {
		t.Fatal("width shrink should mark host dirty")
	}
}

func TestNewPlainTextWrapsByDefault(t *testing.T) {
	rt := NewPlainText("cap", "form-field-caption", "Display name", 0, 0, 120, 0)
	if !rt.Wrap {
		t.Fatal("Wrap = false")
	}
	if rt.Selectable {
		t.Fatal("Selectable = true")
	}
	if rt.styleName != "form-field-caption" {
		t.Fatalf("style = %q", rt.styleName)
	}
}
