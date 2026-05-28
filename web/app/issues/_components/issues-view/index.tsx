"use client";

import { useTranslation } from "react-i18next";
import type { IssueStatus, IssueSummary, Summary } from "@/lib/types";
import { issueStatuses } from "@/lib/types";
import styles from "./index.module.css";

type StatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;

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
  const { t } = useTranslation();

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
        <PanelMessage title={t("issues.noIssueSelected")} />
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
  const { t } = useTranslation();

  return (
    <section className={styles.board} aria-label={t("header.board")}>
      {summary.columns.map((column) => (
        <div className={styles.column} key={column.status}>
          <div className={styles.columnHeader}>
            <h2>{t(`statuses.${column.status}`)}</h2>
            <span>{column.issues.length}</span>
          </div>
          <div className={styles.taskList}>
            {column.issues.length === 0 ? (
              <p className={styles.empty}>{t("issues.noIssues")}</p>
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
  const { t } = useTranslation();

  return (
    <article className={styles.taskCard}>
      <button type="button" className={styles.taskTitle} onClick={onSelect}>
        #{issue.id} {issue.title}
      </button>
      <p>{issue.description || t("issues.noDescription")}</p>
      <div className={styles.metaRow}>
        <span className={priorityClassName(issue.priority)}>{t(`priorities.${issue.priority}`)}</span>
        <span>{issue.run ? t(`runStatuses.${issue.run.status}`) : t("issues.noRun")}</span>
      </div>
      <select
        aria-label={t("issues.moveLabel", { title: issue.title })}
        value={issue.status}
        onChange={(event) =>
          void onStatusChange(issue.id, event.target.value as IssueStatus)
        }
      >
        {issueStatuses.map((status) => (
          <option key={status} value={status}>
            {t(`statuses.${status}`)}
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
  const { t } = useTranslation();

  return (
    <section className={styles.detailLayout}>
      <article className={styles.widePanel}>
        <h2>#{issue.id} {issue.title}</h2>
        <p className={styles.description}>{issue.description || t("issues.noDescription")}</p>
        <dl className={styles.detailList}>
          <div><dt>{t("issues.detail.issueStatus")}</dt><dd>{t(`statuses.${issue.status}`)}</dd></div>
          <div><dt>{t("issues.detail.priority")}</dt><dd>{t(`priorities.${issue.priority}`)}</dd></div>
          <div><dt>{t("issues.assignee")}</dt><dd>{issue.assignee || t("issues.unassigned")}</dd></div>
          <div><dt>{t("issues.detail.runStatus")}</dt><dd>{issue.run ? t(`runStatuses.${issue.run.status}`) : t("issues.noRun")}</dd></div>
          <div><dt>{t("issues.detail.workspace")}</dt><dd>{issue.run?.workspace || t("common.pending")}</dd></div>
          <div><dt>{t("issues.attempt")}</dt><dd>{issue.run?.attempt ?? 0}</dd></div>
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
              {t(`statuses.${status}`)}
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
