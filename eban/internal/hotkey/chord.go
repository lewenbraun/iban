// Package hotkey turns raw keyboard events into dictation actions for the
// Alt+Space chord family. Recording is toggle-driven: one chord press starts
// the recording, the next chord press stops it. Holding V as part of the
// chord starts in paste-only mode, B in clipboard-only mode; pressing V or B
// while the chord is held during a recording switches the delivery mode.
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
// events belong to it. A chord press fires once, when both keys transition
// to held; auto-repeat and releases never retrigger it. While a toggle
// recording runs, keys pressed without the full chord held pass through to
// the system untouched. It is not safe for concurrent use; feed it from the
// same thread that receives the keyboard events.
type Chord struct {
	alt            bool
	space          bool
	v              bool
	b              bool
	started        bool
	mode           state.Mode
	swallowSpaceUp bool
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
		wasHeld := c.alt
		c.alt = true
		if c.space && !wasHeld {
			return c.press(), c.mode, false
		}
		return ActionNone, c.mode, false
	}
	c.alt = false
	return ActionNone, c.mode, false
}

func (c *Chord) spaceKey(down bool) (Action, state.Mode, bool) {
	if down {
		wasHeld := c.space
		c.space = true
		if c.alt && !wasHeld {
			c.swallowSpaceUp = true
			return c.press(), c.mode, true
		}
		return ActionNone, c.mode, false
	}
	c.space = false
	swallow := c.swallowSpaceUp
	c.swallowSpaceUp = false
	return ActionNone, c.mode, swallow
}

// press resolves one full chord press: it toggles the recording and picks
// the delivery mode from the selector keys held at press time.
func (c *Chord) press() Action {
	if c.started {
		c.started = false
		return ActionStop
	}
	c.started = true
	c.mode = state.ModePasteEnter
	if c.v {
		c.mode = state.ModePaste
	}
	if c.b {
		c.mode = state.ModeCopy
	}
	return ActionStart
}

// selectorKey only reacts while the full chord is physically held, so V and
// B typed during a toggle recording reach the active application normally.
func (c *Chord) selectorKey(vk uint32, down bool) (Action, state.Mode, bool) {
	held := c.alt && c.space
	if vk == VKKeyV {
		c.v = down
	} else {
		c.b = down
	}
	if !down || !held || !c.started {
		return ActionNone, c.mode, held
	}
	c.mode = state.ModePaste
	if vk == VKKeyB {
		c.mode = state.ModeCopy
	}
	return ActionSelect, c.mode, true
}
