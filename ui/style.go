// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// Theme metric scale (CurrentTheme)
//
// Use these tiers when adding or tuning styles so demos and widgets stay visually
// coherent. Values are suggestions, not hard limits.
//
//	CornerRadius:
//	  6  — compact controls: button, primary, danger, icon-button, input,
//	       dropdown, datepicker, checkbox, datatable frame, colorpicker swatch.
//	  8  — medium surfaces: panel, card, treeview, accordion chrome, toolbar,
//	       context menu, filepicker root (soft panel).
//	  8–10 — large chrome: card / modal dialogs (~10); toasts (~8); vector-shadow
//	         tiles (~12) stay more “card-like”.
//
//	Padding (uniform):
//	  0–4  — transparent flex rows, list/tree insets, colorpicker tap pad.
//	  8–10 — tabs, badge chips, control interiors, contextmenu rows.
//	  12–14 — accordion body, auxiliary panels.
//	  16–20 — panel, form, card, modal body.
//
//	BorderWidth:
//	  0     — ghost / interior-only.
//	  1     — hairline: list, card, modal shell, separators.
//	  1.5   — panel emphasis, standard buttons.
//	  2     — text fields (input, dropdown, datepicker), toggle, checkbox.
//
// Style describes the visual properties of a widget — analogous to a CSS class.
// Every widget has a styleName that is looked up in CurrentTheme each frame.
// Styles are value types; changing CurrentTheme[name] immediately affects all
// widgets that reference that name.
type Style struct {
	BackgroundColor rl.Color `json:"backgroundColor"`
	TextColor       rl.Color `json:"textColor"`
	FontSize        int32    `json:"fontSize"`
	Padding         float32  `json:"padding"` // uniform for simplicity
	BorderWidth     float32  `json:"borderWidth"`
	BorderColor     rl.Color `json:"borderColor"`
	FlexGrow        float32  `json:"flexGrow"`
	CornerRadius    float32  `json:"cornerRadius"` // for rounded rectangles
	// Bold triggers the heavier SDF shader variant (edge shift ≈+9% weight).
	Bold bool `json:"bold"`
	// Italic selects the italic UI face when loaded (Poppins Italic).
	Italic bool `json:"italic"`
	// Mono selects Fira Code for code spans.
	Mono bool `json:"mono"`
	// PreviewFont selects the Inter preview stack (markdown preview lane).
	PreviewFont bool `json:"previewFont"`
	// FontDensity scales the rendered font size without changing layout sizes
	// (0 == 1.0, i.e. unscaled). Values like 1.1 make text slightly denser.
	FontDensity float32 `json:"fontDensity"`
	// MinFontSize, when > 0, overrides the global minRenderPx floor for this
	// particular style. Use on styles where even 13px would be too large.
	MinFontSize int32 `json:"minFontSize"`
	// MaxFontSize, when > 0, caps EffectiveFontSize so rem scaling does not
	// grow past a fixed pixel size (status bar separators, fixed chrome labels).
	MaxFontSize int32 `json:"maxFontSize"`
	// NoCaptionBold disables automatic bold SDF for FontSize < 15 (chrome captions).
	NoCaptionBold bool `json:"noCaptionBold"`
}

// DefaultStyle provides sensible defaults.
var DefaultStyle = Style{
	BackgroundColor: rl.NewColor(255, 255, 255, 255),
	TextColor:       rl.NewColor(30, 30, 30, 255),
	FontSize:        18,
	Padding:         10,
	BorderWidth:     1,
	BorderColor:     rl.NewColor(180, 180, 180, 255),
	FlexGrow:        0,
	CornerRadius:    0,
}

// Theme is a named collection of Styles.
// Swap CurrentTheme at runtime to implement light/dark mode or brand themes.
type Theme map[string]Style

// CurrentTheme is the active theme. Widgets call Element.GetStyle() which
// reads from this map. Override or extend it before the main loop to apply
// a custom look.
//
// Typography scale (all rendering goes through drawTextS → SDF / DefaultFont):
//   - Theme FontSize tokens are rem-like units on a 16 px grid ([TypeScaleReference]).
//   - At runtime, EffectiveFontSize = token × (RootFontSize / 16). At 640 px width,
//     RootFontSize is 22 px ([TypeScaleMinRoot]), so default (18) ≈ 25 px on screen.
//   - Body & fields: "default", "input", "dropdown", "datepicker" — 18 pt token
//   - Captions & hints: "form-label" (15 pt); "form-value" (18 pt) for readouts
//   - Preset/card body: "surface-body" — 18 pt (synced from tinted chrome)
//   - Chrome: "panel-title" 16 pt; "header" 28 pt; "header-subtitle" 18 pt; "button" 20 pt
//
// Demos: form-label / form-value for field captions and hints; default for paragraph body.
//
// Visual design references (Bootstrap 2/5, DevExpress Ribbon): docs/DESIGN_REFERENCE.md
//
// ── Future: SDF / MSDF Font Rendering ────────────────────────────────────────
// When upgrading to multi-channel SDF fonts for crisp text at all scales:
//  1. Replace drawText / measureText in font.go with an MSDF shader pipeline.
//     Style.FontSize values are semantic (pt) — they carry through unchanged.
//  2. Increase InitFonts atlasSize to 1 for SDF (single-channel) or keep 64
//     and switch to an off-screen MSDF texture baked with msdf-atlas-gen.
//  3. For supersampling / SSAA: render to a 2× FBO in main.go and blit down
//     with a box filter — no widget code changes required.
//  4. Mark styles that need sharp edges (separators, borders) separately so
//     the SDF shader can skip anti-aliasing for hairlines.
//
// ─────────────────────────────────────────────────────────────────────────────
var CurrentTheme = Theme{
	// Default: clean white card, dark text, 18 pt — base for all widgets
	"default": DefaultStyle,
	// Transparent: zero-background layout box — use instead of "separator" for
	// invisible flex containers (cbRow pairs, toolbar rows, etc.).
	"transparent": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         0,
	},
	// editor-scroll — notepad / code editor host viewport (no inset border).
	"editor-scroll": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		BorderWidth:     0,
		Padding:         0,
	},
	// titlebar — borderless window chrome (TitleBar.Draw reads these via SetAppearance).
	"titlebar": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		TextColor:       rl.NewColor(28, 30, 48, 255),
		BorderColor:     rl.NewColor(222, 226, 230, 255),
	},
	"titlebar-hover": {
		BackgroundColor: rl.NewColor(0, 0, 0, 14),
	},
	// notepad-preview — Markdown preview lane inner padding.
	"notepad-preview": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		BorderWidth:     0,
		Padding:         20,
	},
	// appshell-content — legacy 8px inset; prefer MountAppShellRoot with page-scroll
	// viewport as a direct shell child (Document + page-scroll own the band).
	"appshell-content": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         8,
	},
	// appshell-footer — pinned shell footer (pagination / page controls).
	"appshell-footer": {
		BackgroundColor: rl.NewColor(252, 253, 255, 240),
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 220, 232, 255),
		Padding:         10,
	},
	// settings-section — grouped list section title (settings_demo.go).
	"settings-section": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(45, 48, 62, 255),
		FontSize:        17,
		Bold:            true,
	},
	// settings-section-wrap — column wrapper; spacer child adds top gap.
	"settings-section-wrap": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         0,
	},
	// settings-chip-row — horizontal badge/filter row; inset avoids scissor clip.
	"settings-chip-row": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         4,
	},
	// settings-row-label — caption beside inline controls (e.g. rating).
	"settings-row-label": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(45, 48, 62, 255),
		FontSize:        15,
	},
	// page-grid — responsive page grid shell under an absolute root: inset
	// content from window edges without drawing a chrome background.
	"page-grid": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         0,
	},
	// page-scroll — page viewport (MountAppPage): fills the client area; padding +
	// scrollbar gutter live inside this box (web-style — one viewport for all page content).
	"page-scroll": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         16,
	},
	// surface-muted — light grey scroll/list band (notepad Open Notes, settings pages).
	"surface-muted": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		BorderWidth:     0,
		Padding:         12,
	},
	// list-pane — docked master-list column: flat grey, no border/radius (list_pane_shell.go).
	"list-pane": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		BorderWidth:     0,
		CornerRadius:    0,
		Padding:         0,
	},
	// list-pane-scroll — flush vertical scroll host; scrollbar sits on the split edge.
	"list-pane-scroll": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         0,
	},
	// list-pane-list — inner list padding (horizontal inset for rows; scroll bar stays outer).
	"list-pane-list": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         12,
	},
	// list-pane-header — title row above the scroll list (flat, no border seam).
	"list-pane-header": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         12,
	},
	// list-pane-header-title — sidebar section title (customize via SetStyleOverrides).
	"list-pane-header-title": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(48, 52, 68, 255),
		FontSize:        18,
		Bold:            true,
		NoCaptionBold:   true,
		Padding:         0,
	},
	// settings-page-title — large in-shell settings heading.
	"settings-page-title": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(28, 32, 48, 255),
		FontSize:        32,
		Bold:            true,
		Padding:         0,
	},
	// settings-scroll — in-shell settings viewport (below MenuBar). Transparent like
	// page-scroll; vertical padding is 0 (flush to MenuBar/StatusBar). Horizontal
	// inset comes from scrollContentPadding (settingsScrollPadH).
	"settings-scroll": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         0,
	},
	// settings-body — main band under MenuBar (opaque; content must not show through).
	"settings-body": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		BorderWidth:     0,
		Padding:         0,
	},
	"settings-flat-band": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		BorderWidth:     0,
		CornerRadius:    0,
		Padding:         16,
	},
	// page-shell — legacy outer inset (20px). Do not wrap a page-scroll viewport;
	// app-shell scenes use appshell-content + NewAppShellScrollViewport instead.
	"page-shell": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         20,
	},
	// Viewport/panel background: light warm grey with a subtle border and rounded corners
	"panel": {
		BackgroundColor: rl.NewColor(245, 245, 248, 255),
		TextColor:       rl.NewColor(30, 30, 30, 255),
		Padding:         16,
		BorderWidth:     1.5,
		BorderColor:     rl.NewColor(210, 210, 215, 255),
		CornerRadius:    8,
	},
	// Standard button: slate blue-grey, white text, rounded corners
	"button": {
		BackgroundColor: rl.NewColor(90, 100, 120, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		FontSize:        17,
		Bold:            true,
		Padding:         8,
		BorderWidth:     0,
		CornerRadius:    6,
	},
	// Icon button: same geometry as "button" (shared control radius).
	// Use SetStyle("primary") or SetStyle("danger") on an IconButton to change
	// the colour family; this entry is only the neutral default.
	"icon-button": {
		BackgroundColor: rl.NewColor(90, 100, 120, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		FontSize:        17,
		Bold:            true,
		Padding:         8,
		BorderWidth:     0,
		CornerRadius:    6,
	},
	// AppBar / toolbar ghost icon (Phosphor); no filled slab — hover draws a soft disc.
	"appbar-icon": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(45, 48, 62, 255),
		FontSize:        26,
		Padding:         8,
		BorderWidth:     0,
		CornerRadius:    8,
	},
	// Primary action button: indigo, same control radius as "button".
	"primary": {
		BackgroundColor: rl.NewColor(79, 70, 229, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		FontSize:        17,
		Bold:            true,
		Padding:         8,
		BorderWidth:     0,
		CornerRadius:    6,
	},
	// Danger / destructive action: red, rounded corners
	"danger": {
		BackgroundColor: rl.NewColor(220, 38, 38, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		FontSize:        17,
		Bold:            true,
		Padding:         8,
		BorderWidth:     0,
		CornerRadius:    6,
	},
	// Header widget: transparent background, large bold title, indigo accent bar
	"header": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0), // transparent — sits on panel bg
		TextColor:       rl.NewColor(28, 28, 38, 255),
		FontSize:        28,
		Bold:            true,
		Padding:         16,
	},
	// Header subtitle: muted secondary text beneath the title
	"header-subtitle": {
		TextColor:   rl.NewColor(100, 102, 120, 255),
		FontSize:    18,
		FontDensity: 1.05,
	},
	// Text input: white bg with blue focus ring (applied in drawInternal).
	// TextInput defaults to this key in NewTextInput; padding drives horizontal inset.
	"input": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(20, 20, 20, 255),
		FontSize:        18,
		Padding:         8,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(190, 192, 200, 255),
		CornerRadius:    6,
	},
	// Panel title bar: deep slate background, near-white text
	"panel-title": {
		BackgroundColor: rl.NewColor(42, 44, 58, 255),
		TextColor:       rl.NewColor(225, 226, 234, 255),
		FontSize:        16,
		Bold:            true,
		FontDensity:     1.05,
		Padding:         10,
	},
	// Progress bar track — pill-shaped, borderless, clean
	"progress": {
		BackgroundColor: rl.NewColor(226, 228, 234, 255),
		TextColor:       rl.NewColor(20, 20, 20, 255),
		BorderWidth:     0,
		CornerRadius:    100, // clamped to 1.0 in draw → full pill shape
	},
	// Progress bar fill — indigo pill matching the track shape
	"progress-fill": {
		BackgroundColor: rl.NewColor(79, 70, 229, 255),
		CornerRadius:    100,
	},
	// List panel inside viewport
	"list": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(33, 37, 41, 255),
		Padding:         0,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(222, 226, 230, 255),
		CornerRadius:    4,
	},
	// list-row — Bootstrap list-group-item row inside VirtualList.
	"list-row": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(33, 37, 41, 255),
		FontSize:        18,
		Padding:         8,
	},
	// list-flush — no inset/border so a nested DataTable fills the viewport edge-to-edge.
	"list-flush": {
		BackgroundColor: rl.NewColor(248, 249, 251, 255),
		TextColor:       rl.NewColor(30, 30, 30, 255),
		Padding:         0,
		BorderWidth:     0,
	},
	// Selected list item
	"list-selected": {
		BackgroundColor: rl.NewColor(237, 236, 254, 255),
		TextColor:       rl.NewColor(79, 70, 229, 255),
		FontSize:        18,
		Padding:         8,
	},
	// Toggle track — off state: cool light grey
	// Toggle ring colors (drawn in toggle.go; matches segmented track).
	"toggle": {
		BackgroundColor: rl.NewColor(235, 237, 245, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		BorderWidth:     1,
		BorderColor:     rl.NewColor(220, 222, 234, 255),
	},
	"toggle-on": {
		BackgroundColor: rl.NewColor(79, 70, 229, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		BorderWidth:     1,
		BorderColor:     rl.NewColor(55, 48, 163, 255),
	},
	// Dropdown button: thin border + caret (matches toolbar-menu / ribbon fields).
	"dropdown": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(73, 80, 87, 255),
		FontSize:        18,
		Padding:         6,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(222, 226, 230, 255),
		CornerRadius:    4,
	},
	// dropdown-selected — popup list item text when the current value is highlighted.
	"dropdown-selected": {
		TextColor: rl.NewColor(79, 70, 229, 255),
	},
	// DatePicker field: same chrome as dropdown (compact menu trigger).
	"datepicker": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 30, 30, 255),
		FontSize:        18,
		Padding:         10,
		BorderWidth:     2,
		BorderColor:     rl.NewColor(190, 192, 200, 255),
		CornerRadius:    6,
	},
	"combobox": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 30, 30, 255),
		FontSize:        18,
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 222, 232, 255),
		CornerRadius:    4,
	},
	"daterangepicker": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 30, 30, 255),
		FontSize:        18,
		Padding:         10,
		BorderWidth:     2,
		BorderColor:     rl.NewColor(190, 192, 200, 255),
		CornerRadius:    6,
	},
	// FilePicker root: soft panel surface for the composed browser chrome.
	"filepicker": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		TextColor:       rl.NewColor(30, 30, 30, 255),
		FontSize:        16,
		Padding:         16,
		BorderWidth:     1.5,
		BorderColor:     rl.NewColor(210, 210, 215, 255),
		CornerRadius:    6,
	},
	// Separator rule: no background, grey line, small muted text
	"separator": {
		TextColor:   rl.NewColor(160, 162, 175, 255),
		FontSize:    15,
		FontDensity: 1.1,
		BorderWidth: 1,
		BorderColor: rl.NewColor(218, 220, 228, 255),
	},
	// Checkbox unchecked — white bg, rounded corners, grey border
	"checkbox": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255), // checkmark colour (used via checked state)
		BorderWidth:     1,
		BorderColor:     rl.NewColor(190, 192, 204, 255),
		CornerRadius:    6,
	},
	// Checkbox checked — indigo fill, white checkmark
	"checkbox-checked": {
		BackgroundColor: rl.NewColor(79, 70, 229, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		BorderWidth:     1,
		BorderColor:     rl.NewColor(190, 192, 204, 255),
		CornerRadius:    6,
	},
	// Card container — white bg; border matches panel chrome (polish §4.1).
	"card": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 30, 40, 255),
		Padding:         16,
		BorderWidth:     1.5,
		BorderColor:     rl.NewColor(210, 210, 215, 255),
		CornerRadius:    8,
	},
	// card-callout / card-code — legacy fallback when resolving card variants (see theme_v2.go).
	"card-callout": {
		BackgroundColor: rl.NewColor(165, 180, 252, 255),
		TextColor:       rl.NewColor(30, 27, 75, 255),
		Padding:         16,
		BorderWidth:     2,
		BorderColor:     rl.NewColor(67, 56, 202, 255),
		CornerRadius:    8,
	},
	"card-code": {
		BackgroundColor: rl.NewColor(248, 249, 251, 255),
		TextColor:       rl.NewColor(40, 42, 54, 255),
		Padding:         16,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 214, 222, 255),
		CornerRadius:    8,
	},
	"card-blockquote": {
		BackgroundColor: rl.NewColor(248, 249, 251, 255),
		TextColor:       rl.NewColor(40, 42, 54, 255),
		Padding:         12,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 214, 222, 255),
		CornerRadius:    6,
	},
	// Card title — slightly darker/bolder than body text, no coloured bg
	"card-title": {
		TextColor:   rl.NewColor(28, 30, 48, 255),
		FontSize:    16,
		Bold:        true,
		FontDensity: 1.08,
	},
	// icon — transparent background, used for standalone icon-only widgets.
	// No border, no padding; size/position is fully controlled by the caller.
	"icon": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		BorderWidth:     0,
		Padding:         0,
	},
	// vector-shadow — a card-like tile for displaying gg-rendered vector
	// graphics with a subtle shadow feel. Slightly off-white bg, soft border,
	// generous corner radius. Pair with DrawRoundedRectWithShadow in vector.go.
	"vector-shadow": {
		BackgroundColor: rl.NewColor(250, 251, 255, 255),
		TextColor:       rl.NewColor(40, 42, 60, 255),
		FontSize:        15,
		Padding:         12,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(200, 204, 220, 255),
		CornerRadius:    12,
	},

	// ── Batch 1 Widget Styles ─────────────────────────────────────────────────

	// tooltip — compact dark overlay for hover popups.
	// Drop shadow is drawn by drawTooltipPopup; CornerRadius 5 gives ~0.3
	// roundness on a typical 31px-tall popup.
	"tooltip": {
		BackgroundColor: rl.NewColor(22, 24, 40, 255),
		TextColor:       rl.NewColor(236, 238, 252, 255),
		FontSize:        15,
		FontDensity:     1.0,
		Padding:         10,
		BorderWidth:     0,
		CornerRadius:    5,
	},

	// toast-label / toast-body — overlay notification typography (toast.go).
	"toast-label": {
		TextColor:   rl.NewColor(55, 65, 81, 255),
		FontSize:    14,
		Bold:        true,
		FontDensity: 1.0,
	},
	"toast-body": {
		TextColor:     rl.NewColor(31, 41, 55, 255),
		FontSize:      16,
		FontDensity:   1.0,
		NoCaptionBold: true,
	},
	// toast-action — trailing action label (ShowToastWithAction).
	"toast-action": {
		TextColor:   rl.NewColor(79, 70, 229, 255),
		FontSize:    16,
		Bold:        true,
		FontDensity: 1.0,
	},

	// list-tile / list-tile-subtitle — settings row typography (listtile.go).
	"list-tile": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(28, 32, 48, 255),
		FontSize:        18,
		NoCaptionBold:   true,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(232, 234, 240, 255),
		CornerRadius:    4,
		FontDensity:     1.0,
	},
	"list-tile-subtitle": {
		TextColor:   rl.NewColor(100, 106, 128, 255),
		FontSize:    16,
		FontDensity: 1.0,
	},
	// list-tile-pane — alias; same chrome as list-tile (docked list panes).
	"list-tile-pane": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(28, 32, 48, 255),
		FontSize:        18,
		NoCaptionBold:   true,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(232, 234, 240, 255),
		CornerRadius:    4,
		FontDensity:     1.0,
	},
	// list-tile-bordered — alias; same chrome as list-tile (settings panel rows).
	"list-tile-bordered": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(28, 32, 48, 255),
		FontSize:        18,
		NoCaptionBold:   true,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(232, 234, 240, 255),
		CornerRadius:    4,
		FontDensity:     1.0,
	},
	// spinbox — numeric stepper chrome (spinbox.go).
	"spinbox": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(31, 41, 55, 255),
		FontSize:        15,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 213, 228, 255),
		CornerRadius:    6,
	},
	// segmented — iOS-style option strip (segmentedcontrol.go).
	"segmented": {
		BackgroundColor: rl.NewColor(235, 237, 245, 255),
		TextColor:       rl.NewColor(90, 94, 118, 255),
		FontSize:        14,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(220, 222, 234, 255),
		CornerRadius:    8,
	},
	// appbar — top shell bar (appbar.go).
	"appbar": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(22, 24, 40, 255),
		FontSize:        22,
		Bold:            true,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 220, 232, 255),
	},
	// bottomnav — inactive bottom navigation item (bottomnavigation.go).
	"bottomnav": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(110, 114, 140, 255),
		FontSize:        16,
		NoCaptionBold:   true,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 220, 232, 255),
	},
	// bottomnav-active — selected bottom navigation item.
	"bottomnav-active": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(79, 70, 229, 255),
		FontSize:        16,
		Bold:            true,
		BorderWidth:     2,
		BorderColor:     rl.NewColor(79, 70, 229, 255),
	},
	// navigationrail — vertical nav strip (navigationrail.go).
	"navigationrail": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		TextColor:       rl.NewColor(110, 114, 140, 255),
		FontSize:        12,
		NoCaptionBold:   true,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 220, 232, 255),
	},
	// menubar — desktop menu strip (menubar.go).
	"menubar": {
		BackgroundColor: rl.NewColor(252, 252, 254, 255),
		TextColor:       rl.NewColor(38, 40, 54, 255),
		FontSize:        18,
		MinFontSize:     TypeScaleMinMenubarPx,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 220, 232, 255),
	},
	"menubar-hover": {
		BackgroundColor: rl.NewColor(232, 234, 244, 255),
		TextColor:       rl.NewColor(22, 24, 40, 255),
		FontSize:        18,
		MinFontSize:     TypeScaleMinMenubarPx,
	},
	// statusbar — bottom desktop status strip (statusbar.go).
	"statusbar": {
		BackgroundColor: rl.NewColor(244, 245, 249, 255),
		TextColor:       rl.NewColor(100, 104, 128, 255),
		FontSize:        16,
		NoCaptionBold:   true,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 220, 232, 255),
	},
	// statusbar-label — status metrics (floor TypeScaleMinStatusPx effective px).
	"statusbar-label": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(90, 94, 118, 255),
		FontSize:        17,
		MinFontSize:     TypeScaleMinStatusPx,
		NoCaptionBold:   true,
		Padding:         0,
		BorderWidth:     0,
	},
	// statusbar-pipe — doc-column separator (literal 16px cap; lighter than labels).
	"statusbar-pipe": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(160, 164, 180, 255),
		FontSize:        16,
		MaxFontSize:     16,
		NoCaptionBold:   true,
		Padding:         0,
		BorderWidth:     0,
	},
	// text-editor — plain document body (token 18; floor TypeScaleMinEditorPx).
	"text-editor": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(28, 30, 44, 255),
		FontSize:        18,
		MinFontSize:     TypeScaleMinEditorPx,
		Padding:         12,
		BorderWidth:     0,
	},
	// fab — floating action button (fab.go).
	"fab": {
		BackgroundColor: rl.NewColor(79, 70, 229, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		FontSize:        14,
		Bold:            true,
		Padding:         16,
		BorderWidth:     0,
		CornerRadius:    28,
	},
	// avatar — circle initials / photo frame (avatar.go).
	"avatar": {
		BackgroundColor: rl.NewColor(199, 210, 254, 255),
		TextColor:       rl.NewColor(55, 48, 163, 255),
		FontSize:        14,
		Bold:            true,
	},
	// rating — star fill color (rating.go); empty stars use neutral grey in Draw.
	"rating": {
		TextColor: rl.NewColor(250, 204, 21, 255),
		FontSize:  24,
	},
	// breadcrumbs — path segments (breadcrumbs.go).
	"breadcrumbs": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(79, 70, 229, 255),
		FontSize:        14,
		Padding:         12,
	},
	// pagination — page button strip (pagination.go).
	"pagination": {
		BackgroundColor: rl.NewColor(245, 246, 250, 255),
		TextColor:       rl.NewColor(45, 48, 62, 255),
		FontSize:        14,
	},

	// tooltip-light — same geometry, light background for widgets on dark surfaces.
	"tooltip-light": {
		BackgroundColor: rl.NewColor(252, 253, 255, 255),
		TextColor:       rl.NewColor(22, 24, 40, 255),
		FontSize:        15,
		Padding:         9,
		BorderWidth:     0,
		CornerRadius:    5,
	},

	// tab — inactive tab header; light grey text on transparent/light bg.
	"tab": {
		BackgroundColor: rl.NewColor(235, 237, 245, 255),
		TextColor:       rl.NewColor(110, 114, 140, 255),
		FontSize:        16,
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 213, 228, 255),
		CornerRadius:    0,
	},

	// tab-active — active tab header; white bg, dark text, indigo bottom accent.
	"tab-active": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        16,
		Bold:            true,
		Padding:         10,
		BorderWidth:     2,
		BorderColor:     rl.NewColor(99, 102, 241, 255),
		CornerRadius:    0,
	},

	// tab-hover — hovered inactive tab header; slightly lighter background.
	"tab-hover": {
		BackgroundColor: rl.NewColor(220, 223, 236, 255),
		TextColor:       rl.NewColor(60, 64, 90, 255),
		FontSize:        16,
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(190, 194, 215, 255),
		CornerRadius:    0,
	},

	// tab-disabled — muted, non-interactive tab header.
	"tab-disabled": {
		BackgroundColor: rl.NewColor(235, 237, 245, 255),
		TextColor:       rl.NewColor(170, 172, 186, 255),
		FontSize:        16,
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 213, 228, 255),
		CornerRadius:    0,
	},

	// drawer — side navigation panel (drawer.go).
	"drawer": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        15,
		Padding:         12,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 220, 232, 255),
		CornerRadius:    0,
	},
	// bottomsheet — slide-up sheet surface (bottomsheet.go).
	"bottomsheet": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        15,
		Padding:         16,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 220, 232, 255),
		CornerRadius:    12,
	},
	// commandpalette — fuzzy command picker overlay (commandpalette.go).
	"commandpalette": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(22, 24, 40, 255),
		FontSize:        15,
		Padding:         12,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(200, 204, 220, 255),
		CornerRadius:    10,
	},
	"commandpalette-input": {
		BackgroundColor: rl.NewColor(244, 245, 249, 255),
		TextColor:       rl.NewColor(22, 24, 40, 255),
		FontSize:        16,
		Padding:         10,
		BorderWidth:     0,
		CornerRadius:    8,
	},
	"commandpalette-item": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(38, 40, 54, 255),
		FontSize:        15,
	},
	"commandpalette-selected": {
		BackgroundColor: rl.NewColor(232, 234, 255, 255),
		TextColor:       rl.NewColor(55, 48, 163, 255),
		FontSize:        15,
	},
	// timeline — vertical activity feed (timeline.go).
	"timeline": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(22, 24, 40, 255),
		FontSize:        16,
		Bold:            true,
	},
	"timeline-subtitle": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(100, 104, 128, 255),
		FontSize:        14,
	},
	"timeline-time": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(130, 134, 158, 255),
		FontSize:        13,
		NoCaptionBold:   true,
	},
	// colorwell — compact swatch row (colorwell.go).
	"colorwell": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        14,
	},
	// carousel — slide viewport chrome (carousel.go).
	"carousel": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		BorderColor:     rl.NewColor(220, 224, 240, 255),
		BorderWidth:     1,
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        15,
	},
	// resizablepanel — multi-pane splitter host (resizablepanel.go).
	"resizablepanel": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        14,
	},

	// modal — large popup shell; same corner radius family as contextmenu.
	"modal": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        18,
		Padding:         20,
		BorderWidth:     2,
		BorderColor:     rl.NewColor(210, 213, 228, 255),
		CornerRadius:    12,
	},

	// modal-title — bold header text inside modal.
	"modal-title": {
		TextColor:   rl.NewColor(20, 22, 38, 255),
		FontSize:    17,
		Bold:        true,
		FontDensity: 1.05,
	},

	// modal-body — muted copy and field captions inside modal content (1× overlay draw).
	"modal-body": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(90, 94, 120, 255),
		FontSize:        15,
		Padding:         0,
		BorderWidth:     0,
	},

	// modal-input — find/replace fields on the modal shell (always readable placeholders).
	"modal-input": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        18,
		Padding:         8,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(190, 192, 200, 255),
		CornerRadius:    6,
	},

	// spinner — transparent background; colour comes from the widget's Color field.
	"spinner": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         0,
	},

	// radio — unselected radio option; grey circle.
	"radio": {
		BackgroundColor: rl.NewColor(240, 241, 248, 255),
		TextColor:       rl.NewColor(60, 62, 80, 255),
		FontSize:        15,
		BorderWidth:     2,
		BorderColor:     rl.NewColor(180, 182, 200, 255),
		CornerRadius:    0,
	},

	// radio-selected — active radio option; indigo fill.
	"radio-selected": {
		BackgroundColor: rl.NewColor(99, 102, 241, 255),
		TextColor:       rl.NewColor(255, 255, 255, 255),
		FontSize:        15,
		BorderWidth:     2,
		BorderColor:     rl.NewColor(79, 82, 200, 255),
		CornerRadius:    0,
	},

	// radio-hover — hovered unselected row highlight (low-opacity indigo wash).
	"radio-hover": {
		BackgroundColor: rl.NewColor(99, 102, 241, 22),
		TextColor:       rl.NewColor(60, 62, 80, 255),
		FontSize:        15,
		BorderWidth:     0,
		CornerRadius:    4,
	},

	// radio-disabled — greyed-out, non-interactive option.
	"radio-disabled": {
		BackgroundColor: rl.NewColor(230, 231, 238, 255),
		TextColor:       rl.NewColor(170, 172, 186, 255),
		FontSize:        15,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(205, 207, 220, 255),
		CornerRadius:    0,
	},

	// treeview — tree container background.
	"treeview": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		TextColor:       rl.NewColor(40, 42, 60, 255),
		FontSize:        15,
		NoCaptionBold:   true,
		Padding:         4,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 213, 228, 255),
		CornerRadius:    6,
	},

	// treeview-selected — highlighted tree row.
	"treeview-selected": {
		BackgroundColor: rl.NewColor(99, 102, 241, 30),
		TextColor:       rl.NewColor(20, 22, 50, 255),
		FontSize:        15,
		BorderWidth:     0,
		CornerRadius:    4,
	},

	// treeview-hover — hovered (non-selected) tree row.
	"treeview-hover": {
		BackgroundColor: rl.NewColor(99, 102, 241, 14),
		TextColor:       rl.NewColor(40, 42, 60, 255),
		FontSize:        15,
		BorderWidth:     0,
		CornerRadius:    0,
	},

	// ── Form Styles ─────────────────────────────────────────────────────────────

	// form — container background for Form widgets.
	// Light warm white, thin border, comfortable padding.
	"form": {
		BackgroundColor: rl.NewColor(252, 253, 255, 255),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        15,
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(214, 217, 232, 255),
		CornerRadius:    6,
	},

	// form-label — right-aligned (or caption) text beside each control.
	// Muted slate, 14 pt minimum, slightly dense.
	"form-label": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(90, 94, 120, 255),
		FontSize:        15,
		NoCaptionBold:   true,
	},

	// toolbar-caption — compact checkbox / toggle label beside a control.
	"toolbar-caption": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(90, 94, 120, 255),
		FontSize:        15,
		NoCaptionBold:   true,
	},

	// form-field-caption — title above a generated form control (Name, Density, etc.).
	"form-field-caption": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        19,
		Bold:            true,
	},

	// form-value — readout / status text beside controls (body size, not caption).
	"form-value": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        18,
	},

	// label — inline text on surfaces (transparent; no white block behind copy).
	"label": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 30, 30, 255),
		FontSize:        18,
		Padding:         0,
		BorderWidth:     0,
	},

	// surface-body — paragraph body inside tinted preset surfaces (neo-glow, glass, …).
	// ApplySurfaceBodyTypography merges chrome TextColor + FontSize onto Label/RichText children.
	"surface-body": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 27, 75, 255),
		FontSize:        18,
		Padding:         0,
		BorderWidth:     0,
	},

	// form-error — inline validation error text shown below an invalid control.
	// Red, 14 pt minimum, no background.
	"form-error": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(220, 38, 38, 255),
		FontSize:        15,
	},

	// contextmenu — popup menu container.
	"contextmenu": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        16,
		Padding:         8,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 213, 228, 255),
		CornerRadius:    12,
	},

	// contextmenu-item — normal menu item row.
	"contextmenu-item": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        16,
		Padding:         10,
		BorderWidth:     0,
	},

	// contextmenu-shortcut — trailing shortcut hint in menu rows.
	"contextmenu-shortcut": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(118, 122, 146, 255),
		FontSize:        14,
		NoCaptionBold:   true,
		Padding:         0,
		BorderWidth:     0,
	},

	// contextmenu-hover — hovered menu item row.
	"contextmenu-hover": {
		BackgroundColor: rl.NewColor(237, 238, 246, 255),
		TextColor:       rl.NewColor(20, 22, 50, 255),
		FontSize:        16,
		Padding:         10,
		BorderWidth:     0,
	},

	// contextmenu-divider — thin separator line between menu sections.
	"contextmenu-divider": {
		BackgroundColor: rl.NewColor(210, 213, 228, 255),
		BorderWidth:     0,
	},

	// ── Batch 2 Widget Styles ─────────────────────────────────────────────────

	// searchbar — pill-shaped search input with light grey-lavender background.
	// CornerRadius 100 is clamped to 1.0 at runtime → full pill ends.
	// The search icon and clear button are drawn by SearchBar.drawInternal;
	// this style controls background, text colour, font size, and border.
	"searchbar": {
		BackgroundColor: rl.NewColor(240, 241, 248, 255),
		TextColor:       rl.NewColor(20, 22, 40, 255),
		FontSize:        16,
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(208, 210, 226, 255),
		CornerRadius:    100, // clamped to 1.0 → pill shape
	},

	// ── Batch 2 continued: ColorPicker ───────────────────────────────────────

	// colorpicker — the swatch widget that sits in the layout tree.
	// The rounded fill and border are drawn programmatically; this style
	// carries padding and corner-radius values used by sub-draw helpers.
	"colorpicker": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		BorderWidth:     1,
		BorderColor:     rl.NewColor(222, 226, 230, 255),
		CornerRadius:    4,
		Padding:         0,
	},

	// ── Batch 3: Badge styles ──────────────────────────────────────────────────
	//
	// All badge styles share the same geometry (pill shape, 24 px default height,
	// 16 pt text).  Only background/text/border colors differ per variant.
	// CornerRadius=100 is a sentinel that badge.drawInternal maps to roundness=1.0
	// (full pill).  BorderWidth=0 means no outline is drawn.

	// badge — neutral grey chip for generic labels or tags.
	"badge": {
		BackgroundColor: rl.NewColor(228, 229, 238, 255),
		TextColor:       rl.NewColor(50, 54, 80, 255),
		FontSize:        16,
		Padding:         12,
		BorderWidth:     0,
		CornerRadius:    100,
	},

	// badge-primary — indigo chip for category or feature badges.
	"badge-primary": {
		BackgroundColor: rl.NewColor(224, 225, 255, 255),
		TextColor:       rl.NewColor(55, 48, 163, 255),
		FontSize:        16,
		Padding:         12,
		BorderWidth:     0,
	},

	// badge-success — green chip for confirmed or completed states.
	"badge-success": {
		BackgroundColor: rl.NewColor(209, 250, 229, 255),
		TextColor:       rl.NewColor(6, 95, 70, 255),
		FontSize:        16,
		Padding:         12,
		BorderWidth:     0,
	},

	// badge-warning — amber chip for pending or caution states.
	"badge-warning": {
		BackgroundColor: rl.NewColor(254, 243, 199, 255),
		TextColor:       rl.NewColor(146, 64, 14, 255),
		FontSize:        16,
		Padding:         12,
		BorderWidth:     0,
	},

	// badge-danger — red chip for errors, failures, or destructive actions.
	"badge-danger": {
		BackgroundColor: rl.NewColor(254, 226, 226, 255),
		TextColor:       rl.NewColor(153, 27, 27, 255),
		FontSize:        16,
		Padding:         12,
		BorderWidth:     0,
	},

	// badge-info — sky blue chip for informational, low-urgency labels.
	"badge-info": {
		BackgroundColor: rl.NewColor(224, 242, 254, 255),
		TextColor:       rl.NewColor(12, 74, 110, 255),
		FontSize:        16,
		Padding:         12,
		BorderWidth:     0,
	},

	// ── Accordion ──────────────────────────────────────────────────────────────

	// accordion — header bar in its default (collapsed) state.
	"accordion": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		TextColor:       rl.NewColor(30, 35, 60, 255),
		FontSize:        16,
		Padding:         14,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 220, 228, 255),
		CornerRadius:    8,
	},

	// accordion-open — header bar when the accordion is expanded or animating.
	"accordion-open": {
		BackgroundColor: rl.NewColor(238, 242, 255, 255),
		TextColor:       rl.NewColor(30, 35, 60, 255),
		FontSize:        16,
		Padding:         14,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(199, 210, 254, 255),
		CornerRadius:    8,
	},

	// accordion-body — content area; matches panel surface when expanded.
	"accordion-body": {
		BackgroundColor: rl.NewColor(245, 245, 248, 255),
		TextColor:       rl.NewColor(30, 30, 30, 255),
		FontSize:        16,
		Padding:         16,
		BorderWidth:     0,
		BorderColor:     rl.NewColor(210, 210, 215, 255),
		CornerRadius:    8,
	},

	// ── Stepper ───────────────────────────────────────────────────────────────

	// stepper — base widget style (transparent background, used by GetStyle).
	"stepper": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 35, 60, 255),
		FontSize:        19,
		Padding:         0,
		BorderWidth:     0,
	},

	// stepper-title — primary label below/beside each step circle.
	"stepper-title": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 35, 60, 255),
		FontSize:        19,
		Padding:         0,
		BorderWidth:     0,
	},

	// stepper-subtitle — secondary description below the title.
	"stepper-subtitle": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(100, 105, 125, 255),
		FontSize:        16,
		Padding:         0,
		BorderWidth:     0,
	},

	// ── KeyboardShortcut ──────────────────────────────────────────────────────

	// keyboard-shortcut — base style; action label text colour and font size.
	// Key pill colours are baked into the widget (kbTokenBg / kbTokenBorder).
	"keyboard-shortcut": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(40, 44, 60, 255),
		FontSize:        18,
		Padding:         4,
		BorderWidth:     0,
	},

	// ── File preview (FilePicker) ─────────────────────────────────────────────

	// preview-frame — Panel chrome for the document preview block (warm paper tone).
	"preview-frame": {
		BackgroundColor: rl.NewColor(246, 245, 240, 255),
		TextColor:       rl.NewColor(31, 41, 55, 255),
		FontSize:        15,
		Padding:         8,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(198, 194, 182, 255),
		CornerRadius:    8,
	},

	// preview-document — text inside TextViewer (mostly line color/size).
	"preview-document": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(42, 44, 53, 255),
		FontSize:        15,
		Padding:         0,
		BorderWidth:     0,
	},

	// richtext — document-style text block for paragraphs/spans.
	"richtext": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(42, 44, 53, 255),
		FontSize:        18,
		Padding:         10,
		BorderWidth:     0,
	},
	"richtext-muted": {
		TextColor: rl.NewColor(100, 105, 125, 255),
	},
	// richtext-on-dark — body copy on tinted cards (e.g. DocBlock style overrides).
	"richtext-on-dark": {
		TextColor: rl.NewColor(229, 231, 235, 255),
	},
	"richtext-on-dark-code": {
		TextColor:   rl.NewColor(254, 243, 199, 255),
		FontDensity: 1.02,
	},
	"richtext-preview-bold": {
		TextColor:   rl.NewColor(28, 30, 48, 255),
		Bold:        true,
		PreviewFont: true,
	},
	"richtext-preview-italic": {
		TextColor:   rl.NewColor(40, 42, 54, 255),
		Italic:      true,
		PreviewFont: true,
		FontDensity: 1.06,
	},
	"richtext-h1": {
		TextColor:   rl.NewColor(28, 30, 48, 255),
		FontSize:    28,
		Bold:        true,
		PreviewFont: true,
		FontDensity: 1.08,
	},
	"richtext-h2": {
		TextColor:   rl.NewColor(34, 37, 58, 255),
		FontSize:    24,
		Bold:        true,
		PreviewFont: true,
		FontDensity: 1.05,
	},
	"richtext-h3": {
		TextColor:   rl.NewColor(42, 44, 64, 255),
		FontSize:    22,
		Bold:        true,
		PreviewFont: true,
		FontDensity: 1.04,
	},
	"richtext-h4": {
		TextColor:   rl.NewColor(48, 50, 72, 255),
		FontSize:    18,
		Bold:        true,
		PreviewFont: true,
		FontDensity: 1.02,
	},
	"richtext-h5": {
		TextColor:   rl.NewColor(54, 56, 78, 255),
		FontSize:    17,
		Bold:        true,
		PreviewFont: true,
	},
	"richtext-h6": {
		TextColor:   rl.NewColor(62, 64, 86, 255),
		FontSize:    16,
		Bold:        true,
		PreviewFont: true,
	},
	"richtext-callout": {
		TextColor: rl.NewColor(72, 74, 92, 255),
		FontSize:  15,
	},
	// richtext-preview — markdown preview body (Inter when assets are present).
	"richtext-preview": {
		TextColor:    rl.NewColor(40, 42, 54, 255),
		FontSize:     16,
		PreviewFont:  true,
		NoCaptionBold: true,
	},
	"richtext-italic": {
		TextColor: rl.NewColor(58, 60, 82, 255),
		Italic:    true,
	},
	"richtext-strike": {
		TextColor: rl.NewColor(120, 124, 142, 255),
	},
	"richtext-list-marker": {
		TextColor: rl.NewColor(79, 70, 229, 255),
		Bold:      true,
	},
	"richtext-code": {
		BackgroundColor: rl.NewColor(237, 233, 254, 255),
		TextColor:       rl.NewColor(109, 40, 217, 255),
		FontSize:        16,
		Mono:            true,
	},
	"richtext-math": {
		BackgroundColor: rl.NewColor(237, 233, 254, 255),
		TextColor:       rl.NewColor(91, 33, 182, 255),
		FontSize:        16,
		Mono:            true,
		Italic:          true,
		Padding:         2,
	},
	"richtext-math-inline": {
		TextColor: rl.NewColor(91, 33, 182, 255),
		FontSize:  16,
		Mono:      true,
		Italic:    true,
		Padding:   4,
	},
	"richtext-math-display": {
		TextColor: rl.NewColor(76, 29, 149, 255),
		FontSize:  18,
		Mono:      true,
		Padding:   4,
	},
	"richtext-code-block": {
		BackgroundColor: rl.NewColor(248, 249, 251, 255),
		TextColor:       rl.NewColor(55, 58, 78, 255),
		FontSize:        14,
		Mono:            true,
		Padding:         6,
	},
	"richtext-blockquote": {
		TextColor:     rl.NewColor(72, 74, 92, 255),
		FontSize:      16,
		PreviewFont:   true,
		FontDensity:   1.0,
		NoCaptionBold: true,
	},
	"richtext-blockquote-italic": {
		TextColor:     rl.NewColor(72, 74, 92, 255),
		FontSize:      16,
		Italic:        true,
		PreviewFont:   true,
		FontDensity:   1.0,
		NoCaptionBold: true,
	},
	"preview-blockquote": {
		BackgroundColor: rl.NewColor(252, 251, 247, 255),
		Padding:         14,
		CornerRadius:    6,
	},
	"blockquote-accent": {
		BackgroundColor: rl.NewColor(109, 40, 217, 255),
		Padding:         0,
	},
	"blockquote-accent-nested": {
		BackgroundColor: rl.NewColor(139, 92, 246, 255),
		Padding:         0,
	},
	"richtext-footnote-back": {
		TextColor:   rl.NewColor(79, 70, 229, 255),
		FontSize:    16,
		Bold:        true,
		PreviewFont: true,
	},
	"preview-footnote-back-icon": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(79, 70, 229, 255),
		Padding:         0,
		BorderWidth:     0,
	},
	"preview-task-icon": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(55, 58, 78, 255),
		Padding:         0,
		BorderWidth:     0,
	},
	"preview-hr-wrap": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		Padding:         10,
	},
	"preview-footnote-shell": {
		BackgroundColor: rl.NewColor(248, 248, 252, 255),
		Padding:         12,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(220, 222, 232, 255),
		CornerRadius:    8,
	},
	"preview-scroll": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		Padding:         0,
	},
	"preview-lane": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		Padding:         12,
	},
	"richtext-footnote": {
		TextColor:   rl.NewColor(90, 92, 108, 255),
		FontSize:    13,
		PreviewFont: true,
	},
	"richtext-footnote-ref": {
		TextColor:   rl.NewColor(79, 70, 229, 255),
		FontSize:    13,
		Bold:        true,
		PreviewFont: true,
	},
	"preview-image-wrap": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(220, 222, 230, 255),
		CornerRadius:    6,
	},
	"preview-image-frame": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     1,
		BorderColor:     rl.NewColor(220, 222, 230, 255),
		CornerRadius:    6,
		Padding:         8,
	},
	"preview-image-caption": {
		TextColor:   rl.NewColor(120, 124, 142, 255),
		FontSize:    14,
		PreviewFont: true,
	},
	"table-header-row": {
		BackgroundColor: rl.NewColor(238, 239, 243, 255),
		TextColor:       rl.NewColor(30, 30, 30, 255),
		Padding:         10,
	},
	"table-body-row": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(40, 42, 54, 255),
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(235, 236, 240, 255),
	},
	"richtext-link": {
		TextColor: rl.NewColor(79, 70, 229, 255),
		Bold:      true,
	},

	// ── DataTable ─────────────────────────────────────────────────────────────

	// datatable — outer table frame.
	"datatable": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(31, 41, 55, 255),
		FontSize:        15,
		Padding:         0,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 213, 228, 255),
		CornerRadius:    6,
	},
	// datatable-flat — edge-to-edge table without outer chrome (Foundry plan diff).
	"datatable-flat": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(31, 41, 55, 255),
		FontSize:        15,
		Padding:         0,
		BorderWidth:     0,
		CornerRadius:    0,
	},
	// foundry-detail-sep — hairline between plan table and detail strip.
	"foundry-detail-sep": {
		BackgroundColor: rl.NewColor(218, 220, 232, 255),
		BorderWidth:     0,
		Padding:         0,
	},
	// foundry-graph-canvas — pan/zoom graph background.
	"foundry-graph-canvas": {
		BackgroundColor: rl.NewColor(250, 251, 253, 255),
		BorderWidth:     0,
		Padding:         0,
	},
	// foundry-plan-chrome — inset for plan diff toolbar / summary / footer.
	"foundry-plan-chrome": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderWidth:     0,
		Padding:         16,
	},
	// foundry-detail — attribute diff readout under the plan table.
	"foundry-detail": {
		BackgroundColor: rl.NewColor(252, 253, 255, 255),
		BorderWidth:     0,
		Padding:         12,
		FontSize:        19,
	},
	// foundry-inspector — resource inspector card (light mode).
	"foundry-inspector": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 45, 255),
		Padding:         12,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(210, 214, 224, 255),
		CornerRadius:    10,
	},
	// foundry-inspector-shell — full inspector column (splitter provides the left edge).
	"foundry-inspector-shell": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 45, 255),
		Padding:         0,
		BorderWidth:     0,
	},
	// foundry-inspector-header — pinned title band.
	"foundry-inspector-header": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 45, 255),
		Padding:         12,
		BorderWidth:     0,
	},
	// foundry-inspector-body — scrollable inspector content.
	"foundry-inspector-body": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 45, 255),
		Padding:         12,
		BorderWidth:     0,
	},
	// foundry-inspector-muted — secondary labels in the inspector.
	"foundry-inspector-muted": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(96, 102, 124, 255),
		FontSize:        13,
	},
	// foundry-inspector-code — attribute diff block (mono, light).
	"foundry-inspector-code": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		TextColor:       rl.NewColor(40, 44, 58, 255),
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(218, 222, 232, 255),
		CornerRadius:    6,
		FontSize:        14,
	},
	// foundry-inspector-sensitive-block — attributes with redacted values.
	"foundry-inspector-sensitive-block": {
		BackgroundColor: rl.NewColor(255, 251, 235, 255),
		TextColor:       rl.NewColor(40, 44, 58, 255),
		Padding:         10,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(252, 211, 77, 255),
		CornerRadius:    6,
		FontSize:        14,
	},
	// foundry-inspector-sensitive — redaction note under attributes.
	"foundry-inspector-sensitive": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(180, 120, 20, 255),
		FontSize:        12,
	},
	// foundry-inspector-footer — pinned action row (JSON / copy).
	"foundry-inspector-footer": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(30, 32, 45, 255),
		Padding:         12,
		BorderWidth:     0,
	},
	// foundry-graph-toolbar — compact controls row above the canvas.
	"foundry-graph-toolbar": {
		BackgroundColor: rl.NewColor(252, 253, 255, 255),
		TextColor:       rl.NewColor(30, 32, 45, 255),
		Padding:         10,
		BorderWidth:     0,
	},
	// foundry-graph-footer — legend under the canvas.
	"foundry-graph-footer": {
		BackgroundColor: rl.NewColor(252, 253, 255, 255),
		TextColor:       rl.NewColor(30, 32, 45, 255),
		Padding:         8,
		BorderWidth:     0,
	},

	// ── Toast ─────────────────────────────────────────────────────────────────
	// toast styles are baked into toast.go (palette per level); these entries
	// are provided for completeness and future theming use only.

	// toast-info — blue notification card.
	"toast-info": {
		BackgroundColor: rl.NewColor(239, 246, 255, 255),
		TextColor:       rl.NewColor(31, 41, 55, 255),
		BorderColor:     rl.NewColor(186, 212, 253, 255),
		BorderWidth:     1.5,
		CornerRadius:    8,
	},
	// toast-success — green notification card.
	"toast-success": {
		BackgroundColor: rl.NewColor(240, 253, 244, 255),
		TextColor:       rl.NewColor(31, 41, 55, 255),
		BorderColor:     rl.NewColor(134, 239, 172, 255),
		BorderWidth:     1.5,
		CornerRadius:    8,
	},
	// toast-warning — amber notification card.
	"toast-warning": {
		BackgroundColor: rl.NewColor(255, 251, 235, 255),
		TextColor:       rl.NewColor(31, 41, 55, 255),
		BorderColor:     rl.NewColor(252, 211, 77, 255),
		BorderWidth:     1.5,
		CornerRadius:    8,
	},
	// toast-error — red notification card.
	"toast-error": {
		BackgroundColor: rl.NewColor(254, 242, 242, 255),
		TextColor:       rl.NewColor(31, 41, 55, 255),
		BorderColor:     rl.NewColor(252, 165, 165, 255),
		BorderWidth:     1.5,
		CornerRadius:    8,
	},

	// ── SplitView ─────────────────────────────────────────────────────────────

	// splitview — background of the entire SplitView area (drawn under the panes).
	"splitview": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0), // transparent; panes draw themselves
		BorderWidth:     0,
		Padding:         0,
	},
	// splitview-splitter — 1px divider line; hover/drag glow is drawn in code.
	"splitview-splitter": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		BorderColor:     rl.NewColor(200, 202, 215, 255),
		BorderWidth:     0,
		CornerRadius:    0,
	},

	// ── Part 4 widgets ────────────────────────────────────────────────────────

	"gauge": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        14,
	},
	"chart": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		BorderColor:     rl.NewColor(220, 224, 240, 255),
		BorderWidth:     1,
		TextColor:       rl.NewColor(30, 32, 50, 255),
		FontSize:        14,
	},
	"filedropzone": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		BorderColor:     rl.NewColor(200, 204, 220, 255),
		TextColor:       rl.NewColor(60, 64, 84, 255),
		FontSize:        14,
	},
	"propertytable": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		BorderColor:     rl.NewColor(220, 224, 240, 255),
		BorderWidth:     1,
	},
	"videoplaceholder": {
		BackgroundColor: rl.NewColor(30, 32, 48, 255),
		TextColor:       rl.NewColor(220, 224, 240, 255),
		FontSize:        12,
	},
	"mapplaceholder": {
		BackgroundColor: rl.NewColor(220, 235, 220, 255),
		BorderColor:     rl.NewColor(180, 200, 180, 255),
		BorderWidth:     1,
	},

	// ── Batch 9 Widget Styles — Toolbar ───────────────────────────────────────

	// toolbar — overall toolbar container background with a subtle bottom border.
	"toolbar": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		BorderColor:     rl.NewColor(210, 212, 222, 255),
		BorderWidth:     1,
		CornerRadius:    6,
		Padding:         8,
	},

	// toolbar-separator — colour of in-row vertical separator lines (panel-adjacent neutrals, polish §4.2).
	"toolbar-separator": {
		BackgroundColor: rl.NewColor(210, 213, 228, 255),
	},

	// toolbar-ribbon-tab — inactive ribbon tab.
	// Reuses the same visual language as the TabView tabs.
	"toolbar-ribbon-tab": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(73, 80, 87, 255),
		BorderColor:     rl.NewColor(222, 226, 230, 255),
		FontSize:        16,
		Padding:         8,
		CornerRadius:    4,
	},
	"toolbar-ribbon-tab-active": {
		// Same strip as inactive tabs; active state is underline + indigo text only.
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(79, 70, 229, 255),
		BorderColor:     rl.NewColor(222, 226, 230, 255),
		FontSize:        16,
		Bold:            true,
		Padding:         8,
		CornerRadius:    4,
	},
	// toolbar-ribbon — edge-to-edge notepad ribbon shell (no outer border seam).
	"toolbar-ribbon": {
		BackgroundColor: rl.NewColor(248, 249, 252, 255),
		BorderColor:     rl.NewColor(210, 212, 222, 255),
		BorderWidth:     0,
		CornerRadius:    0,
		Padding:         4,
	},
	// toolbar-cell — ribbon stacked cell (icon above caption, flat ghost chrome).
	"toolbar-cell": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(55, 65, 81, 255),
		FontSize:        14,
		Padding:         4,
		CornerRadius:    4,
	},
	// toolbar-btn — flat command-bar control (Bootstrap btn-light).
	"toolbar-btn": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(73, 80, 87, 255),
		FontSize:        18,
		Padding:         8,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(222, 226, 230, 255),
		CornerRadius:    4,
	},
	// toolbar-toggle-label — flat toolbar word toggle (accent when on, no switch chrome).
	"toolbar-toggle-label": {
		BackgroundColor: rl.NewColor(0, 0, 0, 0),
		TextColor:       rl.NewColor(107, 114, 128, 255),
		FontSize:        18,
		Padding:         8,
		CornerRadius:    4,
	},
	// toolbar-menu — dropdown / find face (same readable size as toolbar-btn).
	"toolbar-menu": {
		BackgroundColor: rl.NewColor(255, 255, 255, 255),
		TextColor:       rl.NewColor(73, 80, 87, 255),
		FontSize:        18,
		MaxFontSize:     18,
		FontDensity:     1.0,
		Padding:         6,
		BorderWidth:     1,
		BorderColor:     rl.NewColor(222, 226, 230, 255),
		CornerRadius:    4,
	},
}

// GetThemeStyle returns the style for name from CurrentTheme, falling back to DefaultStyle.
func GetThemeStyle(name string) Style {
	if s, ok := CurrentTheme[name]; ok {
		return s
	}
	return DefaultStyle
}

// DropdownSelectedTextColor is the popup highlight color for the current dropdown value.
func DropdownSelectedTextColor() rl.Color {
	if c := GetThemeStyle("dropdown-selected").TextColor; c.A > 0 {
		return c
	}
	return focusRingIndigo
}

// lerpColor linearly interpolates between two colours (used by Toggle, Button, SplitView).
func lerpColor(a, b rl.Color, t float32) rl.Color {
	lerp := func(x, y uint8, t float32) uint8 {
		return uint8(float32(x) + (float32(y)-float32(x))*t)
	}
	return rl.NewColor(
		lerp(a.R, b.R, t),
		lerp(a.G, b.G, t),
		lerp(a.B, b.B, t),
		lerp(a.A, b.A, t),
	)
}
