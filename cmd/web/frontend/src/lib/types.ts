import {
  IssueStatus as IssueStatusValues,
  Priority as PriorityValues,
  type Column,
  type Comment,
  type CommentListMeta,
  type CommentListResponse,
  type CreateProjectInput,
  type CommentType,
  type CreateIssueInput,
  type Issue,
  type IssueStatus,
  type IssueSummary,
  type Priority,
  type Project,
  type Summary,
} from "@/lib/generated/issue-tracker";

export const issueStatuses = Object.values(IssueStatusValues);
export const priorities = Object.values(PriorityValues);

export type {
  Column,
  Comment,
  CommentListMeta,
  CommentListResponse,
  CreateProjectInput,
  CommentType,
  CreateIssueInput,
  Issue,
  IssueStatus,
  IssueSummary,
  Priority,
  Project,
  Summary,
};
