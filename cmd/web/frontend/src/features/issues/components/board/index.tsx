import { useTranslation } from "react-i18next";
import { IssueCard } from "@/features/issues/components/card";
import { IconProxy, type IconProxyName } from "@/components/ui/icon-proxy";
import type { IssueStatus, Summary } from "@/lib/types";
import { boardColumns } from "./board-columns";
import styles from "./index.module.css";

type StatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;

const boardActions = [
  { icon: "filter", titleKey: "issues.board.filter" },
  { icon: "arrow-up-down", titleKey: "issues.board.sort" },
  { icon: "ellipsis", showLabel: false, titleKey: "issues.board.view" },
] satisfies Array<{ icon: IconProxyName; showLabel?: boolean; titleKey: string }>;

export function IssueBoard({
  showFilterSortActions = true,
  summary,
  onAddIssue,
  onStatusChange,
}: {
  showFilterSortActions?: boolean;
  summary: Summary;
  onAddIssue: (status?: IssueStatus) => void;
  onStatusChange: StatusChangeHandler;
}) {
  const { t } = useTranslation();
  const issues = summary.columns.flatMap((column) => column.issues);
  const visibleBoardActions = showFilterSortActions
    ? boardActions
    : boardActions.filter((action) =>
      action.titleKey !== "issues.board.filter" && action.titleKey !== "issues.board.sort"
    );

  return (
    <section className={styles.board} aria-label={t("header.board")}>
      <div className={styles.boardToolbar}>
        <div className={styles.boardActions}>
          {visibleBoardActions.map((action) => (
            <button key={action.titleKey} type="button" aria-label={t(action.titleKey)}>
              <IconProxy name={action.icon} size={15} />
              {action.showLabel === false ? null : t(action.titleKey)}
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
