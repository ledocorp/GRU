// Package ui (continued) — shared surface features facade helpers (Phase C3).
package ui

// ensureSurfaceFeatures attaches the default feature controller when missing.
func ensureSurfaceFeatures(sh *SurfaceShell) *PanelFeaturesBehavior {
	if sh.panelFeatures == nil {
		sh.AttachBehavior(NewPanelFeaturesBehavior())
	}
	return sh.panelFeatures
}

func surfaceFeaturesConfig(sh *SurfaceShell) *PanelFeatures {
	return ensureSurfaceFeatures(sh).config
}

func surfaceSyncFeatures(sh *SurfaceShell) {
	if sh.panelFeatures != nil {
		sh.panelFeatures.Apply()
	}
}

func surfaceContentTarget(sh *SurfaceShell) Node {
	if pf := sh.panelFeatures; pf != nil {
		return pf.panelContentTarget(sh)
	}
	if sh.body != nil {
		return sh.body
	}
	return sh
}

func surfaceCollapseBehavior(sh *SurfaceShell) *CollapseBehavior {
	if sh.panelFeatures == nil {
		return sh.collapse
	}
	return sh.panelFeatures.collapse
}

func surfaceDismissBehavior(sh *SurfaceShell) *DismissBehavior {
	return sh.dismiss
}
