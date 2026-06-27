import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/ui/markdown";
import type { OrchestratorConversationEvent } from "@/lib/types";
import styles from "./index.module.css";

type Translator = (key: string, options?: Record<string, unknown>) => string;

type ApprovalRequestDetails = {
  reason: string;
  command: string;
};

export function EventCard({ event }: { event: OrchestratorConversationEvent }) {
  const { t } = useTranslation();

  return (
    <article
      className={[
        styles.card,
        event.event === "item/commandExecution/requestApproval" ? styles.approval : "",
      ].join(" ")}
    >
      <div className={styles.header}>
        <EventTitle event={event} t={t} />
        <span className={styles.time}>{formatDateTime(event.at)}</span>
      </div>
      <EventBody event={event} t={t} />
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

function EventBody({
  event,
  t,
}: {
  event: OrchestratorConversationEvent;
  t: Translator;
}) {
  if (event.event === "turn_completed") {
    return (
      <Markdown
        className={styles.markdown}
        content={extractAggregatedOutput(event.payload_json)}
        emptyText={event.message || t("issues.detailPage.emptyConversationEvent")}
      />
    );
  }

  if (event.event === "item/completed") {
    const item = itemCompletedView(event.payload_json);
    const content = [item.aggregatedOutput, item.text].filter(Boolean).join("\n\n");
    return (
      <div className={styles.itemBody}>
        <Markdown
          className={styles.markdown}
          content={content}
          emptyText={event.message || t("issues.detailPage.emptyConversationEvent")}
        />
        {item.exitCode !== null && item.exitCode !== 0 ? (
          <p className={styles.exitCode}>{t("issues.detailPage.exitCode", { code: item.exitCode })}</p>
        ) : null}
      </div>
    );
  }

  if (event.event === "item/commandExecution/requestApproval") {
    return <ApprovalRequest event={event} t={t} />;
  }

  return (
    <p className={styles.statusMessage}>
      {event.message || t(`runStatuses.${event.event}`)}
    </p>
  );
}

function ApprovalRequest({
  event,
  t,
}: {
  event: OrchestratorConversationEvent;
  t: Translator;
}) {
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

function eventLabel(event: OrchestratorConversationEvent["event"], t: Translator): string {
  if (event === "turn_completed") {
    return t("issues.detailPage.turnCompleted");
  }
  return t(`runStatuses.${event}`);
}

function extractAggregatedOutput(payloadJSON: string): string {
  const payload = parsePayloadJSON(payloadJSON);
  return typeof payload === "string" ? payload : findAggregatedOutput(payload);
}

function extractApprovalRequestDetails(payloadJSON: string): ApprovalRequestDetails {
  const payload = parsePayloadJSON(payloadJSON);
  if (typeof payload === "string") {
    return { reason: "", command: "" };
  }
  return {
    reason: findStringField(payload, "reason"),
    command: findStringField(payload, "command"),
  };
}

type ItemCompletedView = {
  type: string;
  command: string;
  aggregatedOutput: string;
  text: string;
  exitCode: number | null;
};

function itemCompletedView(payloadJSON: string): ItemCompletedView {
  const payload = parsePayloadJSON(payloadJSON);
  const item = typeof payload === "string" ? null : findItemPayload(payload);
  const source = item ?? (isRecord(payload) ? payload : null);

  return {
    type: findStringField(source, "type") || "item/completed",
    command: findStringField(source, "command"),
    aggregatedOutput: source ? findAggregatedOutput(source) : "",
    text: findStringField(source, "text"),
    exitCode: findNumberField(source, "exitCode"),
  };
}

function parsePayloadJSON(payloadJSON: string): unknown {
  if (payloadJSON.trim() === "") {
    return {};
  }
  try {
    return JSON.parse(payloadJSON) as unknown;
  } catch {
    return payloadJSON;
  }
}

function findAggregatedOutput(value: unknown): string {
  if (typeof value === "string") {
    return "";
  }
  if (!value || typeof value !== "object") {
    return "";
  }
  if ("aggregatedOutput" in value && typeof value.aggregatedOutput === "string") {
    return value.aggregatedOutput;
  }
  if ("aggregated_output" in value && typeof value.aggregated_output === "string") {
    return value.aggregated_output;
  }
  for (const child of Object.values(value)) {
    const output = findAggregatedOutput(child);
    if (output !== "") {
      return output;
    }
  }
  return "";
}

function findItemPayload(value: unknown): Record<string, unknown> | null {
  if (!isRecord(value)) {
    return null;
  }
  if (isRecord(value.item)) {
    return value.item;
  }
  if (typeof value.type === "string") {
    return value;
  }
  for (const child of Object.values(value)) {
    const item = findItemPayload(child);
    if (item) {
      return item;
    }
  }
  return null;
}

function findStringField(value: unknown, field: string): string {
  if (!isRecord(value)) {
    return "";
  }
  const direct = value[field];
  if (typeof direct === "string") {
    return direct;
  }
  for (const child of Object.values(value)) {
    const result = findStringField(child, field);
    if (result !== "") {
      return result;
    }
  }
  return "";
}

function findNumberField(value: unknown, field: string): number | null {
  if (!isRecord(value)) {
    return null;
  }
  const direct = value[field];
  if (typeof direct === "number" && Number.isFinite(direct)) {
    return direct;
  }
  for (const child of Object.values(value)) {
    const result = findNumberField(child, field);
    if (result !== null) {
      return result;
    }
  }
  return null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
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
