# Controlled Homebrew arm64 acceptance — 2026-10-27

Environment: macOS 27.0.1 (26A434), arm64. Manager: Homebrew 7.0.10,
source commit `8f10429b56be28a25032b87463fffc95e556b5bc`, using its matching
already-installed portable Ruby copied to an isolated temporary prefix.

Result: the real-manager **cask-only** acceptance run passed. Formula build,
real vendor application authentication, native Intel and upmacx formula
acceptance remain separate, unpassed gates. These fixtures are local test
packages, not software from the user's ordinary Homebrew installation.

## Isolation and procedure

The [opt-in runner](../../scripts/native_brew_acceptance.py) created a new test
prefix, HOME, cache, logs, build-temp, and local-only fixture tap/remotes. Its
four casks contain only small local shell-script payloads linked within that
prefix. No application bundle was installed into `/Applications`. Existing user
packages were never enumerated as update data or upgraded.

The runner freezes the Homebrew checkout at its reviewed tag. Empty local
fixture `homebrew/core` and `homebrew/cask` taps prevent bootstrap of external
taps. Those are explicitly test fixtures, not snapshots of the official taps.
A dedicated Git wrapper enforces file-only remotes, including through Homebrew's
subprocess environment filtering. All payload URLs are local file URLs. No
sandbox-disable option, dependency tool installation, sudo or reboot is used.

Reproduction (root must not exist; provide the actual source/binary paths):

```sh
python3 scripts/native_brew_acceptance.py \
  --root /private/tmp/upmacx-cask-acceptance-new-run \
  --brew-source /opt/homebrew \
  --upmacx "$PWD/dist/upmacx-arm64" --casks-only
```

## Observed successful run

- Installed four local fixture casks through real Homebrew.
- Captured an actual empty `outdated --json=v2` result with exit 0.
- Pinned the designated cask and advanced local fixture definitions.
- `upmacx --check --providers brew` returned 0, showed the eligible cask and
  pinned exclusion, and left the installed script at 1.0.
- Inherited `HOMEBREW_UPGRADE_GREEDY=1` and a per-cask greedy setting did not add
  the self-updating fixture to the plan. Its installed script stayed at 1.0.
- Nonterminal invocation without `--yes` returned 2, without an upgrade.
- Approved invocation upgraded the explicitly planned cask from 1.0 to 2.0,
  returned 0 and reported the completed candidate. The pinned cask stayed 1.0.
- `--yes --clean` with no eligible upgrades ran explicit cleanup, removed old
  fixture cache archives and returned 0. Pinned exclusions remained visible.
- A later plan upgraded the eligible cask to 3.0, then the broken cask's
  intentional preflight failure occurred. upmacx returned 1, retained the first
  completed candidate, reported failure and did not run approved cleanup.
  Homebrew had already unlinked the broken cask's old binary before failure;
  no rollback or repair was attempted. This demonstrates why failed operations
  are described as partial/unknown rather than no change.

Native JSON snapshots are stored under
[provider testdata](../../internal/providers/brew/testdata/README.md).
The runner records each command's argv, exit code, stdout and stderr locally in
its test root. No credentials, private account data or full environment dump
are recorded in this document.

## Failed attempts and limitations

The formula-inclusive runner failed at its first fixture build. The initial
command sandbox denied a build socket; an approved retry reached Homebrew's own
`Inherited sandbox permits writes to .../brew/bin/brew` guard. That guard was
left enabled. Formula install/upgrade and old-keg cleanup-suppression assertions
have not passed. Cask version migration itself purges old cask versions, so the
cask run is not evidence of formula keg retention.

Early cask harness attempts triggered automatic `homebrew/core` lookup during
unqualified pinning or explicit refresh. Those runs were stopped, no external
package was installed, and no user's ordinary tap/prefix was modified. The final
harness uses qualified fixture names, empty local core/cask fixture taps and a
Git wrapper rather than relying only on a filtered environment variable.

The fake-brew automated matrix, Darwin pseudo-terminal tests and Linux tests are
separate evidence. Script-binary casks do not prove vendor app installation,
authorization denial or dialog handling. The native upmacx formula still requires
a real immutable project tag/source checksum and strict audit/install/test.
