//go:build windows

package app

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/lewenbraun/eban/eban/internal/state"
	"github.com/lewenbraun/eban/eban/internal/tray"
)

const trayDaemonArg = "--daemon"

func init() {
	commands["tray"] = cmdTray
}

func cmdTray(args []string) error {
	if len(args) == 1 && args[0] == trayDaemonArg {
		file, err := openTrayLog()
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		log.SetOutput(file)
		return tray.Run(newService(), state.New(defaultStateDir))
	}
	return spawnTrayDaemon()
}

func openTrayLog() (*os.File, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, "Eban")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath.Join(dir, "eban.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
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
