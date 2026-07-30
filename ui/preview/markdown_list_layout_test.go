package preview

import (
	"strings"
	"testing"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestMarkdownListItemsNoVerticalOverlap(t *testing.T) {
	src := "- " + strings.Repeat("first item ", 35) + "\n- " + strings.Repeat("second item ", 35)
	lane := buildPreviewLane(t, src)
	col, ok := lane.Children()[0].(*ui.Container)
	if !ok {
		t.Fatalf("list root = %T, want Container", lane.Children()[0])
	}
	col.SetBounds(rl.NewRectangle(0, 0, 320, 900))
	col.Layout()

	if len(col.Children()) < 2 {
		t.Fatalf("list rows = %d, want >= 2", len(col.Children()))
	}
	r0 := col.Children()[0].Bounds()
	r1 := col.Children()[1].Bounds()
	if r1.Y < r0.Y+r0.Height-0.5 {
		t.Fatalf("list rows overlap: row0 bottom=%.1f row1 top=%.1f", r0.Y+r0.Height, r1.Y)
	}
}

func TestMarkdownHeadingLevelChangeRelayouts(t *testing.T) {
	w := float32(360)
	layoutHeading := func(level int) *ui.RichText {
		src := strings.Repeat("#", level) + " Title " + strings.Repeat("wrap ", 20)
		nodes := BuildMarkdownNodes("h", src)
		if len(nodes) != 1 {
			t.Fatalf("blocks = %d, want 1 heading", len(nodes))
		}
		rt, ok := nodes[0].(*ui.RichText)
		if !ok {
			t.Fatalf("heading = %T, want RichText", nodes[0])
		}
		lane := ui.NewContainer("lane", 0, 0, w, 0)
		lane.LayoutType = ui.LayoutFlex
		lane.FlexDirection = ui.FlexColumn
		lane.AutoHeight = true
		lane.AddChild(rt)
		lane.SetBounds(rl.NewRectangle(0, 0, w, 800))
		lane.Layout()
		return rt
	}
	h1 := layoutHeading(1)
	h2 := layoutHeading(2)
	if h1.Bounds().Height <= h2.Bounds().Height {
		t.Fatalf("h1 height %.1f should exceed h2 %.1f", h1.Bounds().Height, h2.Bounds().Height)
	}
}
