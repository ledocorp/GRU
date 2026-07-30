// Package ui (continued) — shared layout metrics for visual preset surfaces.
package ui

// Visual preset surfaces (card + panel) share one rhythm so gallery/showcase rows align.
const (
	PresetSurfacePadding      = float32(16)
	PresetSurfaceTitleHeight  = float32(36)
	PresetSurfaceCornerRadius = float32(10)
	PresetSurfaceBorderWidth  = float32(2)
	PresetSurfaceFontSize     = int32(18)
	PresetSurfaceBodyGap      = float32(8)
	// PresetBackdropPadding insets preset tiles from the demo gradient strip edges.
	PresetBackdropPadding = float32(12)
)

// IsVisualSurfacePreset reports presets that use PresetSurface* layout metrics.
func IsVisualSurfacePreset(name string) bool {
	switch name {
	case "neo-glow-card", "glass-panel", "glass-panel-dark", "glass-card", "surface-card", "callout-tip", "code-block":
		return true
	default:
		return false
	}
}

// normalizeVisualSurfaceStyle applies shared padding, typography, radius, and border width.
func normalizeVisualSurfaceStyle(st Style) Style {
	st.Padding = PresetSurfacePadding
	st.FontSize = PresetSurfaceFontSize
	if st.BorderWidth > 0 || st.BorderColor.A > 0 {
		st.BorderWidth = PresetSurfaceBorderWidth
	}
	if st.CornerRadius > 0 {
		st.CornerRadius = PresetSurfaceCornerRadius
	}
	return st
}

func applyVisualSurfaceLayout(titleHeight, gap *float32, presetName string) {
	if !IsVisualSurfacePreset(presetName) {
		return
	}
	*titleHeight = PresetSurfaceTitleHeight
	*gap = PresetSurfaceBodyGap
}
