// Package examples provides self-contained Gru demo scenes.
//
// Go is the engine; declarative pages compile to the same widgets via
// ui.BuildDocumentSpec (see docs/GO_FIRST_UI.md). Scenes here are the native
// reference for layout patterns that work.
//
// Each demo implements the Scene interface. Register a factory function with
// Register(); main.go calls Build() on the active scene and Destroy() when
// switching away — ensuring textures and signals are cleaned up promptly.
//
// To add a new demo:
//  1. Create a new file in this package.
//  2. Define a struct that embeds BaseScene (or implements Scene directly).
//  3. Implement Title(), Build(*ui.Document), and Destroy().
//  4. Call Register() in an init() function.
//
// Prefer MountAppPage or MountPageGrid (page_shell.go) in Build so the
// tree follows Root (absolute) → page shell → Viewport → Panel → flex.
package examples

import (
	"github.com/ledocorp/gru/ui"
)

// Scene is the lifecycle interface every demo scene must implement.
//
//   - Title   — short name shown in the window title bar and nav strip.
//   - Build   — construct the widget tree into doc; called once on activation.
//   - Destroy — release any GPU resources (textures, render targets) not
//               managed by doc itself; called before the scene is replaced.
//   - OnUpdate — optional per-frame hook; called after doc.Root.Update(dt).
//               Use it for tween ticks, focus management, key shortcuts, etc.
//               Implementations that don't need per-frame work can embed
//               BaseScene, which provides a no-op OnUpdate.
type Scene interface {
	Title() string
	Build(doc *ui.Document)
	Destroy()
	OnUpdate(doc *ui.Document, dt float32)
	// HideDemoNav reports whether the launcher footer (Tab hint + Directory) should
	// be hidden so the scene uses the full client area (reference apps, utilities).
	HideDemoNav() bool
}

// BaseScene is an embeddable struct that provides no-op implementations of
// Destroy and OnUpdate. Embed it in your scene struct to avoid boilerplate
// when you have nothing to clean up or update manually.
type BaseScene struct{}

// Destroy is a no-op. Override in your scene struct if you hold GPU resources.
func (BaseScene) Destroy() {}

// OnUpdate is a no-op. Override when you need per-frame logic (tweens, focus).
func (BaseScene) OnUpdate(_ *ui.Document, _ float32) {}

// HideDemoNav is false by default; override on product-style scenes (e.g. Notepad).
func (BaseScene) HideDemoNav() bool { return false }

// focusClickedTextInput routes document focus to the TextInput under the mouse.
// Widgets like Dropdown handle clicks internally, but TextInput requires
// document focus before it can consume keyboard input.
func focusClickedTextInput(d *ui.Document) {
	focusEditableAt(d)
}

// focusEditableAt routes focus to TextInput, TextEditor, or WebViewPanel under the cursor.
func focusEditableAt(d *ui.Document) {
	ui.RouteScenePointerFocus(d)
}

// ─── Global registry ──────────────────────────────────────────────────────────

type sceneEntry struct {
	title   string
	factory func() Scene
}

// registry holds registered scenes with titles resolved at Register time
// (one lightweight probe + Destroy per scene at package init).
var registry []sceneEntry

// DirectorySceneTitle is the launcher hub scene; main.go starts here when registered.
const DirectorySceneTitle = "Demo Directory"

func registerScene(factory func() Scene, prepend bool) {
	sc := factory()
	entry := sceneEntry{title: sc.Title(), factory: factory}
	sc.Destroy()
	if prepend {
		registry = append([]sceneEntry{entry}, registry...)
	} else {
		registry = append(registry, entry)
	}
}

// Register appends a scene factory to the global list. Call from init()
// in each demo file so main.go doesn't need to import individual scenes.
func Register(factory func() Scene) {
	registerScene(factory, false)
}

// RegisterFirst prepends a scene so it appears first in Tab order (use for the demo directory).
func RegisterFirst(factory func() Scene) {
	registerScene(factory, true)
}

// Registered returns a snapshot of all registered factory functions.
// main.go calls this once after init() functions have run.
func Registered() []func() Scene {
	out := make([]func() Scene, len(registry))
	for i, e := range registry {
		out[i] = e.factory
	}
	return out
}

// RegisteredSceneCount returns the number of registered demo scenes.
func RegisteredSceneCount() int { return len(registry) }

// RegistryTitle returns the title for a registered scene index.
func RegistryTitle(index int) string {
	if index < 0 || index >= len(registry) {
		return ""
	}
	return registry[index].title
}
