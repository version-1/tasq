"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { fetchSummary, updateIssueStatus } from "@/lib/api";
import type { IssueStatus, IssueSummary, RunSnapshot, Summary } from "@/lib/types";
import { issueStatuses } from "@/lib/types";

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; summary: Summary }
  | { kind: "error"; message: string };

const tabs = ["issues", "runs", "detail"] as const;
type Tab = (typeof tabs)[number];

const statusLabels: Record<IssueStatus, string> = {
  backlog: "Backlog",
  ready: "Ready",
  in_progress: "In Progress",
  review: "Review",
  blocked: "Blocked",
  failed: "Failed",
  done: "Done",
};

export default function Home() {
  const [loadState, setLoadState] = useState<LoadState>({ kind: "loading" });
  const [activeTab, setActiveTab] = useState<Tab>("issues");
  const [selectedIssueID, setSelectedIssueID] = useState<number | null>(null);
  const [notice, setNotice] = useState("");

  const load = useCallback(async () => {
    try {
      const summary = await fetchSummary();
      setLoadState({ kind: "ready", summary });
      setSelectedIssueID((current) => current ?? firstIssueID(summary));
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
  const issues = useMemo(() => {
    if (!summary) return [];
    return summary.columns.flatMap((column) => column.issues);
  }, [summary]);
  const selectedIssue =
    issues.find((issue) => issue.id === selectedIssueID) ?? issues[0] ?? null;

  async function handleStatusChange(id: number, status: IssueStatus) {
    setNotice("");
    try {
      await updateIssueStatus(id, status);
      await load();
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "failed to update issue");
    }
  }

  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Tasq Issue Tracker</p>
          <h1>Issue Operations</h1>
        </div>
        <div className="status-strip">
          <span>{summary ? `${issues.length} issues` : "loading"}</span>
          <span>{summary ? `${summary.runs.length} tracked runs` : "..."}</span>
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
          selectedIssue={selectedIssue}
          onSelectIssue={(issue) => {
            setSelectedIssueID(issue.id);
            setActiveTab("detail");
          }}
          onStatusChange={handleStatusChange}
        />
      ) : null}
    </main>
  );
}

function DashboardView({
  activeTab,
  summary,
  selectedIssue,
  onSelectIssue,
  onStatusChange,
}: {
  activeTab: Tab;
  summary: Summary;
  selectedIssue: IssueSummary | null;
  onSelectIssue: (issue: IssueSummary) => void;
  onStatusChange: (id: number, status: IssueStatus) => Promise<void>;
}) {
  if (activeTab === "runs") {
    return <RunStatusPanel runs={summary.runs} />;
  }
  if (activeTab === "detail") {
    return selectedIssue ? (
      <IssueDetail issue={selectedIssue} onStatusChange={onStatusChange} />
    ) : (
      <PanelMessage title="No issue selected" />
    );
  }
  return (
    <IssueBoard
      summary={summary}
      onSelectIssue={onSelectIssue}
      onStatusChange={onStatusChange}
    />
  );
}

function IssueBoard({
  summary,
  onSelectIssue,
  onStatusChange,
}: {
  summary: Summary;
  onSelectIssue: (issue: IssueSummary) => void;
  onStatusChange: (id: number, status: IssueStatus) => Promise<void>;
}) {
  return (
    <section className="board" aria-label="Issue board">
      {summary.columns.map((column) => (
        <div className="column" key={column.status}>
          <div className="column-header">
            <h2>{column.title}</h2>
            <span>{column.issues.length}</span>
          </div>
          <div className="task-list">
            {column.issues.length === 0 ? (
              <p className="empty">No issues</p>
            ) : (
              column.issues.map((issue) => (
                <IssueCard
                  key={issue.id}
                  issue={issue}
                  onSelect={() => onSelectIssue(issue)}
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

function IssueCard({
  issue,
  onSelect,
  onStatusChange,
}: {
  issue: IssueSummary;
  onSelect: () => void;
  onStatusChange: (id: number, status: IssueStatus) => Promise<void>;
}) {
  return (
    <article className="task-card">
      <button type="button" className="task-title" onClick={onSelect}>
        #{issue.id} {issue.title}
      </button>
      <p>{issue.description || "No description"}</p>
      <div className="meta-row">
        <span className={`priority ${issue.priority}`}>{issue.priority}</span>
        <span>{issue.run?.status ?? "no run"}</span>
      </div>
      <select
        aria-label={`Move ${issue.title}`}
        value={issue.status}
        onChange={(event) =>
          void onStatusChange(issue.id, event.target.value as IssueStatus)
        }
      >
        {issueStatuses.map((status) => (
          <option key={status} value={status}>
            {statusLabels[status]}
          </option>
        ))}
      </select>
    </article>
  );
}

function RunStatusPanel({ runs }: { runs: RunSnapshot[] }) {
  return (
    <section className="panel-grid">
      <div className="wide-panel">
        <h2>Run Status</h2>
        {runs.length === 0 ? (
          <p className="empty">No orchestrator runs yet</p>
        ) : (
          <div className="agent-list">
            {runs.map((run) => (
              <article className="agent-row" key={`${run.workItemId}-${run.runId}`}>
                <div>
                  <strong>Issue #{run.issueId}</strong>
                  <span>{run.workspace || "workspace pending"}</span>
                </div>
                <span className="agent-state">{run.status}</span>
              </article>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}

function IssueDetail({
  issue,
  onStatusChange,
}: {
  issue: IssueSummary;
  onStatusChange: (id: number, status: IssueStatus) => Promise<void>;
}) {
  return (
    <section className="detail-layout">
      <article className="wide-panel">
        <h2>#{issue.id} {issue.title}</h2>
        <p className="description">{issue.description || "No description"}</p>
        <dl className="detail-list">
          <div><dt>Issue Status</dt><dd>{statusLabels[issue.status]}</dd></div>
          <div><dt>Priority</dt><dd>{issue.priority}</dd></div>
          <div><dt>Assignee</dt><dd>{issue.assignee || "unassigned"}</dd></div>
          <div><dt>Run Status</dt><dd>{issue.run?.status ?? "no run"}</dd></div>
          <div><dt>Workspace</dt><dd>{issue.run?.workspace || "pending"}</dd></div>
          <div><dt>Attempt</dt><dd>{issue.run?.attempt ?? 0}</dd></div>
        </dl>
        {issue.run?.error ? <p className="error-text">{issue.run.error}</p> : null}
        <div className="actions">
          {issueStatuses.map((status) => (
            <button
              key={status}
              type="button"
              onClick={() => void onStatusChange(issue.id, status)}
              disabled={issue.status === status}
            >
              {statusLabels[status]}
            </button>
          ))}
        </div>
      </article>
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

function firstIssueID(summary: Summary): number | null {
  return summary.columns.flatMap((column) => column.issues)[0]?.id ?? null;
}

function tabTitle(tab: Tab): string {
  switch (tab) {
    case "issues":
      return "Issues";
    case "runs":
      return "Runs";
    case "detail":
      return "Detail";
  }
}
