package examples

import (
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/ledocorp/gru/ui"
)

// GRUPageReloader watches a .gru file and swaps the compiled page body when it changes.
// Go-owned shell (header, BuildContext actions) stays mounted; only the spec body rebuilds.
//
// Reload detection uses fsnotify when available, with per-frame ModTime polling as backup.
// When PreserveControls is true (default), form values and page scroll position survive reload.
type GRUPageReloader struct {
	LogicalPath string
	VP          *ui.Viewport
	Ctx         *ui.BuildContext
	MutateSpec  func(*ui.DocumentSpec)
	AfterBuild  func(ui.Node)

	// PreserveControls keeps ControlSnapshot values and viewport ScrollY across reloads.
	PreserveControls bool

	resolved      string
	lastMod       time.Time
	pendingReload atomic.Bool
	watcher       *fsnotify.Watcher
	stopWatcher   chan struct{}
}

// Compile reads and builds the current .gru file.
func (r *GRUPageReloader) Compile() (ui.Node, error) {
	data, err := ui.ReadGRUFile(r.LogicalPath)
	if err != nil {
		return nil, err
	}
	spec, err := ui.ParseDocumentSpec(data)
	if err != nil {
		return nil, err
	}
	if r.MutateSpec != nil {
		r.MutateSpec(&spec)
	}
	return ui.BuildDocumentSpec(spec, r.Ctx)
}

// MountShell mounts the page viewport shell and records the file mod time.
func (r *GRUPageReloader) MountShell(doc *ui.Document, shellID, hdrTitle, hdrSubtitle string, body ui.Node) *ui.Viewport {
	r.VP = mountGRUPageShell(doc, shellID, hdrTitle, hdrSubtitle, body)
	r.recordModTime()
	r.startWatcher()
	return r.VP
}

// Close stops the fsnotify watcher. Call from Scene.Destroy when leaving a .gru demo.
func (r *GRUPageReloader) Close() {
	if r.stopWatcher != nil {
		close(r.stopWatcher)
		r.stopWatcher = nil
	}
	if r.watcher != nil {
		_ = r.watcher.Close()
		r.watcher = nil
	}
}

// Poll reloads the body when the .gru file changes on disk.
func (r *GRUPageReloader) Poll(doc *ui.Document) {
	if r.VP == nil || r.Ctx == nil || r.LogicalPath == "" {
		return
	}
	if !r.needsReload() {
		return
	}

	preserve := r.PreserveControls
	var snap map[string]any
	var scrollY float32
	if preserve {
		snap = r.Ctx.ControlSnapshot()
		scrollY = r.VP.ScrollY
	}

	body, err := r.Compile()
	if err != nil {
		log.Printf("gru reload %s: %v", r.LogicalPath, err)
		return
	}
	if preserve && len(snap) > 0 {
		ui.ApplyControlSnapshot(body, snap)
	}

	replaceViewportBody(r.VP, body)
	if r.AfterBuild != nil {
		r.AfterBuild(body)
	}
	r.recordModTime()
	r.pendingReload.Store(false)

	ui.InvalidateViewportScrollFastPath(doc.Root)
	doc.ForceFullLayout()

	if preserve {
		r.VP.ScrollY = scrollY
	}
}

func (r *GRUPageReloader) needsReload() bool {
	if r.pendingReload.Load() {
		return true
	}
	changed, err := r.fileChanged()
	return err == nil && changed
}

func (r *GRUPageReloader) recordModTime() {
	path, err := ui.GRUResolvedPath(r.LogicalPath)
	if err != nil {
		return
	}
	r.resolved = path
	if st, err := os.Stat(path); err == nil {
		r.lastMod = st.ModTime()
	}
}

func (r *GRUPageReloader) fileChanged() (bool, error) {
	path := r.resolved
	if path == "" {
		var err error
		path, err = ui.GRUResolvedPath(r.LogicalPath)
		if err != nil {
			return false, err
		}
		r.resolved = path
	}
	st, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if st.ModTime().Equal(r.lastMod) {
		return false, nil
	}
	return true, nil
}

func (r *GRUPageReloader) startWatcher() {
	if r.watcher != nil || r.resolved == "" {
		return
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	dir := filepath.Dir(r.resolved)
	if err := w.Add(dir); err != nil {
		_ = w.Close()
		return
	}
	r.watcher = w
	r.stopWatcher = make(chan struct{})
	go r.watchLoop(w, filepath.Base(r.resolved))
}

func (r *GRUPageReloader) watchLoop(w *fsnotify.Watcher, base string) {
	for {
		select {
		case <-r.stopWatcher:
			return
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			if filepath.Base(ev.Name) != base {
				continue
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
				r.pendingReload.Store(true)
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Printf("gru watcher %s: %v", r.LogicalPath, err)
		}
	}
}

func replaceViewportBody(vp *ui.Viewport, body ui.Node) {
	ch := vp.Children()
	if len(ch) < 2 {
		return
	}
	if root, ok := body.(*ui.Container); ok {
		root.SetFlexGrow(0)
	}
	vp.ReplaceChildAt(1, body)
}
