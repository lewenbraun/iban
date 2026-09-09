package app

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lewenbraun/eban/eban/internal/recorder"
)

// RunWatchdog sleeps for the requested timeout and then stops the recording
// process if it is still active. It runs as a detached re-exec of the binary.
// A recording ended before the timeout clears its pid file, so a late
// watchdog stands down instead of killing an unrelated recycled pid and
// clobbering the next session's audio.
func RunWatchdog(args []string) {
	if len(args) != 2 {
		os.Exit(0)
	}
	pid, pidErr := strconv.Atoi(args[0])
	sec, secErr := strconv.Atoi(args[1])
	if pidErr != nil || secErr != nil || pid <= 0 || sec <= 0 {
		os.Exit(0)
	}
	time.Sleep(time.Duration(sec) * time.Second)
	if !ownsRecording(pid) {
		os.Exit(0)
	}
	if recorder.Alive(pid) {
		recorder.StopPID(pid)
	}
	os.Exit(0)
}

// ownsRecording reports whether the recording pid file still names this
// watchdog's process.
func ownsRecording(pid int) bool {
	path := filepath.Join(defaultStateDir, "recording.pid")
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	current, err := strconv.Atoi(strings.TrimSpace(string(b)))
	return err == nil && current == pid
}
