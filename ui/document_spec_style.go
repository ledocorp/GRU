package ui

import "sort"

// DocumentSpecStyleFields lists JSON keys supported on block-level `"style"` objects.
// Same schema as Theme v2 styleJSON — see docs/DOCUMENT_SPEC_AUTHORING.md §Per-block style.
var DocumentSpecStyleFields = []string{
	"backgroundColor",
	"textColor",
	"fontSize",
	"padding",
	"borderWidth",
	"borderColor",
	"flexGrow",
	"cornerRadius",
	"bold",
	"fontDensity",
	"minFontSize",
}

// DocumentSpecPresetProps lists parametric `"props"` keys for visual presets.
var DocumentSpecPresetProps = []string{
	"glowIntensity",
	"hoverLift",
	"radiusScale",
}

// DocumentSpecSpanStyleKeys lists inline TextSpan styling keys in `.gru` JSON spans.
// Block-level `style` does not apply to spans — use variant/style on each run.
var DocumentSpecSpanStyleKeys = []string{
	"text",
	"style",
	"variant",
	"bold",
	"italic",
	"strike",
	"link",
	"linkTitle",
	"color",
}

// CollectDocumentSpecPresets returns unique preset names referenced in a spec tree.
func CollectDocumentSpecPresets(spec DocumentSpec) []string {
	seen := make(map[string]bool)
	for _, block := range spec.Children {
		collectDocBlockPresets(block, seen)
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func collectDocBlockPresets(block DocBlock, seen map[string]bool) {
	if block.Preset != "" {
		seen[block.Preset] = true
	}
	for _, item := range block.PresetItems {
		if item.Preset != "" {
			seen[item.Preset] = true
		}
	}
	for _, ch := range block.Children {
		collectDocBlockPresets(ch, seen)
	}
}

// ThemeV2ComponentVariants returns sorted variant names for a Theme v2 component.
func ThemeV2ComponentVariants(component string) []string {
	t := CurrentThemeV2()
	if t == nil {
		return nil
	}
	comp, ok := t.Components[component]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(comp.Variants))
	for name := range comp.Variants {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// ThemeV2ComponentNames returns sorted Theme v2 component keys.
func ThemeV2ComponentNames() []string {
	t := CurrentThemeV2()
	if t == nil {
		return nil
	}
	out := make([]string, 0, len(t.Components))
	for name := range t.Components {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
