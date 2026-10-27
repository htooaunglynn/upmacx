# Homebrew compatibility evidence

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
and amd64 smoke tests before publication or broadening compatibility. No public
release compatibility is claimed.
