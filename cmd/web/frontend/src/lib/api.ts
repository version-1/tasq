import {
  getApiV1IssuesId,
  getApiV1IssuesIssueIdComments,
  getApiV1Projects,
  getApiV1Summary,
  patchApiV1IssuesId,
  postApiV1Issues,
  postApiV1Projects,
  type CommentListResponse,
  type CreateIssueInput,
  type CreateProjectInput,
  type ErrorResponse,
  type Issue,
  type IssueStatus,
  type Project,
  type Summary,
} from "@/lib/generated/issue-tracker";
import type {
  OrchestratorConversation as OrchestratorConversationPayload,
  OrchestratorIssueRuntime as OrchestratorIssueRuntimePayload,
} from "@/lib/types";

type ApiResponse<T> = {
  data: ApiEnvelope<T> | ErrorResponse;
  status: number;
};

type ApiEnvelope<T, M = unknown> = {
  data: T;
  meta: M;
};

const noStore: RequestInit = {
  cache: "no-store",
};

export function fetchSummary(): Promise<Summary> {
  return unwrapResponse(getApiV1Summary(noStore));
}

export function createIssue(input: CreateIssueInput): Promise<Issue> {
  return unwrapResponse(postApiV1Issues(input, noStore));
}

export function fetchIssue(id: number): Promise<Issue> {
  return unwrapResponse(getApiV1IssuesId(id, noStore));
}

export function fetchComments(
  issueID: number,
  cursor?: number,
  limit = 20,
): Promise<CommentListResponse> {
  return unwrapEnvelope<CommentListResponse>(
    getApiV1IssuesIssueIdComments(issueID, { cursor, limit }, noStore),
  );
}

export function fetchProjects(): Promise<Project[]> {
  return unwrapResponse(getApiV1Projects(noStore));
}

export function createProject(input: CreateProjectInput): Promise<Project> {
  return unwrapResponse(postApiV1Projects(input, noStore));
}

export function updateIssueStatus(id: number, status: IssueStatus): Promise<Issue> {
  return unwrapResponse(patchApiV1IssuesId(id, { status }, noStore));
}

export function fetchOrchestratorIssueRuntime(
  issueID: number,
): Promise<OrchestratorIssueRuntimePayload> {
  return fetchOrchestrator<OrchestratorIssueRuntimePayload>(`/api/v1/issue-${issueID}`);
}

export function fetchOrchestratorConversation(
  issueID: number,
  runID: string,
): Promise<OrchestratorConversationPayload> {
  return fetchOrchestrator<OrchestratorConversationPayload>(
    `/api/v1/issue-${issueID}/runs/${encodeURIComponent(runID)}/conversations`,
  );
}

async function unwrapResponse<T>(response: Promise<ApiResponse<T>>): Promise<T> {
  const resolved = await response;
  if (resolved.status >= 400) {
    const payload = resolved.data as ErrorResponse;
    throw new Error(`${payload.error.code}: ${payload.error.message}`);
  }
  return (resolved.data as ApiEnvelope<T>).data;
}

async function unwrapEnvelope<T extends ApiEnvelope<unknown>>(
  response: Promise<{ data: T | ErrorResponse; status: number }>,
): Promise<T> {
  const resolved = await response;
  if (resolved.status >= 400) {
    const payload = resolved.data as ErrorResponse;
    throw new Error(`${payload.error.code}: ${payload.error.message}`);
  }
  return resolved.data as T;
}

async function fetchOrchestrator<T>(path: string): Promise<T> {
  const response = await fetch(`/orchestrator${path}`, noStore);
  const payload = await response.json();
  if (!response.ok) {
    const error = payload as { error?: { code?: string; message?: string } };
    const code = error.error?.code ?? "orchestrator_error";
    const message = error.error?.message ?? "orchestrator request failed";
    throw new Error(`${code}: ${message}`);
  }
  return payload as T;
}
