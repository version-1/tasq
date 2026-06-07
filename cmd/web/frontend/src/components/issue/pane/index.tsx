import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { IssueStatus, IssueSummary } from "@/lib/types";
import { issueStatuses } from "@/lib/types";
import { Markdown } from "@/app/issues/_components/markdown";
import styles from "./index.module.css";

type IssueStatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;

export function IssueDetail({
  issue,
  onStatusChange,
}: {
  issue: IssueSummary;
  onStatusChange: IssueStatusChangeHandler;
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
        <Link className={styles.detailLink} to={`/issues/${issue.id}`}>
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
