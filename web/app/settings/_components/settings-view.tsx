import styles from "./settings-view.module.css";

export function SettingsView({
  refreshIntervalMs,
  generatedAt,
  onRefreshIntervalChange,
}: {
  refreshIntervalMs: number;
  generatedAt: string;
  onRefreshIntervalChange: (intervalMs: number) => void;
}) {
  return (
    <section className={styles.panelGrid}>
      <form className={`${styles.widePanel} ${styles.settingsForm}`} onSubmit={(event) => event.preventDefault()}>
        <h2>Web UI Settings</h2>
        <label>
          Refresh interval
          <select
            value={refreshIntervalMs}
            onChange={(event) => onRefreshIntervalChange(Number(event.target.value))}
          >
            <option value={1000}>1 second</option>
            <option value={3000}>3 seconds</option>
            <option value={5000}>5 seconds</option>
            <option value={10000}>10 seconds</option>
          </select>
        </label>
        <label>
          API origin
          <input value={apiOriginLabel()} readOnly />
        </label>
        <label>
          Last summary
          <input value={generatedAt || "pending"} readOnly />
        </label>
      </form>
    </section>
  );
}

function apiOriginLabel(): string {
  return process.env.NEXT_PUBLIC_ISSUE_TRACKER_URL || window.location.origin;
}
