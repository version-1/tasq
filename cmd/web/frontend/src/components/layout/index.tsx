import { useLocation, useNavigate } from "react-router-dom";
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
import { modalIDs } from "@/constants";
import { ModalOutlet } from "@/components/ui/modal";
import { PanelMessage } from "@/components/ui/pannel-message";
import { ToastStack } from "@/components/ui/toast";
import {
  createIssue,
  createProject,
  fetchProjects,
  fetchSummary,
  updateIssueStatus,
} from "@/lib/api";
import "@/lib/i18n";
import { i18n, type SupportedLanguage } from "@/lib/i18n";
import { ModalProvider, useModal } from "@/lib/modal";
import { toast } from "@/lib/toast";
import type {
  CreateIssueInput,
  CreateProjectInput,
  IssueStatus,
  IssueSummary,
  Project,
  Summary,
} from "@/lib/types";
import { AddIssueDialog } from "./add-issue-dialog";
import { AddProjectDialog } from "./add-project-dialog";
import { Header } from "./header";
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
  addIssueError: string;
  addIssueInitialStatus: IssueStatus;
  addProjectError: string;
  isIssueDetailPage: boolean;
  isProjectIssueScope: boolean;
  issues: IssueSummary[];
  layoutData: LayoutData | null;
  loadState: LoadState;
  projects: Project[];
  summary: Summary | null;
  title: string | null;
  onAddIssue: (status?: IssueStatus) => void;
  onAddProject: () => void;
  onCloseModal: () => void;
  onCreateIssue: (input: CreateIssueInput) => Promise<void>;
  onCreateProject: (input: CreateProjectInput) => Promise<void>;
};

const layoutDataContext = createContext<LayoutData | null>(null);
const layoutShellContext = createContext<LayoutShellData | null>(null);

const defaultRefreshIntervalMs = 3000;
type LoadOptions = {
  silent?: boolean;
};

export function Layout({ children }: { children: ReactNode }) {
  return (
    <ModalProvider>
      <LayoutContent>{children}</LayoutContent>
    </ModalProvider>
  );
}

function LayoutContent({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const activePage = activePageFromPathname(pathname);
  const issueScope = issueScopeFromPathname(pathname);
  const isIssueDetailPage = /^\/issues\/\d+$/.test(pathname);
  const [loadState, setLoadState] = useState<LoadState>({ kind: "loading" });
  const [selectedIssueID, setSelectedIssueID] = useState<number | null>(null);
  const [addIssueInitialStatus, setAddIssueInitialStatus] =
    useState<IssueStatus>("backlog");
  const [addIssueError, setAddIssueError] = useState("");
  const [addProjectError, setAddProjectError] = useState("");
  const [refreshIntervalMs, setRefreshIntervalMs] = useState(
    defaultRefreshIntervalMs,
  );
  const [language, setLanguage] = useState<SupportedLanguage>(
    i18n.language === "en" ? "en" : "ja",
  );
  const modal = useModal();

  useEffect(() => {
    const stored = window.localStorage.getItem("tasq.refreshIntervalMs");
    const parsed = stored
      ? Number.parseInt(stored, 10)
      : defaultRefreshIntervalMs;
    if (Number.isFinite(parsed) && parsed >= 1000) {
      setRefreshIntervalMs(parsed);
    }
  }, []);

  useEffect(() => {
    document.documentElement.lang = language;
  }, [language]);

  const load = useCallback(
    async (options?: LoadOptions) => {
      try {
        const [summary, projects] = await Promise.all([
          fetchSummary(options),
          fetchProjects(options),
        ]);
        setLoadState({ kind: "ready", projects, summary });
        setSelectedIssueID((current) => current ?? firstIssueID(summary));
      } catch (error) {
        setLoadState((current) => {
          if (!options?.silent && current.kind === "ready") {
            return current;
          }
          return {
            kind: "error",
            message:
              error instanceof Error
                ? error.message
                : t("layout.failedToLoadSummary"),
          };
        });
      }
    },
    [t],
  );

  useEffect(() => {
    void load({ silent: true });
    const id = window.setInterval(() => {
      void load();
    }, refreshIntervalMs);
    return () => window.clearInterval(id);
  }, [load, refreshIntervalMs]);

  const loadedSummary = loadState.kind === "ready" ? loadState.summary : null;
  const projects = loadState.kind === "ready" ? loadState.projects : [];
  const isProjectIssueScope = issueScope.kind === "project";
  const scopedProject = isProjectIssueScope
    ? (projects.find((project) => project.key === issueScope.projectKey) ??
      null)
    : null;
  const activeProject = isProjectIssueScope
    ? scopedProject
    : activePage === "issues"
      ? null
      : (projects[0] ?? null);
  const summary = useMemo(() => {
    if (!loadedSummary) return null;
    if (!isProjectIssueScope) return loadedSummary;
    return scopedProject
      ? filterSummaryByProject(loadedSummary, scopedProject.id)
      : emptySummary(loadedSummary);
  }, [isProjectIssueScope, loadedSummary, scopedProject]);
  const issues = useMemo(() => {
    if (!summary) return [];
    return summary.columns.flatMap((column) => column.issues);
  }, [summary]);
  const selectedIssue =
    issues.find((issue) => issue.id === selectedIssueID) ?? issues[0] ?? null;

  async function handleStatusChange(id: number, status: IssueStatus) {
    try {
      await updateIssueStatus(id, status);
      await load();
      toast.success({ message: t("toast.success.issueStatusUpdated") });
    } catch {
      // Error toast is emitted by the API wrapper.
    }
  }

  async function handleCreateIssue(input: CreateIssueInput) {
    try {
      const created = await createIssue(input, { silent: true });
      await load({ silent: true });
      setSelectedIssueID(created.id);
      setAddIssueError("");
      modal.closeModal();
      toast.success({ message: t("toast.success.issueCreated") });
    } catch (error) {
      setAddIssueError(
        error instanceof Error
          ? error.message
          : t("layout.failedToCreateIssue"),
      );
    }
  }

  async function handleCreateProject(input: CreateProjectInput) {
    try {
      const created = await createProject(input, { silent: true });
      await load({ silent: true });
      setAddProjectError("");
      modal.closeModal();
      void navigate(`/projects/${encodeURIComponent(created.key)}/issues`);
      toast.success({ message: t("toast.success.projectCreated") });
    } catch (error) {
      setAddProjectError(
        error instanceof Error
          ? error.message
          : t("layout.failedToCreateProject"),
      );
    }
  }

  function handleAddIssue(status?: IssueStatus) {
    setAddIssueInitialStatus(status ?? "backlog");
    setAddIssueError("");
    modal.openModal(modalIDs.addIssue);
  }

  function handleAddProject() {
    setAddProjectError("");
    modal.openModal(modalIDs.addProject);
  }

  function handleCloseModal() {
    setAddIssueError("");
    setAddProjectError("");
    modal.closeModal();
  }

  function handleRefreshIntervalChange(nextIntervalMs: number) {
    setRefreshIntervalMs(nextIntervalMs);
    window.localStorage.setItem(
      "tasq.refreshIntervalMs",
      String(nextIntervalMs),
    );
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
        onAddIssue: handleAddIssue,
        onStatusChange: handleStatusChange,
      }
    : null;

  const shellData: LayoutShellData = {
    activePage,
    activeProject,
    addIssueError,
    addIssueInitialStatus,
    addProjectError,
    isIssueDetailPage,
    isProjectIssueScope,
    issues,
    layoutData,
    loadState,
    projects,
    summary,
    title: issueScopeTitle(
      issueScope,
      activeProject?.name ?? null,
      t("sidebar.allProjects"),
    ),
    onAddIssue: handleAddIssue,
    onAddProject: handleAddProject,
    onCloseModal: handleCloseModal,
    onCreateIssue: handleCreateIssue,
    onCreateProject: handleCreateProject,
  };

  return (
    <layoutShellContext.Provider value={shellData}>
      {children}
      <ToastStack />
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
  showAddTaskButton = true,
  showViewNavigation = true,
}: {
  children: ReactNode;
  shellData: LayoutShellData;
  showAddTaskButton?: boolean;
  showViewNavigation?: boolean;
}) {
  const { t } = useTranslation();

  return (
    <div className={styles.appFrame}>
      <Sidebar
        activePage={shellData.activePage}
        activeProjectID={
          shellData.isProjectIssueScope
            ? (shellData.activeProject?.id ?? null)
            : null
        }
        onAddProject={shellData.onAddProject}
        projects={shellData.projects}
      />
      <main className={styles.shell}>
        <Header
          activePage={shellData.activePage}
          projectName={shellData.title}
          issueCount={shellData.summary ? shellData.issues.length : null}
          onAddTask={() => shellData.onAddIssue("backlog")}
          showAddTaskButton={showAddTaskButton}
          showViewNavigation={showViewNavigation}
        />

        <ModalOutlet>
          <LayoutModalContent shellData={shellData} />
        </ModalOutlet>

        <div className={styles.content}>
          {shellData.loadState.kind === "loading" ? (
            <PanelMessage title={t("layout.loading")} />
          ) : null}
          {shellData.loadState.kind === "error" ? (
            <PanelMessage
              title={t("layout.apiUnavailable")}
              detail={shellData.loadState.message}
            />
          ) : null}
          {shellData.layoutData ? (
            <layoutDataContext.Provider value={shellData.layoutData}>
              {children}
            </layoutDataContext.Provider>
          ) : null}
        </div>
      </main>
    </div>
  );
}

function LayoutModalContent({ shellData }: { shellData: LayoutShellData }) {
  const modal = useModal();

  if (modal.activeModalID === modalIDs.addIssue) {
    return (
      <AddIssueDialog
        error={shellData.addIssueError}
        initialStatus={shellData.addIssueInitialStatus}
        project={shellData.activeProject}
        onCancel={shellData.onCloseModal}
        onSubmit={shellData.onCreateIssue}
      />
    );
  }

  if (modal.activeModalID === modalIDs.addProject) {
    return (
      <AddProjectDialog
        error={shellData.addProjectError}
        onCancel={shellData.onCloseModal}
        onSubmit={shellData.onCreateProject}
      />
    );
  }

  return null;
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
  const match = /^\/projects\/([^/]+)(?:\/(?:issues|settings))?\/?$/.exec(
    pathname,
  );
  if (!match) {
    return { kind: "all" };
  }
  return { kind: "project", projectKey: decodeURIComponent(match[1]) };
}

function issueScopeTitle(
  issueScope: IssueScope,
  activeProjectName: string | null,
  allProjectsTitle: string,
): string | null {
  if (issueScope.kind === "project") {
    return activeProjectName ?? issueScope.projectKey;
  }
  return activeProjectName ?? allProjectsTitle;
}
