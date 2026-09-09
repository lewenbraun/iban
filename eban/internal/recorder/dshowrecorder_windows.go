//go:build windows

package recorder

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

const (
	rawSuffix       = ".raw"
	ffmpegLogSuffix = ".ffmpeg.log"
	detachedProcess = 0x00000008
	detachedFlags   = syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess
)

// DShowRecorder records 16 kHz mono audio through ffmpeg's dshow input.
type DShowRecorder struct {
	mu  sync.Mutex
	wav string
	raw string
}

// New returns a DirectShow-backed recorder.
func New() *DShowRecorder {
	return &DShowRecorder{}
}

// Start launches a detached ffmpeg process capturing raw PCM samples.
func (r *DShowRecorder) Start(wavPath string) (int, error) {
	if err := os.MkdirAll(filepath.Dir(wavPath), 0o700); err != nil {
		return 0, err
	}
	device := audioDevice()
	if device == "" {
		return 0, errors.New("no dshow audio device found (install ffmpeg or set EBAN_DSHOW_AUDIO)")
	}
	_ = os.Remove(wavPath)
	rawPath := wavPath + rawSuffix
	_ = os.Remove(rawPath)
	pid, err := launchFFmpeg(device, rawPath)
	if err != nil {
		return 0, err
	}
	r.mu.Lock()
	r.wav, r.raw = wavPath, rawPath
	r.mu.Unlock()
	return pid, nil
}

// launchFFmpeg starts the detached capture process; its stderr goes to a
// truncating log beside the raw file so device failures stay diagnosable.
func launchFFmpeg(device, rawPath string) (int, error) {
	cmd := exec.Command("ffmpeg", dshowArgs(device, rawPath)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: detachedFlags}
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return 0, err
	}
	defer func() { _ = devnull.Close() }()
	ffmpegLog, err := os.OpenFile(rawPath+ffmpegLogSuffix, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return 0, err
	}
	defer func() { _ = ffmpegLog.Close() }()
	cmd.Stdout = devnull
	cmd.Stderr = ffmpegLog
	if err := cmd.Start(); err != nil {
		return 0, errors.New("ffmpeg not available")
	}
	go func() { _ = cmd.Wait() }()
	return cmd.Process.Pid, nil
}

// Stop terminates ffmpeg and wraps the raw samples into a WAV file.
func (r *DShowRecorder) Stop(pid int) {
	killAndWait(pid)
	r.mu.Lock()
	wav, raw := r.wav, r.raw
	r.wav, r.raw = "", ""
	r.mu.Unlock()
	if wav != "" {
		_ = WrapWAV(wav, raw)
	}
}

func dshowArgs(device, rawPath string) []string {
	return []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-f", "dshow", "-i", "audio=" + device,
		"-ar", "16000", "-ac", "1", "-f", "s16le", rawPath,
	}
}

func audioDevice() string {
	var devices []DShowDevice
	if out, err := exec.Command("ffmpeg", "-hide_banner", "-list_devices", "true", "-f", "dshow", "-i", "dummy").CombinedOutput(); err == nil || len(out) > 0 {
		devices = parseDShowDevices(out)
	}
	return dshowAudioInput(devices)
}
