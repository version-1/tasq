import {
  getApiV1Projects,
  getApiV1Summary,
  patchApiV1IssuesId,
  type ErrorResponse,
  type Issue,
  type IssueStatus,
  type Project,
  type Summary,
} from "@/lib/generated/issue-tracker";

type ApiResponse<T> = {
  data: ApiEnvelope<T> | ErrorResponse;
  status: number;
};

type ApiEnvelope<T> = {
  data: T;
  meta: Record<string, unknown>;
};

const noStore: RequestInit = {
  cache: "no-store",
};

export function fetchSummary(): Promise<Summary> {
  return unwrapResponse(getApiV1Summary(noStore));
}

export function fetchProjects(): Promise<Project[]> {
  return unwrapResponse(getApiV1Projects(noStore));
}

export function updateIssueStatus(id: number, status: IssueStatus): Promise<Issue> {
  return unwrapResponse(patchApiV1IssuesId(id, { status }, noStore));
}

async function unwrapResponse<T>(response: Promise<ApiResponse<T>>): Promise<T> {
  const resolved = await response;
  if (resolved.status >= 400) {
    const payload = resolved.data as ErrorResponse;
    throw new Error(`${payload.error.code}: ${payload.error.message}`);
  }
  return (resolved.data as ApiEnvelope<T>).data;
}
