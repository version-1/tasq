import { useTranslation } from "react-i18next";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "./index.module.css";
import { MetricGrid } from "./metric-grid";
import { formatNumber, tokenUsageView } from "./payload";

export function TokenUsageEventBody({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();
  const payload = tokenUsageView(event.payload_json);

  if (!payload) {
    return (
      <p className={styles.statusMessage}>
        {event.message || t("issues.detailPage.emptyConversationEvent")}
      </p>
    );
  }

  return (
    <section className={styles.metrics} aria-label={t("issues.detailPage.tokenUsage")}>
      <div className={styles.metricGroup}>
        <h3 className={styles.metricHeading}>{t("issues.detailPage.tokenUsageTotal")}</h3>
        <MetricGrid
          items={[
            [t("issues.detailPage.totalTokens"), formatNumber(payload.total.totalTokens)],
            [t("dashboard.inputTokens"), formatNumber(payload.total.inputTokens)],
            [t("issues.detailPage.cachedInputTokens"), formatNumber(payload.total.cachedInputTokens)],
            [t("dashboard.outputTokens"), formatNumber(payload.total.outputTokens)],
            [t("issues.detailPage.reasoningOutputTokens"), formatNumber(payload.total.reasoningOutputTokens)],
          ]}
        />
      </div>
      <div className={styles.metricGroup}>
        <h3 className={styles.metricHeading}>{t("issues.detailPage.tokenUsageLast")}</h3>
        <MetricGrid
          items={[
            [t("issues.detailPage.totalTokens"), formatNumber(payload.last.totalTokens)],
            [t("dashboard.inputTokens"), formatNumber(payload.last.inputTokens)],
            [t("issues.detailPage.cachedInputTokens"), formatNumber(payload.last.cachedInputTokens)],
            [t("dashboard.outputTokens"), formatNumber(payload.last.outputTokens)],
            [t("issues.detailPage.reasoningOutputTokens"), formatNumber(payload.last.reasoningOutputTokens)],
          ]}
        />
      </div>
      <p className={styles.metricNote}>
        {t("issues.detailPage.modelContextWindow", {
          count: formatNumber(payload.modelContextWindow),
        })}
      </p>
    </section>
  );
}
