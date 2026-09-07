//go:build windows

package tray

import (
	"syscall"
	"unsafe"
)

const trayClassName = "eban-tray"

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type msgStruct struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       pointStruct
	lPrivate uint32
}

type pointStruct struct {
	x int32
	y int32
}

type notifyIconData struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         [16]byte
	hBalloonIcon     uintptr
}

type trayWindow struct {
	hwnd  uintptr
	hinst uintptr
	icons iconSet
}

func createTrayWindow(icons iconSet) (*trayWindow, error) {
	hinst, _, _ := procGetModuleHandle.Call(0)
	if hinst == 0 {
		return nil, syscall.GetLastError()
	}
	w := &trayWindow{hinst: hinst, icons: icons}
	className, _ := syscall.UTF16PtrFromString(trayClassName)
	title, _ := syscall.UTF16PtrFromString("eban")
	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   syscall.NewCallback(w.wndProc),
		hInstance:     hinst,
		lpszClassName: className,
	}
	if ret, _, _ := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); ret == 0 {
		return nil, syscall.GetLastError()
	}
	//nolint:gosec // CreateWindowEx takes the class name by pointer
	hwnd, _, _ := procCreateWindowEx.Call(
		0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(title)),
		0, 0, 0, 0, 0, 0, 0, hinst, 0)
	if hwnd == 0 {
		return nil, syscall.GetLastError()
	}
	w.hwnd = hwnd
	if err := w.addTrayIcon(); err != nil {
		w.dispose()
		return nil, err
	}
	return w, nil
}

func (w *trayWindow) wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmTrayCallback:
		if low := lParam & 0xFFFF; low == wmContextmenu || low == wmRbuttonup {
			_, _, _ = procPostMessage.Call(hwnd, wmClose, 0, 0)
		}
		return 0
	case wmDestroy:
		_, _, _ = procPostQuitMessage.Call(0)
		return 0
	default:
		ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
		return ret
	}
}

func (w *trayWindow) baseNotify(uFlags uint32) notifyIconData {
	var nid notifyIconData
	nid.cbSize = uint32(unsafe.Sizeof(nid))
	nid.hWnd = w.hwnd
	nid.uID = trayIconID
	nid.uFlags = uFlags
	return nid
}

func (w *trayWindow) addTrayIcon() error {
	nid := w.baseNotify(nifMessage | nifIcon | nifTip)
	nid.uCallbackMessage = wmTrayCallback
	nid.hIcon = w.icons.idle
	copy(nid.szTip[:], syscall.StringToUTF16("eban dictation"))
	if ret, _, _ := procShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&nid))); ret == 0 {
		return syscall.GetLastError()
	}
	return nil
}

func (w *trayWindow) setIcon(icon uintptr) {
	nid := w.baseNotify(nifIcon)
	nid.hIcon = icon
	_, _, _ = procShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

func (w *trayWindow) balloon(text string) {
	nid := w.baseNotify(nifInfo)
	copy(nid.szInfo[:], syscall.StringToUTF16(text))
	copy(nid.szInfoTitle[:], syscall.StringToUTF16("eban"))
	_, _, _ = procShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

func (w *trayWindow) dispose() {
	if w.hwnd != 0 {
		nid := w.baseNotify(0)
		_, _, _ = procShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
		_, _, _ = procDestroyWindow.Call(w.hwnd)
		className, _ := syscall.UTF16PtrFromString(trayClassName)
		_, _, _ = procUnregisterClass.Call(uintptr(unsafe.Pointer(className)), w.hinst)
	}
}
