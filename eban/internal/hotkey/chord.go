// Package hotkey turns raw keyboard events into dictation actions for the
// Alt+Space chord family: holding Alt+Space records and pastes with Enter,
// adding V while held switches to paste-only, adding B to clipboard-only.
// Releasing any chord key stops the recording.
package hotkey

import (
	"github.com/lewenbraun/eban/eban/internal/state"
)

// Windows virtual-key codes used by the chord.
const (
	VKLMenu uint32 = 0xA4
	VKRMenu uint32 = 0xA5
	VKSpace uint32 = 0x20
	VKKeyV  uint32 = 0x56
	VKKeyB  uint32 = 0x42
)

// Action is the dictation action triggered by a keyboard event.
type Action string

// Chord-triggered actions.
const (
	ActionNone   Action = "none"
	ActionStart  Action = "start"
	ActionSelect Action = "select"
	ActionStop   Action = "stop"
)

// Event is one raw keyboard event from the OS input stream.
type Event struct {
	VK       uint32
	Down     bool
	Injected bool
}

// Chord tracks the physical state of the Alt+Space chord and decides which
// events belong to it. It is not safe for concurrent use; feed it from the
// same thread that receives the keyboard events.
type Chord struct {
	alt     bool
	space   bool
	started bool
	mode    state.Mode
}

// Key feeds one keyboard event and returns the triggered action, the
// delivery mode chosen so far, and whether the event must be swallowed
// so the rest of the system never sees it.
func (c *Chord) Key(e Event) (Action, state.Mode, bool) {
	if e.Injected {
		return ActionNone, c.mode, false
	}
	switch e.VK {
	case VKLMenu, VKRMenu:
		return c.altKey(e.Down)
	case VKSpace:
		return c.spaceKey(e.Down)
	case VKKeyV, VKKeyB:
		return c.selectorKey(e.VK, e.Down)
	default:
		return ActionNone, c.mode, false
	}
}

// Mode returns the delivery mode chosen for the running chord.
func (c *Chord) Mode() state.Mode {
	return c.mode
}

func (c *Chord) altKey(down bool) (Action, state.Mode, bool) {
	if down {
		c.alt = true
		if c.space && !c.started {
			return c.start(), c.mode, false
		}
		return ActionNone, c.mode, false
	}
	c.alt = false
	if c.started {
		return c.finish(false)
	}
	return ActionNone, c.mode, false
}

func (c *Chord) spaceKey(down bool) (Action, state.Mode, bool) {
	if down {
		armed := c.alt || c.started
		if !c.space {
			c.space = true
			if armed && !c.started {
				return c.start(), c.mode, true
			}
		}
		return ActionNone, c.mode, armed
	}
	wasChord := c.started
	c.space = false
	if wasChord {
		return c.finish(true)
	}
	return ActionNone, c.mode, false
}

func (c *Chord) selectorKey(vk uint32, down bool) (Action, state.Mode, bool) {
	if !c.started {
		return ActionNone, c.mode, false
	}
	if down {
		c.mode = state.ModePaste
		if vk == VKKeyB {
			c.mode = state.ModeCopy
		}
		return ActionSelect, c.mode, true
	}
	return ActionNone, c.mode, true
}

func (c *Chord) start() Action {
	c.started = true
	c.mode = state.ModePasteEnter
	return ActionStart
}

func (c *Chord) finish(swallow bool) (Action, state.Mode, bool) {
	mode := c.mode
	c.reset()
	return ActionStop, mode, swallow
}

func (c *Chord) reset() {
	c.alt = false
	c.space = false
	c.started = false
	c.mode = ""
}
