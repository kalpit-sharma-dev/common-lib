package pluginUtils

import (
	"runtime"
	"syscall"
	"unsafe"
)

var (
	kerneldll             = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentProcess = kerneldll.NewProc("GetCurrentProcess")
	procIsWow64Process    = kerneldll.NewProc("IsWow64Process")
)

// Is64BitOS verifies os architecture
func Is64BitOS() bool {
	if "amd64" == runtime.GOARCH {
		return true
	} else if "386" == runtime.GOARCH {
		f64 := false
		return isWow64Process(getCurrentProcess(), &f64)
	}
	return false
}

// DisableRedirection disables file system redirection for the calling thread. File system redirection is enabled by default.
// OldValue is file system redirection value.
// The system uses this parameter to store information necessary to re-enable file system redirection.
func DisableRedirection(OldValue *uintptr) {
	kerneldll.NewProc("Wow64DisableWow64FsRedirection").Call(uintptr(unsafe.Pointer(OldValue)))
}

// RevertRedirection restores file system redirection for the calling thread.
// This function should not be called without a previous call to the DisableRedirection function.
func RevertRedirection(OldValue *uintptr) {
	kerneldll.NewProc("Wow64RevertWow64FsRedirection").Call(uintptr(unsafe.Pointer(OldValue)))
}

func getCurrentProcess() syscall.Handle {
	p0, _, _ := procGetCurrentProcess.Call()

	return syscall.Handle(p0)
}

func isWow64Process(hProcess syscall.Handle, wow64Process *bool) bool {
	p0, _, _ := procIsWow64Process.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(wow64Process)))

	return p0 != 0
}
