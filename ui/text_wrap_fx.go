package ui

import (
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/ledocorp/gru/ui/fxmeasure"
)

func fxWrapEnvEnabled() bool {
	if fxOptOut("GRU_FX_WRAP", "GORY_FX_WRAP") {
		return false
	}
	return fxmeasure.Enabled()
}

// EditorWrapPathName is the live numeric wrap kernel: "pod-go" or "pod-fx".
func EditorWrapPathName() string {
	if fxWrapEnvEnabled() && fxmeasure.Enabled() {
		return "pod-fx"
	}
	return "pod-go"
}

// EditorWrapReport is classic prefix-measure vs POD advance wrap (demo + tests).
type EditorWrapReport struct {
	Path           string
	Match          bool
	PODLines       int
	ClassicLines   int
	PODNs          int64
	ClassicNs      int64
	MismatchLine   int // -1 if Match
	PODJoinedOK    bool
	ClassicJoinedOK bool
}

type wrapScratch struct {
	adv []float32
	can []int32
	off []int
	out []int32
}

var wrapScratchPool = sync.Pool{
	New: func() any { return &wrapScratch{} },
}

// wrapAdvancesLines is the shared greedy wrap over POD advances (Go reference).
// outEnds[k] = exclusive end index; returns line count.
func wrapAdvancesLines(advances []float32, canBreak []int32, maxW float32, outEnds []int32) int {
	n := len(advances)
	if len(canBreak) < n || len(outEnds) < 1 || maxW <= 0 {
		return -1
	}
	if n == 0 {
		return 0
	}
	lines := 0
	i := 0
	for i < n {
		lineW := float32(0)
		start := i
		lastBreak := -1
		for i < n {
			adv := advances[i]
			if lineW+adv > maxW {
				if lineW > 0 {
					break
				}
				i++
				if lines >= len(outEnds) {
					return -2
				}
				outEnds[lines] = int32(i)
				lines++
				lastBreak = -1
				lineW = 0
				start = i
				continue
			}
			lineW += adv
			if canBreak[i] != 0 {
				lastBreak = i + 1
			}
			i++
		}
		cut := i
		if i < n && lastBreak > start {
			cut = lastBreak
			i = cut
		}
		if cut <= start {
			cut = start + 1
			i = cut
		}
		if lines >= len(outEnds) {
			return -2
		}
		outEnds[lines] = int32(cut)
		lines++
	}
	return lines
}

func tryWrapAdvancesFX(advances []float32, canBreak []int32, maxW float32, outEnds []int32) (int, bool) {
	if !fxWrapEnvEnabled() || !fxmeasure.Enabled() {
		return 0, false
	}
	n := fxmeasure.WrapAdvances(advances, canBreak, maxW, outEnds)
	if n < 0 {
		return 0, false
	}
	return int(n), true
}

func wrapAdvances(advances []float32, canBreak []int32, maxW float32, outEnds []int32) int {
	if n, ok := tryWrapAdvancesFX(advances, canBreak, maxW, outEnds); ok {
		return n
	}
	return wrapAdvancesLines(advances, canBreak, maxW, outEnds)
}

func joinWrapLines(para string, lines []string) bool {
	n := 0
	for _, ln := range lines {
		n += len(ln)
	}
	if n != len(para) {
		return false
	}
	off := 0
	for _, ln := range lines {
		if para[off:off+len(ln)] != ln {
			return false
		}
		off += len(ln)
	}
	return true
}

func wrapLinesEqual(a, b []string) (bool, int) {
	if len(a) != len(b) {
		return false, 0
	}
	for i := range a {
		if a[i] != b[i] {
			return false, i
		}
	}
	return true, -1
}

// CompareEditorWrap times classic prefix-measure vs POD wrap (loops ≥ 1).
func CompareEditorWrap(para string, maxW float32, style Style, loops int) EditorWrapReport {
	if loops < 1 {
		loops = 1
	}
	rep := EditorWrapReport{Path: EditorWrapPathName(), MismatchLine: -1}
	var pod, classic []string
	t0 := time.Now()
	for i := 0; i < loops; i++ {
		pod = wrapParagraphByAdvances(para, maxW, style)
	}
	rep.PODNs = time.Since(t0).Nanoseconds() / int64(loops)
	t0 = time.Now()
	for i := 0; i < loops; i++ {
		classic = wrapEditorLinesClassic(para, maxW, style)
	}
	rep.ClassicNs = time.Since(t0).Nanoseconds() / int64(loops)
	if pod == nil {
		pod = []string{}
	}
	if classic == nil {
		classic = []string{}
	}
	rep.PODLines = len(pod)
	rep.ClassicLines = len(classic)
	rep.PODJoinedOK = joinWrapLines(para, pod)
	rep.ClassicJoinedOK = joinWrapLines(para, classic)
	ok, idx := wrapLinesEqual(pod, classic)
	rep.Match = ok
	rep.MismatchLine = idx
	return rep
}

// wrapParagraphByAdvances soft-wraps a single paragraph (no '\n') via POD advances.
func wrapParagraphByAdvances(para string, maxW float32, style Style) []string {
	if maxW < 8 {
		if para == "" {
			return nil
		}
		return []string{para}
	}
	if para == "" {
		return []string{""}
	}
	limit := maxW - 4
	sc := wrapScratchPool.Get().(*wrapScratch)
	defer wrapScratchPool.Put(sc)
	sc.adv = sc.adv[:0]
	sc.can = sc.can[:0]
	sc.off = sc.off[:0]
	if shapedTextUsesComplexScript(para) {
		if !fillWrapAdvancesShaped(para, style, sc) {
			return nil
		}
	} else if !fillWrapAdvancesLatin(para, style, sc) {
		return nil
	}
	needOut := len(sc.adv) + 1
	if cap(sc.out) < needOut {
		sc.out = make([]int32, needOut)
	} else {
		sc.out = sc.out[:needOut]
	}
	n := wrapAdvances(sc.adv, sc.can, limit, sc.out)
	if n <= 0 {
		return nil
	}
	return wrapLinesFromOffsets(para, sc.off, sc.out, n)
}

func wrapLinesFromOffsets(para string, offs []int, outEnds []int32, n int) []string {
	lines := make([]string, 0, n)
	start := 0
	for k := 0; k < n; k++ {
		end := int(outEnds[k])
		if end < 0 || end >= len(offs) {
			return nil
		}
		lines = append(lines, para[offs[start]:offs[end]])
		start = end
	}
	return lines
}

func fillWrapAdvancesLatin(para string, style Style, sc *wrapScratch) bool {
	nRunes := utf8.RuneCountInString(para)
	if cap(sc.adv) < nRunes {
		sc.adv = make([]float32, 0, nRunes)
		sc.can = make([]int32, 0, nRunes)
	}
	if cap(sc.off) < nRunes+1 {
		sc.off = make([]int, 0, nRunes+1)
	}
	sc.off = append(sc.off, 0)
	for i := 0; i < len(para); {
		r, size := utf8.DecodeRuneInString(para[i:])
		if size < 1 {
			size = 1
		}
		sc.adv = append(sc.adv, measureWrapRuneAdvance(r, style))
		var brk int32
		if r == ' ' || r == '\t' {
			brk = 1
		}
		sc.can = append(sc.can, brk)
		i += size
		sc.off = append(sc.off, i)
	}
	return len(sc.adv) > 0
}

type shapedWrapCluster struct {
	startRune int
	endRune   int
	advance   float32
}

// fillWrapAdvancesShaped builds POD wrap arrays from HarfBuzz cluster advances (Arabic/Devanagari).
func fillWrapAdvancesShaped(para string, style Style, sc *wrapScratch) bool {
	if shapedNeedsSDFFallbackMeasure(para) || !shapedTextReady() {
		return false
	}
	fs := EffectiveFontSize(style)
	bold := styleDrawBold(style)
	italic, mono, preview := style.Italic, style.Mono, style.PreviewFont
	mono = shapedEffectiveMono(mono, para)
	run, ok := shapedShapeText(para, fs, bold, italic, mono, preview)
	if !ok || len(run.out.Glyphs) == 0 {
		return false
	}
	ratio := shapedMeasureRatioFor(fs, preview, para)

	var clusters []shapedWrapCluster
	var cur *shapedWrapCluster
	for _, g := range run.out.Glyphs {
		startRune := g.TextIndex()
		endRune := startRune + g.RunesCount()
		adv := float32(g.XAdvance.Ceil())
		if adv == 0 {
			adv = float32(g.Advance.Ceil())
		}
		if cur == nil || cur.startRune != startRune {
			if cur != nil {
				clusters = append(clusters, *cur)
			}
			cur = &shapedWrapCluster{startRune: startRune, endRune: endRune, advance: adv}
		} else {
			if endRune > cur.endRune {
				cur.endRune = endRune
			}
			cur.advance += adv
		}
	}
	if cur != nil {
		clusters = append(clusters, *cur)
	}
	if len(clusters) == 0 {
		return false
	}
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].startRune < clusters[j].startRune
	})

	need := len(clusters)
	if cap(sc.adv) < need {
		sc.adv = make([]float32, 0, need)
		sc.can = make([]int32, 0, need)
	}
	if cap(sc.off) < need+1 {
		sc.off = make([]int, 0, need+1)
	}
	sc.off = append(sc.off, 0)
	for _, cl := range clusters {
		sc.adv = append(sc.adv, cl.advance*ratio)
		var brk int32
		if cl.endRune > 0 && cl.endRune <= len(run.runes) {
			last := run.runes[cl.endRune-1]
			if last == ' ' || last == '\t' {
				brk = 1
			}
		}
		sc.can = append(sc.can, brk)
		sc.off = append(sc.off, byteOffsetForRuneIndex(para, cl.endRune))
	}
	return len(sc.adv) > 0 && sc.off[len(sc.off)-1] == len(para)
}
