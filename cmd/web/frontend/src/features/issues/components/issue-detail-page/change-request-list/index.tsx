import { useTranslation } from "react-i18next";
import type { ChangeRequest } from "@/lib/types";
import { Markdown } from "@/components/ui/markdown";
import { formatDateTime } from "../format";
import styles from "./index.module.css";

export function ChangeRequestList({
  changeRequests,
  error,
  isLoading,
}: {
  changeRequests: ChangeRequest[];
  error: string;
  isLoading: boolean;
}) {
  const { t } = useTranslation();

  return (
    <section className={styles.section} aria-labelledby="issue-change-requests">
      <div className={styles.sectionHeader}>
        <h3 id="issue-change-requests">{t("issues.detailPage.changeRequests")}</h3>
        <span>{t("issues.detailPage.changeRequestCount", { count: changeRequests.length })}</span>
      </div>
      {changeRequests.length === 0 && !isLoading ? (
        <p className={styles.emptyText}>{t("issues.detailPage.noChangeRequests")}</p>
      ) : null}
      <div className={styles.requestList}>
        {changeRequests.map((changeRequest) => (
          <article key={changeRequest.id} className={styles.request}>
            <div className={styles.requestHeader}>
              <div>
                <strong>{changeRequest.author}</strong>
                <span>{t(`changeRequests.statuses.${changeRequest.status}`)}</span>
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
        ))}
      </div>
      {error ? <p className={styles.errorText}>{error}</p> : null}
    </section>
  );
}
