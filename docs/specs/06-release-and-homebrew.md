# Release and Homebrew distribution

## Repositories and prerequisites

Future repositories: `https://github.com/OWNER/upmacx` for source, docs and releases; `https://github.com/OWNER/homebrew-tap` for the formula. The existing Git remote confirms source owner `htooaunglynn`; the user approved `htooaunglynn` as the MIT copyright holder. Tap availability, source checksum and binary checksums remain unresolved. `OWNER` in the templates is now a substitution for the confirmed source owner, not evidence a tap exists. Resolve repository/name availability before publishing. Do not infer GitHub identity from a filesystem username.

Adopt SemVer, starting at `v0.1.0`. `main.version` defaults to `dev`; tagged builds inject the exact `v`-prefixed tag. README must identify shipped providers and unsupported capabilities per release. Add actual MIT LICENSE before the first tag. Git, Go and GitHub CLI are maintainer tools; brew is required only for brew integration. Other providers remain optional.

## Build and artifact contract

From a clean tested tagged checkout, build with `CGO_ENABLED=0`, `GOOS=darwin`, `GOARCH=arm64`, `-trimpath` and `-ldflags="-s -w -X main.version=TAG"`, targeting `./cmd/upmacx`. Replace TAG with the real tag; do not publish a `dev` binary. Do not execute foreign-architecture binaries as if cross-builds were native validation.

Artifacts: `upmacx_TAG_darwin_arm64.tar.gz` and `checksums.txt`. Each archive contains executable `upmacx`, README and LICENSE at its root. Generate SHA-256 lines over the final archives with `shasum -a 256`; verify them again before upload. Source archive checksum is independently calculated from the exact tag URL consumed by Homebrew.

Archive determinism is a build-system goal, not an unverified claim; record compiler version and build inputs in release notes. Code signing/notarization is deferred; don't claim either. Test direct-download execution on controlled Macs and document any platform-required manual steps honestly.

## Local preparation helper

[releaseprep](../../cmd/releaseprep/main.go) implements local preparation with
[isolated tests](../../internal/release/release_test.go). After native acceptance,
from a clean tagged checkout, provide the exact downloaded GitHub tag archive:

```sh
go run ./cmd/releaseprep --tag v0.1.0 \
  --source-archive /absolute/path/to/upmacx-v0.1.0-source.tar.gz \
  --out /absolute/path/to/new-artifact-directory
```

The helper rejects a dirty checkout, missing/mismatched tag, existing output,
unformatted code, failed tests/vet/builds, or a source archive differing from
tracked files. It rejects source traversal, links, duplicates, missing/extra
files, and corrupt gzip data. Generated output includes the arm64 binary
archive (root entries `upmacx`, `README`, `LICENSE`), disk-verified SHA-256 lines,
build-info metadata and `Formula/upmacx.rb` using the supplied source archive's
independent SHA-256. Output is staged before the complete directory is exposed.

The source archive is supplied by the maintainer from the exact formula URL;
matching content does not authenticate where its bytes were downloaded. No
network fetch, dependency installation, tagging, tap update or publication is
performed by the helper. It does not attest that native acceptance passed.
The helper does not create a GitHub draft release. The [tag workflow](../../.github/workflows/release.yml) runs preparation with read-only permissions and grants release-content writes only to a separate draft job. It never promotes a public release or writes to a tap. The generated formula remains in the workflow artifact for maintainer validation; it is not uploaded as an installable tap change. This workflow is authored but has not run on GitHub.

## Ordered publishing procedure

1. Run phase acceptance gates, check clean source state, finalize README/LICENSE/changelog and version. Commit before tagging.
2. Create and push immutable `vX.Y.Z` tag. Initial releases use manual maintainer publication; CI validates tags and builds/uploads artifacts to a draft release, without automatic public promotion or tap writes.
3. Inspect draft artifacts, verify SHA-256, inspect archive contents and native version output. Publish the release only after these checks. Never move a published tag to repair it; create a new patch release.
4. Download `https://github.com/OWNER/upmacx/archive/refs/tags/TAG.tar.gz`, verify contents, compute its SHA-256 and put that value in the tap formula. Never reuse a binary archive checksum.
5. Audit and install/test the formula in an isolated Homebrew environment, then commit/push the tap change. No tap credentials are needed by source CI under this initial manual flow.
6. Verify public installation instructions from a clean environment. Record evidence in release notes. Future updates repeat the same sequence and replace URL/checksum together.

GitHub repository creation, authentication, public release and push are future maintainer actions; these specs do not perform them. CI should use read-only permissions for pull requests and grant release-content writes only to the tag workflow's draft-release job. No secrets in forks, archives, output or fixtures.

## Formula template contract

The tap contains `Formula/upmacx.rb` with class `Upmacx`, description, source homepage, immutable tag URL, real source `sha256`, `license "MIT"`, `depends_on "go" => :build`, `depends_on :macos`, and `depends_on arch: :arm64`. No linuxbrew dependency, optional-provider dependencies, updater execution in install, or bottles initially.

Its install method invokes Go using separate arguments:

```ruby
system "go", "build", "-trimpath",
       "-ldflags=-s -w -X main.version=v#{version}",
       "-o", bin/"upmacx", "./cmd/upmacx"
```

The test block asserts exact `upmacx v#{version}` output from `--version` and help text from `--help`; neither invokes providers. Keep version inference from the `vX.Y.Z` source URL consistent with the injected prefix.

Future tap validation commands (substitute real owner):

```sh
brew audit --strict OWNER/tap/upmacx
brew install --build-from-source OWNER/tap/upmacx
brew test OWNER/tap/upmacx
upmacx --version
```

The formula supports only Apple Silicon, enforced with `depends_on arch: :arm64`. It must be source-built and tested natively on arm64. Do not generate or publish Intel artifacts.

## User installation and updates

```sh
brew install OWNER/tap/upmacx
upmacx --check
upmacx
# Later, update upmacx itself through its distribution manager:
brew update
brew upgrade OWNER/tap/upmacx
```

README also explains `--providers`, `--clean`, `--system`, provider prerequisites, noninteractive limitations, authentication, restart/manual completion, failure exits and unsupported apps. Explain that the brew provider can update upmacx itself as a formula; no separate self-update downloader is implemented. A running invocation continues its original process and does not re-exec a newly installed version.

References: [Homebrew tap guide](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap), [Formula Cookbook](https://docs.brew.sh/Formula-Cookbook), [Go build command](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies). This file defines the project's release policy, rather than claiming publication has occurred.
