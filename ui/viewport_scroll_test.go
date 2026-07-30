package ui

import (
	"testing"
)

func TestScrollToShowNodeIgnoresCurrentScrollY(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 200, 100)
	lane := NewContainer("lane", 0, 0, 200, 0)
	lane.AutoHeight = true
	lane.LayoutType = LayoutFlex
	lane.FlexDirection = FlexColumn
	lane.Gap = 0
	for i := 0; i < 12; i++ {
		lane.AddChild(NewLabel("sp"+itoa(i), "line", 0, 0, 200, 24))
	}
	target := NewLabel("target", "heading", 0, 0, 200, 24)
	lane.AddChild(target)
	vp.AddChild(lane)

	vp.Layout()
	docY := vp.measureContentOffsetY(target, lane)
	vp.ScrollY = 180
	vp.scrollDirty = true
	vp.Layout()

	vp.ScrollToShowNode(target, lane)
	want := docY - 12
	if max := vp.overflowScrollY(); want > max {
		want = max
	}
	if vp.ScrollY < want-1 || vp.ScrollY > want+1 {
		t.Fatalf("ScrollY = %v, want ~%v (docY=%v)", vp.ScrollY, want, docY)
	}
}

func TestNodeOffsetYWithinNested(t *testing.T) {
	lane := NewContainer("lane", 0, 0, 200, 0)
	lane.AutoHeight = true
	lane.LayoutType = LayoutFlex
	lane.FlexDirection = FlexColumn
	lane.Gap = 0
	for i := 0; i < 5; i++ {
		lane.AddChild(NewLabel("sp"+itoa(i), "line", 0, 0, 200, 20))
	}
	wrap := NewContainer("wrap", 0, 0, 200, 0)
	wrap.AutoHeight = true
	wrap.LayoutType = LayoutFlex
	wrap.FlexDirection = FlexColumn
	target := NewLabel("target", "fn", 0, 0, 200, 20)
	wrap.AddChild(target)
	lane.AddChild(wrap)

	vp := NewViewport("vp", 0, 0, 200, 100)
	vp.AddChild(lane)
	vp.Layout()

	got := NodeOffsetYWithin(target, lane)
	if got < 100 || got > 130 {
		t.Fatalf("NodeOffsetYWithin nested = %v, want ~100-130", got)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [12]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}
