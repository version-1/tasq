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

export type OrchestratorRunStatus =
  | "queued"
  | "starting"
  | "running"
  | "waiting_for_input"
  | "succeeded"
  | "failed"
  | "cancelled";

export type OrchestratorConversationEventType = "turn_completed" | OrchestratorRunStatus;

export type OrchestratorTokenSummary = {
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
};

export type OrchestratorRunSummary = {
  issue_id: string;
  issue_identifier: string;
  state: OrchestratorRunStatus;
  session_id: string;
  turn_count: number;
  last_event: string;
  last_message: string;
  started_at: string;
  last_event_at: string | null;
  tokens: OrchestratorTokenSummary;
};

export type OrchestratorState = {
  generated_at: string;
  counts: {
    running: number;
    retrying: number;
  };
  running: OrchestratorRunSummary[];
  retrying: unknown[];
  codex_totals: OrchestratorTokenSummary | null;
  rate_limits: Record<string, unknown> | null;
};

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
  event: OrchestratorConversationEventType;
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
