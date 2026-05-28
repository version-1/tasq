import type { RunSnapshot } from "@/lib/types";
import styles from "./index.module.css";

export function AgentsView({ runs }: { runs: RunSnapshot[] }) {
  return (
    <section className={styles.panelGrid}>
      <div className={styles.widePanel}>
        <h2>Agent Runs</h2>
        {runs.length === 0 ? (
          <p className={styles.empty}>No orchestrator runs yet</p>
        ) : (
          <div className={styles.agentList}>
            {runs.map((run) => (
              <article className={styles.agentRow} key={`${run.workItemId}-${run.runId}`}>
                <div>
                  <strong>Issue #{run.issueId}</strong>
                  <span>{run.workspace || "workspace pending"}</span>
                </div>
                <span className={styles.agentState}>{run.status}</span>
              </article>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
