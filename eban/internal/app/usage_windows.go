//go:build windows

package app

const usageText = `eban - push-to-talk dictation via ElevenLabs Scribe v2

Usage:
  eban <command> [flags]

Commands:
  toggle    start/stop recording (bind to a hotkey)
  start     start recording
  stop      stop, transcribe, copy/paste text
  status    print state (idle|recording|transcribing)
  tray      run the tray indicator and Alt+Space hotkeys (right click quits)
  help      show this help

Flags (toggle, start, stop):
  -paste      paste text into the window active when recording started
  -enter      paste text and press Enter in that window
  -lang code  speech language, auto-detected by default
  -timeout    max recording length (default 10m)

Tray hotkeys (press once to start, press again to stop and deliver):
  Alt+Space     toggle dictation; paste into the invoking window + Enter
  Alt+Space+V   toggle in paste-only mode (hold V as part of the press)
  Alt+Space+B   toggle in clipboard-only mode (hold B as part of the press)
  While still holding the starting press, V or B switches the delivery mode.
  Keys typed without the chord held reach the active window normally.

Setup:
  scoop install ffmpeg (recording backend)
  setx ELEVENLABS_API_KEY "<key>" (or put the key into
  the apikey file under .config\eban in your home directory)

Saying "paste" or one of its russian equivalents at the end of a phrase
switches the result from clipboard-only to pasting into the saved window.
`
