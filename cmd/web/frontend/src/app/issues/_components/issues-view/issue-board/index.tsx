import { useTranslation } from "react-i18next";
import { IssueCard } from "@/components/issue/card";
import type { IssueStatus, Summary } from "@/lib/types";
import type { StatusChangeHandler } from "../types";
import { boardColumns } from "./board-columns";
import styles from "./index.module.css";

export function IssueBoard({
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
