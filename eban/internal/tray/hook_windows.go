//go:build windows

package tray

import (
	"syscall"
	"time"
	"unsafe"

	"github.com/lewenbraun/eban/eban/internal/dictation"
	"github.com/lewenbraun/eban/eban/internal/hotkey"
	"github.com/lewenbraun/eban/eban/internal/state"
)

type kbdllHookStruct struct {
	vkCode      uint32
	scanCode    uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

type keyHook struct {
	chord  hotkey.Chord
	events chan<- chordEvent
}

const doneLinger = 4 * time.Second

func installKeyboardHook(w *trayWindow, events chan<- chordEvent) (uintptr, error) {
	h := &keyHook{events: events}
	hook, _, err := procSetWindowsHookEx.Call(
		whKeyboardLL, syscall.NewCallback(h.proc), w.hinst, 0)
	if hook == 0 {
		return 0, err
	}
	return hook, nil
}

func uninstallKeyboardHook(hook uintptr) {
	if hook != 0 {
		_, _, _ = procUnhookWindowsHookEx.Call(hook)
	}
}

func (h *keyHook) proc(nCode int32, wParam, lParam uintptr) uintptr {
	if nCode < 0 {
		return callNextHook(nCode, wParam, lParam)
	}
	action, mode, swallow := h.key(wParam, lParam)
	if action != hotkey.ActionNone {
		h.post(chordEvent{action: action, mode: mode})
	}
	if swallow {
		return 1
	}
	return callNextHook(nCode, wParam, lParam)
}

func (h *keyHook) key(wParam, lParam uintptr) (hotkey.Action, state.Mode, bool) {
	//nolint:gosec // the hook receives the KBDLLHOOKSTRUCT by pointer
	kbd := (*kbdllHookStruct)(unsafe.Pointer(lParam))
	e := hotkey.Event{
		VK:       kbd.vkCode,
		Down:     wParam == wmKeydown || wParam == wmSyskeydown,
		Injected: kbd.flags&llkhfInjected != 0,
	}
	action, mode, swallow := h.chord.Key(e)
	return action, mode, swallow
}

func (h *keyHook) post(ev chordEvent) {
	select {
	case h.events <- ev:
	default:
	}
}

func callNextHook(nCode int32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

func runWorker(svc *dictation.Service, store *state.Store, w *trayWindow, events <-chan chordEvent) {
	for ev := range events {
		runEvent(svc, store, w, ev)
	}
}

func runEvent(svc *dictation.Service, store *state.Store, w *trayWindow, ev chordEvent) {
	switch ev.action {
	case hotkey.ActionStart:
		startRecording(svc, w)
	case hotkey.ActionSelect:
		store.SetMode(ev.mode)
	case hotkey.ActionStop:
		stopRecording(svc, w)
	case hotkey.ActionNone:
	}
}

func startRecording(svc *dictation.Service, w *trayWindow) {
	w.setIcon(w.icons.record)
	if err := svc.Start(state.ModePasteEnter, "", 0); err != nil {
		w.balloon(err.Error())
		w.setIcon(w.icons.idle)
	}
}

func stopRecording(svc *dictation.Service, w *trayWindow) {
	w.setIcon(w.icons.working)
	if err := svc.Stop("", false, false); err != nil {
		w.balloon(err.Error())
		w.setIcon(w.icons.idle)
		return
	}
	w.setIcon(w.icons.done)
	time.AfterFunc(doneLinger, func() { w.setIcon(w.icons.idle) })
}
