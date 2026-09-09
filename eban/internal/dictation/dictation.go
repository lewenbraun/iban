// Package dictation orchestrates the record -> transcribe -> deliver flow
// and defines the seams (consumer-side interfaces) every integration plugs
// into: recorder, transcriber, clipboard, and target-aware paste.
package dictation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lewenbraun/eban/eban/internal/output"
	"github.com/lewenbraun/eban/eban/internal/recorder"
	"github.com/lewenbraun/eban/eban/internal/state"
)

// Recorder captures audio into a WAV file as a detached process.
type Recorder interface {
	Start(wavPath string) (int, error)
	Stop(pid int)
}

// Transcriber converts an audio file into text.
type Transcriber interface {
	Transcribe(ctx context.Context, audioPath, lang string) (string, error)
}

// TranscriberFactory builds a Transcriber lazily at stop time.
type TranscriberFactory func() (Transcriber, error)

// Config carries the tunable timeouts and paths of the service.
type Config struct {
	StateDir          string
	DefaultTimeout    time.Duration
	MaxTimeout        time.Duration
	MinRecording      time.Duration
	TranscribeTimeout time.Duration
}

// DefaultConfig returns the production defaults for the service.
func DefaultConfig(stateDir string) Config {
	return Config{
		StateDir:          stateDir,
		DefaultTimeout:    10 * time.Minute,
		MaxTimeout:        time.Hour,
		MinRecording:      400 * time.Millisecond,
		TranscribeTimeout: 2 * time.Minute,
	}
}

// Service orchestrates the record-transcribe-deliver flow.
type Service struct {
	store          *state.Store
	recorder       Recorder
	newTranscriber TranscriberFactory
	copier         output.Copier
	paster         output.Paster
	defaultTimeout time.Duration
	maxTimeout     time.Duration
	minRecording   time.Duration
	transcribeWait time.Duration
}

// New wires the service with its integration seams.
func New(cfg Config, rec Recorder, tf TranscriberFactory, c output.Copier, p output.Paster) *Service {
	return &Service{
		store:          state.New(cfg.StateDir),
		recorder:       rec,
		newTranscriber: tf,
		copier:         c,
		paster:         p,
		defaultTimeout: cfg.DefaultTimeout,
		maxTimeout:     cfg.MaxTimeout,
		minRecording:   cfg.MinRecording,
		transcribeWait: cfg.TranscribeTimeout,
	}
}

// Active reports whether a recording session is running.
func (s *Service) Active() bool {
	_, ok := s.store.RecordingPID()
	return ok
}

// Status returns the current state machine value.
func (s *Service) Status() state.State {
	return s.store.State()
}

// Start begins a recording session in the given mode.
func (s *Service) Start(mode state.Mode, lang string, timeout time.Duration) error {
	if s.Active() {
		return errors.New("recording is already active")
	}
	target, err := s.captureTarget(mode)
	if err != nil {
		return err
	}
	timeout = s.clampTimeout(timeout)
	pid, err := s.recorder.Start(s.store.WavPath())
	if err != nil {
		return err
	}
	if err := s.store.SaveRecording(pid, mode, lang, target); err != nil {
		s.recorder.Stop(pid)
		return fmt.Errorf("save recording state: %w", err)
	}
	recorder.SpawnWatchdog(pid, timeout)
	s.store.SetState(state.StateRecording)
	s.store.SetIndicator(state.IndicatorRecording)
	return nil
}

// Stop ends the session, transcribes and delivers the result.
func (s *Service) Stop(lang string, paste, pressEnter bool) error {
	pid, active := s.store.RecordingPID()
	if !active {
		return errors.New("no active recording")
	}
	mode := s.store.Mode()
	target := s.store.PasteTarget()
	if lang == "" {
		lang = s.store.Lang()
	}
	elapsed := s.store.Elapsed()
	s.recorder.Stop(pid)
	s.store.ClearRecording()
	if elapsed < s.minRecording {
		s.store.SetIndicator(state.IndicatorIdle)
		return fmt.Errorf("recording too short (min %d ms)", s.minRecording.Milliseconds())
	}
	if err := s.requireAudio(); err != nil {
		return s.fail(err)
	}
	s.store.SetState(state.StateTranscribing)
	s.store.SetIndicator(state.IndicatorTranscribing)
	paste = paste || mode == state.ModePaste || mode == state.ModePasteEnter
	pressEnter = pressEnter || mode == state.ModePasteEnter
	return s.transcribeAndDeliver(lang, paste, pressEnter, target)
}

func (s *Service) transcribeAndDeliver(lang string, paste, pressEnter bool, target string) error {
	text, err := s.transcribe(lang)
	if err != nil {
		return s.fail(err)
	}
	text, pasteCommand := SplitPasteCommand(text)
	if text == "" {
		return s.fail(errors.New("empty transcript"))
	}
	if err := s.copier.Copy(text); err != nil {
		return s.fail(fmt.Errorf("copy to clipboard: %w", err))
	}
	if paste || pasteCommand {
		if err := s.pasteTarget(target, pressEnter); err != nil {
			return s.fail(err)
		}
	}
	s.store.SetState(state.StateIdle)
	s.store.SetIndicator(state.IndicatorDone)
	return nil
}

func (s *Service) captureTarget(mode state.Mode) (string, error) {
	if s.paster == nil {
		if mode == state.ModePaste || mode == state.ModePasteEnter {
			return "", errors.New("paste target unavailable")
		}
		return "", nil
	}
	target, err := s.paster.CaptureTarget()
	if err == nil {
		return target, nil
	}
	if mode != state.ModePaste {
		return "", nil
	}
	return "", fmt.Errorf("capture paste target: %w", err)
}

func (s *Service) pasteTarget(target string, pressEnter bool) error {
	if target == "" {
		return errors.New("missing paste target")
	}
	if s.paster == nil {
		return errors.New("paste target unavailable")
	}
	if err := s.paster.PasteTarget(target, pressEnter); err != nil {
		return fmt.Errorf("paste into saved target: %w", err)
	}
	return nil
}

func (s *Service) fail(err error) error {
	s.store.SetState(state.StateIdle)
	s.store.SetIndicator(state.IndicatorIdle)
	return err
}

func (s *Service) transcribe(lang string) (string, error) {
	tr, err := s.newTranscriber()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.transcribeWait)
	defer cancel()
	return tr.Transcribe(ctx, s.store.WavPath(), lang)
}

func (s *Service) requireAudio() error {
	info, err := os.Stat(s.store.WavPath())
	if err != nil || info.Size() <= recorder.WAVHeaderSize {
		return errors.New("no audio captured (check microphone and the ffmpeg log next to the recording)")
	}
	return nil
}

func (s *Service) clampTimeout(d time.Duration) time.Duration {
	if d <= 0 {
		return s.defaultTimeout
	}
	return min(d, s.maxTimeout)
}
