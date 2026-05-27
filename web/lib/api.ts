import type { Settings, Summary, Task, TaskStatus } from "@/lib/types";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_ORCHESTRATOR_URL ?? "http://localhost:8080";

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

export function updateTaskStatus(id: number, status: TaskStatus): Promise<Task> {
  return request<Task>(`/api/v1/tasks/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

export function updateSettings(settings: Settings): Promise<Settings> {
  return request<Settings>("/api/v1/settings", {
    method: "PUT",
    body: JSON.stringify(settings),
  });
}
