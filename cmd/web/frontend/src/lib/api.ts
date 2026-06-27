import {
  getApiV1Issues,
  getApiV1IssuesId,
  getApiV1IssuesIssueIdComments,
  getApiV1Projects,
  getApiV1ProjectsIdWorkflow,
  getApiV1Summary,
  patchApiV1IssuesId,
  postApiV1Issues,
  postApiV1Projects,
  type CommentListResponse,
  type CreateIssueInput,
  type CreateProjectInput,
  type ErrorResponse,
  type GetApiV1IssuesParams,
  type Issue,
  type IssueListResponse,
  type IssueStatus,
  type Project,
  type ProjectWorkflow,
  type Summary,
} from "@/lib/generated/issue-tracker";
import {
  getApiV1State,
  getApiV1IssueIdentifier,
  getApiV1IssueIdentifierRunsRunIdConversations,
  type ConversationResponse,
  type ErrorResponse as OrchestratorErrorResponse,
  type IssueRuntimeResponse,
  type StateResponse,
} from "@/lib/generated/orchestrator";
import { i18n } from "@/lib/i18n";
import { toast } from "@/lib/toast";

type ApiResponse<T> = {
  data: ApiEnvelope<T> | ErrorResponse;
  status: number;
};

type ApiEnvelope<T, M = unknown> = {
  data: T;
  meta: M;
};

type ApiRequestOptions = {
  silent?: boolean;
};

type ApiErrorInfo = {
  code: string;
  message: string;
};

const noStore: RequestInit = {
  cache: "no-store",
};

export class ApiRequestError extends Error {
  readonly code: string;

  constructor(error: ApiErrorInfo) {
    super(error.message);
    this.name = "ApiRequestError";
    this.code = error.code;
  }
}

export function fetchSummary(options?: ApiRequestOptions): Promise<Summary> {
  return unwrapResponse(getApiV1Summary(noStore), options);
}

export function createIssue(input: CreateIssueInput, options?: ApiRequestOptions): Promise<Issue> {
  return unwrapResponse(postApiV1Issues(input, noStore), options);
}

export function fetchIssue(id: number, options?: ApiRequestOptions): Promise<Issue> {
  return unwrapResponse(getApiV1IssuesId(id, noStore), options);
}

export function fetchIssues(
  params: GetApiV1IssuesParams,
  options?: ApiRequestOptions,
): Promise<IssueListResponse> {
  return unwrapEnvelope<IssueListResponse>(getApiV1Issues(params, noStore), options);
}

export function fetchComments(
  issueID: number,
  cursor?: number,
  limit = 20,
  options?: ApiRequestOptions,
): Promise<CommentListResponse> {
  return unwrapEnvelope<CommentListResponse>(
    getApiV1IssuesIssueIdComments(issueID, { cursor, limit }, noStore),
    options,
  );
}

export function fetchProjects(options?: ApiRequestOptions): Promise<Project[]> {
  return unwrapResponse(getApiV1Projects(noStore), options);
}

export function fetchProjectWorkflow(
  projectID: number,
  options?: ApiRequestOptions,
): Promise<ProjectWorkflow> {
  return unwrapResponse(getApiV1ProjectsIdWorkflow(projectID, noStore), options);
}

export function createProject(input: CreateProjectInput, options?: ApiRequestOptions): Promise<Project> {
  return unwrapResponse(postApiV1Projects(input, noStore), options);
}

export function updateIssueStatus(
  id: number,
  status: IssueStatus,
  options?: ApiRequestOptions,
): Promise<Issue> {
  return unwrapResponse(patchApiV1IssuesId(id, { status }, noStore), options);
}

export function fetchOrchestratorState(options?: ApiRequestOptions): Promise<StateResponse> {
  return unwrapOrchestratorResponse(getApiV1State(noStore), options);
}

export function fetchOrchestratorIssueRuntime(
  issueID: number,
  options?: ApiRequestOptions,
): Promise<IssueRuntimeResponse> {
  return unwrapOrchestratorResponse(getApiV1IssueIdentifier(`issue-${issueID}`, noStore), options);
}

export function fetchOrchestratorConversation(
  issueID: number,
  runID: string,
  options?: ApiRequestOptions,
): Promise<ConversationResponse> {
  return unwrapOrchestratorResponse(
    getApiV1IssueIdentifierRunsRunIdConversations(
      `issue-${issueID}`,
      encodeURIComponent(runID),
      noStore,
    ),
    options,
  );
}

async function unwrapResponse<T>(
  response: Promise<ApiResponse<T>>,
  options?: ApiRequestOptions,
): Promise<T> {
  const resolved = await response;
  if (resolved.status >= 400) {
    const payload = resolved.data as ErrorResponse;
    throw notifyAndCreateError(payload.error, options);
  }
  return (resolved.data as ApiEnvelope<T>).data;
}

async function unwrapEnvelope<T extends ApiEnvelope<unknown>>(
  response: Promise<{ data: T | ErrorResponse; status: number }>,
  options?: ApiRequestOptions,
): Promise<T> {
  const resolved = await response;
  if (resolved.status >= 400) {
    const payload = resolved.data as ErrorResponse;
    throw notifyAndCreateError(payload.error, options);
  }
  return resolved.data as T;
}

async function unwrapOrchestratorResponse<T>(
  response: Promise<{ data: T | OrchestratorErrorResponse; status: number }>,
  options?: ApiRequestOptions,
): Promise<T> {
  const resolved = await response;
  if (resolved.status >= 400) {
    const payload = resolved.data as OrchestratorErrorResponse;
    const code = payload.error?.code ?? "orchestrator_error";
    const message = payload.error?.message ?? "orchestrator request failed";
    throw notifyAndCreateError({ code, message }, options);
  }
  return resolved.data as T;
}

function notifyAndCreateError(error: ApiErrorInfo, options?: ApiRequestOptions): ApiRequestError {
  const requestError = new ApiRequestError(error);
  if (!options?.silent) {
    toast.error({ message: toastErrorMessage(error) });
  }
  return requestError;
}

function toastErrorMessage(error: ApiErrorInfo): string {
  const key = `toast.error.${error.code}`;
  return i18n.exists(key) ? i18n.t(key) : error.message;
}
