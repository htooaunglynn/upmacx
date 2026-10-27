# Architecture and execution

## Project layout

- `cmd/upmacx/main.go`: version variable, OS signals, dependency wiring, exit status only.
- `internal/cli`: parsing, usage, terminal detection, confirmation and rendering.
- `internal/updater`: orchestration, plan/result types and provider registry.
- `internal/process`: executable lookup, subprocess capture/streaming and cancellation.
- `internal/providers/{brew,softwareupdate,mas,npm,pnpm,pipx,cargo}`: adapters and adjacent fixture tests.
- `testdata/`: sanitized version-labeled command output; `docs/`, `go.mod`, README, LICENSE and CI added in phase 1.

Use the standard library for flags, JSON, process execution and signals. A small terminal-detection dependency is allowed only if documented and tested on both Darwin architectures; prefer an injected terminal predicate so tests need no real TTY. Keep Go version selection in phase 1 tied to the stable supported toolchain then available, and record it in go.mod and CI together.

## Internal interfaces

These are internal Go contracts, not a public plugin API. Use constructors to inject a Runner into each provider.

```go
type Provider interface {
    ID() string
    Detect(context.Context) (Availability, error)
    Discover(context.Context) (Plan, error)
    Apply(context.Context, Plan) (Outcome, error)
    Cleanup(context.Context) error
}
type Runner interface {
    LookPath(string) (string, error)
    Run(context.Context, Command) (CommandResult, error)
}
```

- `Availability`: available bool, tool version, reason. Missing executable/unsupported capability is unavailable; unexpected lookup/probe failure is an error.
- `Command`: resolved executable path, argument slice, environment overrides, working directory, mode (`capture` or `stream`), stdin policy (`closed` or `inherit`). Capture stdout/stderr separately with a 16 MiB limit per stream; overflow fails discovery rather than parsing truncation. Stream mode routes directly to injected output writers.
- `CommandResult`: captured byte streams and child exit code; errors distinguish start failure, nonzero exit, overflow and context cancellation. Providers alone normalize documented nonzero discovery statuses.
- `Candidate`: opaque validated ID, display name, installed/target version strings (unknown allowed), kind, restart/manual-action note. Never derive executable names from candidate text.
- `Plan`: provider ID, candidates, exclusions with reasons, report-only bool. Candidate IDs are kept separate from display strings; empty plan cannot trigger Apply.
- `Outcome`: provider ID, completed candidate IDs if provable, operation status and manual-action notes. Aggregate results also record failed stage, diagnostic and whether earlier mutation occurred.

Cleanup is a no-op on providers other than brew; the orchestrator calls it only for selected brew with `--clean`, consent and no brew failure. Discovery must not set mutation flags or invoke cleanup. Rendering consumes plans/results; provider logic never prompts for upmacx consent.

## Process and data flow

Resolve tools once per run, retain absolute paths, and invoke with `exec.Command`/argument slices, never a shell or joined command string. Use a neutral temporary working directory for provider commands so local project configuration is not accidentally selected; retain user package-manager configuration and active installation prefixes. Remove that temporary directory after children exit. Limit overrides to documented adapter needs; never overwrite HOME or CARGO_HOME.

Use `LC_ALL=C` for text parsers, structured output where supported, and fixtures for every supported schema. Reject malformed nonempty output, missing required IDs, duplicates with conflicting targets and option-like IDs. Don't equate empty stdout with success when stderr/exit status indicates failure. Pass tool-specific option terminators where supported; otherwise validate IDs against the provider's identifier grammar.

Build immutable plans for all providers, render once, confirm once, execute in canonical order. Refresh checks may access network and caches. Before each mutation ensure its original executable still exists; failure stops that provider. Child tools remain authoritative on applicability and dependency resolution; no silent broadening to newly discovered top-level packages. A failed multi-item command may already have changed packages: label its outcome partial/unknown, not “nothing changed.”

Set a process group on Darwin and propagate SIGINT/SIGTERM to that group. Allow five seconds to exit, then kill and reap the group. Observe the direct child exit/signal before draining output inherited by descendants; an open descendant pipe must not delay signal notification until the capture timeout. Keep the cancellation watcher alive until output drains and group cleanup completes. Inject the timer in tests. No fixed timeout for interactive upgrades; cancellation remains available. Captured checks have a five-minute timeout per command and report timeout as provider failure. Never reuse prompt input buffering for child stdin.

No persistent state/database, plugin loading, automatic retry, recovery engine or telemetry is required. A rerun rediscovers actual package-manager state.
