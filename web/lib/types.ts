import {
  IssueStatus as IssueStatusValues,
  Priority as PriorityValues,
  RunStatus as RunStatusValues,
  WorkspaceStatus as WorkspaceStatusValues,
  type Column,
  type Issue,
  type IssueStatus,
  type IssueSummary,
  type Priority,
  type Project,
  type RunSnapshot,
  type RunStatus,
  type Summary,
  type Workspace,
  type WorkspaceStatus,
} from "@/lib/generated/issue-tracker";

export const issueStatuses = Object.values(IssueStatusValues);
export const priorities = Object.values(PriorityValues);
export const runStatuses = Object.values(RunStatusValues);
export const workspaceStatuses = Object.values(WorkspaceStatusValues);

export type {
  Column,
  Issue,
  IssueStatus,
  IssueSummary,
  Priority,
  Project,
  RunSnapshot,
  RunStatus,
  Summary,
  Workspace,
  WorkspaceStatus,
};
