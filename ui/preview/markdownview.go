package preview

import (
	"fmt"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/ledocorp/gru/ui"
)

func markdownLayoutDebug() bool {
	return ui.EnvTruthy("GRU_MD_LAYOUT", "GORY_MD_LAYOUT")
}

const (
	markdownBuildBatch        = 8
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
	parsePending   bool // goldmark parse in flight (SetMarkdownAsync)
	statusLabel    *ui.RichText
	aboveFoldDone  bool
	lastLayoutW    float32
	layoutSyncedGen int // last buildGen fully laid out at current bounds
	settleFrames   int  // force remeasure for a few frames after build (resize nudge)
	pendingBuildCompleteReflow bool // SimulateResizeReflow once after incremental build ends
	rebuildHadPrior            bool // true when this build replaces existing preview content (live edit)
	savedScrollY               float32
	savedScrollAtBottom        bool
	pendingScrollRestore       bool
	displayedSource            string // last source committed to the preview lane
}

// IsBuilding reports whether incremental block mount is still in progress.
func (m *MarkdownView) IsBuilding() bool {
	return m != nil && m.building
}

// IsPreviewBusy reports parse or incremental build work in flight.
func (m *MarkdownView) IsPreviewBusy() bool {
	return m != nil && (m.building || m.parsePending)
}

// DisplayedSource returns the markdown source last committed to the preview lane.
func (m *MarkdownView) DisplayedSource() string {
	if m == nil {
		return ""
	}
	return m.displayedSource
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
	// Mid-drag: still run Container.Layout so scroll/lane screen positions track the
	// pane (otherwise scissor clips everything → blank preview). RichText/Label
	// already skip AutoHeight remasure while ChromeWindowMoving.
	if ui.ChromeWindowMoving() {
		m.Container.Layout()
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
		// Settling alone: keep lastMeasuredW so applyFlexAutoHeightLayout can early-out.
		if widthChanged || contentChanged {
			m.invalidatePreviewTextMeasure(m.lane)
		}
		m.invalidatePreviewScrollHosts(m.lane)
		ui.ScheduleLayoutPass(m.lane)
		if markdownLayoutDebug() {
			println(fmt.Sprintf("[md-layout] reflow w=%.1f lastW=%.1f settle=%d synced=%d gen=%d width=%v content=%v",
				w, m.lastLayoutW, m.settleFrames, m.layoutSyncedGen, m.buildGen, widthChanged, contentChanged))
		}
	}
	m.Container.Layout()
	if needsReflow {
		// One follow-up pass after parent assigned lane width (resize nudge).
		// Avoid a thrashing double-invalidate loop — that broke LaTeX card sizing.
		// Use ScheduleLayoutPass (no bubble) — MarkDirty here re-dirtied MarkdownView
		// after Container.Layout cleared it and left notepad-md-preview[layout] forever.
		if widthChanged || contentChanged {
			m.invalidatePreviewTextMeasure(m.lane)
		}
		m.invalidatePreviewScrollHosts(m.lane)
		if m.lane != nil {
			m.lane.InvalidateLayoutPassCache()
			ui.ScheduleLayoutPass(m.lane)
			m.lane.Layout()
		}
		if m.scroll != nil {
			ui.ScheduleLayoutPass(m.scroll)
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
	m.beginBuildGeneration()
	if strings.TrimSpace(source) == "" {
		m.clearPreviewContent()
		return
	}
	blocks, src := ParseMarkdownBlocksCached(source)
	m.startIncrementalBuild(blocks, src)
}

// SetMarkdownAsync parses on a worker, then builds on the main thread via doc.QueueMain.
// Existing preview content stays visible until the replacement lane is ready (no blank flash).
func (m *MarkdownView) SetMarkdownAsync(doc *ui.Document, source string) {
	if m == nil || m.scroll == nil || doc == nil {
		m.SetMarkdown(source)
		return
	}
	m.uiDoc = doc
	m.beginBuildGeneration()
	if strings.TrimSpace(source) == "" {
		m.parsePending = false
		m.clearPreviewContent()
		return
	}
	gen := m.buildGen
	m.parsePending = true

	ui.SubmitAsyncBg(func() {
		blocks, src := ParseMarkdownBlocks(source)
		doc.QueueMain(func() {
			m.parsePending = false
			if gen != m.buildGen {
				return
			}
			m.startIncrementalBuild(blocks, src)
		})
	})
}

func (m *MarkdownView) beginBuildGeneration() {
	m.buildGen++
	m.rebuildHadPrior = len(m.scroll.Children()) > 0
	if m.rebuildHadPrior && m.scroll != nil {
		m.savedScrollY = m.scroll.ScrollY
		max := m.scroll.OverflowScrollY()
		m.savedScrollAtBottom = max > 8 && m.savedScrollY >= max-16
		m.pendingScrollRestore = true
	} else {
		m.savedScrollY = 0
		m.savedScrollAtBottom = false
		m.pendingScrollRestore = false
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
	m.pendingBuildCompleteReflow = false
}

func (m *MarkdownView) clearPreviewContent() {
	if m == nil || m.scroll == nil {
		return
	}
	for _, ch := range m.scroll.Children() {
		m.scroll.RemoveChild(ch.ID())
	}
	m.lane = nil
	m.building = false
	m.displayedSource = ""
	m.sections = nil
	m.buildCtx = nil
	m.rebuildHadPrior = false
	m.scroll.MarkDirty()
	m.MarkDirty()
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
	if m.settleFrames < 2 {
		m.settleFrames = 2
	}
	if w := m.Bounds().Width; w > 1 {
		m.lastLayoutW = w - 0.26
	}
	m.Layout()
}

// NudgeLayoutAfterWidthChange marks preview stale after pane width changes without
// the full SimulateResizeReflow nudge. Used for throttled mid-drag tracking;
// settle still uses SimulateResizeReflow once chrome is idle.
func (m *MarkdownView) NudgeLayoutAfterWidthChange() {
	if m == nil || m.Container == nil {
		return
	}
	m.ensurePreviewLayoutAfterBuild()
	if m.settleFrames < 1 {
		m.settleFrames = 2
	}
	if w := m.Bounds().Width; w > 1 {
		m.lastLayoutW = w - 0.26
	}
	m.MarkDirty()
}

// SimulateResizeReflow applies the same dirty/remeasure path as a 1px window
// resize without changing the OS window — imperceptible, but unsticks first-paint
// and post-edit AutoHeight / LaTeX card sizing that otherwise needs a manual nudge.
// Call once after mount / build complete and after coalesced host resize settle —
// not every incremental batch and never mid-drag (ChromeWindowMoving).
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

func (m *MarkdownView) laneLayoutWidth() float32 {
	if m == nil {
		return 320
	}
	if m.scroll != nil {
		if clip := m.scroll.ClipBounds(); clip.Width > 80 {
			return clip.Width
		}
		if b := m.scroll.Bounds(); b.Width > 80 {
			return b.Width
		}
	}
	if b := m.Bounds(); b.Width > 80 {
		return b.Width
	}
	return 320
}

func (m *MarkdownView) swapScrollLane(lane *ui.Container) {
	if m == nil || m.scroll == nil || lane == nil {
		return
	}
	for _, ch := range m.scroll.Children() {
		m.scroll.RemoveChild(ch.ID())
	}
	m.scroll.AddChild(lane)
}

func (m *MarkdownView) applyPendingScrollRestore() {
	if m == nil || !m.pendingScrollRestore || m.building || m.scroll == nil {
		return
	}
	m.pendingScrollRestore = false
	if m.savedScrollAtBottom {
		m.scroll.ScrollToEnd()
		return
	}
	if m.savedScrollY > 0 {
		m.scroll.RestoreScrollY(m.savedScrollY)
	}
}

func (m *MarkdownView) startIncrementalBuild(blocks []ast.Node, src []byte) {
	// Prefer the caller-supplied parse. Only re-parse when empty (defensive).
	if len(blocks) == 0 && len(src) > 0 {
		blocks, src = ParseMarkdownBlocks(string(src))
	}
	m.displayedSource = string(src)
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
		m.buildCtx.footnoteIndexToRef = collectFootnoteIndexFromBlocks(blocks)
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

	laneW := m.laneLayoutWidth()
	lane.SetBounds(rl.NewRectangle(0, 0, laneW, 0))
	m.buildAboveFold()
	m.swapScrollLane(lane)
	m.RelayoutAfterContentChange()
	if !m.building {
		m.finishBuild()
		m.applyPendingScrollRestore()
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
		// One-shot 1px settle after every completed mount (first paint and live edits).
		// RelayoutAfterContentChange alone can leave cards/RichText one frame stale until
		// a real window resize — dogfood: type → preview refresh → misaligned until nudge.
		m.pendingBuildCompleteReflow = true
		m.pendingScrollRestore = m.rebuildHadPrior
		return added
	}
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

// MarkIdleStable clears draw-only dirty flags when incremental build is done.
// Does not sync-build the full document — that blocked Notepad preview toggles.
func (m *MarkdownView) MarkIdleStable() {
	if m == nil {
		return
	}
	if !m.building {
		if m.lane != nil {
			m.layoutSyncedGen = m.buildGen
		}
		if m.Container != nil {
			m.lastLayoutW = m.Bounds().Width
		}
		ui.ClearDrawDirtySubtree(m)
	}
}

// EnsureAboveFoldBuilt synchronously materializes the first viewport of blocks.
func (m *MarkdownView) EnsureAboveFoldBuilt() {
	if m == nil || !m.building {
		return
	}
	if !m.aboveFoldDone {
		m.buildAboveFold()
	}
	if m.aboveFoldDone && m.statusLabel != nil {
		m.finishBuild()
	}
	m.RelayoutAfterContentChange()
}

// Update implements Node.Update — appends the next batch of blocks while building.
func (m *MarkdownView) Update(dt float32) {
	if ui.ChromeWindowMoving() {
		m.Container.Update(dt)
		return
	}
	// Session 14: pause incremental mount while the editor is actively typing so
	// keystrokes are not competing with RelayoutAfterContentChange every frame.
	if m.building && ui.TypingGestureActive() {
		m.Container.Update(dt)
		m.applyPendingScroll()
		return
	}
	wasBuilding := m.building
	aboveBefore := m.aboveFoldDone
	if m.building {
		if !m.aboveFoldDone {
			m.buildAboveFold()
		} else {
			m.appendSectionBlocks(markdownBuildBatch)
		}
		m.flushTrailingFootnotes()
	} else if m.settleFrames > 0 {
		m.settlePreviewLayout()
	}
	// Full remeasure only when above-fold first lands or build completes — not every batch.
	if wasBuilding && (!aboveBefore && m.aboveFoldDone || !m.building) {
		m.RelayoutAfterContentChange()
	}
	if m.pendingBuildCompleteReflow && !ui.ChromeWindowMoving() && !ui.TypingGestureActive() {
		m.pendingBuildCompleteReflow = false
		m.SimulateResizeReflow()
	}
	m.applyPendingScrollRestore()
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
	if m.pendingBuildCompleteReflow {
		m.pendingBuildCompleteReflow = false
		m.SimulateResizeReflow()
	} else if m.lane != nil && m.layoutSyncedGen != m.buildGen {
		m.RelayoutAfterContentChange()
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
