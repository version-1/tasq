import { useTranslation } from "react-i18next";
import { attachmentContentURL } from "@/lib/api";
import type { Attachment } from "@/lib/types";
import { formatDateTime } from "../format";
import styles from "./index.module.css";

type AttachmentsSectionProps = {
  attachments: Attachment[] | null | undefined;
  error: string;
  isLoading: boolean;
};

export function AttachmentsSection({
  attachments,
  error,
  isLoading,
}: AttachmentsSectionProps) {
  const { t } = useTranslation();
  const visibleAttachments = attachments ?? [];

  return (
    <section className={styles.section} aria-labelledby="issue-attachments">
      <div className={styles.sectionHeader}>
        <h3 id="issue-attachments">{t("issues.detailPage.attachments")}</h3>
        <span>{t("issues.detailPage.attachmentCount", { count: visibleAttachments.length })}</span>
      </div>
      {isLoading ? <p className={styles.muted}>{t("layout.loading")}</p> : null}
      {error ? <p className={styles.error}>{error}</p> : null}
      {!isLoading && !error && visibleAttachments.length === 0 ? (
        <p className={styles.muted}>{t("issues.detailPage.noAttachments")}</p>
      ) : null}
      {visibleAttachments.length > 0 ? (
        <ul className={styles.attachmentList}>
          {visibleAttachments.map((attachment) => (
            <li key={attachment.id} className={styles.attachmentItem}>
              <div className={styles.attachmentMain}>
                <a href={attachmentContentURL(attachment.id)} target="_blank" rel="noreferrer">
                  {attachment.filename}
                </a>
                <span>{attachment.contentType}</span>
              </div>
              <div className={styles.attachmentMeta}>
                <span>{formatFileSize(attachment.size)}</span>
                <time dateTime={attachment.createdAt}>{formatDateTime(attachment.createdAt)}</time>
              </div>
            </li>
          ))}
        </ul>
      ) : null}
    </section>
  );
}

function formatFileSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) {
    return "";
  }
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
