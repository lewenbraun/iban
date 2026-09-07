package hotkey

import (
	"testing"

	"github.com/lewenbraun/eban/eban/internal/state"
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

func TestChordHoldAndRelease(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionStop, state.ModePasteEnter, true}},
		{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionNone, "", false}},
	})
}

func TestChordAltReleasedFirst(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKRMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKRMenu, Down: false}, actionExpectation{ActionStop, state.ModePasteEnter, false}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, "", false}},
	})
}

func TestChordSelectorV(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKKeyV, Down: true}, actionExpectation{ActionSelect, state.ModePaste, true}},
		{Event{VK: VKKeyV, Down: false}, actionExpectation{ActionNone, state.ModePaste, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionStop, state.ModePaste, true}},
	})
}

func TestChordSelectorB(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKKeyB, Down: true}, actionExpectation{ActionSelect, state.ModeCopy, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionStop, state.ModeCopy, true}},
	})
}

func TestChordSpaceAlonePassesThrough(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, "", false}},
	})
}

func TestChordSelectorsOutsideChordPassThrough(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKKeyV, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKKeyB, Down: true}, actionExpectation{ActionNone, "", false}},
	})
}

func TestChordAutoRepeatIsSwallowed(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionNone, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionStop, state.ModePasteEnter, true}},
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
	feed(t, []keyStep{
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionStop, state.ModePasteEnter, true}},
		{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, true}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionStop, state.ModePasteEnter, true}},
	})
}

func TestChordSpaceThenAltStarts(t *testing.T) {
	t.Parallel()
	feed(t, []keyStep{
		{Event{VK: VKSpace, Down: true}, actionExpectation{ActionNone, "", false}},
		{Event{VK: VKLMenu, Down: true}, actionExpectation{ActionStart, state.ModePasteEnter, false}},
		{Event{VK: VKLMenu, Down: false}, actionExpectation{ActionStop, state.ModePasteEnter, false}},
		{Event{VK: VKSpace, Down: false}, actionExpectation{ActionNone, "", false}},
	})
}
