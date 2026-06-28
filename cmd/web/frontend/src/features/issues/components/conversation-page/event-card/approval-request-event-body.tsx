import { useTranslation } from "react-i18next";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "./index.module.css";
import { extractApprovalRequestDetails } from "./payload";

export function ApprovalRequestEventBody({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();
  const details = extractApprovalRequestDetails(event.payload_json);
  const reason = details.reason || event.message || t("issues.detailPage.approvalRequestReasonFallback");
  const command = details.command || t("issues.detailPage.approvalRequestCommandFallback");

  return (
    <section className={styles.approvalRequest} aria-label={t("issues.detailPage.approvalRequest")}>
      <h3 className={styles.approvalReason}>{reason}</h3>
      <pre className={styles.approvalCommand}>
        <code>{command}</code>
      </pre>
    </section>
  );
}
