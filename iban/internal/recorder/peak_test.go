package recorder

import (
	"os"
	"path/filepath"
	"testing"
)

func writeWAV(t *testing.T, samples []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "recording.wav")
	header := make([]byte, WAVHeaderSize)
	if err := os.WriteFile(path, append(header, samples...), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPeakAmplitude(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		samples []byte
		want    int
	}{
		{name: "silence", samples: make([]byte, 8000), want: 0},
		{name: "single loud sample", samples: []byte{0x00, 0x7F}, want: 0x7F00},
		{name: "negative peak", samples: []byte{0x00, 0x80, 0x01, 0x00}, want: 0x8000},
		{name: "empty data", samples: nil, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := PeakAmplitude(writeWAV(t, tc.samples))
			if got != tc.want {
				t.Fatalf("PeakAmplitude() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPeakAmplitudeMissingFile(t *testing.T) {
	t.Parallel()
	if got := PeakAmplitude(filepath.Join(t.TempDir(), "absent.wav")); got != 0 {
		t.Fatalf("PeakAmplitude() = %d, want 0", got)
	}
}
