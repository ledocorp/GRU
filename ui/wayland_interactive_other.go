//go:build !linux

package ui

func platformInteractiveMoveResizeSupported() bool { return false }

func platformHealPointerButtons() bool { return false }

func platformBeginInteractiveMove() bool { return false }

func platformBeginInteractiveResize(edge resizeEdge) bool {
	_ = edge
	return false
}
