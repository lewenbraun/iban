//go:build !windows

package recorder

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/lewenbraun/eban/eban/internal/procs"
)

const stopGrace = 3 * time.Second

// StopPID interrupts a recording process gracefully and falls back to
// SIGKILL after the grace period expires.
func StopPID(pid int) {
	_ = syscall.Kill(pid, syscall.SIGINT)
	deadline := time.Now().Add(stopGrace)
	for time.Now().Before(deadline) {
		if !procs.Alive(pid) {
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
}

// Alive reports whether the process with the given pid exists.
func Alive(pid int) bool {
	return procs.Alive(pid)
}

// SpawnWatchdog re-executes the binary detached to stop the recording
// process after the timeout passes.
func SpawnWatchdog(pid int, timeout time.Duration) {
	self, err := os.Executable()
	if err != nil {
		return
	}
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return
	}
	defer func() { _ = devnull.Close() }()
	cmd := exec.Command(self, "watchdog", strconv.Itoa(pid), strconv.Itoa(int(timeout.Seconds())))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = devnull
	cmd.Stderr = devnull
	_ = cmd.Start()
}
