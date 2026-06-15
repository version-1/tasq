"use client";

import { useParams, useSearchParams } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/issue/markdown";
import { fetchOrchestratorConversation, fetchOrchestratorIssueRuntime } from "@/lib/api";
import type { OrchestratorConversation, OrchestratorConversationEvent } from "@/lib/types";
import styles from "./index.module.css";

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; conversation: OrchestratorConversation }
  | { kind: "empty" }
  | { kind: "error"; message: string };

type Translator = (key: string, options?: Record<string, unknown>) => string;

type ApprovalRequestDetails = {
  reason: string;
  command: string;
};

export function ConversationPage() {
  const { t } = useTranslation();
  const { id, runId } = useParams();
  const [searchParams] = useSearchParams();
  const issueID = parseIssueID(id);
  const selectedRunID = runId ?? searchParams.get("runId");
  const [state, setState] = useState<LoadState>({ kind: "loading" });

  useEffect(() => {
    if (!issueID) {
      setState({ kind: "error", message: t("issues.detailPage.invalidIssue") });
      return;
    }

    let active = true;
    setState({ kind: "loading" });
    void resolveConversation(issueID, selectedRunID)
      .then((conversation) => {
        if (active) {
          setState(conversation ? { kind: "ready", conversation } : { kind: "empty" });
        }
      })
      .catch((error) => {
        if (active) {
          setState({
            kind: "error",
            message:
              error instanceof Error ? error.message : t("issues.detailPage.failedToLoadConversation"),
          });
        }
      });
    return () => {
      active = false;
    };
  }, [issueID, selectedRunID, t]);

  const title = useMemo(() => {
    if (state.kind !== "ready") {
      return t("issues.detailPage.conversation");
    }
    return `${t("issues.detailPage.conversation")} · ${state.conversation.run_id}`;
  }, [state, t]);

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h2>{title}</h2>
      </header>
      {state.kind === "loading" ? <p className={styles.muted}>{t("layout.loading")}</p> : null}
      {state.kind === "error" ? <p className={styles.error}>{state.message}</p> : null}
      {state.kind === "empty" ? (
        <p className={styles.muted}>{t("issues.detailPage.noRuns")}</p>
      ) : null}
      {state.kind === "ready" ? (
        <ConversationEvents events={state.conversation.events} />
      ) : null}
    </div>
  );
}

async function resolveConversation(
  issueID: number,
  selectedRunID: string | null,
): Promise<OrchestratorConversation | null> {
  if (selectedRunID) {
    return fetchOrchestratorConversation(issueID, selectedRunID, { silent: true });
  }

  const runtime = await fetchOrchestratorIssueRuntime(issueID, { silent: true });
  const latestRun = runtime.runs[0];
  if (!latestRun) {
    return null;
  }
  return fetchOrchestratorConversation(issueID, latestRun.run_id, { silent: true });
}

function ConversationEvents({ events }: { events: OrchestratorConversationEvent[] }) {
  const { t } = useTranslation();

  if (events.length === 0) {
    return <p className={styles.muted}>{t("issues.detailPage.noConversationEvents")}</p>;
  }

  return (
    <ol className={styles.timeline}>
      {events.map((event, index) => (
        <li
          key={`${event.at}-${event.event}-${index}`}
          className={[
            styles.timelineItem,
            event.event === "item/commandExecution/requestApproval" ? styles.approvalItem : "",
          ].join(" ")}
        >
          <div className={styles.eventHeader}>
            <EventTitle event={event} t={t} />
            <span className={styles.eventTime}>{formatDateTime(event.at)}</span>
          </div>
          <EventBody event={event} t={t} />
        </li>
      ))}
    </ol>
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
      <span className={[styles.eventType, styles.approvalBadge].join(" ")}>
        {eventLabel(event.event, t)}
      </span>
    );
  }

  if (event.event !== "item/completed") {
    return <span className={styles.eventType}>{eventLabel(event.event, t)}</span>;
  }

  const item = itemCompletedView(event.payload_json);
  return (
    <span className={styles.eventType}>
      {item.type}
      {item.command ? <code className={styles.eventCommand}>{item.command}</code> : null}
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

function parseIssueID(value: string | undefined): number | null {
  const id = value ? Number.parseInt(value, 10) : Number.NaN;
  return Number.isSafeInteger(id) && id > 0 ? id : null;
}
