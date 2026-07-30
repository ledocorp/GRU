package ui

import (
	"math/rand"
	"testing"
	"testing/quick"
)

// TestAppBarScrollIdleQuickWidth asserts AppBar+scroll shells idle across random widths.
func TestAppBarScrollIdleQuickWidth(t *testing.T) {
	f := func(width uint16) bool {
		w := float32(int(width)%580 + 220) // 220..799
		root := buildAppBarScrollFixture(w, 320)
		root.Layout()
		SimulateCacheHitFrame(root)
		return IdleReady(root)
	}
	cfg := &quick.Config{MaxCount: 40, Rand: rand.New(rand.NewSource(42))}
	if err := quick.Check(f, cfg); err != nil {
		t.Fatalf("AppBar+scroll not idle-ready for some width: %v", err)
	}
}
