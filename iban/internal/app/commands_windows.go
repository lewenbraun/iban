//go:build windows

package app

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/lewenbraun/iban/iban/internal/state"
	"github.com/lewenbraun/iban/iban/internal/tray"
)

const trayDaemonArg = "--daemon"

var procCreateMutexW = syscall.NewLazyDLL("kernel32.dll").NewProc("CreateMutexW")

func init() {
	commands["tray"] = cmdTray
}

func cmdTray(args []string) error {
	if len(args) == 1 && args[0] == trayDaemonArg {
		if err := acquireTrayMutex(); err != nil {
			return err
		}
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

// acquireTrayMutex reserves the single tray slot for this session. The mutex
// handle lives until process exit, so a second tray daemon refuses to start
// instead of installing a duplicate keyboard hook over the same state files.
func acquireTrayMutex() error {
	name, _ := syscall.UTF16PtrFromString(`Local\iban-tray`)
	handle, _, err := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		return err
	}
	if errors.Is(err, syscall.ERROR_ALREADY_EXISTS) {
		return errors.New("tray already running")
	}
	return nil
}

func openTrayLog() (*os.File, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, "Iban")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath.Join(dir, "iban.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
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
