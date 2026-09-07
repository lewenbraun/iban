package recorder

import (
	"os"
	"strings"
)

// DShowDevice is a DirectShow capture device entry.
type DShowDevice struct {
	Name string
	Alt  string
}

// dshowAudioInput resolves the ffmpeg dshow audio input specifier: the
// EBAN_DSHOW_AUDIO override wins, otherwise the first capture device's
// stable alternative name (or friendly name as a fallback).
func dshowAudioInput(devices []DShowDevice) string {
	if override := strings.TrimSpace(os.Getenv("EBAN_DSHOW_AUDIO")); override != "" {
		return override
	}
	if len(devices) == 0 {
		return ""
	}
	if devices[0].Alt != "" {
		return devices[0].Alt
	}
	return devices[0].Name
}

// parseDShowDevices extracts the audio capture devices from the stderr
// output of "ffmpeg -list_devices true -f dshow -i dummy". Video devices
// are ignored.
func parseDShowDevices(out []byte) []DShowDevice {
	var devices []DShowDevice
	audio := false
	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.Contains(line, "DirectShow audio devices"):
			audio = true
		case strings.Contains(line, "DirectShow video devices"):
			audio = false
		case !audio:
		case strings.Contains(line, "Alternative name"):
			if alt, ok := firstQuoted(line); ok && len(devices) > 0 {
				devices[len(devices)-1].Alt = alt
			}
		default:
			if name, ok := firstQuoted(line); ok {
				devices = append(devices, DShowDevice{Name: name})
			}
		}
	}
	return devices
}

func firstQuoted(line string) (string, bool) {
	start := strings.IndexByte(line, '"')
	if start < 0 {
		return "", false
	}
	end := strings.IndexByte(line[start+1:], '"')
	if end < 0 {
		return "", false
	}
	return line[start+1 : start+1+end], true
}
