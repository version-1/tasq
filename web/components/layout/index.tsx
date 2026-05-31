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
  type FormEvent,
  type ReactNode,
} from "react";
import { createIssue, fetchProjects, fetchSummary, updateIssueStatus } from "@/lib/api";
import "@/lib/i18n";
import { i18n, type SupportedLanguage } from "@/lib/i18n";
import { issueStatuses, priorities } from "@/lib/types";
import type { CreateIssueInput, IssueStatus, IssueSummary, Priority, Project, Summary } from "@/lib/types";
import { Header } from "./header";
import { Sidebar } from "./sidebar";
import styles from "./index.module.css";

export type TasqPage = "issues" | "agents" | "workspace" | "settings";

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
const defaultAddIssuePriority: Priority = "normal";

type AddIssueDialogState =
  | { kind: "closed" }
  | { kind: "open"; initialStatus: IssueStatus; error?: string };

type AddIssueFormValues = {
  title: string;
  description: string;
  status: IssueStatus;
  priority: Priority;
  assignee: string;
};

export function Layout({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const pathname = usePathname();
  const activePage = activePageFromPathname(pathname);
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
        />

        <AddIssueDialog
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

function PanelMessage({ title, detail }: { title: string; detail?: string }) {
  return (
    <section className={styles.widePanel}>
      <h2>{title}</h2>
      {detail ? <p>{detail}</p> : null}
    </section>
  );
}

function AddIssueDialog({
  state,
  onCancel,
  onSubmit,
}: {
  state: AddIssueDialogState;
  onCancel: () => void;
  onSubmit: (input: CreateIssueInput) => Promise<void>;
}) {
  const { t } = useTranslation();
  const initialStatus = state.kind === "open" ? state.initialStatus : "backlog";
  const [values, setValues] = useState<AddIssueFormValues>(() => initialAddIssueValues(initialStatus));
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationError, setValidationError] = useState("");

  useEffect(() => {
    if (state.kind === "open") {
      setValues(initialAddIssueValues(state.initialStatus));
      setValidationError("");
      setIsSubmitting(false);
    }
  }, [state.kind, initialStatus]);

  if (state.kind === "closed") {
    return null;
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const title = values.title.trim();
    if (!title) {
      setValidationError(t("addIssue.errors.titleRequired"));
      return;
    }

    setIsSubmitting(true);
    setValidationError("");
    await onSubmit(toCreateIssueInput(values, title));
    setIsSubmitting(false);
  }

  const errorMessage = validationError || state.error || "";

  return (
    <div className={styles.dialogBackdrop} role="presentation">
      <section
        aria-labelledby="add-issue-title"
        aria-modal="true"
        className={styles.dialog}
        role="dialog"
      >
        <form className={styles.addIssueForm} onSubmit={(event) => void handleSubmit(event)}>
          <div className={styles.dialogHeader}>
            <h2 id="add-issue-title">{t("addIssue.title")}</h2>
            <button type="button" aria-label={t("addIssue.close")} onClick={onCancel}>
              ×
            </button>
          </div>

          <label className={styles.field}>
            <span>{t("addIssue.fields.title")}</span>
            <input
              autoFocus
              value={values.title}
              onChange={(event) => setValues({ ...values, title: event.target.value })}
              placeholder={t("addIssue.placeholders.title")}
            />
          </label>

          <label className={styles.field}>
            <span>{t("addIssue.fields.description")}</span>
            <textarea
              value={values.description}
              onChange={(event) => setValues({ ...values, description: event.target.value })}
              placeholder={t("addIssue.placeholders.description")}
              rows={4}
            />
          </label>

          <div className={styles.formGrid}>
            <label className={styles.field}>
              <span>{t("addIssue.fields.status")}</span>
              <select
                value={values.status}
                onChange={(event) =>
                  setValues({ ...values, status: event.target.value as IssueStatus })
                }
              >
                {issueStatuses.map((status) => (
                  <option key={status} value={status}>
                    {t(`statuses.${status}`)}
                  </option>
                ))}
              </select>
            </label>

            <label className={styles.field}>
              <span>{t("addIssue.fields.priority")}</span>
              <select
                value={values.priority}
                onChange={(event) =>
                  setValues({ ...values, priority: event.target.value as Priority })
                }
              >
                {priorities.map((priority) => (
                  <option key={priority} value={priority}>
                    {t(`priorities.${priority}`)}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <label className={styles.field}>
            <span>{t("addIssue.fields.assignee")}</span>
            <input
              value={values.assignee}
              onChange={(event) => setValues({ ...values, assignee: event.target.value })}
              placeholder={t("addIssue.placeholders.assignee")}
            />
          </label>

          {errorMessage ? <p className={styles.formError}>{errorMessage}</p> : null}

          <div className={styles.dialogActions}>
            <button type="button" onClick={onCancel} disabled={isSubmitting}>
              {t("addIssue.cancel")}
            </button>
            <button type="submit" disabled={isSubmitting}>
              {isSubmitting ? t("addIssue.saving") : t("addIssue.submit")}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}

function initialAddIssueValues(status: IssueStatus): AddIssueFormValues {
  return {
    title: "",
    description: "",
    status,
    priority: defaultAddIssuePriority,
    assignee: "",
  };
}

function toCreateIssueInput(values: AddIssueFormValues, title: string): CreateIssueInput {
  return {
    title,
    description: values.description.trim() || undefined,
    status: values.status,
    priority: values.priority,
    assignee: values.assignee.trim() || undefined,
  };
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
