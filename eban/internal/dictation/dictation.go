// Package dictation orchestrates the record -> transcribe -> deliver flow
// and defines the seams (consumer-side interfaces) every integration plugs
// into: recorder, transcriber, clipboard, typer and notifier.
package dictation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lewenbraun/go-skeleton/eban/internal/output"
	"github.com/lewenbraun/go-skeleton/eban/internal/recorder"
	"github.com/lewenbraun/go-skeleton/eban/internal/state"
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
	typer          output.Typer
	notifier       output.Notifier
	defaultTimeout time.Duration
	maxTimeout     time.Duration
	minRecording   time.Duration
	transcribeWait time.Duration
}

// New wires the service with its integration seams.
func New(cfg Config, rec Recorder, tf TranscriberFactory, c output.Copier, t output.Typer, n output.Notifier) *Service {
	return &Service{
		store:          state.New(cfg.StateDir),
		recorder:       rec,
		newTranscriber: tf,
		copier:         c,
		typer:          t,
		notifier:       n,
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
	timeout = s.clampTimeout(timeout)
	pid, err := s.recorder.Start(s.store.WavPath())
	if err != nil {
		return err
	}
	if err := s.store.SaveRecording(pid, mode, lang); err != nil {
		s.recorder.Stop(pid)
		return fmt.Errorf("save recording state: %w", err)
	}
	recorder.SpawnWatchdog(pid, timeout)
	s.store.SetState(state.StateRecording)
	s.notifier.Notify("recording started", output.UrgencyNormal)
	return nil
}

// Stop ends the session, transcribes and delivers the result.
func (s *Service) Stop(lang string, paste bool) error {
	pid, active := s.store.RecordingPID()
	if !active {
		return errors.New("no active recording")
	}
	mode := s.store.Mode()
	if lang == "" {
		lang = s.store.Lang()
	}
	elapsed := s.store.Elapsed()
	s.recorder.Stop(pid)
	s.store.ClearRecording()
	if elapsed < s.minRecording {
		return fmt.Errorf("recording too short (min %d ms)", s.minRecording.Milliseconds())
	}
	s.store.SetState(state.StateTranscribing)
	s.notifier.Notify("transcribing...", output.UrgencyNormal)
	return s.transcribeAndDeliver(lang, paste || mode == state.ModePaste)
}

func (s *Service) transcribeAndDeliver(lang string, paste bool) error {
	text, err := s.transcribe(lang)
	if err != nil {
		return err
	}
	text, pasteCommand := SplitPasteCommand(text)
	if text == "" {
		return errors.New("empty transcript")
	}
	if err := s.copier.Copy(text); err != nil {
		return fmt.Errorf("copy to clipboard: %w", err)
	}
	if paste || pasteCommand {
		if err := s.typer.Type(text); err != nil {
			return fmt.Errorf("type text: %w", err)
		}
	}
	s.notifier.Notify("done: "+truncate(text, previewLimit), output.UrgencyNormal)
	s.store.SetState(state.StateIdle)
	return nil
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

func (s *Service) clampTimeout(d time.Duration) time.Duration {
	if d <= 0 {
		return s.defaultTimeout
	}
	return min(d, s.maxTimeout)
}
