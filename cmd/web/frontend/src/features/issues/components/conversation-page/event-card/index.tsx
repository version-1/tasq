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

  if (event.event === "thread/tokenUsage/updated") {
    return <TokenUsage event={event} t={t} />;
  }

  if (event.event === "account/rateLimits/updated") {
    return <RateLimits event={event} t={t} />;
  }

  return (
    <p className={styles.statusMessage}>
      {event.message || t(`runStatuses.${event.event}`)}
    </p>
  );
}

function TokenUsage({
  event,
  t,
}: {
  event: OrchestratorConversationEvent;
  t: Translator;
}) {
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

function RateLimits({
  event,
  t,
}: {
  event: OrchestratorConversationEvent;
  t: Translator;
}) {
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

function MetricGrid({ items }: { items: Array<[string, string]> }) {
  return (
    <dl className={styles.metricGrid}>
      {items.map(([label, value]) => (
        <div key={label} className={styles.metric}>
          <dt>{label}</dt>
          <dd>{value}</dd>
        </div>
      ))}
    </dl>
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

type TokenUsageBucket = {
  totalTokens: number;
  inputTokens: number;
  cachedInputTokens: number;
  outputTokens: number;
  reasoningOutputTokens: number;
};

type TokenUsageView = {
  total: TokenUsageBucket;
  last: TokenUsageBucket;
  modelContextWindow: number;
};

type RateLimitWindow = {
  usedPercent: number;
  windowDurationMins: number;
  resetsAt: number;
};

type RateLimitsView = {
  limitId: string;
  planType: string;
  primary: RateLimitWindow;
  secondary: RateLimitWindow;
  rateLimitReachedType: string;
};

function tokenUsageView(payloadJSON: string): TokenUsageView | null {
  const payload = parsePayloadJSON(payloadJSON);
  if (!isRecord(payload) || !isRecord(payload.tokenUsage)) {
    return null;
  }
  const total = tokenUsageBucket(payload.tokenUsage.total);
  const last = tokenUsageBucket(payload.tokenUsage.last);
  const modelContextWindow = findNumberField(payload.tokenUsage, "modelContextWindow");
  if (!total || !last || modelContextWindow === null) {
    return null;
  }
  return { total, last, modelContextWindow };
}

function tokenUsageBucket(value: unknown): TokenUsageBucket | null {
  if (!isRecord(value)) {
    return null;
  }
  const totalTokens = findNumberField(value, "totalTokens");
  const inputTokens = findNumberField(value, "inputTokens");
  const cachedInputTokens = findNumberField(value, "cachedInputTokens");
  const outputTokens = findNumberField(value, "outputTokens");
  const reasoningOutputTokens = findNumberField(value, "reasoningOutputTokens");
  if (
    totalTokens === null ||
    inputTokens === null ||
    cachedInputTokens === null ||
    outputTokens === null ||
    reasoningOutputTokens === null
  ) {
    return null;
  }
  return {
    totalTokens,
    inputTokens,
    cachedInputTokens,
    outputTokens,
    reasoningOutputTokens,
  };
}

function rateLimitsView(payloadJSON: string): RateLimitsView | null {
  const payload = parsePayloadJSON(payloadJSON);
  if (!isRecord(payload) || !isRecord(payload.rateLimits)) {
    return null;
  }
  const primary = rateLimitWindow(payload.rateLimits.primary);
  const secondary = rateLimitWindow(payload.rateLimits.secondary);
  const limitId = findStringField(payload.rateLimits, "limitId");
  const planType = findStringField(payload.rateLimits, "planType");
  if (!primary || !secondary || limitId === "" || planType === "") {
    return null;
  }
  return {
    limitId,
    planType,
    primary,
    secondary,
    rateLimitReachedType: findStringField(payload.rateLimits, "rateLimitReachedType"),
  };
}

function rateLimitWindow(value: unknown): RateLimitWindow | null {
  if (!isRecord(value)) {
    return null;
  }
  const usedPercent = findNumberField(value, "usedPercent");
  const windowDurationMins = findNumberField(value, "windowDurationMins");
  const resetsAt = findNumberField(value, "resetsAt");
  if (usedPercent === null || windowDurationMins === null || resetsAt === null) {
    return null;
  }
  return { usedPercent, windowDurationMins, resetsAt };
}

function formatRateLimitWindow(window: RateLimitWindow, t: Translator): string {
  return t("issues.detailPage.rateLimitWindow", {
    percent: window.usedPercent,
    minutes: formatNumber(window.windowDurationMins),
    resetsAt: formatUnixSeconds(window.resetsAt),
  });
}

function formatUnixSeconds(value: number): string {
  const date = new Date(value * 1000);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat().format(value);
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
