# Homebrew compatibility evidence

Current evidence: the [full protected-prefix native run](../../../../docs/validation/2026-10-28-homebrew-full-arm64.md)
passed on 2026-10-28. Earlier unverified/blocked statements below describe their
original baseline and are superseded by the dated acquisitions.

## Original source-reviewed baseline

Development adapter: Homebrew 7.0.10, inspected 2026-10-26 on Darwin arm64.
The version and four help text files were captured with `brew --version` and
`brew help outdated`, `brew help list`, `brew help upgrade`, `brew help cleanup`.
These are read-only probes; no installed-package listing, metadata refresh,
cleanup or upgrade was run against the host.

`mixed.json` is a synthetic sanitized fixture, not a real installed-package
snapshot. Its schema was checked against Homebrew 7.0.10 source:
`Library/Homebrew/cmd/outdated.rb` (`json_info`) and
`Library/Homebrew/cask/cask.rb` (`outdated_info`). Both arrays have `name`,
`installed_versions`, `current_version`, `pinned` and `pinned_version`.
Both formula and cask pinning are represented. Unknown/missing pin fields fail.
Homebrew's normal non-greedy outdated filtering supplies cask eligibility;
explicit upgrades use only that filtered set. The parser additionally excludes
unbounded latest/HEAD targets. Duplicate or unsafe names fail discovery.

`Library/Homebrew/env_config.rb` confirms this version treats `0` as false for
HOMEBREW_UPGRADE_GREEDY. HOMEBREW_UPGRADE_GREEDY_CASKS must also be cleared.
The fake tests assert effective child environment overrides, planned typed IDs,
exit-0 discovery, error exits, partial failure, consent and cleanup suppression.
Fake subprocess tests run the Go test executable with isolated PATH, HOME and
prefix; they cannot reach the host Homebrew executable.

Only the source-reviewed 7.0.10 version is accepted by the development adapter;
help capabilities are also required. This is not native installation evidence.
Actual empty/outdated discovery exit statuses and output, self-updating cask
eligibility, cleanup suppression during real installations, and controlled
formula/cask upgrades remain unverified. Acquire those fixtures in a disposable
Homebrew prefix/Mac, record OS version and architecture, then run native arm64
smoke tests before publication or broadening compatibility. No public
release compatibility is claimed.

## Native cask fixtures acquired 2026-10-27

`homebrew-7.0.10-darwin-arm64-empty.json` and
`homebrew-7.0.10-darwin-arm64-casks.json` were acquired from real Homebrew 7.0.10
(`8f10429b56be28a25032b87463fffc95e556b5bc`) on macOS 27.0.1 (26A434), arm64.
The [controlled native run](../../../../docs/validation/2026-10-27-homebrew-arm64.md)
used only isolated local script-cask fixtures, separate HOME/cache/prefix, and
file-only Git remotes. Acquisition was `brew outdated --json=v2`, exit 0 for both
empty and outdated results, with greedy settings disabled. The cask snapshot has
one eligible 1.0 → 2.0 candidate and one pinned cask; the self-updating cask is
absent. No installed user packages or private account data appear in fixtures.

The real updater also passed cask upgrades, explicit cleanup and a controlled
partial failure. These results supersede the earlier “unverified” statements
above for this narrow cask-fixture scope only. Formula builds were blocked by
Homebrew's inherited-sandbox guard, and real vendor application authentication,
formula keg-retention evidence, upmacx formula acceptance remain
pending. The original `mixed.json` remains explicitly synthetic.

## Full native fixture run acquired 2026-10-28

`homebrew-7.0.10-darwin-arm64-mixed.json` is an actual `brew outdated --json=v2`
result (exit 0) from the full protected-prefix fixture run, with one formula and
one cask eligible for 1.0 → 2.0 and pinned formula/cask exclusions. Self-updating
casks are absent. This is distinct from the original synthetic `mixed.json`.

[Full native evidence](../../../../docs/validation/2026-10-28-homebrew-full-arm64.md)
records successful source formula and cask fixture upgrades, check-only behavior,
old-keg retention without cleanup, explicit cleanup and partial failure with
cleanup skipped. The prior formula-build blocker was caused by locating the
manager in `/private/tmp`, which Homebrew's sandbox intentionally permits writes
to. A protected project-local prefix fixes it without disabling sandboxing.
Real vendor authorization and tagged upmacx source-formula acceptance remain
pending; earlier blanket claims that fixture formula builds were blocked are
historical and superseded for the corrected layout.
