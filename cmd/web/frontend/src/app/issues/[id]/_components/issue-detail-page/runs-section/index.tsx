"use client";

import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { OrchestratorIssueRun } from "@/lib/types";
import styles from "./index.module.css";

type RunsSectionProps = {
  issueID: number;
  error: string;
  isLoading: boolean;
  runs: OrchestratorIssueRun[];
};

export function RunsSection({ issueID, error, isLoading, runs }: RunsSectionProps) {
  const { t } = useTranslation();

  return (
    <section className={styles.section} aria-labelledby="issue-runs">
      <div className={styles.sectionHeader}>
        <h3 id="issue-runs">{t("issues.detailPage.runs")}</h3>
      </div>
      {isLoading ? <p className={styles.muted}>{t("layout.loading")}</p> : null}
      {error ? <p className={styles.error}>{error}</p> : null}
      {!isLoading && !error && runs.length === 0 ? (
        <p className={styles.muted}>{t("issues.detailPage.noRuns")}</p>
      ) : null}
      {runs.length > 0 ? (
        <div className={styles.runList}>
          {runs.map((run) => (
            <Link
              key={run.run_id}
              className={styles.runRow}
              to={`/issues/${issueID}/runs/${encodeURIComponent(run.run_id)}/conversations`}
            >
              <span className={styles.runID}>{run.run_id}</span>
              <span className={styles.runMeta}>
                {t(`runStatuses.${run.status}`)} · {t("issues.attempt")} {run.attempt}
              </span>
              <span className={styles.runTime}>{formatDateTime(run.updated_at)}</span>
            </Link>
          ))}
        </div>
      ) : null}
    </section>
  );
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
