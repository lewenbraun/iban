//go:build !windows

package recorder

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

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
