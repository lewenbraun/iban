package app

import (
	"os"
	"strconv"
	"time"

	"github.com/lewenbraun/go-skeleton/eban/internal/recorder"
)

// RunWatchdog sleeps for the requested timeout and then stops the recording
// process if it is still alive. It runs as a detached re-exec of the binary.
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
	if recorder.Alive(pid) {
		recorder.StopPID(pid)
	}
	os.Exit(0)
}
