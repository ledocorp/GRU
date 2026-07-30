package ui

import "testing"

func TestMenuPopupFlipsAboveWhenBelowHost(t *testing.T) {
	vp := NewViewport("vp", 0, 200, 400, 300)
	vp.styleName = "default"
	vp.contentHeight = 800
	vp.lastFlexValid = true
	vp.lastFlexW = 400
	vp.lastFlexH = 300
	vp.ScrollY = 500

	cb := NewComboBox("cb", []string{"One", "Two", "Three", "Four"}, NewSignal("One"), 40, 430, 160, 40)
	vp.AddChild(cb)
	cb.SetParent(vp)

	cb.isOpen = true
	cb.syncMenuPopupPlacement()
	if !cb.popupOpenAbove {
		t.Fatal("combobox near host bottom should flip popup above the face")
	}
	pop := cb.PopupBounds()
	if pop.Y+cb.PopupBounds().Height > cb.Bounds().Y+1 {
		t.Fatalf("flipped popup should sit above face: pop=%v face=%v", pop, cb.Bounds())
	}
	host := menuPopupHostRect(cb)
	if pop.Y < host.Y-1 {
		t.Fatalf("popup top %v above host top %v", pop.Y, host.Y)
	}
}

func TestDropdownPopupFlipsAboveNearBottom(t *testing.T) {
	vp := NewViewport("vp", 0, 100, 400, 250)
	vp.styleName = "default"

	dd := NewDropdown("dd", []string{"A", "B", "C", "D"}, 0, 40, 300, 160, 40)
	vp.AddChild(dd)
	dd.SetParent(vp)

	dd.isOpen = true
	dd.syncMenuPopupPlacement()
	if !dd.popupOpenAbove {
		t.Fatal("dropdown near viewport bottom should open upward")
	}
	if dd.PopupBounds().Y+dd.PopupBounds().Height > dd.Bounds().Y+1 {
		t.Fatal("popup should be above dropdown face")
	}
}

func TestComputeMenuPopupPlacementPrefersBelowWhenRoom(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 400, 600)
	vp.styleName = "default"
	dd := NewDropdown("dd", []string{"A", "B", "C"}, 0, 40, 80, 160, 40)
	vp.AddChild(dd)
	dd.SetParent(vp)

	p := computeMenuPopupPlacement(dd, dd.Bounds(), 0, 120, 40)
	if p.openAbove {
		t.Fatal("should open below when plenty of room under face in viewport")
	}
}
