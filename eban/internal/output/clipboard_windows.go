//go:build windows

package output

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

var (
	procOpenClipboard    = syscall.NewLazyDLL("user32.dll").NewProc("OpenClipboard")
	procCloseClipboard   = syscall.NewLazyDLL("user32.dll").NewProc("CloseClipboard")
	procEmptyClipboard   = syscall.NewLazyDLL("user32.dll").NewProc("EmptyClipboard")
	procSetClipboardData = syscall.NewLazyDLL("user32.dll").NewProc("SetClipboardData")
	procGlobalAlloc      = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalAlloc")
	procGlobalLock       = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalLock")
	procGlobalUnlock     = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalUnlock")
	procGlobalFree       = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalFree")
)

type winClipboard struct{}

// Copy places text on the system clipboard.
func (winClipboard) Copy(text string) error {
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}
	if ret, _, _ := procOpenClipboard.Call(0); ret == 0 {
		return errors.New("open clipboard")
	}
	defer func() { _, _, _ = procCloseClipboard.Call() }()
	if ret, _, _ := procEmptyClipboard.Call(); ret == 0 {
		return errors.New("empty clipboard")
	}
	return allocClipboardText(utf16)
}

func allocClipboardText(utf16 []uint16) error {
	h, _, _ := procGlobalAlloc.Call(gmemMoveable, uintptr(len(utf16))*2)
	if h == 0 {
		return errors.New("allocate clipboard memory")
	}
	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		_, _, _ = procGlobalFree.Call(h)
		return errors.New("lock clipboard memory")
	}
	//nolint:gosec // clipboard copy requires a raw memory write
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(utf16))
	copy(dst, utf16)
	_, _, _ = procGlobalUnlock.Call(h)
	if ret, _, _ := procSetClipboardData.Call(cfUnicodeText, h); ret == 0 {
		_, _, _ = procGlobalFree.Call(h)
		return errors.New("set clipboard data")
	}
	return nil
}
