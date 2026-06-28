import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/ui/markdown";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "../../index.module.css";
import { extractApprovalRequestDetails } from "../../payload";

export function ApprovalRequestEventBody({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();
  const details = extractApprovalRequestDetails(event.payload_json);
  const reason = details.reason || event.message || t("issues.detailPage.approvalRequestReasonFallback");
  const command = details.command || t("issues.detailPage.approvalRequestCommandFallback");

  return (
    <section className={styles.approvalRequest} aria-label={t("issues.detailPage.approvalRequest")}>
      <div className={styles.approvalReasonBlock}>
        <span className={styles.approvalLabel}>{t("issues.detailPage.approvalReason")}</span>
        <p className={styles.approvalReason}>{reason}</p>
      </div>
      <Markdown
        className={styles.approvalCommandMarkdown}
        content={`\`\`\`shell\n${command}\n\`\`\``}
        emptyText={t("issues.detailPage.approvalRequestCommandFallback")}
      />
    </section>
  );
}
