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

func TestMatchDefaultDevice(t *testing.T) {
	t.Parallel()
	devices := []DShowDevice{
		{Name: "Mic USB", Alt: "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\\wave_{441F2CB2-6750-4715-B75D-895788AEAB93}"},
		{Name: "Mic Steam", Alt: "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\\wave_{75687F12-C3E1-42F5-9342-C0EE22C67980}"},
		{Name: "Headset BT", Alt: "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\\wave_{E550EF6C-5CF2-4CE0-81B2-984B2B8F5836}"},
	}
	cases := []struct {
		name     string
		defaults []string
		want     int
	}{
		{
			name:     "console role matches second device",
			defaults: []string{"{0.0.1.00000000}.{75687f12-c3e1-42f5-9342-c0ee22c67980}"},
			want:     1,
		},
		{
			name:     "communications fallback consulted after console",
			defaults: []string{"{0.0.1.00000000}.{e550ef6c-5cf2-4ce0-81b2-984b2b8f5836}", "garbage"},
			want:     2,
		},
		{
			name:     "no guid match falls back to first device",
			defaults: []string{"{0.0.1.00000000}.{00000000-0000-0000-0000-000000000000}"},
			want:     0,
		},
		{name: "no defaults falls back to first device", want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := matchDefaultDevice(tc.defaults, devices); got != tc.want {
				t.Fatalf("matchDefaultDevice() = %d, want %d", got, tc.want)
			}
		})
	}
}
