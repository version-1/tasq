import { useTranslation } from "react-i18next";
import type { Issue } from "@/lib/types";
import { formatDateTime } from "../format";
import { MetaItem } from "./meta-item";
import styles from "./index.module.css";

export function IssueHeader({ issue }: { issue: Issue }) {
  const { t } = useTranslation();

  return (
    <header className={styles.header}>
      <div className={styles.titleBlock}>
        <p className={styles.issueID}>#{issue.id}</p>
        <h2>{issue.title}</h2>
      </div>
      <div className={styles.badges}>
        <span className={styles.statusBadge}>{t(`statuses.${issue.status}`)}</span>
        <span className={priorityClassName(issue.priority)}>{t(`priorities.${issue.priority}`)}</span>
      </div>
      <dl className={styles.metaGrid}>
        <MetaItem label={t("issues.detail.project")} value={issue.projectKey} />
        <MetaItem label={t("issues.assignee")} value={issue.assignee || t("issues.unassigned")} />
        <MetaItem label={t("issues.detailPage.createdAt")} value={formatDateTime(issue.createdAt)} />
        <MetaItem label={t("issues.detailPage.updatedAt")} value={formatDateTime(issue.updatedAt)} />
      </dl>
    </header>
  );
}

function priorityClassName(priority: Issue["priority"]): string {
  if (priority === "high" || priority === "urgent") {
    return `${styles.priorityBadge} ${styles.warningPriority}`;
  }
  return styles.priorityBadge;
}
