//go:build !windows

package procs

import "syscall"

// Alive reports whether the process with the given pid exists.
func Alive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
