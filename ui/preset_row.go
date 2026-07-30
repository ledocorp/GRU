// Package ui (continued) — engine-built responsive preset comparison rows.
//
// Authors declare preset names + body copy; the engine builds backdrop, flex
// wrap rows, and uniform Card tiles. See DocumentSpec block type "presetRow".
package ui

import "fmt"

// DefaultPresetRowMinTileWidth is the minimum tile width before wrap/reflow.
const DefaultPresetRowMinTileWidth = float32(200)

// PresetTileSpec is one tile in a presetRow block or BuildPresetRow call.
type PresetTileSpec struct {
	Preset   string         `json:"preset"`
	Title    string         `json:"title"`
	Text     string         `json:"text"`
	Spans    []TextSpan     `json:"spans"`
	Props    map[string]any `json:"props"`
	MinWidth float32        `json:"minWidth"`
}

// PresetRowOptions configures BuildPresetRow layout.
type PresetRowOptions struct {
	Gap          float32
	MinTileWidth float32
	// Columns chunks items into fixed rows (e.g. 3 → 3+2 for five tiles).
	// Zero uses one wrap row for all items.
	Columns      int
	UseBackdrop  bool
	ClipBackdrop bool
}

// DefaultPresetRowOptions returns gallery/showcase defaults.
func DefaultPresetRowOptions() PresetRowOptions {
	return PresetRowOptions{
		Gap:          PresetSurfaceBodyGap,
		MinTileWidth: DefaultPresetRowMinTileWidth,
		UseBackdrop:  true,
		ClipBackdrop: false,
	}
}

// BuildPresetRow compiles a responsive preset comparison strip.
func BuildPresetRow(id string, items []PresetTileSpec, opts PresetRowOptions) (Node, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("preset row %q: no items", id)
	}
	if opts.MinTileWidth <= 0 {
		opts.MinTileWidth = DefaultPresetRowMinTileWidth
	}
	if opts.Gap <= 0 {
		opts.Gap = PresetSurfaceBodyGap
	}

	rowNodes, err := buildPresetRowContainers(id, items, opts)
	if err != nil {
		return nil, err
	}

	body := rowNodes[0]
	if len(rowNodes) > 1 {
		col := NewContainer(id+"-rows", 0, 0, 0, 0)
		col.SetStyle("transparent")
		col.FlexDirection = FlexColumn
		col.Gap = opts.Gap
		col.AutoHeight = true
		for _, row := range rowNodes {
			col.AddChild(row)
		}
		body = col
	}

	if !opts.UseBackdrop {
		return body, nil
	}
	frame := newPresetStripFrame(id+"-strip", opts)
	frame.AddChild(body)
	return frame, nil
}

func buildPresetRowContainers(id string, items []PresetTileSpec, opts PresetRowOptions) ([]Node, error) {
	chunks := chunkPresetTiles(items, opts.Columns)
	rows := make([]Node, 0, len(chunks))
	for i, chunk := range chunks {
		rowID := id + "-row"
		if len(chunks) > 1 {
			rowID = fmt.Sprintf("%s-row-%d", id, i)
		}
		row := NewContainer(rowID, 0, 0, 0, 0)
		row.SetStyle("transparent")
		row.FlexDirection = FlexRow
		row.SetFlexWrap(true)
		row.Gap = opts.Gap
		row.AutoHeight = true
		for j, tile := range chunk {
			tileID := fmt.Sprintf("%s-tile-%d", rowID, j)
			if tile.Preset != "" {
				tileID = id + "-" + sanitizePresetID(tile.Preset)
				if len(chunks) > 1 {
					tileID = fmt.Sprintf("%s-%d", tileID, i)
				}
			}
			card, err := BuildPresetTileCard(tileID, tile, opts.MinTileWidth)
			if err != nil {
				return nil, fmt.Errorf("tile %q: %w", tile.Preset, err)
			}
			row.AddChild(card)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func sanitizePresetID(preset string) string {
	out := make([]byte, 0, len(preset))
	for i := 0; i < len(preset); i++ {
		c := preset[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			out = append(out, c)
		} else if c == '-' || c == '_' {
			out = append(out, c)
		} else {
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "tile"
	}
	return string(out)
}

func chunkPresetTiles(items []PresetTileSpec, columns int) [][]PresetTileSpec {
	if columns <= 0 || columns >= len(items) {
		return [][]PresetTileSpec{items}
	}
	chunks := make([][]PresetTileSpec, 0, (len(items)+columns-1)/columns)
	for i := 0; i < len(items); i += columns {
		end := i + columns
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

// BuildPresetTileCard builds one uniform Card tile for a visual surface preset.
func BuildPresetTileCard(id string, tile PresetTileSpec, defaultMinW float32) (*Card, error) {
	if tile.Preset == "" {
		return nil, fmt.Errorf("preset name required")
	}
	if _, ok := LookupPreset(tile.Preset); !ok {
		return nil, fmt.Errorf("unknown preset %q", tile.Preset)
	}
	title := tile.Title
	if title == "" {
		title = tile.Preset
	}
	card := NewCard(id, title, 0, 0, 0, 0)
	minW := tile.MinWidth
	if minW <= 0 {
		minW = defaultMinW
	}
	card.MinWidth = minW
	if err := card.SetStylePreset(tile.Preset, PresetPropsFromMap(tile.Props)); err != nil {
		return nil, err
	}
	spans := tile.Spans
	if len(spans) == 0 && tile.Text != "" {
		spans = []TextSpan{{Text: tile.Text}}
	}
	if len(spans) > 0 {
		rt := NewRichText(id+"-body", spans, 0, 0, 0, 0)
		card.AddChild(rt)
	}
	return card, nil
}
