# Go Architecture Rules (eban)

Mandatory before touching any `.go` file. General Go knowledge lives in
`.agents/skills/cc-skills-golang/` — load the relevant skill; this file only
contains project-specific constraints.

## Repository shape

```
<skeleton root>            agent rules (.opencode/rules/), skills (.agents/)
└── eban/                  the Go project (module github.com/lewenbraun/go-skeleton/eban)
    ├── go.mod             tool directives: golangci-lint, govulncheck, stringer, modernize, goreleaser
    ├── .golangci.yml      20 linters + formatters
    ├── Makefile           build/install/test/vet/lint/fmt/vuln/modernize/check
    ├── cmd/eban/          main.go — thin entry, no logic
    └── internal/
        ├── app/           CLI surface: dispatch, commands, flags, usage, watchdog, wiring
        ├── dictation/     orchestration Service + consumer-side seams (Recorder, Transcriber)
        ├── recorder/      pw-record process lifecycle
        ├── state/         State/Mode enums + file-backed Store (dir injectable)
        ├── elevenlabs/    Scribe v2 HTTP client + API key loading
        └── output/        clipboard/typing/notify with Wayland and X11 backends
```

- **stdlib only** for runtime code. New dependencies require explicit user
  approval. Dev tools are NOT dependencies: they live in the go.mod `tool`
  block (`go get -tool ...`, run via `go tool <name>`)

## Extensibility seams

- `dictation.Service` consumes small interfaces defined where they are
  consumed: `Recorder` (dictation), `Transcriber` (dictation), and the
  output package's `Copier`/`Typer`/`Notifier`. Swapping ElevenLabs for
  another provider = new package under `internal/` (e.g. `internal/whisper`),
  zero changes in dictation or app
- Platform backends (Wayland/X11, future Windows) live only in
  `internal/output/`. Never leak platform checks outside that package
- `state.Store` takes its directory as a constructor arg — tests use
  `t.TempDir()`, never `/tmp/eban`
- New CLI commands: add to the `commands` map in `internal/app/commands.go`,
  keep builtin fallbacks in `runBuiltin`

## External processes and state

- Recording is a detached child (`Setsid`), tracked by PID file. State
  machine: `idle -> recording -> transcribing -> idle`
- Cross-invocation communication is filesystem state files only. No
  daemons, no sockets, no IPC
- Kill sequence is always SIGINT (graceful, WAV header finalization) ->
  wait loop -> SIGKILL fallback (`recorder.StopPID`)
- Target: Arch Linux + Wayland + PipeWire; X11 fallback kept
