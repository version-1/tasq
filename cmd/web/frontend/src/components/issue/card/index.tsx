import { useTranslation } from "react-i18next";
import type { IssueStatus, IssueSummary } from "@/lib/types";
import { issueStatuses } from "@/lib/types";
import { Markdown } from "@/app/issues/_components/markdown";
import styles from "./index.module.css";

type IssueStatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;

export function IssueCard({
  issue,
  onSelect,
  onStatusChange,
}: {
  issue: IssueSummary;
  onSelect: () => void;
  onStatusChange: IssueStatusChangeHandler;
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

function priorityClassName(priority: IssueSummary["priority"]): string {
  if (priority === "high" || priority === "urgent") {
    return `${styles.priority} ${styles.warningPriority}`;
  }
  return styles.priority;
}
