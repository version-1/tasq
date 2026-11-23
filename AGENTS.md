Respond in Japanese unless instructed otherwise.

# Development Style

When KPI or coverage goals are provided, keep iterating until they are met.
Ask questions to clarify ambiguous instructions.

# Branch and Release Workflow

- Use `develop` as the default base branch when creating Tasq pull requests. Use `main` as the base for pull requests that promote release-ready changes from `develop`.
- Keep the tip of `main` ready for release. Create official `tq` releases from the tip of `origin/main`.
- Use `develop` for rapid proof-of-concept and beta work on `tqdev`. Base that work on `develop` and target `develop` in its pull requests.
- Promote changes from `develop` to `main` when they are ready for an official `tq` release.

# Documentation

- Write documentation in English and Japanese.
  - English documentation is the primary source of truth for development and design decisions.
  - All documents linked from AGENTS.md must also be written in English.
  - Create two documents for each topic, one in English and one in Japanese, and keep them synchronized. See [docs/development.md](docs/development.md) for the pairing rule and its exceptions.
- See [docs/development.md](docs/development.md) for the repository documentation index, task flow, and documentation update rules.
- See [docs/design.md](docs/design.md) for the design documentation index.
- Symphony-related orchestration, workflow, workspace, agent-runner, tracker, and observability changes must comply with [docs/symphony/SPEC.md](docs/symphony/SPEC.md).
- Do not change [docs/symphony/SPEC.md](docs/symphony/SPEC.md) to match Tasq-specific implementation differences. Record any intentional Tasq-vs-Symphony differences in [docs/symphony/DEVIATIONS.md](docs/symphony/DEVIATIONS.md) instead.
- See [docs/symphony/CODEX_APP_SERVER.md](docs/symphony/CODEX_APP_SERVER.md) for the Tasq Codex app-server start/resume contract.
- See [cmd/web/frontend/docs/design.md](cmd/web/frontend/docs/design.md) for frontend routing and component placement rules.

# Code Design

- Maintain separation of concerns.
- Separate state from logic.
- Prioritize readability and maintainability.
- Replace conditional branching with class structures or interfaces when it separates behavior by responsibility.
- Keep functions small and split them into testable units.
- Define contract layers (APIs and types) strictly, and keep implementation layers regenerable.

# Tools

- gh: GitHub CLI
