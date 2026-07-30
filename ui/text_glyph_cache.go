package ui

import (
	"hash/maphash"
	"image"
	"math"
	"sync"
	"unsafe"

	"github.com/fogleman/gg"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
)

// T1.8 — shaped glyph cache: rasterize go-text outlines → dynamic atlas → batched quads.
// Complex-script draw is always on when shaped engine is active; GRU_SHAPED_GLYPH_CACHE=1
// (GORY_SHAPED_GLYPH_CACHE alias) forces Latin onto the outline path (dev experiment only).

const shapedGlyphCacheMaxEntries = 4096

type shapedGlyphCacheKey struct {
	faceSeed uint64
	gid      font.GID
	sizePx   uint16
}

type shapedGlyphCacheEntry struct {
	tex     rl.Texture2D
	src     rl.Rectangle // atlas sub-rect; zero width => use full tex
	owned   bool         // true when tex is per-glyph (atlas full fallback)
	width   float32      // logical px (SSAA-normalized for draw)
	height  float32
	originX float32
	originY float32
}

type shapedGlyphCache struct {
	mu      sync.Mutex
	seed    maphash.Seed
	entries map[shapedGlyphCacheKey]shapedGlyphCacheEntry
	order   []shapedGlyphCacheKey
}

var shapedGlyphs shapedGlyphCache

func shapedGlyphCacheEnabled() bool {
	return EnvOr("GRU_SHAPED_GLYPH_CACHE", "GORY_SHAPED_GLYPH_CACHE") == "1"
}

// ShapedGlyphCacheEnvEnabled reports GRU_SHAPED_GLYPH_CACHE (GORY alias; Latin outline draw — demo only).
func ShapedGlyphCacheEnvEnabled() bool {
	return shapedGlyphCacheEnabled()
}

func initShapedGlyphCache() {
	shapedGlyphs.seed = maphash.MakeSeed()
	shapedGlyphs.entries = make(map[shapedGlyphCacheKey]shapedGlyphCacheEntry)
	shapedGlyphs.order = shapedGlyphs.order[:0]
	initShapedGlyphAtlas()
}

func unloadShapedGlyphCache() {
	shapedGlyphs.mu.Lock()
	defer shapedGlyphs.mu.Unlock()
	for _, e := range shapedGlyphs.entries {
		if e.owned && e.tex.ID != 0 {
			rl.UnloadTexture(e.tex)
		}
	}
	shapedGlyphs.entries = nil
	shapedGlyphs.order = shapedGlyphs.order[:0]
	unloadShapedGlyphAtlas()
}

func faceCacheSeed(face *font.Face) uint64 {
	if face == nil {
		return 0
	}
	var h maphash.Hash
	h.SetSeed(shapedGlyphs.seed)
	// Face pointer is stable for process lifetime; sufficient for cache namespacing.
	p := uintptr(unsafe.Pointer(face))
	h.Write([]byte{byte(p), byte(p >> 8), byte(p >> 16), byte(p >> 24)})
	return h.Sum64()
}

func shapedGlyphCacheLookup(face *font.Face, gid font.GID, sizePx float32) (shapedGlyphCacheEntry, bool) {
	if face == nil {
		return shapedGlyphCacheEntry{}, false
	}
	key := shapedGlyphCacheKey{
		faceSeed: faceCacheSeed(face),
		gid:      gid,
		sizePx:   uint16(sizePx + 0.5),
	}
	shapedGlyphs.mu.Lock()
	defer shapedGlyphs.mu.Unlock()
	e, ok := shapedGlyphs.entries[key]
	return e, ok
}

func shapedGlyphCacheStore(face *font.Face, gid font.GID, sizePx float32, e shapedGlyphCacheEntry) {
	key := shapedGlyphCacheKey{
		faceSeed: faceCacheSeed(face),
		gid:      gid,
		sizePx:   uint16(sizePx + 0.5),
	}
	shapedGlyphs.mu.Lock()
	defer shapedGlyphs.mu.Unlock()
	if _, exists := shapedGlyphs.entries[key]; exists {
		shapedGlyphs.entries[key] = e
		return
	}
	for len(shapedGlyphs.order) >= shapedGlyphCacheMaxEntries {
		old := shapedGlyphs.order[0]
		shapedGlyphs.order = shapedGlyphs.order[1:]
		if evicted, ok := shapedGlyphs.entries[old]; ok {
			if evicted.owned && evicted.tex.ID != 0 {
				rl.UnloadTexture(evicted.tex)
			}
			delete(shapedGlyphs.entries, old)
		}
	}
	shapedGlyphs.entries[key] = e
	shapedGlyphs.order = append(shapedGlyphs.order, key)
}

func shapedGlyphSSAA() float32 {
	s := EffectiveSupersamplingScale()
	if s < 1 {
		return 1
	}
	return s
}

func shapedGlyphRasterPx(fontSize float32) float32 {
	return fontSize * shapedGlyphSSAA() * EffectiveGlyphDisplayScale()
}
func shapedEnsureGlyphTexture(face *font.Face, gid font.GID, fontSize float32) (shapedGlyphCacheEntry, bool) {
	rasterPx := shapedGlyphRasterPx(fontSize)
	if e, ok := shapedGlyphCacheLookup(face, gid, rasterPx); ok {
		return e, true
	}
	e, ok := rasterShapedGlyph(face, gid, fontSize, rasterPx)
	if !ok || e.tex.ID == 0 {
		return shapedGlyphCacheEntry{}, false
	}
	shapedGlyphCacheStore(face, gid, rasterPx, e)
	return e, true
}

func rasterShapedGlyph(face *font.Face, gid font.GID, fontSize, rasterPx float32) (shapedGlyphCacheEntry, bool) {
	if face == nil {
		return shapedGlyphCacheEntry{}, false
	}
	outline, ok := shapedGlyphOutline(face, gid)
	if !ok {
		return shapedGlyphCacheEntry{}, false
	}
	ext, ok := face.GlyphExtents(gid)
	if !ok {
		return shapedGlyphCacheEntry{}, false
	}
	scale, w, h, ox, oy := shapedGlyphRasterLayout(outline, ext, rasterPx, face.Upem())
	gc := gg.NewContext(w, h)
	gc.SetRGBA(0, 0, 0, 0)
	gc.Clear()
	gc.SetRGBA(1, 1, 1, 1)
	drawGlyphOutlineGG(gc, outline, scale, ox, oy)
	ssaa := shapedGlyphSSAA()
	img := gc.Image()
	tex, src, ok := shapedAtlasPackGlyph(img)
	entry := shapedGlyphCacheEntry{
		width:   float32(w) / ssaa,
		height:  float32(h) / ssaa,
		originX: float32(ox) / ssaa,
		originY: float32(oy) / ssaa,
	}
	if ok && tex.ID != 0 {
		entry.tex = tex
		entry.src = src
		entry.owned = false
		return entry, true
	}
	tex = GoImageToTexture(img)
	if tex.ID == 0 {
		return shapedGlyphCacheEntry{}, false
	}
	entry.tex = tex
	entry.src = rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
	entry.owned = true
	return entry, true
}

func shapedGlyphRasterLayout(outline font.GlyphOutline, ext font.GlyphExtents, sizePx float32, upem uint16) (scale float64, w, h int, ox, oy float64) {
	scale = float64(sizePx) / float64(upem)
	minX, minY, maxX, maxY := shapedOutlineBounds(outline)
	if minX > maxX {
		minX, maxX = float64(ext.XBearing), float64(ext.XBearing+ext.Width)
		minY, maxY = float64(ext.YBearing-ext.Height), float64(ext.YBearing)
	}
	const pad = 2.0
	w = int(math.Ceil((maxX-minX)*scale)) + int(pad*2)
	h = int(math.Ceil((maxY-minY)*scale)) + int(pad*2)
	if w < 4 {
		w = 4
	}
	if h < 4 {
		h = 4
	}
	ox = pad - minX*scale
	oy = pad + maxY*scale // font Y-up → image Y-down
	return scale, w, h, ox, oy
}

func shapedOutlineBounds(outline font.GlyphOutline) (minX, minY, maxX, maxY float64) {
	first := true
	for _, seg := range outline.Segments {
		for _, p := range seg.ArgsSlice() {
			x, y := float64(p.X), float64(p.Y)
			if first {
				minX, maxX = x, x
				minY, maxY = y, y
				first = false
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	return minX, minY, maxX, maxY
}

func drawGlyphOutlineGG(gc *gg.Context, outline font.GlyphOutline, scale, offsetX, baselineY float64) {
	var current ot.SegmentPoint
	for _, seg := range outline.Segments {
		args := seg.ArgsSlice()
		switch seg.Op {
		case ot.SegmentOpMoveTo:
			if len(args) < 1 {
				continue
			}
			current = args[0]
			x, y := glyphUnitToPixel(current.X, current.Y, scale, offsetX, baselineY)
			gc.MoveTo(x, y)
		case ot.SegmentOpLineTo:
			if len(args) < 1 {
				continue
			}
			current = args[0]
			x, y := glyphUnitToPixel(current.X, current.Y, scale, offsetX, baselineY)
			gc.LineTo(x, y)
		case ot.SegmentOpQuadTo:
			if len(args) < 2 {
				continue
			}
			cx, cy := glyphUnitToPixel(args[0].X, args[0].Y, scale, offsetX, baselineY)
			x, y := glyphUnitToPixel(args[1].X, args[1].Y, scale, offsetX, baselineY)
			gc.QuadraticTo(cx, cy, x, y)
			current = args[1]
		case ot.SegmentOpCubeTo:
			if len(args) < 3 {
				continue
			}
			c1x, c1y := glyphUnitToPixel(args[0].X, args[0].Y, scale, offsetX, baselineY)
			c2x, c2y := glyphUnitToPixel(args[1].X, args[1].Y, scale, offsetX, baselineY)
			x, y := glyphUnitToPixel(args[2].X, args[2].Y, scale, offsetX, baselineY)
			gc.CubicTo(c1x, c1y, c2x, c2y, x, y)
			current = args[2]
		}
	}
	gc.Fill()
}

func glyphUnitToPixel(x, y float32, scale, offsetX, baselineY float64) (float64, float64) {
	return float64(x)*scale + offsetX, baselineY - float64(y)*scale
}

// shapedGlyphPlacement holds one glyph draw quad in line-local coordinates.
type shapedGlyphPlacement struct {
	left, top float32
	entry     shapedGlyphCacheEntry
}

func shapedRunGlyphPlacements(run shapedTextRun, fontSize float32) ([]shapedGlyphPlacement, float32, float32, bool) {
	if run.out.Face == nil || len(run.out.Glyphs) == 0 {
		return nil, 0, 0, false
	}
	face := run.out.Face
	rtl := run.out.Direction == di.DirectionRTL
	ascent := float32(run.out.LineBounds.Ascent.Ceil())
	penX := float32(0)
	if rtl {
		penX = float32(run.out.Advance.Ceil())
	}
	left := float32(1e9)
	right := float32(-1e9)
	var out []shapedGlyphPlacement
	for _, g := range run.out.Glyphs {
		adv := float32(g.XAdvance.Ceil())
		if adv == 0 {
			adv = float32(g.Advance.Ceil())
		}
		if rtl {
			penX -= adv
		}
		gx := penX + float32(g.XOffset.Ceil())
		gy := ascent - float32(g.YOffset.Ceil())
		entry, ok := shapedEnsureGlyphTexture(face, g.GlyphID, fontSize)
		if !ok {
			if !rtl {
				penX += adv
			}
			continue
		}
		x0 := gx - entry.originX
		x1 := x0 + entry.width
		if x0 < left {
			left = x0
		}
		if x1 > right {
			right = x1
		}
		out = append(out, shapedGlyphPlacement{
			left:  x0,
			top:   gy - entry.originY,
			entry: entry,
		})
		if !rtl {
			penX += adv
		}
	}
	if len(out) == 0 {
		return nil, 0, 0, false
	}
	if right <= left {
		right = float32(run.out.Advance.Ceil())
		left = 0
	}
	return out, left, right, true
}

// shapedRunVisualWidth returns ink width including glyph side-bearings (complex script).
// Uses font metrics only — no GPU raster (safe for measure/layout).
func shapedRunVisualWidth(run shapedTextRun, fontSize float32) (float32, bool) {
	_, left, right, ok := shapedRunInkExtents(run, fontSize)
	if !ok {
		return 0, false
	}
	return right - left, true
}

func shapedRunInkExtents(run shapedTextRun, fontSize float32) (placements []shapedGlyphPlacement, left, right float32, ok bool) {
	if run.out.Face == nil || len(run.out.Glyphs) == 0 {
		return nil, 0, 0, false
	}
	face := run.out.Face
	rtl := run.out.Direction == di.DirectionRTL
	ascent := float32(run.out.LineBounds.Ascent.Ceil())
	scale := fontSize / float32(face.Upem())
	penX := float32(0)
	if rtl {
		penX = float32(run.out.Advance.Ceil())
	}
	left = float32(1e9)
	right = float32(-1e9)
	for _, g := range run.out.Glyphs {
		adv := float32(g.XAdvance.Ceil())
		if adv == 0 {
			adv = float32(g.Advance.Ceil())
		}
		if rtl {
			penX -= adv
		}
		gx := penX + float32(g.XOffset.Ceil())
		gy := ascent - float32(g.YOffset.Ceil())
		gw, gh, ox, oy, gok := shapedGlyphLogicalBounds(face, g.GlyphID, scale)
		if !gok {
			if !rtl {
				penX += adv
			}
			continue
		}
		x0 := gx - ox
		x1 := x0 + gw
		if x0 < left {
			left = x0
		}
		if x1 > right {
			right = x1
		}
		_ = gy
		_ = gh
		_ = oy
		if !rtl {
			penX += adv
		}
	}
	if right <= left {
		right = float32(run.out.Advance.Ceil())
		left = 0
	}
	return nil, left, right, true
}

func shapedGlyphLogicalBounds(face *font.Face, gid font.GID, scale float32) (w, h, originX, originY float32, ok bool) {
	outline, ok := shapedGlyphOutline(face, gid)
	if !ok {
		return 0, 0, 0, 0, false
	}
	ext, ok := face.GlyphExtents(gid)
	if !ok {
		return 0, 0, 0, 0, false
	}
	rasterScale := float64(scale) * float64(shapedGlyphSSAA())
	_, bw, bh, boxOx, boxOy := shapedGlyphRasterLayout(outline, ext, float32(rasterScale), face.Upem())
	ssaa := shapedGlyphSSAA()
	return float32(bw) / ssaa, float32(bh) / ssaa, float32(boxOx) / ssaa, float32(boxOy) / ssaa, true
}

func shapedGlyphEntrySrc(entry shapedGlyphCacheEntry) rl.Rectangle {
	if entry.src.Width > 0 && entry.src.Height > 0 {
		return entry.src
	}
	return rl.NewRectangle(0, 0, float32(entry.tex.Width), float32(entry.tex.Height))
}

// shapedDrawRunGlyphCache draws a shaped run using cached outline rasters (T1.8).
// x,y follow raylib DrawTextEx convention: y is the top of the line box, not baseline.
func shapedDrawRunGlyphCache(run shapedTextRun, x, y, fontSize float32, color rl.Color) bool {
	placements, left, _, ok := shapedRunGlyphPlacements(run, fontSize)
	if !ok {
		return false
	}
	shiftX := x - left
	tint := rl.NewColor(color.R, color.G, color.B, color.A)
	for _, p := range placements {
		dst := rl.NewRectangle(p.left+shiftX, y+p.top, p.entry.width, p.entry.height)
		src := shapedGlyphEntrySrc(p.entry)
		rl.DrawTexturePro(p.entry.tex, src, dst, rl.NewVector2(0, 0), 0, tint)
	}
	return true
}

func shapedGlyphOutline(face *font.Face, gid font.GID) (font.GlyphOutline, bool) {
	data := face.GlyphData(gid)
	if data == nil {
		return font.GlyphOutline{}, false
	}
	outline, ok := data.(font.GlyphOutline)
	return outline, ok
}

// rasterShapedGlyphImage is a CPU-only helper for unit tests (no GPU upload).
func rasterShapedGlyphImage(face *font.Face, gid font.GID, fontSize float32) (image.Image, bool) {
	if face == nil {
		return nil, false
	}
	rasterPx := shapedGlyphRasterPx(fontSize)
	outline, ok := shapedGlyphOutline(face, gid)
	if !ok {
		return nil, false
	}
	ext, ok := face.GlyphExtents(gid)
	if !ok {
		return nil, false
	}
	scale, w, h, ox, oy := shapedGlyphRasterLayout(outline, ext, rasterPx, face.Upem())
	gc := gg.NewContext(w, h)
	gc.SetRGBA(0, 0, 0, 0)
	gc.Clear()
	gc.SetRGBA(1, 1, 1, 1)
	drawGlyphOutlineGG(gc, outline, scale, ox, oy)
	return gc.Image(), true
}

// shapedGlyphCacheHasInk reports whether a test raster produced non-transparent pixels.
func shapedGlyphCacheHasInk(img image.Image) bool {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				return true
			}
		}
	}
	return false
}
