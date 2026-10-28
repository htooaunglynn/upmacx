# upmacx

Your macOS update assistant for **Apple Silicon (arm64) only**. Intel Macs are unsupported. Phase 1 development implementation covers Homebrew
formulae and eligible casks. No public release or Homebrew tap is available yet.
Native formula/cask fixture tests pass; vendor authorization and tagged upmacx formula acceptance remain pending.

## Build and inspect

Requires Go 1.27.3 and Git. The CLI uses only the Go standard library.

```sh
go build -trimpath -o dist/upmacx ./cmd/upmacx
./dist/upmacx --help
./dist/upmacx --version
```

Help and version work without Homebrew and on other operating systems. Updater
execution requires macOS on arm64 and a non-root user. Only the source-reviewed Homebrew
7.0.10 adapter is currently accepted; other versions report unavailable until
compatibility is reviewed. Homebrew must already be installed; follow the
[official Homebrew instructions](https://brew.sh/) manually if needed. upmacx
never installs its own dependencies.

```sh
./dist/upmacx --check                 # refresh metadata and display a plan
./dist/upmacx                         # inspect, confirm once, then upgrade
./dist/upmacx --providers brew --yes   # explicit selection, skip upmacx prompt
./dist/upmacx --clean                 # include opt-in cleanup in confirmation
```

`--check` may use the network and update Homebrew metadata/cache. It never
upgrades packages or runs cleanup. `--check --clean` is rejected. Cleanup is
suppressed during other Homebrew operations and only explicitly run after a
successful Homebrew stage, with consent. It can run even when no upgrades exist.
Declining the prompt cancels every planned change, including cleanup.
Nonterminal input requires `--yes` when changes are planned; it is never read
for confirmation. Only `y` and `yes` (case-insensitive, trimmed) approve changes.

Pinned packages are excluded. Normal non-greedy Homebrew discovery excludes
self-updating/latest casks. Explicit planned identifiers are upgraded one at a
time; no broad upgrade fallback exists. Homebrew may also change dependencies
and run installer scripts, and the displayed target versions are previews.
Cask installers may request authentication or require an app to close.
`--yes` only skips the upmacx prompt; it does not provide credentials, accept
licenses or guarantee unattended installation. upmacx never initiates privilege
escalation, repairs manager state, or reboots.

Only Homebrew-managed software is covered. `softwareupdate`, `mas`, `npm`,
`pnpm`, `pipx`, `cargo` and `--system` remain unavailable in this phase.
Manually installed applications must use their vendor's updater. A Homebrew
installation of upmacx can eventually be upgraded through Homebrew; the running
process continues its original version without re-executing.

The final summary reports provider outcomes. Failures may follow successful
changes; no rollback is promised. On interruption, inspect Homebrew state before
rerunning. Exit codes: 0 success/no work/decline, 1 runtime/provider failure,
2 invalid invocation or missing confirmation, 130 SIGINT, 143 SIGTERM.

## Development checks

```sh
go test ./...
go test -race ./...
go vet ./...
test -z "$(gofmt -l cmd internal)"
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -o dist/upmacx-arm64 ./cmd/upmacx
```

Darwin process tests also use a real pseudo-terminal to verify fake-child input,
Ctrl-C and terminal foreground restoration. This is not real installer or
authorization acceptance evidence.

Tests use injected providers and the Go test binary as fake tools with isolated
PATH/homes/prefixes. They never update host software. CI runs on Linux and macOS;
the arm64 target is built. Intel builds and acceptance are out of scope.
See [compatibility evidence](internal/providers/brew/testdata/README.md) for the
difference between source-reviewed fixtures and outstanding native tests.

The [specification index](docs/specs/00-index.md),
[progress](docs/specs/08-progress.md), and [release procedure](docs/specs/06-release-and-homebrew.md)
describe requirements and remaining gates. Tagged archives, checksums and a
separate tap formula require native acceptance and an immutable release tag.
No fabricated source checksum or installable placeholder formula is included.

Licensed under [MIT](LICENSE).

## Prepare a release locally

After native acceptance, commit and tag a clean checkout using the release
procedure. Download the exact GitHub tag source archive separately. Then run:

```sh
go run ./cmd/releaseprep --tag v0.1.0 \
  --source-archive /absolute/path/to/upmacx-v0.1.0-source.tar.gz \
  --out /absolute/path/to/new-artifact-directory
```

The helper requires an existing exact tag at HEAD, no tracked/untracked changes,
a new output directory, and source archive contents matching every tracked file.
It runs tests, vet and a formatting check, builds Darwin arm64 with
the tag injected, and generates archives, verified binary checksums, build input
metadata and `Formula/upmacx.rb`. The formula checksum is calculated from the
supplied source archive, separately from binary checksums. Archive roots contain
`upmacx`, `README` (from README.md), and `LICENSE`.

This command does not download dependencies, create tags, publish a release or
change a tap. Run the generated formula's audit/install/test in isolated native
Homebrew environments before publishing. The source archive must be the exact
bytes obtained from the formula URL; content validation does not authenticate
its download origin. No real project release tag or source checksum exists yet.

The tag workflow prepares artifacts and creates a GitHub **draft** only after its
checks pass. Maintainers must verify native acceptance and review the draft
before public promotion. Tap updates remain manual. The workflow has not yet
been validated on GitHub.

## Phase 1 acceptance status

The [acceptance ledger](docs/validation/phase-1-acceptance.md) lists each implemented
component, its evidence, and the remaining native/release gates. The opt-in
`native_brew_acceptance.py` command there creates local-only formula/cask fixtures
in a new isolated prefix. Keep its root outside `/private/tmp`, `/private/var/tmp` and macOS temporary
cache directories so Homebrew can protect its executable. Homebrew's sandbox
remains enabled. Only Apple Silicon CI is configured; hosted runs are pending.

Native harness layout regressions can be run without Homebrew:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts -p 'test_native_brew_paths.py'
```
