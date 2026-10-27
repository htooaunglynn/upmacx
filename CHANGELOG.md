# Changelog

## Unreleased — Phase 1 development

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
- Add explicit native Intel CI and opt-in isolated native Homebrew acceptance.
- Provider installation, native Intel validation, real tagged release archives and
  Homebrew formula audit/install/test remain pending. No release is published.
