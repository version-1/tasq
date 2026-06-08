import {
  IssueStatus as IssueStatusValues,
  Priority as PriorityValues,
  type Column,
  type Comment,
  type CommentListMeta,
  type CommentListResponse,
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

export type OrchestratorRunStatus = "queued" | "running" | "succeeded" | "failed" | "cancelled";

export type OrchestratorIssueRun = {
  run_id: string;
  status: OrchestratorRunStatus;
  attempt: number;
  created_at: string;
  updated_at: string;
};

export type OrchestratorIssueRuntime = {
  issue_identifier: string;
  issue_id: string;
  status: OrchestratorRunStatus;
  runs: OrchestratorIssueRun[];
};

export type OrchestratorConversationEvent = {
  at: string;
  event: "turn_completed" | OrchestratorRunStatus;
  message: string;
  payload_json: string;
};

export type OrchestratorConversation = {
  issue_identifier: string;
  issue_id: string;
  run_id: string;
  events: OrchestratorConversationEvent[];
};

export type {
  Column,
  Comment,
  CommentListMeta,
  CommentListResponse,
  CommentType,
  CreateIssueInput,
  Issue,
  IssueStatus,
  IssueSummary,
  Priority,
  Project,
  Summary,
};
