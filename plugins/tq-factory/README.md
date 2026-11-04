# TQ Factory plugin

Japanese counterpart: [README.ja.md](README.ja.md).

TQ Factory is a coding-agent workflow plugin for projects that use Tasq's `tq`
CLI as their work tracker. It coordinates bounded work through issues and
comments while external agents or humans perform the work. It does not start
Tasq's legacy orchestrator, run an agent daemon, or generate documents through
an embedded model.

## Install for local development

The package includes a portable root `plugin.json`, a Codex compatibility
manifest, and a Claude manifest. To try it in a local Codex client, add this
package to a local or repository marketplace and install it from that source;
Codex supports `$REPO_ROOT/.agents/plugins/marketplace.json` for a repository
marketplace. Then invoke `tq-factory` for a multi-issue coding goal. This
package intentionally does not add a marketplace catalog or publish itself.

See [the official plugin packaging guide](https://developers.openai.com/plugins/build/plugins)
for marketplace shape and local-client availability.

The target project needs:

- a running issue-tracker reachable by `tq`;
- a registered project key (`tq project list`); and
- permission to read and write the target issues and comments.

Set `TQ_API_URL` or use `--api-url` for a tracker other than the local default.
Before writing, verify the endpoint with `tq --output json issue list --project
<project-key>`.

## Minimal flow

1. Ask the agent to use `tq-factory` for a concrete goal and project key.
2. It creates or resumes a goal issue and child issues in Tasq.
3. An external agent works on one assigned child in an isolated workspace and
   records progress and handoff evidence with `tq comment add`.
4. An independent reviewer records verification. An authorized owner closes the
   verified issue and updates the goal index.

Read [the skill](skills/tq-factory/SKILL.md) and its linked workflow reference
for exact command shapes and state rules.

## Offline validation scenario

Review [the scenario](skills/tq-factory/evals/offline-scenario.md) without
calling `tq`. It checks that a user request is turned into safe issue-state and
handoff decisions rather than a fictitious agent runtime.
