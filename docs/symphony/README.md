# Symphony Specification

This directory stores a local copy of the Symphony service specification used by this repository.

## Source

- Upstream repository: https://github.com/openai/symphony
- Upstream document: https://github.com/openai/symphony/blob/main/SPEC.md
- Raw source used for `SPEC.md`: https://raw.githubusercontent.com/openai/symphony/main/SPEC.md
- Retrieved on: 2026-05-31

## Files

- [SPEC.md](SPEC.md): English source specification copied from upstream.
- [SPEC.ja.md](SPEC.ja.md): Japanese translation maintained alongside the English copy.
- [DEVIATIONS.md](DEVIATIONS.md): Tasq-specific deviations from the upstream specification.
- [DEVIATIONS.ja.md](DEVIATIONS.ja.md): Japanese translation of the Tasq-specific deviations.
- [CODEX_APP_SERVER.md](CODEX_APP_SERVER.md): Tasq's Codex app-server transport and JSON-RPC contract.
- [CODEX_APP_SERVER.ja.md](CODEX_APP_SERVER.ja.md): Japanese translation of the Codex app-server contract.
- [WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md): Supported Tasq workflow front matter fields and prompt template guide.
- [WORKFLOW_CONTRACT.ja.md](WORKFLOW_CONTRACT.ja.md): Japanese translation of the workflow contract.
- [ENTITY_MAPPING.md](ENTITY_MAPPING.md): Mapping between Symphony SPEC domain model and Tasq entities.
- [ENTITY_MAPPING.ja.md](ENTITY_MAPPING.ja.md): Japanese translation of the Tasq entity mapping.

## Repository Guidance

When implementing Symphony-related orchestration, workflow, workspace, agent-runner, tracker, or observability behavior in this repository, treat [SPEC.md](SPEC.md) as the governing contract.

If upstream `SPEC.md` is refreshed, update both local files in this directory and keep the Japanese translation synchronized.
