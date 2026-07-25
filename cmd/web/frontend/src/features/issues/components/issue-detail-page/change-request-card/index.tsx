import { useTranslation } from "react-i18next";
import type { ChangeRequest } from "@/lib/types";
import { Markdown } from "@/components/ui/markdown";
import { ChangeRequestStatusBadge } from "@/features/issues/components/change-request-status-badge";
import { formatDateTime } from "../format";
import styles from "./index.module.css";

export function ChangeRequestCard({ changeRequest }: { changeRequest: ChangeRequest }) {
  const { t } = useTranslation();

  return (
    <article className={styles.request}>
      <div className={styles.requestHeader}>
        <div>
          <strong>{changeRequest.author}</strong>
          <ChangeRequestStatusBadge status={changeRequest.status} />
        </div>
        <time dateTime={changeRequest.createdAt}>{formatDateTime(changeRequest.createdAt)}</time>
      </div>
      <Markdown
        className={styles.markdown}
        content={changeRequest.body}
        emptyText={t("issues.detailPage.emptyChangeRequest")}
      />
      {changeRequest.resolvedAt ? (
        <p className={styles.resolvedMeta}>
          {t("issues.detailPage.changeRequestResolved", {
            runID: changeRequest.resolvedByRunId ?? t("issues.detailPage.noRun"),
            time: formatDateTime(changeRequest.resolvedAt),
          })}
        </p>
      ) : null}
    </article>
  );
}
