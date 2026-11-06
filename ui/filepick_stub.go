//go:build !android

package ui

import "errors"

var (
	errFilePickCancelled = errors.New("file pick cancelled")
	errFilePickBusy      = errors.New("file pick already in progress")
)

// ErrFilePickCancelled reports a dismissed file dialog (Android SAF parity).
func ErrFilePickCancelled() error { return errFilePickCancelled }

// PlatformStartTextFileOpen is unavailable off Android.
func PlatformStartTextFileOpen(onDone func(path string, err error)) {
	if onDone != nil {
		onDone("", errFilePickBusy)
	}
}

// PlatformStartTextFileSave is unavailable off Android.
func PlatformStartTextFileSave(defaultName string, initial []byte, onDone func(path string, err error)) {
	_ = defaultName
	_ = initial
	if onDone != nil {
		onDone("", errFilePickBusy)
	}
}

// PlatformPickTextFileOpen is unavailable off Android.
func PlatformPickTextFileOpen() (string, error) {
	return "", errFilePickBusy
}

// PlatformPickTextFileSave is unavailable off Android.
func PlatformPickTextFileSave(defaultName string, initial []byte) (string, error) {
	_ = defaultName
	_ = initial
	return "", errFilePickBusy
}
