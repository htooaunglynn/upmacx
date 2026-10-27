# Project progress and handoff

Last updated: 2026-10-27 (Asia/Yangon).

## Current state

Phase 1 local application and isolated test implementation is present. Go module, CLI, Homebrew adapter, MIT license, README and CI configuration have been added. Native provider installation and formula validation remain pending, so Phase 1 is not complete or release-ready. No package updates, real project tags/releases, tap updates or publication were performed. Development binaries are ignored build outputs. Release preparation tooling and a temporary tagged-fixture rehearsal now exist; no real release artifacts or source formula have been prepared.

Follow [AGENTS.md](../../AGENTS.md), [the specification index](00-index.md) and [the delivery sequence](07-step-by-step-delivery.md). This file records evidence and status; it does not override requirements.

## Milestone status

| Milestone | Implementation / authoring | Validation | Publication |
| --- | --- | --- | --- |
| Specification baseline, 00–07 | Complete | Internal links, code fences and file count checked; provider documentation reviewed | Not applicable |
| Agent instructions and progress tracking | Complete | Local Markdown links and expected files checked | Not applicable |
| Phase 1: foundation and Homebrew | In progress: application, tests, CI, docs and release preparation tooling implemented | Local isolated tests/race, vet, formatting and Darwin cross-builds pass; Ubuntu arm64 native test binaries pass; macOS provider/formula gates pending | Not started |
| Phase 2: macOS software updates | Not started | Not run | Not started |
| Phase 3: Mac App Store | Not started | Not run | Not started |
| Phase 4: developer tools | Not started | Not run | Not started |
| Phase 5: stable release | Not started | Not run | Not started |

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
- Homebrew 7.0.10 read-only version/help evidence and synthetic source-reviewed fixtures exist; automated isolated tests have run. Native arm64 help/version have run; native Intel execution, actual provider discovery/install fixtures and formula audit/install/test remain unrun. No public release readiness is claimed.
- Controlled disposable Homebrew environments for both architectures and an immutable release tag/source checksum are required for the remaining Phase 1 release gates.

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

## Next step

Ubuntu portable test execution is complete; no further Linux testing is needed for the current source. For the remaining macOS acceptance gates, first resolve Computer Use access to the designated macOS UTM guest (its Apple backend rejects guest-agent execution), then verify guest OS/architecture and installed tools. Continue Phase 1 acceptance in [07](07-step-by-step-delivery.md): use disposable designated Homebrew environments to acquire actual empty/outdated/pinned/self-updating fixtures and controlled formula/cask upgrade, automatic-cleanup suppression and interactive authorization evidence. Validate native Intel execution and both native formula installs/tests. Only then use the release preparation helper with a real clean tagged checkout and the exact GitHub source archive under [06](06-release-and-homebrew.md). The authored tag-to-draft-release workflow still needs hosted validation. Phase 2 remains unstarted.

## Maintenance format

For each future update, adjust the milestone table and append a dated entry with: work completed, affected file links, exact checks and outcomes, unresolved issues, and next step. Use `Not started`, `In progress`, `Blocked` (with a concrete reason), or `Complete` for task status. Keep validation and publication separate; never infer them from code completion.
