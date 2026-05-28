"use client";

import { usePathname } from "next/navigation";
import { useTranslation } from "react-i18next";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { fetchSummary, updateIssueStatus } from "@/lib/api";
import "@/lib/i18n";
import { i18n, type SupportedLanguage } from "@/lib/i18n";
import type { IssueStatus, IssueSummary, Summary } from "@/lib/types";
import { Header } from "./header";
import { Sidebar } from "./sidebar";
import styles from "./index.module.css";

export type TasqPage = "issues" | "agents" | "workspace" | "settings";

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; summary: Summary }
  | { kind: "error"; message: string };

type LayoutData = {
  summary: Summary;
  issues: IssueSummary[];
  selectedIssue: IssueSummary | null;
  refreshIntervalMs: number;
  language: SupportedLanguage;
  onRefreshIntervalChange: (intervalMs: number) => void;
  onLanguageChange: (language: SupportedLanguage) => void;
  onSelectIssue: (issueID: number) => void;
  onStatusChange: (id: number, status: IssueStatus) => Promise<void>;
};

const layoutDataContext = createContext<LayoutData | null>(null);

const defaultRefreshIntervalMs = 3000;

export function Layout({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const pathname = usePathname();
  const activePage = activePageFromPathname(pathname);
  const [loadState, setLoadState] = useState<LoadState>({ kind: "loading" });
  const [selectedIssueID, setSelectedIssueID] = useState<number | null>(null);
  const [notice, setNotice] = useState("");
  const [refreshIntervalMs, setRefreshIntervalMs] = useState(defaultRefreshIntervalMs);
  const [language, setLanguage] = useState<SupportedLanguage>(i18n.language === "en" ? "en" : "ja");

  useEffect(() => {
    const stored = window.localStorage.getItem("tasq.refreshIntervalMs");
    const parsed = stored ? Number.parseInt(stored, 10) : defaultRefreshIntervalMs;
    if (Number.isFinite(parsed) && parsed >= 1000) {
      setRefreshIntervalMs(parsed);
    }
  }, []);

  useEffect(() => {
    document.documentElement.lang = language;
  }, [language]);

  const load = useCallback(async () => {
    try {
      const summary = await fetchSummary();
      setLoadState({ kind: "ready", summary });
      setSelectedIssueID((current) => current ?? firstIssueID(summary));
    } catch (error) {
      setLoadState({
        kind: "error",
        message: error instanceof Error ? error.message : t("layout.failedToLoadSummary"),
      });
    }
  }, [t]);

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
      setNotice(error instanceof Error ? error.message : t("layout.failedToUpdateIssue"));
    }
  }

  function handleRefreshIntervalChange(nextIntervalMs: number) {
    setRefreshIntervalMs(nextIntervalMs);
    window.localStorage.setItem("tasq.refreshIntervalMs", String(nextIntervalMs));
  }

  function handleLanguageChange(nextLanguage: SupportedLanguage) {
    setLanguage(nextLanguage);
    window.localStorage.setItem("tasq.language", nextLanguage);
    void i18n.changeLanguage(nextLanguage);
    document.documentElement.lang = nextLanguage;
  }

  const layoutData = summary
    ? {
        summary,
        issues,
        selectedIssue,
        refreshIntervalMs,
        language,
        onRefreshIntervalChange: handleRefreshIntervalChange,
        onLanguageChange: handleLanguageChange,
        onSelectIssue: setSelectedIssueID,
        onStatusChange: handleStatusChange,
      }
    : null;

  return (
    <div className={styles.appFrame}>
      <Sidebar />
      <main className={styles.shell}>
        <Header
          activePage={activePage}
          issueCount={summary ? issues.length : null}
          runCount={summary ? summary.runs.length : null}
          onRefresh={() => void load()}
        />

        {notice ? <p className={styles.notice}>{notice}</p> : null}

        {loadState.kind === "loading" ? <PanelMessage title={t("layout.loading")} /> : null}
        {loadState.kind === "error" ? (
          <PanelMessage title={t("layout.apiUnavailable")} detail={loadState.message} />
        ) : null}
        {layoutData ? (
          <layoutDataContext.Provider value={layoutData}>
            {children}
          </layoutDataContext.Provider>
        ) : null}
      </main>
    </div>
  );
}

export function useLayoutData(): LayoutData {
  const layoutData = useContext(layoutDataContext);
  if (!layoutData) {
    throw new Error(i18n.t("layout.useLayoutDataError"));
  }
  return layoutData;
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

function activePageFromPathname(pathname: string): TasqPage {
  const segment = pathname.split("/").filter(Boolean)[0];
  if (segment === "agents" || segment === "workspace" || segment === "settings") {
    return segment;
  }
  return "issues";
}
