import type {
  Comment,
  CommentListResponse,
  CreateIssueInput,
  Issue,
  IssueState,
  IssueStatus,
  Project,
  Summary,
  UpdateIssueInput,
} from "@/lib/generated/issue-tracker";
import { commentFixtures } from "./fixtures/comments";
import { issueFixtures } from "./fixtures/issues";
import { projectFixtures } from "./fixtures/projects";
import { summaryColumnFixtures } from "./fixtures/summary";

type IssueListFilters = {
  projectID?: number;
  states?: IssueStatus[];
};

let projects = clone(projectFixtures);
let issues = clone(issueFixtures);
let comments = clone(commentFixtures);
let nextIssueID = nextNumericID(issues);
let nextCommentID = nextNumericID(comments);

export function listProjects(): Project[] {
  return clone(projects);
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
      issues: listIssues({ states: [column.status] }),
    })),
    generatedAt: new Date().toISOString(),
  };
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
