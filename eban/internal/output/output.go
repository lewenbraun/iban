// Package output delivers text to the user through clipboard and target-aware
// paste backends for Wayland and X11.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

type waylandPaster struct{}

type hyprlandWindow struct {
	Address string `json:"address"`
}

type hyprlandOption struct {
	Int int `json:"int"`
}

func (waylandPaster) CaptureTarget() (string, error) {
	return activeHyprlandWindow()
}

func activeHyprlandWindow() (string, error) {
	b, err := exec.Command("hyprctl", "activewindow", "-j").Output()
	if err != nil {
		return "", fmt.Errorf("read active hyprland window: %w", err)
	}
	return parseHyprlandTarget(b)
}

func parseHyprlandTarget(b []byte) (string, error) {
	var window hyprlandWindow
	if err := json.Unmarshal(b, &window); err != nil {
		return "", fmt.Errorf("parse active hyprland window: %w", err)
	}
	if !isHyprlandAddress(window.Address) {
		return "", errors.New("invalid active hyprland window address")
	}
	return window.Address, nil
}

func (waylandPaster) PasteTarget(target string) error {
	if !isHyprlandAddress(target) {
		return errors.New("invalid hyprland window address")
	}
	current, err := activeHyprlandWindow()
	if err != nil {
		return err
	}
	if current == target {
		return sendPasteShortcut(target)
	}
	noWarps, err := cursorNoWarps()
	if err != nil {
		return err
	}
	if err := setCursorNoWarps(true); err != nil {
		return err
	}
	return pasteWithFocus(target, current, noWarps)
}

func pasteWithFocus(target, current string, noWarps bool) error {
	if err := focusHyprlandWindow(target); err != nil {
		return errors.Join(err, setCursorNoWarps(noWarps))
	}
	pasteErr := sendPasteShortcut(target)
	restoreFocusErr := focusHyprlandWindow(current)
	restoreCursorErr := setCursorNoWarps(noWarps)
	return errors.Join(pasteErr, restoreFocusErr, restoreCursorErr)
}

func cursorNoWarps() (bool, error) {
	b, err := exec.Command("hyprctl", "getoption", "cursor:no_warps", "-j").Output()
	if err != nil {
		return false, fmt.Errorf("read hyprland cursor warping setting: %w", err)
	}
	return parseHyprlandBool(b)
}

func parseHyprlandBool(b []byte) (bool, error) {
	var option hyprlandOption
	if err := json.Unmarshal(b, &option); err != nil {
		return false, fmt.Errorf("parse hyprland option: %w", err)
	}
	if option.Int != 0 && option.Int != 1 {
		return false, errors.New("invalid hyprland boolean option")
	}
	return option.Int == 1, nil
}

func setCursorNoWarps(enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	if err := exec.Command("hyprctl", "keyword", "cursor:no_warps", value).Run(); err != nil {
		return fmt.Errorf("set hyprland cursor warping setting: %w", err)
	}
	return nil
}

func focusHyprlandWindow(target string) error {
	if err := exec.Command("hyprctl", "dispatch", "focuswindow", "address:"+target).Run(); err != nil {
		return fmt.Errorf("focus hyprland window: %w", err)
	}
	return nil
}

func sendPasteShortcut(target string) error {
	if err := exec.Command("hyprctl", "dispatch", "sendshortcut", pasteShortcut(target)).Run(); err != nil {
		return fmt.Errorf("send paste shortcut: %w", err)
	}
	return nil
}

type x11Paster struct{}

func (x11Paster) CaptureTarget() (string, error) {
	b, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return "", fmt.Errorf("read active x11 window: %w", err)
	}
	target := strings.TrimSpace(string(b))
	if !isX11Window(target) {
		return "", errors.New("invalid active x11 window")
	}
	return target, nil
}

func (x11Paster) PasteTarget(target string) error {
	if !isX11Window(target) {
		return errors.New("invalid x11 window")
	}
	return exec.Command("xdotool", "key", "--window", target, "ctrl+v").Run()
}

// NewCopier returns the clipboard backend for the current session.
func NewCopier() Copier {
	if isWayland() {
		return waylandCopier{}
	}
	return x11Copier{}
}

// NewPaster returns the target-aware paste backend for the current session.
func NewPaster() Paster {
	if isWayland() {
		return waylandPaster{}
	}
	return x11Paster{}
}

func isHyprlandAddress(value string) bool {
	if !strings.HasPrefix(value, "0x") || len(value) == 2 {
		return false
	}
	for _, r := range value[2:] {
		if !isHexDigit(r) {
			return false
		}
	}
	return true
}

func isHexDigit(r rune) bool {
	return r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
}

func isX11Window(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func pasteShortcut(target string) string {
	return "CTRL,code:55,address:" + target
}
