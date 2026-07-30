// Package ui (continued)
package ui

// SetIcon draws from the shared Remix icon registry (Icons).
func (ib *IconButton) SetIcon(name string, weight IconWeight) {
	ib.Symbol = ""
	ib.IconPath = ""
	ib.VectorPainter = nil
	ib.phosphorName = name
	ib.phosphorWeight = weight
	ib.usePhosphor = name != ""
	ib.MarkDrawDirty()
}

// Deprecated: use SetIcon.
func (ib *IconButton) SetPhosphorIcon(name string, weight IconWeight) {
	ib.SetIcon(name, weight)
}

// phosphorIconReady reports whether the Remix glyph is in the atlas.
func phosphorIconReady(name string, weight IconWeight) bool {
	if name == "" {
		return false
	}
	return remixHasGlyph(name, weight)
}
