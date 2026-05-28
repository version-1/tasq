"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { AgentsView } from "@/app/agents/_components/agents-view";
import { IssuesView } from "@/app/issues/_components/issues-view";
import { SettingsView } from "@/app/settings/_components/settings-view";
import { fetchSummary, updateIssueStatus } from "@/lib/api";
import type { IssueStatus, IssueSummary, Summary } from "@/lib/types";
import styles from "./tasq-shell.module.css";

export type TasqPage = "issues" | "agents" | "settings";

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; summary: Summary }
  | { kind: "error"; message: string };

const pages = [
  { key: "issues", href: "/issues", title: "Issues" },
  { key: "agents", href: "/agents", title: "Agents" },
  { key: "settings", href: "/settings", title: "Settings" },
] as const;

const defaultRefreshIntervalMs = 3000;

export function TasqShell({ activePage }: { activePage: TasqPage }) {
  const [loadState, setLoadState] = useState<LoadState>({ kind: "loading" });
  const [selectedIssueID, setSelectedIssueID] = useState<number | null>(null);
  const [notice, setNotice] = useState("");
  const [refreshIntervalMs, setRefreshIntervalMs] = useState(defaultRefreshIntervalMs);

  useEffect(() => {
    const stored = window.localStorage.getItem("tasq.refreshIntervalMs");
    const parsed = stored ? Number.parseInt(stored, 10) : defaultRefreshIntervalMs;
    if (Number.isFinite(parsed) && parsed >= 1000) {
      setRefreshIntervalMs(parsed);
    }
  }, []);

  const load = useCallback(async () => {
    try {
      const summary = await fetchSummary();
      setLoadState({ kind: "ready", summary });
      setSelectedIssueID((current) => current ?? firstIssueID(summary));
    } catch (error) {
      setLoadState({
        kind: "error",
        message: error instanceof Error ? error.message : "failed to load summary",
      });
    }
  }, []);

  useEffect(() => {
    void load();
    const id = window.setInterval(() => {
      void load();
    }, refreshIntervalMs);
    return () => window.clearInterval(id);
  }, [load, refreshIntervalMs]);

  const summary = loadState.kind === "ready" ? loadState.summary : null;
  const issues = useMemo(() => {
    if (!summary) return [];
    return summary.columns.flatMap((column) => column.issues);
  }, [summary]);
  const selectedIssue =
    issues.find((issue) => issue.id === selectedIssueID) ?? issues[0] ?? null;

  async function handleStatusChange(id: number, status: IssueStatus) {
    setNotice("");
    try {
      await updateIssueStatus(id, status);
      await load();
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "failed to update issue");
    }
  }

  function handleRefreshIntervalChange(nextIntervalMs: number) {
    setRefreshIntervalMs(nextIntervalMs);
    window.localStorage.setItem("tasq.refreshIntervalMs", String(nextIntervalMs));
  }

  return (
    <main className={styles.shell}>
      <header className={styles.topbar}>
        <div>
          <p className={styles.eyebrow}>Tasq Issue Tracker</p>
          <h1>{pageHeading(activePage)}</h1>
        </div>
        <div className={styles.statusStrip}>
          <span>{summary ? `${issues.length} issues` : "loading"}</span>
          <span>{summary ? `${summary.runs.length} tracked runs` : "..."}</span>
          <button className={styles.iconButton} type="button" onClick={() => void load()} aria-label="Refresh">
            Refresh
          </button>
        </div>
      </header>

      <nav className={styles.tabs} aria-label="Views">
        {pages.map((page) => (
          <Link
            key={page.key}
            className={activePage === page.key ? `${styles.tab} ${styles.active}` : styles.tab}
            href={page.href}
          >
            {page.title}
          </Link>
        ))}
      </nav>

      {notice ? <p className={styles.notice}>{notice}</p> : null}

      {loadState.kind === "loading" ? <PanelMessage title="Loading" /> : null}
      {loadState.kind === "error" ? (
        <PanelMessage title="API unavailable" detail={loadState.message} />
      ) : null}
      {summary ? (
        <PageView
          activePage={activePage}
          summary={summary}
          selectedIssue={selectedIssue}
          refreshIntervalMs={refreshIntervalMs}
          onRefreshIntervalChange={handleRefreshIntervalChange}
          onSelectIssue={(issueID) => setSelectedIssueID(issueID)}
          onStatusChange={handleStatusChange}
        />
      ) : null}
    </main>
  );
}

function PageView({
  activePage,
  summary,
  selectedIssue,
  refreshIntervalMs,
  onRefreshIntervalChange,
  onSelectIssue,
  onStatusChange,
}: {
  activePage: TasqPage;
  summary: Summary;
  selectedIssue: IssueSummary | null;
  refreshIntervalMs: number;
  onRefreshIntervalChange: (intervalMs: number) => void;
  onSelectIssue: (issueID: number) => void;
  onStatusChange: (id: number, status: IssueStatus) => Promise<void>;
}) {
  switch (activePage) {
    case "agents":
      return <AgentsView runs={summary.runs} />;
    case "settings":
      return (
        <SettingsView
          refreshIntervalMs={refreshIntervalMs}
          generatedAt={summary.generatedAt}
          onRefreshIntervalChange={onRefreshIntervalChange}
        />
      );
    case "issues":
      return (
        <IssuesView
          summary={summary}
          selectedIssue={selectedIssue}
          onSelectIssue={onSelectIssue}
          onStatusChange={onStatusChange}
        />
      );
  }
}

function PanelMessage({ title, detail }: { title: string; detail?: string }) {
  return (
    <section className={styles.widePanel}>
      <h2>{title}</h2>
      {detail ? <p>{detail}</p> : null}
    </section>
  );
}

function firstIssueID(summary: Summary): number | null {
  return summary.columns.flatMap((column) => column.issues)[0]?.id ?? null;
}

function pageHeading(page: TasqPage): string {
  switch (page) {
    case "issues":
      return "Issue Operations";
    case "agents":
      return "Agent Operations";
    case "settings":
      return "Settings";
  }
}
