import { useTranslation } from "react-i18next";
import { issueStatuses } from "@/lib/types";
import type { IssueStatus } from "@/lib/types";
import styles from "./index.module.css";

export function StatusActions({
  currentStatus,
  disabled,
  embedded = false,
  onStatusChange,
}: {
  currentStatus: IssueStatus;
  disabled: boolean;
  embedded?: boolean;
  onStatusChange: (status: IssueStatus) => Promise<void>;
}) {
  const { t } = useTranslation();

  return (
    <section
      className={embedded ? `${styles.section} ${styles.embeddedSection}` : styles.section}
      aria-labelledby="issue-status-actions"
    >
      <div className={styles.sectionHeader}>
        <h3 id="issue-status-actions">{t("issues.detailPage.statusActions")}</h3>
      </div>
      <div className={styles.actions}>
        {issueStatuses.map((status) => (
          <button
            key={status}
            type="button"
            onClick={() => void onStatusChange(status)}
            disabled={disabled || status === currentStatus}
          >
            {t(`statuses.${status}`)}
          </button>
        ))}
      </div>
    </section>
  );
}
