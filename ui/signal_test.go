package ui

import "testing"

func TestSignalSetSkipsUnchanged(t *testing.T) {
	n := 0
	s := NewSignal(1)
	s.Subscribe(func() { n++ })
	s.Set(1)
	if n != 0 {
		t.Fatalf("unchanged Set notified %d times, want 0", n)
	}
	s.Set(2)
	if n != 1 {
		t.Fatalf("changed Set notified %d times, want 1", n)
	}
}

func TestSignalMirrorNoStackOverflow(t *testing.T) {
	a := NewSignal(0)
	b := NewSignal(0)
	a.Subscribe(func() { b.Set(a.Get()) })
	b.Subscribe(func() { a.Set(b.Get()) })
	a.Set(7) // must not recurse until stack overflow
	if a.Get() != 7 || b.Get() != 7 {
		t.Fatalf("mirror sync failed: a=%d b=%d", a.Get(), b.Get())
	}
}
