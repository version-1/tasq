import { useTranslation } from "react-i18next";
import type { RunSnapshot } from "@/lib/types";
import styles from "./index.module.css";

export function AgentsView({ runs }: { runs: RunSnapshot[] }) {
  const { t } = useTranslation();

  return (
    <section className={styles.panelGrid}>
      <div className={styles.widePanel}>
        <h2>{t("agents.title")}</h2>
        {runs.length === 0 ? (
          <p className={styles.empty}>{t("agents.empty")}</p>
        ) : (
          <div className={styles.agentList}>
            {runs.map((run) => (
              <article className={styles.agentRow} key={`${run.workItemId}-${run.runId}`}>
                <div>
                  <strong>{t("agents.issueLabel", { id: run.issueId })}</strong>
                  <span>{run.workspace || t("agents.workspacePending")}</span>
                </div>
                <span className={styles.agentState}>{t(`runStatuses.${run.status}`)}</span>
              </article>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
