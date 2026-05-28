import {
  IssueStatus as IssueStatusValues,
  Priority as PriorityValues,
  RunStatus as RunStatusValues,
  type Column,
  type Issue,
  type IssueStatus,
  type IssueSummary,
  type Priority,
  type RunSnapshot,
  type RunStatus,
  type Summary,
} from "@/lib/generated/issue-tracker";

export const issueStatuses = Object.values(IssueStatusValues);
export const priorities = Object.values(PriorityValues);
export const runStatuses = Object.values(RunStatusValues);

export type {
  Column,
  Issue,
  IssueStatus,
  IssueSummary,
  Priority,
  RunSnapshot,
  RunStatus,
  Summary,
};
