package state

import (
	"os"
	"testing"
)

func TestStoreRecordingRoundTrip(t *testing.T) {
	t.Parallel()
	s := New(t.TempDir())

	if _, ok := s.RecordingPID(); ok {
		t.Fatal("expected no active recording on a fresh store")
	}
	if err := s.SaveRecording(os.Getpid(), ModePaste, "ru", "0x123"); err != nil {
		t.Fatalf("SaveRecording: %v", err)
	}
	if pid, ok := s.RecordingPID(); !ok || pid != os.Getpid() {
		t.Fatalf("RecordingPID = (%d, %v), want (%d, true)", pid, ok, os.Getpid())
	}
	if s.Mode() != ModePaste {
		t.Errorf("Mode = %q, want %q", s.Mode(), ModePaste)
	}
	if s.Lang() != "ru" {
		t.Errorf("Lang = %q, want %q", s.Lang(), "ru")
	}
	if s.PasteTarget() != "0x123" {
		t.Errorf("PasteTarget = %q, want %q", s.PasteTarget(), "0x123")
	}
	if s.Elapsed() < 0 {
		t.Errorf("Elapsed = %v, want non-negative", s.Elapsed())
	}

	s.ClearRecording()
	if _, ok := s.RecordingPID(); ok {
		t.Error("expected no active recording after ClearRecording")
	}
	if s.Mode() != ModeCopy {
		t.Errorf("Mode after clear = %q, want %q", s.Mode(), ModeCopy)
	}
	if s.PasteTarget() != "" {
		t.Errorf("PasteTarget after clear = %q, want empty", s.PasteTarget())
	}
}

func TestStorePasteEnterMode(t *testing.T) {
	t.Parallel()
	s := New(t.TempDir())
	if err := s.SaveRecording(os.Getpid(), ModePasteEnter, "", "0x123"); err != nil {
		t.Fatalf("SaveRecording: %v", err)
	}
	if got := s.Mode(); got != ModePasteEnter {
		t.Errorf("Mode() = %q, want %q", got, ModePasteEnter)
	}
}

func TestStoreSetModeOverwrites(t *testing.T) {
	t.Parallel()
	s := New(t.TempDir())
	if err := s.SaveRecording(os.Getpid(), ModePasteEnter, "", "0x123"); err != nil {
		t.Fatalf("SaveRecording: %v", err)
	}
	s.SetMode(ModeCopy)
	if got := s.Mode(); got != ModeCopy {
		t.Errorf("Mode() = %q, want %q", got, ModeCopy)
	}
}

func TestStoreState(t *testing.T) {
	t.Parallel()
	s := New(t.TempDir())

	if s.State() != StateIdle {
		t.Errorf("fresh State = %q, want %q", s.State(), StateIdle)
	}
	s.SetState(StateTranscribing)
	if s.State() != StateTranscribing {
		t.Errorf("State = %q, want %q", s.State(), StateTranscribing)
	}
	s.SetIndicator(IndicatorDone)
	if s.Indicator() != IndicatorDone {
		t.Errorf("Indicator = %q, want %q", s.Indicator(), IndicatorDone)
	}
}

func TestParseState(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  State
	}{
		{"recording", StateRecording},
		{"transcribing", StateTranscribing},
		{"idle", StateIdle},
		{"garbage", StateIdle},
		{"", StateIdle},
	}
	for _, tc := range cases {
		if got := ParseState(tc.input); got != tc.want {
			t.Errorf("ParseState(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseIndicator(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  IndicatorState
	}{
		{"recording", IndicatorRecording},
		{"transcribing", IndicatorTranscribing},
		{"done", IndicatorDone},
		{"idle", IndicatorIdle},
		{"garbage", IndicatorIdle},
		{"", IndicatorIdle},
	}
	for _, tc := range cases {
		if got := ParseIndicator(tc.input); got != tc.want {
			t.Errorf("ParseIndicator(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestRecordingPIDGarbage(t *testing.T) {
	t.Parallel()
	s := New(t.TempDir())
	if err := os.WriteFile(s.path("recording.pid"), []byte("not-a-pid"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.RecordingPID(); ok {
		t.Error("expected garbage pid file to report no active recording")
	}
}
