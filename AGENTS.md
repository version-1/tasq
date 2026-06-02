Respond in Japanese unless instructed otherwise.

# Development Style

When KPI or coverage goals are provided, keep iterating until they are met.
Ask questions to clarify ambiguous instructions.

# Documentation

- Write documentation in English.
- See [docs/design.md](docs/design.md) for the design documentation index.
  - [docs/design/architecture.md](docs/design/architecture.md): system architecture and ownership boundaries.
  - [docs/design/api.md](docs/design/api.md): issue-tracker API surface and CLI command mapping.
  - [docs/design/operations.md](docs/design/operations.md): local development environment, verification, and open decisions.
- Symphony-related orchestration, workflow, workspace, agent-runner, tracker, and observability changes must comply with [docs/symphony/SPEC.md](docs/symphony/SPEC.md).
- Do not change [docs/symphony/SPEC.md](docs/symphony/SPEC.md) to match Tasq-specific implementation differences. Record any intentional Tasq-vs-Symphony differences in [docs/symphony/DEVIATIONS.md](docs/symphony/DEVIATIONS.md) instead.
- See [docs/schema.md](docs/schema.md) for entity field specifications and validation rules.
- See [web/docs/design.md](web/docs/design.md) for Web UI structure and styling conventions.
- All documents linked from AGENTS.md must also be written in English.
- When a Japanese version of the same content is needed, create it as `*.ja.md` in the same location as the English version.
- Keep the English and Japanese versions synchronized.

# Code Design

- Maintain separation of concerns.
- Separate state from logic.
- Prioritize readability and maintainability.
- Replace conditional branching with class structures or interfaces when it separates behavior by responsibility.
- Keep functions small and split them into testable units.
- Define contract layers (APIs and types) strictly, and keep implementation layers regenerable.

# Tools

- gh: GitHub CLI
