// Resize propagation debugger for the page shell → viewport → grid chain.
//
// Go excludes files named *_test.go from normal builds (go run / go build),
// so this lives as debug_resize.go. Press F8 in the demo to print bounds.
package main

import (
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/ui"
	"log"
)

func logBounds(tag string, n ui.Node) {
	if n == nil {
		log.Printf("[resize-debug] %s: <nil>", tag)
		return
	}
	r := n.Bounds()
	log.Printf("[resize-debug] %s id=%s X=%.1f Y=%.1f W=%.1f H=%.1f", tag, n.ID(), r.X, r.Y, r.Width, r.Height)
}

// TestResizePropagation builds an isolated Document (shell → main Viewport →
// 12-column grid → sample panel), logs bounds, simulates resizes via doc.Resize,
// and logs again. Does not modify the running demo document.
func TestResizePropagation() {
	const w0, h0 int32 = 1280, 720
	doc := ui.NewDocument(w0, h0)
	shell, vp := examples.MountFlexPageShell(doc, "rzdbg")
	vp.SetStyle("transparent")

	grid := ui.NewContainer("rzdbg-grid", 0, 0, 0, 0)
	grid.LayoutType = ui.LayoutGrid
	grid.GridColumns = 12
	grid.Gap = 12
	grid.SetStyle("page-grid")

	sample := ui.NewPanel("rzdbg-panel", "Sample", 0, 0, 0, 120)
	sample.SetColSpan(ui.BreakpointXS, 12)
	sample.SetColSpan(ui.BreakpointMD, 6)
	sample.SetColSpan(ui.BreakpointLG, 4)

	grid.AddChild(sample)
	vp.AddChild(grid)

	doc.Root.MarkDirty()
	doc.Root.Layout()

	log.Println("[resize-debug] === BEFORE (initial layout, 1280×720) ===")
	logBounds("shell", shell)
	logBounds("main viewport", vp)
	logBounds("grid", grid)
	logBounds("sample panel", sample)

	doc.Resize(640, h0)
	log.Println("[resize-debug] === AFTER doc.Resize(640, 720) ===")
	logBounds("shell", shell)
	logBounds("main viewport", vp)
	logBounds("grid", grid)
	logBounds("sample panel", sample)

	doc.Resize(1600, h0)
	log.Println("[resize-debug] === AFTER doc.Resize(1600, 720) ===")
	logBounds("shell", shell)
	logBounds("main viewport", vp)
	logBounds("grid", grid)
	logBounds("sample panel", sample)

	doc.Resize(w0, h0)
	log.Println("[resize-debug] === AFTER doc.Resize(restore 1280×720) ===")
	logBounds("shell", shell)
	logBounds("main viewport", vp)
	logBounds("grid", grid)
	logBounds("sample panel", sample)
	log.Println("[resize-debug] === done (isolated doc discarded) ===")
}
