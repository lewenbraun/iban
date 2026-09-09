package hotkey

import (
	"testing"

	"github.com/lewenbraun/iban/iban/internal/state"
)

type keyStep struct {
	event Event
	want  actionExpectation
}

type actionExpectation struct {
	action  Action
	mode    state.Mode
	swallow bool
}

func feed(t *testing.T, steps []keyStep) {
	t.Helper()
	var c Chord
	for i, s := range steps {
		action, mode, swallow := c.Key(s.event)
		got := actionExpectation{action, mode, swallow}
		if got != s.want {
			t.Fatalf("step %d (%v): got %+v, want %+v", i, s.event, got, s.want)
		}
	}
}

func TestChordPressStartsSecondPressStops(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, true}},
		{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStop, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, true}},
		{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
	})
}

func TestChordSelectorHeldAtStart(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKKeyV, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePaste, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePaste, true}},
		{Event{VK: VKKeyV, Down: false}, actionExpectation{ActionNone, state.ModePaste, false}},
		{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionNone, state.ModePaste, false}},
	})
}

func TestChordSelectorBHeldAtStart(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKKeyB, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModeCopy, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModeCopy, true}},
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, state.ModeCopy, false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStop, state.ModeCopy, true}},
	})
}

func TestChordLiveSelectorWhileHoldingStartPress(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKKeyV, Down: true}, actionExpectation{ActionSelect, state.ModePaste, true}},
		{Event{VK: VKKeyV, Down: false}, actionExpectation{ActionNone, state.ModePaste, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePaste, true}},
		{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionNone, state.ModePaste, false}},
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, state.ModePaste, false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStop, state.ModePaste, true}},
	})
}

func startPressSteps() []keyStep {
	return []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, true}},
	}
}

func TestChordTypingDuringRecordingPassesThrough(t *testing.T) {
	t.Parallel()
	steps := append(startPressSteps(),
		keyStep{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKSpace, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKKeyV, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKKeyV, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKRMenu, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKRMenu, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
	)
	feed(t, steps)
}

func TestChordSpaceAlonePassesThrough(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, "", false}},
	})
}

func TestChordAutoRepeatDoesNotRetrigger(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStop, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, true}},
	})
}

func TestChordInjectedEventsPassThrough(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true, Injected: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true, Injected: true}, actionExpectation{ActionNone, "", false}},
	})
}

func TestChordRestartAfterStop(t *testing.T) {
	t.Parallel()
	steps := append(startPressSteps(),
		keyStep{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStop, state.ModePasteEnter, true}},
		keyStep{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, true}},
		keyStep{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		keyStep{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
	)
	feed(t, steps)
}

func TestChordSpaceThenAltToggles(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, false}},
		{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, false}},
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionStop, state.ModePasteEnter, false}},
	})
}
