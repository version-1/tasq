"use client";

import { useParams, useSearchParams } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchOrchestratorConversation, fetchOrchestratorIssueRuntime } from "@/lib/api";
import type { OrchestratorConversation, OrchestratorConversationEvent } from "@/lib/types";
import { EventCard } from "./event-card";
import styles from "./index.module.css";

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; conversation: OrchestratorConversation }
  | { kind: "empty" }
  | { kind: "error"; message: string };

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

export function ConversationEvents({ events }: { events: OrchestratorConversationEvent[] }) {
  const { t } = useTranslation();

  if (events.length === 0) {
    return <p className={styles.muted}>{t("issues.detailPage.noConversationEvents")}</p>;
  }

  return (
    <ol className={styles.timeline}>
      {events.map((event, index) => (
        <li key={`${event.at}-${event.event}-${index}`}>
          <EventCard event={event} />
        </li>
      ))}
    </ol>
  );
}

function parseIssueID(value: string | undefined): number | null {
  const id = value ? Number.parseInt(value, 10) : Number.NaN;
  return Number.isSafeInteger(id) && id > 0 ? id : null;
}
