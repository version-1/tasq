# Tasq Design

Tasq design documentation is split by concern:

- [Architecture](design/architecture.md): ownership boundaries, components, dependencies, state ownership, and current MVP behavior.
- [API](design/api.md): issue-tracker API surface and CLI command mapping.
- [Configuration](design/configuration.md): local home directory, runtime state, and Compose development configuration.
- [Schema](design/schema.md): entity field specifications and validation rules.
- [Status](design/status.md): issue status, derived queue status, and expected workflow transitions.
- [Operations](design/operations.md): local development environment, verification, and open decisions.
- [Deployment](design/deployment.md): release tag creation, GitHub Actions, and GoReleaser deployment flow.
- [Release Binary Startup](design/release-binary-startup.md): binary-only startup requirements for the README Getting Started flow.

For Web UI structure and styling conventions, see [design/web.md](design/web.md).
