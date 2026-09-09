//go:build windows

package recorder

import (
	"syscall"
	"unsafe"
)

var (
	ole32                = syscall.NewLazyDLL("ole32.dll")
	procCoInitializeEx   = ole32.NewProc("CoInitializeEx")
	procCoUninitialize   = ole32.NewProc("CoUninitialize")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")
	procCoTaskMemFree    = ole32.NewProc("CoTaskMemFree")
)

const (
	coinitMultithreaded = 0
	clsctxAll           = 0x17
	eCapture            = 1
	roleConsole         = 0
	roleCommunications  = 1
)

var clsidMMDeviceEnumerator = syscall.GUID{
	Data1: 0xBCDE0395, Data2: 0xE52F, Data3: 0x467C,
	Data4: [8]byte{0x8E, 0x3D, 0xC4, 0x57, 0x92, 0x91, 0x69, 0x2E},
}

var iidIMMDeviceEnumerator = syscall.GUID{
	Data1: 0xA95664D2, Data2: 0x9614, Data3: 0x4F35,
	Data4: [8]byte{0xA7, 0x46, 0xDE, 0x8D, 0xB6, 0x36, 0x17, 0xE6},
}

// DefaultCaptureEndpointIDs returns the endpoint IDs of the Windows default
// capture devices: the console role (the input device in Sound settings)
// first, then the communications role. Failures yield a shorter or empty
// list; callers fall back to the first DirectShow device.
func DefaultCaptureEndpointIDs() []string {
	ids := make([]string, 0, 2)
	for _, role := range []uint32{roleConsole, roleCommunications} {
		if id, err := endpointID(role); err == nil && id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// endpointID queries the MMDeviceEnumerator for the default capture
// endpoint ID of one role.
//
//nolint:gosec // manual COM vtable calls through syscall
func endpointID(role uint32) (string, error) {
	hrCo, _, _ := procCoInitializeEx.Call(0, coinitMultithreaded)
	needUninit := hrCo == 0
	if needUninit {
		defer func() { _, _, _ = procCoUninitialize.Call() }()
	}
	var enumPtr uintptr
	hrCreate, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidMMDeviceEnumerator)), 0, clsctxAll,
		uintptr(unsafe.Pointer(&iidIMMDeviceEnumerator)), uintptr(unsafe.Pointer(&enumPtr)))
	if hrCreate != 0 || enumPtr == 0 {
		return "", syscall.Errno(hrCreate)
	}
	defer func() { _ = comRelease(unsafe.Pointer(enumPtr)) }()

	var dev uintptr
	if err := comCall(unsafe.Pointer(enumPtr), 4, eCapture, uintptr(role), uintptr(unsafe.Pointer(&dev))); err != nil {
		return "", err
	}
	defer func() { _ = comRelease(unsafe.Pointer(dev)) }()

	var wstr uintptr
	if err := comCall(unsafe.Pointer(dev), 5, uintptr(unsafe.Pointer(&wstr))); err != nil {
		return "", err
	}
	defer func() { _, _, _ = procCoTaskMemFree.Call(wstr) }()
	return utf16PtrString(wstr), nil
}

// comCall invokes one COM method by vtable index with the object pointer as
// the implicit first argument; a nonzero HRESULT becomes an error.
func comCall(obj unsafe.Pointer, index int, args ...uintptr) error {
	vt := *(*uintptr)(obj)
	fn := *(*uintptr)(unsafe.Pointer(vt + uintptr(index)*unsafe.Sizeof(uintptr(0))))
	r1, _, _ := syscall.SyscallN(fn, append([]uintptr{uintptr(obj)}, args...)...)
	hr := uintptr(int32(r1)) //nolint:gosec // HRESULT sign extension to uintptr
	if hr != 0 {
		return syscall.Errno(hr)
	}
	return nil
}

func comRelease(obj unsafe.Pointer) error {
	return comCall(obj, 2)
}

func utf16PtrString(p uintptr) string {
	s := (*[1 << 16]uint16)(unsafe.Pointer(p))
	n := 0
	for s[n] != 0 {
		n++
	}
	return syscall.UTF16ToString(s[:n])
}
