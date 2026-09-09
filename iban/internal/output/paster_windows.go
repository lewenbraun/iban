//go:build windows

package output

import (
	"errors"
	"strconv"
	"syscall"
	"time"
)

const focusSettleDelay = 150 * time.Millisecond

var (
	procGetForegroundWindow = syscall.NewLazyDLL("user32.dll").NewProc("GetForegroundWindow")
	procSetForegroundWindow = syscall.NewLazyDLL("user32.dll").NewProc("SetForegroundWindow")
)

type winPaster struct{}

// CaptureTarget returns the foreground window handle as the paste target.
func (winPaster) CaptureTarget() (string, error) {
	//nolint:gosec // GetForegroundWindow takes no arguments
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return "", errors.New("no foreground window")
	}
	return strconv.FormatInt(int64(hwnd), 10), nil
}

// PasteTarget focuses the saved window and pastes the clipboard into it,
// optionally pressing Enter afterwards.
func (winPaster) PasteTarget(target string, pressEnter bool) error {
	hwnd, err := strconv.ParseUint(target, 10, 64)
	if err != nil || hwnd == 0 {
		return errors.New("invalid window handle")
	}
	if err := focusWindow(hwnd); err != nil {
		return err
	}
	if err := sendTap(vkControl); err != nil {
		return errors.New("paste into window")
	}
	if !pressEnter {
		return nil
	}
	time.Sleep(focusSettleDelay)
	if err := sendTap(vkReturn); err != nil {
		return errors.New("send Enter to window")
	}
	return nil
}

func focusWindow(hwnd uint64) error {
	if err := sendTap(vkMenu); err != nil {
		return errors.New("arm focus switch")
	}
	//nolint:gosec // SetForegroundWindow takes the target window handle
	ret, _, _ := procSetForegroundWindow.Call(uintptr(hwnd))
	if ret == 0 {
		return errors.New("focus target window")
	}
	time.Sleep(focusSettleDelay)
	return nil
}
