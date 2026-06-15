Respond in Japanese unless instructed otherwise.

# Development Style

When KPI or coverage goals are provided, keep iterating until they are met.
Ask questions to clarify ambiguous instructions.

# Documentation

- Write documentation in English and Japanese.
  - English documentation is the primary source of truth for development and design decisions.
  - All documents linked from AGENTS.md must also be written in English.
  - Create two documents for each topic, one in English and one in Japanese, and keep them synchronized.
- See [docs/development.md](docs/development.md) for repository workflow, task flow, documentation update rules, and component workflow links.
- See [docs/design.md](docs/design.md) for the design documentation index.
  - [docs/design/architecture.md](docs/design/architecture.md): system architecture and ownership boundaries.
  - [docs/design/api.md](docs/design/api.md): issue-tracker API surface and CLI command mapping.
  - [docs/design/operations.md](docs/design/operations.md): local development environment, verification, and open decisions.
- Symphony-related orchestration, workflow, workspace, agent-runner, tracker, and observability changes must comply with [docs/symphony/SPEC.md](docs/symphony/SPEC.md).
- Do not change [docs/symphony/SPEC.md](docs/symphony/SPEC.md) to match Tasq-specific implementation differences. Record any intentional Tasq-vs-Symphony differences in [docs/symphony/DEVIATIONS.md](docs/symphony/DEVIATIONS.md) instead.
- See [docs/design/schema.md](docs/design/schema.md) for entity field specifications and validation rules.
- See [docs/design/web.md](docs/design/web.md) for Web UI structure and styling conventions.
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
