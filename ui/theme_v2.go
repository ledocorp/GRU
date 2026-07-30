package ui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// StyleState describes transient visual state used by Theme v2 resolution.
// It is intentionally draw-focused; layout-affecting state styles should be
// introduced only after a widget explicitly opts into that contract.
type StyleState uint32

const (
	StyleStateNone StyleState = 0
	StyleStateHover StyleState = 1 << iota
	StyleStatePressed
	StyleStateFocus
	StyleStateOpen
	StyleStateChecked
	StyleStateDisabled
)

// TokenSet is the optional design-token source for Theme v2. Tokens are not
// interpreted in hot draw paths; they are for JSON/theme authoring and future
// tooling. Resolved widgets still receive concrete Style values.
type TokenSet struct {
	Colors     map[string]rl.Color
	Spacing    map[string]float32
	Typography map[string]TypographyToken
	Radii      map[string]float32
}

// TypographyToken captures semantic text settings for future JSON themes.
type TypographyToken struct {
	Size    int32
	Density float32
	Bold    bool
	MinSize int32
}

// ComponentStyle is a predictable, non-cascading style group.
type ComponentStyle struct {
	Base     Style
	Variants map[string]Style
	States   map[string]Style
}

// ThemeV2 is the token/component theme layer that sits above the legacy flat
// CurrentTheme map. The legacy map remains the compatibility fallback.
type ThemeV2 struct {
	Name       string
	Tokens     TokenSet
	Components map[string]ComponentStyle
	Revision   uint64
}

var (
	currentThemeV2  *ThemeV2
	themeRevisionV2 uint64
)

func init() {
	SetThemeV2(NewDefaultThemeV2())
}

// NewDefaultThemeV2 builds a small component/variant layer from the existing
// polished flat theme. This preserves the current visual baseline while allowing
// new code to opt into component + variant naming.
func NewDefaultThemeV2() *ThemeV2 {
	return &ThemeV2{
		Name:       "Gru Default",
		Tokens:     defaultTokenSet(),
		Components: map[string]ComponentStyle{
			"button": {
				Base: CurrentTheme["button"],
				Variants: map[string]Style{
					"default": CurrentTheme["button"],
					"primary": CurrentTheme["primary"],
					"danger":  CurrentTheme["danger"],
					"ghost": {
						BackgroundColor: rl.NewColor(0, 0, 0, 0),
						TextColor:       rl.NewColor(183, 189, 201, 255),
						BorderColor:     rl.NewColor(220, 224, 232, 255),
						BorderWidth:     1.5,
					},
				},
			},
			"icon-button": {
				Base: CurrentTheme["icon-button"],
				Variants: map[string]Style{
					"default": CurrentTheme["icon-button"],
					"primary": CurrentTheme["primary"],
					"danger":  CurrentTheme["danger"],
				},
			},
			"panel": {
				Base: CurrentTheme["panel"],
				Variants: map[string]Style{
					"default": CurrentTheme["panel"],
					"glass": {
						BackgroundColor: rl.NewColor(238, 242, 255, 125),
						TextColor:       rl.NewColor(30, 27, 75, 255),
						FontSize:        18,
						Padding:         CurrentTheme["panel"].Padding,
						BorderWidth:     2,
						BorderColor:     rl.NewColor(165, 180, 252, 240),
						CornerRadius:    10,
					},
					"glass-dark": {
						BackgroundColor: rl.NewColor(15, 18, 42, 170),
						TextColor:       rl.NewColor(224, 231, 255, 255),
						FontSize:        18,
						Padding:         CurrentTheme["panel"].Padding,
						BorderWidth:     2,
						BorderColor:     rl.NewColor(99, 102, 241, 210),
						CornerRadius:    10,
					},
				},
			},
			"card": {
				Base: CurrentTheme["card"],
				Variants: map[string]Style{
					"default": CurrentTheme["card"],
					"callout":           docMutedCardStyle(),
					"code":              docCodeCardStyle(),
					"table":             docMutedCardStyle(),
					"blockquote":        docBlockquoteCardStyle(),
					"blockquote-nested": docBlockquoteNestedCardStyle(),
					"image":             docImageCardStyle(),
					"neo-glow": {
						BackgroundColor: rl.NewColor(30, 27, 75, 255),
						TextColor:       rl.NewColor(224, 231, 255, 255),
						FontSize:        18,
						Padding:         16,
						BorderWidth:     2,
						BorderColor:     rl.NewColor(99, 102, 241, 255),
						CornerRadius:    10,
					},
					"glass": {
						BackgroundColor: rl.NewColor(255, 255, 255, 72),
						TextColor:       rl.NewColor(30, 27, 75, 255),
						FontSize:        18,
						Padding:         16,
						BorderWidth:     2,
						BorderColor:     rl.NewColor(165, 180, 252, 230),
						CornerRadius:    10,
					},
				},
			},
			"input": {
				Base: CurrentTheme["input"],
				Variants: map[string]Style{
					"default": CurrentTheme["input"],
				},
				States: map[string]Style{
					"hover": {
						BorderColor: rl.NewColor(129, 140, 248, 255),
					},
					"focus": {
						BorderColor: rl.NewColor(79, 70, 229, 255),
					},
					"disabled": {
						BackgroundColor: rl.NewColor(241, 242, 247, 255),
						TextColor:       rl.NewColor(145, 148, 165, 255),
						BorderColor:     rl.NewColor(210, 212, 222, 255),
					},
				},
			},
			"dropdown": {
				Base: CurrentTheme["dropdown"],
				Variants: map[string]Style{
					"default": CurrentTheme["dropdown"],
				},
				States: map[string]Style{
					"hover": {
						BackgroundColor: rl.NewColor(247, 248, 255, 255),
						BorderColor:     rl.NewColor(129, 140, 248, 255),
					},
					"open": {
						BackgroundColor: rl.NewColor(238, 242, 255, 255),
						BorderColor:     rl.NewColor(79, 70, 229, 255),
					},
					"disabled": {
						BackgroundColor: rl.NewColor(241, 242, 247, 255),
						TextColor:       rl.NewColor(145, 148, 165, 255),
						BorderColor:     rl.NewColor(210, 212, 222, 255),
					},
				},
			},
			"checkbox": {
				Base: CurrentTheme["checkbox"],
				Variants: map[string]Style{
					"default": CurrentTheme["checkbox"],
				},
				States: map[string]Style{
					"checked":  CurrentTheme["checkbox-checked"],
					"disabled": disabledControlStyle(),
				},
			},
			"toggle": {
				Base: CurrentTheme["toggle"],
				Variants: map[string]Style{
					"default": CurrentTheme["toggle"],
				},
				States: map[string]Style{
					"checked":  CurrentTheme["toggle-on"],
					"disabled": disabledControlStyle(),
				},
			},
			"radio": {
				Base: CurrentTheme["radio"],
				Variants: map[string]Style{
					"default": CurrentTheme["radio"],
				},
				States: map[string]Style{
					"hover":    CurrentTheme["radio-hover"],
					"checked":  CurrentTheme["radio-selected"],
					"disabled": CurrentTheme["radio-disabled"],
				},
			},
			"slider": {
				Base: CurrentTheme["default"],
				Variants: map[string]Style{
					"default": CurrentTheme["default"],
				},
				States: map[string]Style{
					"hover": {
						BorderColor: rl.NewColor(99, 90, 249, 255),
					},
					"pressed": {
						BorderColor: rl.NewColor(79, 70, 229, 255),
					},
					"disabled": disabledControlStyle(),
				},
			},
			"badge": {
				Base: CurrentTheme["badge"],
				Variants: map[string]Style{
					"default": CurrentTheme["badge"],
					"primary": CurrentTheme["badge-primary"],
					"success": CurrentTheme["badge-success"],
					"warning": CurrentTheme["badge-warning"],
					"danger":  CurrentTheme["badge-danger"],
					"info":    CurrentTheme["badge-info"],
				},
			},
		},
	}
}

// SetThemeV2 swaps the active Theme v2 layer and bumps the global revision so
// element-level resolved-style caches naturally invalidate.
func SetThemeV2(theme *ThemeV2) {
	currentThemeV2 = theme
	themeRevisionV2++
	if currentThemeV2 != nil {
		currentThemeV2.Revision = themeRevisionV2
	}
}

// CurrentThemeV2 returns the active component/variant theme layer.
func CurrentThemeV2() *ThemeV2 { return currentThemeV2 }

// DocMarkdownCardPadding is uniform inset for preview code/table/callout cards.
const DocMarkdownCardPadding = float32(10)

// docMutedCardStyle is the shared light-grey chrome for markdown callout, code, and table blocks.
func docMutedCardStyle() Style {
	return Style{
		BackgroundColor: rl.NewColor(248, 249, 251, 255),
		TextColor:       rl.NewColor(40, 42, 54, 255),
		Padding:         DocMarkdownCardPadding,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 214, 222, 255),
		CornerRadius:    8,
	}
}

func docCodeCardStyle() Style {
	return docMutedCardStyle()
}

func darkDocMutedCardStyle() Style {
	return Style{
		BackgroundColor: rl.NewColor(36, 38, 50, 255),
		TextColor:       rl.NewColor(232, 234, 244, 255),
		Padding:         DocMarkdownCardPadding,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(64, 68, 84, 255),
		CornerRadius:    8,
	}
}

// SyncThemeV2MarkdownCards updates card code/table/callout variants for light or dark appearance.
func SyncThemeV2MarkdownCards(dark bool) {
	if currentThemeV2 == nil {
		return
	}
	comp, ok := currentThemeV2.Components["card"]
	if !ok {
		return
	}
	if dark {
		muted := darkDocMutedCardStyle()
		comp.Variants["callout"] = muted
		comp.Variants["code"] = muted
		comp.Variants["table"] = muted
		comp.Variants["blockquote"] = muted
		nested := muted
		nested.Padding = 10
		comp.Variants["blockquote-nested"] = nested
	} else {
		comp.Variants["callout"] = docMutedCardStyle()
		comp.Variants["code"] = docCodeCardStyle()
		comp.Variants["table"] = docMutedCardStyle()
		comp.Variants["blockquote"] = docBlockquoteCardStyle()
		comp.Variants["blockquote-nested"] = docBlockquoteNestedCardStyle()
	}
	currentThemeV2.Components["card"] = comp
}

// SyncThemeV2FormControls copies CurrentTheme field/modal bases into Theme v2 after SetAppearance.
func SyncThemeV2FormControls() {
	if currentThemeV2 == nil {
		return
	}
	for _, name := range []string{"input", "dropdown", "combobox", "modal", "searchbar"} {
		s, ok := CurrentTheme[name]
		if !ok {
			continue
		}
		comp, ok := currentThemeV2.Components[name]
		if !ok {
			continue
		}
		comp.Base = s
		if comp.Variants == nil {
			comp.Variants = map[string]Style{}
		}
		comp.Variants["default"] = s
		currentThemeV2.Components[name] = comp
	}
}

func docBlockquoteCardStyle() Style {
	return docMutedCardStyle()
}

func docBlockquoteNestedCardStyle() Style {
	st := docMutedCardStyle()
	st.Padding = 10
	return st
}

func docImageCardStyle() Style {
	return Style{
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(40, 42, 54, 255),
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(220, 222, 230, 255),
		CornerRadius:    6,
	}
}

// MergedComponentVariantStyle returns base+variant merged Style for a Theme v2 component.
// Used by Card callout/code recipes and Draw fallback when variant chrome must be visible.
func MergedComponentVariantStyle(component, variant string) (Style, bool) {
	if currentThemeV2 == nil || component == "" || variant == "" || variant == "default" {
		return Style{}, false
	}
	comp, ok := currentThemeV2.Components[component]
	if !ok {
		return Style{}, false
	}
	vs, ok := comp.Variants[variant]
	if !ok {
		return Style{}, false
	}
	return mergeStyle(comp.Base, vs), true
}

type themeV2JSON struct {
	Name       string                         `json:"name"`
	Tokens     tokenSetJSON                   `json:"tokens"`
	Components map[string]componentStyleJSON `json:"components"`
}

type tokenSetJSON struct {
	Colors     map[string]string          `json:"colors"`
	Spacing    map[string]float32         `json:"spacing"`
	Typography map[string]TypographyToken `json:"typography"`
	Radii      map[string]float32         `json:"radii"`
}

type componentStyleJSON struct {
	Base     styleJSON            `json:"base"`
	Variants map[string]styleJSON `json:"variants"`
	States   map[string]styleJSON `json:"states"`
}

type styleJSON struct {
	BackgroundColor *string  `json:"backgroundColor"`
	TextColor       *string  `json:"textColor"`
	FontSize        *int32   `json:"fontSize"`
	Padding         *float32 `json:"padding"`
	BorderWidth     *float32 `json:"borderWidth"`
	BorderColor     *string  `json:"borderColor"`
	FlexGrow        *float32 `json:"flexGrow"`
	CornerRadius    *float32 `json:"cornerRadius"`
	Bold            *bool    `json:"bold"`
	FontDensity     *float32 `json:"fontDensity"`
	MinFontSize     *int32   `json:"minFontSize"`
}

// LoadThemeV2 loads a component/variant theme from JSON and makes it active.
// It is separate from LoadTheme, which remains the legacy flat map loader.
func LoadThemeV2(data []byte) error {
	var spec themeV2JSON
	if err := json.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("ui/theme_v2: parse error: %w", err)
	}
	theme, err := spec.toTheme()
	if err != nil {
		return err
	}
	SetThemeV2(theme)
	return nil
}

func (j themeV2JSON) toTheme() (*ThemeV2, error) {
	tokens, err := j.Tokens.toTokens()
	if err != nil {
		return nil, err
	}
	components := make(map[string]ComponentStyle, len(j.Components))
	for name, compJSON := range j.Components {
		comp, err := compJSON.toComponent()
		if err != nil {
			return nil, fmt.Errorf("ui/theme_v2: component %q: %w", name, err)
		}
		components[name] = comp
	}
	return &ThemeV2{
		Name:       j.Name,
		Tokens:     tokens,
		Components: components,
	}, nil
}

func (j tokenSetJSON) toTokens() (TokenSet, error) {
	colors := make(map[string]rl.Color, len(j.Colors))
	for name, raw := range j.Colors {
		col, err := parseHexColor(raw)
		if err != nil {
			return TokenSet{}, fmt.Errorf("ui/theme_v2: token color %q: %w", name, err)
		}
		colors[name] = col
	}
	return TokenSet{
		Colors:     colors,
		Spacing:    j.Spacing,
		Typography: j.Typography,
		Radii:      j.Radii,
	}, nil
}

func (j componentStyleJSON) toComponent() (ComponentStyle, error) {
	base, err := j.Base.toStyle(DefaultStyle)
	if err != nil {
		return ComponentStyle{}, fmt.Errorf("base: %w", err)
	}
	variants := make(map[string]Style, len(j.Variants))
	for name, raw := range j.Variants {
		st, err := raw.toStyle(Style{})
		if err != nil {
			return ComponentStyle{}, fmt.Errorf("variant %q: %w", name, err)
		}
		variants[name] = st
	}
	states := make(map[string]Style, len(j.States))
	for name, raw := range j.States {
		st, err := raw.toStyle(Style{})
		if err != nil {
			return ComponentStyle{}, fmt.Errorf("state %q: %w", name, err)
		}
		states[name] = st
	}
	return ComponentStyle{Base: base, Variants: variants, States: states}, nil
}

func (j styleJSON) toStyle(base Style) (Style, error) {
	if j.BackgroundColor != nil {
		col, err := resolveStyleColor(*j.BackgroundColor)
		if err != nil {
			return Style{}, fmt.Errorf("backgroundColor: %w", err)
		}
		base.BackgroundColor = col
	}
	if j.TextColor != nil {
		col, err := resolveStyleColor(*j.TextColor)
		if err != nil {
			return Style{}, fmt.Errorf("textColor: %w", err)
		}
		base.TextColor = col
	}
	if j.FontSize != nil {
		base.FontSize = *j.FontSize
	}
	if j.Padding != nil {
		base.Padding = *j.Padding
	}
	if j.BorderWidth != nil {
		base.BorderWidth = *j.BorderWidth
	}
	if j.BorderColor != nil {
		col, err := resolveStyleColor(*j.BorderColor)
		if err != nil {
			return Style{}, fmt.Errorf("borderColor: %w", err)
		}
		base.BorderColor = col
	}
	if j.FlexGrow != nil {
		base.FlexGrow = *j.FlexGrow
	}
	if j.CornerRadius != nil {
		base.CornerRadius = *j.CornerRadius
	}
	if j.Bold != nil {
		base.Bold = *j.Bold
	}
	if j.FontDensity != nil {
		base.FontDensity = *j.FontDensity
	}
	if j.MinFontSize != nil {
		base.MinFontSize = *j.MinFontSize
	}
	return base, nil
}

func parseHexColor(raw string) (rl.Color, error) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 && len(s) != 8 {
		return rl.Color{}, fmt.Errorf("expected #RRGGBB or #RRGGBBAA, got %q", raw)
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return rl.Color{}, err
	}
	if len(s) == 6 {
		return rl.NewColor(uint8(v>>16), uint8(v>>8), uint8(v), 255), nil
	}
	return rl.NewColor(uint8(v>>24), uint8(v>>16), uint8(v>>8), uint8(v)), nil
}

// mergeStyleJSON merges DocBlock style patches; nil pointer fields in add are ignored.
func mergeStyleJSON(base *styleJSON, add styleJSON) *styleJSON {
	out := styleJSON{}
	if base != nil {
		out = *base
	}
	if add.BackgroundColor != nil {
		out.BackgroundColor = add.BackgroundColor
	}
	if add.TextColor != nil {
		out.TextColor = add.TextColor
	}
	if add.FontSize != nil {
		out.FontSize = add.FontSize
	}
	if add.Padding != nil {
		out.Padding = add.Padding
	}
	if add.BorderWidth != nil {
		out.BorderWidth = add.BorderWidth
	}
	if add.BorderColor != nil {
		out.BorderColor = add.BorderColor
	}
	if add.FlexGrow != nil {
		out.FlexGrow = add.FlexGrow
	}
	if add.CornerRadius != nil {
		out.CornerRadius = add.CornerRadius
	}
	if add.Bold != nil {
		out.Bold = add.Bold
	}
	if add.FontDensity != nil {
		out.FontDensity = add.FontDensity
	}
	if add.MinFontSize != nil {
		out.MinFontSize = add.MinFontSize
	}
	return &out
}

// mergeStyle overlays non-zero fields from override onto base. Use stylePatch on
// Element for explicit zeros (e.g. padding: 0) from DocumentSpec blocks.
func mergeStyle(base, override Style) Style {
	if override.BackgroundColor.A != 0 {
		base.BackgroundColor = override.BackgroundColor
	}
	if override.TextColor.A != 0 {
		base.TextColor = override.TextColor
	}
	if override.FontSize != 0 {
		base.FontSize = override.FontSize
	}
	if override.Padding != 0 {
		base.Padding = override.Padding
	}
	if override.BorderWidth != 0 {
		base.BorderWidth = override.BorderWidth
	}
	if override.BorderColor.A != 0 {
		base.BorderColor = override.BorderColor
	}
	if override.FlexGrow != 0 {
		base.FlexGrow = override.FlexGrow
	}
	if override.CornerRadius != 0 {
		base.CornerRadius = override.CornerRadius
	}
	if override.Bold {
		base.Bold = true
	}
	if override.FontDensity != 0 {
		base.FontDensity = override.FontDensity
	}
	if override.MinFontSize != 0 {
		base.MinFontSize = override.MinFontSize
	}
	if override.MaxFontSize != 0 {
		base.MaxFontSize = override.MaxFontSize
		// Capped spans (status bar pipe) must not inherit a parent MinFontSize floor.
		if override.MinFontSize == 0 {
			base.MinFontSize = 0
		}
	}
	if override.NoCaptionBold {
		base.NoCaptionBold = true
	}
	if override.Italic {
		base.Italic = true
	}
	if override.Mono {
		base.Mono = true
	}
	if override.PreviewFont {
		base.PreviewFont = true
	}
	return base
}

func styleStateNames(state StyleState) []string {
	if state == StyleStateNone {
		return nil
	}
	names := make([]string, 0, 4)
	if state&StyleStateHover != 0 {
		names = append(names, "hover")
	}
	if state&StyleStatePressed != 0 {
		names = append(names, "pressed")
	}
	if state&StyleStateFocus != 0 {
		names = append(names, "focus")
	}
	if state&StyleStateOpen != 0 {
		names = append(names, "open")
	}
	if state&StyleStateChecked != 0 {
		names = append(names, "checked")
	}
	if state&StyleStateDisabled != 0 {
		names = append(names, "disabled")
	}
	return names
}

func disabledControlStyle() Style {
	return Style{
		BackgroundColor: rl.NewColor(241, 242, 247, 255),
		TextColor:       rl.NewColor(145, 148, 165, 255),
		BorderColor:     rl.NewColor(210, 212, 222, 255),
	}
}

