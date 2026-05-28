import type { Issue, IssueStatus, Project, Summary, Workspace } from "@/lib/types";

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

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export function fetchSummary(): Promise<Summary> {
  return request<Summary>("/api/v1/summary");
}

export type CreateProjectInput = {
  key: string;
  name: string;
  description?: string;
};

export type UpdateProjectInput = Partial<CreateProjectInput>;

export function fetchProjects(): Promise<Project[]> {
  return request<Project[]>("/api/v1/projects");
}

export function createProject(input: CreateProjectInput): Promise<Project> {
  return request<Project>("/api/v1/projects", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateProject(id: number, input: UpdateProjectInput): Promise<Project> {
  return request<Project>(`/api/v1/projects/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deleteProject(id: number): Promise<void> {
  return request<void>(`/api/v1/projects/${id}`, {
    method: "DELETE",
  });
}

export type CreateWorkspaceInput = {
  projectId: number;
  name: string;
  path: string;
  status?: Workspace["status"];
};

export type UpdateWorkspaceInput = Partial<CreateWorkspaceInput>;

export function fetchWorkspaces(): Promise<Workspace[]> {
  return request<Workspace[]>("/api/v1/workspaces");
}

export function createWorkspace(input: CreateWorkspaceInput): Promise<Workspace> {
  return request<Workspace>("/api/v1/workspaces", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateWorkspace(id: number, input: UpdateWorkspaceInput): Promise<Workspace> {
  return request<Workspace>(`/api/v1/workspaces/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deleteWorkspace(id: number): Promise<void> {
  return request<void>(`/api/v1/workspaces/${id}`, {
    method: "DELETE",
  });
}

export function updateIssueStatus(id: number, status: IssueStatus): Promise<Issue> {
  return request<Issue>(`/api/v1/issues/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}
