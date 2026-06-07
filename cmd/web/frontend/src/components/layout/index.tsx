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

export type TasqPage = "issues" | "dashboard" | "settings";
type IssueScope = { kind: "all" } | { kind: "project"; projectKey: string };

export type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; projects: Project[]; summary: Summary }
  | { kind: "error"; message: string };

export type LayoutData = {
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

export type LayoutShellData = {
  activePage: TasqPage;
  activeProject: Project | null;
  addIssueState: AddIssueDialogState;
  isIssueDetailPage: boolean;
  issues: IssueSummary[];
  layoutData: LayoutData | null;
  loadState: LoadState;
  notice: string;
  projects: Project[];
  summary: Summary | null;
  title: string | null;
  onAddIssue: (status?: IssueStatus) => void;
  onCancelAddIssue: () => void;
  onCreateIssue: (input: CreateIssueInput) => Promise<void>;
};

const layoutDataContext = createContext<LayoutData | null>(null);
const layoutShellContext = createContext<LayoutShellData | null>(null);

const defaultRefreshIntervalMs = 3000;

export function Layout({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const activePage = activePageFromPathname(pathname);
  const issueScope = issueScopeFromPathname(pathname);
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

  const loadedSummary = loadState.kind === "ready" ? loadState.summary : null;
  const projects = loadState.kind === "ready" ? loadState.projects : [];
  const isProjectIssueScope = issueScope.kind === "project";
  const scopedProject =
    isProjectIssueScope
      ? projects.find((project) => project.key === issueScope.projectKey) ?? null
      : null;
  const activeProject = isProjectIssueScope
    ? scopedProject
    : activePage === "issues"
      ? null
      : projects[0] ?? null;
  const summary = useMemo(() => {
    if (!loadedSummary) return null;
    if (!isProjectIssueScope) return loadedSummary;
    return scopedProject ? filterSummaryByProject(loadedSummary, scopedProject.id) : emptySummary(loadedSummary);
  }, [isProjectIssueScope, loadedSummary, scopedProject]);
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

  const shellData: LayoutShellData = {
    activePage,
    activeProject,
    addIssueState,
    isIssueDetailPage,
    issues,
    layoutData,
    loadState,
    notice,
    projects,
    summary,
    title: issueScopeTitle(issueScope, activeProject?.name ?? null, t("sidebar.allProjects")),
    onAddIssue: (status) => setAddIssueState({ kind: "open", initialStatus: status ?? "backlog" }),
    onCancelAddIssue: () => setAddIssueState({ kind: "closed" }),
    onCreateIssue: handleCreateIssue,
  };

  return (
    <layoutShellContext.Provider value={shellData}>
      {children}
    </layoutShellContext.Provider>
  );
}

export function useLayoutData(): LayoutData {
  const layoutData = useContext(layoutDataContext);
  if (!layoutData) {
    throw new Error(i18n.t("layout.useLayoutDataError"));
  }
  return layoutData;
}

export function useLayoutShellData(): LayoutShellData {
  const shellData = useContext(layoutShellContext);
  if (!shellData) {
    throw new Error(i18n.t("layout.useLayoutDataError"));
  }
  return shellData;
}

export function ShellLayout({
  children,
  shellData,
  showViewNavigation,
}: {
  children: ReactNode;
  shellData: LayoutShellData;
  showViewNavigation: boolean;
}) {
  const { t } = useTranslation();

  return (
    <div className={styles.appFrame}>
      <Sidebar
        activePage={shellData.activePage}
        activeProjectID={shellData.activeProject?.id ?? null}
        projects={shellData.projects}
      />
      <main className={styles.shell}>
        <Header
          activePage={shellData.activePage}
          projectName={shellData.title}
          issueCount={shellData.summary ? shellData.issues.length : null}
          onAddTask={() => shellData.onAddIssue("backlog")}
          showViewNavigation={showViewNavigation}
        />

        <AddIssueDialog
          project={shellData.activeProject}
          state={shellData.addIssueState}
          onCancel={shellData.onCancelAddIssue}
          onSubmit={shellData.onCreateIssue}
        />

        {shellData.notice ? <p className={styles.notice}>{shellData.notice}</p> : null}

        {shellData.loadState.kind === "loading" ? <PanelMessage title={t("layout.loading")} /> : null}
        {shellData.loadState.kind === "error" ? (
          <PanelMessage title={t("layout.apiUnavailable")} detail={shellData.loadState.message} />
        ) : null}
        {shellData.layoutData ? (
          <layoutDataContext.Provider value={shellData.layoutData}>
            {children}
          </layoutDataContext.Provider>
        ) : null}
      </main>
    </div>
  );
}

function firstIssueID(summary: Summary): number | null {
  return summary.columns.flatMap((column) => column.issues)[0]?.id ?? null;
}

function filterSummaryByProject(summary: Summary, projectID: number): Summary {
  return {
    ...summary,
    columns: summary.columns.map((column) => ({
      ...column,
      issues: column.issues.filter((issue) => issue.projectId === projectID),
    })),
  };
}

function emptySummary(summary: Summary): Summary {
  return {
    ...summary,
    columns: summary.columns.map((column) => ({
      ...column,
      issues: [],
    })),
  };
}

function activePageFromPathname(pathname: string): TasqPage {
  const segment = pathname.split("/").filter(Boolean)[0];
  if (segment === "dashboard" || segment === "settings") {
    return segment;
  }
  return "issues";
}

function issueScopeFromPathname(pathname: string): IssueScope {
  const match = /^\/projects\/([^/]+)\/issues\/?$/.exec(pathname);
  if (!match) {
    return { kind: "all" };
  }
  return { kind: "project", projectKey: decodeURIComponent(match[1]) };
}

function issueScopeTitle(issueScope: IssueScope, activeProjectName: string | null, allProjectsTitle: string): string | null {
  if (issueScope.kind === "project") {
    return activeProjectName ?? issueScope.projectKey;
  }
  return activeProjectName ?? allProjectsTitle;
}
