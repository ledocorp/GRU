//go:build android

package ui

/*
#include <stdint.h>
#include <stdlib.h>
void gruCallPickOpen(void);
void gruCallPickSave(const char *name, const void *data, int n);
*/
import "C"

import (
	"sync"
	"unsafe"
)

type filePickResult struct {
	path      string
	cancelled bool
}

type filePickCallback func(path string, err error)

var (
	filePickMu    sync.Mutex
	fileOpenDone  filePickCallback
	fileSaveDone  filePickCallback
)

//export goFileOpenResult
func goFileOpenResult(path *C.char, cancelled C.int) {
	p := C.GoString(path)
	if path != nil {
		C.free(unsafe.Pointer(path))
	}
	filePickMu.Lock()
	cb := fileOpenDone
	fileOpenDone = nil
	filePickMu.Unlock()
	if cb == nil {
		return
	}
	if cancelled != 0 || p == "" {
		cb("", errFilePickCancelled)
		return
	}
	cb(p, nil)
}

//export goFileSaveResult
func goFileSaveResult(path *C.char, cancelled C.int) {
	p := C.GoString(path)
	if path != nil {
		C.free(unsafe.Pointer(path))
	}
	filePickMu.Lock()
	cb := fileSaveDone
	fileSaveDone = nil
	filePickMu.Unlock()
	if cb == nil {
		return
	}
	if cancelled != 0 || p == "" {
		cb("", errFilePickCancelled)
		return
	}
	cb(p, nil)
}

// PlatformStartTextFileOpen launches SAF OPEN_DOCUMENT without blocking the
// NativeActivity game loop (blocking caused ANRs). onDone may run on the JNI /
// UI thread — callers should QueueMain before touching the scene.
func PlatformStartTextFileOpen(onDone func(path string, err error)) {
	if onDone == nil {
		return
	}
	filePickMu.Lock()
	if fileOpenDone != nil {
		filePickMu.Unlock()
		onDone("", errFilePickBusy)
		return
	}
	fileOpenDone = onDone
	filePickMu.Unlock()
	C.gruCallPickOpen()
}

// PlatformStartTextFileSave launches SAF CREATE_DOCUMENT without blocking.
func PlatformStartTextFileSave(defaultName string, initial []byte, onDone func(path string, err error)) {
	if onDone == nil {
		return
	}
	filePickMu.Lock()
	if fileSaveDone != nil {
		filePickMu.Unlock()
		onDone("", errFilePickBusy)
		return
	}
	fileSaveDone = onDone
	filePickMu.Unlock()
	cname := C.CString(defaultName)
	defer C.free(unsafe.Pointer(cname))
	var ptr unsafe.Pointer
	n := len(initial)
	if n > 0 {
		ptr = C.CBytes(initial)
		defer C.free(ptr)
	}
	C.gruCallPickSave(cname, ptr, C.int(n))
}

// PlatformPickTextFileOpen is a sync wrapper for non-loop callers; prefer Start.
func PlatformPickTextFileOpen() (string, error) {
	ch := make(chan filePickResult, 1)
	PlatformStartTextFileOpen(func(path string, err error) {
		ch <- filePickResult{path: path, cancelled: err != nil}
	})
	r := <-ch
	if r.cancelled || r.path == "" {
		return "", errFilePickCancelled
	}
	return r.path, nil
}

// PlatformPickTextFileSave is a sync wrapper for non-loop callers; prefer Start.
func PlatformPickTextFileSave(defaultName string, initial []byte) (string, error) {
	ch := make(chan filePickResult, 1)
	PlatformStartTextFileSave(defaultName, initial, func(path string, err error) {
		ch <- filePickResult{path: path, cancelled: err != nil}
	})
	r := <-ch
	if r.cancelled || r.path == "" {
		return "", errFilePickCancelled
	}
	return r.path, nil
}
