export const issueStatuses = [
  "backlog",
  "ready",
  "in_progress",
  "review",
  "blocked",
  "failed",
  "done",
] as const;

export type IssueStatus = (typeof issueStatuses)[number];

export const priorities = ["low", "normal", "high", "urgent"] as const;

export type Priority = (typeof priorities)[number];

export const runStatuses = [
  "queued",
  "starting",
  "running",
  "waiting_for_input",
  "succeeded",
  "failed",
  "cancelled",
] as const;

export type RunStatus = (typeof runStatuses)[number];

export const workspaceStatuses = ["active", "inactive", "archived"] as const;

export type WorkspaceStatus = (typeof workspaceStatuses)[number];

export type Project = {
  id: number;
  key: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
};

export type Workspace = {
  id: number;
  projectId: number;
  name: string;
  path: string;
  status: WorkspaceStatus;
  createdAt: string;
  updatedAt: string;
};

export type Issue = {
  id: number;
  title: string;
  description: string;
  status: IssueStatus;
  priority: Priority;
  assignee: string;
  createdAt: string;
  updatedAt: string;
};

export type RunSnapshot = {
  issueId: number;
  workItemId: number;
  runId: string;
  status: RunStatus;
  workspace: string;
  attempt: number;
  error: string;
  orchestratorId: string;
  updatedAt: string;
};

export type IssueSummary = Issue & {
  run?: RunSnapshot;
};

export type Column = {
  status: IssueStatus;
  title: string;
  issues: IssueSummary[];
};

export type Summary = {
  columns: Column[];
  runs: RunSnapshot[];
  generatedAt: string;
};
