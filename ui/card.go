// Package ui (continued)
// See node.go for the full package documentation.
package ui

// Card is a facade over SurfaceShell with an inset title header (HeaderModeInset).
//
// Example:
//
//	c := ui.NewCard("info-card", "Quick Stats", 0, 0, 560, 180)
//	c.AddChild(ui.NewLabel("stat1", "Users: 4 218", 0, 0, 480, 28))
//	viewport.AddChild(c)
type Card struct {
	SurfaceShell
}

// NewCard creates a Card with the given title and bounds.
func NewCard(id, title string, x, y, w, h float32) *Card {
	c := NewContainer(id, x, y, w, h)
	card := &Card{
		SurfaceShell: SurfaceShell{
			Container:   *c,
			Title:       title,
			TitleHeight: 40,
			headerMode:  HeaderModeInset,
		},
	}
	card.Gap = 12
	card.cachePolicy = CacheAuto
	card.attachBody(id)
	card.styleName = "card"
	card.Element.SetStyleVariant("card", "default")
	card.AttachBehavior(NewPanelFeaturesBehavior())
	return card
}

// AddChild attaches content to the body or scroll host; parent is the attach target
// so ancestor clip walks (viewport scissor) stay correct.
func (card *Card) AddChild(child Node) {
	target := surfaceContentTarget(&card.SurfaceShell)
	if vp, ok := target.(*Viewport); ok {
		vp.AddChild(child)
	} else if card.body != nil {
		card.body.AddChild(child)
	}
	child.SetParent(target)
	card.applySurfaceBodyTypographyToChild(child)
}

// ApplyCardBodyTextColor syncs direct Label/RichText children to card chrome.
func (card *Card) ApplyCardBodyTextColor() {
	card.applySurfaceBodyTypographyToChildren()
}

// SetStylePreset applies a named visual preset to the card chrome.
func (card *Card) SetStylePreset(name string, props PresetProps) error {
	if err := card.Element.SetStylePreset(name, props); err != nil {
		return err
	}
	applyVisualSurfaceLayout(&card.TitleHeight, &card.Gap, name)
	card.ApplyCardBodyTextColor()
	return nil
}
