# Go Code Rules (eban)

Mandatory before editing any `.go` file. Style baseline is the
golang-code-style and golang-naming skills; this file adds project rules.

## Hard gates

Every change must pass `make check` (go vet + golangci-lint + tests + build).
Configured limits are law, not suggestions:

- funlen: max 40 lines / 25 statements per function — split instead of nolint
- gocyclo: max 12, nestif: max 4
- goconst: no string literal repeated 3+ times — make it a typed constant
- errcheck: every error is either handled, wrapped with `%w`, or explicitly
  discarded as `_ =` (only for best-effort side effects like notifications)

## Language rules

- Enums: `type State string` + const block. Never scatter raw strings that
  are part of a known set
- Enum-to-string for logs/CLI: stringer-generated, not hand-written
- All user-facing strings, errors and identifiers: **English**
- No comments in code unless the user explicitly asks for them
- Command dispatch: `map[string]func([]string) error`, switch only for
  fallback/builtin cases
- Timeouts and cancellation via `context.WithTimeout` on API calls; HTTP
  client timeout as a second belt
- Errors: `fmt.Errorf("...: %w", err)` for wrapping, sentinel `errors.New`
  for static conditions. Compare with `errors.Is/As`
- Concurrency: goroutines only with a defined exit (Wait, ctx, or
  fire-and-forget with explicit `_ =` rationale)

## Testing

- stdlib `testing` only, table-driven with `t.Run` subtests and `t.Parallel`
  (testify is available as a skill but NOT a dependency of this project)
- Pure logic gets tests (regexp parsing, clamps, state parsing). Process and
  network code is covered manually — do not mock os/exec or net/http here
- Test file name mirrors the source: `cmd.go` -> `cmd_test.go`

## Formatting

`make fmt` (gofmt + gofumpt + goimports via golangci-lint) before every
commit. CI/`make check` fails on unformatted code.
