package recorder

import (
	"os"
	"regexp"
	"strings"
)

// DShowDevice is a DirectShow capture device entry.
type DShowDevice struct {
	Name string
	Alt  string
}

var endpointGUID = regexp.MustCompile(`\{[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\}`)

// dshowAudioInput resolves the ffmpeg dshow audio input specifier: the
// EBAN_DSHOW_AUDIO override wins, otherwise the Windows default capture
// device when one of the default endpoint GUIDs matches, otherwise the
// first capture device.
func dshowAudioInput(devices []DShowDevice) string {
	if override := strings.TrimSpace(os.Getenv("EBAN_DSHOW_AUDIO")); override != "" {
		return override
	}
	if len(devices) == 0 {
		return ""
	}
	return pickDevice(devices, matchDefaultDevice(DefaultCaptureEndpointIDs(), devices))
}

func pickDevice(devices []DShowDevice, i int) string {
	if devices[i].Alt != "" {
		return devices[i].Alt
	}
	return devices[i].Name
}

// matchDefaultDevice returns the index of the device bound to the Windows
// default capture endpoint, or 0 when no endpoint GUID matches. WASAPI
// endpoint IDs and DirectShow alternative names carry the same device GUID,
// so matching on it is stable across rename and locale changes.
func matchDefaultDevice(defaults []string, devices []DShowDevice) int {
	for _, def := range defaults {
		for _, guid := range endpointGUID.FindAllString(def, -1) {
			guid = strings.ToLower(guid)
			for i, d := range devices {
				if strings.Contains(strings.ToLower(d.Alt+" "+d.Name), guid) {
					return i
				}
			}
		}
	}
	return 0
}

// parseDShowDevices extracts the audio capture devices from the stderr
// output of "ffmpeg -list_devices true -f dshow -i dummy". Video devices
// are ignored.
func parseDShowDevices(out []byte) []DShowDevice {
	var devices []DShowDevice
	audio := false
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.HasSuffix(strings.TrimSpace(line), "(audio)") {
			audio = true
		}
		if strings.HasSuffix(strings.TrimSpace(line), "(video)") {
			audio = false
		}
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
