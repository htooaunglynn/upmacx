# Changelog

## Unreleased — Phase 1 development

- Fix the native test-prefix layout: reject broadly writable temporary/cache
  roots and validate real formula/cask upgrades under normal Homebrew sandboxing.

- Scope changed to Apple Silicon only: reject Intel execution, build/publish only
  arm64, restrict the Homebrew formula to arm64, and remove Intel CI/gates.

- Add Go CLI, Homebrew plan/discovery, explicit upgrades and opt-in cleanup.
- Add safe confirmation, check-only execution, signal cancellation and summaries.
- Add isolated fake-tool tests, Homebrew 7.0.10 source-reviewed fixtures and CI.
- Preserve earlier failure summaries during cancellation and honor the full
  descendant termination grace period.
- Notify cancellation promptly when an interrupted child leaves descendant output
  pipes open; verify terminal input, Ctrl-C and foreground restoration on macOS.
- Report Homebrew probe timeouts, overflows and execution errors as failures.
- Add local release preparation with clean-tag/source verification, archives,
  separate source/binary checksums and generated source formula.
- Enforce canonical provider ordering and clone discovered/approved plans.
- Report stage progress and interrupted cleanup while preserving completed IDs.
- Accept reviewed Homebrew version headers with trailing tap-version details.
- Add opt-in isolated native Homebrew acceptance.
- Provider installation, real tagged release archives and
  Homebrew formula audit/install/test remain pending. No release is published.
