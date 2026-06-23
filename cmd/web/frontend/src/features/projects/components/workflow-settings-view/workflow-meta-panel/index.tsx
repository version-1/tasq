import { useTranslation } from "react-i18next";
import { formatDateTime } from "../format";
import styles from "./index.module.css";

type WorkflowMetaPanelProps = {
  updatedAt: string;
};

export function WorkflowMetaPanel({ updatedAt }: WorkflowMetaPanelProps) {
  const { t } = useTranslation();

  return (
    <div className={styles.metaPanel}>
      <span className={styles.metaLabel}>{t("projectSettings.syncedAt")}</span>
      <time className={styles.metaValue} dateTime={updatedAt}>
        {formatDateTime(updatedAt)}
      </time>
    </div>
  );
}
