package recorder

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
)

// WAV parameters matching the recorder output on every platform.
const (
	wavSampleRate = 16000
	wavChannels   = 1
	wavBits       = 16
	// WAVHeaderSize is the byte length of the canonical PCM header.
	WAVHeaderSize = 44
)

// WrapWAV prepends a canonical PCM WAV header to the raw little-endian
// samples in rawPath and writes the result to dst.
func WrapWAV(dst, rawPath string) error {
	raw, err := os.Open(rawPath)
	if err != nil {
		return err
	}
	defer func() { _ = raw.Close() }()
	info, err := raw.Stat()
	if err != nil {
		return err
	}
	if info.Size() > math.MaxUint32-36 {
		return errors.New("recording too large for WAV")
	}
	tmp := dst + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	//nolint:gosec // bounded by the MaxUint32 size guard above
	if _, err := out.Write(wavHeader(uint32(info.Size()))); err != nil {
		_ = out.Close()
		return err
	}
	if _, err := io.Copy(out, raw); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func wavHeader(dataSize uint32) []byte {
	const rate = uint32(wavSampleRate)
	const channels = uint16(wavChannels)
	const bits = uint16(wavBits)
	byteRate := rate * uint32(channels) * uint32(bits) / 8
	blockAlign := channels * bits / 8
	h := make([]byte, WAVHeaderSize)
	copy(h[0:], "RIFF")
	binary.LittleEndian.PutUint32(h[4:], 36+dataSize)
	copy(h[8:], "WAVE")
	copy(h[12:], "fmt ")
	binary.LittleEndian.PutUint32(h[16:], 16)
	binary.LittleEndian.PutUint16(h[20:], 1)
	binary.LittleEndian.PutUint16(h[22:], channels)
	binary.LittleEndian.PutUint32(h[24:], rate)
	binary.LittleEndian.PutUint32(h[28:], byteRate)
	binary.LittleEndian.PutUint16(h[32:], blockAlign)
	binary.LittleEndian.PutUint16(h[34:], bits)
	copy(h[36:], "data")
	binary.LittleEndian.PutUint32(h[40:], dataSize)
	return h
}
