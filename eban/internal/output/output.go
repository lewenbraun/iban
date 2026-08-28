// Package output delivers text to the user: clipboard, typing into the
// focused window and desktop notifications, with Wayland/X11 backends.
package output

import (
	"os"
	"os/exec"
	"strings"
)

// Notification urgency levels understood by notify-send.
const (
	UrgencyNormal   = "normal"
	UrgencyCritical = "critical"
)

func isWayland() bool {
	return os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "wayland"
}

type waylandCopier struct{}

func (waylandCopier) Copy(text string) error {
	cmd := exec.Command("wl-copy", "--type", "text/plain")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

type x11Copier struct{}

func (x11Copier) Copy(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

type waylandTyper struct{}

func (waylandTyper) Type(text string) error {
	return exec.Command("wtype", "--", text).Run()
}

type x11Typer struct{}

func (x11Typer) Type(text string) error {
	return exec.Command("xdotool", "type", "--delay", "0", "--", text).Run()
}

type notifySender struct{}

func (notifySender) Notify(body, urgency string) {
	_ = exec.Command("notify-send", "-a", "eban", "-u", urgency, "-t", "3000", "Eban", body).Run()
}

// NewCopier returns the clipboard backend for the current session.
func NewCopier() Copier {
	if isWayland() {
		return waylandCopier{}
	}
	return x11Copier{}
}

// NewTyper returns the keyboard-typing backend for the current session.
func NewTyper() Typer {
	if isWayland() {
		return waylandTyper{}
	}
	return x11Typer{}
}

// NewNotifier returns the desktop notification backend.
func NewNotifier() Notifier {
	return notifySender{}
}
