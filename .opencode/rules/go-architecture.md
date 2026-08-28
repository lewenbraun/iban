# Go Architecture Rules (eban)

Mandatory before touching any `.go` file. General Go knowledge lives in
`.agents/skills/cc-skills-golang/` — load the relevant skill; this file only
contains project-specific constraints.

## Project shape

- Single binary, module `eban`, Go >= 1.27 (toolchain pinned in go.mod)
- **stdlib only** for runtime code. New dependencies require explicit user
  approval first. Dev tools are NOT dependencies: they go into the `tool`
  directive block in go.mod (`go get -tool ...`, run via `go tool <name>`)
- Flat package layout on purpose (7 files, no packages). Do not create
  `cmd/`, `internal/`, `pkg/` unless the user asks — see golang-project-layout
  skill for when flat stops working

## Files and ownership

| File | Owns |
|---|---|
| `main.go` | entry, command dispatch (map, not switch), usage text |
| `cmd.go` | command implementations, options parsing, voice-command regexp |
| `state.go` | State/Mode enums, state files in /tmp/eban |
| `recorder.go` | pw-record lifecycle, watchdog |
| `scribe.go` | ElevenLabs client, API key loading |
| `out.go` | clipboard / typing / notifications, Wayland vs X11 |

Adding a new system integration = new file, do not grow existing ones.

## External processes and state

- Recording is a detached child (`Setsid`), controlled by PID file in
  `/tmp/eban/`. State machine: `idle -> recording -> transcribing -> idle`
- Cross-invocation communication is filesystem-based state files only.
  No daemons, no sockets, no IPC
- Kill sequence is always SIGINT (graceful, WAV header finalization) ->
  wait loop -> SIGKILL fallback

## Platform

- Wayland is primary (wl-copy, wtype), X11 is fallback (xclip, xdotool),
  detected via env. Windows support will arrive as `*_windows.go` files
  with build tags — keep platform code isolated in out.go so this stays true
