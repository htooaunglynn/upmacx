# Step-by-step delivery handoff

Each phase builds on the previous phase. Start by reading [00](00-index.md), then the CLI, architecture, provider and testing contracts. Implement only the providers assigned to that phase. Keep future provider IDs unavailable until their phase ships; reject unshipped IDs clearly. Every phase updates help, README, fixtures, compatibility evidence and release notes.

## Phase 1 — v0.1 foundation and Homebrew

Prerequisites: approved specs; locally available supported stable Go and Git; actual GitHub owner needed only for a publishable module/release. Resolve module owner before committing go.mod; no invented public path.

1. Create Go module and layout from 03, README, actual MIT LICENSE and build-output ignore rules. Set the selected supported stable Go version consistently in module and CI.
2. Implement CLI parsing, version/help, injected Runner, result types, plan rendering, terminal consent and signal cancellation. Add tests as each behavior is introduced.
3. Implement brew refresh, structured discovery, eligibility filtering, explicit upgrades and opt-in cleanup with automatic cleanup disabled.
4. Add CI for tests, vet, nonmutating formatting check and a Darwin arm64 build. Add a macOS process/cancellation test job.
5. Complete fake-brew workflow matrix and controlled native smoke validation. Prepare an Apple Silicon arm64 archive and source formula using 06; publish only after the release gates pass.

Completion: `--check` cannot install or clean, declined confirmation changes nothing, partial failure exits correctly, cleanup is opt-in, help/version need no brew, the arm64 target builds, and controlled formula audit/install/test pass. Artifacts: Go application, tests/fixtures, CI, README/LICENSE, v0.1 release candidate archives/checksums and a separate tap formula. Publication is a later action, not a prerequisite for local correctness.

## Phase 2 — v0.2 macOS software updates

Prerequisites: phase 1 gates; local Apple manual/help evidence and disposable supported-OS test environment.

1. Add softwareupdate detection/parser and report-only candidates to default discovery.
2. Add `--system`, selection validation, terminal/auth handling, restart/manual notes, and install command restrictions.
3. Test labels, command failures, check-only exclusions, opt-in and cancellation. Verify absence of restart, force, password and agreement flags.
4. Validate installation without reboot in a disposable environment for each supported OS family. If evidence is unavailable or behavior is unsafe, mark install capability unavailable there while retaining discovery; explain why in compatibility docs.

Completion: default runs never install system updates, explicit installation is bounded to planned labels, and summary accurately reports required manual completion. Artifacts: adapter, fixtures, compatibility evidence, updated CLI/docs, release candidate.

## Phase 3 — v0.3 Mac App Store

Prerequisites: previous gates; supported mas, controlled entitled test app and account managed by the tester.

1. Add capability detection and JSON-stream parser, using non-accurate discovery.
2. Add planned-ID upgrades, terminal interaction and clear account/Spotlight/privilege diagnostics.
3. Add fixture and integration cases; perform a controlled native app-update test. Do not automate login, service resets or indexing repair.

Completion: mas is optional by default, explicit unavailability fails, and a mas failure does not block independent providers. Artifacts: adapter/tests, prerequisites and compatibility evidence, updated release candidate.

## Phase 4 — v0.4 developer tools

Prerequisites: previous gates; isolated prefixes/environments and tested package-manager versions. Implement one adapter at a time in order npm, pnpm, pipx, Cargo.

1. Add global-only discovery and ownership/source exclusions, with versioned fixtures.
2. Implement only the explicit approved-package operations in 04. Validate compatible targets and avoid manager self-replacement.
3. Require modern pipx outdated/maintenance controls and the separate cargo-update helper; missing capabilities never trigger dependency installation or a guessed fallback.
4. Test per-provider failures and a combined run. Verify caller projects remain byte-for-byte unchanged and no alternate user/toolchain prefix is traversed.

Completion: all seven providers are represented in help/default selection, supported global tools update, excluded origins/pins are visible, and no project dependencies are modified. Artifacts: four adapters and tests, compatibility records, comprehensive prerequisites and v0.4 release candidate.

## Phase 5 — v1.0 stable public release

Prerequisites: all provider gates and native compatibility evidence, resolved owner/license holder and real artifact checksums.

1. Review end-to-end UX and cross-provider ordering, error summaries, authentication/manual action and interruption handling.
2. Run the full acceptance matrix and clean-environment installation on Apple Silicon arm64; document tested OS/tool versions and limitations.
3. Validate tag-to-draft-release CI, artifact checksums and manual tap update process. Rehearse a patch release without moving existing tags.
4. Publish source/release/tap through the documented maintainer workflow and verify public installation instructions. Declare only proven compatibility.

Completion: users can install and update upmacx through the tap, understand coverage, inspect before changing software, and get accurate outcomes. Artifacts: stable release, verified formula and installation guide, complete compatibility/test evidence. Persistent logs, rich TUI and additional integrations remain future work, not hidden v1.0 gates.

## Final documentation and handoff checks

- All source requirements map to 01's traceability table and a phase or explicit deferral.
- CLI options, IDs, consent and exit codes match across docs, implementation and tests.
- Check-only never installs or cleans; system installation and cleanup each require their explicit flag and consent.
- No fabricated owner, checksum, tested version, completed gate or publication claim.
- An engineer can execute phases in order; upstream compatibility evidence is collected during implementation rather than assumed from prose.
