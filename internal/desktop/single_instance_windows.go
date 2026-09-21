//go:build windows

package desktop

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	user32       = syscall.NewLazyDLL("user32.dll")
	createMutexW = kernel32.NewProc("CreateMutexW")
	messageBoxW  = user32.NewProc("MessageBoxW")
)

type InstanceLock struct {
	handle syscall.Handle
}

func AcquireSingleInstance(name string) (*InstanceLock, bool, error) {
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, false, err
	}
	handle, _, callErr := createMutexW.Call(0, 0, uintptr(unsafe.Pointer(namePtr)))
	if handle == 0 {
		return nil, false, fmt.Errorf("create mutex: %w", callErr)
	}
	alreadyRunning := callErr == syscall.ERROR_ALREADY_EXISTS
	return &InstanceLock{handle: syscall.Handle(handle)}, alreadyRunning, nil
}

func (l *InstanceLock) Close() error {
	if l == nil || l.handle == 0 {
		return nil
	}
	err := syscall.CloseHandle(l.handle)
	l.handle = 0
	return err
}

func ShowAlreadyRunningMessage() {
	text, _ := syscall.UTF16PtrFromString("鼠标保活已在系统托盘中运行。")
	title, _ := syscall.UTF16PtrFromString("鼠标保活")
	messageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x40)
}
