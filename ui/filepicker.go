// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// FileMode selects how [FilePicker] interprets paths and primary actions.
// Use the typed constants [FileModeFile], [FileModeFolder], and [FileModeSave].
type FileMode string

const (
	// FileModeFile picks an existing file. Single-click a file for selection and
	// preview; double-click the same file within ~450ms to confirm (OnOpen).
	// The primary Open button is hidden in this mode as redundant.
	FileModeFile FileMode = "file"
	// FileModeFolder picks a directory. The list shows folders only; use
	// "Select folder" or Open to confirm the directory currently being browsed.
	FileModeFolder FileMode = "folder"
	// FileModeSave picks a directory plus a file name (Save commits the join).
	FileModeSave FileMode = "save"
)

// fpEntry is one row in the file list [DataTable].
type fpEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// FilePicker is a composite browser: folder tree, sortable file table, and
// optional action buttons. It embeds [*Container] so it drops into any flex
// layout or modal body like a normal panel.
//
// # Path signal
//
// [CurrentPath] always reflects the directory being browsed (cleaned with
// [filepath.Clean]). Call [CurrentPath].Set from application code to jump the
// picker to another folder. When the user confirms (double-click file in
// [FileModeFile], Open/Save in other modes, or "Select folder"), the same signal
// is set to the chosen file or folder path, then [OnOpen] runs with that path.
// Use [OnCancel] for the Cancel button.
//
// # Document
//
// Async directory reads and image previews require a [Document] for
// [Document.QueueMain]. Pass a non-nil doc to [NewFilePicker].
//
// The toolbar **Refresh** control re-reads the current folder on the main thread
// and bumps the async load generation so stale worker callbacks are ignored.
//
// # Document preview widgets
//
// Pass previewBody and previewImage to [NewFilePicker] (may be nil). When
// non-nil, async text/image preview updates those widgets; place them in your
// own Panel/Viewport so scrollbars and layout are under your control.
//
// # LLM Prompt Template
//
//	fp := ui.NewFilePicker("open", doc, ui.FileModeFile, homeDir, previewText, previewImg, 0, 0, 0, 480)
//	fp.OnOpen = func(path string) { loadFile(path) }
//	fp.OnCancel = func() { closeModal() }
//	modalBody.AddChild(fp)
//
// Demo scenes: **Batch 3 FilePicker Demo**, Notepad Open dialog.
type FilePicker struct {
	*Container

	doc         *Document
	mode        FileMode
	CurrentPath *Signal[string]
	Selection   *Signal[string] // highlighted file path (empty when none / folder browse)

	OnOpen   func(path string)
	OnCancel func()

	// Internal navigation state (mirrors CurrentPath while browsing).
	browseDir    string
	loadGen      uint64
	loading      atomic.Bool
	lastTreeSel  *TreeNode
	tree         *TreeView
	filesBinding *ListBinding[fpEntry]
	fileTable    *DataTable[fpEntry]

	saveNameInput *TextInput
	saveRow       *Container

	previewBody *TextViewer
	previewImg  *Image

	btnUp      *Button
	btnRefresh *Button
	btnSelFold *Button
	btnOpen    *Button
	btnCancel  *Button

	pathLabel *Label

	// Double-click detection for FileModeFile (binding index + time).
	lastFileClickIdx  int
	lastFileClickTime time.Time
}

// NewFilePicker constructs a FilePicker bound to currentPath for the directory
// being shown and (after confirm) the chosen path. mode selects file, folder,
// or save behaviour. doc must be non-nil so async work can call [Document.QueueMain].
// previewBody and previewImg receive text and image previews; pass nil to disable.
func NewFilePicker(id string, currentPath *Signal[string], mode FileMode, doc *Document, previewBody *TextViewer, previewImg *Image, x, y, w, h float32) *FilePicker {
	if currentPath == nil {
		panic("ui.NewFilePicker: currentPath must not be nil")
	}
	if doc == nil {
		panic("ui.NewFilePicker: doc must not be nil")
	}
	switch mode {
	case FileModeFile, FileModeFolder, FileModeSave:
	default:
		mode = FileModeFile
	}

	base := NewContainer(id, x, y, w, h)
	base.FlexDirection = FlexColumn
	base.Gap = 10
	base.styleName = "filepicker"

	fp := &FilePicker{
		Container:   base,
		doc:         doc,
		mode:        mode,
		CurrentPath: currentPath,
		Selection:   NewSignal(""),
		previewBody: previewBody,
		previewImg:  previewImg,
	}

	// Path label shares the CurrentPath signal for live directory display.
	fp.pathLabel = &Label{
		Element: NewElement(id+"-path", 0, 0, 0, 28),
		Text:    currentPath,
	}
	fp.pathLabel.styleName = "form-value"
	fp.pathLabel.SetFlexGrow(1)
	currentPath.Subscribe(func() { fp.pathLabel.MarkDirty() })

	fp.btnUp = NewButton(id+"-up", "Up", 0, 0, 56, 32)
	fp.btnUp.SetStyle("button")
	fp.btnUp.OnClick = func() { fp.navigateUp() }

	fp.btnRefresh = NewButton(id+"-ref", "Refresh", 0, 0, 88, 32)
	fp.btnRefresh.SetStyle("button")
	fp.btnRefresh.OnClick = func() { fp.reload() }

	toolbar := NewContainer(id+"-toolbar", 0, 0, 0, 36)
	toolbar.FlexDirection = FlexRow
	toolbar.Gap = 8
	toolbar.SetStyle("transparent")
	toolbar.AddChild(fp.pathLabel)
	toolbar.AddChild(fp.btnUp)
	toolbar.AddChild(fp.btnRefresh)

	root := NewTreeNode(id+"-troot", ".")
	fp.tree = NewTreeView(id+"-tree", root, 0, 0, 200, 0)
	fp.tree.ShowRoot = true
	fp.tree.SetFlexGrow(0)
	fp.tree.MinWidth = 180
	fp.tree.MaxWidth = 220

	fp.filesBinding = NewListBinding([]fpEntry{})
	cols := []Column[fpEntry]{
		{Title: "Name", Width: 200, Sortable: true, SortLess: func(a, b fpEntry) bool { return strings.ToLower(a.Name) < strings.ToLower(b.Name) },
			Render: func(e fpEntry) string {
				if e.IsDir {
					return e.Name + string(filepath.Separator)
				}
				return e.Name
			}},
		{Title: "Size", Width: 90, Align: ColumnAlignRight, Sortable: true,
			SortLess: func(a, b fpEntry) bool { return a.Size < b.Size },
			Render: func(e fpEntry) string {
				if e.IsDir {
					return "-"
				}
				return formatFileSize(e.Size)
			}},
		{Title: "Modified", Width: 150, Sortable: true,
			SortLess: func(a, b fpEntry) bool { return a.ModTime.Before(b.ModTime) },
			Render: func(e fpEntry) string {
				if e.ModTime.IsZero() {
					return "-"
				}
				return e.ModTime.Local().Format("2006-01-02 15:04")
			}},
	}
	fp.fileTable = NewDataTable(id+"-table", cols, fp.filesBinding, 0, 0, 0, 0)
	fp.fileTable.SetFlexGrow(1)
	fp.fileTable.OnRowClick = fp.onFileRowClick
	fp.filesBinding.SubscribeSelection(fp.onFileSelectionChange)

	mainRow := NewContainer(id+"-main", 0, 0, 0, 0)
	mainRow.FlexDirection = FlexRow
	mainRow.Gap = 8
	mainRow.SetStyle("transparent")
	mainRow.SetFlexGrow(1)
	mainRow.AutoHeight = false
	mainRow.AddChild(fp.tree)
	mainRow.AddChild(fp.fileTable)

	fp.saveRow = NewContainer(id+"-saverow", 0, 0, 0, 40)
	fp.saveRow.FlexDirection = FlexRow
	fp.saveRow.Gap = 10
	fp.saveRow.SetStyle("transparent")
	nameLbl := NewLabel(id+"-sn-lbl", "File name", 0, 0, 90, 32)
	nameLbl.SetStyle("form-label")
	fp.saveNameInput = NewTextInput(id+"-sn", "document.txt", 0, 0, 0, 32)
	fp.saveNameInput.SetFlexGrow(1)
	fp.saveNameInput.SetStyle("input")
	fp.saveRow.AddChild(nameLbl)
	fp.saveRow.AddChild(fp.saveNameInput)
	if mode != FileModeSave {
		fp.saveRow.Hide()
	}

	fp.btnSelFold = NewButton(id+"-selfold", "Select folder", 0, 0, 140, 36)
	fp.btnSelFold.SetStyle("primary")
	fp.btnSelFold.OnClick = func() { fp.confirmFolderBrowse() }
	if mode != FileModeFolder {
		fp.btnSelFold.Hide()
	}

	fp.btnOpen = NewButton(id+"-open", "Open", 0, 0, 120, 36)
	fp.btnOpen.SetStyle("primary")
	fp.btnOpen.OnClick = func() { fp.confirmOpen() }

	fp.btnCancel = NewButton(id+"-cancel", "Cancel", 0, 0, 100, 36)
	fp.btnCancel.SetStyle("button")
	fp.btnCancel.OnClick = func() {
		if fp.OnCancel != nil {
			fp.OnCancel()
		}
	}

	btnRow := NewContainer(id+"-btnrow", 0, 0, 0, 40)
	btnRow.FlexDirection = FlexRow
	btnRow.Gap = 10
	btnRow.SetStyle("transparent")
	btnRow.AddChild(fp.btnSelFold)
	btnRow.AddChild(fp.btnOpen)
	btnRow.AddChild(fp.btnCancel)
	sp := NewContainer(id+"-btnsp", 0, 0, 0, 1)
	sp.SetFlexGrow(1)
	sp.SetStyle("transparent")
	btnRow.AddChild(sp)

	base.AddChild(toolbar)
	base.AddChild(mainRow)
	base.AddChild(fp.saveRow)
	base.AddChild(btnRow)

	fp.applyModeLabels()
	fp.lastFileClickIdx = -1

	start := strings.TrimSpace(currentPath.Get())
	if start == "" {
		start = "."
	}
	if st, err := os.Stat(start); err == nil && !st.IsDir() {
		start = filepath.Dir(start)
	}
	fp.browseDir = filepath.Clean(start)
	if currentPath.Get() != fp.browseDir {
		currentPath.Set(fp.browseDir)
	}

	currentPath.Subscribe(fp.onExternalPathChange)

	fp.loadDirAsync()
	return fp
}

func (fp *FilePicker) applyModeLabels() {
	switch fp.mode {
	case FileModeSave:
		fp.btnOpen.Text.Set("Save")
		fp.btnOpen.Show()
	case FileModeFolder:
		fp.btnOpen.Text.Set("Open")
		fp.btnOpen.Show()
	default:
		fp.btnOpen.Text.Set("Open")
		fp.btnOpen.Hide()
	}
}

func (fp *FilePicker) onExternalPathChange() {
	if fp == nil || fp.Container == nil {
		return
	}
	v := strings.TrimSpace(fp.CurrentPath.Get())
	if v == "" {
		v = "."
	}
	v = filepath.Clean(v)
	if st, err := os.Stat(v); err == nil && !st.IsDir() {
		v = filepath.Clean(filepath.Dir(v))
	}
	if v == fp.browseDir {
		return
	}
	fp.browseDir = v
	fp.loadDirAsync()
}

// Mode returns the picker mode (file, folder, or save).
func (fp *FilePicker) Mode() FileMode { return fp.mode }

// IsLoading reports whether a directory listing job is in flight.
func (fp *FilePicker) IsLoading() bool {
	if fp == nil {
		return false
	}
	return fp.loading.Load()
}

// IsInteractive implements [Node] — the picker contains interactive children.
func (fp *FilePicker) IsInteractive() bool { return true }

// Update forwards to children after syncing tree-driven navigation.
func (fp *FilePicker) Update(dt float32) {
	if fp.IsHidden() {
		return
	}
	fp.Container.Update(dt)
	fp.syncTreeNav()
}

func (fp *FilePicker) syncTreeNav() {
	sel := fp.tree.Selected.Get()
	if sel == nil {
		return
	}
	if sel == fp.lastTreeSel {
		return
	}
	fp.lastTreeSel = sel
	if p, ok := sel.Data.(string); ok && p != "" {
		if filepath.Clean(p) == filepath.Clean(fp.browseDir) {
			return
		}
		fp.navigateAbs(p)
	}
}

func (fp *FilePicker) onFileRowClick(bindingIdx int) {
	if bindingIdx < 0 || bindingIdx >= fp.filesBinding.Len() {
		return
	}
	ent := fp.filesBinding.GetItem(bindingIdx)
	if ent.IsDir {
		fp.lastFileClickIdx = -1
		fp.navigateAbs(ent.Path)
		return
	}
	if fp.mode != FileModeFile {
		return
	}
	now := time.Now()
	if bindingIdx == fp.lastFileClickIdx && now.Sub(fp.lastFileClickTime) < 450*time.Millisecond {
		fp.lastFileClickIdx = -1
		fp.commitFileSelection(ent.Path)
		return
	}
	fp.lastFileClickIdx = bindingIdx
	fp.lastFileClickTime = now
}

func (fp *FilePicker) onFileSelectionChange() {
	idx := fp.filesBinding.GetSelectedIndex()
	if idx < 0 || idx >= fp.filesBinding.Len() {
		fp.Selection.Set("")
		fp.collapsePreview()
		return
	}
	ent := fp.filesBinding.GetItem(idx)
	if ent.IsDir {
		fp.Selection.Set("")
		fp.collapsePreview()
		return
	}
	if fp.mode == FileModeFolder {
		return
	}
	fp.Selection.Set(ent.Path)
	fp.runPreview(ent.Path)
}

func (fp *FilePicker) navigateUp() {
	parent := filepath.Dir(fp.browseDir)
	if parent == fp.browseDir {
		return
	}
	fp.navigateAbs(parent)
}

// reload re-reads the current directory on the main thread and updates the tree
// and table immediately. It also bumps the async load generation so any
// in-flight worker callback from [loadDirAsync] self-discards (fixes Refresh
// appearing to do nothing when a stale QueueMain result would otherwise win).
func (fp *FilePicker) reload() {
	if fp.doc == nil {
		return
	}
	atomic.AddUint64(&fp.loadGen, 1)

	fp.lastFileClickIdx = -1
	fp.resetPreview()
	fp.Selection.Set("")
	fp.tree.ScrollY = 0

	dir := fp.browseDir
	entries, subs, err := readDirForPicker(dir, fp.mode)
	if err != nil {
		entries, subs = nil, nil
	}
	if entries == nil {
		entries = []fpEntry{}
	}
	if subs == nil {
		subs = []fpEntry{}
	}
	sortFpEntries(entries)
	fp.loading.Store(false)
	fp.filesBinding.SetItems(entries)
	fp.filesBinding.SetSelectedIndex(-1)
	fp.rebuildTree(dir, subs)
	fp.MarkDirty()
}

func (fp *FilePicker) navigateAbs(abs string) {
	abs = filepath.Clean(abs)
	if abs == fp.browseDir {
		return
	}
	st, err := os.Stat(abs)
	if err != nil || !st.IsDir() {
		return
	}
	fp.lastFileClickIdx = -1
	fp.browseDir = abs
	if fp.CurrentPath.Get() != abs {
		fp.CurrentPath.Set(abs)
	}
	fp.Selection.Set("")
	fp.resetPreview()
	fp.loadDirAsync()
}

func (fp *FilePicker) confirmFolderBrowse() {
	if fp.mode != FileModeFolder {
		return
	}
	path := filepath.Clean(fp.browseDir)
	fp.CurrentPath.Set(path)
	if fp.OnOpen != nil {
		fp.OnOpen(path)
	}
}

func (fp *FilePicker) confirmOpen() {
	switch fp.mode {
	case FileModeFolder:
		fp.confirmFolderBrowse()
	case FileModeSave:
		name := strings.TrimSpace(fp.saveNameInput.Text.Get())
		if name == "" {
			return
		}
		path := filepath.Clean(filepath.Join(fp.browseDir, name))
		fp.CurrentPath.Set(path)
		if fp.OnOpen != nil {
			fp.OnOpen(path)
		}
	default: // file — Open button is hidden; use double-click on a row instead.
		sel := strings.TrimSpace(fp.Selection.Get())
		if sel == "" {
			idx := fp.filesBinding.GetSelectedIndex()
			if idx >= 0 && idx < fp.filesBinding.Len() {
				ent := fp.filesBinding.GetItem(idx)
				if !ent.IsDir {
					sel = ent.Path
				}
			}
		}
		if sel == "" {
			return
		}
		fp.commitFileSelection(sel)
	}
}

func (fp *FilePicker) commitFileSelection(filePath string) {
	filePath = filepath.Clean(filePath)
	if st, err := os.Stat(filePath); err != nil || st.IsDir() {
		return
	}
	fp.CurrentPath.Set(filePath)
	if fp.OnOpen != nil {
		fp.OnOpen(filePath)
	}
}

func (fp *FilePicker) loadDirAsync() {
	if fp.doc == nil {
		return
	}
	gen := atomic.AddUint64(&fp.loadGen, 1)
	dir := fp.browseDir
	fp.loading.Store(true)
	fp.MarkDrawDirty()

	ctx := context.Background()
	SubmitAsync(ctx, func() {
		entries, treeSubs, err := readDirForPicker(dir, fp.mode)
		if err != nil {
			entries = nil
			treeSubs = nil
		}
		g := gen
		fp.doc.QueueMain(func() {
			if atomic.LoadUint64(&fp.loadGen) != g {
				return
			}
			fp.loading.Store(false)
			if entries == nil {
				entries = []fpEntry{}
			}
			sortFpEntries(entries)
			fp.filesBinding.SetItems(entries)
			fp.filesBinding.SetSelectedIndex(-1)
			fp.lastFileClickIdx = -1
			fp.rebuildTree(dir, treeSubs)
			fp.MarkDirty()
		})
	})
}

func (fp *FilePicker) rebuildTree(dir string, subs []fpEntry) {
	root := NewTreeNode(fp.ID()+"-root", filepath.Base(dir))
	if root.Label == "" || root.Label == "." {
		root.Label = dir
	}
	root.Data = dir
	for _, sub := range subs {
		if !sub.IsDir {
			continue
		}
		id := sanitizeTreeID(sub.Path)
		n := NewTreeNode(id, sub.Name)
		n.FolderGlyph = true
		n.Data = sub.Path
		root.AddChild(n)
	}
	root.SetExpanded(true)
	fp.tree.Root = root
	fp.tree.dirty = true
	fp.tree.Rebuild()
	fp.tree.Selected.Set(root)
	fp.lastTreeSel = root
}

func sanitizeTreeID(p string) string {
	s := strings.Map(func(r rune) rune {
		if r == filepath.Separator || r == '/' || r == ':' || r == ' ' {
			return '_'
		}
		return r
	}, p)
	if s == "" {
		return "n"
	}
	return s
}

func sortFpEntries(entries []fpEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}

// collapsePreview clears preview widgets when selection is cleared or on navigation.
func (fp *FilePicker) collapsePreview() {
	if fp.previewBody != nil {
		fp.previewBody.ResetScroll()
		fp.previewBody.Text.Set("")
		// Do not Hide previewBody: parent scene controls document panel visibility;
		// hiding here prevented layout/Draw after re-select (blank preview).
	}
	if fp.previewImg != nil {
		fp.previewImg.Hide()
		fp.previewImg.Unload()
		fp.previewImg.failed = false
		fp.previewImg.loaded = false
		fp.previewImg.FilePath = ""
	}
	fp.MarkDirty()
}

// resetPreview clears selection visuals after navigation or reload.
func (fp *FilePicker) resetPreview() {
	fp.collapsePreview()
}

func readDirForPicker(dir string, mode FileMode) (entries []fpEntry, subdirs []fpEntry, err error) {
	de, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	for _, d := range de {
		name := d.Name()
		if name == "." || name == ".." {
			continue
		}
		full := filepath.Join(dir, name)
		info, e := d.Info()
		if e != nil {
			continue
		}
		isDir := info.IsDir()
		if mode == FileModeFolder && !isDir {
			continue
		}
		ent := fpEntry{
			Name:    name,
			Path:    full,
			IsDir:   isDir,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		entries = append(entries, ent)
		if isDir {
			subdirs = append(subdirs, ent)
		}
	}
	return entries, subdirs, nil
}

func (fp *FilePicker) runPreview(path string) {
	if fp.previewBody == nil && fp.previewImg == nil {
		return
	}
	if fp.previewBody != nil {
		fp.previewBody.ResetScroll()
		fp.previewBody.Text.Set("")
	}
	if fp.previewImg != nil {
		fp.previewImg.Hide()
		fp.previewImg.Unload()
		fp.previewImg.failed = false
		fp.previewImg.loaded = false
		fp.previewImg.FilePath = ""
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".hdr", ".webp":
		if fp.previewImg == nil {
			return
		}
		if fp.previewBody != nil {
			fp.previewBody.Hide()
		}
		fp.previewImg.Unload()
		fp.previewImg.failed = false
		fp.previewImg.loaded = false
		fp.previewImg.FilePath = path
		fp.previewImg.Show()
		fp.previewImg.LoadAsync(fp.doc)
		fp.MarkDirty()
		return
	}

	if fp.previewBody == nil {
		return
	}
	fp.previewBody.Show()
	fp.previewBody.Text.Set("Loading...")
	fp.MarkDrawDirty()

	ctx := context.Background()
	SubmitAsync(ctx, func() {
		data, err := os.ReadFile(path)
		if err != nil {
			fp.doc.QueueMain(func() {
				if fp.previewBody != nil {
					fp.previewBody.Show()
					fp.previewBody.Text.Set("Could not read file: " + err.Error())
					fp.previewBody.MarkDirty()
				}
				fp.MarkDirty()
			})
			return
		}
		summary := summarizeFilePreview(data)
		fp.doc.QueueMain(func() {
			if fp.previewBody != nil {
				fp.previewBody.Show()
				fp.previewBody.Text.Set(summary)
				fp.previewBody.MarkDirty()
			}
			fp.MarkDirty()
		})
	})
}

func summarizeFilePreview(data []byte) string {
	const maxChars = 20000
	if len(data) == 0 {
		return "(empty file)"
	}
	// crude binary detection
	for i := 0; i < len(data) && i < 800; i++ {
		if data[i] == 0 {
			return "Binary file (no text preview)"
		}
	}
	s := string(data)
	if len(s) > maxChars {
		s = trimToValidUTF8(s[:maxChars]) + "\n..."
	}
	return s
}

func trimToValidUTF8(s string) string {
	for !utf8.ValidString(s) {
		if len(s) == 0 {
			return ""
		}
		s = s[:len(s)-1]
	}
	return s
}

func formatFileSize(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	v := float64(n) / 1024
	u := 0
	for v >= 1024 && u < len(units)-1 {
		v /= 1024
		u++
	}
	return fmt.Sprintf("%.1f %s", v, units[u])
}
