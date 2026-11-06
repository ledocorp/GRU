//go:build android

// Android OpenBrowser via Activity.openGruUrl → ACTION_VIEW.
// C implementation: gruOpenBrowser in ime_android_jni.go. See docs/ANDROID_CODE.md.
package ui

/*
#include <stdlib.h>
void gruOpenBrowser(const char *url);
*/
import "C"

import (
	"errors"
	"unsafe"
)

// OpenBrowser opens url in an external browser (or Custom Tabs handler) via ACTION_VIEW.
func OpenBrowser(url string) error {
	if url == "" {
		return errors.New("empty url")
	}
	c := C.CString(url)
	defer C.free(unsafe.Pointer(c))
	C.gruOpenBrowser(c)
	return nil
}
