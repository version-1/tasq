"use client";

import { useTranslation } from "react-i18next";
import type { OrchestratorIssueRun } from "@/lib/types";
import styles from "./index.module.css";
import { RunRow } from "./run-row";

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
            <RunRow key={run.run_id} issueID={issueID} run={run} />
          ))}
        </div>
      ) : null}
    </section>
  );
}
