import { useTranslation } from "react-i18next";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "../../index.module.css";
import {
  formatNumber,
  itemCompletedView,
  rateLimitsView,
  tokenUsageView,
} from "../../payload";

const PREVIEW_LENGTH = 100;

export function EventBodyPreview({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();
  const text = eventPreviewText(event, t);

  return <p className={styles.bodyPreview}>{truncatePreview(text)}</p>;
}

function eventPreviewText(
  event: OrchestratorConversationEvent,
  t: (key: string, options?: Record<string, unknown>) => string,
): string {
  if (event.event === "item/completed") {
    const item = itemCompletedView(event.payload_json);
    return [item.aggregatedOutput, item.text, event.message].find((value) => value.trim() !== "")
      ?? t("issues.detailPage.emptyConversationEvent");
  }

  if (event.event === "thread/tokenUsage/updated") {
    const payload = tokenUsageView(event.payload_json);
    if (payload) {
      return t("issues.detailPage.tokenUsagePreview", {
        total: formatNumber(payload.total.totalTokens),
        last: formatNumber(payload.last.totalTokens),
      });
    }
  }

  if (event.event === "account/rateLimits/updated") {
    const payload = rateLimitsView(event.payload_json);
    if (payload) {
      return t("issues.detailPage.rateLimitsPreview", {
        plan: payload.planType,
        primary: payload.primary.usedPercent,
        secondary: payload.secondary.usedPercent,
      });
    }
  }

  return event.message || t(`runStatuses.${event.event}`);
}

function truncatePreview(value: string): string {
  const normalized = value.replace(/\s+/g, " ").trim();
  const chars = Array.from(normalized);
  if (chars.length <= PREVIEW_LENGTH) {
    return normalized;
  }
  return `${chars.slice(0, PREVIEW_LENGTH).join("")}...`;
}
