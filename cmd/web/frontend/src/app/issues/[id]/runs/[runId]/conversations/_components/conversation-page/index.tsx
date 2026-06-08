"use client";

import { Link, useParams } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/issue/markdown";
import { fetchOrchestratorConversation } from "@/lib/api";
import type { OrchestratorConversation, OrchestratorConversationEvent } from "@/lib/types";
import styles from "./index.module.css";

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; conversation: OrchestratorConversation }
  | { kind: "error"; message: string };

export function ConversationPage() {
  const { t } = useTranslation();
  const { id, runId } = useParams();
  const issueID = parseIssueID(id);
  const [state, setState] = useState<LoadState>({ kind: "loading" });

  useEffect(() => {
    if (!issueID || !runId) {
      setState({ kind: "error", message: t("issues.detailPage.invalidIssue") });
      return;
    }

    let active = true;
    setState({ kind: "loading" });
    void fetchOrchestratorConversation(issueID, runId)
      .then((conversation) => {
        if (active) {
          setState({ kind: "ready", conversation });
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
  }, [issueID, runId, t]);

  const title = useMemo(() => {
    if (!runId) {
      return t("issues.detailPage.conversation");
    }
    return `${t("issues.detailPage.conversation")} · ${runId}`;
  }, [runId, t]);

  return (
    <div className={styles.page}>
      <Link className={styles.backLink} to={issueID ? `/issues/${issueID}` : "/issues"}>
        {t("issues.detailPage.backToIssue")}
      </Link>
      <header className={styles.header}>
        <h2>{title}</h2>
      </header>
      {state.kind === "loading" ? <p className={styles.muted}>{t("layout.loading")}</p> : null}
      {state.kind === "error" ? <p className={styles.error}>{state.message}</p> : null}
      {state.kind === "ready" ? (
        <ConversationEvents events={state.conversation.events} />
      ) : null}
    </div>
  );
}

function ConversationEvents({ events }: { events: OrchestratorConversationEvent[] }) {
  const { t } = useTranslation();

  if (events.length === 0) {
    return <p className={styles.muted}>{t("issues.detailPage.noConversationEvents")}</p>;
  }

  return (
    <ol className={styles.timeline}>
      {events.map((event, index) => (
        <li key={`${event.at}-${event.event}-${index}`} className={styles.timelineItem}>
          <div className={styles.eventHeader}>
            <span className={styles.eventType}>{eventLabel(event.event, t)}</span>
            <span className={styles.eventTime}>{formatDateTime(event.at)}</span>
          </div>
          {event.event === "turn_completed" ? (
            <Markdown
              className={styles.markdown}
              content={extractAggregatedOutput(event.payload_json)}
              emptyText={event.message || t("issues.detailPage.emptyConversationEvent")}
            />
          ) : (
            <p className={styles.statusMessage}>
              {event.message || t(`runStatuses.${event.event}`)}
            </p>
          )}
        </li>
      ))}
    </ol>
  );
}

function eventLabel(event: OrchestratorConversationEvent["event"], t: (key: string) => string): string {
  if (event === "turn_completed") {
    return t("issues.detailPage.turnCompleted");
  }
  return t(`runStatuses.${event}`);
}

function extractAggregatedOutput(payloadJSON: string): string {
  if (payloadJSON.trim() === "") {
    return "";
  }
  try {
    const payload = JSON.parse(payloadJSON) as unknown;
    return findAggregatedOutput(payload);
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
