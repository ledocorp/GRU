//go:build windows

package appinstance

import (
	"strings"
	"syscall"
	"unsafe"
)

const mutexName = "Local\\GruNotepad_SingleInstance_v1"

var (
	modKernel32             = syscall.NewLazyDLL("kernel32.dll")
	modUser32               = syscall.NewLazyDLL("user32.dll")
	procCreateMutexW        = modKernel32.NewProc("CreateMutexW")
	procEnumWindows         = modUser32.NewProc("EnumWindows")
	procGetWindowTextW      = modUser32.NewProc("GetWindowTextW")
	procIsWindowVisible     = modUser32.NewProc("IsWindowVisible")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
)

func startup(forwardPath string, onOpen func(path string)) (exit bool, stop func()) {
	handle, already := createAppMutex(true)
	if handle == 0 {
		return false, func() {}
	}
	if already {
		syscall.CloseHandle(handle)
		_ = writePendingOpen(forwardPath)
		foregroundExistingInstance()
		return true, func() {}
	}

	stopPoll := startPendingPoll(onOpen)
	return false, func() {
		stopPoll()
		syscall.CloseHandle(handle)
	}
}

func createAppMutex(initialOwner bool) (handle syscall.Handle, alreadyExists bool) {
	name, _ := syscall.UTF16PtrFromString(mutexName)
	owner := uintptr(0)
	if initialOwner {
		owner = 1
	}
	r, _, err := procCreateMutexW.Call(0, owner, uintptr(unsafe.Pointer(name)))
	if r == 0 {
		return 0, false
	}
	return syscall.Handle(r), err == syscall.ERROR_ALREADY_EXISTS
}

func foregroundExistingInstance() {
	if hwnd := findMainWindowHWND(); hwnd != 0 {
		procSetForegroundWindow.Call(hwnd)
	}
}

func findMainWindowHWND() uintptr {
	var found uintptr
	cb := syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}
		var buf [512]uint16
		procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		title := syscall.UTF16ToString(buf[:])
		if title == "Gru Notepad" || strings.Contains(title, "Notepad (Go)") {
			found = hwnd
			return 0
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return found
}
