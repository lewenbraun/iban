package recorder

import (
	"os"
	"path/filepath"
	"testing"
)

const dshowSample = `[dshow @ 000001f3e4a2f340]  "DirectShow video devices (some may be both video and audio devices)"
[dshow @ 000001f3e4a2f340]  "USB2.0 HD UVC WebCam" (video)
[dshow @ 000001f3e4a2f340]     Alternative name "@device_pnp_\\?\usb#vid_046d&pid_0825&mi_00#6&16f4ad2&0&0000#{65e8773d-8f56-11d0-a3b9-00a0c9223196}\global"
[dshow @ 000001f3e4a2f340]  "DirectShow audio devices (some may be both video and audio devices)"
[dshow @ 000001f3e4a2f340]  "Microphone (Realtek(R) Audio)" (audio)
[dshow @ 000001f3e4a2f340]     Alternative name "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\{9B9F4AC0-1E5B-4E36-A12C-6C1FDFE2A476}"
[dshow @ 000001f3e4a2f340]  " Stereo Mix (Realtek(R) Audio)" (audio)
[dshow @ 000001f3e4a2f340]     Alternative name "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\{ABCD1234-1E5B-4E36-A12C-6C1FDFE2A476}"
dummy: Immediate exit requested
`

func TestParseDShowDevices(t *testing.T) {
	t.Parallel()
	devices := parseDShowDevices([]byte(dshowSample))
	if len(devices) != 2 {
		t.Fatalf("got %d devices, want 2: %+v", len(devices), devices)
	}
	wantName := "Microphone (Realtek(R) Audio)"
	if devices[0].Name != wantName {
		t.Errorf("device name = %q, want %q", devices[0].Name, wantName)
	}
	wantAlt := "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\\{9B9F4AC0-1E5B-4E36-A12C-6C1FDFE2A476}"
	if devices[0].Alt != wantAlt {
		t.Errorf("device alt = %q, want %q", devices[0].Alt, wantAlt)
	}
	if devices[1].Name != " Stereo Mix (Realtek(R) Audio)" {
		t.Errorf("second device name = %q", devices[1].Name)
	}
}

func TestParseDShowDevicesEmpty(t *testing.T) {
	t.Parallel()
	if devices := parseDShowDevices(nil); len(devices) != 0 {
		t.Errorf("got %+v, want none", devices)
	}
}

func TestDShowAudioInputPrefersAltName(t *testing.T) {
	t.Parallel()
	devices := parseDShowDevices([]byte(dshowSample))
	if got := dshowAudioInput(devices); got != devices[0].Alt {
		t.Errorf("dshowAudioInput() = %q, want %q", got, devices[0].Alt)
	}
}

func TestDShowAudioInputOverride(t *testing.T) {
	t.Setenv("IBAN_DSHOW_AUDIO", "My Mic")
	if got := dshowAudioInput(nil); got != "My Mic" {
		t.Errorf("dshowAudioInput() = %q, want override", got)
	}
}

func TestWrapWAV(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rawPath := filepath.Join(dir, "recording.raw")
	samples := make([]byte, 3200)
	if err := os.WriteFile(rawPath, samples, 0o600); err != nil {
		t.Fatalf("write raw: %v", err)
	}
	dst := filepath.Join(dir, "recording.wav")
	if err := WrapWAV(dst, rawPath); err != nil {
		t.Fatalf("WrapWAV: %v", err)
	}
	out, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read wav: %v", err)
	}
	if len(out) != 44+len(samples) {
		t.Fatalf("wav size = %d, want %d", len(out), 44+len(samples))
	}
	if string(out[0:4]) != "RIFF" || string(out[8:12]) != "WAVE" {
		t.Errorf("missing RIFF/WAVE magic: %q", out[0:12])
	}
	if got := le16(out, 22); got != wavChannels {
		t.Errorf("channels = %d, want %d", got, wavChannels)
	}
	if got := le32(out, 24); got != wavSampleRate {
		t.Errorf("sample rate = %d, want %d", got, wavSampleRate)
	}
	if got := le32(out, 40); int(got) != len(samples) {
		t.Errorf("data size = %d, want %d", got, len(samples))
	}
}

func le16(b []byte, off int) uint16 {
	return uint16(b[off]) | uint16(b[off+1])<<8
}

func le32(b []byte, off int) uint32 {
	return uint32(b[off]) | uint32(b[off+1])<<8 | uint32(b[off+2])<<16 | uint32(b[off+3])<<24
}
