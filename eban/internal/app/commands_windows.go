//go:build windows

package app

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/lewenbraun/go-skeleton/eban/internal/state"
	"github.com/lewenbraun/go-skeleton/eban/internal/tray"
)

const trayDaemonArg = "--daemon"

func init() {
	commands["tray"] = cmdTray
}

func cmdTray(args []string) error {
	if len(args) == 1 && args[0] == trayDaemonArg {
		return tray.Run(newService(), state.New(defaultStateDir))
	}
	return spawnTrayDaemon()
}

func spawnTrayDaemon() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer func() { _ = devnull.Close() }()
	cmd := exec.Command(self, "tray", trayDaemonArg)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: detachedTrayFlags}
	cmd.Stdout = devnull
	cmd.Stderr = devnull
	return cmd.Start()
}
