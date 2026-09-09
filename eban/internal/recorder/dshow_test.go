package recorder

import "testing"

func TestParseDShowDevicesModern(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		input string
	}{
		{name: "modern", input: `[in#0] "Microphone" (audio)
[in#0] Alternative name "audio-id"
[in#0] "Camera" (video)
[in#0] Alternative name "video-id"`},
		{name: "legacy", input: `[dshow] DirectShow audio devices
[dshow] "Microphone"
[dshow] Alternative name "audio-id"
[dshow] DirectShow video devices
[dshow] "Camera"
[dshow] Alternative name "video-id"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			devices := parseDShowDevices([]byte(tc.input))
			if len(devices) != 1 {
				t.Fatalf("devices = %+v, want one microphone", devices)
			}
			if devices[0].Name != "Microphone" || devices[0].Alt != "audio-id" {
				t.Fatalf("incorrect microphone: %+v", devices[0])
			}
		})
	}
}
