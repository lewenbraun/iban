package dictation

import (
	"context"
	"errors"
	"testing"

	"github.com/lewenbraun/iban/iban/internal/output"
	"github.com/lewenbraun/iban/iban/internal/state"
)

const (
	testText   = "hello"
	testTarget = "0x123"
)

type testTranscriber struct {
	text string
	err  error
}

func (t testTranscriber) Transcribe(context.Context, string, string) (string, error) {
	return t.text, t.err
}

type testCopier struct {
	err error
}

func (t testCopier) Copy(string) error {
	return t.err
}

type testPaster struct {
	target       string
	captureErr   error
	pasteErr     error
	pastedTarget string
	pressedEnter bool
}

func (p *testPaster) CaptureTarget() (string, error) {
	return p.target, p.captureErr
}

func (p *testPaster) PasteTarget(target string, pressEnter bool) error {
	p.pastedTarget = target
	p.pressedEnter = pressEnter
	return p.pasteErr
}

type testDependencies struct {
	transcriber testTranscriber
	copier      testCopier
	paster      *testPaster
}

type deliveryExpectation struct {
	paste          bool
	pressEnter     bool
	wantErr        bool
	state          state.State
	indicatorState state.IndicatorState
	target         string
}

func newTestService(t *testing.T, deps testDependencies) *Service {
	t.Helper()
	var paster output.Paster
	if deps.paster != nil {
		paster = deps.paster
	}
	return New(
		DefaultConfig(t.TempDir()),
		nil,
		func() (Transcriber, error) { return deps.transcriber, nil },
		deps.copier,
		paster,
	)
}

func assertDeliveryState(t *testing.T, svc *Service, expected deliveryExpectation) {
	t.Helper()
	svc.store.SetState(state.StateTranscribing)

	err := svc.transcribeAndDeliver("", expected.paste, expected.pressEnter, expected.target)
	if (err != nil) != expected.wantErr {
		t.Fatalf("transcribeAndDeliver() error = %v, want error: %t", err, expected.wantErr)
	}
	if got := svc.Status(); got != expected.state {
		t.Errorf("Status() = %q, want %q", got, expected.state)
	}
	if got := svc.store.Indicator(); got != expected.indicatorState {
		t.Errorf("Indicator() = %q, want %q", got, expected.indicatorState)
	}
}

func TestServiceTranscribeAndDeliverDone(t *testing.T) {
	t.Parallel()
	svc := newTestService(t, testDependencies{transcriber: testTranscriber{text: testText}})

	assertDeliveryState(t, svc, deliveryExpectation{
		state:          state.StateIdle,
		indicatorState: state.IndicatorDone,
	})
}

func TestServiceTranscribeAndDeliverTranscriptionFailure(t *testing.T) {
	t.Parallel()
	svc := newTestService(t, testDependencies{
		transcriber: testTranscriber{err: errors.New("transcription failed")},
	})

	assertDeliveryState(t, svc, deliveryExpectation{
		wantErr:        true,
		state:          state.StateIdle,
		indicatorState: state.IndicatorIdle,
	})
}

func TestServiceTranscribeAndDeliverCopyFailure(t *testing.T) {
	t.Parallel()
	svc := newTestService(t, testDependencies{
		transcriber: testTranscriber{text: testText},
		copier:      testCopier{err: errors.New("copy failed")},
	})

	assertDeliveryState(t, svc, deliveryExpectation{
		wantErr:        true,
		state:          state.StateIdle,
		indicatorState: state.IndicatorIdle,
	})
}

func TestServiceTranscribeAndDeliverPasteFailure(t *testing.T) {
	t.Parallel()
	paster := &testPaster{pasteErr: errors.New("paste failed")}
	svc := newTestService(t, testDependencies{
		transcriber: testTranscriber{text: testText},
		paster:      paster,
	})

	assertDeliveryState(t, svc, deliveryExpectation{
		paste:          true,
		wantErr:        true,
		state:          state.StateIdle,
		indicatorState: state.IndicatorIdle,
		target:         testTarget,
	})
}

func TestServiceTranscribeAndDeliverPastesSavedTarget(t *testing.T) {
	t.Parallel()
	paster := &testPaster{}
	svc := newTestService(t, testDependencies{
		transcriber: testTranscriber{text: testText},
		paster:      paster,
	})

	assertDeliveryState(t, svc, deliveryExpectation{
		paste:          true,
		state:          state.StateIdle,
		indicatorState: state.IndicatorDone,
		target:         testTarget,
	})
	if paster.pastedTarget != testTarget {
		t.Errorf("PasteTarget() = %q, want %q", paster.pastedTarget, testTarget)
	}
}

func TestServiceTranscribeAndDeliverPastesAndPressesEnter(t *testing.T) {
	t.Parallel()
	paster := &testPaster{}
	svc := newTestService(t, testDependencies{
		transcriber: testTranscriber{text: testText},
		paster:      paster,
	})

	assertDeliveryState(t, svc, deliveryExpectation{
		paste:          true,
		pressEnter:     true,
		state:          state.StateIdle,
		indicatorState: state.IndicatorDone,
		target:         testTarget,
	})
	if !paster.pressedEnter {
		t.Error("PasteTarget() did not request Enter")
	}
}

func TestServiceCaptureTarget(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		mode       state.Mode
		paster     *testPaster
		wantTarget string
		wantErr    bool
	}{
		{"copy mode", state.ModeCopy, &testPaster{target: testTarget}, testTarget, false},
		{"copy without target", state.ModeCopy, &testPaster{captureErr: errors.New("missing")}, "", false},
		{"paste mode", state.ModePaste, &testPaster{target: testTarget}, testTarget, false},
		{"paste without target", state.ModePaste, &testPaster{captureErr: errors.New("missing")}, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestService(t, testDependencies{paster: tc.paster})
			got, err := svc.captureTarget(tc.mode)
			if (err != nil) != tc.wantErr {
				t.Fatalf("captureTarget() error = %v, want error: %t", err, tc.wantErr)
			}
			if got != tc.wantTarget {
				t.Errorf("captureTarget() = %q, want %q", got, tc.wantTarget)
			}
		})
	}
}

func TestServiceTranscribeAndDeliverEmptyTranscript(t *testing.T) {
	t.Parallel()
	svc := newTestService(t, testDependencies{})

	assertDeliveryState(t, svc, deliveryExpectation{
		wantErr:        true,
		state:          state.StateIdle,
		indicatorState: state.IndicatorIdle,
	})
}
