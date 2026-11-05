// Public export — assemble staging/export without PowerShell.
//
//	gru export [-o DIR]
//
// Maintainer-only workflow. The public tree must never include scripts/ or
// staging planning docs. See staging/MANIFEST.md (private).
package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func runExport(args []string) error {
	flags := flag.NewFlagSet("export", flag.ContinueOnError)
	outFlag := flags.String("o", "", "output directory (default: <repo>/staging/export)")
	flags.SetOutput(os.Stderr)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("usage: gru export [-o DIR]")
	}

	root, err := findRepoRoot()
	if err != nil {
		return err
	}
	out := *outFlag
	if out == "" {
		out = filepath.Join(root, "staging", "export")
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Export → %s\n", out)
	tmp := out + ".tmp"
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return err
	}
	// Build into tmp, then replace out (avoids Windows locks when cwd is inside out).
	defer func() { _ = os.RemoveAll(tmp) }()

	copyFile := func(rel string) error {
		return copyFileTo(filepath.Join(root, rel), filepath.Join(tmp, rel))
	}
	copyTree := func(rel string) error {
		return copyTreeTo(filepath.Join(root, rel), filepath.Join(tmp, rel))
	}

	for _, f := range []string{"LICENSE", "NOTICE", "go.mod", "go.sum", "VERSION"} {
		if err := copyFile(f); err != nil {
			return err
		}
	}
	for _, t := range []string{
		"ui", "media/imaging", "internal/go-webview2", "internal/appicon",
		"internal/tray", "internal/version", "devtools/studio",
		"cmd/hello", "samples/hello", "cmd/webviewhello", "samples/webviewhello", "cmd/gru",
		// S1 / fx defaults (CGO): emit-C bodies + zspec consumer pack. Without these,
		// public clones fail to link ui/fxglyph + ui/fxmeasure + ui/fxbridge.
		"gen/fxlib",
		"third_party/fx/emit-go-abi",
		// Optional: fx sources for maintainers who regenerate gen/ (not required to build).
		"fxlib",
	} {
		if err := copyTree(t); err != nil {
			return err
		}
	}

	rootGo := []string{
		"main.go", "main_imports_grudemo.go",
		// Do NOT export main_imports_studio.go — blank-imports Foundry/Prism (private).
		"main_instance.go", // !notepad StartupInstance wiring (not main_instance_notepad.go)
		"platform_app.go", "platform_windows.go", "platform_linux.go", "platform_other.go",
		"platform_desktop_title_default.go",
		"process_metrics_windows.go", "process_metrics_other.go",
		"release_mode_dev.go", "text_engine_rollout.go", "env_alias.go",
		"debug.go", // consolidated F5–F11 / env demo diagnostics (was debug_*.go)
		"resource.syso", // Windows version/icon resource — Task Manager FileDescription
	}
	for _, f := range rootGo {
		if err := copyFile(f); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
	}

	demoFiles := []string{
		"shell_desktop_demo.go", "appshell_demo.go", "settings_desktop_demo.go",
		"list_pane_demo.go", "filters_demo.go", "form_demo.go", "card_nest_demo.go",
		"responsive_demo.go", "settings_demo.go", "timeline_demo.go", "counter_demo.go",
		"theme_v2_demo.go", "document_gallery_demo.go", "public_directory_demo.go",
		"batch1_demo.go", "batch2_demo.go", "batch3_demo.go", "batch3b_demo.go",
		"batch3_datepicker_demo.go", "batch4_demo.go", "batch6_demo.go", "batch7_demo.go",
		"batch11_date_range_demo.go", "batch12_rating_demo.go", "batch13_pagination_demo.go",
		"batch14_segmented_demo.go", "batch15_combobox_demo.go", "batch16_spinbox_demo.go",
		"batch17_breadcrumbs_demo.go", "batch18_colorwell_demo.go", "batch19_dropdown_demo.go",
		"batch20_radiogroup_demo.go",
		"batch21_listtile_demo.go", "batch22_toggle_demo.go", "batch23_checkbox_demo.go",
		"batch24_slider_demo.go", "batch25_progressbar_demo.go", "batch26_colorpicker_demo.go",
		"webview_demo.go", "webview_focus_demo.go",
	}
	glueFiles := []string{
		"scene.go", "scene_nav.go", "page_shell.go", "page_grid.go", "batch_caption.go",
		"flex_copy.go", "phosphor_chrome.go", "shell_chrome.go", "list_pane_shell.go",
		"settings_desktop_rows.go", "gru_page.go", "gru_hot_reload.go",
		"public_catalog.go", "public_catalog_smoke_test.go",
		"public_demo_mode.go", "public_demo_mode_grudemo.go",
		"dev_env.go", "dev_env_startup_default.go", "packaging_paths.go",
		"scene_product_titles.go",
		"sample_hosts_grudemo.go", // stubs Studio footer hosts under -tags grudemo
		"startup_instance_stub.go", // !notepad StartupInstance no-op
		"filedialog_common.go", "filedialog_desktop.go", "filedialog_android.go",
		"titlebar_chrome.go", "split_demo_helpers.go",
	}
	for _, f := range append(demoFiles, glueFiles...) {
		rel := filepath.Join("examples", f)
		if err := copyFile(rel); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
	}
	for _, t := range []string{"examples/appinstance", "examples/startuppath"} {
		if err := copyTree(t); err != nil {
			return err
		}
	}

	if err := exportFonts(root, tmp); err != nil {
		return err
	}
	if err := copyTree("assets/web"); err != nil {
		return err
	}
	for _, banned := range []string{"assets/test_video", "assets/test_images"} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(banned))); err == nil {
			return fmt.Errorf("export must not include %s", banned)
		}
	}
	if err := copyFile("pages/gallery.gru"); err != nil {
		return err
	}
	if err := exportPublicDocs(root, tmp); err != nil {
		return err
	}
	if err := guardExport(tmp); err != nil {
		return err
	}

	n := 0
	_ = filepath.WalkDir(tmp, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			n++
		}
		return nil
	})

	_ = os.RemoveAll(out)
	if err := os.Rename(tmp, out); err != nil {
		// Windows may refuse rename across locks — fall back to copy+remove.
		if err := copyTreeTo(tmp, out); err != nil {
			return fmt.Errorf("install export: %w", err)
		}
		_ = os.RemoveAll(tmp)
	}

	fmt.Fprintf(os.Stderr, "Done. Files: %d\n", n)
	fmt.Fprintln(os.Stderr, "Next:")
	fmt.Fprintln(os.Stderr, "  cd staging/export")
	fmt.Fprintln(os.Stderr, "  go build -tags grudemo .")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/hello")
	fmt.Fprintln(os.Stderr, "  go run -tags webview2 ./cmd/webviewhello")
	return nil
}

func exportPublicDocs(root, out string) error {
	docs := filepath.Join(out, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		return err
	}
	if err := copyFileTo(filepath.Join(root, "staging", "PUBLIC_README.md"), filepath.Join(out, "README.md")); err != nil {
		return err
	}
	for _, pair := range [][2]string{
		{"PUBLIC_CONTRIBUTING.md", "CONTRIBUTING.md"},
		{"PUBLIC_SECURITY.md", "SECURITY.md"},
	} {
		src := filepath.Join(root, "staging", pair[0])
		dst := filepath.Join(out, pair[1])
		if err := copyFileTo(src, dst); err != nil {
			fmt.Fprintf(os.Stderr, "warning: missing %s\n", pair[0])
		}
	}
	pairs := [][2]string{
		{"BUILD.md", "BUILD.md"},
		{"GETTING_STARTED.md", "GETTING_STARTED.md"},
		{"CLI.md", "CLI.md"},
		{"WEBVIEW.md", "WEBVIEW.md"},
		{"COMPOSITION.md", "COMPOSITION.md"},
		{"DEMO_INDEX.md", "DEMO_INDEX.md"},
		{"CHEATSHEET.md", "CHEATSHEET.md"},
		{"LAYOUT.md", "LAYOUT.md"},
	}
	for _, p := range pairs {
		src := filepath.Join(root, "staging", p[0])
		dst := filepath.Join(docs, p[1])
		if err := copyFileTo(src, dst); err != nil {
			fmt.Fprintf(os.Stderr, "warning: missing public doc %s\n", p[0])
			continue
		}
	}
	for _, sub := range []string{"architecture", "widgets", "api", "licensing"} {
		src := filepath.Join(root, "staging", "docs", sub)
		dst := filepath.Join(docs, sub)
		if _, err := os.Stat(src); err != nil {
			fmt.Fprintf(os.Stderr, "warning: missing docs subtree staging/docs/%s\n", sub)
			continue
		}
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
		if err := copyTreeTo(src, dst); err != nil {
			return err
		}
	}
	if err := copyFileTo(filepath.Join(root, "staging", "docs", "README.md"), filepath.Join(docs, "README.md")); err != nil {
		fmt.Fprintf(os.Stderr, "warning: missing staging/docs/README.md\n")
	}
	return sanitizePublicREADME(filepath.Join(out, "README.md"))
}

func sanitizePublicREADME(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	t := string(b)
	reDraft := regexp.MustCompile(`(?m)^> This file is the \*\*draft\*\* public GitHub README\..*\r?\n\r?\n`)
	t = reDraft.ReplaceAllString(t, "")
	reMaint := regexp.MustCompile(`(?m)^Maintainer note \(private monorepo\):.*\r?\n`)
	t = reMaint.ReplaceAllString(t, "")
	t = strings.ReplaceAll(t, "staging/PUBLIC_README.md", "README.md")
	return os.WriteFile(path, []byte(t), 0o644)
}

func exportFonts(root, out string) error {
	must := []string{
		"assets/fonts/remixicon.ttf",
		"assets/fonts/remixicon.css",
		"assets/fonts/poppins/Poppins-Regular.ttf",
		"assets/fonts/poppins/Poppins-Bold.ttf",
		"assets/fonts/poppins/Poppins-Italic.ttf",
		"assets/fonts/poppins/Poppins-BoldItalic.ttf",
	}
	for _, rel := range must {
		src := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := copyFileTo(src, filepath.Join(out, filepath.FromSlash(rel))); err != nil {
			return err
		}
	}
	firaCandidates := []string{
		"assets/fonts/firacode/static/FiraCode-Regular.ttf",
		"assets/fonts/firacode/FiraCode-Regular.ttf",
	}
	for _, rel := range firaCandidates {
		src := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(src); err != nil {
			continue
		}
		dst := filepath.Join(out, "assets", "fonts", "firacode", "static", "FiraCode-Regular.ttf")
		if err := copyFileTo(src, dst); err != nil {
			return err
		}
		break
	}
	var interVar string
	_ = filepath.WalkDir(filepath.Join(root, "assets", "fonts"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), "Inter-Variable") && strings.HasSuffix(d.Name(), ".ttf") {
			interVar = path
			return fs.SkipAll
		}
		return nil
	})
	if interVar != "" {
		dst := filepath.Join(out, "assets", "fonts", "Inter", filepath.Base(interVar))
		if err := copyFileTo(interVar, dst); err != nil {
			return err
		}
	} else {
		inter18 := filepath.Join(root, "assets", "fonts", "Inter", "static")
		entries, _ := os.ReadDir(inter18)
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "Inter_18pt-") && strings.HasSuffix(e.Name(), ".ttf") {
				src := filepath.Join(inter18, e.Name())
				dst := filepath.Join(out, "assets", "fonts", "Inter", "static", e.Name())
				if err := copyFileTo(src, dst); err != nil {
					return err
				}
			}
		}
	}
	return filepath.WalkDir(filepath.Join(root, "assets", "fonts"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		lower := strings.ToLower(name)
		if !(name == "OFL.txt" || strings.HasPrefix(name, "LICENSE") || strings.HasPrefix(lower, "readme")) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if !regexp.MustCompile(`(?i)poppins|firacode|Inter|remixicon|dm.?mono`).MatchString(rel) {
			return nil
		}
		return copyFileTo(path, filepath.Join(out, rel))
	})
}

func guardExport(out string) error {
	var hits []string
	bannedExact := map[string]bool{
		filepath.FromSlash("examples/directory_demo.go"): true,
		filepath.FromSlash("examples/catalog.go"):        true,
		"main_imports_studio.go":                         true, // Foundry/Prism blank-imports
	}
	bannedSub := []string{
		"notepad_demo.go", filepath.FromSlash("apps/notepad"), filepath.FromSlash("apps/prism"),
		filepath.FromSlash("foundry") + string(os.PathSeparator),
		filepath.FromSlash("scripts") + string(os.PathSeparator),
		"PUBLIC_DOCS_SHIP.md", "GITHUB_CUT_PLAN.md", "MANIFEST.md",
		"IMPLEMENTATION_PLAN.md", "PLATFORM_SHIP.md", "READINESS.md", "MAINTAINER_LOCAL.md",
		"PUBLIC_TIP_SYNC_PLAN.md", "MASTER_SEPARATION_PLAN.md",
		filepath.FromSlash("cmd/calc") + string(os.PathSeparator),
		filepath.FromSlash("cmd/leandemo") + string(os.PathSeparator),
		filepath.FromSlash("leanpublic") + string(os.PathSeparator),
		filepath.FromSlash("samples/calc") + string(os.PathSeparator),
		filepath.FromSlash("cmd/ghostclock") + string(os.PathSeparator),
		filepath.FromSlash("cmd/ghostisland") + string(os.PathSeparator),
		filepath.FromSlash("cmd/transparentdemo") + string(os.PathSeparator),
	}
	err := filepath.WalkDir(out, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(out, path)
		if err != nil {
			return err
		}
		if bannedExact[rel] {
			hits = append(hits, rel)
		}
		for _, b := range bannedSub {
			if strings.Contains(rel, b) {
				hits = append(hits, rel)
			}
		}
		base := filepath.Base(rel)
		if strings.HasPrefix(base, "notepad_") && !strings.HasSuffix(base, "_stub.go") &&
			strings.Contains(rel, "examples"+string(os.PathSeparator)) {
			hits = append(hits, rel)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(out, "scripts")); err == nil {
		hits = append(hits, "scripts/")
	}
	if len(hits) > 0 {
		return fmt.Errorf("export contains private paths:\n%s", strings.Join(hits, "\n"))
	}
	return nil
}

func resetDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

func copyFileTo(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err == nil {
		_ = os.Chmod(dst, info.Mode())
	}
	return nil
}

func copyTreeTo(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: missing: %s\n", src)
		return nil
	}
	if !info.IsDir() {
		return copyFileTo(src, dst)
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFileTo(path, target)
	})
}
