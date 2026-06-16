import { useTranslation } from "react-i18next";
import { IssueCard } from "@/components/issue/card";
import { IconProxy, type IconProxyName } from "@/components/ui/icon-proxy";
import type { IssueStatus, Summary } from "@/lib/types";
import { boardColumns } from "./board-columns";
import styles from "./index.module.css";

type StatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;

const boardActions = [
  { icon: "filter", titleKey: "issues.board.filter" },
  { icon: "arrow-up-down", titleKey: "issues.board.sort" },
  { icon: "layout-grid", titleKey: "issues.board.view" },
] satisfies Array<{ icon: IconProxyName; titleKey: string }>;

export function IssueBoard({
  summary,
  onAddIssue,
  onStatusChange,
}: {
  summary: Summary;
  onAddIssue: (status?: IssueStatus) => void;
  onStatusChange: StatusChangeHandler;
}) {
  const { t } = useTranslation();
  const issues = summary.columns.flatMap((column) => column.issues);

  return (
    <section className={styles.board} aria-label={t("header.board")}>
      <div className={styles.boardToolbar}>
        <div className={styles.boardActions}>
          {boardActions.map((action) => (
            <button key={action.titleKey} type="button">
              <IconProxy name={action.icon} size={15} />
              {t(action.titleKey)}
            </button>
          ))}
        </div>
      </div>

      <div className={styles.columns}>
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
                      commentCount={issue.stats.commentCount}
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
      </div>
    </section>
  );
}
