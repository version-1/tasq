import type { Issue, IssueStatus, Summary } from "@/lib/types";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_ISSUE_TRACKER_URL ?? "";

async function request<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
    cache: "no-store",
  });

  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as
      | { error?: string }
      | null;
    throw new Error(payload?.error ?? `${response.status} ${response.statusText}`);
  }

  return (await response.json()) as T;
}

export function fetchSummary(): Promise<Summary> {
  return request<Summary>("/api/v1/summary");
}

export function updateIssueStatus(id: number, status: IssueStatus): Promise<Issue> {
  return request<Issue>(`/api/v1/issues/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}
