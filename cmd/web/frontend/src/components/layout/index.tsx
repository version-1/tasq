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
  deleteProject,
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
  IssueStatus,
  IssueSummary,
  Project,
  Summary,
} from "@/lib/types";
import { AddIssueDialog } from "@/components/dialog/add-issue";
import { AddProjectDialog } from "@/components/dialog/add-project";
import { DeleteProjectDialog } from "@/components/dialog/delete-project";
import { ChangeRequestDialog } from "@/features/issues/components/change-request-dialog";
import { Header } from "./header";
import { Sidebar } from "./sidebar";
import { useTheme } from "./use-theme";
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
  onRejectIssue: (issueID: number) => void;
  onStatusChange: (id: number, status: IssueStatus) => Promise<void>;
};

export type LayoutShellData = {
  activePage: TasqPage;
  activeProject: Project | null;
  addIssueError: string;
  addIssueInitialStatus: IssueStatus;
  deleteProjectError: string;
  isDeletingProject: boolean;
  isMovingRejectedIssue: boolean;
  isIssueDetailPage: boolean;
  isProjectIssueScope: boolean;
  issues: IssueSummary[];
  layoutData: LayoutData | null;
  loadState: LoadState;
  projects: Project[];
  rejectIssue: IssueSummary | null;
  rejectIssueError: string;
  summary: Summary | null;
  title: string | null;
  onIssueDetailTitleChange: (title: string | null) => void;
  onAddIssue: (status?: IssueStatus) => void;
  onAddProject: () => void;
  onCloseModal: () => void;
  onCreateIssue: (input: CreateIssueInput) => Promise<void>;
  onDeleteProject: () => void;
  onConfirmDeleteProject: () => Promise<void>;
  onMoveRejectedIssueReady: () => Promise<void>;
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
  const issueDetailID = issueIDFromPathname(pathname);
  const isIssueDetailPage = issueDetailID !== null;
  const [loadState, setLoadState] = useState<LoadState>({ kind: "loading" });
  const [selectedIssueID, setSelectedIssueID] = useState<number | null>(null);
  const [addIssueInitialStatus, setAddIssueInitialStatus] =
    useState<IssueStatus>("backlog");
  const [addIssueError, setAddIssueError] = useState("");
  const [deleteProjectError, setDeleteProjectError] = useState("");
  const [isDeletingProject, setIsDeletingProject] = useState(false);
  const [rejectIssueID, setRejectIssueID] = useState<number | null>(null);
  const [rejectIssueError, setRejectIssueError] = useState("");
  const [isMovingRejectedIssue, setIsMovingRejectedIssue] = useState(false);
  const [issueDetailTitleOverride, setIssueDetailTitleOverride] = useState<string | null>(null);
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
    : activePage === "settings"
      ? (projects[0] ?? null)
      : null;
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
  const issueDetailTitle =
    issueDetailID === null
      ? null
      : (issues.find((issue) => issue.id === issueDetailID)?.title ?? null);

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

  function handleAddIssue(status?: IssueStatus) {
    setAddIssueInitialStatus(status ?? "backlog");
    setAddIssueError("");
    modal.openModal(modalIDs.addIssue);
  }

  function handleAddProject() {
    modal.openModal(modalIDs.addProject);
  }

  function handleDeleteProject() {
    setDeleteProjectError("");
    modal.openModal(modalIDs.deleteProject);
  }

  function handleRejectIssue(issueID: number) {
    setRejectIssueID(issueID);
    setRejectIssueError("");
    modal.openModal(modalIDs.rejectIssue);
  }

  async function handleMoveRejectedIssueReady() {
    if (rejectIssueID === null) return;
    setIsMovingRejectedIssue(true);
    setRejectIssueError("");
    try {
      await updateIssueStatus(rejectIssueID, "ready", { silent: true });
      await load({ silent: true });
      toast.success({ message: t("toast.success.issueRejected") });
    } catch (error) {
      const message = error instanceof Error ? error.message : t("layout.failedToRejectIssue");
      setRejectIssueError(message);
      throw new Error(message);
    } finally {
      setIsMovingRejectedIssue(false);
    }
  }

  async function handleConfirmDeleteProject() {
    if (!activeProject) return;
    setIsDeletingProject(true);
    setDeleteProjectError("");
    try {
      await deleteProject(activeProject.id, { silent: true });
      await load({ silent: true });
      modal.closeModal();
      navigate("/dashboard");
      toast.success({ message: t("toast.success.projectDeleted") });
    } catch (error) {
      setDeleteProjectError(
        error instanceof Error
          ? error.message
          : t("layout.failedToDeleteProject"),
      );
    } finally {
      setIsDeletingProject(false);
    }
  }

  function handleCloseModal() {
    setAddIssueError("");
    setDeleteProjectError("");
    setRejectIssueError("");
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

  const handleIssueDetailTitleChange = useCallback((title: string | null) => {
    setIssueDetailTitleOverride((current) => (current === title ? current : title));
  }, []);

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
        onRejectIssue: handleRejectIssue,
        onStatusChange: handleStatusChange,
      }
    : null;

  const shellData: LayoutShellData = {
    activePage,
    activeProject,
    addIssueError,
    addIssueInitialStatus,
    deleteProjectError,
    isIssueDetailPage,
    isDeletingProject,
    isMovingRejectedIssue,
    isProjectIssueScope,
    issues,
    layoutData,
    loadState,
    projects,
    rejectIssue: issues.find((issue) => issue.id === rejectIssueID) ?? null,
    rejectIssueError,
    summary,
    title: issueDetailTitleOverride ?? issueDetailTitle ?? issueScopeTitle(
      issueScope,
      activeProject?.name ?? null,
      t("sidebar.allProjects"),
    ),
    onIssueDetailTitleChange: handleIssueDetailTitleChange,
    onAddIssue: handleAddIssue,
    onAddProject: handleAddProject,
    onCloseModal: handleCloseModal,
    onCreateIssue: handleCreateIssue,
    onDeleteProject: handleDeleteProject,
    onConfirmDeleteProject: handleConfirmDeleteProject,
    onMoveRejectedIssueReady: handleMoveRejectedIssueReady,
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

export function useOptionalLayoutShellData(): LayoutShellData | null {
  return useContext(layoutShellContext);
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
  const { isDark, setIsDark } = useTheme();

  return (
    <div className={styles.appFrame}>
      <Sidebar
        activePage={shellData.activePage}
        activeProjectID={
          shellData.isProjectIssueScope
            ? (shellData.activeProject?.id ?? null)
            : null
        }
        isDarkMode={isDark}
        onAddProject={shellData.onAddProject}
        onDarkModeChange={setIsDark}
        projects={shellData.projects}
      />
      <main className={styles.shell}>
        <Header
          activePage={shellData.activePage}
          projectName={shellData.title}
          issueCount={shellData.summary ? shellData.issues.length : null}
          isIssueDetailPage={shellData.isIssueDetailPage}
          language={shellData.layoutData?.language ?? "ja"}
          onAddTask={() => shellData.onAddIssue("backlog")}
          onDeleteProject={shellData.onDeleteProject}
          onLanguageChange={shellData.layoutData?.onLanguageChange ?? (() => undefined)}
          canDeleteProject={shellData.isProjectIssueScope && shellData.activeProject !== null}
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
        dependencyOptions={shellData.issues}
        error={shellData.addIssueError}
        initialStatus={shellData.addIssueInitialStatus}
        project={shellData.activeProject}
        projects={shellData.projects}
        onCancel={shellData.onCloseModal}
        onSubmit={shellData.onCreateIssue}
      />
    );
  }

  if (modal.activeModalID === modalIDs.addProject) {
    return <AddProjectDialog onCancel={shellData.onCloseModal} />;
  }

  if (modal.activeModalID === modalIDs.deleteProject && shellData.activeProject) {
    return (
      <DeleteProjectDialog
        error={shellData.deleteProjectError}
        isDeleting={shellData.isDeletingProject}
        project={shellData.activeProject}
        onCancel={shellData.onCloseModal}
        onConfirm={shellData.onConfirmDeleteProject}
      />
    );
  }

  if (modal.activeModalID === modalIDs.rejectIssue && shellData.rejectIssue) {
    return (
      <ChangeRequestDialog
        error={shellData.rejectIssueError}
        isMovingIssue={shellData.isMovingRejectedIssue}
        issueID={shellData.rejectIssue.id}
        issueTitle={shellData.rejectIssue.title}
        onCancel={shellData.onCloseModal}
        onMoveIssueReady={shellData.onMoveRejectedIssueReady}
        onSuccess={shellData.onCloseModal}
        variant="reject"
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
  if (/^\/issues\/\d+(?:\/|$)/.test(pathname)) {
    return "dashboard";
  }
  return "issues";
}

function issueScopeFromPathname(pathname: string): IssueScope {
  const match = /^\/projects\/([^/]+)(?:\/(?:issues|settings|table))?\/?$/.exec(
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

function issueIDFromPathname(pathname: string): number | null {
  const match =
    /^\/issues\/(\d+)(?:\/(?:conversations|runs\/[^/]+\/conversations))?\/?$/.exec(
      pathname,
    );
  if (!match) {
    return null;
  }
  const id = Number.parseInt(match[1], 10);
  return Number.isSafeInteger(id) && id > 0 ? id : null;
}
