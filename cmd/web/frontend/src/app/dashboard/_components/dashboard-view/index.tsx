import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchOrchestratorState } from "@/lib/api";
import {
  issueStatuses,
  priorities,
  type IssueStatus,
  type IssueSummary,
  type OrchestratorRunSummary,
  type OrchestratorState,
  type OrchestratorTokenSummary,
  type Priority,
  type Summary,
} from "@/lib/types";
import styles from "./index.module.css";

type DashboardViewProps = {
  summary: Summary;
  issues: IssueSummary[];
  refreshIntervalMs: number;
};

type OrchestratorLoadState =
  | { kind: "loading" }
  | { kind: "ready"; state: OrchestratorState }
  | { kind: "error"; message: string };

type DistributionRow<Key extends string> = {
  key: Key;
  label: string;
  count: number;
  percent: number;
};

export function DashboardView({ summary, issues, refreshIntervalMs }: DashboardViewProps) {
  const { t } = useTranslation();
  const [orchestratorState, setOrchestratorState] = useState<OrchestratorLoadState>({ kind: "loading" });
  const activeIssues = issues.filter((issue) => !isTerminalIssueStatus(issue.status)).length;
  const statusRows = useMemo(
    () => buildStatusDistribution(issues, t),
    [issues, t],
  );
  const priorityRows = useMemo(
    () => buildPriorityDistribution(issues, t),
    [issues, t],
  );
  const agentState = orchestratorState.kind === "ready" ? orchestratorState.state : null;
  const tokenTotals = tokenTotalsFromState(agentState);
  const runningAgents = agentState?.counts.running ?? null;
  const retryingAgents = agentState?.counts.retrying ?? null;

  useEffect(() => {
    let active = true;

    async function loadOrchestratorState() {
      try {
        const state = await fetchOrchestratorState({ silent: true });
        if (active) {
          setOrchestratorState({ kind: "ready", state });
        }
      } catch (error) {
        if (active) {
          setOrchestratorState({
            kind: "error",
            message: error instanceof Error ? error.message : t("dashboard.agentMetricsUnavailable"),
          });
        }
      }
    }

    void loadOrchestratorState();
    const id = window.setInterval(() => {
      void loadOrchestratorState();
    }, refreshIntervalMs);

    return () => {
      active = false;
      window.clearInterval(id);
    };
  }, [refreshIntervalMs, t]);

  return (
    <section className={styles.dashboard} aria-labelledby="dashboard-title">
      <header className={styles.header}>
        <div>
          <h2 className={styles.title} id="dashboard-title">{t("dashboard.title")}</h2>
          <p className={styles.syncText}>{t("dashboard.generatedAt")}: {summary.generatedAt || t("common.pending")}</p>
        </div>
      </header>

      <dl className={styles.metricGrid}>
        <MetricCard label={t("dashboard.totalIssues")} value={formatCount(issues.length)} />
        <MetricCard label={t("dashboard.activeIssues")} value={formatCount(activeIssues)} />
        <MetricCard label={t("dashboard.runningAgents")} value={formatOptionalCount(runningAgents)} />
        <MetricCard label={t("dashboard.retryingAgents")} value={formatOptionalCount(retryingAgents)} />
      </dl>

      <div className={styles.detailGrid}>
        <DistributionPanel
          title={t("dashboard.statusDistribution")}
          total={issues.length}
          rows={statusRows}
        />
        <DistributionPanel
          title={t("dashboard.priorityDistribution")}
          total={issues.length}
          rows={priorityRows}
        />
        <TokenPanel tokens={tokenTotals} />
        <RunningRunsPanel loadState={orchestratorState} />
      </div>
    </section>
  );
}

function MetricCard({
  label,
  value,
  variant = "default",
}: {
  label: string;
  value: string;
  variant?: "default" | "compact";
}) {
  return (
    <div className={variant === "compact" ? styles.compactMetricCard : styles.metricCard}>
      <dt className={styles.metricLabel}>{label}</dt>
      <dd className={variant === "compact" ? styles.compactMetricValue : styles.metricValue}>{value}</dd>
    </div>
  );
}

function DistributionPanel<Key extends string>({
  title,
  total,
  rows,
}: {
  title: string;
  total: number;
  rows: DistributionRow<Key>[];
}) {
  const { t } = useTranslation();

  return (
    <section className={styles.panel}>
      <div className={styles.panelHeader}>
        <h3 className={styles.panelTitle}>{title}</h3>
        <span className={styles.panelMeta}>{t("dashboard.totalCount", { count: total })}</span>
      </div>
      <div className={styles.distributionList}>
        {rows.map((row) => (
          <div key={row.key} className={styles.distributionRow}>
            <div className={styles.distributionMeta}>
              <span>{row.label}</span>
              <span>{row.count}</span>
            </div>
            <div
              className={styles.barTrack}
              aria-label={t("dashboard.distributionValue", {
                label: row.label,
                count: row.count,
                percent: row.percent,
              })}
            >
              <span
                className={styles.barFill}
                style={{ inlineSize: `${row.percent}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

function TokenPanel({ tokens }: { tokens: OrchestratorTokenSummary }) {
  const { t } = useTranslation();

  return (
    <section className={styles.panel}>
      <div className={styles.panelHeader}>
        <h3 className={styles.panelTitle}>{t("dashboard.tokenConsumption")}</h3>
        <span className={styles.panelMeta}>{t("dashboard.tokenSource")}</span>
      </div>
      <dl className={styles.tokenGrid}>
        <MetricCard label={t("dashboard.inputTokens")} value={formatCount(tokens.input_tokens)} variant="compact" />
        <MetricCard label={t("dashboard.outputTokens")} value={formatCount(tokens.output_tokens)} variant="compact" />
        <MetricCard label={t("dashboard.totalTokens")} value={formatCount(tokens.total_tokens)} variant="compact" />
      </dl>
    </section>
  );
}

function RunningRunsPanel({ loadState }: { loadState: OrchestratorLoadState }) {
  const { t } = useTranslation();
  const runs = loadState.kind === "ready" ? loadState.state.running : [];

  return (
    <section className={styles.panel}>
      <div className={styles.panelHeader}>
        <h3 className={styles.panelTitle}>{t("dashboard.runningRuns")}</h3>
        {loadState.kind === "loading" ? <span className={styles.panelMeta}>{t("layout.loading")}</span> : null}
      </div>
      {loadState.kind === "error" ? (
        <p className={styles.errorMessage}>{t("dashboard.agentMetricsUnavailable")}: {loadState.message}</p>
      ) : null}
      {loadState.kind === "ready" && runs.length === 0 ? (
        <p className={styles.emptyMessage}>{t("dashboard.noRunningRuns")}</p>
      ) : null}
      {runs.length > 0 ? <RunningRunsTable runs={runs} /> : null}
    </section>
  );
}

function RunningRunsTable({ runs }: { runs: OrchestratorRunSummary[] }) {
  const { t } = useTranslation();

  return (
    <div className={styles.tableWrap}>
      <table className={styles.runTable}>
        <thead>
          <tr>
            <th className={styles.tableHeadCell}>{t("dashboard.issueID")}</th>
            <th className={styles.tableHeadCell}>{t("dashboard.state")}</th>
            <th className={styles.tableHeadCell}>{t("dashboard.turnCount")}</th>
            <th className={styles.tableHeadCell}>{t("dashboard.lastEvent")}</th>
            <th className={styles.tableHeadCell}>{t("dashboard.startedAt")}</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={`${run.issue_identifier}-${run.session_id}-${run.started_at}`}>
              <td className={styles.tableCell}>{run.issue_id || run.issue_identifier}</td>
              <td className={styles.tableCell}>{t(`runStatuses.${run.state}`)}</td>
              <td className={styles.tableCell}>{run.turn_count}</td>
              <td className={styles.tableCell}>{run.last_event || t("common.pending")}</td>
              <td className={styles.tableCell}>{formatDateTime(run.started_at)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function isTerminalIssueStatus(status: IssueStatus): boolean {
  return status === "done" || status === "cancelled" || status === "duplicate";
}

function buildStatusDistribution(
  issues: IssueSummary[],
  t: (key: string) => string,
): DistributionRow<IssueStatus>[] {
  return issueStatuses.map((status) => {
    const count = issues.filter((issue) => issue.status === status).length;
    return {
      key: status,
      label: t(`statuses.${status}`),
      count,
      percent: percentOf(count, issues.length),
    };
  });
}

function buildPriorityDistribution(
  issues: IssueSummary[],
  t: (key: string) => string,
): DistributionRow<Priority>[] {
  return priorities.map((priority) => {
    const count = issues.filter((issue) => issue.priority === priority).length;
    return {
      key: priority,
      label: t(`priorities.${priority}`),
      count,
      percent: percentOf(count, issues.length),
    };
  });
}

function tokenTotalsFromState(state: OrchestratorState | null): OrchestratorTokenSummary {
  if (state?.codex_totals && isTokenSummary(state.codex_totals)) {
    return state.codex_totals;
  }

  return (state?.running ?? []).reduce<OrchestratorTokenSummary>(
    (totals, run) => ({
      input_tokens: totals.input_tokens + run.tokens.input_tokens,
      output_tokens: totals.output_tokens + run.tokens.output_tokens,
      total_tokens: totals.total_tokens + run.tokens.total_tokens,
    }),
    { input_tokens: 0, output_tokens: 0, total_tokens: 0 },
  );
}

function isTokenSummary(value: unknown): value is OrchestratorTokenSummary {
  if (!value || typeof value !== "object") {
    return false;
  }
  const tokenSummary = value as Partial<OrchestratorTokenSummary>;
  return (
    typeof tokenSummary.input_tokens === "number" &&
    typeof tokenSummary.output_tokens === "number" &&
    typeof tokenSummary.total_tokens === "number"
  );
}

function percentOf(count: number, total: number): number {
  if (total === 0) {
    return 0;
  }
  return Math.round((count / total) * 100);
}

function formatOptionalCount(value: number | null): string {
  return value === null ? "—" : formatCount(value);
}

function formatCount(value: number): string {
  return new Intl.NumberFormat().format(value);
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
