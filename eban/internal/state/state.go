// Package state persists the dictation session (pid, mode, language,
// started-at and the current state machine value) as files in a directory.
package state

// State is the dictation state machine value.
type State string

// Recording session states.
const (
	StateIdle         State = "idle"
	StateRecording    State = "recording"
	StateTranscribing State = "transcribing"
)

// Mode selects the delivery behaviour for a recording session.
type Mode string

// Delivery modes for a recording session.
const (
	ModeCopy  Mode = "copy"
	ModePaste Mode = "paste"
)
