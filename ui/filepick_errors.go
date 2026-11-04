//go:build android

package ui

import "errors"

var (
	errFilePickCancelled = errors.New("file pick cancelled")
	errFilePickBusy      = errors.New("file pick already in progress")
)

// ErrFilePickCancelled reports a dismissed SAF dialog.
func ErrFilePickCancelled() error { return errFilePickCancelled }
