import { useTranslation } from "react-i18next";
import type { ReactNode } from "react";
import type { Issue, IssueStatus } from "@/lib/types";
import { formatDateTime } from "../format";
import { MetaItem } from "./meta-item";
import styles from "./index.module.css";

export function IssueHeader({ children, issue }: { children?: ReactNode; issue: Issue }) {
  const { t } = useTranslation();

  return (
    <header className={styles.header}>
      <div className={styles.badges}>
        <span className={statusClassName(issue.status)}>{t(`statuses.${issue.status}`)}</span>
        <span className={priorityClassName(issue.priority)}>{t(`priorities.${issue.priority}`)}</span>
      </div>
      <dl className={styles.metaGrid}>
        <MetaItem label={t("issues.table.id")} value={`#${issue.id}`} />
        <MetaItem label={t("issues.detail.project")} value={issue.projectKey} />
        <MetaItem label={t("issues.assignee")} value={issue.assignee || t("issues.unassigned")} />
        <MetaItem label={t("issues.detailPage.createdAt")} value={formatDateTime(issue.createdAt)} />
        <MetaItem label={t("issues.detailPage.updatedAt")} value={formatDateTime(issue.updatedAt)} />
      </dl>
      {children ? <div className={styles.actionsSlot}>{children}</div> : null}
    </header>
  );
}

function priorityClassName(priority: Issue["priority"]): string {
  if (priority === "high" || priority === "urgent") {
    return `${styles.priorityBadge} ${styles.warningPriority}`;
  }
  return styles.priorityBadge;
}

function statusClassName(status: IssueStatus): string {
  if (status === "cancelled" || status === "duplicate") {
    return `${styles.statusBadge} ${styles.neutralStatus}`;
  }
  return styles.statusBadge;
}
