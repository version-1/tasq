import type {
  Comment,
  CommentListResponse,
  CreateIssueInput,
  CreateProjectInput,
  Issue,
  IssueState,
  IssueStatus,
  Project,
  ProjectWorkflow,
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
  projectID?: number;
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
  return clone(
    issues.filter((issue) => {
      const matchesProject = filters.projectID === undefined || issue.projectId === filters.projectID;
      const matchesState =
        filters.states === undefined ||
        filters.states.length === 0 ||
        filters.states.includes(issue.status);

      return matchesProject && matchesState;
    }),
  );
}

export function getIssue(id: number): Issue | null {
  return clone(issues.find((issue) => issue.id === id) ?? null);
}

export function createIssue(input: CreateIssueInput): Issue | null {
  const project = projects.find((candidate) => candidate.id === input.projectId);
  if (!project || input.title.trim() === "") {
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
        stats: {
          commentCount: comments.filter((comment) => comment.issueId === issue.id).length,
        },
      })),
    })),
    generatedAt: new Date().toISOString(),
  };
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
