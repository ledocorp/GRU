//go:build android

// Package examples — Android SAF file dialogs (M4). Async so NativeActivity
// does not ANR while DocumentsUI is open.
package examples

import (
	"errors"
	"time"

	"github.com/ledocorp/gru/ui"
)

func mapFilePickErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ui.ErrFilePickCancelled()) {
		return ErrFileDialogCancelled
	}
	return err
}

// StartTextFileOpen launches ACTION_OPEN_DOCUMENT. onDone may run off the
// game loop — queue scene work with Document.QueueMain.
func StartTextFileOpen(onDone func(path string, err error)) {
	ui.PlatformStartTextFileOpen(func(path string, err error) {
		if onDone != nil {
			onDone(path, mapFilePickErr(err))
		}
	})
}

// StartTextFileSave launches ACTION_CREATE_DOCUMENT with initial content.
func StartTextFileSave(defaultName string, content []byte, onDone func(path string, err error)) {
	if defaultName == "" {
		defaultName = "Untitled.txt"
	}
	ui.PlatformStartTextFileSave(defaultName, content, func(path string, err error) {
		if onDone != nil {
			onDone(path, mapFilePickErr(err))
		}
	})
}

// PickPlanJSONOpen opens any text-like document via SAF (sync; avoid from game loop).
func PickPlanJSONOpen() (string, error) {
	return PickTextFileOpen()
}

// PickTextFileOpen is sync; prefer StartTextFileOpen from the game loop.
func PickTextFileOpen() (string, error) {
	path, err := ui.PlatformPickTextFileOpen()
	return path, mapFilePickErr(err)
}

// PickTextFileSave is sync; prefer StartTextFileSave from the game loop.
func PickTextFileSave(defaultName string) (string, error) {
	return PickTextFileSaveWith(defaultName, nil)
}

// PickTextFileSaveWith is sync; prefer StartTextFileSave from the game loop.
func PickTextFileSaveWith(defaultName string, content []byte) (string, error) {
	if defaultName == "" {
		defaultName = "Untitled.txt"
	}
	path, err := ui.PlatformPickTextFileSave(defaultName, content)
	return path, mapFilePickErr(err)
}

// PickHTMLReportSave uses the same SAF create path as text save.
func PickHTMLReportSave(defaultName string) (string, error) {
	if defaultName == "" {
		defaultName = "report.html"
	}
	return PickTextFileSaveWith(defaultName, nil)
}

// PickPNGSave is not wired to an image encoder on Android yet.
func PickPNGSave(defaultName string) (string, error) {
	_ = defaultName
	ui.ShowToast("PNG export is not available on Android yet", ui.ToastInfo, 3*time.Second)
	return "", ErrFileDialogUnavailable
}
