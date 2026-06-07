import { useLocation } from "react-router-dom";
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
import { createIssue, fetchProjects, fetchSummary, updateIssueStatus } from "@/lib/api";
import "@/lib/i18n";
import { i18n, type SupportedLanguage } from "@/lib/i18n";
import type { CreateIssueInput, IssueStatus, IssueSummary, Project, Summary } from "@/lib/types";
import { AddIssueDialog, type AddIssueDialogState } from "./add-issue-dialog";
import { Header } from "./header";
import { PanelMessage } from "./panel-message";
import { Sidebar } from "./sidebar";
import styles from "./index.module.css";

export type TasqPage = "issues" | "agents" | "dashboard" | "settings";

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; projects: Project[]; summary: Summary }
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
  onAddIssue: (status?: IssueStatus) => void;
  onStatusChange: (id: number, status: IssueStatus) => Promise<void>;
};

const layoutDataContext = createContext<LayoutData | null>(null);

const defaultRefreshIntervalMs = 3000;

export function Layout({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const activePage = activePageFromPathname(pathname);
  const isIssueDetailPage = /^\/issues\/\d+$/.test(pathname);
  const [loadState, setLoadState] = useState<LoadState>({ kind: "loading" });
  const [selectedIssueID, setSelectedIssueID] = useState<number | null>(null);
  const [notice, setNotice] = useState("");
  const [addIssueState, setAddIssueState] = useState<AddIssueDialogState>({ kind: "closed" });
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
      const [summary, projects] = await Promise.all([fetchSummary(), fetchProjects()]);
      setLoadState({ kind: "ready", projects, summary });
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
  const projects = loadState.kind === "ready" ? loadState.projects : [];
  const activeProject = projects[0] ?? null;
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

  async function handleCreateIssue(input: CreateIssueInput) {
    setNotice("");
    try {
      const created = await createIssue(input);
      await load();
      setSelectedIssueID(created.id);
      setAddIssueState({ kind: "closed" });
    } catch (error) {
      setAddIssueState((current) => ({
        kind: "open",
        initialStatus: current.kind === "open" ? current.initialStatus : "backlog",
        error: error instanceof Error ? error.message : t("layout.failedToCreateIssue"),
      }));
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

  const layoutData: LayoutData | null = summary
    ? {
        summary,
        issues,
        selectedIssue,
        refreshIntervalMs,
        language,
        onRefreshIntervalChange: handleRefreshIntervalChange,
        onLanguageChange: handleLanguageChange,
        onSelectIssue: setSelectedIssueID,
        onAddIssue: (status) => setAddIssueState({ kind: "open", initialStatus: status ?? "backlog" }),
        onStatusChange: handleStatusChange,
      }
    : null;

  return (
    <div className={styles.appFrame}>
      <Sidebar activeProjectID={activeProject?.id ?? null} projects={projects} />
      <main className={styles.shell}>
        <Header
          activePage={activePage}
          projectName={activeProject?.name ?? null}
          issueCount={summary ? issues.length : null}
          onAddTask={() => setAddIssueState({ kind: "open", initialStatus: "backlog" })}
          showViewNavigation={!isIssueDetailPage}
        />

        <AddIssueDialog
          project={activeProject}
          state={addIssueState}
          onCancel={() => setAddIssueState({ kind: "closed" })}
          onSubmit={handleCreateIssue}
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

function firstIssueID(summary: Summary): number | null {
  return summary.columns.flatMap((column) => column.issues)[0]?.id ?? null;
}

function activePageFromPathname(pathname: string): TasqPage {
  const segment = pathname.split("/").filter(Boolean)[0];
  if (segment === "agents" || segment === "dashboard" || segment === "settings") {
    return segment;
  }
  return "issues";
}
