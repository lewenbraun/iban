//go:build !windows

package app

const usageText = `iban - push-to-talk dictation via ElevenLabs Scribe v2

Usage:
  iban <command> [flags]

Commands:
  toggle    start/stop recording (bind to a hotkey)
  start     start recording
  stop      stop, transcribe, copy/paste text
  status    print state (idle|recording|transcribing)
  help      show this help

Flags (toggle, start, stop):
  -paste      paste text into the window active when recording started
  -enter      paste text and press Enter in that window
  -lang code  speech language, auto-detected by default
  -timeout    max recording length (default 10m)

API key:
  $ELEVENLABS_API_KEY or ~/.config/iban/apikey (chmod 600)

Saying "paste" or one of its russian equivalents at the end of a phrase
switches the result from clipboard-only to pasting into the saved window.
`
