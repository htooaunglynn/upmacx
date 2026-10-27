# Phase 1 acceptance ledger

Updated 2026-10-27. This separates implemented code from acceptance evidence.
It does not waive any requirement in the numbered specifications.

| Gate | Current evidence | Status |
| --- | --- | --- |
| Module, MIT license, README and ignored build output | Present in source | Implemented |
| Parsing, pure help/version, root/OS refusal, consent | Isolated unit and fake-tool tests pass | Validated locally |
| Provider registry order and immutable approved plans | Direct updater regressions pass | Validated locally |
| Check-only, explicit typed upgrades, opt-in cleanup, partial failure | Isolated fake-brew workflow tests pass | Validated locally |
| Capture limits, timeouts, process-group cancellation | Darwin tests/race and Linux native test binaries pass | Validated locally |
| Interactive input, Ctrl-C, parent foreground restoration | Native Darwin pseudo-terminal tests pass with fake child | Terminal plumbing validated; real installer pending |
| Real Homebrew version/help | 7.0.10 version and help captured; multiline version probe regression added | Read-only validation only |
| Real outdated/pinned/auto-updating cask schemas | [Real arm64 cask run](2026-10-27-homebrew-arm64.md), empty/pinned schemas and updater plans | Cask validation passed; real formula schema pending |
| Real formula/cask upgrades and automatic cleanup suppression | Real local cask upgrades, pinning, self-update exclusion and explicit cleanup passed; formula build blocked | Cask subset passed; formula/keg suppression pending |
| Authorization denial and partial mutation through real manager | Real cask preflight failure after an earlier completed upgrade passed; fake permission tests pass | Partial mutation validated; real vendor authorization pending |
| Test/race/vet/format, Darwin arm64/amd64 builds | Local checks pass | Validated locally |
| Native Intel executable and process tests | Explicit Intel CI job authored | Hosted/native Intel execution pending |
| Dual archives, binary checksums, source-checksummed formula generator | Temporary tagged fixture rehearsal passed; helper tests pass | Tooling validated; real tagged artifacts pending |
| Formula strict audit, source install and test on native architectures | Ruby syntax check of temporary fixture recipe passed | Homebrew acceptance pending |
| Tag-to-draft workflow | Source authored; separate write-permission draft job | Hosted validation pending |
| Public release/tap publication | Not requested or performed | Separate future action |

## Controlled Homebrew runner

[scripts/native_brew_acceptance.py](../../scripts/native_brew_acceptance.py) is an
explicit maintainer command, never run by normal Go tests or the installed CLI.
Use a disposable macOS environment or dedicated test prefix, with an existing
Homebrew source checkout containing tag 7.0.10 and the matching already-installed
portable Ruby. The command refuses an existing test root. It does not download
or install dependency tools, invoke sudo, or use installed user packages as data.

From an ordinary terminal outside the current inherited agent sandbox:

```sh
go build -o dist/upmacx-native-acceptance ./cmd/upmacx
python3 scripts/native_brew_acceptance.py \
  --root /private/tmp/upmacx-native-acceptance-new-run \
  --brew-source /opt/homebrew \
  --upmacx "$PWD/dist/upmacx-native-acceptance"
```

Substitute an existing Homebrew source path on the test machine. On Intel this
is commonly `/usr/local/Homebrew`; do not assume a path without checking it.
The selected Go toolchain must already exist for the build command.

The runner creates a separate Homebrew checkout, HOME, caches, logs, build-temp,
local Git remotes, empty core/cask fixture taps, fixture tap and payloads beneath
the new root. A Git wrapper restricts child Git operations to file remotes. Homebrew is
frozen at 7.0.10, so explicit refresh uses only local fixture remotes. It verifies
its effective prefix before installs. Test formulae have no dependencies and
use `--ignore-dependencies`; casks contain only a local script binary linked
inside the test prefix. No app is installed to `/Applications`. The application
itself still preserves HOME; only this maintainer harness isolates test HOME.

The default formula-inclusive path remains blocked. `--casks-only` runs the cask
subset independently; its successful results are recorded [here](2026-10-27-homebrew-arm64.md).
Assertions cover empty/outdated schema, formula and cask pins,
self-updating cask exclusion despite inherited greedy variables, check-only and
nonterminal refusal, explicit upgrades, old-keg retention without cleanup,
explicit cleanup, and real partial failure with cleanup skipped. Each command's
argv/status/stdout/stderr is saved beneath `evidence/` in the test root. Treat
that as local raw evidence; sanitize paths and any sensitive diagnostics before
checking fixtures into source.

## Attempt on the development Mac

Host: macOS 27.0.1, build 26A434, arm64. Homebrew 7.0.10 source commit
`8f10429b56be28a25032b87463fffc95e556b5bc` and its matching existing portable Ruby
were copied into temporary test prefixes. `brew --prefix`, version and help
worked in isolation. An initial prefix setup needed a distinct HOMEBREW_TEMP;
that was corrected before fixture installation.

The first install attempt failed because the command sandbox denied Homebrew's
local build socket. An approved unsandboxed retry then failed with Homebrew's
own error: `Inherited sandbox permits writes to .../brew/bin/brew`.
Inspection of that checkout's `Library/Homebrew/sandbox.rb` confirmed this is an
explicit inherited-confinement check. No sandbox-disable option or code patch
was used. Neither formula-inclusive run completed even the first formula fixture install.
A separate cask-only run subsequently passed; the full formula-inclusive runner
is not claimed end-to-end validated.

The user-designated macOS UTM VM was no longer registered when rechecked; only
the Ubuntu VM remained. Ubuntu execution is recorded [separately](2026-10-27-ubuntu-arm64.md)
and cannot replace these native macOS checks.

## Remaining release inputs

Real `v0.1.0` tagged source and its exact GitHub archive bytes do not yet exist
for this work. Do not commit an invented checksum, create an installable formula
pointing to a nonexistent tag, or treat a temporary fixture tag as a project
release. Use [releaseprep](../../cmd/releaseprep/main.go) once the documented gates
and clean-tag prerequisites are satisfied. It will generate both archives,
verified checksum lines and the separate source formula. Native formula
acceptance and public promotion remain distinct.
