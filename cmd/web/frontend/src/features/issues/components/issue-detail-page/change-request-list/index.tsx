import { useTranslation } from "react-i18next";
import type { ChangeRequest } from "@/lib/types";
import { ChangeRequestCard } from "../change-request-card";
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
          <ChangeRequestCard key={changeRequest.id} changeRequest={changeRequest} />
        ))}
      </div>
      {error ? <p className={styles.errorText}>{error}</p> : null}
    </section>
  );
}
