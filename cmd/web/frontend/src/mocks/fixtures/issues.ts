import type { Issue } from "@/lib/generated/issue-tracker";

export const issueFixtures: Issue[] = [
  {
    id: 1,
    projectId: 1,
    projectKey: "tasq",
    title: "Design standalone mock workflow",
    description: `## Goal

Define a standalone workflow that lets the web UI run without a local issue-tracker or orchestrator service. The mock layer should feel close enough to production that route, layout, and board behavior can be reviewed in the browser.

### Scope

- Map each screen to the API calls it needs.
- Keep fixture data realistic across multiple projects.
- Make refresh, create, and status-update flows visible during manual checks.

### Done when

The development notes explain how to start the mock UI, which handlers are active, and what a reviewer should verify before switching back to the real backend.`,
    status: "backlog",
    priority: "normal",
    assignee: "frontend",
    dependency_ids: [],
    createdAt: "2026-06-01T01:00:00.000Z",
    updatedAt: "2026-06-01T01:00:00.000Z",
  },
  {
    id: 2,
    projectId: 1,
    projectKey: "tasq",
    title: "Wire issue board to generated client",
    description: `## Goal

Load issue board data through the generated issue-tracker client so the standalone UI and production UI exercise the same contract. Mocking should happen at the HTTP boundary, not inside route components.

### Implementation notes

- Use the generated response types as the source of truth.
- Keep the board summary shape aligned with the OpenAPI schema.
- Avoid UI-only fixture shapes that would hide contract drift.

### Acceptance criteria

The board renders from generated client calls, status changes survive the next refresh, and type checking fails if the API response contract changes incompatibly.`,
    status: "ready",
    priority: "high",
    assignee: "web",
    dependency_ids: [],
    createdAt: "2026-06-01T02:00:00.000Z",
    updatedAt: "2026-06-01T02:00:00.000Z",
  },
  {
    id: 3,
    projectId: 1,
    projectKey: "tasq",
    title: "Verify status transitions",
    description: `## Goal

Verify that status transitions update both the individual issue and the board summary after refresh. This fixture is meant to catch regressions where a card appears to move locally but returns to the old column after the next poll.

### Checks

- Move an issue from Ready to In Progress.
- Confirm the mock state stores the new status.
- Wait for the summary refresh and confirm the card remains in the expected column.

### Notes

The behavior should match production semantics: the UI asks the API to update the issue, then reloads board data from the same source instead of mutating a separate local copy.`,
    status: "in_progress",
    priority: "normal",
    assignee: "qa",
    dependency_ids: [],
    createdAt: "2026-06-01T03:00:00.000Z",
    updatedAt: "2026-06-01T03:00:00.000Z",
  },
  {
    id: 4,
    projectId: 1,
    projectKey: "tasq",
    title: "Review detail page loading state",
    description: `## Goal

Review the issue detail page while running only the frontend mock worker. The detail route should load issue metadata, render Markdown descriptions, and fetch comments without depending on backend services.

### Expected behavior

- Direct navigation to the detail URL works after a page reload.
- Invalid issue IDs show the error panel.
- Comments render in chronological order with empty and Markdown bodies handled cleanly.

### Review focus

Use this issue to inspect spacing, metadata labels, status actions, and long-description wrapping in both the card preview and the full detail page.`,
    status: "review",
    priority: "low",
    assignee: "",
    dependency_ids: [],
    createdAt: "2026-06-01T04:00:00.000Z",
    updatedAt: "2026-06-01T04:00:00.000Z",
  },
  {
    id: 5,
    projectId: 2,
    projectKey: "agent-docs",
    title: "Outline agent onboarding guide",
    description: `## Goal

Create an onboarding guide that helps a new contributor run an agent-assisted workflow from a clean checkout. The document should be practical, command-oriented, and explicit about which steps are safe to run locally.

### Suggested sections

- Repository setup and prerequisites.
- Common task flow from branch creation to PR.
- How to interpret agent status messages.
- Troubleshooting notes for missing tools, stale worktrees, and failed checks.

### Done when

A new contributor can follow the guide, start the web UI, run the core verification commands, and understand when to ask before using destructive Git operations.`,
    status: "backlog",
    priority: "normal",
    assignee: "docs",
    dependency_ids: [],
    createdAt: "2026-06-01T05:00:00.000Z",
    updatedAt: "2026-06-01T05:00:00.000Z",
  },
  {
    id: 6,
    projectId: 2,
    projectKey: "agent-docs",
    title: "Review command reference examples",
    description: `## Goal

Review the command reference examples for accuracy against the current agent workflow. The docs should avoid stale flags, ambiguous prompts, and examples that imply unsafe Git or GitHub operations.

### Review checklist

- Command names and arguments match the current tooling.
- Examples distinguish read-only checks from write operations.
- Output snippets describe the important signal without copying noisy logs.
- Safety notes cover branch deletion, PR editing, and push behavior.

### Done when

The reference can be used during a live task without forcing the reader to cross-check every example against source code or recent chat history.`,
    status: "review",
    priority: "high",
    assignee: "docs",
    dependency_ids: [],
    createdAt: "2026-06-01T06:00:00.000Z",
    updatedAt: "2026-06-01T06:00:00.000Z",
  },
  {
    id: 7,
    projectId: 3,
    projectKey: "some-project",
    title: "Prepare sample project backlog",
    description: `## Goal

Prepare a small but believable backlog for the sample project so the all-project board does not look like a single-project demo. The fixture should make cross-project filtering obvious during visual review.

### Fixture intent

- Include work that belongs outside the Tasq product itself.
- Use a different assignee and priority mix from the Tasq issues.
- Keep titles specific enough that a reviewer can spot the project boundary at a glance.

### Done when

The /issues route shows this project alongside Tasq and agent-docs work, while /projects/some-project/issues narrows the board to only the sample-project cards.`,
    status: "ready",
    priority: "low",
    assignee: "product",
    dependency_ids: [],
    createdAt: "2026-06-01T07:00:00.000Z",
    updatedAt: "2026-06-01T07:00:00.000Z",
  },
  {
    id: 8,
    projectId: 3,
    projectKey: "some-project",
    title: "Validate scoped issue navigation",
    description: `## Goal

Validate the project-scoped issue navigation from the sidebar. Selecting a project should update the URL, active sidebar state, header title, issue count, and board contents without showing unrelated project cards.

### Manual checks

- Open /issues and confirm all mock projects are represented.
- Open /projects/tasq/issues and confirm only Tasq issues remain.
- Open /projects/some-project/issues and confirm this card is visible while Tasq cards are hidden.

### Edge case

If a project key is unknown, the board should not fall back to all issues; it should show an empty scoped board with the requested key as context.`,
    status: "in_progress",
    priority: "normal",
    assignee: "web",
    dependency_ids: [],
    createdAt: "2026-06-01T08:00:00.000Z",
    updatedAt: "2026-06-01T08:00:00.000Z",
  },
];
