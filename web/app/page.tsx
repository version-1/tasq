"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { fetchSummary, updateSettings, updateTaskStatus } from "@/lib/api";
import type { Settings, Summary, Task, TaskStatus } from "@/lib/types";
import { taskStatuses } from "@/lib/types";

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; summary: Summary }
  | { kind: "error"; message: string };

const tabs = ["kanban", "agents", "detail", "settings"] as const;
type Tab = (typeof tabs)[number];

const statusLabels: Record<TaskStatus, string> = {
  backlog: "Backlog",
  ready: "Ready",
  running: "Running",
  review: "Review",
  blocked: "Blocked",
  failed: "Failed",
  done: "Done",
};

export default function Home() {
  const [loadState, setLoadState] = useState<LoadState>({ kind: "loading" });
  const [activeTab, setActiveTab] = useState<Tab>("kanban");
  const [selectedTaskID, setSelectedTaskID] = useState<number | null>(null);
  const [notice, setNotice] = useState("");

  const load = useCallback(async () => {
    try {
      const summary = await fetchSummary();
      setLoadState({ kind: "ready", summary });
      setSelectedTaskID((current) => current ?? firstTaskID(summary));
    } catch (error) {
      setLoadState({
        kind: "error",
        message: error instanceof Error ? error.message : "failed to load summary",
      });
    }
  }, []);

  useEffect(() => {
    void load();
    const id = window.setInterval(() => {
      void load();
    }, 3000);
    return () => window.clearInterval(id);
  }, [load]);

  const summary = loadState.kind === "ready" ? loadState.summary : null;
  const tasks = useMemo(() => {
    if (!summary) return [];
    return summary.columns.flatMap((column) => column.tasks);
  }, [summary]);
  const selectedTask =
    tasks.find((task) => task.id === selectedTaskID) ?? tasks[0] ?? null;

  async function handleStatusChange(id: number, status: TaskStatus) {
    setNotice("");
    try {
      await updateTaskStatus(id, status);
      await load();
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "failed to update task");
    }
  }

  async function handleSettingsSubmit(settings: Settings) {
    setNotice("");
    try {
      await updateSettings(settings);
      await load();
      setNotice("Settings saved");
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "failed to save settings");
    }
  }

  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Tasq Orchestrator</p>
          <h1>Task Operations</h1>
        </div>
        <div className="status-strip">
          <span>{summary ? `${tasks.length} tasks` : "loading"}</span>
          <span>{summary ? `${summary.agents.length} active agents` : "..."}</span>
          <button className="icon-button" type="button" onClick={() => void load()} aria-label="Refresh">
            Refresh
          </button>
        </div>
      </header>

      <nav className="tabs" aria-label="Views">
        {tabs.map((tab) => (
          <button
            key={tab}
            type="button"
            className={activeTab === tab ? "tab active" : "tab"}
            onClick={() => setActiveTab(tab)}
          >
            {tabTitle(tab)}
          </button>
        ))}
      </nav>

      {notice ? <p className="notice">{notice}</p> : null}

      {loadState.kind === "loading" ? <PanelMessage title="Loading" /> : null}
      {loadState.kind === "error" ? (
        <PanelMessage title="API unavailable" detail={loadState.message} />
      ) : null}
      {summary ? (
        <DashboardView
          activeTab={activeTab}
          summary={summary}
          selectedTask={selectedTask}
          onSelectTask={(task) => {
            setSelectedTaskID(task.id);
            setActiveTab("detail");
          }}
          onStatusChange={handleStatusChange}
          onSettingsSubmit={handleSettingsSubmit}
        />
      ) : null}
    </main>
  );
}

function DashboardView({
  activeTab,
  summary,
  selectedTask,
  onSelectTask,
  onStatusChange,
  onSettingsSubmit,
}: {
  activeTab: Tab;
  summary: Summary;
  selectedTask: Task | null;
  onSelectTask: (task: Task) => void;
  onStatusChange: (id: number, status: TaskStatus) => Promise<void>;
  onSettingsSubmit: (settings: Settings) => Promise<void>;
}) {
  if (activeTab === "agents") {
    return <AgentStatusPanel agents={summary.agents} />;
  }
  if (activeTab === "detail") {
    return selectedTask ? (
      <TaskDetail task={selectedTask} onStatusChange={onStatusChange} />
    ) : (
      <PanelMessage title="No task selected" />
    );
  }
  if (activeTab === "settings") {
    return (
      <SettingsPanel
        settings={summary.settings}
        onSubmit={onSettingsSubmit}
      />
    );
  }
  return (
    <KanbanBoard
      summary={summary}
      onSelectTask={onSelectTask}
      onStatusChange={onStatusChange}
    />
  );
}

function KanbanBoard({
  summary,
  onSelectTask,
  onStatusChange,
}: {
  summary: Summary;
  onSelectTask: (task: Task) => void;
  onStatusChange: (id: number, status: TaskStatus) => Promise<void>;
}) {
  return (
    <section className="board" aria-label="Kanban board">
      {summary.columns.map((column) => (
        <div className="column" key={column.status}>
          <div className="column-header">
            <h2>{column.title}</h2>
            <span>{column.tasks.length}</span>
          </div>
          <div className="task-list">
            {column.tasks.length === 0 ? (
              <p className="empty">No tasks</p>
            ) : (
              column.tasks.map((task) => (
                <TaskCard
                  key={task.id}
                  task={task}
                  onSelect={() => onSelectTask(task)}
                  onStatusChange={onStatusChange}
                />
              ))
            )}
          </div>
        </div>
      ))}
    </section>
  );
}

function TaskCard({
  task,
  onSelect,
  onStatusChange,
}: {
  task: Task;
  onSelect: () => void;
  onStatusChange: (id: number, status: TaskStatus) => Promise<void>;
}) {
  return (
    <article className="task-card">
      <button type="button" className="task-title" onClick={onSelect}>
        #{task.id} {task.title}
      </button>
      <p>{task.description || "No description"}</p>
      <div className="meta-row">
        <span className={`priority ${task.priority}`}>{task.priority}</span>
        <span>{task.agentStatus}</span>
      </div>
      <select
        aria-label={`Move ${task.title}`}
        value={task.status}
        onChange={(event) =>
          void onStatusChange(task.id, event.target.value as TaskStatus)
        }
      >
        {taskStatuses.map((status) => (
          <option key={status} value={status}>
            {statusLabels[status]}
          </option>
        ))}
      </select>
    </article>
  );
}

function AgentStatusPanel({ agents }: { agents: Task[] }) {
  return (
    <section className="panel-grid">
      <div className="wide-panel">
        <h2>Agent Realtime Status</h2>
        {agents.length === 0 ? (
          <p className="empty">No queued or running agents</p>
        ) : (
          <div className="agent-list">
            {agents.map((agent) => (
              <article className="agent-row" key={agent.id}>
                <div>
                  <strong>#{agent.id} {agent.title}</strong>
                  <span>{agent.workspace || "workspace pending"}</span>
                </div>
                <span className="agent-state">{agent.agentStatus}</span>
              </article>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}

function TaskDetail({
  task,
  onStatusChange,
}: {
  task: Task;
  onStatusChange: (id: number, status: TaskStatus) => Promise<void>;
}) {
  return (
    <section className="detail-layout">
      <article className="wide-panel">
        <h2>#{task.id} {task.title}</h2>
        <p className="description">{task.description || "No description"}</p>
        <dl className="detail-list">
          <div><dt>Status</dt><dd>{statusLabels[task.status]}</dd></div>
          <div><dt>Priority</dt><dd>{task.priority}</dd></div>
          <div><dt>Agent</dt><dd>{task.agentStatus}</dd></div>
          <div><dt>Assignee</dt><dd>{task.assignee || "unassigned"}</dd></div>
          <div><dt>Source</dt><dd>{task.source || "manual"} {task.sourceId}</dd></div>
          <div><dt>Workspace</dt><dd>{task.workspace || "pending"}</dd></div>
          <div><dt>Attempts</dt><dd>{task.attempts}</dd></div>
        </dl>
        {task.lastError ? <p className="error-text">{task.lastError}</p> : null}
        <div className="actions">
          {taskStatuses.map((status) => (
            <button
              key={status}
              type="button"
              onClick={() => void onStatusChange(task.id, status)}
              disabled={task.status === status}
            >
              {statusLabels[status]}
            </button>
          ))}
        </div>
      </article>
    </section>
  );
}

function SettingsPanel({
  settings,
  onSubmit,
}: {
  settings: Settings;
  onSubmit: (settings: Settings) => Promise<void>;
}) {
  const [draft, setDraft] = useState(settings);

  useEffect(() => {
    setDraft(settings);
  }, [settings]);

  return (
    <section className="detail-layout">
      <form
        className="wide-panel settings-form"
        onSubmit={(event) => {
          event.preventDefault();
          void onSubmit(draft);
        }}
      >
        <h2>Global Settings</h2>
        <label>
          Poll interval seconds
          <input
            type="number"
            min="1"
            value={draft.pollIntervalSeconds}
            onChange={(event) =>
              setDraft({ ...draft, pollIntervalSeconds: Number(event.target.value) })
            }
          />
        </label>
        <label>
          Max concurrent runs
          <input
            type="number"
            min="1"
            value={draft.maxConcurrentRuns}
            onChange={(event) =>
              setDraft({ ...draft, maxConcurrentRuns: Number(event.target.value) })
            }
          />
        </label>
        <label>
          Workspace root
          <input
            value={draft.workspaceRoot}
            onChange={(event) =>
              setDraft({ ...draft, workspaceRoot: event.target.value })
            }
          />
        </label>
        <label>
          Tracker provider
          <input
            value={draft.trackerProvider}
            onChange={(event) =>
              setDraft({ ...draft, trackerProvider: event.target.value })
            }
          />
        </label>
        <label>
          Agent command
          <input
            value={draft.agentCommand}
            onChange={(event) =>
              setDraft({ ...draft, agentCommand: event.target.value })
            }
          />
        </label>
        <button type="submit">Save Settings</button>
      </form>
    </section>
  );
}

function PanelMessage({ title, detail }: { title: string; detail?: string }) {
  return (
    <section className="wide-panel">
      <h2>{title}</h2>
      {detail ? <p>{detail}</p> : null}
    </section>
  );
}

function firstTaskID(summary: Summary): number | null {
  return summary.columns.flatMap((column) => column.tasks)[0]?.id ?? null;
}

function tabTitle(tab: Tab): string {
  switch (tab) {
    case "kanban":
      return "Kanban";
    case "agents":
      return "Agents";
    case "detail":
      return "Detail";
    case "settings":
      return "Settings";
  }
}
