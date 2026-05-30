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
  data: T | ErrorResponse;
  status: number;
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
    throw new Error(payload.error ?? String(resolved.status));
  }
  return resolved.data as T;
}
