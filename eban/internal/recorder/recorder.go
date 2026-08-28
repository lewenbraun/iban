// Package recorder captures microphone audio through pw-record as a
// detached process and manages its lifecycle plus a watchdog re-exec.
package recorder

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

const stopGrace = 3 * time.Second

// PWRecorder records 16 kHz mono WAV files via PipeWire.
type PWRecorder struct{}

// New returns a PipeWire-backed recorder.
func New() *PWRecorder {
	return &PWRecorder{}
}

// Start launches a detached pw-record process writing to wavPath.
func (r *PWRecorder) Start(wavPath string) (int, error) {
	_ = os.Remove(wavPath)
	cmd := exec.Command("pw-record", "--rate", "16000", "--channels", "1", wavPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return 0, errors.New("pw-record not available")
	}
	go func() { _ = cmd.Wait() }()
	return cmd.Process.Pid, nil
}

// Stop interrupts the recording process gracefully.
func (r *PWRecorder) Stop(pid int) {
	StopPID(pid)
}

// StopPID interrupts a recording process gracefully and falls back to
// SIGKILL after the grace period expires.
func StopPID(pid int) {
	_ = syscall.Kill(pid, syscall.SIGINT)
	deadline := time.Now().Add(stopGrace)
	for time.Now().Before(deadline) {
		if !Alive(pid) {
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
}

// Alive reports whether the process with the given pid exists.
func Alive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
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
