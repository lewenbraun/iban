package state

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Store persists dictation session state as files in a directory.
type Store struct {
	dir string
}

// New creates a Store rooted at dir.
func New(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) path(name string) string {
	return filepath.Join(s.dir, name)
}

// WavPath is the fixed location the recorder writes audio to.
func (s *Store) WavPath() string {
	return s.path("recording.wav")
}

// SetState writes the current state machine value.
func (s *Store) SetState(v State) {
	_ = os.WriteFile(s.path("state"), []byte(v), 0o600)
}

// SetIndicator writes the current indicator state.
func (s *Store) SetIndicator(v IndicatorState) {
	_ = os.WriteFile(s.path("indicator"), []byte(v), 0o600)
}

// State reads the current state machine value.
func (s *Store) State() State {
	b, err := os.ReadFile(s.path("state"))
	if err != nil {
		return StateIdle
	}
	return ParseState(string(b))
}

// ParseState normalizes raw state file contents into a valid State.
func ParseState(raw string) State {
	v := State(strings.TrimSpace(raw))
	if v == StateRecording || v == StateTranscribing {
		return v
	}
	return StateIdle
}

// Indicator reads the current indicator state.
func (s *Store) Indicator() IndicatorState {
	b, err := os.ReadFile(s.path("indicator"))
	if err != nil {
		return IndicatorIdle
	}
	return ParseIndicator(string(b))
}

// ParseIndicator normalizes raw indicator data into a valid state.
func ParseIndicator(raw string) IndicatorState {
	v := IndicatorState(strings.TrimSpace(raw))
	if v == IndicatorRecording || v == IndicatorTranscribing || v == IndicatorDone {
		return v
	}
	return IndicatorIdle
}

// SaveRecording persists the active recording session metadata.
func (s *Store) SaveRecording(pid int, mode Mode, lang, target string) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	entries := []struct{ path, data string }{
		{s.path("recording.pid"), strconv.Itoa(pid)},
		{s.path("mode"), string(mode)},
		{s.path("lang"), lang},
		{s.path("target"), target},
		{s.path("started_at"), strconv.FormatInt(time.Now().UnixNano(), 10)},
	}
	for _, e := range entries {
		if err := os.WriteFile(e.path, []byte(e.data), 0o600); err != nil {
			return err
		}
	}
	return nil
}

// RecordingPID returns the pid of the active recording, if any.
func (s *Store) RecordingPID() (int, bool) {
	b, err := os.ReadFile(s.path("recording.pid"))
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	if syscall.Kill(pid, 0) != nil {
		return 0, false
	}
	return pid, true
}

// Mode returns the delivery mode saved for the active session.
func (s *Store) Mode() Mode {
	b, err := os.ReadFile(s.path("mode"))
	if err != nil {
		return ModeCopy
	}
	mode := Mode(strings.TrimSpace(string(b)))
	if mode == ModePaste || mode == ModePasteEnter {
		return mode
	}
	return ModeCopy
}

// Lang returns the language hint saved for the active session.
func (s *Store) Lang() string {
	b, err := os.ReadFile(s.path("lang"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// PasteTarget returns the saved destination window for the active session.
func (s *Store) PasteTarget() string {
	b, err := os.ReadFile(s.path("target"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// Elapsed returns how long the active recording has been running.
func (s *Store) Elapsed() time.Duration {
	b, err := os.ReadFile(s.path("started_at"))
	if err != nil {
		return 0
	}
	ns, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	if err != nil {
		return 0
	}
	return time.Since(time.Unix(0, ns))
}

// ClearRecording removes the session files and resets the state.
func (s *Store) ClearRecording() {
	for _, name := range []string{"recording.pid", "mode", "lang", "target", "started_at"} {
		_ = os.Remove(s.path(name))
	}
	s.SetState(StateIdle)
}
