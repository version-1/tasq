import {
  IssueStatus as IssueStatusValues,
  Priority as PriorityValues,
  WorkspaceStatus as WorkspaceStatusValues,
  type Column,
  type CreateIssueInput,
  type Issue,
  type IssueStatus,
  type IssueSummary,
  type Priority,
  type Project,
  type Summary,
  type Workspace,
  type WorkspaceStatus,
} from "@/lib/generated/issue-tracker";

export const issueStatuses = Object.values(IssueStatusValues);
export const priorities = Object.values(PriorityValues);
export const workspaceStatuses = Object.values(WorkspaceStatusValues);

export type {
  Column,
  CreateIssueInput,
  Issue,
  IssueStatus,
  IssueSummary,
  Priority,
  Project,
  Summary,
  Workspace,
  WorkspaceStatus,
};
