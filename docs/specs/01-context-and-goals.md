# Context and goals

## Source and intent

Source: [Create Homebrew Package — shared conversation](https://chatgpt.com/share/6adf574d-c100-83ea-98e2-6ec5d1f90877). The complete visible conversation was read during planning. This document summarizes it; the tutorial code is not an accepted implementation.

The user wants one command to update supported software on a Mac and to learn the complete Go-to-Homebrew publishing process. The working name is `upmacx`, with the description “Your macOS update assistant.” Audience: Mac users comfortable running terminal commands, using Apple Silicon Macs only. Intel support was explicitly removed by the user on 2026-10-28.

Success means a user can inspect available updates, understand the planned changes, confirm once, execute supported updates, and receive an honest per-provider outcome. Installation through a custom Homebrew tap and reproducible release instructions are required. “Every installed app” is an aspiration limited by available package-manager coverage, permissions, compatibility, and account access.

## Requirement traceability

| Source requirement | Specification destination |
| --- | --- |
| Go CLI, one command, name | 02, 03 |
| Formulae and casks | 04 Homebrew, phase 1 |
| Confirmation, check mode, version | 02, 05 |
| Cleanup | 02, 04; revised to opt-in |
| macOS updates and restart considerations | 04 softwareupdate, phase 2 |
| App Store through mas | 04 mas, phase 3 |
| npm, pnpm, pipx, Cargo | 04 developer providers, phase 4 |
| Project structure and implementation guide | 03, 07 |
| Apple Silicon arm64 binary and checksums | 06 |
| GitHub releases, tap, installation, future upgrades | 06 |
| Tests, CI, better output, errors and recovery | 03, 05, 07 |
| Manually installed DMG applications | Unsupported; use vendor updater |
| Persistent logs and richer progress UI | Deferred; streamed output and summary are required now |

## Corrections and accepted refinements

- The tutorial defaults to cleanup. The user selected opt-in cleanup instead. Suppress Homebrew's automatic install cleanup as well as omitting an explicit cleanup command.
- The final formula must use `depends_on :macos`; the illustrative `linuxbrew` dependency is invalid for this project and must never be copied.
- Homebrew formula source checksums and downloadable binary checksums are different artifacts.
- Add the actual MIT license before tagging a release; an example license declaration alone is insufficient.
- Broader provider coverage is now specified in detail, but phased delivery remains approved.
- System discovery is normal; installation is explicitly selected and never includes a reboot flag.
- No single success message may imply all Mac software is updated. Provider detection and update checks can be incomplete or fail.

## Boundaries

Only software owned by the selected package manager is eligible. npm/pnpm operate on global installations in their active prefix; pipx operates on the current user's app environments; Cargo operates on installed executable crates. Do not traverse projects, other users, alternate Node versions, or arbitrary application directories. Do not install dependency tools, change accounts, unpin packages, force casks, or repair package-manager state automatically.
