package recorder

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
)

// PeakAmplitude returns the loudest absolute PCM sample in the WAV file at
// wavPath, or 0 when the file cannot be read. It skips the canonical header
// and scans the data chunk as little-endian 16-bit samples.
func PeakAmplitude(wavPath string) int {
	f, err := os.Open(wavPath)
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReaderSize(io.NewSectionReader(f, WAVHeaderSize, 1<<62), 64<<10)
	var peak int
	for {
		var buf [4096]byte
		n, err := r.Read(buf[:])
		for i := 0; i+1 < n; i += 2 {
			v := int(binary.LittleEndian.Uint16(buf[i : i+2]))
			if v > 32767 {
				v -= 65536
			}
			if v < 0 {
				v = -v
			}
			if v > peak {
				peak = v
			}
		}
		if err != nil {
			return peak
		}
	}
}
