# eban — agent instructions

Agents read this file before doing work in this repository. These instructions
are mandatory.

## Mandatory language-specific rules

Code rules live in `.opencode/rules/`. They are mandatory and not ranked.
Before editing any `.go` file, read both in full:

- @.opencode/rules/go-architecture.md
- @.opencode/rules/go-code-rules.md

## Skills

Community Go skills (samber/cc-skills-golang) are installed in
`.agents/skills/cc-skills-golang/skills/` and auto-discovered by OpenCode.
Recommended combos:

- **Writing/editing Go code**: golang-code-style + golang-naming +
  golang-error-handling + golang-safety
- **CLI work**: golang-cli + golang-project-layout
- **Concurrency/timers/watchdog**: golang-concurrency + golang-context
- **Testing**: golang-testing
- **Before commit**: golang-lint (config is .golangci.yml) +
  golang-modernize (`go tool modernize ./...`)

Project constraints override skills: stdlib-only, no new deps without
explicit user approval.

## Behavioural discipline

- Answer every explicit request in the user's message
- "в чат / in chat / на аппрув" means output text in chat. Do not create or
  edit files unless the user explicitly says to write into the project
- After finishing, re-scan the latest user message and verify nothing was
  skipped
- Run `make -C eban check` after any code change; report linter output
  verbatim

## Project essentials

- Skeleton root holds agent infrastructure only; the Go project lives in
  `eban/` (module `github.com/lewenbraun/go-skeleton/eban`)
- Binary `eban`: push-to-talk dictation, pw-record -> ElevenLabs Scribe v2
  -> clipboard (wl-copy) / typing (wtype)
- Tooling pinned in eban/go.mod `tool` block: golangci-lint, govulncheck,
  stringer, modernize, goreleaser. Run via `go tool <name>` or make targets
- Target: Arch Linux + Wayland + PipeWire. X11 fallback kept
