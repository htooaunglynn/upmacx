# upmacx specification index

Status: documentation baseline; no application exists yet. Prepared 2026-10-26.

Read these documents in order. Requirements describe intended behavior, not implemented or tested capabilities. Provider command references were inspected on the preparation date; compatibility must be demonstrated with fixtures and versioned smoke-test evidence before shipping each adapter.

| Step | Document | Purpose / dependency |
| --- | --- | --- |
| 1 | [Context and goals](01-context-and-goals.md) | Source, audience, scope, traceability |
| 2 | [Product and CLI](02-product-and-cli-spec.md) | Authoritative user-facing behavior |
| 3 | [Architecture](03-architecture-and-execution.md) | Implements the CLI contract |
| 4 | [Providers](04-provider-specifications.md) | Adapter contracts using that architecture |
| 5 | [Testing](05-testing-and-acceptance.md) | Acceptance gates for steps 2–4 |
| 6 | [Release](06-release-and-homebrew.md) | Distribution after acceptance gates pass |
| 7 | [Delivery](07-step-by-step-delivery.md) | Ordered work and completion evidence |
| 8 | [Progress](08-progress.md) | Current status, validation evidence and next handoff |

Every agent and model must follow [the root agent instructions](../../AGENTS.md) and the specifications in this directory. Read progress at the start of work and update it when project status changes.

## Approved decisions

- Go CLI named `upmacx`, macOS arm64 and amd64; prefer standard library.
- Specify all seven providers now; ship them in phases.
- Cleanup requires `--clean`; system installation requires `--system`.
- Never automatically install missing dependencies, escalate privileges, or reboot.
- Default license MIT; two GitHub repositories; source-build Homebrew formula.
- Documentation only in this task. No Go module, application code, workflows, formula file, Git initialization, package update, or publication is performed.

## Authority and readiness

The CLI spec controls behavior; provider specs control command mapping. If future upstream behavior conflicts, change the specification and tests together before enabling that adapter. Never substitute an unsafe command to make a test pass. Publication is blocked on real owner, immutable tag, license holder, and checksums. These placeholders do not block local implementation.

All numbered phases are required for the full updater. Deferred ideas are GUI/TUI, scheduling, persistent configuration/log files, JSON public output, rollback, Linux support, arbitrary scripts, Rust toolchain updates, project dependencies, and universal updating of manually installed applications. They are not implied requirements.
