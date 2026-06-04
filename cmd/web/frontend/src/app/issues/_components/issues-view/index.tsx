"use client";

import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { IssueStatus, IssueSummary, Summary } from "@/lib/types";
import { issueStatuses } from "@/lib/types";
import { Markdown } from "../markdown";
import styles from "./index.module.css";

type StatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;
type BoardColumnKey = "draft" | "todo" | "inProgress" | "inReview";

type BoardColumn = {
  key: BoardColumnKey;
  titleKey: string;
  statuses: IssueStatus[];
};

const boardColumns: BoardColumn[] = [
  { key: "draft", titleKey: "issues.board.draft", statuses: ["backlog", "blocked", "failed"] },
  { key: "todo", titleKey: "issues.board.todo", statuses: ["ready"] },
  { key: "inProgress", titleKey: "issues.board.inProgress", statuses: ["in_progress"] },
  { key: "inReview", titleKey: "issues.board.inReview", statuses: ["review", "done"] },
];

export function IssuesView({
  summary,
  selectedIssue,
  onSelectIssue,
  onAddIssue,
  onStatusChange,
}: {
  summary: Summary;
  selectedIssue: IssueSummary | null;
  onSelectIssue: (issueID: number) => void;
  onAddIssue: (status?: IssueStatus) => void;
  onStatusChange: StatusChangeHandler;
}) {
  const { t } = useTranslation();

  return (
    <div className={styles.issuesLayout}>
      <IssueBoard
        summary={summary}
        onSelectIssue={onSelectIssue}
        onAddIssue={onAddIssue}
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
  onAddIssue,
  onStatusChange,
}: {
  summary: Summary;
  onSelectIssue: (issueID: number) => void;
  onAddIssue: (status?: IssueStatus) => void;
  onStatusChange: StatusChangeHandler;
}) {
  const { t } = useTranslation();
  const issues = summary.columns.flatMap((column) => column.issues);

  return (
    <section className={styles.board} aria-label={t("header.board")}>
      {boardColumns.map((column) => {
        const columnIssues = issues.filter((issue) => column.statuses.includes(issue.status));

        return (
          <div className={styles.column} key={column.key}>
            <div className={styles.columnHeader}>
              <div className={styles.columnTitle}>
                {column.key === "inReview" ? <span className={styles.reviewIcon} aria-hidden="true">○</span> : null}
                <h2>{t(column.titleKey)}</h2>
                <span>{columnIssues.length}</span>
              </div>
              <div className={styles.columnActions}>
                <button
                  type="button"
                  aria-label={t("issues.board.addTask")}
                  onClick={() => onAddIssue(column.statuses[0])}
                >
                  ＋
                </button>
                <button type="button" aria-label={t("issues.board.columnActions")}>···</button>
              </div>
            </div>
            <div className={styles.taskList}>
              {columnIssues.length === 0 ? (
                <p className={styles.empty}>{t("issues.noIssues")}</p>
              ) : (
                columnIssues.map((issue) => (
                  <IssueCard
                    key={issue.id}
                    issue={issue}
                    onSelect={() => onSelectIssue(issue.id)}
                    onStatusChange={onStatusChange}
                  />
                ))
              )}
            </div>
            <button
              className={styles.addTaskButton}
              type="button"
              onClick={() => onAddIssue(column.statuses[0])}
            >
              <span aria-hidden="true">＋</span>
              {t("issues.board.addTask")}
            </button>
          </div>
        );
      })}
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
      <Markdown
        className={styles.cardMarkdown}
        content={issue.description}
        emptyText={t("issues.noDescription")}
      />
      <div className={styles.metaRow}>
        <span className={styles.projectKey}>{issue.projectKey}</span>
        <span className={priorityClassName(issue.priority)}>{t(`priorities.${issue.priority}`)}</span>
        <span>{t(`statuses.${issue.status}`)}</span>
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
        <Markdown
          className={styles.description}
          content={issue.description}
          emptyText={t("issues.noDescription")}
        />
        <Link className={styles.detailLink} to={`/issues/detail?id=${issue.id}`}>
          {t("issues.detail.openDetail")}
        </Link>
        <dl className={styles.detailList}>
          <div><dt>{t("issues.detail.project")}</dt><dd>{issue.projectKey}</dd></div>
          <div><dt>{t("issues.detail.issueStatus")}</dt><dd>{t(`statuses.${issue.status}`)}</dd></div>
          <div><dt>{t("issues.detail.priority")}</dt><dd>{t(`priorities.${issue.priority}`)}</dd></div>
          <div><dt>{t("issues.assignee")}</dt><dd>{issue.assignee || t("issues.unassigned")}</dd></div>
        </dl>
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
