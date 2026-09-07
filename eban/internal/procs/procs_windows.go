//go:build windows

package procs

import "syscall"

const (
	processQueryLimitedInfo = 0x1000
	stillActive             = 259
)

// Alive reports whether the process with the given pid exists.
func Alive(pid int) bool {
	h, err := syscall.OpenProcess(processQueryLimitedInfo, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)
	var code uint32
	if err := syscall.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	return code == stillActive
}
