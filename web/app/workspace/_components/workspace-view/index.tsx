import { useTranslation } from "react-i18next";
import type { IssueSummary, Summary } from "@/lib/types";
import styles from "./index.module.css";

type WorkspaceViewProps = {
  summary: Summary;
  issues: IssueSummary[];
};

export function WorkspaceView({ summary, issues }: WorkspaceViewProps) {
  const { t } = useTranslation();
  const activeIssues = issues.filter((issue) => issue.status !== "done").length;
  const activeRuns = summary.runs.filter((run) => run.status === "running" || run.status === "starting").length;

  return (
    <section className={styles.panelGrid}>
      <div className={styles.widePanel}>
        <h2>{t("workspace.title")}</h2>
        <dl className={styles.metricGrid}>
          <div>
            <dt>{t("workspace.totalIssues")}</dt>
            <dd>{issues.length}</dd>
          </div>
          <div>
            <dt>{t("workspace.activeIssues")}</dt>
            <dd>{activeIssues}</dd>
          </div>
          <div>
            <dt>{t("workspace.totalRuns")}</dt>
            <dd>{summary.runs.length}</dd>
          </div>
          <div>
            <dt>{t("workspace.activeRuns")}</dt>
            <dd>{activeRuns}</dd>
          </div>
        </dl>
        <p className={styles.generatedAt}>
          {t("workspace.generatedAt")}: {summary.generatedAt || t("common.pending")}
        </p>
      </div>
    </section>
  );
}
