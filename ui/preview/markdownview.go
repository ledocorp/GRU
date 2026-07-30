package preview

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/ledocorp/gru/ui"
)

func markdownLayoutDebug() bool {
	return ui.EnvTruthy("GRU_MD_LAYOUT", "GORY_MD_LAYOUT")
}

const (
	markdownBuildBatch        = 6
	markdownBuildFirstBatch   = 4
	markdownAboveFoldEstimate = float32(520)
)

func markdownBlocksNeedFootnotes(blocks []ast.Node) bool {
	for _, b := range blocks {
		if _, ok := b.(*extast.FootnoteList); ok {
			return true
		}
	}
	return false
}

// MarkdownView renders GFM source into a scrollable column of preview nodes.
//
// Goldmark parses source to an AST on the main thread or a worker; this view walks
// that AST and builds Gru widgets (Card, RichText, etc.). That is an extension of
// Goldmark, not a fork — individual block types can use custom renderers when the AST
// is awkward (footnotes list, raw HTML) or when a hand-built widget is clearer.
type MarkdownView struct {
	*ui.Container
	scroll  *ui.Viewport
	lane    *ui.Container
	anchors map[string]ui.Node
	OnLink  func(string)

	buildCtx       *mdBuildCtx
	uiDoc          *ui.Document
	buildGen       int // bumped on reset; async builds ignore stale generations
	pendingScroll  string
	sections       []markdownSection
	sectionIndex   int
	blockInSection int
	globalBlockIdx int
	building       bool
	statusLabel    *ui.RichText
	aboveFoldDone  bool
	lastLayoutW    float32
	layoutSyncedGen int // last buildGen fully laid out at current bounds
	settleFrames   int  // force remeasure for a few frames after build (resize nudge)
}

// NewMarkdownView creates a scroll host for markdown preview blocks.
func NewMarkdownView(id string, x, y, w, h float32) *MarkdownView {
	root := ui.NewContainer(id, x, y, w, h)
	root.LayoutType = ui.LayoutFlex
	root.FlexDirection = ui.FlexColumn
	root.SetStyle("transparent")
	root.SetFlexGrow(1)

	scroll := ui.NewViewport(id+"-scroll", 0, 0, 0, 0)
	scroll.SetStyle("preview-scroll")
	scroll.ContentClipBleed = 0 // avoid sibling card fills bleeding over images while scrolling
	scroll.SetFlexGrow(1)
	scroll.Gap = 0
	root.AddChild(scroll)

	return &MarkdownView{Container: root, scroll: scroll}
}

// Layout reflows preview RichText when pane width or markdown content changes.
func (m *MarkdownView) Layout() {
	if m == nil || m.Container == nil {
		return
	}
	w := m.Bounds().Width
	settling := m.settleFrames > 0 && w > 1
	firstRealWidth := m.lastLayoutW <= 1 && w > 1
	widthChanged := firstRealWidth || (m.lastLayoutW > 1 && (w > m.lastLayoutW+0.5 || w < m.lastLayoutW-0.5))
	contentChanged := m.lane != nil && m.layoutSyncedGen != m.buildGen
	needsReflow := (widthChanged || contentChanged || settling) && m.lane != nil && w > 1
	if needsReflow {
		m.lane.InvalidateLayoutPassCache()
		m.invalidatePreviewTextMeasure(m.lane)
		m.invalidatePreviewScrollHosts(m.lane)
		m.lane.MarkDirty()
		if markdownLayoutDebug() {
			println(fmt.Sprintf("[md-layout] reflow w=%.1f lastW=%.1f settle=%d synced=%d gen=%d width=%v content=%v",
				w, m.lastLayoutW, m.settleFrames, m.layoutSyncedGen, m.buildGen, widthChanged, contentChanged))
		}
	}
	m.Container.Layout()
	if needsReflow {
		// One follow-up pass after parent assigned lane width (resize nudge).
		// Avoid a thrashing double-invalidate loop — that broke LaTeX card sizing.
		m.invalidatePreviewTextMeasure(m.lane)
		m.invalidatePreviewScrollHosts(m.lane)
		if m.lane != nil {
			m.lane.InvalidateLayoutPassCache()
			m.lane.MarkDirty()
			m.lane.Layout()
		}
		if m.scroll != nil {
			m.scroll.MarkDirty()
			m.scroll.Layout()
		}
		if m.settleFrames > 0 {
			m.settleFrames--
		}
		if m.settleFrames == 0 {
			m.layoutSyncedGen = m.buildGen
		}
	}
	if w > 1 {
		m.lastLayoutW = w
	}
}

func (m *MarkdownView) invalidatePreviewTextMeasure(n ui.Node) {
	if rt, ok := n.(*ui.RichText); ok {
		rt.InvalidateAutoHeightMeasure()
	}
	for _, ch := range n.Children() {
		m.invalidatePreviewTextMeasure(ch)
	}
}

func (m *MarkdownView) invalidatePreviewScrollHosts(n ui.Node) {
	if vp, ok := n.(*ui.Viewport); ok && vp.Orientation == ui.ScrollHorizontal {
		vp.InvalidateLayoutPassCache()
		vp.MarkDirty()
	}
	for _, ch := range n.Children() {
		m.invalidatePreviewScrollHosts(ch)
	}
}

// SetDocument enables async LaTeX rendering inside markdown math blocks.
func (m *MarkdownView) SetDocument(doc *ui.Document) {
	m.uiDoc = doc
	if m.buildCtx != nil {
		m.buildCtx.doc = doc
		m.refreshMathEquations(m.lane)
	}
}

func (m *MarkdownView) refreshMathEquations(root ui.Node) {
	if root == nil {
		return
	}
	if eq, ok := root.(*MathEquation); ok {
		eq.SetDocument(m.uiDoc)
	}
	for _, ch := range root.Children() {
		m.refreshMathEquations(ch)
	}
}

// SetMarkdown parses source and replaces preview content (incremental build).
func (m *MarkdownView) SetMarkdown(source string) {
	if m == nil || m.scroll == nil {
		return
	}
	m.resetBuildState()
	blocks, src := ParseMarkdownBlocksCached(source)
	m.startIncrementalBuild(blocks, src)
}

// SetMarkdownAsync parses on a worker, then builds on the main thread via doc.QueueMain.
func (m *MarkdownView) SetMarkdownAsync(doc *ui.Document, source string) {
	if m == nil || m.scroll == nil || doc == nil {
		m.SetMarkdown(source)
		return
	}
	m.uiDoc = doc
	m.resetBuildState()
	gen := m.buildGen

	ui.SubmitAsyncBg(func() {
		// Fresh parse (no AST cache): live editor buffers change often and must not
		// reuse blocks warmed for the markdown showcase fixture in another scene.
		blocks, src := ParseMarkdownBlocks(source)
		doc.QueueMain(func() {
			if gen != m.buildGen {
				return
			}
			m.startIncrementalBuild(blocks, src)
		})
	})
}

func (m *MarkdownView) resetBuildState() {
	m.buildGen++
	for _, ch := range m.scroll.Children() {
		m.scroll.RemoveChild(ch.ID())
	}
	m.anchors = make(map[string]ui.Node)
	m.lane = nil
	m.building = false
	m.sections = nil
	m.buildCtx = nil
	m.sectionIndex = 0
	m.blockInSection = 0
	m.globalBlockIdx = 0
	m.aboveFoldDone = false
	m.statusLabel = nil
	m.pendingScroll = ""
	m.lastLayoutW = 0
	m.layoutSyncedGen = 0
	m.settleFrames = 0
}

// invalidatePreviewLayout marks the preview lane stale after content rebuilds
// (heading level changes, list edits) so the next layout pass remeasures widths
// and heights without waiting for a window resize.
func (m *MarkdownView) invalidatePreviewLayout() {
	if m.lane == nil {
		return
	}
	m.lane.InvalidateLayoutPassCache()
	m.invalidatePreviewTextMeasure(m.lane)
	m.lane.MarkDirty()
	if m.scroll != nil {
		m.scroll.MarkDirty()
	}
	m.MarkDirty()
}

// ensurePreviewLayoutAfterBuild marks preview stale after a content rebuild.
// Layout runs from the normal parent chain once bounds are assigned — do not
// layout the lane here (zero-width measure freezes headings until resize).
func (m *MarkdownView) ensurePreviewLayoutAfterBuild() {
	m.layoutSyncedGen = 0
	m.invalidatePreviewLayout()
	m.MarkDrawDirty()
}

// RelayoutAfterContentChange remeasures preview blocks after editor-driven rebuilds.
func (m *MarkdownView) RelayoutAfterContentChange() {
	if m == nil || m.Container == nil {
		return
	}
	m.ensurePreviewLayoutAfterBuild()
	if m.settleFrames < 1 {
		m.settleFrames = 4
	}
	if w := m.Bounds().Width; w > 1 {
		m.lastLayoutW = w - 0.26
	}
	m.Layout()
}

// SimulateResizeReflow applies the same dirty/remeasure path as a 1px window
// resize without changing the OS window — imperceptible, but unsticks first-paint
// AutoHeight / LaTeX card sizing that previously needed a manual nudge.
// Call once after mount / build complete — not every incremental batch.
func (m *MarkdownView) SimulateResizeReflow() {
	if m == nil || m.Container == nil {
		return
	}
	b := m.Bounds()
	m.ensurePreviewLayoutAfterBuild()
	if m.settleFrames < 3 {
		m.settleFrames = 3
	}
	if b.Width <= 1 {
		m.MarkDirty()
		return
	}
	ui.MarkResizeLayoutDirtySubtree(m)
	ui.InvalidateAutoHeightTextMeasures(m)
	// Bump width one pixel, layout, restore — same code path as Document.Resize.
	bumped := b
	bumped.Width += 1
	m.SetBounds(bumped)
	m.lastLayoutW = 0
	m.MarkDirty()
	m.Layout()
	m.SetBounds(b)
	ui.MarkResizeLayoutDirtySubtree(m)
	ui.InvalidateAutoHeightTextMeasures(m)
	m.lastLayoutW = 0
	m.layoutSyncedGen = 0
	m.MarkDirty()
	m.Layout()
	if markdownLayoutDebug() {
		println(fmt.Sprintf("[md-layout] simulate-resize w=%.1f→%.1f→%.1f settle=%d",
			b.Width, bumped.Width, m.Bounds().Width, m.settleFrames))
	}
}

// settlePreviewLayout continues width-aware remeasure without resetting settleFrames.
func (m *MarkdownView) settlePreviewLayout() {
	if m == nil || m.Container == nil || m.settleFrames < 1 {
		return
	}
	m.MarkDirty()
	if w := m.Bounds().Width; w > 1 {
		m.lastLayoutW = w - 0.26
	}
	m.Layout()
}

func (m *MarkdownView) startIncrementalBuild(blocks []ast.Node, src []byte) {
	// Prefer the caller-supplied parse. Only re-parse when empty (defensive).
	if len(blocks) == 0 && len(src) > 0 {
		blocks, src = ParseMarkdownBlocks(string(src))
	}
	for _, ch := range m.scroll.Children() {
		m.scroll.RemoveChild(ch.ID())
	}
	lane := ui.NewContainer(m.ID()+"-lane", 0, 0, 0, 0)
	lane.SetStyle("preview-lane")
	lane.LayoutType = ui.LayoutFlex
	lane.FlexDirection = ui.FlexColumn
	lane.Gap = 16
	lane.AutoHeight = true
	m.lane = lane

	m.buildCtx = newBuildCtx(m.ID()+"-md", m.anchors, nil)
	m.buildCtx.doc = m.uiDoc
	m.buildCtx.source = src
	if markdownBlocksNeedFootnotes(blocks) {
		gmDoc, _ := parseDocument(string(src))
		m.buildCtx.footnoteIndexToRef = collectFootnoteIndexToRef(gmDoc)
	}
	baseLink := MarkdownLinkHandler(m.scroll, lane, m.anchors)
	m.buildCtx.onLink = func(link string) {
		link = strings.TrimSpace(link)
		if strings.HasPrefix(link, "#") {
			m.ensureFullyBuilt()
			slug := markdownAnchorSlug(strings.TrimPrefix(link, "#"))
			if target := resolvePreviewJumpTarget(m.lane, m.anchors, slug); target != nil {
				m.scrollToNode(target)
			} else {
				m.pendingScroll = slug
			}
			return
		}
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			baseLink(link)
			return
		}
		if m.OnLink != nil {
			m.OnLink(link)
		}
	}

	m.sections = partitionMarkdownBlocks(blocks)
	m.sectionIndex = 0
	m.blockInSection = 0
	m.globalBlockIdx = 0
	m.aboveFoldDone = false

	totalContentBlocks := 0
	for _, sec := range m.sections {
		totalContentBlocks += len(sec.blocks)
	}
	m.building = totalContentBlocks > 0

	if totalContentBlocks > markdownBuildFirstBatch+4 {
		m.statusLabel = ui.NewRichText(m.ID()+"-loading", []ui.TextSpan{
			{Text: "Loading preview…", Variant: "muted"},
		}, 0, 0, 0, 0)
		m.statusLabel.AutoHeight = true
		m.lane.AddChild(m.statusLabel)
	}

	m.scroll.AddChild(lane)
	m.buildAboveFold()
	m.RelayoutAfterContentChange()
	if !m.building {
		m.finishBuild()
		m.SimulateResizeReflow()
	}
	m.scroll.MarkDirty()
	m.MarkDirty()
}

func (m *MarkdownView) buildAboveFold() {
	fold := m.viewportHeightEstimate()
	var built float32
	for m.sectionIndex < len(m.sections) {
		secLen := len(m.sections[m.sectionIndex].blocks)
		n := m.appendSectionBlocks(secLen)
		if n == 0 {
			break
		}
		built += float32(n) * 48
		if m.sectionIndex >= len(m.sections) {
			break
		}
		if m.sectionIndex > 0 && built >= fold {
			break
		}
	}
	m.aboveFoldDone = true
	m.flushTrailingFootnotes()
}

// flushTrailingFootnotes materializes deferred footnote definitions at the document tail.
func (m *MarkdownView) flushTrailingFootnotes() {
	for m.sectionIndex < len(m.sections) {
		sec := m.sections[m.sectionIndex]
		if len(sec.blocks) != 1 {
			break
		}
		if _, ok := sec.blocks[0].(*extast.FootnoteList); !ok {
			break
		}
		m.appendSectionBlocks(len(sec.blocks))
	}
}

func (m *MarkdownView) viewportHeightEstimate() float32 {
	if m.scroll != nil {
		if h := m.scroll.Bounds().Height; h > 80 {
			return h * 1.15
		}
	}
	if m.Container != nil {
		if h := m.Bounds().Height; h > 80 {
			return h * 0.85
		}
	}
	return markdownAboveFoldEstimate
}

func (m *MarkdownView) appendSectionBlocks(n int) int {
	if m.buildCtx == nil || m.lane == nil || n <= 0 {
		return 0
	}
	added := 0
	for added < n && m.sectionIndex < len(m.sections) {
		sec := m.sections[m.sectionIndex]
		for added < n && m.blockInSection < len(sec.blocks) {
			block := sec.blocks[m.blockInSection]
			blockID := fmtID(m.buildCtx.idPrefix, m.globalBlockIdx)
			if fl, ok := block.(*extast.FootnoteList); ok {
				for _, def := range m.buildCtx.buildFootnoteDefs(blockID, fl) {
					if def != nil {
						m.insertContentNode(def)
					}
				}
			} else if node := BuildMarkdownBlock(m.buildCtx, blockID, block); node != nil {
				m.insertContentNode(node)
			}
			m.blockInSection++
			m.globalBlockIdx++
			added++
		}
		if m.blockInSection >= len(sec.blocks) {
			m.sectionIndex++
			m.blockInSection = 0
		}
	}
	if m.sectionIndex >= len(m.sections) {
		m.building = false
		m.finishBuild()
		m.SimulateResizeReflow()
		return added
	}
	m.RelayoutAfterContentChange()
	m.lane.MarkDirty()
	m.scroll.MarkDirty()
	m.MarkDirty()
	return added
}

// insertContentNode appends blocks in document order (before loading label if present).
func (m *MarkdownView) insertContentNode(node ui.Node) {
	if m.statusLabel != nil {
		m.lane.InsertChildBefore(m.statusLabel, node)
		return
	}
	m.lane.AddChild(node)
}

func (m *MarkdownView) finishBuild() {
	if m.statusLabel != nil && m.lane != nil {
		m.lane.RemoveChild(m.statusLabel.ID())
		m.statusLabel = nil
	}
}

// MarkIdleStable finishes any incremental build and clears draw-only dirty flags
// so the UI cache can drop to deep idle with the preview pane open.
func (m *MarkdownView) MarkIdleStable() {
	if m == nil {
		return
	}
	m.ensureFullyBuilt()
	m.building = false
	if m.lane != nil {
		m.layoutSyncedGen = m.buildGen
	}
	if m.Container != nil {
		m.lastLayoutW = m.Bounds().Width
	}
	ui.ClearDrawDirtySubtree(m)
}

// Update implements Node.Update — appends the next batch of blocks while building.
func (m *MarkdownView) Update(dt float32) {
	if m.building {
		if !m.aboveFoldDone {
			m.buildAboveFold()
		} else {
			m.appendSectionBlocks(markdownBuildBatch)
		}
		m.flushTrailingFootnotes()
		m.RelayoutAfterContentChange()
	} else if m.layoutSyncedGen != m.buildGen && m.Bounds().Width > 1 {
		m.RelayoutAfterContentChange()
	} else if m.settleFrames > 0 {
		m.settlePreviewLayout()
	}
	m.Container.Update(dt)
	m.applyPendingScroll()
}

// EnsureFullyBuilt completes any incremental block build synchronously.
// Notepad uses this after Open so the preview never shows a half-built document.
func (m *MarkdownView) EnsureFullyBuilt() {
	m.ensureFullyBuilt()
}

func (m *MarkdownView) ensureFullyBuilt() {
	for m.building {
		m.appendSectionBlocks(9999)
	}
}

func (m *MarkdownView) scrollToNode(target ui.Node) {
	if target == nil || m.scroll == nil || m.lane == nil {
		return
	}
	target.Show()
	m.lane.Layout()
	m.scroll.Layout()
	m.scroll.ScrollToShowNode(target, m.lane)
	m.scroll.MarkDirty()
}

func (m *MarkdownView) applyPendingScroll() {
	if m.pendingScroll == "" || m.scroll == nil || m.lane == nil {
		return
	}
	slug := m.pendingScroll
	m.pendingScroll = ""
	m.scrollToNode(resolvePreviewJumpTarget(m.lane, m.anchors, slug))
}

// Scroll returns the inner viewport (for tests or scroll sync).
func (m *MarkdownView) Scroll() *ui.Viewport { return m.scroll }
