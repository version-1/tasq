export const taskStatuses = [
  "backlog",
  "ready",
  "running",
  "review",
  "blocked",
  "failed",
  "done",
] as const;

export type TaskStatus = (typeof taskStatuses)[number];

export const priorities = ["low", "normal", "high", "urgent"] as const;

export type Priority = (typeof priorities)[number];

export const agentStatuses = [
  "idle",
  "queued",
  "running",
  "waiting_for_input",
  "succeeded",
  "failed",
] as const;

export type AgentStatus = (typeof agentStatuses)[number];

export type Task = {
  id: number;
  title: string;
  description: string;
  status: TaskStatus;
  priority: Priority;
  agentStatus: AgentStatus;
  assignee: string;
  source: string;
  sourceId: string;
  workspace: string;
  attempts: number;
  lastError: string;
  createdAt: string;
  updatedAt: string;
};

export type Column = {
  status: TaskStatus;
  title: string;
  tasks: Task[];
};

export type Settings = {
  pollIntervalSeconds: number;
  maxConcurrentRuns: number;
  workspaceRoot: string;
  trackerProvider: string;
  agentCommand: string;
};

export type Summary = {
  columns: Column[];
  agents: Task[];
  settings: Settings;
  generatedAt: string;
};
