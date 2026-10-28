# Full controlled Homebrew arm64 fixture acceptance — 2026-10-28

Result: the corrected native fixture runner completed successfully, exit 0,
with Homebrew's sandbox enabled. This covers actual local fixture formulae and
casks, not the tagged upmacx distribution formula or vendor authorization UI.

## Root cause and correction

Earlier runs placed the test Homebrew repository under `/private/tmp`. Homebrew
7.0.10's `Library/Homebrew/extend/os/mac/sandbox.rb`,
`allow_write_system_temp`, grants writes beneath `/private/tmp`,
`/private/var/tmp`, and the per-user macOS temporary/cache directories. Those
allowances apply during formula build sandbox setup. Consequently, the test
Homebrew executable remained writable and `Sandbox.inherited_sandbox?` correctly
rejected it with `Inherited sandbox permits writes to .../brew/bin/brew`.

The user reproduced that error in an ordinary Terminal, disproving the earlier
attribution to Codex's inherited sandbox. The script layout caused this failure;
no additional user permission or separate VM was required to fix it. The first
agent-side socket error was a separate command-sandbox restriction.

The runner now validates the canonical test root before creating any files and
rejects these broadly writable locations, including Xcode/Swift cache paths.
Use a new protected project-local prefix, for example
`$PWD/.native-tests/acceptance-new-run`. The repo ignores `.native-tests/`.
No Homebrew source patches, sandbox-disable flags or launcher permission hacks
were used. [Layout regression tests](../../scripts/test_native_brew_paths.py)
cover denied paths, allowed project prefixes and component boundaries.

The first corrected native run reached the intended formula failure, but the
harness asserted that the diagnostic must appear on stderr. Homebrew's build
PTY forwarded it to stdout. The assertion now checks both captured streams and
still requires the exact intentional fixture failure text. A fresh full rerun
then passed.

## Successful environment and procedure

- Apple Silicon arm64, macOS 27.0.1 (26A434).
- Homebrew 7.0.10 source commit `8f10429b56be28a25032b87463fffc95e556b5bc`.
- Existing matching portable Ruby copied into the isolated manager checkout.
- New project-local `.native-tests/` root with separate HOME, prefix, caches,
  build-temp, logs, and local fixture repositories/payloads.
- Empty local core/cask fixture taps and a Git wrapper enforcing file-only
  protocols; no user package installations were enumerated or updated.
- Only local script payloads installed. No dependency tools, vendor application,
  system package, sudo operation, reboot, real tag or public release was created.

The full [runner](../../scripts/native_brew_acceptance.py), without `--casks-only`,
was executed against the existing arm64 development CLI. It recorded the
following successful assertions:

1. Actual zero-update JSON discovery, exit 0, before advancing definitions.
2. Formula and cask pin exclusions; a self-updating cask excluded even with
   inherited greedy settings enabled outside the application.
3. Check-only discovery and nonterminal refusal, exit 2, both retaining the
   original installed fixture versions.
4. Explicit approved formula and cask upgrades from 1.0 to 2.0, exit 0.
5. Formula 1.0 keg remained after normal upgrade, proving automatic cleanup
   suppression for that operation; pinned/self-updating fixtures stayed 1.0.
6. Empty eligible plan with explicit approved cleanup, exit 0; formula 1.0 keg
   was removed only at this stage.
7. A later approved formula upgrade to 3.0 succeeded before an intentional
   second-formula build failure. upmacx returned 1, retained completed IDs,
   skipped cleanup, and left the previous 2.0 keg intact.

The [mixed native JSON snapshot](../../internal/providers/brew/testdata/homebrew-7.0.10-darwin-arm64-mixed.json)
contains actual formula/cask candidates and both pin exclusions from this run.
Paths/account data are absent. Raw command records remain in the ignored local
test root; no credentials or full environment dumps are copied into docs.

## Remaining scope

The native fixture formula/cask and cleanup gate is validated. Real vendor
installer authorization handling, strict audit/install/test of the real tagged
upmacx source formula, hosted draft workflow validation and publication remain
separate. This run is not evidence that those operations passed. Intel remains
unsupported and is not an acceptance prerequisite.
