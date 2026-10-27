# Product and CLI contract

## Invocation

`upmacx [--check] [--yes] [--clean] [--system] [--providers LIST]`

| Option | Contract |
| --- | --- |
| no options | Discover enabled providers, show update plan, confirm, update eligible non-system software, summarize |
| `--help`, `-h` | Print usage, provider IDs and limitations; no subprocesses; exit 0 |
| `--version` | Print `upmacx VERSION`; no subprocesses; exit 0 |
| `--check` | Refresh metadata and discover only; no install or cleanup, including implicit maintenance |
| `--yes` | Bypass only upmacx confirmation; does not supply credentials or accept installer agreements |
| `--clean` | Request Homebrew cleanup after a successful Homebrew update stage, including when no upgrades exist |
| `--system` | Include softwareupdate installation in the plan; no automatic reboot |
| `--providers LIST` | Comma-separated subset: `brew,softwareupdate,mas,npm,pnpm,pipx,cargo` |

Reject unknown flags, positional arguments, empty/unknown provider IDs, repeated `--providers`, `--check --clean`, and `--check --system` with exit 2 before any subprocess. Deduplicate IDs and use canonical execution order, not user order. `--yes --check` is valid and has no additional effect. `--system` without explicit providers includes softwareupdate; with explicit providers it requires that ID. `--clean` requires brew in the selected set. `--no-clean` is unsupported. Help takes precedence over version; otherwise version ignores valid execution options after syntax validation. Version does not waive incompatible flag combinations or options unavailable in the current phase.

Only providers shipped in the current release appear in help and are valid IDs. Default selection is all shipped providers. Missing optional executables are skipped by default; explicitly requested missing or incompatible providers are failures. An installed but incompatible provider is reported as unavailable; skip by default with a compatibility reason. A compatible provider whose discovery fails is a failure even in default selection.

## Run sequence and consent

1. Parse without invoking tools. Help/version work on any OS; execution requires Darwin and refuses effective UID 0 so Homebrew/user prefixes are not accidentally managed as root.
2. Detect, refresh, and discover sequentially in order: brew, mas, npm, pnpm, pipx, cargo, softwareupdate. Discover all before starting changes. A failed provider does not block independent providers.
3. Print per-provider candidates with current/target version where available, exclusions and scope. Display system candidates as report-only unless `--system`. List cleanup as a separate action. Package managers may resolve dependency changes beyond top-level candidates; state this before confirmation.
4. If eligible work exists, prompt `Continue with planned changes? [y/N]: `. Accept only trimmed case-insensitive `y` or `yes`; all other input including EOF declines. Decline cancels all mutation, including cleanup. With nonterminal stdin and no `--yes`, do not read input: report confirmation required and exit 2 if work exists. No work requires no prompt.
5. Execute only the approved plan. Keep explicit candidate IDs; never replace a failed parse with an “upgrade everything” fallback. A new top-level candidate discovered after confirmation requires a new run. Skip later actions of a failed provider; continue other providers.
6. Summarize every selected provider. Return failure if any failed even when other work succeeded or consent was declined. Exit 0 for a clean declined run, no updates, successful check, or successful execution.

No background work, retries, dependency installation, automatic privilege escalation, or reboot commands. Child installers can request authentication themselves; forward terminal interaction but do not collect/store passwords. `--yes` does not imply unattended success. If a provider requires a terminal for installation and none is available, fail that provider and continue. Do not promise rollback or transactional updates.

## Output and status

Human-readable plain text is the initial public output. Stdout: plan, stage progress, child stdout, final summary. Stderr: warnings, errors, child stderr. Prefix upmacx messages with provider ID; no spinner or ANSI dependency. Sanitize control characters in parsed names before display and never print credentials or complete environment dumps.

Each provider ends as unavailable, up-to-date, checked, report-only, declined, succeeded, failed, or interrupted. A successful subprocess means its operation completed; do not invent exact package counts or claim an unverified version. Excluded pinned/linked/unsupported items remain visible as exclusions rather than failures. System completion may require manual action and must say so.

| Exit | Meaning |
| --- | --- |
| 0 | Completed without failures, including clean decline/check/no work |
| 1 | Provider/OS/runtime failure, including partial failure |
| 2 | Invalid invocation or required confirmation unavailable |
| 130 | SIGINT cancellation |
| 143 | SIGTERM cancellation |

Signal exits take precedence; usage errors precede execution; runtime failure outranks a normal decline. On cancellation stop scheduling new work, terminate the child process group, wait for it, and summarize completed/unfinished stages. Warn that interrupted installations may require package-manager repair; do not attempt repair or rollback.
