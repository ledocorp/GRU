//go:build android

// Package examples — Android file dialog stubs (Phase A5 / §9).
//
// Android-only file (//go:build android). Desktop uses filedialog_desktop.go.
// v1: toast stub; SAF later. See docs/ANDROID_CODE.md §3 and ANDROID_PHASE.md.
package examples

import (
	"github.com/ledocorp/gru/ui"
	"time"
)

func androidFileDialogStub(action string) (string, error) {
	ui.ShowToast(action+" is not available on Android yet — use New or paste text",
		ui.ToastInfo, 3*time.Second)
	return "", ErrFileDialogUnavailable
}

// PickPlanJSONOpen is a stub on Android until SAF or in-app picker lands.
func PickPlanJSONOpen() (string, error) {
	return androidFileDialogStub("Open plan JSON")
}

// PickTextFileOpen is a stub on Android until SAF or in-app picker lands.
func PickTextFileOpen() (string, error) {
	return androidFileDialogStub("Open")
}

// PickTextFileSave is a stub on Android until SAF or in-app picker lands.
func PickTextFileSave(defaultName string) (string, error) {
	_ = defaultName
	return androidFileDialogStub("Save As")
}

// PickHTMLReportSave is a stub on Android until SAF or in-app picker lands.
func PickHTMLReportSave(defaultName string) (string, error) {
	_ = defaultName
	return androidFileDialogStub("Export plan report")
}

// PickPNGSave is a stub on Android until SAF or in-app picker lands.
func PickPNGSave(defaultName string) (string, error) {
	_ = defaultName
	return androidFileDialogStub("Export graph PNG")
}
