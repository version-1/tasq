import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconProxy } from "@/components/ui/icon-proxy";
import type { OrchestratorConversationEvent } from "@/lib/types";
import { ApprovalRequestEventBody } from "./approval-request-event-body";
import { ItemCompletedEventBody } from "./item-completed-event-body";
import styles from "./index.module.css";
import { itemCompletedView, type Translator } from "./payload";
import { RateLimitsEventBody } from "./rate-limits-event-body";
import { StatusEventBody } from "./status-event-body";
import { TokenUsageEventBody } from "./token-usage-event-body";
import { TurnCompletedEventBody } from "./turn-completed-event-body";

export { ApprovalRequestEventBody } from "./approval-request-event-body";
export { ItemCompletedEventBody } from "./item-completed-event-body";
export { RateLimitsEventBody } from "./rate-limits-event-body";
export { StatusEventBody } from "./status-event-body";
export { TokenUsageEventBody } from "./token-usage-event-body";
export { TurnCompletedEventBody } from "./turn-completed-event-body";

export function EventCard({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();
  const bodyID = useId();
  const [isBodyOpen, setIsBodyOpen] = useState(true);
  const foldLabel = isBodyOpen
    ? t("issues.detailPage.collapseConversationEvent")
    : t("issues.detailPage.expandConversationEvent");

  return (
    <article
      className={[
        styles.card,
        event.event === "item/commandExecution/requestApproval" ? styles.approval : "",
      ].join(" ")}
    >
      <div className={styles.header}>
        <EventTitle event={event} t={t} />
        <div className={styles.headerMeta}>
          <span className={styles.time}>{formatDateTime(event.at)}</span>
          <button
            type="button"
            className={styles.foldButton}
            aria-controls={bodyID}
            aria-expanded={isBodyOpen}
            aria-label={foldLabel}
            title={foldLabel}
            onClick={() => setIsBodyOpen((current) => !current)}
          >
            <IconProxy className={styles.foldIcon} name="chevron-down" size={16} strokeWidth={2.3} />
          </button>
        </div>
      </div>
      <div id={bodyID} className={styles.body} hidden={!isBodyOpen}>
        <EventBody event={event} />
      </div>
    </article>
  );
}

function EventTitle({
  event,
  t,
}: {
  event: OrchestratorConversationEvent;
  t: Translator;
}) {
  if (event.event === "item/commandExecution/requestApproval") {
    return (
      <span className={[styles.type, styles.approvalBadge].join(" ")}>
        {eventLabel(event.event, t)}
      </span>
    );
  }

  if (event.event !== "item/completed") {
    return <span className={styles.type}>{eventLabel(event.event, t)}</span>;
  }

  const item = itemCompletedView(event.payload_json);
  return (
    <span className={styles.type}>
      {item.type}
      {item.command ? <code className={styles.command}>{item.command}</code> : null}
    </span>
  );
}

function EventBody({ event }: { event: OrchestratorConversationEvent }) {
  if (event.event === "turn_completed") {
    return <TurnCompletedEventBody event={event} />;
  }

  if (event.event === "item/completed") {
    return <ItemCompletedEventBody event={event} />;
  }

  if (event.event === "item/commandExecution/requestApproval") {
    return <ApprovalRequestEventBody event={event} />;
  }

  if (event.event === "thread/tokenUsage/updated") {
    return <TokenUsageEventBody event={event} />;
  }

  if (event.event === "account/rateLimits/updated") {
    return <RateLimitsEventBody event={event} />;
  }

  return <StatusEventBody event={event} />;
}

function eventLabel(event: OrchestratorConversationEvent["event"], t: Translator): string {
  if (event === "turn_completed") {
    return t("issues.detailPage.turnCompleted");
  }
  return t(`runStatuses.${event}`);
}

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}
