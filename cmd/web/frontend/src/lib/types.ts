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
  type ProjectWorkflow,
  type ProjectWorkflowFrontmatter,
  type QueueStatus,
  type Summary,
} from "@/lib/generated/issue-tracker";
import type {
  ConversationEvent,
  ConversationResponse,
  IssueRunSummary,
  IssueRuntimeResponse,
  RunSummary,
  RunState,
  StateResponse,
  TokenSummary,
} from "@/lib/generated/orchestrator";

export const issueStatuses = Object.values(IssueStatusValues);
export const priorities = Object.values(PriorityValues);

export type OrchestratorRunStatus = RunState;
export type OrchestratorTokenSummary = TokenSummary;
export type OrchestratorRunSummary = RunSummary;
export type OrchestratorState = StateResponse;
export type OrchestratorIssueRun = IssueRunSummary;
export type OrchestratorIssueRuntime = IssueRuntimeResponse;
export type OrchestratorConversationEvent = ConversationEvent;
export type OrchestratorConversation = ConversationResponse;

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
  ProjectWorkflow,
  ProjectWorkflowFrontmatter,
  QueueStatus,
  Summary,
};
