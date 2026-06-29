import type {
  Comment,
  CommentListResponse,
  CreateIssueInput,
  CreateProjectInput,
  Issue,
  Priority,
  IssueState,
  IssueStatus,
  Project,
  ProjectWorkflow,
  QueueStatus,
  Summary,
  UpdateIssueInput,
} from "@/lib/generated/issue-tracker";
import type {
  ConversationResponse,
  IssueRuntimeResponse,
  RunState,
} from "@/lib/generated/orchestrator";
import { commentFixtures } from "./fixtures/comments";
import { issueFixtures } from "./fixtures/issues";
import { projectFixtures } from "./fixtures/projects";
import { summaryColumnFixtures } from "./fixtures/summary";
import { workflowFixtures } from "./fixtures/workflows";

type IssueListFilters = {
  assignee?: string;
  limit?: number;
  offset?: number;
  projectID?: number;
  projectIDs?: number[];
  priorities?: Priority[];
  search?: string;
  sortBy?: "id" | "priority" | "created_at" | "updated_at";
  sortDirection?: "asc" | "desc";
  states?: IssueStatus[];
};

let projects = clone(projectFixtures);
let issues = clone(issueFixtures);
const workflows = new Map<number, ProjectWorkflow>(
  workflowFixtures.map((workflow) => [workflow.projectId, workflow]),
);
let comments = clone(commentFixtures);
let nextProjectID = nextNumericID(projects);
let nextIssueID = nextNumericID(issues);
let nextCommentID = nextNumericID(comments);

export function listProjects(): Project[] {
  return clone(projects);
}

export function createProject(input: CreateProjectInput): Project | null {
  const key = input.key.trim();
  const name = input.name.trim();
  const location = input.location.trim();
  const description = input.description?.trim() ?? "";

  if (key === "" || name === "" || location === "") {
    return null;
  }

  if (projects.some((project) => project.key === key)) {
    return null;
  }

  const now = new Date().toISOString();
  const project: Project = {
    id: nextProjectID,
    key,
    name,
    description,
    location,
    workflowChecksum: "",
    createdAt: now,
    updatedAt: now,
  };

  nextProjectID += 1;
  projects = [...projects, project];
  return clone(project);
}

export function getProjectWorkflow(projectID: number): ProjectWorkflow | null {
  return clone(workflows.get(projectID) ?? null);
}

export function listIssues(filters: IssueListFilters = {}): Issue[] {
  const filtered = issues.filter((issue) => matchesIssueFilters(issue, filters));
  const sorted = sortIssues(filtered, filters.sortBy ?? "updated_at", filters.sortDirection ?? "desc");
  if (filters.limit === undefined) {
    return clone(sorted);
  }
  const offset = filters.offset ?? 0;
  return clone(sorted.slice(offset, offset + filters.limit));
}

function sortIssues(
  items: Issue[],
  sortBy: "id" | "priority" | "created_at" | "updated_at",
  direction: "asc" | "desc",
): Issue[] {
  const factor = direction === "asc" ? 1 : -1;
  return [...items].sort((a, b) => {
    const primary = compareIssueField(a, b, sortBy);
    if (primary !== 0) {
      return primary * factor;
    }
    return (a.id - b.id) * factor;
  });
}

function compareIssueField(a: Issue, b: Issue, sortBy: "id" | "priority" | "created_at" | "updated_at"): number {
  switch (sortBy) {
    case "id":
      return a.id - b.id;
    case "priority":
      return priorityRank(a.priority) - priorityRank(b.priority);
    case "created_at":
      return a.createdAt.localeCompare(b.createdAt);
    case "updated_at":
      return a.updatedAt.localeCompare(b.updatedAt);
  }
}

function priorityRank(priority: Priority): number {
  switch (priority) {
    case "urgent":
      return 4;
    case "high":
      return 3;
    case "normal":
      return 2;
    case "low":
      return 1;
  }
}

export function countIssues(filters: IssueListFilters = {}): number {
  return issues.filter((issue) => matchesIssueFilters(issue, filters)).length;
}

function matchesIssueFilters(issue: Issue, filters: IssueListFilters): boolean {
  const matchesProject = matchesProjectFilter(issue, filters);
  const matchesState =
    filters.states === undefined ||
    filters.states.length === 0 ||
    filters.states.includes(issue.status);
  const matchesPriority =
    filters.priorities === undefined ||
    filters.priorities.length === 0 ||
    filters.priorities.includes(issue.priority);
  const matchesAssignee = filters.assignee === undefined || issue.assignee === filters.assignee;
  const matchesSearch = matchesIssueSearch(issue, filters.search);

  return matchesProject && matchesState && matchesPriority && matchesAssignee && matchesSearch;
}

function matchesProjectFilter(issue: Issue, filters: IssueListFilters): boolean {
  if (filters.projectID !== undefined) {
    return issue.projectId === filters.projectID;
  }
  return filters.projectIDs === undefined ||
    filters.projectIDs.length === 0 ||
    filters.projectIDs.includes(issue.projectId);
}

function matchesIssueSearch(issue: Issue, search: string | undefined): boolean {
  const query = search?.trim() ?? "";
  if (query === "") {
    return true;
  }
  const parsedID = Number.parseInt(query, 10);
  const matchesID = /^\d+$/.test(query) && Number.isSafeInteger(parsedID) && parsedID > 0 && issue.id === parsedID;
  return matchesID || issue.title.toLocaleLowerCase().includes(query.toLocaleLowerCase());
}

export function getIssue(id: number): Issue | null {
  return clone(issues.find((issue) => issue.id === id) ?? null);
}

export function createIssue(input: CreateIssueInput): Issue | null {
  const project = projects.find((candidate) => candidate.id === input.projectId);
  if (!project || input.title.trim() === "") {
    return null;
  }
  const dependencyIDs = input.dependency_ids ?? [];
  const uniqueDependencyIDs = new Set(dependencyIDs);
  if (
    uniqueDependencyIDs.size !== dependencyIDs.length ||
    dependencyIDs.includes(nextIssueID) ||
    dependencyIDs.some((id) => !issues.some((issue) => issue.id === id))
  ) {
    return null;
  }

  const now = new Date().toISOString();
  const issue: Issue = {
    id: nextIssueID,
    projectId: project.id,
    projectKey: project.key,
    title: input.title,
    description: input.description ?? "",
    status: input.status ?? "backlog",
    priority: input.priority ?? "normal",
    assignee: input.assignee ?? "",
    dependency_ids: [...dependencyIDs].sort((left, right) => left - right),
    createdAt: now,
    updatedAt: now,
  };

  nextIssueID += 1;
  issues = [issue, ...issues];
  return clone(issue);
}

export function updateIssue(id: number, input: UpdateIssueInput): Issue | null {
  const current = issues.find((issue) => issue.id === id);
  if (!current) {
    return null;
  }

  const updated: Issue = {
    ...current,
    ...definedFields(input),
    updatedAt: new Date().toISOString(),
  };

  issues = issues.map((issue) => (issue.id === id ? updated : issue));
  return clone(updated);
}

export function listIssueStates(ids: number[]): IssueState[] {
  return issues
    .filter((issue) => ids.includes(issue.id))
    .map((issue) => ({ id: issue.id, status: issue.status }));
}

export function listComments(issueID: number, cursor = 0, limit = 20): CommentListResponse | null {
  if (!issues.some((issue) => issue.id === issueID)) {
    return null;
  }

  const filtered = comments
    .filter((comment) => comment.issueId === issueID && comment.id > cursor)
    .sort((left, right) => left.id - right.id);
  const data = filtered.slice(0, limit);
  const remaining = filtered.slice(limit);

  return {
    data: clone(data),
    meta: {
      cursor,
      limit,
      nextCursor: remaining[0]?.id ?? null,
    },
  };
}

export function createComment(
  issueID: number,
  input: { author: string; body: string; type?: Comment["type"] },
): Comment | null {
  if (!issues.some((issue) => issue.id === issueID) || input.author.trim() === "" || input.body.trim() === "") {
    return null;
  }

  const comment: Comment = {
    id: nextCommentID,
    issueId: issueID,
    author: input.author,
    type: input.type ?? "general",
    body: input.body,
    createdAt: new Date().toISOString(),
  };

  nextCommentID += 1;
  comments = [...comments, comment];
  return clone(comment);
}

export function buildSummary(): Summary {
  return {
    columns: summaryColumnFixtures.map((column) => ({
      ...column,
      issues: listIssues({ states: [column.status] }).map((issue) => ({
        ...issue,
        queueStatus: queueStatusForIssue(issue),
        stats: {
          commentCount: comments.filter((comment) => comment.issueId === issue.id).length,
        },
      })),
    })),
    generatedAt: new Date().toISOString(),
  };
}

function queueStatusForIssue(issue: Issue): QueueStatus {
  if (issue.status === "backlog") {
    return "backlog";
  }
  if (issue.status === "ready") {
    return hasActiveDependency(issue) ? "pending" : "queued";
  }
  if (issue.status === "in_progress") {
    return "processing";
  }
  if (issue.status === "done") {
    return "completed";
  }
  return "inactive";
}

function hasActiveDependency(issue: Issue): boolean {
  return issue.dependency_ids.some((dependencyID) => {
    const dependency = issues.find((candidate) => candidate.id === dependencyID);
    return dependency !== undefined && !isSatisfiedDependencyStatus(dependency.status);
  });
}

function isSatisfiedDependencyStatus(status: IssueStatus): boolean {
  return (
    status === "done" ||
    status === "cancelled" ||
    status === "duplicate"
  );
}

export function buildOrchestratorIssueRuntime(issueID: number): IssueRuntimeResponse | null {
  const issue = issues.find((candidate) => candidate.id === issueID);
  if (!issue) {
    return null;
  }

  const runs = buildIssueRuns(issue);
  return {
    issue_identifier: `issue-${issue.id}`,
    issue_id: String(issue.id),
    status: runs[0]?.status ?? "queued",
    workspace: {
      path: `/tmp/tasq/workspaces/issue-${issue.id}`,
    },
    attempts: {
      restart_count: Math.max(runs.length - 1, 0),
      current_retry_attempt: runs[0]?.attempt ?? 0,
    },
    runs,
    running: null,
    retry: null,
    logs: {
      codex_session_logs: [],
    },
    recent_events: [
      {
        at: issue.updatedAt,
        event: runs[0]?.status ?? "queued",
        message: `mock run ${runs[0]?.run_id ?? "pending"}`,
      },
    ],
    last_error: null,
    tracked: {},
  };
}

export function getOrchestratorConversation(
  issueID: number,
  runID: string,
): ConversationResponse | null {
  const issue = issues.find((candidate) => candidate.id === issueID);
  if (!issue || !buildIssueRuns(issue).some((run) => run.run_id === runID)) {
    return null;
  }

  if (issue.id === 1) {
    return buildIssueOneConversation(issue, runID);
  }

  return {
    issue_identifier: `issue-${issue.id}`,
    issue_id: String(issue.id),
    run_id: runID,
    events: [
      {
        at: issue.createdAt,
        event: "running",
        message: "mock runner started",
        payload_json: "",
      },
      {
        at: issue.updatedAt,
        event: "turn_completed",
        message: "turn_id=mock-turn-1",
        payload_json: JSON.stringify({
          aggregatedOutput: `## ${issue.title}\n\nMock conversation output for ${runID}.`,
        }),
      },
      {
        at: issue.updatedAt,
        event: "succeeded",
        message: "mock runner completed",
        payload_json: "",
      },
    ],
  };
}

function buildIssueOneConversation(issue: Issue, runID: string): ConversationResponse {
  return {
    issue_identifier: `issue-${issue.id}`,
    issue_id: String(issue.id),
    run_id: runID,
    events: [
      {
        at: "2026-06-01T08:00:00.000Z",
        event: "running",
        message: "mock runner started with a fresh workspace",
        payload_json: "",
      },
      {
        at: "2026-06-01T08:00:18.000Z",
        event: "item/completed",
        message: "read repository context",
        payload_json: JSON.stringify({
          item: {
            type: "reasoning",
            text: [
              "I am checking the mock transport boundary before changing the UI.",
              "The route should continue to call the generated clients, while MSW returns stable fixtures for issue detail, comments, runtime, and conversation history.",
              "Important edge cases to preserve: missing issue IDs should still return not_found errors, and non-issue-<id> orchestrator identifiers should stay invalid.",
            ].join("\n\n"),
            exitCode: 0,
          },
        }),
      },
      {
        at: "2026-06-01T08:01:04.000Z",
        event: "item/completed",
        message: "listed relevant files",
        payload_json: JSON.stringify({
          item: {
            type: "commandExecution",
            command: "rg -n \"getOrchestratorConversation|ConversationEvents|IssueDetailPage\" cmd/web/frontend/src",
            aggregatedOutput: [
              "cmd/web/frontend/src/mocks/state.ts:367:export function getOrchestratorConversation(",
              "cmd/web/frontend/src/features/issues/components/conversation-page/index.tsx:100:export function ConversationEvents",
              "cmd/web/frontend/src/features/issues/components/issue-detail-page/index.tsx:327:<ConversationTab",
              "",
              "The search confirms that the mock conversation payload is centralized and can be expanded without changing route-level components.",
            ].join("\n"),
            exitCode: 0,
          },
        }),
      },
      {
        at: "2026-06-01T08:02:21.000Z",
        event: "turn_completed",
        message: "turn_id=mock-turn-1",
        payload_json: JSON.stringify({
          aggregatedOutput: `## Investigation summary

Issue **#${issue.id}** uses the mock state module as the source of truth for conversation history. The current fixture should exercise a dense timeline, mixed event types, long Markdown output, command snippets, and approval request rendering.

### Findings

| Area | Current behavior | Refine target |
| --- | --- | --- |
| Runtime lookup | Uses \`issue-${issue.id}\` and \`${runID}\` | Keep identifiers stable |
| Timeline density | A few events | Many realistic turns |
| Markdown output | Short paragraph | Headings, lists, tables, code, links |
| Approval event | Rarely visible | Include one request with a clear reason |

### Proposed approach

1. Keep generic conversations short for other issues.
2. Make issue 1 a rich manual-review fixture.
3. Preserve the generated \`ConversationResponse\` contract.
4. Avoid fixture-only fields that the UI does not understand.

> The mock should read like an actual agent session, not like lorem ipsum. This makes spacing, filtering, and payload parsing problems easier to notice.
`,
        }),
      },
      {
        at: "2026-06-01T08:03:09.000Z",
        event: "item/completed",
        message: "opened implementation file",
        payload_json: JSON.stringify({
          item: {
            type: "fileRead",
            text: [
              "Read cmd/web/frontend/src/mocks/state.ts.",
              "The function already validates issue existence and run ID before returning a response.",
              "The safest extension point is after validation and before the generic fixture response.",
            ].join("\n"),
            exitCode: 0,
          },
        }),
      },
      {
        at: "2026-06-01T08:04:30.000Z",
        event: "item/commandExecution/requestApproval",
        message: "approval requested",
        payload_json: JSON.stringify({
          reason: "Need to run the frontend typechecker after changing generated-shape mock data.",
          command: "npm run typecheck",
          availableDecisions: ["approve", "deny"],
          proposedExecpolicyAmendment: {
            sandbox_permissions: "workspace-write",
          },
        }),
      },
      {
        at: "2026-06-01T08:05:12.000Z",
        event: "item/completed",
        message: "edited mock fixture",
        payload_json: JSON.stringify({
          item: {
            type: "patch",
            aggregatedOutput: `### Patch notes

- Added an issue-specific conversation builder for issue 1.
- Covered \`running\`, \`item/completed\`, \`turn_completed\`, approval, failure, and final success events.
- Kept every event inside the generated payload shape.

\`\`\`ts
function buildIssueOneConversation(issue: Issue, runID: string): ConversationResponse {
  return {
    issue_identifier: \`issue-\${issue.id}\`,
    issue_id: String(issue.id),
    run_id: runID,
    events: []
  };
}
\`\`\`
`,
            text: "The mock remains deterministic so visual review screenshots are stable.",
            exitCode: 0,
          },
        }),
      },
      {
        at: "2026-06-01T08:06:37.000Z",
        event: "item/completed",
        message: "ran focused verification",
        payload_json: JSON.stringify({
          item: {
            type: "commandExecution",
            command: "npm run typecheck",
            aggregatedOutput: [
              "> tasq-web@0.1.0 typecheck",
              "> tsc --noEmit",
              "",
              "No TypeScript errors were reported.",
            ].join("\n"),
            exitCode: 0,
          },
        }),
      },
      {
        at: "2026-06-01T08:06:49.000Z",
        event: "thread/tokenUsage/updated",
        message: "token usage updated",
        payload_json: JSON.stringify({
          threadId: "019ec3e5-376c-7521-a5ee-f2af2cd0f57c",
          turnId: "019ec3e5-37ca-7632-a978-42e14ed98b66",
          tokenUsage: {
            total: {
              totalTokens: 8712831,
              inputTokens: 8684367,
              cachedInputTokens: 8415104,
              outputTokens: 28464,
              reasoningOutputTokens: 5186,
            },
            last: {
              totalTokens: 146759,
              inputTokens: 146356,
              cachedInputTokens: 145280,
              outputTokens: 403,
              reasoningOutputTokens: 203,
            },
            modelContextWindow: 258400,
          },
        }),
      },
      {
        at: "2026-06-01T08:06:54.000Z",
        event: "account/rateLimits/updated",
        message: "rate limits updated",
        payload_json: JSON.stringify({
          rateLimits: {
            limitId: "codex",
            limitName: null,
            primary: {
              usedPercent: 13,
              windowDurationMins: 300,
              resetsAt: 1781415835,
            },
            secondary: {
              usedPercent: 13,
              windowDurationMins: 10080,
              resetsAt: 1781572462,
            },
            credits: null,
            individualLimit: null,
            planType: "pro",
            rateLimitReachedType: null,
          },
        }),
      },
      {
        at: "2026-06-01T08:07:18.000Z",
        event: "item/completed",
        message: "captured a failing exploratory command",
        payload_json: JSON.stringify({
          item: {
            type: "commandExecution",
            command: "npm test -- conversation-history-fixture",
            aggregatedOutput: [
              "No test files matched pattern conversation-history-fixture.",
              "",
              "This is expected because the change is mock data only; the existing conversation component tests cover payload rendering behavior.",
            ].join("\n"),
            text: "Kept this failed command in the fixture to verify non-zero exit code rendering.",
            exitCode: 1,
          },
        }),
      },
      {
        at: "2026-06-01T08:08:44.000Z",
        event: "turn_completed",
        message: "turn_id=mock-turn-2",
        payload_json: JSON.stringify({
          response: {
            aggregatedOutput: `## Review handoff

The fixture now gives manual reviewers a long conversation with realistic variation:

- status events for start and completion;
- reasoning and file-read style text events;
- command execution with success and failure exit codes;
- nested approval request data;
- Markdown with tables, quotes, task lists, links, and fenced code blocks;
- enough vertical length to test scrolling and timeline rhythm.

### Manual route

Open [issue 1 conversation](/issues/1?tab=conversation) while mock mode is enabled. The selected run should be \`${runID}\`.

### Checklist

- [x] Long output wraps inside the timeline.
- [x] Command labels stay compact.
- [x] Approval request uses the highlighted approval treatment.
- [ ] Confirm mobile viewport spacing.
- [ ] Confirm message-type filter remains usable with many event types.
`,
          },
        }),
      },
      {
        at: "2026-06-01T08:09:33.000Z",
        event: "failed",
        message: "mock retry marker: first visual pass found crowded spacing",
        payload_json: "",
      },
      {
        at: "2026-06-01T08:10:55.000Z",
        event: "item/completed",
        message: "documented follow-up",
        payload_json: JSON.stringify({
          item: {
            type: "text",
            text: [
              "Follow-up notes:",
              "",
              "- If the approval block feels too dominant, adjust the component style instead of shrinking this fixture.",
              "- If code blocks overflow, inspect the Markdown component's pre/code rules.",
              "- If filtering becomes noisy, consider sorting message-type options by first appearance.",
            ].join("\n"),
            exitCode: 0,
          },
        }),
      },
      {
        at: "2026-06-01T08:12:00.000Z",
        event: "succeeded",
        message: "mock runner completed after conversation fixture refinement",
        payload_json: "",
      },
    ],
  };
}

function buildIssueRuns(issue: Issue): IssueRuntimeResponse["runs"] {
  const status: RunState = issue.status === "done" ? "succeeded" : "running";
  return [
    {
      run_id: `run-${issue.id}-latest`,
      thread_id: `thread-${issue.id}-latest`,
      status,
      attempt: 1,
      created_at: issue.createdAt,
      updated_at: issue.updatedAt,
    },
  ];
}

function definedFields<T extends object>(input: T): Partial<T> {
  return Object.fromEntries(
    Object.entries(input).filter(([, value]) => value !== undefined),
  ) as Partial<T>;
}

function clone<T>(value: T): T {
  return structuredClone(value);
}

function nextNumericID(items: Array<{ id: number }>): number {
  return Math.max(0, ...items.map((item) => item.id)) + 1;
}
