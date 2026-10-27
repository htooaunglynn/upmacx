# Provider specifications

## Shared adapter rules

Commands below are argument vectors expressed in shell notation for readability; the application never invokes a shell. Probe the tool version and required command help during Detect. Capability detection, not an invented minimum version, controls availability. Record tested tool versions and sanitized fixtures before enabling an adapter. A parser schema change is a compatibility failure, never permission to run a broad update.

All adapters must pass the missing-tool, no-update, candidate, malformed-output, command-failure, check-only, cancellation and permission tests in [05](05-testing-and-acceptance.md). Installation uses explicit planned IDs, sequentially per candidate unless stated otherwise. On failure stop that provider. Targets are previews of manager resolution, not a lockfile guarantee. Show this distinction where targets cannot be pinned.

Tool installation instructions belong in README; upmacx never executes them. Optional dependencies are not runtime dependencies of the Homebrew formula. Credentials remain with the provider. No adapter invokes sudo, unpins software, or accepts licenses. Package-manager child processes may invoke their own authorization flows after consent.

## Homebrew (`brew`)

- Detect `brew` on PATH and probe `brew --version`. Support formula and cask discovery through `brew outdated --json=v2`; refresh with `brew update` first.
- Set `HOMEBREW_NO_INSTALL_CLEANUP=1`, `HOMEBREW_NO_AUTO_UPDATE=1`, and `HOMEBREW_UPGRADE_GREEDY=0` on discovery/apply commands; explicit refresh still runs. Also clear `HOMEBREW_UPGRADE_GREEDY_CASKS` and set `HOMEBREW_UPGRADE_GREEDY_LATEST=0` / `HOMEBREW_UPGRADE_GREEDY_AUTO_UPDATES=0` so inherited per-cask/greedy settings cannot broaden the plan. Homebrew 7.0.10 source confirms `0` disables boolean greedy settings. Verify effective behavior in fixtures/smoke tests.
- Parse separate formula/cask arrays and their identifiers/versions. Use `brew list --pinned` to exclude pinned formulae. Respect cask pinning where supported, and normal non-greedy eligibility; do not force `auto_updates`/`latest` casks. Capability tests must cover upstream pinning representation.
- Apply each candidate using `brew upgrade --formula ID` or `brew upgrade --cask ID`. Validate names including tap-qualified tokens; never pass display names. Dependency changes remain Homebrew-managed. Check-only runs neither command.
- With `--clean` and consent, run `brew cleanup` once after a successful brew stage, including an empty upgrade plan. Skip after any brew refresh/discovery/apply failure. Cleanup failure marks brew failed but preserves previous upgrade outcomes. Never run autoremove or aggressive prune options.
- Cask installers may request authentication or require closing an app; forward their interaction. Failed installers are errors, not automatic reasons to retry with force.
- Acceptance: formula-only, cask-only, mixed, pinned, self-updating, no changes, upgrade failure, cleanup failure, and automatic-cleanup suppression are covered. Explicit IDs can behave differently from bulk upgrades, so eligibility filters must be tested before shipping.

Development evidence: [7.0.10 compatibility record](../../internal/providers/brew/testdata/README.md) and [controlled arm64 cask acceptance](../validation/2026-10-27-homebrew-arm64.md). Real local cask discovery, pins, non-greedy eligibility, upgrades, cleanup and partial failure passed. Formula builds, real vendor authentication, Intel and source-formula release gates remain outstanding, so this is not a release-readiness claim. The synthetic mixed schema is still distinguished from acquired native cask fixtures. Other versions remain unavailable until their schema/environment semantics and help capabilities are reviewed.

Reference: [Homebrew manual](https://docs.brew.sh/Manpage). The manual documents automatic cleanup controls; opt-in cleanup is an upmacx policy layered on those controls.

## macOS software updates (`softwareupdate`)

- Detect `/usr/sbin/softwareupdate`; use that fixed path rather than a PATH replacement. Probe `--help`. Discover with `--list`, with C locale.
- Parse `Label:` records, title/version and action metadata; recognize the documented no-updates response. Preserve each complete label as one argument. Malformed nonempty listing is failure. Required-restart records get a manual-action note; unknown action metadata also requires review.
- Without `--system`, all candidates are report-only. With it, run `--install LABEL` per approved candidate, without `--all`, `--restart`, `--force`, `--stdinpass`, installer launch, or license-acceptance flags. Require an interactive terminal for this operation; `--yes` does not remove that requirement.
- macOS may require administrator/owner authentication or reject an install that cannot finish without restart. Pass through the native interaction; report failure/manual next steps without retrying as root. Do not claim absence of a restart flag guarantees every OS release's behavior: non-rebooting application must pass a controlled test for each supported OS family before enabling installation there. Discovery remains usable when installation is unsupported.
- Do not fetch full installers, install Rosetta, alter automatic-update settings, or schedule restarts. When completion is uncertain, state that System Settings must be checked; never report all system updates installed solely from a zero exit code.
- Acceptance: multiline labels, no updates, restart and unknown-action notes, default report-only, explicit opt-in, authorization denial, unsupported OS installation mode, and absence of every prohibited argument.

Primary evidence: Apple's locally installed `softwareupdate(8)` manual inspected 2026-10-26 (`man softwareupdate`), including list/install labels, authentication and restart options. [Apple command-line guide](https://www.apple.com/server/docs/Command_Line.pdf) supplies historical command semantics; it is not a modern compatibility guarantee. [Apple's current installation guide](https://support.apple.com/en-us/102662) distinguishes full-installer downloads. Modern OS fixtures/manual evidence are a release gate.

## Mac App Store (`mas`)

- Detect `mas`, probe `mas version` and help for `outdated`, JSON, and update/upgrade alias support. Requires eligible installed App Store apps, valid Spotlight indexing and an account entitled to updates.
- Discover `mas outdated --json`; decode a stream of JSON objects, not an assumed JSON array. Normalize numeric application IDs plus installed/available versions from the tested schema. Use default detection mode; never use `--accurate` during checks because it can initiate downloads/dialogs. Show that default detection can miss or misidentify updates.
- Apply `mas upgrade ID` per approved numeric ID (documented alias of update). Require a terminal for installation. mas may perform its own privilege request; upmacx neither initiates sudo nor bypasses it. Account/auth/index errors are provider failures with manual guidance; never sign in, reindex or reset services automatically.
- Acceptance: multiple JSON records, no records, incompatible schema, account/Spotlight failure, alias capability, and explicit IDs. Cleanup is unsupported.

Reference: [mas project documentation](https://github.com/mas-cli/mas), including JSON streams, accuracy tradeoffs and authentication requirements.

## npm (`npm`)

- Detect `npm --version`. Discover `npm outdated --global --json`; inspect `npm ls --global --depth=0 --json` to exclude linked/local/git sources and identify current installed direct packages.
- Parse the package-name map; current/wanted/latest are distinct. Treat exit 1 as an outdated result only when valid candidate JSON is present without an error envelope; all other nonzero results fail. Scoped names are supported. Exclude npm, pnpm and corepack themselves to avoid replacing manager installations owned by another installer.
- Choose the numeric `wanted` target for eligible registry packages; show major changes in the plan. Apply `npm install --global NAME@TARGET` per candidate, pinning the previewed target. No unqualified local install/update commands. Reject nonregistry sources and nonversion targets as exclusions.
- Use active user prefix/configuration. Access-denied/registry/authentication errors fail this provider; do not change ownership or invoke sudo. Installer lifecycle scripts remain the manager's behavior and must be disclosed in README.
- Acceptance: scope names, empty map, outdated exit 1, error envelope, links, manager exclusions, and exact global targets.

References: [npm outdated](https://docs.npmjs.com/cli/v11/commands/npm-outdated/), [npm update](https://docs.npmjs.com/cli/v11/commands/npm-update/), [npm install](https://docs.npmjs.com/cli/v11/commands/npm-install/). Exact-version installation is the selected policy rather than an unbounded global update.

## pnpm (`pnpm`)

- Detect `pnpm --version`, help for global outdated JSON formatting and global add. Discover `pnpm outdated --global --format json`; obtain ownership/source metadata with `pnpm list --global --depth 0 --json`.
- Normalize the version-specific JSON structure with fixtures; do not assume npm's schema. Accept outdated-specific nonzero status only for a tested version/status plus valid candidate data and no error envelope. Otherwise fail. Exclude links, nonregistry sources, manager packages npm/pnpm/corepack and explicitly pinned global specs. Use compatible targets; do not opt into `--latest` range breaking.
- Apply `pnpm add --global NAME@TARGET` per approved eligible registry package. No workspace traversal and no commands that could modify the caller's project. Capture global prefix/setup failures as errors with manual guidance, never run `pnpm setup`.
- Acceptance: global scope, compatible target, JSON variants, linked/pinned packages, access denial, and no local manifest/lockfile modification. Cleanup is unsupported.

References: [pnpm outdated](https://pnpm.io/cli/outdated), [pnpm update](https://pnpm.io/cli/update), [pnpm add](https://pnpm.io/cli/add). Global exact-version add is chosen to constrain installation to the previewed set.

## pipx (`pipx`)

- Detect `pipx --version` and required help capabilities. Discover using `pipx list --outdated --short --skip-maintenance`, plus `pipx list --json --skip-maintenance` for environment metadata. Require support for all of these flags; old versions are unavailable with a manual upgrade hint.
- Outdated short output identifies installed environment names; cross-reference every name against JSON `venvs`. Exclude pinned, editable, local, VCS and injected-only packages. Inspect metadata for pinned status; missing required fields are schema incompatibility. No ad hoc PyPI requests or pip traversal fallback. Target may be unknown; show “manager resolves target” before consent.
- Apply `pipx upgrade NAME --skip-maintenance` per approved environment. Preserve pipx's recorded install constraints; do not include injected dependencies, shared-interpreter upgrades or all-user `--global` scope. “Global tools” here means current-user CLI apps, not pipx's all-user flag.
- No cleanup. Missing interpreter, broken environment, network or authentication problems fail this provider; do not reinstall or repair automatically.
- Acceptance: outdated names cross-reference, pin exclusion, custom indexes, empty output, missing metadata, old capabilities and shared-maintenance suppression. Test that check-only does not upgrade shared libraries.

Reference: [pipx CLI reference](https://pipx.pypa.io/stable/reference/cli.html). The discovery strategy requires modern outdated-list support and deliberately avoids assuming JSON snapshot metadata contains latest versions.

## Cargo-installed tools (`cargo`)

- Detect `cargo --version` and executable `cargo-install-update` from the separately installed cargo-update project; probe its help. Cargo alone is insufficient. Explain manual installation of the helper in README, never install it automatically.
- Discover `cargo install-update --list`; parse its package/installed/latest/needs-update table in C locale using versioned fixtures. Only default-registry executable crates are eligible initially. Exclude git/path/custom-registry origins, rustup components and cargo-update itself. Use `cargo install --list` and helper metadata/output to establish ownership; do not infer all installed crates are registry crates. If origin cannot be established, exclude with a reason.
- Apply `cargo install-update NAME:TARGET` per eligible candidate; do not use all, force, downdate, allow-no-update or git flags. Preserve helper package configuration and Cargo's configured home. Do not run `cargo update`, which targets project dependencies. Surface helper constraints/build/auth failures; no uninstall/reinstall fallback.
- No additional helper is required by upmacx, though cargo-update may use an already installed cargo-binstall. Detect configured helper install options and disclose them; reject options that broaden selected packages or downgrade targets. No cleanup.
- Acceptance: missing helper, table parsing, registry-source exclusion, helper self-exclusion, explicit targets, compilation failure and no project mutation.

References: [cargo-update project](https://github.com/nabijaczleweli/cargo-update), [maintainer's command manual](https://raw.githubusercontent.com/nabijaczleweli/cargo-update/master/man/cargo-install-update.md). The helper provides listing and version-qualified package selection; it is not part of Cargo itself.

## Compatibility evidence required per adapter

Before enabling a provider, record in its testdata README: tool version, OS/architecture, required help output, fixture acquisition command, success/no-updates/outdated exit statuses and parser schema. Unknown capabilities default to unavailable, not a guessed command variant. This document verifies upstream command intent, not real-machine behavior. Controlled installation tests remain required before publication.
