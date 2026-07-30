//go:build !notepad

package examples

// Public demo allowlist for build tag "grudemo".
// Canonical list: staging/SCENE_LIST.md — keep titles in sync with Scene.Title().

// PublicDemoStartTitle is the first scene when running with -tags grudemo.
const PublicDemoStartTitle = "Counter Demo"

// PublicDirectorySceneTitle is the grudemo scene picker (footer Scenes button).
const PublicDirectorySceneTitle = "Demo Index"

// publicSceneTitles is the frozen v0 public set (35 scenes incl. Demo Index).
var publicSceneTitles = []string{
	"Demo Index",
	"Desktop Shell (Go)",
	"App Shell (Go)",
	"Settings · Desktop (Go)",
	"List Pane (Go)",
	"Filters (Go)",
	"Form Demo",
	"Card Nest (Go)",
	"Responsive - Breakpoints - Grid",
	"Settings (Go)",
	"Timeline (Go)",
	"Counter Demo",
	"Theme v2 Foundation",
	"Gallery (.gru)",
	"Batch 1 · Tooltip / TabView / Modal",
	"Batch 2 · SearchBar",
	"Batch 3 · Badge",
	"Batch 3b · Accordion",
	"Batch 3 · DatePicker",
	"Batch 4 · Stepper",
	"Batch 6 · DataTable",
	"Batch 7 · Toast / Notification",
	"Batch 11 · DateRangePicker",
	"Batch 12 · Rating",
	"Batch 13 · Pagination",
	"Batch 14 · SegmentedControl",
	"Batch 15 · ComboBox",
	"Batch 16 · SpinBox",
	"Batch 21 · ListTile",
	"Batch 22 · Toggle",
	"Batch 23 · Checkbox",
	"Batch 24 · Slider",
	"Batch 25 · ProgressBar",
	"WebView Module Demo",
	"WebView Focus Handoff",
}

var publicSceneTitleSet map[string]struct{}

func init() {
	publicSceneTitleSet = make(map[string]struct{}, len(publicSceneTitles))
	for _, t := range publicSceneTitles {
		publicSceneTitleSet[t] = struct{}{}
	}
}

// IsPublicDemoTitle reports whether title is on the frozen public allowlist.
func IsPublicDemoTitle(title string) bool {
	_, ok := publicSceneTitleSet[title]
	return ok
}

// PublicSceneTitles returns a copy of the frozen allowlist (stable order).
func PublicSceneTitles() []string {
	out := make([]string, len(publicSceneTitles))
	copy(out, publicSceneTitles)
	return out
}

// FilterPublicFactories keeps only allowlisted scene factories, in allowlist order.
// Factories whose Title() is not on the list are dropped. Missing titles are skipped
// (so a not-yet-registered demo does not break the launcher).
func FilterPublicFactories(all []func() Scene) []func() Scene {
	byTitle := make(map[string]func() Scene, len(all))
	for i, f := range all {
		title := RegistryTitle(i)
		if title == "" && f != nil {
			sc := f()
			title = sc.Title()
			sc.Destroy()
		}
		if title == "" || !IsPublicDemoTitle(title) {
			continue
		}
		byTitle[title] = f
	}
	out := make([]func() Scene, 0, len(publicSceneTitles))
	for _, title := range publicSceneTitles {
		if f, ok := byTitle[title]; ok {
			out = append(out, f)
		}
	}
	return out
}

// IndexOfPublicStart returns the index of PublicDemoStartTitle in factories,
// or 0 if not found.
func IndexOfPublicStart(factories []func() Scene) int {
	for i, f := range factories {
		sc := f()
		title := sc.Title()
		sc.Destroy()
		if title == PublicDemoStartTitle {
			return i
		}
	}
	return 0
}
