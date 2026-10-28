# Project progress and handoff

Last updated: 2026-10-28 (Asia/Yangon).

## Current state

Phase 1 local application and isolated test implementation is present. Go module, CLI, Homebrew adapter, MIT license, README and CI configuration have been added. Controlled native arm64 formula and cask fixture acceptance now passes, including source builds, old-keg retention and cleanup; real vendor authentication and tagged upmacx source-formula validation remain pending, so Phase 1 is not complete or release-ready. Only isolated local test formulae/casks were installed/upgraded; no ordinary user software, real project tags/releases, tap publication or public release was changed. Development binaries are ignored build outputs. Release preparation tooling and a temporary tagged-fixture rehearsal now exist; no real release artifacts or source formula have been prepared.

Follow [AGENTS.md](../../AGENTS.md), [the specification index](00-index.md) and [the delivery sequence](07-step-by-step-delivery.md). This file records evidence and status; it does not override requirements.

## Milestone status

| Milestone | Implementation / authoring | Validation | Publication |
| --- | --- | --- | --- |
| Specification baseline, 00–07 | Complete | Internal links, code fences and file count checked; provider documentation reviewed | Not applicable |
| Agent instructions and progress tracking | Complete | Local Markdown links and expected files checked | Not applicable |
| Phase 1: foundation and Homebrew | In progress: application, tests, CI, docs and release preparation tooling implemented | Local isolated tests/race, vet, formatting and Darwin cross-builds pass; Ubuntu arm64 native test binaries pass; arm64 real formula/cask fixture suite passed; vendor/tagged-formula/release gates pending | Not started |
| Phase 2: macOS software updates | Not started | Not run | Not started |
| Phase 3: Mac App Store | Not started | Not run | Not started |
| Phase 4: developer tools | Not started | Not run | Not started |
| Phase 5: stable release | Not started | Not run | Not started |

Historical entries below retain the architecture scope in effect on their dates. The user changed current support to Apple Silicon only on 2026-10-28; earlier Intel and dual-architecture prerequisites are superseded.

## Completed work and evidence

### 2026-10-26 — Specifications

- Read the shared chat linked in [context and goals](01-context-and-goals.md).
- Created the eight numbered documents listed in [00-index.md](00-index.md): context, CLI contract, architecture, seven providers, tests, release workflow and phased handoff.
- Incorporated the user's choices: all chat providers, phased delivery, opt-in cleanup and explicit opt-in system installation.
- Reviewed official provider documentation and the local Apple `softwareupdate(8)` manual. References and compatibility gates are recorded in [04-provider-specifications.md](04-provider-specifications.md).
- Checked the eight baseline files for headings, paired code fences and valid local Markdown links. Confirmed no application/module was created. These are documentation checks, not provider compatibility or runtime tests.

### 2026-10-26 — Agent guidance and tracking

- Added root [AGENTS.md](../../AGENTS.md), requiring every agent/model to read and follow `docs/`, preserve the project contracts, and maintain this handoff.
- Added this progress file and linked it from the specification index.
- Checked local links and code-fence pairing across the documentation and agent instructions; confirmed only the requested documentation files exist.

## Outstanding prerequisites and limits

- Module owner confirmed from existing Git remote (`github.com/htooaunglynn/upmacx`); user approved MIT holder `htooaunglynn`. Tap repository availability remains unverified.
- Release tags, actual source/archive checksums, repositories and publication credentials are future maintainer inputs.
- Homebrew 7.0.10 read-only version/help evidence, original synthetic fixtures and acquired native mixed formula/cask fixtures exist. Automated tests and full controlled arm64 fixture acceptance passed. Real vendor authorization and tagged upmacx source-formula audit/install/test remain outstanding. No public release readiness is claimed.
- The Apple Silicon fixture environment works with a protected project-local prefix. Real vendor authorization evidence and an immutable release tag/source checksum are still needed for the remaining gates.

### 2026-10-26 — Phase 1 local implementation

- Added [Go module](../../go.mod) using existing Git remote ownership and Go 1.27.3 (installed toolchain, confirmed stable via official Go release listing). Added user-approved [MIT license](../../LICENSE), [README](../../README.md), [changelog](../../CHANGELOG.md), and build-output ignore rules.
- Added [CLI](../../internal/cli/run.go), [orchestration](../../internal/updater/run.go), [process runner](../../internal/process/process.go), [entry point](../../cmd/upmacx/main.go) and [Homebrew adapter](../../internal/providers/brew/brew.go). Includes strict parsing, pure help/version, root/OS refusal, single consent, bounded captures, neutral working directory, explicit typed candidate upgrades, opt-in cleanup, partial summaries and signal-group cancellation with a test-injectable grace timer. Terminal foreground handoff supports native installer interaction; controlled interactive installer validation remains pending.
- Added unit and fake-brew integration tests beside implementation. The test executable is the fake tool, with isolated PATH/HOME/prefix and command/environment logs. Tested check-only, consent/decline/EOF/nonterminal, empty plan, pinned/unsafe/malformed schemas, missing/incompatible tools, permission/refresh/apply/cleanup errors, partial completion, stream/capture, overflow/timeout, SIGINT/SIGTERM, child signals and descendant termination. Tests never run host update commands.
- Captured only read-only Homebrew version/help on Darwin arm64 (7.0.10). [Compatibility record](../../internal/providers/brew/testdata/README.md) distinguishes those probes and synthetic schemas from unrun controlled discovery/install tests. Inherited greedy cask settings are cleared; the provider specification records this safety refinement. Other Homebrew versions remain unavailable pending review.
- Added [CI](../../.github/workflows/ci.yml) for Linux/macOS tests/race, vet, nonmutating formatting and both Darwin cross-builds. Workflow is authored; hosted CI has not run.
- Validation: `go test ./...`, `go test -race ./...`, `go vet ./...`, `gofmt -l cmd internal`, and CGO-disabled Darwin arm64/amd64 builds passed locally. Native arm64 development binary `--help` and `--version` passed (`upmacx dev`); no foreign-architecture binary was executed. Initial process tests exposed temporary-path symlink differences and an overly short helper startup timeout; both test assumptions were corrected and the full suite passed. The sandbox initially blocked Go's default build cache; checks then ran with approved cache access.
- Documentation local-link and code-fence checks passed. No live provider update/check, formula audit/install/test, release-candidate archive, checksum, tag, tap or publication is claimed.

### 2026-10-26 — Phase 1 continuation: cancellation and release preparation

- Fixed [orchestration](../../internal/updater/run.go) so cancellation preserves earlier independent provider failures; active/unfinished providers retain interrupted outcomes and no later mutation starts. Added a regression in [CLI tests](../../internal/cli/run_test.go).
- Fixed [process cleanup](../../internal/process/process.go) to honor the injected/full five-second grace period even when the process-group leader exits first, and notify child-signal cancellation before the watcher exits. [Process tests](../../internal/process/process_test.go) confirm a signal-ignoring descendant remains alive during an injected grace period and is killed after expiry.
- Added [releaseprep command](../../cmd/releaseprep/main.go), [release implementation](../../internal/release/release.go) and [tests](../../internal/release/release_test.go). Requires a clean checkout and matching existing exact tag, checks the supplied GitHub source archive against all tracked files, runs maintainer gates, builds both architectures, creates verified archive checksums and a formula with the separately computed source checksum. No download, dependency installation, tag creation, tap mutation or publication is implemented.
- Rehearsed the real helper in a separate temporary Git test fixture tagged `v0.0.0` with a Git-generated source archive. Both real cross-build archives were generated and independently verified for SHA-256, root contents and executable permissions. The native arm64 fixture executable passed help/version (`upmacx v0.0.0`); generated formula passed `ruby -c`. This is fixture tooling evidence, not a real project release, authenticated GitHub source download, Homebrew audit/install/test or native Intel evidence. A later source-validation hardening change adds gzip-footer verification; its regression is covered by the final suite.
- Added [tag-to-draft-release CI](../../.github/workflows/release.yml): a read-only preparation job runs isolated checks, downloads the exact tag source and prepares artifacts; a separate contents-write job rechecks checksums and creates only a draft release. Formula remains an Actions artifact for manual tap validation. No hosted run, real tag push or draft creation has occurred.
- Validation: `go test ./...`, `go test -race ./...`, `go vet ./...`, nonmutating formatting, Darwin arm64/amd64 builds, workflow YAML syntax, documentation links/code fences and `git diff --check` pass locally. Hosted CI, controlled native Homebrew operations and formula acceptance remain unrun. No ordinary host software was used as update test data.

### 2026-10-27 — UTM native-test environment connection attempt

- User designated their UTM macOS VM for Phase 1 tests. UTM CLI listed the macOS guest as stopped; starting it succeeded and status returned `started`.
- Read-only guest architecture probe through `utmctl exec --cmd /usr/bin/uname -m` was rejected by UTM with “Operation not supported by the backend.” No guest command, package-manager operation or acceptance test executed successfully.
- UI control was attempted through Computer Use, which reported “Computer Use permissions are not granted.” Granting that permission is the next connection prerequisite. No remote-login setting, guest dependency installation, privilege escalation, reboot or host package update was performed. The VM was left started for the user.
- Retried after the user changed permissions: UTM is listed as running, but two UI connections failed with ScreenCaptureKit error -3811 (“Failed to start stream due to audio/video capture failure”), replacing the earlier permission-denied result. No guest UI or test output was accessible.
- A subsequent retry obtained the UTM accessibility tree and a screenshot of the guest Finder desktop. The first input attempt and all subsequent observations failed with ScreenCaptureKit error -3812 (“Failed due to an invalid parameter”), including after a fresh control-session reset. No guest command or test ran; desktop visibility alone is not acceptance evidence. Next attempt should use a smaller normal UTM window with guest Terminal open.
- Guest OS/architecture shown in the user's screenshot are not recorded as command-verified compatibility evidence. Phase 1 native Homebrew and formula gates remain unrun.

### 2026-10-27 — Ubuntu Server UTM validation

- Connected to the user-designated Ubuntu Server VM over SSH. Guest reported Ubuntu 26.10 and `Linux aarch64`; Go and gcc were absent from PATH. No guest dependency installation, privilege escalation or package updates were performed.
- Built all four existing test suites using Go 1.27.3 on the Mac with `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go test -c`, transferred them with the required Homebrew fixtures and verified the bundle checksum on the guest. Executed each suite natively with `-test.v -test.timeout=120s`: CLI/orchestration, process/cancellation, fake Homebrew and release tests all passed. No source change was needed.
- The Linux CLI passed help/version (`upmacx dev`) and correctly refused normal execution with exit 1, empty stdout and the expected macOS/non-root diagnostic. The enclosing SSH validation command exited 0. Extraction emitted harmless Apple metadata/clock-skew warnings; tests were unaffected.
- [Validation record](../validation/2026-10-27-ubuntu-arm64.md) records exact procedures, input fingerprints, results and limits without credentials or private account/network data. Temporary guest test files remain available for inspection.
- Linux race detection, vet, formatting and compilation/cross-builds inside the VM were not run because guest Go/gcc were absent. These results do not satisfy macOS Homebrew, cask, native Intel or formula acceptance gates. Linux updater execution remains unsupported.

### 2026-10-27 — Phase 1 cancellation, terminal and probe hardening

- Reproduced an interrupted-child pipe bug with a new fake process regression: an interrupted leader plus a signal-ignoring descendant holding stdout caused the old runner to wait five seconds and return a capture deadline error instead of signal cancellation.
- Updated [process runner](../../internal/process/process.go) and added [output pipe lifecycle handling](../../internal/process/output.go). Direct-child status is observed before descendant output drains, so signal notification cancels the run promptly while the watcher preserves group grace/termination. The regression now passes, including under the race detector.
- Added native Darwin [pseudo-terminal tests](../../internal/process/terminal_darwin_test.go), using only the Go test executable. Verified child input, terminal Ctrl-C producing SIGINT, cancellation callback and restoration of the parent foreground process group. An initial assertion was corrected to account for the terminal's echoed `^C`; the final suite passes. This is genuine terminal plumbing evidence, not real cask installation or authentication acceptance.
- Fixed [Homebrew detection](../../internal/providers/brew/brew.go) so probe timeout, cancellation, overflow and execution failures propagate as provider failures instead of being silently classified unavailable. Unsupported help exits still report unavailable. Added runtime-error regressions in [provider tests](../../internal/providers/brew/brew_test.go).
- Validation passed: `go test ./...`, `go test -race ./...`, `go vet ./...`, `gofmt -l cmd internal`, Darwin arm64/amd64 cross-builds, documentation links/code fences and `git diff --check`. Recompiled the two changed test suites for Linux arm64 and reran them natively in the designated Ubuntu VM: both PASS. Updated the [Ubuntu validation record](../validation/2026-10-27-ubuntu-arm64.md).
- No package updates, dependency installs, native Intel execution, real Homebrew/formula acceptance or publication occurred. Phase 1 remains in progress solely pending its outstanding acceptance/release gates; no later phase was started.

### 2026-10-27 — Phase 1 native acceptance and contract completion

- Added [canonical registry/plan cloning](../../internal/updater/registry.go) and [updater regression tests](../../internal/updater/run_test.go): enforced canonical discovery/apply order, a single plan after all discovery, immutable approved candidate slices, mutation-attempt tracking, stage progress and interrupted-cleanup summaries retaining completed IDs. Future provider IDs remain unavailable through the CLI.
- Tightened version handling so invalid execution flag combinations/unshipped options are still rejected with `--version`; help remains informational. Updated CLI contract/tests together. Fixed the actual Homebrew version probe to accept the reviewed first-line version when optional tap details follow; added a regression.
- Added the opt-in [native Homebrew acceptance runner](../../scripts/native_brew_acceptance.py) using a new isolated prefix, HOME/cache/build-temp and local fixture repositories/payloads. Formula-inclusive attempts were blocked first by a build socket restriction and then by Homebrew's own inherited-sandbox write-confinement check. No sandbox bypass was used. Early cask attempts triggered automatic core-tap lookups; those subprocesses were stopped, then qualified pinning, empty local core/cask fixture taps and a file-only Git wrapper closed that bootstrap path. No external package was installed.
- Completed the final cask-only native run on macOS 27.0.1 (26A434), arm64, Homebrew 7.0.10. Real discovery, pinning, self-updating exclusion with inherited greedy variables, check-only, nonterminal refusal without mutation, explicit upgrades, empty-plan explicit cleanup, and intentional cask preflight failure after an earlier success passed. Failed-run cleanup was skipped. The failed cask had already unlinked its old binary; no rollback/repair was attempted. Only test-owned script casks beneath the temporary prefix were changed.
- Acquired actual empty/outdated JSON cask fixtures in [testdata](../../internal/providers/brew/testdata/README.md) and added parser regressions. [Native evidence](../validation/2026-10-27-homebrew-arm64.md) distinguishes these results from synthetic formula data, vendor-app authorization and unpassed release gates. The full [Phase 1 ledger](../validation/phase-1-acceptance.md) records each implemented component and remaining gate.
- Added explicit `macos-15-intel` CI alongside arm64 macOS/Linux and native macOS help/version smoke commands. This is workflow authoring, not a hosted/native Intel pass. The designated macOS UTM guest is no longer registered; only Ubuntu was listed.
- Validation: standard tests, race tests, vet, nonmutating formatting, both Darwin builds, arm64 help/version with an empty executable search PATH, workflow YAML/Python syntax, Markdown links/fences and `git diff --check` passed. Normal Homebrew source checkout remains unchanged. No real project tag, release candidate, source checksum, tap change or publication was produced; the existing staged changes were preserved.

### 2026-10-28 — Apple Silicon-only scope and full test authorization

- User explicitly removed Intel support and authorized full access for testing. Updated specifications 00–07, README, changelog, acceptance ledger and native-runner requirements. Historical 2026-10-27 validation records are marked as preceding this scope change; no earlier evidence is rewritten as a current support claim.
- Added an injected architecture to [CLI execution](../../internal/cli/run.go) and wired actual GOARCH in [main](../../cmd/upmacx/main.go). Non-arm64 execution is refused before provider operations or temporary setup; help/version remain pure informational invocations. [CLI tests](../../internal/cli/run_test.go) cover Intel refusal and zero provider calls.
- [Release preparation](../../internal/release/release.go) now produces only the Darwin arm64 archive/checksum and generates `depends_on arch: :arm64`. [Release tests](../../internal/release/release_test.go) assert one arm64 build/checksum and the formula restriction. Removed Intel CI/build/upload paths and added a native arm64 runner assertion. Removed the obsolete ignored local amd64 build generated by earlier work; no tracked user work was discarded.
- Rehearsed the actual release helper in a separate temporary test repository tagged `v0.0.0`: one arm64 archive, its checksum, archive members/mode, arm64-only build metadata, native injected version output and formula Ruby syntax passed. No real project tag, release or tap mutation occurred.
- Retried the complete local fixture formula/cask acceptance command with approved unsandboxed execution under the user's full testing authorization. The first formula install still failed Homebrew's own `Inherited sandbox permits writes to .../brew/bin/brew` check. No security-disable flag or Homebrew patch was used. An attempt to open ordinary Terminal through Computer Use was separately blocked by the tool with “not allowed to use the app ... for safety reasons”; no alternative GUI shell was used to evade that restriction.
- Validation passed: standard/race tests, vet, formatting, Darwin arm64 build and native help/version, one-archive release rehearsal, YAML/Python syntax, documentation links/fences and diff whitespace checks. Intel validation is no longer a gate. Real formula build/keg retention, vendor authorization and tagged source-formula audit/install/test remain outstanding; hosted CI/publication have not run.

### 2026-10-28 — Corrected native fixture layout; full formula/cask pass

- User reproduced the inherited-sandbox error in an ordinary Terminal. Corrected the earlier attribution to Codex: Homebrew's macOS `allow_write_system_temp` grants writes under `/private/tmp`, `/private/var/tmp` and per-user temporary/cache paths. Locating the manager there left its own executable writable, so its protection check correctly failed.
- Added [layout validation](../../scripts/native_brew_paths.py) before test-root setup, [Python regressions](../../scripts/test_native_brew_paths.py) and a CI step. Unsafe temporary/Xcode cache roots are rejected with a useful error before cloning/installing. Updated runnable examples to `$PWD/.native-tests/<new-run>` and ignored native test artifacts/bytecode. No Homebrew patches, sandbox disable flags, privilege changes or executable-permission workarounds were used.
- The first protected-prefix run completed the intended manager operations but exposed a harness assertion expecting the intentional error only on stderr. Homebrew's PTY delivered the diagnostic on stdout; the assertion now checks both streams and still requires the specific failure text.
- Fresh full native rerun exited 0: formula/cask source installation and discovery, both pin exclusions, non-greedy/self-update exclusion, check-only, nonterminal refusal without mutation, upgrades, old formula keg retention with implicit cleanup suppressed, explicit cleanup, and an intentional later formula build failure retaining earlier successes while skipping cleanup all passed.
- Added acquired [native mixed JSON](../../internal/providers/brew/testdata/homebrew-7.0.10-darwin-arm64-mixed.json) and a parser regression. [Full evidence](../validation/2026-10-28-homebrew-full-arm64.md) and [acceptance ledger](../validation/phase-1-acceptance.md) distinguish these fixture results from unpassed distribution-formula/vendor-authorization gates.
- Validation: layout unit tests, unsafe-root rejection before setup, full real native fixture run, standard/race Go tests, vet, Python/YAML syntax, formatting, documentation links/fences and diff whitespace checks pass. Existing user package installations and normal Homebrew source were not modified. No real project tag, tap update or publication occurred.

## Next step

The formula/cask fixture gate is complete; no further permission grant or VM is needed for it. Keep future roots outside system temporary/cache allowances and use the documented project-local path. Remaining Phase 1 gates are real vendor installer authorization evidence, actual arm64 tagged upmacx source-formula audit/install/test and hosted draft workflow validation. Prepare real source/archive checksums only from an authorized clean project tag; temporary fixture tags are not release evidence. Intel remains unsupported. Phase 1 is still in progress; Phase 2 has not started.

## Maintenance format

For each future update, adjust the milestone table and append a dated entry with: work completed, affected file links, exact checks and outcomes, unresolved issues, and next step. Use `Not started`, `In progress`, `Blocked` (with a concrete reason), or `Complete` for task status. Keep validation and publication separate; never infer them from code completion.
