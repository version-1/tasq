"use client";

import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Pagination } from "@/components/ui/pagination";
import {
  IssueTable,
  type IssueTableSortBy,
  type IssueTableSortDirection,
} from "@/components/ui/table";
import { IssueFilterOptions } from "@/features/issues/components/filter-options";
import { fetchIssues } from "@/lib/api";
import {
  issueStatuses,
  priorities,
  type IssueListResponse,
  type IssueStatus,
  type Priority,
  type Project,
} from "@/lib/types";
import styles from "./index.module.css";

const pageSize = 50;
const defaultStatuses = ["backlog", "ready", "in_progress"] satisfies IssueStatus[];

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; response: IssueListResponse }
  | { kind: "error"; message: string };

export function IssuesTableView({
  projectOptions,
  projectID,
  refreshIntervalMs,
}: {
  projectOptions: Project[];
  projectID: number | null;
  refreshIntervalMs: number;
}) {
  const { t } = useTranslation();
  const [selectedStatuses, setSelectedStatuses] = useState<IssueStatus[]>([...defaultStatuses]);
  const [selectedProjectIDs, setSelectedProjectIDs] = useState<number[]>([]);
  const [selectedPriorities, setSelectedPriorities] = useState<Priority[]>([]);
  const [sortBy, setSortBy] = useState<IssueTableSortBy>("updated_at");
  const [sortDirection, setSortDirection] = useState<IssueTableSortDirection>("desc");
  const [offset, setOffset] = useState(0);
  const [loadState, setLoadState] = useState<LoadState>({ kind: "loading" });
  const statesParam = selectedStatuses.length > 0 ? selectedStatuses.join(",") : undefined;
  const projectIDsParam = projectID === null && selectedProjectIDs.length > 0
    ? selectedProjectIDs.join(",")
    : undefined;
  const prioritiesParam = selectedPriorities.length > 0 ? selectedPriorities.join(",") : undefined;

  useEffect(() => {
    setOffset(0);
  }, [prioritiesParam, projectID, projectIDsParam, sortBy, sortDirection, statesParam]);

  useEffect(() => {
    let active = true;

    async function load() {
      try {
        const response = await fetchIssues(
          {
            limit: pageSize,
            offset,
            priorities: prioritiesParam,
            project_id: projectID ?? undefined,
            project_ids: projectIDsParam,
            sort_by: sortBy,
            sort_direction: sortDirection,
            states: statesParam,
          },
          { silent: true },
        );
        if (active) {
          setLoadState({ kind: "ready", response });
        }
      } catch (error) {
        if (active) {
          setLoadState({
            kind: "error",
            message: error instanceof Error ? error.message : t("issues.table.errorTitle"),
          });
        }
      }
    }

    void load();
    const id = window.setInterval(() => {
      void load();
    }, refreshIntervalMs);

    return () => {
      active = false;
      window.clearInterval(id);
    };
  }, [offset, prioritiesParam, projectID, projectIDsParam, refreshIntervalMs, sortBy, sortDirection, statesParam, t]);

  const response = loadState.kind === "ready" ? loadState.response : null;
  const issues = response?.data ?? [];
  const total = response?.meta.total ?? 0;
  const start = total === 0 ? 0 : offset + 1;
  const end = Math.min(offset + issues.length, total);
  const pageSummary = t("issues.table.pageSummary", { start, end, total });

  function selectedCountLabel(count: number) {
    return t("issues.table.selectedCount", { count });
  }

  function handleResetFilters() {
    setSelectedStatuses([...defaultStatuses]);
    setSelectedProjectIDs([]);
    setSelectedPriorities([]);
  }

  function handleStatusChange(values: Array<string | number>) {
    setSelectedStatuses(values as IssueStatus[]);
  }

  function handleProjectChange(values: Array<string | number>) {
    setSelectedProjectIDs(values.map((value) => Number(value)));
  }

  function handlePriorityChange(values: Array<string | number>) {
    setSelectedPriorities(values as Priority[]);
  }

  return (
    <section className={styles.tableView} aria-label={t("issues.table.tableTab")}>
      <div className={styles.toolbar}>
        <fieldset className={styles.filters}>
          <legend>{t("issues.table.filters")}</legend>
          <div className={styles.filterGroup}>
            <IssueFilterOptions
              allLabel={t("issues.table.allStatuses")}
              applyLabel={t("issues.table.apply")}
              cancelLabel={t("issues.table.cancel")}
              clearLabel={t("issues.table.clearAll")}
              label={t("issues.table.status")}
              onChange={handleStatusChange}
              options={issueStatuses.map((status) => ({ label: status, value: status }))}
              selectedCountLabel={selectedCountLabel}
              selectedValues={selectedStatuses}
            />
            {projectID === null ? (
              <IssueFilterOptions
                allLabel={t("issues.table.allProjects")}
                applyLabel={t("issues.table.apply")}
                cancelLabel={t("issues.table.cancel")}
                clearLabel={t("issues.table.clearAll")}
                label={t("issues.table.project")}
                onChange={handleProjectChange}
                options={projectOptions.map((project) => ({ label: project.name, value: project.id }))}
                selectedCountLabel={selectedCountLabel}
                selectedValues={selectedProjectIDs}
              />
            ) : null}
            <IssueFilterOptions
              allLabel={t("issues.table.allPriorities")}
              applyLabel={t("issues.table.apply")}
              cancelLabel={t("issues.table.cancel")}
              clearLabel={t("issues.table.clearAll")}
              label={t("issues.table.priority")}
              onChange={handlePriorityChange}
              options={priorities.map((priority) => ({ label: t(`priorities.${priority}`), value: priority }))}
              selectedCountLabel={selectedCountLabel}
              selectedValues={selectedPriorities}
            />
          </div>
          <button type="button" className={styles.resetButton} onClick={handleResetFilters}>
            {t("issues.table.reset")}
          </button>
        </fieldset>
      </div>

      {loadState.kind === "loading" ? <p className={styles.message}>{t("issues.table.loading")}</p> : null}
      {loadState.kind === "error" ? (
        <p className={`${styles.message} ${styles.error}`}>{loadState.message}</p>
      ) : null}
      {loadState.kind === "ready" ? (
        <>
          <IssueTable
            issues={issues}
            sortBy={sortBy}
            sortDirection={sortDirection}
            onSortChange={(nextSortBy) => {
              if (nextSortBy === sortBy) {
                setSortDirection((current) => current === "asc" ? "desc" : "asc");
                return;
              }
              setSortBy(nextSortBy);
              setSortDirection("desc");
            }}
          />
          {issues.length === 0 ? <p className={styles.message}>{t("issues.table.empty")}</p> : null}
          <Pagination
            nextDisabled={response?.meta.nextOffset === null}
            nextLabel={t("issues.table.next")}
            onNext={() => setOffset(response?.meta.nextOffset ?? offset)}
            onPrevious={() => setOffset(Math.max(0, offset - pageSize))}
            previousDisabled={offset === 0}
            previousLabel={t("issues.table.previous")}
            summary={pageSummary}
          />
        </>
      ) : null}
    </section>
  );
}
