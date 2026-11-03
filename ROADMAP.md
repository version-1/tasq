# Tasq Roadmap

Japanese counterpart: [ROADMAP.ja.md](ROADMAP.ja.md).

## Direction

Tasq will focus on being an AI-friendly, local-first issue tracker: a shared place for humans and AI agents to record work objectives, state, decisions, and outcomes. External execution tools should be able to use the CLI and API without adopting a Tasq-managed agent lifecycle.

The transition will happen in this repository. The repository is not being archived, and orchestrator separation is not a prerequisite for tracker development. Stages below are milestones, not scheduled release dates or promises of features already available.

## 1. Announce the direction

- Publish the tracker-first direction in the README and this roadmap.
- Pause new orchestrator features and additional Symphony conformance work. Existing orchestration contracts continue to apply to maintenance changes.
- Keep the current orchestrator available during the transition. Support is limited to fixes for security, data integrity, and regressions affecting existing workflows, without a guaranteed response time or compatibility with every future agent runtime.
- Use `v0.4.4-dev.20261103T1511.9131446` as the pre-transition prerelease reference. It does not yet provide the planned tracker-only default.

## 2. Make standalone tracker operation complete

- Make installation, startup, CLI use, and Web UI use work without an orchestrator process or agent credentials.
- Make tracker-only operation the documented default for new users, with an explicit migration path for existing installations.
- Prioritize issues, dependencies, comments, attachments, and stable CLI/API contracts. Evaluate additional handoff needs through actual use rather than rebuilding agent execution management inside the tracker.

## 3. Validate external-agent workflows

Use external agents with `tq` in routine development and check that:

- A fresh session can understand the objective, acceptance criteria, remaining work, and blockers from an issue.
- Humans and agents can identify responsibility and avoid conflicting updates or duplicate work.
- Progress, decisions, verification results, and links to outcomes remain available after a session ends.
- Existing installations can follow the documented migration steps and retain issue data and IDs.

## 4. Decide the orchestrator's future

The next decision checkpoint is after standalone tracker validation and before a stable release that makes tracker-only operation the default. Publish the validation results, known orchestrator usage and maintenance needs, and the selected path in this roadmap and release notes.

- Keep it optional if continued in-repository maintenance has clear value.
- Separate it if continued usage and an identified maintainer justify its own releases and compatibility support.
- Retire it if continued maintenance is not justified, documenting the last supported version and migration options.

No separation, retirement, removal date, or next stable version is committed at this stage. If validation is incomplete, report the gaps at the checkpoint rather than treating removal as automatic.

## Compatibility during the transition

This announcement changes documentation only. Current commands, packaging, persisted data, and runtime behavior remain unchanged.

Preserving issue data and IDs is a migration priority. Before removing orchestrator functionality or making breaking CLI/API, startup, packaging, or data-format changes, publish the affected versions, migration steps, and support boundaries in release notes. Changes that affect stored data must include backup and recovery guidance. Do not require users to discover the transition through a broken existing setup.
