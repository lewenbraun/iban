//go:build windows

package tray

import (
	"syscall"
	"unsafe"
)

const (
	iconSize     = 32
	iconBytesPP  = 4
	maskBytesPP  = 1
	trueColor    = 32
	monochrome   = 1
	circleRadius = 12.5
)

var (
	procCreateBitmap       = syscall.NewLazyDLL("gdi32.dll").NewProc("CreateBitmap")
	procDeleteObject       = syscall.NewLazyDLL("gdi32.dll").NewProc("DeleteObject")
	procCreateIconIndirect = syscall.NewLazyDLL("user32.dll").NewProc("CreateIconIndirect")
	procDestroyIcon        = syscall.NewLazyDLL("user32.dll").NewProc("DestroyIcon")
)

type iconInfo struct {
	fIcon    int32
	xHotspot uint32
	yHotspot uint32
	hbmMask  uintptr
	hbmColor uintptr
}

// solidIcon builds a 32x32 tray icon with an opaque circle of the given
// 0x00BBGGRR color on a transparent background.
func solidIcon(color uint32) uintptr {
	const center = float64(iconSize-1) / 2
	stride := iconSize * iconBytesPP
	pixels := make([]byte, iconSize*stride)
	for y := 0; y < iconSize; y++ {
		for x := 0; x < iconSize; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			if dx*dx+dy*dy > circleRadius*circleRadius {
				continue
			}
			off := y*stride + x*iconBytesPP
			pixels[off] = byte(color)
			pixels[off+1] = byte(color >> 8)
			pixels[off+2] = byte(color >> 16)
			pixels[off+3] = 0xFF
		}
	}
	//nolint:gosec // CreateBitmap reads the pixel buffers by pointer
	colorBM, _, _ := procCreateBitmap.Call(iconSize, iconSize, 1, trueColor, uintptr(unsafe.Pointer(&pixels[0])))
	mask := make([]byte, iconSize*iconSize*maskBytesPP)
	for i := range mask {
		mask[i] = 0xFF
	}
	//nolint:gosec // CreateBitmap reads the mask buffer by pointer
	maskBM, _, _ := procCreateBitmap.Call(iconSize, iconSize, 1, monochrome, uintptr(unsafe.Pointer(&mask[0])))
	ii := iconInfo{fIcon: 1, hbmMask: maskBM, hbmColor: colorBM}
	//nolint:gosec // CreateIconIndirect reads ICONINFO by pointer
	icon, _, _ := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))
	_, _, _ = procDeleteObject.Call(colorBM)
	_, _, _ = procDeleteObject.Call(maskBM)
	return icon
}

// destroyIcon releases an icon created by solidIcon.
func destroyIcon(icon uintptr) {
	if icon != 0 {
		_, _, _ = procDestroyIcon.Call(icon)
	}
}
