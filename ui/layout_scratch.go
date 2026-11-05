package ui

import (
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// gridLayoutScratch reuses flex/grid row buffers to cut resize alloc churn (Phase 5).
type gridLayoutScratch struct {
	rowPending  []gridPendingItem
	gridRows    []gridRowScratch
	rowChildren []Node
}

type gridPendingItem struct {
	child  Node
	bounds rl.Rectangle
}

type gridRowScratch struct {
	children []Node
	topY     float32
	height   float32
}

var gridLayoutScratchPool = sync.Pool{
	New: func() any { return &gridLayoutScratch{} },
}

func acquireGridLayoutScratch() *gridLayoutScratch {
	return gridLayoutScratchPool.Get().(*gridLayoutScratch)
}

func releaseGridLayoutScratch(s *gridLayoutScratch) {
	s.rowPending = s.rowPending[:0]
	s.gridRows = s.gridRows[:0]
	s.rowChildren = s.rowChildren[:0]
	gridLayoutScratchPool.Put(s)
}
