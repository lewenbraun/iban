//go:build windows

// Package tray runs the Windows dictation surface: a tray icon mirroring
// the session state and a global Alt+Space recording chord (V selects
// paste-only, B selects clipboard-only).
package tray

import (
	"runtime"
	"syscall"
	"unsafe"

	"github.com/lewenbraun/iban/iban/internal/dictation"
	"github.com/lewenbraun/iban/iban/internal/hotkey"
	"github.com/lewenbraun/iban/iban/internal/state"
)

const (
	wmTrayCallback = 0x8000
	wmDestroy      = 0x0002
	wmClose        = 0x0010
	wmContextmenu  = 0x007B
	wmRbuttonup    = 0x0205
	whKeyboardLL   = 13
	llkhfInjected  = 0x10
	wmKeydown      = 0x0100
	wmKeyup        = 0x0101
	wmSyskeydown   = 0x0104
	wmSyskeyup     = 0x0105
	nimAdd         = 0
	nimModify      = 1
	nimDelete      = 2
	nifMessage     = 0x1
	nifIcon        = 0x2
	nifTip         = 0x4
	nifInfo        = 0x10
	trayIconID     = 1
)

var (
	procGetModuleHandle     = syscall.NewLazyDLL("kernel32.dll").NewProc("GetModuleHandleW")
	procRegisterClassEx     = syscall.NewLazyDLL("user32.dll").NewProc("RegisterClassExW")
	procUnregisterClass     = syscall.NewLazyDLL("user32.dll").NewProc("UnregisterClassW")
	procCreateWindowEx      = syscall.NewLazyDLL("user32.dll").NewProc("CreateWindowExW")
	procDestroyWindow       = syscall.NewLazyDLL("user32.dll").NewProc("DestroyWindow")
	procDefWindowProc       = syscall.NewLazyDLL("user32.dll").NewProc("DefWindowProcW")
	procPostMessage         = syscall.NewLazyDLL("user32.dll").NewProc("PostMessageW")
	procPostQuitMessage     = syscall.NewLazyDLL("user32.dll").NewProc("PostQuitMessage")
	procGetMessage          = syscall.NewLazyDLL("user32.dll").NewProc("GetMessageW")
	procTranslateMessage    = syscall.NewLazyDLL("user32.dll").NewProc("TranslateMessage")
	procDispatchMessage     = syscall.NewLazyDLL("user32.dll").NewProc("DispatchMessageW")
	procShellNotifyIcon     = syscall.NewLazyDLL("shell32.dll").NewProc("Shell_NotifyIconW")
	procSetWindowsHookEx    = syscall.NewLazyDLL("user32.dll").NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = syscall.NewLazyDLL("user32.dll").NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = syscall.NewLazyDLL("user32.dll").NewProc("CallNextHookEx")
)

// Tray indicator colors (0x00BBGGRR).
const (
	colorIdle    uint32 = 0x00A0A0A0
	colorRecord  uint32 = 0x00202CE8 // red
	colorWorking uint32 = 0x0000C8E6 // yellow
	colorDone    uint32 = 0x0000A34A // green
)

type chordEvent struct {
	action hotkey.Action
	mode   state.Mode
}

// Run starts the tray icon and the Alt+Space chord listener and blocks
// until the user exits through the tray icon (right click quits).
func Run(svc *dictation.Service, store *state.Store) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	icons, err := newIconSet()
	if err != nil {
		return err
	}
	defer icons.dispose()
	w, err := createTrayWindow(icons)
	if err != nil {
		return err
	}
	defer w.dispose()
	events := make(chan chordEvent, 8)
	hook, err := installKeyboardHook(w, events)
	if err != nil {
		return err
	}
	defer uninstallKeyboardHook(hook)
	go runWorker(svc, store, w, events)
	runMessageLoop()
	close(events)
	return nil
}

func newIconSet() (iconSet, error) {
	idle := solidIcon(colorIdle)
	record := solidIcon(colorRecord)
	working := solidIcon(colorWorking)
	done := solidIcon(colorDone)
	if idle == 0 || record == 0 || working == 0 || done == 0 {
		return iconSet{}, syscall.GetLastError()
	}
	return iconSet{idle: idle, record: record, working: working, done: done}, nil
}

type iconSet struct {
	idle    uintptr
	record  uintptr
	working uintptr
	done    uintptr
}

func (s iconSet) dispose() {
	destroyIcon(s.idle)
	destroyIcon(s.record)
	destroyIcon(s.working)
	destroyIcon(s.done)
}

func runMessageLoop() {
	var msg msgStruct
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 || ret == 0xFFFFFFFF {
			return
		}
		_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		_, _, _ = procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
