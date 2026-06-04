import { useTranslation } from "react-i18next";
import type { IssueSummary, Summary } from "@/lib/types";
import styles from "./index.module.css";

type DashboardViewProps = {
  summary: Summary;
  issues: IssueSummary[];
};

export function DashboardView({ summary, issues }: DashboardViewProps) {
	const { t } = useTranslation();
	const activeIssues = issues.filter((issue) => issue.status !== "done").length;

	return (
    <section className={styles.panelGrid}>
      <div className={styles.widePanel}>
        <h2>{t("dashboard.title")}</h2>
        <dl className={styles.metricGrid}>
          <div>
            <dt>{t("dashboard.totalIssues")}</dt>
            <dd>{issues.length}</dd>
          </div>
          <div>
            <dt>{t("dashboard.activeIssues")}</dt>
            <dd>{activeIssues}</dd>
          </div>
				</dl>
        <p className={styles.generatedAt}>
          {t("dashboard.generatedAt")}: {summary.generatedAt || t("common.pending")}
        </p>
      </div>
    </section>
  );
}
