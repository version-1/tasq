"use client";

import type { IssueStatus, IssueSummary, Summary } from "@/lib/types";
import { issueStatuses } from "@/lib/types";
import styles from "./issues-view.module.css";

type StatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;

const statusLabels: Record<IssueStatus, string> = {
  backlog: "Backlog",
  ready: "Ready",
  in_progress: "In Progress",
  review: "Review",
  blocked: "Blocked",
  failed: "Failed",
  done: "Done",
};

export function IssuesView({
  summary,
  selectedIssue,
  onSelectIssue,
  onStatusChange,
}: {
  summary: Summary;
  selectedIssue: IssueSummary | null;
  onSelectIssue: (issueID: number) => void;
  onStatusChange: StatusChangeHandler;
}) {
  return (
    <div className={styles.issuesLayout}>
      <IssueBoard
        summary={summary}
        onSelectIssue={onSelectIssue}
        onStatusChange={onStatusChange}
      />
      {selectedIssue ? (
        <IssueDetail issue={selectedIssue} onStatusChange={onStatusChange} />
      ) : (
        <PanelMessage title="No issue selected" />
      )}
    </div>
  );
}

function IssueBoard({
  summary,
  onSelectIssue,
  onStatusChange,
}: {
  summary: Summary;
  onSelectIssue: (issueID: number) => void;
  onStatusChange: StatusChangeHandler;
}) {
  return (
    <section className={styles.board} aria-label="Issue board">
      {summary.columns.map((column) => (
        <div className={styles.column} key={column.status}>
          <div className={styles.columnHeader}>
            <h2>{column.title}</h2>
            <span>{column.issues.length}</span>
          </div>
          <div className={styles.taskList}>
            {column.issues.length === 0 ? (
              <p className={styles.empty}>No issues</p>
            ) : (
              column.issues.map((issue) => (
                <IssueCard
                  key={issue.id}
                  issue={issue}
                  onSelect={() => onSelectIssue(issue.id)}
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
  onStatusChange: StatusChangeHandler;
}) {
  return (
    <article className={styles.taskCard}>
      <button type="button" className={styles.taskTitle} onClick={onSelect}>
        #{issue.id} {issue.title}
      </button>
      <p>{issue.description || "No description"}</p>
      <div className={styles.metaRow}>
        <span className={priorityClassName(issue.priority)}>{issue.priority}</span>
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

function IssueDetail({
  issue,
  onStatusChange,
}: {
  issue: IssueSummary;
  onStatusChange: StatusChangeHandler;
}) {
  return (
    <section className={styles.detailLayout}>
      <article className={styles.widePanel}>
        <h2>#{issue.id} {issue.title}</h2>
        <p className={styles.description}>{issue.description || "No description"}</p>
        <dl className={styles.detailList}>
          <div><dt>Issue Status</dt><dd>{statusLabels[issue.status]}</dd></div>
          <div><dt>Priority</dt><dd>{issue.priority}</dd></div>
          <div><dt>Assignee</dt><dd>{issue.assignee || "unassigned"}</dd></div>
          <div><dt>Run Status</dt><dd>{issue.run?.status ?? "no run"}</dd></div>
          <div><dt>Workspace</dt><dd>{issue.run?.workspace || "pending"}</dd></div>
          <div><dt>Attempt</dt><dd>{issue.run?.attempt ?? 0}</dd></div>
        </dl>
        {issue.run?.error ? <p className={styles.errorText}>{issue.run.error}</p> : null}
        <div className={styles.actions}>
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

function PanelMessage({ title }: { title: string }) {
  return (
    <section className={styles.widePanel}>
      <h2>{title}</h2>
    </section>
  );
}

function priorityClassName(priority: IssueSummary["priority"]): string {
  if (priority === "high" || priority === "urgent") {
    return `${styles.priority} ${styles.warningPriority}`;
  }
  return styles.priority;
}
