//go:build !android

// Package examples — native file dialogs for reference apps (Windows, macOS, Linux).
package examples

import (
	"errors"

	"github.com/ledocorp/gru/ui"

	"github.com/ncruces/zenity"
)

var textDocFilters = []zenity.FileFilter{
	{Name: "Text documents (*.txt, *.md)", Patterns: []string{"*.txt", "*.md"}},
	{Name: "All files (*.*)", Patterns: []string{"*.*"}},
}

var planJSONFilters = []zenity.FileFilter{
	{Name: "Terraform plan JSON (*.json)", Patterns: []string{"*.json"}},
	{Name: "All files (*.*)", Patterns: []string{"*.*"}},
}

var pngExportFilters = []zenity.FileFilter{
	{Name: "PNG image (*.png)", Patterns: []string{"*.png"}},
	{Name: "All files (*.*)", Patterns: []string{"*.*"}},
}

var htmlReportFilters = []zenity.FileFilter{
	{Name: "HTML report (*.html)", Patterns: []string{"*.html"}},
	{Name: "All files (*.*)", Patterns: []string{"*.*"}},
}

// PickPlanJSONOpen shows the OS open-file dialog for terraform plan JSON.
func PickPlanJSONOpen() (string, error) {
	ui.CloseContextMenu()
	ui.PresentScreen()
	path, err := zenity.SelectFile(
		zenity.Title("Open Terraform plan JSON"),
		zenity.FileFilters(planJSONFilters),
	)
	if errors.Is(err, zenity.ErrCanceled) {
		return "", ErrFileDialogCancelled
	}
	return path, err
}

// PickTextFileOpen shows the OS open-file dialog for text documents.
func PickTextFileOpen() (string, error) {
	ui.CloseContextMenu()
	ui.PresentScreen()
	path, err := zenity.SelectFile(
		zenity.Title("Open"),
		zenity.FileFilters(textDocFilters),
	)
	if errors.Is(err, zenity.ErrCanceled) {
		return "", ErrFileDialogCancelled
	}
	return path, err
}

// PickTextFileSave shows the OS save-file dialog. defaultName may be empty.
func PickTextFileSave(defaultName string) (string, error) {
	ui.CloseContextMenu()
	opts := []zenity.Option{
		zenity.Title("Save As"),
		zenity.FileFilters(textDocFilters),
	}
	if defaultName != "" {
		opts = append(opts, zenity.Filename(defaultName))
	}
	path, err := zenity.SelectFileSave(opts...)
	if errors.Is(err, zenity.ErrCanceled) {
		return "", ErrFileDialogCancelled
	}
	return path, err
}

// PickHTMLReportSave shows the OS save-file dialog for an HTML plan report.
func PickHTMLReportSave(defaultName string) (string, error) {
	ui.CloseContextMenu()
	opts := []zenity.Option{
		zenity.Title("Export plan report"),
		zenity.FileFilters(htmlReportFilters),
	}
	if defaultName != "" {
		opts = append(opts, zenity.Filename(defaultName))
	}
	path, err := zenity.SelectFileSave(opts...)
	if errors.Is(err, zenity.ErrCanceled) {
		return "", ErrFileDialogCancelled
	}
	return path, err
}

// PickPNGSave shows the OS save-file dialog for a PNG image.
func PickPNGSave(defaultName string) (string, error) {
	ui.CloseContextMenu()
	opts := []zenity.Option{
		zenity.Title("Export graph PNG"),
		zenity.FileFilters(pngExportFilters),
	}
	if defaultName != "" {
		opts = append(opts, zenity.Filename(defaultName))
	}
	path, err := zenity.SelectFileSave(opts...)
	if errors.Is(err, zenity.ErrCanceled) {
		return "", ErrFileDialogCancelled
	}
	return path, err
}
