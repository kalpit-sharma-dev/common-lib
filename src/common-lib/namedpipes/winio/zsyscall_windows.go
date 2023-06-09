package winio

// nolint
import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	modwinmm    = windows.NewLazySystemDLL("winmm.dll")
	modadvapi32 = windows.NewLazySystemDLL("advapi32.dll")

	procCancelIoEx                                           = modkernel32.NewProc("CancelIoEx")
	procCreateIoCompletionPort                               = modkernel32.NewProc("CreateIoCompletionPort")
	procGetQueuedCompletionStatus                            = modkernel32.NewProc("GetQueuedCompletionStatus")
	procSetFileCompletionNotificationModes                   = modkernel32.NewProc("SetFileCompletionNotificationModes")
	proctimeBeginPeriod                                      = modwinmm.NewProc("timeBeginPeriod")
	procConnectNamedPipe                                     = modkernel32.NewProc("ConnectNamedPipe")
	procCreateNamedPipeW                                     = modkernel32.NewProc("CreateNamedPipeW")
	procCreateFileW                                          = modkernel32.NewProc("CreateFileW")
	procWaitNamedPipeW                                       = modkernel32.NewProc("WaitNamedPipeW")
	procGetNamedPipeInfo                                     = modkernel32.NewProc("GetNamedPipeInfo")
	procGetNamedPipeHandleStateW                             = modkernel32.NewProc("GetNamedPipeHandleStateW")
	procLocalAlloc                                           = modkernel32.NewProc("LocalAlloc")
	procLookupAccountNameW                                   = modadvapi32.NewProc("LookupAccountNameW")
	procConvertSidToStringSidW                               = modadvapi32.NewProc("ConvertSidToStringSidW")
	procConvertStringSecurityDescriptorToSecurityDescriptorW = modadvapi32.NewProc("ConvertStringSecurityDescriptorToSecurityDescriptorW")
	procConvertSecurityDescriptorToStringSecurityDescriptorW = modadvapi32.NewProc("ConvertSecurityDescriptorToStringSecurityDescriptorW")
	procLocalFree                                            = modkernel32.NewProc("LocalFree")
	procGetSecurityDescriptorLength                          = modadvapi32.NewProc("GetSecurityDescriptorLength")
)

func cancelIoEx(file syscall.Handle, o *syscall.Overlapped) (err error) {
	r1, _, e1 := syscall.Syscall(procCancelIoEx.Addr(), 2, uintptr(file), uintptr(unsafe.Pointer(o)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func createIoCompletionPort(file syscall.Handle, port syscall.Handle, key uintptr, threadCount uint32) (newport syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall6(procCreateIoCompletionPort.Addr(), 4, uintptr(file), uintptr(port), uintptr(key), uintptr(threadCount), 0, 0)
	newport = syscall.Handle(r0)
	if newport == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func getQueuedCompletionStatus(port syscall.Handle, bytes *uint32, key *uintptr, o **ioOperation, timeout uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procGetQueuedCompletionStatus.Addr(), 5, uintptr(port), uintptr(unsafe.Pointer(bytes)), uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(o)), uintptr(timeout), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func setFileCompletionNotificationModes(h syscall.Handle, flags uint8) (err error) {
	r1, _, e1 := syscall.Syscall(procSetFileCompletionNotificationModes.Addr(), 2, uintptr(h), uintptr(flags), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}

	return
}

func timeBeginPeriod(period uint32) (n int32) {
	r0, _, _ := syscall.Syscall(proctimeBeginPeriod.Addr(), 1, uintptr(period), 0, 0)
	n = int32(r0)
	return
}

func connectNamedPipe(pipe syscall.Handle, o *syscall.Overlapped) (err error) {
	r1, _, e1 := syscall.Syscall(procConnectNamedPipe.Addr(), 2, uintptr(pipe), uintptr(unsafe.Pointer(o)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func createNamedPipe(name string, flags uint32, pipeMode uint32, maxInstances uint32, outSize uint32, inSize uint32, defaultTimeout uint32, sa *syscall.SecurityAttributes) (handle syscall.Handle, err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(name)
	if err != nil {
		return
	}
	return _createNamedPipe(_p0, flags, pipeMode, maxInstances, outSize, inSize, defaultTimeout, sa)
}

func _createNamedPipe(name *uint16, flags uint32, pipeMode uint32, maxInstances uint32, outSize uint32, inSize uint32, defaultTimeout uint32, sa *syscall.SecurityAttributes) (handle syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall9(procCreateNamedPipeW.Addr(), 8, uintptr(unsafe.Pointer(name)), uintptr(flags), uintptr(pipeMode), uintptr(maxInstances), uintptr(outSize), uintptr(inSize), uintptr(defaultTimeout), uintptr(unsafe.Pointer(sa)), 0)
	handle = syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func createFile(name string, access uint32, mode uint32, sa *syscall.SecurityAttributes, createmode uint32, attrs uint32, templatefile syscall.Handle) (handle syscall.Handle, err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(name)
	if err != nil {
		return
	}
	return _createFile(_p0, access, mode, sa, createmode, attrs, templatefile)
}

func _createFile(name *uint16, access uint32, mode uint32, sa *syscall.SecurityAttributes, createmode uint32, attrs uint32, templatefile syscall.Handle) (handle syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall9(procCreateFileW.Addr(), 7, uintptr(unsafe.Pointer(name)), uintptr(access), uintptr(mode), uintptr(unsafe.Pointer(sa)), uintptr(createmode), uintptr(attrs), uintptr(templatefile), 0, 0)
	handle = syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func waitNamedPipe(name string, timeout uint32) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(name)
	if err != nil {
		return
	}
	return _waitNamedPipe(_p0, timeout)
}

func _waitNamedPipe(name *uint16, timeout uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procWaitNamedPipeW.Addr(), 2, uintptr(unsafe.Pointer(name)), uintptr(timeout), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func getNamedPipeInfo(pipe syscall.Handle, flags *uint32, outSize *uint32, inSize *uint32, maxInstances *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procGetNamedPipeInfo.Addr(), 5, uintptr(pipe), uintptr(unsafe.Pointer(flags)), uintptr(unsafe.Pointer(outSize)), uintptr(unsafe.Pointer(inSize)), uintptr(unsafe.Pointer(maxInstances)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func getNamedPipeHandleState(pipe syscall.Handle, state *uint32, curInstances *uint32, maxCollectionCount *uint32, collectDataTimeout *uint32, userName *uint16, maxUserNameSize uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procGetNamedPipeHandleStateW.Addr(), 7, uintptr(pipe), uintptr(unsafe.Pointer(state)), uintptr(unsafe.Pointer(curInstances)), uintptr(unsafe.Pointer(maxCollectionCount)), uintptr(unsafe.Pointer(collectDataTimeout)), uintptr(unsafe.Pointer(userName)), uintptr(maxUserNameSize), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func localAlloc(uFlags uint32, length uint32) (ptr uintptr) {
	r0, _, _ := syscall.Syscall(procLocalAlloc.Addr(), 2, uintptr(uFlags), uintptr(length), 0)
	ptr = uintptr(r0)
	return
}

func lookupAccountName(systemName *uint16, accountName string, sid *byte, sidSize *uint32, refDomain *uint16, refDomainSize *uint32, sidNameUse *uint32) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(accountName)
	if err != nil {
		return
	}
	return _lookupAccountName(systemName, _p0, sid, sidSize, refDomain, refDomainSize, sidNameUse)
}

func _lookupAccountName(systemName *uint16, accountName *uint16, sid *byte, sidSize *uint32, refDomain *uint16, refDomainSize *uint32, sidNameUse *uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procLookupAccountNameW.Addr(), 7, uintptr(unsafe.Pointer(systemName)), uintptr(unsafe.Pointer(accountName)), uintptr(unsafe.Pointer(sid)), uintptr(unsafe.Pointer(sidSize)), uintptr(unsafe.Pointer(refDomain)), uintptr(unsafe.Pointer(refDomainSize)), uintptr(unsafe.Pointer(sidNameUse)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func convertSidToStringSid(sid *byte, str **uint16) (err error) {
	r1, _, e1 := syscall.Syscall(procConvertSidToStringSidW.Addr(), 2, uintptr(unsafe.Pointer(sid)), uintptr(unsafe.Pointer(str)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func convertStringSecurityDescriptorToSecurityDescriptor(str string, revision uint32, sd *uintptr, size *uint32) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(str)
	if err != nil {
		return
	}
	return _convertStringSecurityDescriptorToSecurityDescriptor(_p0, revision, sd, size)
}

func _convertStringSecurityDescriptorToSecurityDescriptor(str *uint16, revision uint32, sd *uintptr, size *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procConvertStringSecurityDescriptorToSecurityDescriptorW.Addr(), 4, uintptr(unsafe.Pointer(str)), uintptr(revision), uintptr(unsafe.Pointer(sd)), uintptr(unsafe.Pointer(size)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func convertSecurityDescriptorToStringSecurityDescriptor(sd *byte, revision uint32, secInfo uint32, sddl **uint16, sddlSize *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procConvertSecurityDescriptorToStringSecurityDescriptorW.Addr(), 5, uintptr(unsafe.Pointer(sd)), uintptr(revision), uintptr(secInfo), uintptr(unsafe.Pointer(sddl)), uintptr(unsafe.Pointer(sddlSize)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func localFree(mem uintptr) {
	syscall.Syscall(procLocalFree.Addr(), 1, uintptr(mem), 0, 0)
	return
}

func getSecurityDescriptorLength(sd uintptr) (len uint32) {
	r0, _, _ := syscall.Syscall(procGetSecurityDescriptorLength.Addr(), 1, uintptr(sd), 0, 0)
	len = uint32(r0)
	return
}
