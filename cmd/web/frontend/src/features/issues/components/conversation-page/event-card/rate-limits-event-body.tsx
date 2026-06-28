import { useTranslation } from "react-i18next";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "./index.module.css";
import { MetricGrid } from "./metric-grid";
import { formatRateLimitWindow, rateLimitsView } from "./payload";

export function RateLimitsEventBody({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();
  const payload = rateLimitsView(event.payload_json);

  if (!payload) {
    return (
      <p className={styles.statusMessage}>
        {event.message || t("issues.detailPage.emptyConversationEvent")}
      </p>
    );
  }

  return (
    <section className={styles.metrics} aria-label={t("issues.detailPage.rateLimits")}>
      <MetricGrid
        items={[
          [t("issues.detailPage.limitId"), payload.limitId],
          [t("issues.detailPage.planType"), payload.planType],
          [
            t("issues.detailPage.primaryLimit"),
            formatRateLimitWindow(payload.primary, t),
          ],
          [
            t("issues.detailPage.secondaryLimit"),
            formatRateLimitWindow(payload.secondary, t),
          ],
          [
            t("issues.detailPage.rateLimitReachedType"),
            payload.rateLimitReachedType || t("issues.detailPage.rateLimitNotReached"),
          ],
        ]}
      />
    </section>
  );
}
