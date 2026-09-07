//go:build windows

package recorder

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/lewenbraun/go-skeleton/eban/internal/procs"
)

const (
	stopGrace    = 3 * time.Second
	stateDirName = "eban"
	wavName      = "recording.wav"
)

// StopPID terminates the recording process and finalizes the default
// recording file so the WAV is valid even after a watchdog kill.
func StopPID(pid int) {
	killAndWait(pid)
	finalizeDefaultRecording()
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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: detachedFlags}
	cmd.Stdout = devnull
	cmd.Stderr = devnull
	_ = cmd.Start()
}

func killAndWait(pid int) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	_ = proc.Kill()
	_, _ = proc.Wait()
	deadline := time.Now().Add(stopGrace)
	for time.Now().Before(deadline) && procs.Alive(pid) {
		time.Sleep(30 * time.Millisecond)
	}
}

func finalizeDefaultRecording() {
	wav := filepath.Join(os.TempDir(), stateDirName, wavName)
	_ = WrapWAV(wav, wav+rawSuffix)
}
