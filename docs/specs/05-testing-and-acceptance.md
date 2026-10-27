# Testing and acceptance

## Isolation and automated checks

Implement tests before each provider is enabled. Inject Runner, streams, terminal predicate, OS/UID, clock and signal controller. Unit tests never invoke real package managers. Process integration tests re-execute the Go test binary as a fake executable in a temporary directory; record argv, environment and invocation order to a temporary log. The registry, including softwareupdate's fixed production path, must be injectable in tests so no test can reach the real system command.

Set fake PATH and temporary package-manager homes/prefixes for subprocess tests. Assert every requested command is fake; fail closed on an unexpected executable. Network access is unnecessary. No live update smoke test runs in the normal CI suite, and `--check` against the host is not an automated test substitute.

CI from phase 1: `go test ./...`, `go vet ./...`, formatting check without rewriting files, and `CGO_ENABLED=0 GOOS=darwin GOARCH=arm64` / `amd64` builds. Run orchestration/process tests on macOS to verify signal groups. On Darwin, include a real pseudo-terminal test using only the fake Go test executable: verify child input, Ctrl-C, signal propagation and foreground restoration. This does not replace controlled installer/authentication tests. Include an interrupted leader whose signal-ignoring descendant retains stdout/stderr, and assert prompt cancellation rather than a capture timeout. Use a Linux job for portable parsing/tests if convenient; Darwin process code needs build tags and actual macOS coverage. CI explicitly includes `macos-15-intel` as well as the arm64 macOS runner and performs native help/version smoke tests. Runner architecture labels are documented by [GitHub](https://docs.github.com/en/actions/reference/runners/github-hosted-runners); authored jobs are not evidence of a hosted run. Cross-build success does not prove an Intel binary runs; native architecture smoke evidence is required for public support claims.

## CLI and orchestrator acceptance matrix

| Case | Required observation |
| --- | --- |
| Help/version | Exact output contract; zero provider subprocesses |
| Invalid args/selection/flag combination | Exit 2 before detection |
| Non-Darwin/root execution | Clear refusal, no provider operations; help/version still work |
| No outdated candidates | Exit 0; no upgrade; cleanup only if requested and confirmed |
| Check with updates | Exit 0, displayed candidates, no upgrade/cleanup/implicit maintenance |
| Yes/y with case/whitespace | Exactly planned operations execute |
| No/empty/EOF/other answer | No mutation, exit 0 unless discovery already failed |
| Nonterminal without --yes | If work exists exit 2 without reading stdin; otherwise exit 0 |
| --yes | Skip only upmacx prompt; native terminal requirement still enforced |
| Missing default provider | Visible unavailable/skip; remaining providers continue |
| Missing explicit provider | Failure summary and exit 1; independent work may continue |
| Refresh/discovery fails | Failed provider never applies; others proceed |
| Apply fails after earlier success | Partial outcome preserved; no cleanup for failed provider |
| Cleanup fails | Earlier upgrades retained, brew failed, exit 1 |
| Multiple providers | Discovery precedes all mutation, fixed ordering and single consent |
| SIGINT/SIGTERM | No next operation, child group signaled/reaped; 130/143 |
| Hung child | Capture timeout or cancellation kills group after grace period |
| Malformed/oversized output | Failure, never broad update fallback |
| Unsafe candidate text | Reject option-like IDs; no shell execution/control-character rendering |
| All providers unavailable | Honest coverage summary, exit 0 in default mode |

## Provider-specific scenarios

| Provider | Fixture and command assertions |
| --- | --- |
| brew | Formula/cask schemas, pin filters, greedy env override, cleanup suppression, explicit typed IDs, cleanup opt-in |
| softwareupdate | Labels with spaces, no updates, restart metadata, unrecognized listing, default report-only, --system install only without restart/force/password flags |
| mas | Multiple JSON objects rather than array assumption, ID validation, version fields, no accurate discovery, account/Spotlight/privilege error |
| npm | Scoped name, wanted versus latest, linked/git exclusion, global ownership, status 1 with valid data versus error JSON |
| pnpm | Actual schema fixtures, compatible/pinned targets, global-only argv, manager exclusion, setup error |
| pipx | Modern capability detection, outdated-short/venv join, pins/editables/injected exclusions, --skip-maintenance on check/apply |
| cargo | Helper absent, listing table, source classification, rustup/helper exclusions, version-qualified update and compile failure |

For each provider include zero/one/many candidates, malformed records, nonzero stderr errors, timeout and interrupted mutation. Assert cleanup is never invoked by any provider during `--check`; assert no test contains a live install executable path.

## Manual release acceptance

Use disposable test Macs/VMs or explicitly designated test accounts/prefixes, never an ordinary developer machine's software as test data. Capture tested versions and outcomes without credentials. Verify Apple Silicon and Intel executable help/version, one controlled candidate per provider, authorization denial and a partial failure. System-update install testing requires a disposable supported OS environment and observation that no automatic reboot occurs; otherwise ship discovery-only support on that OS and state the limitation.

In a clean Homebrew environment, audit, build from tagged source, run formula test and verify help/version. Real `upmacx --check` is a separate controlled network/metadata smoke test, not part of the formula test. Verify installing the formula itself never executes the updater. Publishing requires the phase's automated gates plus documented native compatibility evidence. No live acceptance operations are performed as part of creating these specs.
