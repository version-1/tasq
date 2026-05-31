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

## Repository Guidance

When implementing Symphony-related orchestration, workflow, workspace, agent-runner, tracker, or observability behavior in this repository, treat [SPEC.md](SPEC.md) as the governing contract.

If upstream `SPEC.md` is refreshed, update both local files in this directory and keep the Japanese translation synchronized.
