# Instructions for all agents and models

These instructions apply to the entire repository and to every agent or model working here, including delegated agents.

## Required context

Treat [docs/](docs/) as the project's source of truth. Before making changes:

1. Read [the specification index](docs/specs/00-index.md).
2. Read [progress](docs/specs/08-progress.md) to establish the current state and next work.
3. Read all numbered specifications in order on your first task in this repository. On subsequent tasks, reread the documents relevant to the change and any documents changed since your last review.
4. Inspect the actual files and applicable instructions before relying on recorded progress.

Follow the requirements in `docs/` for implementation, interfaces, CLI behavior, provider commands, testing, delivery order and releases. Do not substitute assumptions from memory or copy tutorial code without checking the specifications. Give delegated agents the relevant documentation and require the same compliance.

## Working rules

- Work within the user's current request. The presence of a roadmap does not authorize implementing every phase or publishing anything.
- Follow the delivery phases in [07-step-by-step-delivery.md](docs/specs/07-step-by-step-delivery.md); keep unshipped providers unavailable.
- Preserve check-only behavior, opt-in cleanup, explicit system-install selection, and the prohibition on automatic reboot, dependency installation and privilege escalation.
- Use isolated fake tools for automated tests. Do not use the host's installed software as update test data. Follow [the testing specification](docs/specs/05-testing-and-acceptance.md).
- Do not invent owners, checksums, compatibility evidence, passing checks or completed releases.
- When a requested change affects the contract, update the affected specifications and tests together. Explicit user instructions and higher-priority instructions take precedence; record resulting specification changes. If documents conflict, resolve using the index's authority rules; ask only when a material decision remains unresolved.
- Distinguish historical context and future requirements from current implementation. A specification is not proof that a feature exists.

## Progress and handoff

Update [08-progress.md](docs/specs/08-progress.md) whenever work changes project status, completes a task, adds a blocker or produces validation evidence. Keep it factual and concise:

- Record completed work with links to real files and checks actually run.
- Mark work complete only when its documented acceptance criteria pass. Track implementation, validation and publication separately.
- Record failed or unrun checks and unresolved external prerequisites without claiming success.
- Leave a concrete next step and enough context for another agent to continue.
- Do not store credentials, private account data or full sensitive command output in progress notes.

Before finishing, verify documentation links, summarize changes and validation, and state any material remaining limits.
