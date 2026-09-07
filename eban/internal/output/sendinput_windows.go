//go:build windows

package output

import (
	"encoding/binary"
	"errors"
	"syscall"
	"unsafe"
)

const (
	inputKeyboard    = 1
	keyeventfKeyup   = 0x0002
	vkControl        = 0x11
	vkMenu           = 0x12
	vkReturn         = 0x0D
	keybdPayloadSize = 32
	inputStructSize  = 40
)

var procSendInput = syscall.NewLazyDLL("user32.dll").NewProc("SendInput")

// keyboardInput mirrors the win32 INPUT structure: the keyboard union arm
// is serialized into a fixed 32-byte payload to avoid unsafe unions.
type keyboardInput struct {
	typ     uint32
	_       [4]byte
	payload [keybdPayloadSize]byte
}

func sendKey(vk uint16, up bool) error {
	var in keyboardInput
	in.typ = inputKeyboard
	flags := uint32(0)
	if up {
		flags = keyeventfKeyup
	}
	binary.LittleEndian.PutUint16(in.payload[0:], vk)
	binary.LittleEndian.PutUint32(in.payload[4:], flags)
	//nolint:gosec // SendInput requires a pointer to the INPUT array
	ret, _, _ := procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), inputStructSize)
	if ret == 0 {
		return errors.New("send keyboard input")
	}
	return nil
}

func sendTap(vk uint16) error {
	if err := sendKey(vk, false); err != nil {
		return err
	}
	return sendKey(vk, true)
}
